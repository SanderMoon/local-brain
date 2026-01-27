package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sandermoonemans/local-brain/pkg/api"
	"github.com/sandermoonemans/local-brain/pkg/config"
	"github.com/sandermoonemans/local-brain/pkg/external"
	"github.com/spf13/cobra"
)

var noteJSONFlag bool

var noteCmd = &cobra.Command{
	Use:   "note",
	Short: "Manage notes in projects",
	Long: `Manage notes in the focused project.

Interactive mode (no subcommand):
  Select and edit a note with fuzzy finder

Subcommands:
  ls       List all notes in focused project
  delete   Delete a note

Notes are stored as separate files in project/notes/ directory.
To create new notes, use: brain add → brain refile

Requires a focused project. Set with: brain project select <name>`,
	Example: `  brain note               # Pick a note to edit
  brain note ls            # List all notes
  brain note ls --json     # List notes as JSON
  brain note delete        # Pick a note to delete`,
	RunE: runNoteInteractive,
}

var noteLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List notes",
	Long:  "List all notes in the focused project",
	RunE:  runNoteLs,
}

var noteDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a note",
	Long:  "Select and delete a note from the focused project",
	RunE:  runNoteDelete,
}

func init() {
	rootCmd.AddCommand(noteCmd)
	noteCmd.AddCommand(noteLsCmd)
	noteCmd.AddCommand(noteDeleteCmd)

	noteLsCmd.Flags().BoolVar(&noteJSONFlag, "json", false, "Output JSON format")
}

func runNoteLs(cmd *cobra.Command, args []string) error {
	projectDir, err := getFocusedProjectDir()
	if err != nil {
		return err
	}

	notes, err := api.ListNotes(projectDir)
	if err != nil {
		return fmt.Errorf("failed to list notes: %w", err)
	}

	if len(notes) == 0 {
		if noteJSONFlag {
			fmt.Println("[]")
		} else {
			fmt.Println("No notes found")
		}
		return nil
	}

	if noteJSONFlag {
		data, err := json.MarshalIndent(notes, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(data))
	} else {
		// Human-readable list
		for _, note := range notes {
			fmt.Printf("%-35s  %s\n", note.Filename, note.Title)
		}
	}

	return nil
}

func runNoteDelete(cmd *cobra.Command, args []string) error {
	if !external.IsFZFAvailable() {
		return fmt.Errorf("fzf not found (required for interactive mode)")
	}

	projectDir, err := getFocusedProjectDir()
	if err != nil {
		return err
	}

	notes, err := api.ListNotes(projectDir)
	if err != nil {
		return fmt.Errorf("failed to list notes: %w", err)
	}

	if len(notes) == 0 {
		fmt.Println("No notes found")
		return nil
	}

	// Format for FZF
	var items []string
	for _, note := range notes {
		item := fmt.Sprintf("%s|||%s  %s", note.Path, note.Filename, note.Title)
		items = append(items, item)
	}

	// Select with FZF
	selected, err := external.SelectOne(items, external.FZFOptions{
		Header:        "Select note to DELETE (Esc to cancel)",
		Preview:       "bat --color=always --style=header,numbers $(echo {} | cut -d'|' -f1) 2>/dev/null || head -30 $(echo {} | cut -d'|' -f1)",
		PreviewWindow: "right:60%",
	})

	if err != nil {
		if err.Error() == "cancelled" {
			return nil
		}
		return err
	}

	// Parse selection
	notePath := strings.Split(selected, "|||")[0]

	// Get filename for display
	filename := filepath.Base(notePath)

	// Confirmation
	fmt.Printf("About to delete: %s\n", filename)
	fmt.Print("Are you sure? [y/N] ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read confirmation: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	if response != "y" && response != "yes" {
		fmt.Println("Cancelled")
		return nil
	}

	// Delete
	if err := api.DeleteNote(notePath); err != nil {
		return fmt.Errorf("failed to delete note: %w", err)
	}

	fmt.Printf("OK: Deleted note: %s\n", filename)
	return nil
}

func runNoteInteractive(cmd *cobra.Command, args []string) error {
	if !external.IsFZFAvailable() {
		return fmt.Errorf("fzf not found (required for interactive mode)")
	}

	projectDir, err := getFocusedProjectDir()
	if err != nil {
		return err
	}

	notes, err := api.ListNotes(projectDir)
	if err != nil {
		return fmt.Errorf("failed to list notes: %w", err)
	}

	if len(notes) == 0 {
		fmt.Println("No notes found in project")
		fmt.Println("Create notes with: brain add → brain refile")
		return nil
	}

	// Format for FZF
	var items []string
	for _, note := range notes {
		item := fmt.Sprintf("%s|||%s  %s", note.Path, note.Filename, note.Title)
		items = append(items, item)
	}

	// Select with FZF and preview
	selected, err := external.SelectOne(items, external.FZFOptions{
		Header:        "Select note to edit (Esc to cancel)",
		Preview:       "bat --color=always --style=header,numbers $(echo {} | cut -d'|' -f1) 2>/dev/null || head -30 $(echo {} | cut -d'|' -f1)",
		PreviewWindow: "right:60%",
	})

	if err != nil {
		if err.Error() == "cancelled" {
			return nil
		}
		return err
	}

	// Parse selection
	notePath := strings.Split(selected, "|||")[0]

	// Open in editor
	return external.OpenFile(notePath)
}

// Helper functions

func getFocusedProjectDir() (string, error) {
	cfg, err := config.Load()
	if err != nil {
		return "", fmt.Errorf("failed to load config: %w", err)
	}

	focused := cfg.GetFocusedProject()
	if focused == "" {
		return "", fmt.Errorf("no focused project. Set one with: brain project select <name>")
	}

	brainPath, err := cfg.GetCurrentBrainPath()
	if err != nil {
		return "", fmt.Errorf("failed to get brain path: %w", err)
	}

	projectDir := filepath.Join(brainPath, "01_active", focused)

	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		return "", fmt.Errorf("project directory not found: %s", projectDir)
	}

	return projectDir, nil
}
