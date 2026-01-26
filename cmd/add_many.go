package cmd

import (
	"fmt"
	"time"

	"github.com/sandepten/work-obsidian-noter/internal/notes"
	"github.com/sandepten/work-obsidian-noter/internal/ui"
	"github.com/spf13/cobra"
)

var addManyCmd = &cobra.Command{
	Use:   "add-many",
	Short: "Add multiple work items interactively",
	Long: `Add multiple pending work items in a loop.
Press Enter after each task to add it.
Press Ctrl+C when done to exit and see a summary.
You will be prompted to select a workplace if multiple are configured.`,
	RunE: runAddMany,
}

func init() {
	rootCmd.AddCommand(addManyCmd)
}

func runAddMany(cmd *cobra.Command, args []string) error {
	today := time.Now().Truncate(24 * time.Hour)

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
		prompter.DisplayMessage(fmt.Sprintf("Creating today's note for %s...", selectedWorkplace))
	}

	// Display header
	fmt.Println()
	fmt.Println(ui.TitleStyle.Render(fmt.Sprintf("📝 Add Multiple Tasks (%s)", selectedWorkplace)))
	fmt.Println(ui.MutedStyle.Render("Enter each task and press Enter. Press Ctrl+C when done."))
	fmt.Println(ui.RenderDivider(50))
	fmt.Println()

	var addedTasks []string
	taskNumber := 1

	for {
		task, interrupted, err := prompter.PromptForTaskInLoop(taskNumber)
		if err != nil {
			return fmt.Errorf("error prompting for task: %w", err)
		}

		if interrupted {
			// User pressed Ctrl+C
			break
		}

		// Skip empty input
		if task == "" {
			fmt.Println(ui.MutedStyle.Render("  (empty input skipped)"))
			continue
		}

		// Add the task
		todayNote.AddPendingItem(task)
		addedTasks = append(addedTasks, task)

		// Show confirmation
		fmt.Println(ui.SuccessStyle.Render(fmt.Sprintf("  %s Added: %s", ui.IconSuccess, task)))
		taskNumber++
	}

	fmt.Println()
	fmt.Println(ui.RenderDivider(50))

	// Save the note if any tasks were added
	if len(addedTasks) > 0 {
		if err := workplaceWriter.WriteNote(todayNote); err != nil {
			return fmt.Errorf("error saving note: %w", err)
		}

		// Show summary
		fmt.Println()
		summary := fmt.Sprintf("Added %d task(s) to %s worklog", len(addedTasks), selectedWorkplace)
		fmt.Println(ui.RenderSuccess(summary))
		fmt.Println()

		// List added tasks
		fmt.Println(ui.InfoStyle.Render("Tasks added:"))
		for i, task := range addedTasks {
			fmt.Println(ui.RenderPendingItem(i+1, task))
		}
	} else {
		fmt.Println()
		fmt.Println(ui.MutedStyle.Render("No tasks added."))
	}

	fmt.Println()
	return nil
}
