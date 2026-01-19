package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List today's work items",
	Long:  `Display all pending and completed work items from today's note.`,
	RunE:  runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	today := time.Now().Truncate(24 * time.Hour)

	// Get today's note
	todayNote, err := parser.FindTodayNote(today)
	if err != nil {
		return fmt.Errorf("error finding today's note: %w", err)
	}

	if todayNote == nil {
		prompter.DisplayMessage("No note found for today. Use 'worklog start' to create one.")
		return nil
	}

	fmt.Printf("Work items for %s\n", today.Format("2006-01-02"))

	if todayNote.YesterdaySummary != "" {
		fmt.Printf("\nYesterday's Summary: %s\n", todayNote.YesterdaySummary)
	}

	prompter.DisplayWorkItems(todayNote.PendingWork, todayNote.CompletedWork)

	return nil
}
