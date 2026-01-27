package cmd

import (
	"fmt"
	"os/exec"

	"github.com/sandermoonemans/local-brain/pkg/config"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync [status|scan|rescan]",
	Short: "Syncthing control",
	Long: `Syncthing control wrapper.

Commands:
  status      Show Syncthing sync status
  scan/rescan Force rescan of brain folder (default)

Requirements:
  - Syncthing installed and running
  - Syncthing CLI available

Note: Syncthing typically auto-syncs. This command is useful for:
  - Forcing immediate sync after changes
  - Checking sync status
  - Troubleshooting sync issues

Setup Syncthing:
  1. Install: brew install syncthing (macOS) or apt install syncthing (Linux)
  2. Start: syncthing (or set up as system service)
  3. Configure: http://localhost:8384
  4. Add ~/brain folder to sync
  5. Set up hub-and-spoke with NAS as hub`,
	Example: `  brain sync          # Force rescan of brain directory
  brain sync scan     # Same as above
  brain sync status   # Show sync status`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSync,
}

func init() {
	rootCmd.AddCommand(syncCmd)
}

func runSync(cmd *cobra.Command, args []string) error {
	subCmd := "scan"
	if len(args) > 0 {
		subCmd = args[0]
	}

	switch subCmd {
	case "status":
		return runSyncStatus()
	case "scan", "rescan":
		return runSyncScan()
	default:
		return fmt.Errorf("unknown command '%s'. Available: status, scan, rescan", subCmd)
	}
}

func runSyncStatus() error {
	// Check if syncthing is available
	if !isSyncthingAvailable() {
		return showSyncthingNotFound()
	}

	// Check if running
	if !isSyncthingRunning() {
		fmt.Println("Warning: Syncthing doesn't appear to be running")
		fmt.Println("")
		fmt.Println("Start Syncthing:")
		fmt.Println("  syncthing")
		fmt.Println("")
		fmt.Println("Or set up as system service for automatic startup")
		return nil
	}

	fmt.Println("Checking Syncthing status...")
	fmt.Println("")

	// Try to get status via REST API
	if _, err := exec.LookPath("curl"); err == nil {
		apiURL := "http://localhost:8384/rest/system/status"

		// Try to connect
		curlCmd := exec.Command("curl", "-s", apiURL)
		if err := curlCmd.Run(); err == nil {
			fmt.Println("OK: Syncthing is running")
			fmt.Println("")
			fmt.Println("Web interface: http://localhost:8384")
			fmt.Println("")
			fmt.Println("For detailed status, visit the web interface")
			fmt.Println("or use: syncthing cli")
		} else {
			fmt.Println("Cannot connect to Syncthing API")
			fmt.Println("Ensure Syncthing is running: syncthing")
		}
	} else {
		fmt.Println("OK: Syncthing process is running")
		fmt.Println("")
		fmt.Println("Web interface: http://localhost:8384")
		fmt.Println("")
		fmt.Println("Install curl for detailed status: brew install curl")
	}

	return nil
}

func runSyncScan() error {
	// Check if syncthing is available
	if !isSyncthingAvailable() {
		return showSyncthingNotFound()
	}

	// Check if running
	if !isSyncthingRunning() {
		fmt.Println("Warning: Syncthing doesn't appear to be running")
		fmt.Println("")
		fmt.Println("Start Syncthing first, then run: brain sync scan")
		return nil
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	brainPath, err := cfg.GetCurrentBrainPath()
	if err != nil {
		return fmt.Errorf("failed to get brain path: %w", err)
	}

	fmt.Println("Forcing rescan of brain directory...")
	fmt.Printf("Location: %s\n", brainPath)
	fmt.Println("")

	// Try to trigger rescan via REST API
	if _, err := exec.LookPath("curl"); err == nil {
		fmt.Println("Note: Triggering rescan via Syncthing REST API")
		fmt.Println("")
		fmt.Println("If this doesn't work, manually trigger rescan:")
		fmt.Println("  1. Visit: http://localhost:8384")
		fmt.Println("  2. Find brain folder")
		fmt.Println("  3. Click 'Rescan' button")
		fmt.Println("")

		// Attempt generic rescan
		apiURL := "http://localhost:8384/rest/db/scan"
		curlCmd := exec.Command("curl", "-s", "-X", "POST", apiURL)

		if err := curlCmd.Run(); err == nil {
			fmt.Println("OK: Rescan triggered")
		} else {
			fmt.Println("Could not trigger automatic rescan")
			fmt.Println("Manual rescan required via web interface")
		}
	} else {
		fmt.Println("Syncthing is running but automatic rescan requires curl")
		fmt.Println("")
		fmt.Println("Option 1: Install curl")
		fmt.Println("  brew install curl (macOS)")
		fmt.Println("  apt install curl (Linux)")
		fmt.Println("")
		fmt.Println("Option 2: Manual rescan")
		fmt.Println("  Visit: http://localhost:8384")
		fmt.Println("  Find brain folder and click 'Rescan'")
	}

	return nil
}

func isSyncthingAvailable() bool {
	_, err := exec.LookPath("syncthing")
	return err == nil
}

func isSyncthingRunning() bool {
	cmd := exec.Command("pgrep", "-x", "syncthing")
	return cmd.Run() == nil
}

func showSyncthingNotFound() error {
	fmt.Println("Error: Syncthing not found")
	fmt.Println("")
	fmt.Println("Install Syncthing:")
	fmt.Println("  macOS: brew install syncthing")
	fmt.Println("  Linux: apt install syncthing (or dnf/pacman)")
	fmt.Println("")
	fmt.Println("After installation:")
	fmt.Println("  1. Run: syncthing")
	fmt.Println("  2. Visit: http://localhost:8384")
	fmt.Println("  3. Add ~/brain to sync folders")
	return nil
}
