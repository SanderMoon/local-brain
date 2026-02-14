package agents

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// JSONMCPInstaller manages MCP server registration in JSON config files.
// Used by agents that store MCP config in JSON format (Claude Code, Gemini CLI, OpenCode).
//
// To use: set ConfigPath to the agent's config file, ServerKey to the top-level
// key holding MCP server entries, and BuildEntry to a function that returns
// the JSON-serializable entry for a given binary path.
type JSONMCPInstaller struct {
	ConfigPath string                              // e.g. ~/.gemini/settings.json
	ServerKey  string                              // e.g. "mcpServers"
	BuildEntry func(binaryPath string) interface{} // returns the entry value
}

func (j *JSONMCPInstaller) Install(serverName, binaryPath string) error {
	config, err := j.readConfig()
	if err != nil {
		return err
	}

	servers, _ := config[j.ServerKey].(map[string]interface{})
	if servers == nil {
		servers = make(map[string]interface{})
	}
	servers[serverName] = j.BuildEntry(binaryPath)
	config[j.ServerKey] = servers

	return j.writeConfig(config)
}

func (j *JSONMCPInstaller) Remove(serverName string) error {
	data, err := os.ReadFile(j.ConfigPath)
	if err != nil {
		return nil // file doesn't exist = nothing to remove
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("parsing %s: %w", j.ConfigPath, err)
	}

	servers, _ := config[j.ServerKey].(map[string]interface{})
	if servers == nil {
		return nil
	}
	delete(servers, serverName)
	if len(servers) == 0 {
		delete(config, j.ServerKey)
	} else {
		config[j.ServerKey] = servers
	}

	return j.writeConfig(config)
}

func (j *JSONMCPInstaller) IsInstalled(serverName string) (bool, error) {
	data, err := os.ReadFile(j.ConfigPath)
	if err != nil {
		return false, nil // file doesn't exist = not installed
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return false, nil // can't parse = assume not installed
	}

	servers, _ := config[j.ServerKey].(map[string]interface{})
	_, exists := servers[serverName]
	return exists, nil
}

// readConfig reads the JSON config file, returning an empty map if it doesn't exist.
func (j *JSONMCPInstaller) readConfig() (map[string]interface{}, error) {
	data, err := os.ReadFile(j.ConfigPath)
	if os.IsNotExist(err) {
		return make(map[string]interface{}), nil
	}
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return make(map[string]interface{}), nil
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", j.ConfigPath, err)
	}
	return config, nil
}

// writeConfig writes the config map as pretty-printed JSON.
func (j *JSONMCPInstaller) writeConfig(config map[string]interface{}) error {
	if err := os.MkdirAll(filepath.Dir(j.ConfigPath), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(j.ConfigPath, data, 0644)
}
