package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/sandepten/work-obsidian-noter/internal/ui"
	"github.com/sandepten/work-obsidian-noter/internal/util"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add [task description]",
	Short: "Add a new pending work item",
	Long:  `Add a new pending work item to today's note.`,
	Args:  cobra.MinimumNArgs(1),
	RunE:  runAdd,
}

func init() {
	rootCmd.AddCommand(addCmd)
}

func runAdd(cmd *cobra.Command, args []string) error {
	today := util.TruncateToLocalDay(time.Now())
	taskText := strings.Join(args, " ")

	// Get or create today's note
	todayNote, err := parser.FindTodayNote(today)
	if err != nil {
		return fmt.Errorf("error finding today's note: %w", err)
	}

	if todayNote == nil {
		todayNote = writer.CreateTodayNote(today)
		fmt.Println(ui.InfoStyle.Render("Creating today's note..."))
	}

	// Add the new item
	todayNote.AddPendingItem(taskText)

	// Save the note
	if err := writer.WriteNote(todayNote); err != nil {
		return fmt.Errorf("error saving note: %w", err)
	}

	fmt.Println()
	fmt.Println(ui.RenderSuccess("Task added successfully!"))
	fmt.Println(ui.RenderPendingItem(len(todayNote.PendingWork), taskText))
	fmt.Println()
	fmt.Println(ui.MutedStyle.Render(fmt.Sprintf("  📋 You now have %d pending task(s)", len(todayNote.PendingWork))))
	fmt.Println()

	return nil
}
