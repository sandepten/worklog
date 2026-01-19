package cmd

import (
	"fmt"
	"path/filepath"
	"sort"
	"time"

	"github.com/sandepten/work-obsidian-noter/internal/notes"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start daily workflow",
	Long: `Start your daily workflow:
1. Review pending items from the most recent previous note
2. Mark items as completed or carry them forward
3. Generate an AI summary of yesterday's completed work
4. Create today's note with the summary`,
	RunE: runStart,
}

func init() {
	rootCmd.AddCommand(startCmd)
}

func runStart(cmd *cobra.Command, args []string) error {
	today := time.Now().Truncate(24 * time.Hour)

	fmt.Printf("Starting daily workflow for %s...\n\n", today.Format("2006-01-02"))

	// Check if today's note already exists
	todayNote, err := parser.FindTodayNote(today)
	if err != nil {
		return fmt.Errorf("error checking for today's note: %w", err)
	}

	// Find the most recent previous note
	previousNote, err := parser.FindMostRecentNote(today)
	if err != nil {
		return fmt.Errorf("error finding previous note: %w", err)
	}

	// Create today's note if it doesn't exist
	if todayNote == nil {
		todayNote = writer.CreateTodayNote(today)
		fmt.Printf("Creating new note: %s\n\n", filepath.Base(todayNote.FilePath))
	} else {
		fmt.Printf("Today's note already exists: %s\n\n", filepath.Base(todayNote.FilePath))
	}

	// Process previous note if it exists
	if previousNote != nil {
		fmt.Printf("Found previous note: %s (Date: %s)\n\n", filepath.Base(previousNote.FilePath), previousNote.Date.Format("2006-01-02"))

		// Review pending items from previous note
		if previousNote.HasPendingWork() {
			fmt.Println("Reviewing pending items from previous note...\n")

			completedIndices, err := prompter.SelectPendingItems(previousNote.PendingWork)
			if err != nil {
				return fmt.Errorf("error reviewing pending items: %w", err)
			}

			// Sort indices in descending order to avoid index shifting during removal
			sort.Sort(sort.Reverse(sort.IntSlice(completedIndices)))

			// Process completed items - move to previous note's completed section
			for _, idx := range completedIndices {
				item := previousNote.PendingWork[idx]
				previousNote.CompletedWork = append(previousNote.CompletedWork, notes.WorkItem{
					Text:      item.Text,
					Completed: true,
				})
			}

			// Remaining pending items go to today's note
			remainingPending := []notes.WorkItem{}
			completedSet := make(map[int]bool)
			for _, idx := range completedIndices {
				completedSet[idx] = true
			}

			for i, item := range previousNote.PendingWork {
				if !completedSet[i] {
					remainingPending = append(remainingPending, item)
					// Add to today's pending
					todayNote.AddPendingItem(item.Text)
				}
			}

			// Update previous note - clear pending (items either completed or moved)
			previousNote.PendingWork = []notes.WorkItem{}
		}

		// Generate summary if there's completed work
		if previousNote.HasCompletedWork() {
			fmt.Println("\nGenerating AI summary of completed work...")

			// Test connection first
			if err := aiClient.TestConnection(); err != nil {
				fmt.Printf("Warning: Could not connect to OpenCode server: %v\n", err)
				fmt.Println("Skipping AI summary generation.")
			} else {
				summary, err := aiClient.SummarizeWorkItems(previousNote.CompletedWork)
				if err != nil {
					fmt.Printf("Warning: Could not generate summary: %v\n", err)
				} else {
					fmt.Printf("\nSummary: %s\n\n", summary)

					// Update both notes with the summary
					previousNote.Summary = summary
					todayNote.YesterdaySummary = summary
				}
			}
		}

		// Save the updated previous note
		if err := writer.WriteNote(previousNote); err != nil {
			return fmt.Errorf("error saving previous note: %w", err)
		}
		fmt.Printf("Updated previous note: %s\n", filepath.Base(previousNote.FilePath))
	} else {
		fmt.Println("No previous notes found. Starting fresh!")
	}

	// Save today's note
	if err := writer.WriteNote(todayNote); err != nil {
		return fmt.Errorf("error saving today's note: %w", err)
	}
	fmt.Printf("Saved today's note: %s\n", filepath.Base(todayNote.FilePath))

	// Show current state
	fmt.Println("\n--- Today's Note ---")
	prompter.DisplayWorkItems(todayNote.PendingWork, todayNote.CompletedWork)

	fmt.Println("Daily workflow complete! Use 'worklog add \"task\"' to add new items.")

	return nil
}
