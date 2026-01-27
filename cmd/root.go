package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

// SetVersion sets the version information (called from main)
func SetVersion(v, c, d string) {
	version = v
	commit = c
	date = d
	rootCmd.Version = buildVersion()
}

// buildVersion constructs a detailed version string
func buildVersion() string {
	result := version
	if commit != "unknown" {
		result += fmt.Sprintf(" (commit: %s)", commit)
	}
	if date != "unknown" {
		result += fmt.Sprintf(" (built: %s)", date)
	}
	return result
}

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "brain",
	Short: "Local Brain - Personal knowledge management CLI",
	Long: `Local Brain is a personal knowledge management system
that helps you capture tasks, notes, and organize projects.

Features:
  - Quick task and note capture
  - Project-based organization
  - Git repository linking
  - Tmux workspace integration
  - Programmatic JSON API`,
	Version: buildVersion(),
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Customize version output
	rootCmd.SetVersionTemplate(`{{printf "%s\n" .Version}}`)

	// Global flags can be added here
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/brain/config.json)")
}
