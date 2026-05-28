package cmd

import (
	"fmt"
	"time"

	"github.com/sandepten/work-obsidian-noter/internal/ui"
	"github.com/sandepten/work-obsidian-noter/internal/util"
	"github.com/spf13/cobra"
)

var summarizeCmd = &cobra.Command{
	Use:   "summarize",
	Short: "Get AI summary of today's completed work",
	Long:  `Generate and display an AI-powered summary of today's completed work items.`,
	RunE:  runSummarize,
}

func init() {
	rootCmd.AddCommand(summarizeCmd)
}

func runSummarize(cmd *cobra.Command, args []string) error {
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

	if !todayNote.HasCompletedWork() {
		fmt.Println()
		fmt.Println(ui.MutedStyle.Render("No completed work items to summarize."))
		fmt.Println(ui.MutedStyle.Render("Use 'worklog done' to mark items as completed first."))
		fmt.Println()
		return nil
	}

	fmt.Println()
	fmt.Println(ui.TitleStyle.Render("📊 Work Summary"))
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
