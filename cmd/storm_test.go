package cmd

import (
	"testing"
)

func TestStormCommand_DefaultAgent(t *testing.T) {
	// Verify the default agent flag value is "claude"
	cmd := stormCmd
	agentFlag := cmd.Flags().Lookup("agent")
	if agentFlag == nil {
		t.Fatal("expected --agent flag to be defined")
	}
	if agentFlag.DefValue != "claude" {
		t.Errorf("expected default agent to be %q, got %q", "claude", agentFlag.DefValue)
	}
}

func TestStormCommand_ShortFlag(t *testing.T) {
	// Verify the -a shorthand is defined
	cmd := stormCmd
	agentFlag := cmd.Flags().ShorthandLookup("a")
	if agentFlag == nil {
		t.Fatal("expected -a shorthand flag to be defined")
	}
}

func TestStormCommand_MissingAgent(t *testing.T) {
	// Verify that an unknown agent returns an error
	orig := stormAgent
	stormAgent = "this-agent-does-not-exist-xyz"
	defer func() { stormAgent = orig }()

	err := runStorm(stormCmd, nil)
	if err == nil {
		t.Fatal("expected error for missing agent, got nil")
	}
}
