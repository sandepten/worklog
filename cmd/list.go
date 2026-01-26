package cmd

import (
	"fmt"
	"time"

	"github.com/sandepten/work-obsidian-noter/internal/notes"
	"github.com/sandepten/work-obsidian-noter/internal/ui"
	"github.com/spf13/cobra"
)

var (
	pendingOnly bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List today's work items",
	Long:  `Display all pending and completed work items from today's note. You will be prompted to select a workplace if multiple are configured.`,
	RunE:  runList,
}

func init() {
	listCmd.Flags().BoolVarP(&pendingOnly, "pending", "p", false, "Show only pending tasks")
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	today := time.Now().Truncate(24 * time.Hour)

	// Ask which workplace
	selectedWorkplace, err := prompter.SelectWorkplace(cfg.Workplaces)
	if err != nil {
		return fmt.Errorf("error selecting workplace: %w", err)
	}

	// Create parser for the selected workplace
	workplaceParser := notes.NewParser(cfg.WorkNotesLocation, selectedWorkplace)

	// Get today's note
	todayNote, err := workplaceParser.FindTodayNote(today)
	if err != nil {
		return fmt.Errorf("error finding today's note: %w", err)
	}

	if todayNote == nil {
		prompter.DisplayWarning(fmt.Sprintf("No note found for today in %s. Use 'worklog start' to create one.", selectedWorkplace))
		return nil
	}

	// Display date header with stats inline
	dateStr := today.Format("Mon, Jan 2")
	statsStr := fmt.Sprintf("%d pending · %d done", len(todayNote.PendingWork), len(todayNote.CompletedWork))
	fmt.Printf("%s  %s  %s\n", ui.TitleStyle.Render("📅 "+dateStr), ui.MutedStyle.Render("•"), ui.InfoStyle.Render(selectedWorkplace))
	fmt.Println(ui.MutedStyle.Render(statsStr))

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
