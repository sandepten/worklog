package cmd

import (
	"fmt"
	"path/filepath"
	"sort"
	"time"

	"github.com/sandepten/work-obsidian-noter/internal/notes"
	"github.com/sandepten/work-obsidian-noter/internal/ui"
	"github.com/sandepten/work-obsidian-noter/internal/util"
	"github.com/spf13/cobra"
)

var reviewCmd = &cobra.Command{
	Use:   "review",
	Short: "Review pending items from previous notes",
	Long: `Manually review and process pending items from previous notes
without creating a new note or generating summaries.`,
	RunE: runReview,
}

func init() {
	rootCmd.AddCommand(reviewCmd)
}

func runReview(cmd *cobra.Command, args []string) error {
	today := util.TruncateToLocalDay(time.Now())

	// Find the most recent previous note
	previousNote, err := parser.FindMostRecentNote(today)
	if err != nil {
		return fmt.Errorf("error finding previous note: %w", err)
	}

	if previousNote == nil {
		prompter.DisplayMessage("No previous notes found.")
		return nil
	}

	fmt.Println()
	fmt.Println(ui.TitleStyle.Render("📝 Review Previous Note"))
	fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("📄 %s (%s)", filepath.Base(previousNote.FilePath), previousNote.Date.Format("January 2, 2006"))))
	fmt.Println(ui.RenderDivider(50))
	fmt.Println()

	if !previousNote.HasPendingWork() {
		fmt.Println(ui.RenderSuccess("No pending items to review — all caught up! 🎉"))
		fmt.Println()
		prompter.DisplayWorkItems(previousNote.PendingWork, previousNote.CompletedWork)
		return nil
	}

	fmt.Println(ui.HeaderStyle.Render("Review Pending Items"))
	fmt.Println(ui.MutedStyle.Render("Mark items you've completed"))
	fmt.Println()

	completedIndices, err := prompter.SelectPendingItems(previousNote.PendingWork)
	if err != nil {
		return fmt.Errorf("error reviewing items: %w", err)
	}

	if len(completedIndices) == 0 {
		fmt.Println()
		fmt.Println(ui.MutedStyle.Render("No items marked as completed."))
		fmt.Println()
		return nil
	}

	// Sort indices in descending order
	sort.Sort(sort.Reverse(sort.IntSlice(completedIndices)))

	// Mark items as completed
	for _, idx := range completedIndices {
		item := previousNote.PendingWork[idx]
		previousNote.CompletedWork = append(previousNote.CompletedWork, notes.WorkItem{
			Text:      item.Text,
			Completed: true,
		})
		// Remove from pending
		previousNote.PendingWork = append(previousNote.PendingWork[:idx], previousNote.PendingWork[idx+1:]...)
	}

	// Save the note
	if err := writer.WriteNote(previousNote); err != nil {
		return fmt.Errorf("error saving note: %w", err)
	}

	fmt.Println()
	fmt.Println(ui.RenderDivider(50))
	fmt.Println(ui.RenderSuccess(fmt.Sprintf("Marked %d item(s) as completed!", len(completedIndices))))
	fmt.Println()

	// Show updated state
	prompter.DisplayWorkItems(previousNote.PendingWork, previousNote.CompletedWork)

	return nil
}
