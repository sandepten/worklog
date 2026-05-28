package cmd

import (
	"fmt"
	"time"

	"github.com/sandepten/work-obsidian-noter/internal/ui"
	"github.com/sandepten/work-obsidian-noter/internal/util"
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
	today := util.TruncateToLocalDay(time.Now())

	// Get today's note
	todayNote, err := parser.FindTodayNote(today)
	if err != nil {
		return fmt.Errorf("error finding today's note: %w", err)
	}

	if todayNote == nil {
		prompter.DisplayWarning("No note found for today. Use 'worklog start' to create one.")
		return nil
	}

	if !todayNote.HasPendingWork() {
		fmt.Println()
		fmt.Println(ui.RenderSuccess("No pending items — you're all caught up! 🎉"))
		fmt.Println()
		return nil
	}

	fmt.Println()
	fmt.Println(ui.TitleStyle.Render("✓ Mark Tasks as Done"))
	fmt.Println(ui.MutedStyle.Render("Select which tasks you've completed"))
	fmt.Println(ui.RenderDivider(50))
	fmt.Println()

	completedIndices, err := prompter.SelectPendingItems(todayNote.PendingWork)
	if err != nil {
		return fmt.Errorf("error selecting items: %w", err)
	}

	if len(completedIndices) == 0 {
		fmt.Println()
		fmt.Println(ui.MutedStyle.Render("No items marked as completed."))
		fmt.Println()
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

	fmt.Println()
	fmt.Println(ui.RenderDivider(50))
	fmt.Println(ui.RenderSuccess(fmt.Sprintf("Marked %d item(s) as completed!", len(completedIndices))))
	fmt.Println()

	// Show updated state
	prompter.DisplayWorkItems(todayNote.PendingWork, todayNote.CompletedWork)

	return nil
}
