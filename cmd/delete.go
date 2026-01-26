package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/sandepten/work-obsidian-noter/internal/notes"
	"github.com/sandepten/work-obsidian-noter/internal/ui"
	"github.com/spf13/cobra"
)

var (
	deleteAll bool
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete tasks from a workplace's worklog",
	Long: `Delete specific tasks from today's worklog for a selected workplace.
By default, you will be prompted to select which tasks to delete.
Use --all flag to delete the entire worklog file.`,
	RunE: runDelete,
}

func init() {
	deleteCmd.Flags().BoolVarP(&deleteAll, "all", "a", false, "Delete the entire worklog file")
	rootCmd.AddCommand(deleteCmd)
}

func runDelete(cmd *cobra.Command, args []string) error {
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
		prompter.DisplayWarning(fmt.Sprintf("No note found for today in %s. Nothing to delete.", selectedWorkplace))
		return nil
	}

	// If --all flag is set, delete the entire file
	if deleteAll {
		return deleteEntireWorklog(todayNote, selectedWorkplace)
	}

	// Otherwise, let user select specific tasks to delete
	return deleteSpecificTasks(todayNote, workplaceWriter, selectedWorkplace)
}

func deleteEntireWorklog(todayNote *notes.Note, selectedWorkplace string) error {
	// Show what will be deleted
	fmt.Println()
	fmt.Println(ui.TitleStyle.Render(fmt.Sprintf("🗑️  Delete Entire Worklog (%s)", selectedWorkplace)))
	fmt.Println(ui.RenderDivider(50))
	fmt.Println()

	fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("File: %s", filepath.Base(todayNote.FilePath))))
	fmt.Println(ui.MutedStyle.Render(fmt.Sprintf("  Pending tasks: %d", len(todayNote.PendingWork))))
	fmt.Println(ui.MutedStyle.Render(fmt.Sprintf("  Completed tasks: %d", len(todayNote.CompletedWork))))
	fmt.Println()

	// Confirm deletion
	confirmed, err := prompter.ConfirmAction(fmt.Sprintf("Are you sure you want to delete today's worklog for %s?", selectedWorkplace))
	if err != nil {
		return fmt.Errorf("error confirming deletion: %w", err)
	}

	if !confirmed {
		fmt.Println()
		fmt.Println(ui.MutedStyle.Render("Deletion cancelled."))
		fmt.Println()
		return nil
	}

	// Delete the file
	if err := os.Remove(todayNote.FilePath); err != nil {
		return fmt.Errorf("error deleting note: %w", err)
	}

	fmt.Println()
	fmt.Println(ui.RenderSuccess(fmt.Sprintf("Deleted today's worklog for %s", selectedWorkplace)))
	fmt.Println()

	return nil
}

func deleteSpecificTasks(todayNote *notes.Note, workplaceWriter *notes.Writer, selectedWorkplace string) error {
	fmt.Println()
	fmt.Println(ui.TitleStyle.Render(fmt.Sprintf("🗑️  Delete Tasks (%s)", selectedWorkplace)))
	fmt.Println(ui.RenderDivider(50))
	fmt.Println()

	// Check if there are any tasks
	if !todayNote.HasPendingWork() && !todayNote.HasCompletedWork() {
		fmt.Println(ui.MutedStyle.Render("No tasks to delete."))
		fmt.Println()
		return nil
	}

	var pendingDeleted, completedDeleted int

	// Delete pending tasks
	if todayNote.HasPendingWork() {
		pendingIndices, err := prompter.SelectTasksToDelete(todayNote.PendingWork, "pending")
		if err != nil {
			return fmt.Errorf("error selecting pending tasks: %w", err)
		}

		// Sort indices in descending order to avoid index shifting
		sort.Sort(sort.Reverse(sort.IntSlice(pendingIndices)))

		for _, idx := range pendingIndices {
			todayNote.RemovePendingItem(idx)
		}
		pendingDeleted = len(pendingIndices)
	}

	fmt.Println()

	// Delete completed tasks
	if todayNote.HasCompletedWork() {
		completedIndices, err := prompter.SelectTasksToDelete(todayNote.CompletedWork, "completed")
		if err != nil {
			return fmt.Errorf("error selecting completed tasks: %w", err)
		}

		// Sort indices in descending order to avoid index shifting
		sort.Sort(sort.Reverse(sort.IntSlice(completedIndices)))

		for _, idx := range completedIndices {
			todayNote.RemoveCompletedItem(idx)
		}
		completedDeleted = len(completedIndices)
	}

	totalDeleted := pendingDeleted + completedDeleted

	if totalDeleted == 0 {
		fmt.Println()
		fmt.Println(ui.MutedStyle.Render("No tasks deleted."))
		fmt.Println()
		return nil
	}

	// Save the updated note
	if err := workplaceWriter.WriteNote(todayNote); err != nil {
		return fmt.Errorf("error saving note: %w", err)
	}

	fmt.Println()
	fmt.Println(ui.RenderDivider(50))
	fmt.Println(ui.RenderSuccess(fmt.Sprintf("Deleted %d task(s) from %s", totalDeleted, selectedWorkplace)))
	if pendingDeleted > 0 {
		fmt.Println(ui.MutedStyle.Render(fmt.Sprintf("  Pending tasks deleted: %d", pendingDeleted)))
	}
	if completedDeleted > 0 {
		fmt.Println(ui.MutedStyle.Render(fmt.Sprintf("  Completed tasks deleted: %d", completedDeleted)))
	}
	fmt.Println()

	// Show remaining tasks
	if todayNote.HasPendingWork() || todayNote.HasCompletedWork() {
		fmt.Println(ui.InfoStyle.Render("Remaining tasks:"))
		prompter.DisplayWorkItems(todayNote.PendingWork, todayNote.CompletedWork)
	}

	return nil
}
