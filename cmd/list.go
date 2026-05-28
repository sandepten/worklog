package cmd

import (
	"fmt"
	"time"

	"github.com/sandepten/work-obsidian-noter/internal/ui"
	"github.com/sandepten/work-obsidian-noter/internal/util"
	"github.com/spf13/cobra"
)

var (
	pendingOnly bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List today's work items",
	Long:  `Display all pending and completed work items from today's note.`,
	RunE:  runList,
}

func init() {
	listCmd.Flags().BoolVarP(&pendingOnly, "pending", "p", false, "Show only pending tasks")
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
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

	// Display date header with stats inline
	dateStr := today.Format("Mon, Jan 2")
	statsStr := fmt.Sprintf("%d pending · %d done", len(todayNote.PendingWork), len(todayNote.CompletedWork))
	fmt.Printf("%s  %s\n", ui.TitleStyle.Render("📅 "+dateStr), ui.MutedStyle.Render(statsStr))

	// Show yesterday's summary only if NOT using --pending flag
	if !pendingOnly && todayNote.YesterdaySummary != "" {
		fmt.Println(ui.RenderSummary("Yesterday", todayNote.YesterdaySummary))
	}

	// Display based on flag
	if pendingOnly {
		prompter.DisplayPendingOnly(todayNote.PendingWork)
	} else {
		prompter.DisplayWorkItems(todayNote.PendingWork, todayNote.CompletedWork)
	}

	// Show tip at the end
	fmt.Println(ui.MutedStyle.Render("💡 Use 'worklog add \"task\"' to add items"))

	return nil
}
