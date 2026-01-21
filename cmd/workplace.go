package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sandepten/work-obsidian-noter/internal/notes"
	"github.com/sandepten/work-obsidian-noter/internal/ui"
	"github.com/spf13/cobra"
)

var workplaceCmd = &cobra.Command{
	Use:   "workplace",
	Short: "Manage workplaces",
	Long:  `Manage your workplaces - add new workplaces or rename existing ones.`,
}

var workplaceAddCmd = &cobra.Command{
	Use:   "add [name]",
	Short: "Add a new workplace",
	Long:  `Add a new workplace to your configuration. If no name is provided, you will be prompted to enter one.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runWorkplaceAdd,
}

var workplaceRenameCmd = &cobra.Command{
	Use:   "rename",
	Short: "Rename an existing workplace",
	Long:  `Rename an existing workplace. This will also rename all associated note files.`,
	RunE:  runWorkplaceRename,
}

var workplaceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured workplaces",
	Long:  `List all workplaces currently configured.`,
	RunE:  runWorkplaceList,
}

func init() {
	rootCmd.AddCommand(workplaceCmd)
	workplaceCmd.AddCommand(workplaceAddCmd)
	workplaceCmd.AddCommand(workplaceRenameCmd)
	workplaceCmd.AddCommand(workplaceListCmd)
}

func runWorkplaceAdd(cmd *cobra.Command, args []string) error {
	var workplaceName string
	var err error

	if len(args) > 0 {
		workplaceName = strings.TrimSpace(args[0])
		// Validate the name
		if workplaceName == "" {
			return fmt.Errorf("workplace name cannot be empty")
		}
		if strings.Contains(workplaceName, ",") {
			return fmt.Errorf("workplace name cannot contain commas")
		}
	} else {
		// Prompt for workplace name
		workplaceName, err = prompter.PromptForWorkplaceName("New workplace name")
		if err != nil {
			if err.Error() == "cancelled" {
				fmt.Println(ui.RenderWarning("Cancelled"))
				return nil
			}
			return fmt.Errorf("error getting workplace name: %w", err)
		}
	}

	// Add the workplace
	if err := cfg.AddWorkplace(workplaceName); err != nil {
		return fmt.Errorf("failed to add workplace: %w", err)
	}

	fmt.Println()
	fmt.Println(ui.RenderSuccess(fmt.Sprintf("Workplace '%s' added successfully!", workplaceName)))
	fmt.Println(ui.MutedStyle.Render(fmt.Sprintf("  You now have %d workplace(s) configured", len(cfg.Workplaces))))
	fmt.Println()

	return nil
}

func runWorkplaceRename(cmd *cobra.Command, args []string) error {
	// Select workplace to rename
	oldName, err := prompter.SelectWorkplaceToRename(cfg.Workplaces)
	if err != nil {
		return fmt.Errorf("error selecting workplace: %w", err)
	}

	fmt.Println()
	fmt.Println(ui.RenderInfo(fmt.Sprintf("Renaming workplace '%s'", oldName)))

	// Prompt for new name
	newName, err := prompter.PromptForWorkplaceName("New name")
	if err != nil {
		if err.Error() == "cancelled" {
			fmt.Println(ui.RenderWarning("Cancelled"))
			return nil
		}
		return fmt.Errorf("error getting new workplace name: %w", err)
	}

	// Confirm the rename with file updates
	confirmed, err := prompter.ConfirmAction(fmt.Sprintf("Rename '%s' to '%s' and update all note files", oldName, newName))
	if err != nil {
		return fmt.Errorf("error confirming action: %w", err)
	}
	if !confirmed {
		fmt.Println(ui.RenderWarning("Cancelled"))
		return nil
	}

	// Rename files first
	renamedCount, err := renameWorkplaceFiles(cfg.WorkNotesLocation, oldName, newName)
	if err != nil {
		return fmt.Errorf("error renaming files: %w", err)
	}

	// Update config
	if err := cfg.RenameWorkplace(oldName, newName); err != nil {
		return fmt.Errorf("failed to rename workplace in config: %w", err)
	}

	fmt.Println()
	fmt.Println(ui.RenderSuccess(fmt.Sprintf("Workplace renamed from '%s' to '%s'!", oldName, newName)))
	if renamedCount > 0 {
		fmt.Println(ui.MutedStyle.Render(fmt.Sprintf("  Renamed %d note file(s)", renamedCount)))
	}
	fmt.Println()

	return nil
}

func runWorkplaceList(cmd *cobra.Command, args []string) error {
	fmt.Println()
	fmt.Println(ui.RenderHeader("Configured Workplaces"))
	fmt.Println()

	if len(cfg.Workplaces) == 0 {
		fmt.Println(ui.RenderWarning("No workplaces configured"))
		return nil
	}

	for i, wp := range cfg.Workplaces {
		fmt.Printf("  %d. %s\n", i+1, ui.SuccessStyle.Render(wp))
	}
	fmt.Println()

	return nil
}

// renameWorkplaceFiles renames all note files for a workplace
func renameWorkplaceFiles(notesDir, oldName, newName string) (int, error) {
	renamedCount := 0

	// Find all files matching the pattern *-OldName.md
	pattern := fmt.Sprintf("*-%s.md", oldName)
	matches, err := filepath.Glob(filepath.Join(notesDir, pattern))
	if err != nil {
		return 0, fmt.Errorf("error finding files: %w", err)
	}

	for _, oldPath := range matches {
		// Get the filename
		filename := filepath.Base(oldPath)

		// Replace the workplace name in the filename
		newFilename := strings.Replace(filename, fmt.Sprintf("-%s.md", oldName), fmt.Sprintf("-%s.md", newName), 1)
		newPath := filepath.Join(notesDir, newFilename)

		// Rename the file
		if err := os.Rename(oldPath, newPath); err != nil {
			return renamedCount, fmt.Errorf("error renaming file %s: %w", filename, err)
		}

		// Update file content (ID and tags)
		if err := updateNoteContent(newPath, oldName, newName); err != nil {
			// Log warning but don't fail
			fmt.Println(ui.RenderWarning(fmt.Sprintf("Warning: could not update content in %s: %v", newFilename, err)))
		}

		renamedCount++
	}

	return renamedCount, nil
}

// updateNoteContent updates the workplace references inside a note file
func updateNoteContent(filePath, oldName, newName string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	contentStr := string(content)

	// Update the ID (WorkplaceName-D-Mon-YYYY)
	contentStr = strings.ReplaceAll(contentStr, oldName+"-", newName+"-")

	// Update the tags (lowercase workplace name)
	oldTag := notes.ToLowerCase(oldName)
	newTag := notes.ToLowerCase(newName)
	contentStr = strings.ReplaceAll(contentStr, fmt.Sprintf("- %s", oldTag), fmt.Sprintf("- %s", newTag))

	return os.WriteFile(filePath, []byte(contentStr), 0644)
}
