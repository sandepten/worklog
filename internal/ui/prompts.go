package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/manifoldco/promptui"
	"github.com/sandepten/work-obsidian-noter/internal/notes"
)

// Prompter handles interactive CLI prompts
type Prompter struct{}

// NewPrompter creates a new prompter
func NewPrompter() *Prompter {
	return &Prompter{}
}

// ConfirmCompletion asks if a work item was completed
func (p *Prompter) ConfirmCompletion(item notes.WorkItem) (bool, error) {
	prompt := promptui.Prompt{
		Label:     fmt.Sprintf("Did you complete: \"%s\"", item.Text),
		IsConfirm: true,
	}

	_, err := prompt.Run()
	if err != nil {
		if err == promptui.ErrAbort {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// SelectPendingItems allows selecting multiple pending items to mark as done
func (p *Prompter) SelectPendingItems(items []notes.WorkItem) ([]int, error) {
	if len(items) == 0 {
		return nil, nil
	}

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "> {{ .Text | cyan }}",
		Inactive: "  {{ .Text }}",
		Selected: "{{ .Text | green }}",
	}

	var selectedIndices []int

	fmt.Println(RenderInfo("Review pending items:"))
	fmt.Println()

	for i, item := range items {
		completed, err := p.ConfirmCompletion(item)
		if err != nil {
			return selectedIndices, err
		}
		if completed {
			selectedIndices = append(selectedIndices, i)
		}
	}

	// Suppress unused variable warning
	_ = templates

	return selectedIndices, nil
}

// PromptForNewItem asks for a new work item
func (p *Prompter) PromptForNewItem() (string, error) {
	prompt := promptui.Prompt{
		Label: "Enter new work item (leave empty to skip)",
	}

	result, err := prompt.Run()
	if err != nil {
		if err == promptui.ErrInterrupt {
			return "", nil
		}
		return "", err
	}

	return result, nil
}

// PromptForTaskInLoop prompts for a task and returns it with a flag indicating if interrupted
func (p *Prompter) PromptForTaskInLoop(taskNumber int) (string, bool, error) {
	label := PromptStyle.Render(fmt.Sprintf("Task #%d", taskNumber))
	prompt := promptui.Prompt{
		Label: label,
	}

	result, err := prompt.Run()
	if err != nil {
		if err == promptui.ErrInterrupt {
			return "", true, nil // Interrupted
		}
		return "", false, err
	}

	return strings.TrimSpace(result), false, nil
}

// ConfirmAction asks for a yes/no confirmation
func (p *Prompter) ConfirmAction(message string) (bool, error) {
	prompt := promptui.Prompt{
		Label:     message,
		IsConfirm: true,
	}

	_, err := prompt.Run()
	if err != nil {
		if err == promptui.ErrAbort {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// SelectFromList allows selecting an item from a list
func (p *Prompter) SelectFromList(label string, items []string) (int, error) {
	prompt := promptui.Select{
		Label: label,
		Items: items,
	}

	index, _, err := prompt.Run()
	if err != nil {
		return -1, err
	}

	return index, nil
}

// DisplayWorkItems shows a formatted list of work items with modern styling
func (p *Prompter) DisplayWorkItems(pending, completed []notes.WorkItem) {
	// Pending section
	pendingHeader := HeaderStyle.Render("Pending") + " " + RenderBadge(len(pending), PendingBadgeStyle)
	fmt.Println(pendingHeader)

	if len(pending) == 0 {
		fmt.Println(RenderEmptyState("  No pending items — you're all caught up!"))
	} else {
		var pendingItems []string
		for i, item := range pending {
			pendingItems = append(pendingItems, RenderPendingItem(i+1, item.Text))
		}
		content := strings.Join(pendingItems, "\n")
		fmt.Println(PendingCardStyle.Render(content))
	}

	// Completed section
	completedHeader := HeaderStyle.Render("Done") + " " + RenderBadge(len(completed), CompletedBadgeStyle)
	fmt.Println(completedHeader)

	if len(completed) == 0 {
		fmt.Println(RenderEmptyState("  No completed items yet"))
	} else {
		var completedItems []string
		for i, item := range completed {
			completedItems = append(completedItems, RenderCompletedItem(i+1, item.Text))
		}
		content := strings.Join(completedItems, "\n")
		fmt.Println(CompletedCardStyle.Render(content))
	}
}

// DisplayPendingOnly shows only pending work items with modern styling
func (p *Prompter) DisplayPendingOnly(pending []notes.WorkItem) {
	// Pending section header
	pendingHeader := HeaderStyle.Render("Pending") + " " + RenderBadge(len(pending), PendingBadgeStyle)
	fmt.Println(pendingHeader)

	if len(pending) == 0 {
		fmt.Println(RenderEmptyState("  No pending items — you're all caught up!"))
	} else {
		var pendingItems []string
		for i, item := range pending {
			pendingItems = append(pendingItems, RenderPendingItem(i+1, item.Text))
		}
		content := strings.Join(pendingItems, "\n")
		fmt.Println(PendingCardStyle.Render(content))
	}
}

// DisplayMessage shows a message to the user
func (p *Prompter) DisplayMessage(message string) {
	fmt.Println(RenderInfo(message))
}

// DisplayError shows an error message
func (p *Prompter) DisplayError(message string) {
	fmt.Println(RenderError(message))
}

// DisplaySuccess shows a success message
func (p *Prompter) DisplaySuccess(message string) {
	fmt.Println(RenderSuccess(message))
}

// DisplayWarning shows a warning message
func (p *Prompter) DisplayWarning(message string) {
	fmt.Println(RenderWarning(message))
}

// DisplayTitle shows a styled title
func (p *Prompter) DisplayTitle(title string) {
	fmt.Println(RenderTitle(title))
}

// DisplayHeader shows a styled header
func (p *Prompter) DisplayHeader(header string) {
	fmt.Println(RenderHeader(header))
}

// DisplaySummaryBox shows a summary in a styled box
func (p *Prompter) DisplaySummaryBox(title, content string) {
	fmt.Println(RenderSummary(title, content))
}

// DisplayDateHeader shows a styled date header
func (p *Prompter) DisplayDateHeader(date string) {
	header := TitleStyle.Render("📅 " + date)
	fmt.Println(header)
}

// DisplayStats shows task statistics
func (p *Prompter) DisplayStats(pending, completed int) {
	stats := lipgloss.JoinHorizontal(
		lipgloss.Center,
		MutedStyle.Render("Tasks: "),
		RenderBadge(pending, PendingBadgeStyle),
		MutedStyle.Render(" pending  "),
		RenderBadge(completed, CompletedBadgeStyle),
		MutedStyle.Render(" completed"),
	)
	fmt.Println(stats)
	fmt.Println()
}

// SelectTasksToDelete allows selecting tasks to delete from a list
func (p *Prompter) SelectTasksToDelete(items []notes.WorkItem, taskType string) ([]int, error) {
	if len(items) == 0 {
		return nil, nil
	}

	var selectedIndices []int

	fmt.Println(RenderInfo(fmt.Sprintf("Select %s tasks to delete:", taskType)))
	fmt.Println()

	for i, item := range items {
		prompt := promptui.Prompt{
			Label:     fmt.Sprintf("Delete %s task: \"%s\"", taskType, item.Text),
			IsConfirm: true,
		}

		_, err := prompt.Run()
		if err != nil {
			if err == promptui.ErrAbort {
				continue // User said no, skip this item
			}
			return selectedIndices, err
		}
		selectedIndices = append(selectedIndices, i)
	}

	return selectedIndices, nil
}

// SelectWorkplace allows selecting a workplace from the configured list
func (p *Prompter) SelectWorkplace(workplaces []string) (string, error) {
	if len(workplaces) == 0 {
		return "", fmt.Errorf("no workplaces configured")
	}

	// If only one workplace, return it directly
	if len(workplaces) == 1 {
		return workplaces[0], nil
	}

	fmt.Println()
	fmt.Println(RenderInfo("Select workplace"))

	prompt := promptui.Select{
		Label: "Workplace",
		Items: workplaces,
		Size:  10,
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}",
			Active:   "> {{ . | cyan | bold }}",
			Inactive: "  {{ . }}",
			Selected: SuccessStyle.Render(IconSuccess) + " {{ . | green }}",
		},
	}

	_, result, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return result, nil
}

// PromptForWorkplaceName prompts the user to enter a workplace name
func (p *Prompter) PromptForWorkplaceName(label string) (string, error) {
	validate := func(input string) error {
		trimmed := strings.TrimSpace(input)
		if trimmed == "" {
			return fmt.Errorf("workplace name cannot be empty")
		}
		if strings.Contains(trimmed, ",") {
			return fmt.Errorf("workplace name cannot contain commas")
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:    label,
		Validate: validate,
	}

	result, err := prompt.Run()
	if err != nil {
		if err == promptui.ErrInterrupt {
			return "", fmt.Errorf("cancelled")
		}
		return "", err
	}

	return strings.TrimSpace(result), nil
}

// SelectWorkplaceToRename allows selecting a workplace to rename
func (p *Prompter) SelectWorkplaceToRename(workplaces []string) (string, error) {
	if len(workplaces) == 0 {
		return "", fmt.Errorf("no workplaces configured")
	}

	fmt.Println()
	fmt.Println(RenderInfo("Select workplace to rename"))

	prompt := promptui.Select{
		Label: "Workplace",
		Items: workplaces,
		Size:  10,
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}",
			Active:   "> {{ . | cyan | bold }}",
			Inactive: "  {{ . }}",
			Selected: WarningStyle.Render("→") + " {{ . | yellow }}",
		},
	}

	_, result, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return result, nil
}
