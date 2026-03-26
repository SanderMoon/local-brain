package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sandermoonemans/local-brain/pkg/config"
	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config [key] [value]",
	Short: "Get or set configuration values",
	Long: `Get or set configuration values.

When called with no arguments, shows all current config values.

Available keys:
  editor      Preferred text editor (e.g., nvim, vim, emacs, nano)
  dev_dir     Directory for cloned repositories (default: ~/dev)
  brain_root  Root directory for brains (read-only, use BRAIN_ROOT env var)
  symlink     Symlink location (read-only, use BRAIN_SYMLINK env var)

Subcommands:
  brain config show     Show all current config values
  brain config setup    Interactive configuration wizard

Examples:
  brain config                       Show all config values
  brain config editor                Show current editor
  brain config editor nvim           Set editor to nvim
  brain config dev_dir ~/projects    Set dev directory
  brain config setup                 Run interactive setup wizard`,
	Args: cobra.MinimumNArgs(0),
	RunE: runConfig,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show all current configuration values",
	RunE:  runConfigShow,
}

var configSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Interactive configuration wizard",
	Long:  `Walk through key settings interactively. Press Enter to keep the current/default value or type a new value.`,
	RunE:  runConfigSetup,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetupCmd)
}

// tildify replaces the home directory prefix with ~ for display purposes
func tildify(path string) string {
	home := os.Getenv("HOME")
	if home != "" && strings.HasPrefix(path, home) {
		return "~" + path[len(home):]
	}
	return path
}

// configSource returns the source label for a config value
func configSource(configValue, envVar string) string {
	if configValue != "" {
		return "(configured)"
	}
	if envVar != "" && os.Getenv(envVar) != "" {
		return "(env: " + envVar + ")"
	}
	return "(default)"
}

func printConfigTable(cfg *config.Config) {
	home := os.Getenv("HOME")
	defaultBrainRoot := filepath.Join(home, "brains")
	defaultSymlink := filepath.Join(home, "brain")
	brainRoot := config.GetBrainRoot()
	symlinkPath := config.GetSymlinkPath()
	devDir := cfg.GetDevDir()
	editor := cfg.GetEditor()
	current := cfg.GetCurrentBrain()

	// Determine sources
	brainRootSource := configSource("", "BRAIN_ROOT")
	if brainRoot != defaultBrainRoot && os.Getenv("BRAIN_ROOT") == "" {
		// dev mode changes default
		brainRootSource = "(default)"
	}

	symlinkSource := configSource("", "BRAIN_SYMLINK")
	if symlinkPath != defaultSymlink && os.Getenv("BRAIN_SYMLINK") == "" {
		symlinkSource = "(default)"
	}

	devDirSource := configSource(cfg.GetRawDevDir(), "BRAIN_DEV_DIR")

	editorDisplay := editor
	editorSource := "(configured)"
	if editor == "" {
		// Try to auto-detect for display
		editorDisplay = detectEditorName()
		if editorDisplay == "" {
			editorDisplay = "(none)"
		}
		editorSource = "(auto-detected)"
	}

	currentDisplay := current
	if currentDisplay == "" {
		currentDisplay = "(none)"
	}

	fmt.Printf("  %-14s %-22s %s\n", "brain_root", tildify(brainRoot), brainRootSource)
	fmt.Printf("  %-14s %-22s %s\n", "symlink", tildify(symlinkPath), symlinkSource)
	fmt.Printf("  %-14s %-22s %s\n", "dev_dir", tildify(devDir), devDirSource)
	fmt.Printf("  %-14s %-22s %s\n", "editor", editorDisplay, editorSource)
	fmt.Printf("  %-14s %-22s\n", "current", currentDisplay)
}

// detectEditorName returns the name of the auto-detected editor (for display only)
func detectEditorName() string {
	if _, err := exec.LookPath("nvim"); err == nil {
		return "nvim"
	}
	if _, err := exec.LookPath("vim"); err == nil {
		return "vim"
	}
	if editorEnv := os.Getenv("EDITOR"); editorEnv != "" {
		return editorEnv
	}
	return ""
}

