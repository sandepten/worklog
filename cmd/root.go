package cmd

import (
	"fmt"
	"os"

	"github.com/sandepten/work-obsidian-noter/internal/config"
	"github.com/sandepten/work-obsidian-noter/internal/notes"
	"github.com/sandepten/work-obsidian-noter/internal/summarizer"
	"github.com/sandepten/work-obsidian-noter/internal/ui"
	"github.com/spf13/cobra"
)

var (
	cfg      *config.Config
	parser   *notes.Parser
	writer   *notes.Writer
	prompter *ui.Prompter
	aiClient *summarizer.Client
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "noter",
	Short: "Work Obsidian Noter - Daily work tracking CLI",
	Long: `A CLI tool for managing daily work notes in Obsidian.
	
Track your pending and completed work items, review yesterday's tasks,
and get AI-powered summaries of your accomplishments.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
}

// initConfig reads configuration and initializes dependencies
func initConfig() {
	var err error
	cfg, err = config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Ensure notes directory exists
	if err := cfg.EnsureNotesDirectory(); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating notes directory: %v\n", err)
		os.Exit(1)
	}

	// Initialize dependencies
	parser = notes.NewParser(cfg.WorkNotesLocation, cfg.WorkplaceName)
	writer = notes.NewWriter(cfg.WorkNotesLocation, cfg.WorkplaceName)
	prompter = ui.NewPrompter()
	aiClient = summarizer.NewClient(cfg.OpenCodeServer, cfg.AIProvider, cfg.AIModel)
}
