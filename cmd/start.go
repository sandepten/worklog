package cmd

import (
	"fmt"
	"path/filepath"
	"sort"
	"time"

	"github.com/sandepten/work-obsidian-noter/internal/notes"
	"github.com/sandepten/work-obsidian-noter/internal/ui"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start daily workflow",
	Long: `Start your daily workflow:
1. Review pending items from the most recent previous note
2. Mark items as completed or carry them forward
3. Generate an AI summary of yesterday's completed work
4. Create today's note with the summary
You will be prompted to select a workplace if multiple are configured.`,
	RunE: runStart,
}

func init() {
	rootCmd.AddCommand(startCmd)
}

func runStart(cmd *cobra.Command, args []string) error {
	today := time.Now().Truncate(24 * time.Hour)

	// Ask which workplace
	selectedWorkplace, err := prompter.SelectWorkplace(cfg.Workplaces)
	if err != nil {
		return fmt.Errorf("error selecting workplace: %w", err)
	}

	// Create parser and writer for the selected workplace
	workplaceParser := notes.NewParser(cfg.WorkNotesLocation, selectedWorkplace)
	workplaceWriter := notes.NewWriter(cfg.WorkNotesLocation, selectedWorkplace)

	fmt.Println()
	fmt.Println(ui.TitleStyle.Render(fmt.Sprintf("🚀 Daily Workflow (%s)", selectedWorkplace)))
	fmt.Println(ui.MutedStyle.Render(today.Format("Monday, January 2, 2006")))
	fmt.Println(ui.RenderDivider(50))
	fmt.Println()

	// Check if today's note already exists
	todayNote, err := workplaceParser.FindTodayNote(today)
	if err != nil {
		return fmt.Errorf("error checking for today's note: %w", err)
	}

	// Find the most recent previous note
	previousNote, err := workplaceParser.FindMostRecentNote(today)
	if err != nil {
		return fmt.Errorf("error finding previous note: %w", err)
	}

	// Create today's note if it doesn't exist
	if todayNote == nil {
		todayNote = workplaceWriter.CreateTodayNote(today)
		fmt.Println(ui.RenderSuccess(fmt.Sprintf("Created new note: %s", filepath.Base(todayNote.FilePath))))
	} else {
		fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("ℹ Today's note already exists: %s", filepath.Base(todayNote.FilePath))))
	}
	fmt.Println()

	// Process previous note if it exists
	if previousNote != nil {
		fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("📄 Found previous note: %s", filepath.Base(previousNote.FilePath))))
		fmt.Println()

		// Review pending items from previous note
		if previousNote.HasPendingWork() {
			fmt.Println(ui.HeaderStyle.Render("Review Pending Items"))
			fmt.Println(ui.MutedStyle.Render("Mark items you completed since last session"))
			fmt.Println()

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
			completedSet := make(map[int]bool)
			for _, idx := range completedIndices {
				completedSet[idx] = true
			}

			for i, item := range previousNote.PendingWork {
				if !completedSet[i] {
					// Add to today's pending
					todayNote.AddPendingItem(item.Text)
				}
			}

			// Update previous note - clear pending (items either completed or moved)
			previousNote.PendingWork = []notes.WorkItem{}

			if len(completedIndices) > 0 {
				fmt.Println()
				fmt.Println(ui.RenderSuccess(fmt.Sprintf("Marked %d item(s) as completed", len(completedIndices))))
			}
		}

		// Generate summary if there's completed work
		if previousNote.HasCompletedWork() {
			fmt.Println()
			fmt.Println(ui.HeaderStyle.Render("AI Summary"))
			fmt.Println(ui.MutedStyle.Render("Generating summary of completed work..."))

			// Test connection first
			if err := aiClient.TestConnection(); err != nil {
				fmt.Println(ui.RenderWarning(fmt.Sprintf("Could not connect to OpenCode server: %v", err)))
				fmt.Println(ui.MutedStyle.Render("Skipping AI summary generation."))
			} else {
				summary, err := aiClient.SummarizeWorkItems(previousNote.CompletedWork)
				if err != nil {
					fmt.Println(ui.RenderWarning(fmt.Sprintf("Could not generate summary: %v", err)))
				} else {
					fmt.Println()
					prompter.DisplaySummaryBox("Summary", summary)

					// Update both notes with the summary
					previousNote.Summary = summary
					todayNote.YesterdaySummary = summary
				}
			}
		}

		// Save the updated previous note
		if err := workplaceWriter.WriteNote(previousNote); err != nil {
			return fmt.Errorf("error saving previous note: %w", err)
		}
		fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("ℹ Updated: %s", filepath.Base(previousNote.FilePath))))
	} else {
		fmt.Println(ui.MutedStyle.Render("No previous notes found. Starting fresh!"))
	}

	// Save today's note
	if err := workplaceWriter.WriteNote(todayNote); err != nil {
		return fmt.Errorf("error saving today's note: %w", err)
	}

	fmt.Println()
	fmt.Println(ui.RenderDivider(50))
	fmt.Println()
	fmt.Println(ui.TitleStyle.Render("📋 Today's Note"))

	// Show current state
	prompter.DisplayWorkItems(todayNote.PendingWork, todayNote.CompletedWork)

	fmt.Println(ui.RenderSuccess("Daily workflow complete!"))
	fmt.Println(ui.MutedStyle.Render("Use 'worklog add \"task\"' to add new items"))
	fmt.Println()

	return nil
}