func runConfig(cmd *cobra.Command, args []string) error {
	// If no args and no subcommand was invoked, show config table
	if len(args) == 0 {
		return runConfigShow(cmd, args)
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	key := args[0]

	// Get operation (no value provided)
	if len(args) == 1 {
		switch key {
		case "editor":
			editor := cfg.GetEditor()
			if editor == "" {
				fmt.Println("(not set - using auto-detection: nvim > vim > $EDITOR)")
			} else {
				fmt.Println(editor)
			}
			return nil
		case "dev_dir":
			fmt.Println(tildify(cfg.GetDevDir()))
			return nil
		case "brain_root":
			fmt.Println(tildify(config.GetBrainRoot()))
			if os.Getenv("BRAIN_ROOT") != "" {
				fmt.Println("  (set via BRAIN_ROOT environment variable)")
			} else {
				fmt.Println("  (default; override with BRAIN_ROOT environment variable)")
			}
			return nil
		case "symlink":
			fmt.Println(tildify(config.GetSymlinkPath()))
			if os.Getenv("BRAIN_SYMLINK") != "" {
				fmt.Println("  (set via BRAIN_SYMLINK environment variable)")
			} else {
				fmt.Println("  (default; override with BRAIN_SYMLINK environment variable)")
			}
			return nil
		default:
			return fmt.Errorf("unknown config key: %s\nAvailable keys: editor, dev_dir, brain_root, symlink", key)
		}
	}

	// Set operation (value provided)
	value := args[1]
	switch key {
	case "editor":
		cfg.SetEditor(value)
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		fmt.Printf("Editor set to: %s\n", value)
		return nil
	case "dev_dir":
		cfg.SetDevDir(value)
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		fmt.Printf("Dev directory set to: %s\n", value)
		return nil
	case "brain_root":
		fmt.Println("brain_root is read-only in config.")
		fmt.Println("Set the BRAIN_ROOT environment variable to override it.")
		return nil
	case "symlink":
		fmt.Println("symlink is read-only in config.")
		fmt.Println("Set the BRAIN_SYMLINK environment variable to override it.")
		return nil
	default:
		return fmt.Errorf("unknown config key: %s\nAvailable keys: editor, dev_dir, brain_root (read-only), symlink (read-only)", key)
	}
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println("Current configuration:")
	fmt.Println()
	printConfigTable(cfg)
	fmt.Println()
	fmt.Println("Config file:", tildify(config.GetConfigFile()))

	return nil
}

func runConfigSetup(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	reader := bufio.NewReader(os.Stdin)
	changed := false

	fmt.Println("Local Brain configuration setup")
	fmt.Println("Press Enter to keep the current value, or type a new value.")
	fmt.Println()

	// 1. Brain root directory (read-only, informational)
	brainRoot := config.GetBrainRoot()
	fmt.Printf("Brain root directory: %s", tildify(brainRoot))
	if os.Getenv("BRAIN_ROOT") != "" {
		fmt.Print("  (via BRAIN_ROOT env var)")
	}
	fmt.Println()
	fmt.Println("  (Set BRAIN_ROOT environment variable to change)")
	fmt.Println()

	// 2. Dev directory
	currentDevDir := tildify(cfg.GetDevDir())
	fmt.Printf("Dev directory for repos [%s]: ", currentDevDir)
	devDirInput, _ := reader.ReadString('\n')
	devDirInput = strings.TrimSpace(devDirInput)
	if devDirInput != "" && devDirInput != currentDevDir {
		cfg.SetDevDir(devDirInput)
		changed = true
	}

	// 3. Editor
	currentEditor := cfg.GetEditor()
	editorDefault := currentEditor
	if editorDefault == "" {
		editorDefault = detectEditorName()
	}
	displayEditor := editorDefault
	if displayEditor == "" {
		displayEditor = "auto-detect"
	}
	fmt.Printf("Editor [%s]: ", displayEditor)
	editorInput, _ := reader.ReadString('\n')
	editorInput = strings.TrimSpace(editorInput)
	if editorInput != "" && editorInput != currentEditor {
		cfg.SetEditor(editorInput)
		changed = true
	}

	// 4. Symlink location (read-only, informational)
	symlinkPath := config.GetSymlinkPath()
	fmt.Printf("Symlink location: %s", tildify(symlinkPath))
	if os.Getenv("BRAIN_SYMLINK") != "" {
		fmt.Print("  (via BRAIN_SYMLINK env var)")
	}
	fmt.Println()
	fmt.Println("  (Set BRAIN_SYMLINK environment variable to change)")
	fmt.Println()

	// Save if anything changed
	if changed {
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		fmt.Println("Configuration saved!")
		fmt.Println()
		fmt.Println("Updated configuration:")
		fmt.Println()
		printConfigTable(cfg)
	} else {
		fmt.Println("No changes made.")
	}

	return nil
}
