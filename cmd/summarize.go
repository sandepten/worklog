package cmd

import (
	"fmt"
	"time"

	"github.com/sandepten/work-obsidian-noter/internal/notes"
	"github.com/sandepten/work-obsidian-noter/internal/ui"
	"github.com/spf13/cobra"
)

var summarizeCmd = &cobra.Command{
	Use:   "summarize",
	Short: "Get AI summary of today's completed work",
	Long:  `Generate and display an AI-powered summary of today's completed work items. You will be prompted to select a workplace if multiple are configured.`,
	RunE:  runSummarize,
}

func init() {
	rootCmd.AddCommand(summarizeCmd)
}

func runSummarize(cmd *cobra.Command, args []string) error {
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

	if !todayNote.HasCompletedWork() {
		fmt.Println()
		fmt.Println(ui.MutedStyle.Render(fmt.Sprintf("No completed work items to summarize in %s.", selectedWorkplace)))
		fmt.Println(ui.MutedStyle.Render("Use 'worklog done' to mark items as completed first."))
		fmt.Println()
		return nil
	}

	fmt.Println()
	fmt.Println(ui.TitleStyle.Render(fmt.Sprintf("📊 Work Summary (%s)", selectedWorkplace)))
	fmt.Println(ui.MutedStyle.Render(today.Format("Monday, January 2, 2006")))
	fmt.Println(ui.RenderDivider(50))
	fmt.Println()

	// Display completed work
	fmt.Println(ui.HeaderStyle.Render("Completed Work"))
	for i, item := range todayNote.CompletedWork {
		fmt.Println(ui.RenderCompletedItem(i+1, item.Text))
	}
	fmt.Println()

	// Generate AI summary
	fmt.Println(ui.InfoStyle.Render("🤖 Generating AI summary..."))
	fmt.Println()

	// Test connection first
	if err := aiClient.TestConnection(); err != nil {
		return fmt.Errorf("could not connect to OpenCode server: %w", err)
	}

	summary, err := aiClient.SummarizeWorkItems(todayNote.CompletedWork)
	if err != nil {
		return fmt.Errorf("could not generate summary: %w", err)
	}

	prompter.DisplaySummaryBox("AI-Generated Summary", summary)

	return nil
}
