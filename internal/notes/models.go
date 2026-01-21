package notes

import (
	"time"
)

// WorkItem represents a single work item (pending or completed)
type WorkItem struct {
	Text      string
	Completed bool
}

// Note represents a daily work note
type Note struct {
	// Frontmatter fields
	ID      string
	Aliases []string
	Tags    []string
	Date    time.Time

	// Content fields
	Title            string
	Summary          string
	YesterdaySummary string
	PendingWork      []WorkItem
	CompletedWork    []WorkItem

	// File info
	FilePath string
}

// NewNote creates a new note for the given date and workplace
func NewNote(date time.Time, workplaceName string) *Note {
	return &Note{
		ID:               generateID(date, workplaceName),
		Aliases:          []string{},
		Tags:             []string{toLowerCase(workplaceName), "job"},
		Date:             date,
		Title:            date.Format("2006-01-02"),
		Summary:          "",
		YesterdaySummary: "",
		PendingWork:      []WorkItem{},
		CompletedWork:    []WorkItem{},
	}
}

// generateID creates the note ID in format: WorkplaceName-D-Mon-YYYY
func generateID(date time.Time, workplaceName string) string {
	return workplaceName + "-" + date.Format("2-Jan-2006")
}

// ToLowerCase converts a string to lowercase (exported for use by other packages)
func ToLowerCase(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			result[i] = c + 32
		} else {
			result[i] = c
		}
	}
	return string(result)
}

// toLowerCase is an alias for internal use
func toLowerCase(s string) string {
	return ToLowerCase(s)
}

// GenerateFilename creates the filename for a note: YYYY-MM-DD-WorkplaceName.md
func GenerateFilename(date time.Time, workplaceName string) string {
	return date.Format("2006-01-02") + "-" + workplaceName + ".md"
}

// HasPendingWork returns true if the note has any pending work items
func (n *Note) HasPendingWork() bool {
	return len(n.PendingWork) > 0
}

// HasCompletedWork returns true if the note has any completed work items
func (n *Note) HasCompletedWork() bool {
	return len(n.CompletedWork) > 0
}

// AddPendingItem adds a new pending work item
func (n *Note) AddPendingItem(text string) {
	n.PendingWork = append(n.PendingWork, WorkItem{Text: text, Completed: false})
}

// AddCompletedItem adds a new completed work item
func (n *Note) AddCompletedItem(text string) {
	n.CompletedWork = append(n.CompletedWork, WorkItem{Text: text, Completed: true})
}

// MarkItemCompleted moves a pending item to completed
func (n *Note) MarkItemCompleted(index int) {
	if index >= 0 && index < len(n.PendingWork) {
		item := n.PendingWork[index]
		item.Completed = true
		n.CompletedWork = append(n.CompletedWork, item)
		// Remove from pending
		n.PendingWork = append(n.PendingWork[:index], n.PendingWork[index+1:]...)
	}
}

// RemovePendingItem removes a pending item at the given index
func (n *Note) RemovePendingItem(index int) {
	if index >= 0 && index < len(n.PendingWork) {
		n.PendingWork = append(n.PendingWork[:index], n.PendingWork[index+1:]...)
	}
}

// RemoveCompletedItem removes a completed item at the given index
func (n *Note) RemoveCompletedItem(index int) {
	if index >= 0 && index < len(n.CompletedWork) {
		n.CompletedWork = append(n.CompletedWork[:index], n.CompletedWork[index+1:]...)
	}
}
