package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/sandepten/work-obsidian-noter/internal/ui"
	"github.com/spf13/cobra"
)

var (
	pendingOnly bool
	jsonOutput  bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List today's work items",
	Long:  `Display all pending and completed work items from today's note.`,
	RunE:  runList,
}

func init() {
	listCmd.Flags().BoolVarP(&pendingOnly, "pending", "p", false, "Show only pending tasks")
	listCmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output tasks in JSON format")
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
		if jsonOutput {
			// Return empty JSON if no note found
			fmt.Println(`{"pending":[],"completed":[]}`)
			return nil
		}
		prompter.DisplayWarning("No note found for today. Use 'worklog start' to create one.")
		return nil
	}

	// JSON output format
	if jsonOutput {
		type Task struct {
			Index int    `json:"index"`
			Text  string `json:"text"`
		}
		type TasksOutput struct {
			Pending   []Task `json:"pending"`
			Completed []Task `json:"completed"`
		}

		output := TasksOutput{
			Pending:   make([]Task, 0, len(todayNote.PendingWork)),
			Completed: make([]Task, 0, len(todayNote.CompletedWork)),
		}

		for i, item := range todayNote.PendingWork {
			output.Pending = append(output.Pending, Task{
				Index: i + 1,
				Text:  item.Text,
			})
		}

		for i, item := range todayNote.CompletedWork {
			output.Completed = append(output.Completed, Task{
				Index: i + 1,
				Text:  item.Text,
			})
		}

		jsonData, err := json.Marshal(output)
		if err != nil {
			return fmt.Errorf("error marshaling JSON: %w", err)
		}
		fmt.Println(string(jsonData))
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
