package main

import "github.com/sandermoonemans/local-brain/cmd"

// Build information. Populated at build-time via ldflags
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	// Set version info in cmd package
	cmd.SetVersion(version, commit, date)
	cmd.Execute()
}
