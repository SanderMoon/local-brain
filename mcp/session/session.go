package session

import (
	"sync"
	"time"

	"github.com/sandermoonemans/local-brain/pkg/api"
	"github.com/sandermoonemans/local-brain/pkg/config"
)

// Session manages state and caching for the MCP server
type Session struct {
	mu            sync.RWMutex
	cfg           *config.Config
	lastRefresh   time.Time
	cacheTTL      time.Duration
	overviewCache *api.BrainOverview
}

// NewSession creates a new session manager
func NewSession(cfg *config.Config) *Session {
	return &Session{
		cfg:      cfg,
		cacheTTL: 30 * time.Second,
	}
}

// GetConfig returns the brain configuration
func (s *Session) GetConfig() *config.Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cfg
}

// RefreshConfig reloads the configuration from disk
func (s *Session) RefreshConfig() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.cfg = cfg
	s.mu.Unlock()

	return nil
}

// GetBrainOverview returns cached brain overview or fetches fresh data
func (s *Session) GetBrainOverview() (*api.BrainOverview, error) {
	s.mu.RLock()
	if s.overviewCache != nil && time.Since(s.lastRefresh) < s.cacheTTL {
		defer s.mu.RUnlock()
		return s.overviewCache, nil
	}
	s.mu.RUnlock()

	// Refresh cache
	cfg := s.GetConfig()
	overview, err := api.GetBrainOverview(cfg)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	s.overviewCache = overview
	s.lastRefresh = time.Now()
	s.mu.Unlock()

	return overview, nil
}

// Invalidate clears the cache (called after mutations)
func (s *Session) Invalidate() {
	s.mu.Lock()
	s.overviewCache = nil
	s.mu.Unlock()
}
