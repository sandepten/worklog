package cmd

import (
	"fmt"
	"time"

	"github.com/sandepten/work-obsidian-noter/internal/notes"
	"github.com/sandepten/work-obsidian-noter/internal/ui"
	"github.com/spf13/cobra"
)

var doneCmd = &cobra.Command{
	Use:   "done",
	Short: "Mark pending items as completed",
	Long:  `Interactively mark pending items as completed in today's note. You will be prompted to select a workplace if multiple are configured.`,
	RunE:  runDone,
}

func init() {
	rootCmd.AddCommand(doneCmd)
}

func runDone(cmd *cobra.Command, args []string) error {
	today := time.Now().Truncate(24 * time.Hour)

	// Ask which workplace
	selectedWorkplace, err := prompter.SelectWorkplace(cfg.Workplaces)
	if err != nil {
		return fmt.Errorf("error selecting workplace: %w", err)
	}

	// Create parser and writer for the selected workplace
	workplaceParser := notes.NewParser(cfg.WorkNotesLocation, selectedWorkplace)
	workplaceWriter := notes.NewWriter(cfg.WorkNotesLocation, selectedWorkplace)

	// Get today's note
	todayNote, err := workplaceParser.FindTodayNote(today)
	if err != nil {
		return fmt.Errorf("error finding today's note: %w", err)
	}

	if todayNote == nil {
		prompter.DisplayWarning(fmt.Sprintf("No note found for today in %s. Use 'worklog start' to create one.", selectedWorkplace))
		return nil
	}

	if !todayNote.HasPendingWork() {
		fmt.Println()
		fmt.Println(ui.RenderSuccess(fmt.Sprintf("No pending items in %s — you're all caught up! 🎉", selectedWorkplace)))
		fmt.Println()
		return nil
	}

	fmt.Println()
	fmt.Println(ui.TitleStyle.Render(fmt.Sprintf("✓ Mark Tasks as Done (%s)", selectedWorkplace)))
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
	if err := workplaceWriter.WriteNote(todayNote); err != nil {
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
