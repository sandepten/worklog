package cmd

import (
	"fmt"
	"time"

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
	today := time.Now().Truncate(24 * time.Hour)

	// Get today's note
	todayNote, err := parser.FindTodayNote(today)
	if err != nil {
		return fmt.Errorf("error finding today's note: %w", err)
	}

	if todayNote == nil {
		prompter.DisplayError("No note found for today. Use 'noter start' to create one.")
		return nil
	}

	if !todayNote.HasCompletedWork() {
		prompter.DisplayMessage("No completed work items to summarize.")
		return nil
	}

	fmt.Printf("Completed work for %s:\n", today.Format("2006-01-02"))
	for i, item := range todayNote.CompletedWork {
		fmt.Printf("  %d. %s\n", i+1, item.Text)
	}

	fmt.Println("\nGenerating AI summary...")

	// Test connection first
	if err := aiClient.TestConnection(); err != nil {
		return fmt.Errorf("could not connect to OpenCode server: %w", err)
	}

	summary, err := aiClient.SummarizeWorkItems(todayNote.CompletedWork)
	if err != nil {
		return fmt.Errorf("could not generate summary: %w", err)
	}

	fmt.Println("\n--- Summary ---")
	fmt.Println(summary)

	return nil
}
