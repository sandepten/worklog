package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var doneCmd = &cobra.Command{
	Use:   "done",
	Short: "Mark pending items as completed",
	Long:  `Interactively mark pending items as completed in today's note.`,
	RunE:  runDone,
}

func init() {
	rootCmd.AddCommand(doneCmd)
}

func runDone(cmd *cobra.Command, args []string) error {
	today := time.Now().Truncate(24 * time.Hour)

	// Get today's note
	todayNote, err := parser.FindTodayNote(today)
	if err != nil {
		return fmt.Errorf("error finding today's note: %w", err)
	}

	if todayNote == nil {
		prompter.DisplayError("No note found for today. Use 'worklog start' to create one.")
		return nil
	}

	if !todayNote.HasPendingWork() {
		prompter.DisplayMessage("No pending items to mark as done.")
		return nil
	}

	fmt.Println("Select items to mark as completed:")

	completedIndices, err := prompter.SelectPendingItems(todayNote.PendingWork)
	if err != nil {
		return fmt.Errorf("error selecting items: %w", err)
	}

	if len(completedIndices) == 0 {
		prompter.DisplayMessage("No items marked as completed.")
		return nil
	}

	// Mark items as completed (process in reverse order to maintain indices)
	for i := len(completedIndices) - 1; i >= 0; i-- {
		idx := completedIndices[i]
		todayNote.MarkItemCompleted(idx)
	}

	// Save the note
	if err := writer.WriteNote(todayNote); err != nil {
		return fmt.Errorf("error saving note: %w", err)
	}

	prompter.DisplaySuccess(fmt.Sprintf("Marked %d item(s) as completed.", len(completedIndices)))

	// Show updated state
	prompter.DisplayWorkItems(todayNote.PendingWork, todayNote.CompletedWork)

	return nil
}
