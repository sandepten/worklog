package summarizer

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sandepten/work-obsidian-noter/internal/notes"
)

// Client handles communication with the OpenCode server for AI summaries
type Client struct {
	baseURL    string
	providerID string
	modelID    string
	httpClient *http.Client
}

// NewClient creates a new OpenCode API client
func NewClient(baseURL, providerID, modelID string) *Client {
	return &Client{
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		providerID: providerID,
		modelID:    modelID,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// Session represents an OpenCode session
type Session struct {
	ID string `json:"id"`
}

// TextPart represents a text message part
type TextPart struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// PromptRequest represents the request body for sending a message
type PromptRequest struct {
	Model *ModelSpec `json:"model,omitempty"`
	Parts []TextPart `json:"parts"`
}

// ModelSpec specifies which model to use
type ModelSpec struct {
	ProviderID string `json:"providerID"`
	ModelID    string `json:"modelID"`
}

// MessageResponse represents a response from the API
type MessageResponse struct {
	Info  MessageInfo `json:"info"`
	Parts []Part      `json:"parts"`
}

// MessageInfo contains message metadata
type MessageInfo struct {
	ID   string `json:"id"`
	Role string `json:"role"`
}

// Part represents a message part in the response
type Part struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// SSEEvent represents a server-sent event
type SSEEvent struct {
	Type       string          `json:"type"`
	Properties json.RawMessage `json:"properties"`
}

// createSession creates a new session for summarization
func (c *Client) createSession() (*Session, error) {
	req, err := http.NewRequest("POST", c.baseURL+"/session", bytes.NewBuffer([]byte("{}")))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create session: status %d, body: %s", resp.StatusCode, string(body))
	}

	var session Session
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		return nil, fmt.Errorf("failed to decode session response: %w", err)
	}

	return &session, nil
}

// sendMessageAsync sends a message to a session (async - returns immediately)
func (c *Client) sendMessageAsync(sessionID string, prompt string) error {
	requestBody := PromptRequest{
		Model: &ModelSpec{
			ProviderID: c.providerID,
			ModelID:    c.modelID,
		},
		Parts: []TextPart{
			{Type: "text", Text: prompt},
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/session/%s/message", c.baseURL, sessionID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to send message: status %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

// waitForIdleWithPolling polls the messages endpoint until we get an assistant response
func (c *Client) waitForIdleWithPolling(sessionID string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for response")
		case <-ticker.C:
			messages, err := c.getMessages(sessionID)
			if err != nil {
				continue
			}

			// Check if we have an assistant message with content
			for _, msg := range messages {
				if msg.Info.Role == "assistant" {
					// Check if the message has text content
					for _, part := range msg.Parts {
						if part.Type == "text" && part.Text != "" {
							return nil // We have a response
						}
					}
				}
			}
		}
	}
}

// startEventListener starts listening to SSE events and returns a channel for idle notifications
func (c *Client) startEventListener(ctx context.Context, sessionID string) <-chan struct{} {
	idleChan := make(chan struct{}, 1)

	go func() {
		defer close(idleChan)

		req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/event", nil)
		if err != nil {
			return
		}
		req.Header.Set("Accept", "text/event-stream")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return
		}
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
			}

			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			var event SSEEvent
			if err := json.Unmarshal([]byte(data), &event); err != nil {
				continue
			}

			if event.Type == "session.idle" {
				var props struct {
					SessionID string `json:"sessionID"`
				}
				if err := json.Unmarshal(event.Properties, &props); err == nil {
					if props.SessionID == sessionID {
						select {
						case idleChan <- struct{}{}:
						default:
						}
						return
					}
				}
			}
		}
	}()

	return idleChan
}

// getMessages retrieves all messages from a session
func (c *Client) getMessages(sessionID string) ([]MessageResponse, error) {
	url := fmt.Sprintf("%s/session/%s/message", c.baseURL, sessionID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get messages: status %d, body: %s", resp.StatusCode, string(body))
	}

	var messages []MessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&messages); err != nil {
		return nil, fmt.Errorf("failed to decode messages: %w", err)
	}

	return messages, nil
}

// extractAssistantResponse extracts text from assistant messages
func (c *Client) extractAssistantResponse(messages []MessageResponse) string {
	var result strings.Builder

	for _, msg := range messages {
		if msg.Info.Role == "assistant" {
			for _, part := range msg.Parts {
				if part.Type == "text" && part.Text != "" {
					result.WriteString(part.Text)
				}
			}
		}
	}

	return strings.TrimSpace(result.String())
}

// SummarizeWorkItems generates an AI summary of completed work items
func (c *Client) SummarizeWorkItems(items []notes.WorkItem) (string, error) {
	if len(items) == 0 {
		return "No work items to summarize.", nil
	}

	// Build the prompt
	var sb strings.Builder
	sb.WriteString("Summarize the following completed work items in 1-2 concise sentences. Focus on the key accomplishments and outcomes. Keep it brief and professional. Do not use any tools, just respond with plain text:\n\n")

	for _, item := range items {
		sb.WriteString(fmt.Sprintf("- %s\n", item.Text))
	}

	// Create session
	session, err := c.createSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Start event listener BEFORE sending message
	idleChan := c.startEventListener(ctx, session.ID)

	// Small delay to ensure listener is ready
	time.Sleep(100 * time.Millisecond)

	// Send message asynchronously
	if err := c.sendMessageAsync(session.ID, sb.String()); err != nil {
		return "", fmt.Errorf("failed to send message: %w", err)
	}

	// Wait for either SSE idle event or timeout
	select {
	case <-idleChan:
		// Session is idle
	case <-ctx.Done():
		// Timeout - but let's still try to get messages in case we missed the event
	}

	// Get messages and extract response
	messages, err := c.getMessages(session.ID)
	if err != nil {
		return "", fmt.Errorf("failed to get messages: %w", err)
	}

	response := c.extractAssistantResponse(messages)
	if response == "" {
		// If no response via SSE, try polling
		if err := c.waitForIdleWithPolling(session.ID, 30*time.Second); err != nil {
			return "", fmt.Errorf("no response received from AI: %w", err)
		}

		// Try getting messages again
		messages, err = c.getMessages(session.ID)
		if err != nil {
			return "", fmt.Errorf("failed to get messages: %w", err)
		}

		response = c.extractAssistantResponse(messages)
		if response == "" {
			return "", fmt.Errorf("no response received from AI")
		}
	}

	return response, nil
}

// TestConnection tests if the OpenCode server is reachable
func (c *Client) TestConnection() error {
	req, err := http.NewRequest("GET", c.baseURL+"/global/health", nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to OpenCode server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("OpenCode server returned status %d", resp.StatusCode)
	}

	return nil
}
