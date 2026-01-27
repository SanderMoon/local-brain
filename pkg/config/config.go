package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/sandermoonemans/local-brain/pkg/fileutil"
)

// BrainInfo contains metadata about a brain
type BrainInfo struct {
	Path    string `json:"path"`
	Created string `json:"created"`
	Focus   string `json:"focus,omitempty"`
}

// Config represents the brain configuration
type Config struct {
	Current string                `json:"current"`
	Brains  map[string]*BrainInfo `json:"brains"`
	mu      sync.RWMutex          `json:"-"`
}

var (
	// Default paths (can be overridden by environment variables for testing)
	defaultConfigDir  = filepath.Join(os.Getenv("HOME"), ".config", "brain")
	defaultConfigFile = filepath.Join(defaultConfigDir, "config.json")
	defaultSymlink    = filepath.Join(os.Getenv("HOME"), "brain")
)

// GetConfigDir returns the configuration directory path
// Can be overridden with BRAIN_CONFIG_DIR environment variable (for testing)
func GetConfigDir() string {
	if configDir := os.Getenv("BRAIN_CONFIG_DIR"); configDir != "" {
		expanded, err := fileutil.ExpandPath(configDir)
		if err == nil {
			return expanded
		}
	}
	return defaultConfigDir
}

// GetConfigFile returns the configuration file path
// Can be overridden with BRAIN_CONFIG_PATH environment variable (for testing)
func GetConfigFile() string {
	if configPath := os.Getenv("BRAIN_CONFIG_PATH"); configPath != "" {
		expanded, err := fileutil.ExpandPath(configPath)
		if err == nil {
			return expanded
		}
	}
	return defaultConfigFile
}

// GetSymlinkPath returns the brain symlink path (can be overridden by BRAIN_SYMLINK env var)
func GetSymlinkPath() string {
	if symlink := os.Getenv("BRAIN_SYMLINK"); symlink != "" {
		expanded, err := fileutil.ExpandPath(symlink)
		if err == nil {
			return expanded
		}
	}
	return defaultSymlink
}

// Load reads the configuration from disk
func Load() (*Config, error) {
	configFile := GetConfigFile()

	// Ensure config directory exists
	if err := fileutil.EnsureDir(GetConfigDir()); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	// If config file doesn't exist, initialize it
	if !fileutil.FileExists(configFile) {
		cfg := &Config{
			Current: "",
			Brains:  make(map[string]*BrainInfo),
		}
		if err := cfg.Save(); err != nil {
			return nil, fmt.Errorf("failed to initialize config: %w", err)
		}
		return cfg, nil
	}

	// Read config file
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse JSON
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Initialize map if nil
	if cfg.Brains == nil {
		cfg.Brains = make(map[string]*BrainInfo)
	}

	return &cfg, nil
}

// Save writes the configuration to disk atomically
func (c *Config) Save() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	configFile := GetConfigFile()

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Add newline at end
	data = append(data, '\n')

	// Write atomically
	if err := fileutil.AtomicWriteFile(configFile, data); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetCurrentBrain returns the name of the current brain
func (c *Config) GetCurrentBrain() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Current
}

// GetBrain returns info for a specific brain
func (c *Config) GetBrain(name string) (*BrainInfo, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	info, exists := c.Brains[name]
	return info, exists
}

// GetBrainPath returns the path to a specific brain
func (c *Config) GetBrainPath(name string) (string, error) {
	info, exists := c.GetBrain(name)
	if !exists {
		return "", fmt.Errorf("brain '%s' not found", name)
	}
	return info.Path, nil
}

// GetCurrentBrainPath returns the path to the current active brain
func (c *Config) GetCurrentBrainPath() (string, error) {
	current := c.GetCurrentBrain()
	if current == "" {
		// Fallback to default location if no brain configured
		return GetSymlinkPath(), nil
	}
	return c.GetBrainPath(current)
}

// ListBrains returns a list of all brain names
func (c *Config) ListBrains() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	names := make([]string, 0, len(c.Brains))
	for name := range c.Brains {
		names = append(names, name)
	}
	return names
}

// BrainExists checks if a brain exists in the configuration
func (c *Config) BrainExists(name string) bool {
	_, exists := c.GetBrain(name)
	return exists
}

// AddBrain adds a new brain to the configuration
func (c *Config) AddBrain(name, path string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Expand ~ in path
	expandedPath, err := fileutil.ExpandPath(path)
	if err != nil {
		return fmt.Errorf("failed to expand path: %w", err)
	}

	// Create brain info
	created := time.Now().Format("2006-01-02")
	c.Brains[name] = &BrainInfo{
		Path:    expandedPath,
		Created: created,
	}

	return nil
}

// SetCurrentBrain sets the active brain
func (c *Config) SetCurrentBrain(name string) error {
	if !c.BrainExists(name) {
		return fmt.Errorf("brain '%s' not found", name)
	}

	c.mu.Lock()
	c.Current = name
	c.mu.Unlock()

	// Update symlink
	return UpdateSymlink(name, c)
}

// SetFocusedProject sets the focused project for the current brain
func (c *Config) SetFocusedProject(projectName string) error {
	currentBrain := c.GetCurrentBrain()
	if currentBrain == "" {
		return fmt.Errorf("no current brain set")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	brain, exists := c.Brains[currentBrain]
	if !exists {
		return fmt.Errorf("current brain '%s' not found", currentBrain)
	}

	brain.Focus = projectName
	return nil
}

// GetFocusedProject returns the focused project for the current brain
func (c *Config) GetFocusedProject() string {
	currentBrain := c.GetCurrentBrain()
	if currentBrain == "" {
		return ""
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	brain, exists := c.Brains[currentBrain]
	if !exists {
		return ""
	}

	return brain.Focus
}

// RenameBrain renames a brain in the configuration
func (c *Config) RenameBrain(oldName, newName, newPath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if old brain exists
	oldBrain, exists := c.Brains[oldName]
	if !exists {
		return fmt.Errorf("brain '%s' not found", oldName)
	}

	// Check if new name already exists
	if _, exists := c.Brains[newName]; exists {
		return fmt.Errorf("brain '%s' already exists", newName)
	}

	// Create new brain entry with updated path
	c.Brains[newName] = &BrainInfo{
		Path:    newPath,
		Created: oldBrain.Created,
		Focus:   oldBrain.Focus,
	}

	// Delete old entry
	delete(c.Brains, oldName)

	// Update current if needed
	if c.Current == oldName {
		c.Current = newName
	}

	return nil
}

// DeleteBrain removes a brain from the configuration
func (c *Config) DeleteBrain(name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if brain exists
	if _, exists := c.Brains[name]; !exists {
		return fmt.Errorf("brain '%s' not found", name)
	}

	// Delete brain
	delete(c.Brains, name)

	// Clear current if deleting current brain
	if c.Current == name {
		c.Current = ""
	}

	return nil
}
