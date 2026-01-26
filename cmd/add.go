package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/sandepten/work-obsidian-noter/internal/notes"
	"github.com/sandepten/work-obsidian-noter/internal/ui"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add [task description]",
	Short: "Add a new pending work item",
	Long:  `Add a new pending work item to today's note. You will be prompted to select a workplace if multiple are configured.`,
	Args:  cobra.MinimumNArgs(1),
	RunE:  runAdd,
}

func init() {
	rootCmd.AddCommand(addCmd)
}

func runAdd(cmd *cobra.Command, args []string) error {
	today := time.Now().Truncate(24 * time.Hour)
	taskText := strings.Join(args, " ")

	// Ask which workplace this task belongs to
	selectedWorkplace, err := prompter.SelectWorkplace(cfg.Workplaces)
	if err != nil {
		return fmt.Errorf("error selecting workplace: %w", err)
	}

	// Create parser and writer for the selected workplace
	workplaceParser := notes.NewParser(cfg.WorkNotesLocation, selectedWorkplace)
	workplaceWriter := notes.NewWriter(cfg.WorkNotesLocation, selectedWorkplace)

	// Get or create today's note for the selected workplace
	todayNote, err := workplaceParser.FindTodayNote(today)
	if err != nil {
		return fmt.Errorf("error finding today's note: %w", err)
	}

	if todayNote == nil {
		todayNote = workplaceWriter.CreateTodayNote(today)
		fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("Creating today's note for %s...", selectedWorkplace)))
	}

	// Add the new item
	todayNote.AddPendingItem(taskText)

	// Save the note
	if err := workplaceWriter.WriteNote(todayNote); err != nil {
		return fmt.Errorf("error saving note: %w", err)
	}

	fmt.Println()
	fmt.Println(ui.RenderSuccess(fmt.Sprintf("Task added to %s!", selectedWorkplace)))
	fmt.Println(ui.RenderPendingItem(len(todayNote.PendingWork), taskText))
	fmt.Println()
	fmt.Println(ui.MutedStyle.Render(fmt.Sprintf("  📋 You now have %d pending task(s) in %s", len(todayNote.PendingWork), selectedWorkplace)))
	fmt.Println()

	return nil
}
