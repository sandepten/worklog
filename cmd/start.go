package cmd

import (
	"fmt"
	"os"
	"os/exec"
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
4. Create today's note with the summary`,
	RunE: runStart,
}

func init() {
	rootCmd.AddCommand(startCmd)
}

func runStart(cmd *cobra.Command, args []string) error {
	today := time.Now().Truncate(24 * time.Hour)

	fmt.Println()
	fmt.Println(ui.TitleStyle.Render("🚀 Daily Workflow"))
	fmt.Println(ui.MutedStyle.Render(today.Format("Monday, January 2, 2006")))
	fmt.Println(ui.RenderDivider(50))
	fmt.Println()

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
		if err := writer.WriteNote(previousNote); err != nil {
			return fmt.Errorf("error saving previous note: %w", err)
		}
		fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("ℹ Updated: %s", filepath.Base(previousNote.FilePath))))
	} else {
		fmt.Println(ui.MutedStyle.Render("No previous notes found. Starting fresh!"))
	}

	// Save today's note
	if err := writer.WriteNote(todayNote); err != nil {
		return fmt.Errorf("error saving today's note: %w", err)
	}

	fmt.Println()
	fmt.Println(ui.RenderDivider(50))
	fmt.Println()
	fmt.Println(ui.TitleStyle.Render("📋 Today's Note"))

	// Show current state
	prompter.DisplayWorkItems(todayNote.PendingWork, todayNote.CompletedWork)

	fmt.Println()
	fmt.Println(ui.RenderDivider(50))
	fmt.Println()

	// Commit all files in the obsidian folder
	if err := commitObsidianFiles(); err != nil {
		fmt.Println(ui.RenderWarning(fmt.Sprintf("Could not commit files: %v", err)))
		fmt.Println(ui.MutedStyle.Render("Continuing without git commit..."))
	} else {
		fmt.Println(ui.RenderSuccess("Committed all files in obsidian folder"))
	}

	fmt.Println()
	fmt.Println(ui.RenderSuccess("Daily workflow complete!"))
	fmt.Println(ui.MutedStyle.Render("Use 'worklog add \"task\"' to add new items"))
	fmt.Println()

	return nil
}

// commitObsidianFiles adds all files in the obsidian folder and commits them
func commitObsidianFiles() error {
	obsidianPath := cfg.WorkNotesLocation

	// Check if the directory exists
	if _, err := os.Stat(obsidianPath); os.IsNotExist(err) {
		return fmt.Errorf("obsidian folder does not exist: %s", obsidianPath)
	}

	// Find the git repository root (the obsidian folder might be inside a git repo)
	gitRoot, err := findGitRoot(obsidianPath)
	if err != nil {
		return fmt.Errorf("not a git repository: %w", err)
	}

	// Get relative path from git root to obsidian folder
	relPath, err := filepath.Rel(gitRoot, obsidianPath)
	if err != nil {
		return fmt.Errorf("error calculating relative path: %w", err)
	}

	// Change to git root directory for git commands
	originalDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %w", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(gitRoot); err != nil {
		return fmt.Errorf("error changing to git root: %w", err)
	}

	// Add all files in the obsidian folder
	addCmd := exec.Command("git", "add", relPath)
	addCmd.Stdout = os.Stdout
	addCmd.Stderr = os.Stderr
	if err := addCmd.Run(); err != nil {
		return fmt.Errorf("error adding files: %w", err)
	}

	// Check if there are any changes to commit
	statusCmd := exec.Command("git", "status", "--porcelain", relPath)
	statusOutput, err := statusCmd.Output()
	if err != nil {
		return fmt.Errorf("error checking git status: %w", err)
	}

	// If no changes, skip commit
	if len(statusOutput) == 0 {
		return nil
	}

	// Create commit message with current date
	commitMessage := fmt.Sprintf("Update work notes: %s", time.Now().Format("2006-01-02"))

	// Commit the changes
	commitCmd := exec.Command("git", "commit", "-m", commitMessage)
	commitCmd.Stdout = os.Stdout
	commitCmd.Stderr = os.Stderr
	if err := commitCmd.Run(); err != nil {
		return fmt.Errorf("error committing files: %w", err)
	}

	return nil
}

// findGitRoot finds the root of the git repository by walking up from the given path
func findGitRoot(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	current := absPath
	for {
		gitDir := filepath.Join(current, ".git")
		if info, err := os.Stat(gitDir); err == nil && info.IsDir() {
			return current, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			// Reached filesystem root
			break
		}
		current = parent
	}

	return "", fmt.Errorf("no git repository found")
}
