package skillscatalog

import "github.com/sandermoonemans/local-brain/pkg/agents"

// AgentInfo describes a supported AI coding agent.
// This is a type alias for agents.Info — all existing code continues to work.
type AgentInfo = agents.Info

// KnownAgents returns all supported agents with paths resolved from HOME.
func KnownAgents() []AgentInfo {
	all := agents.All()
	result := make([]AgentInfo, len(all))
	for i, a := range all {
		result[i] = a.Info
	}
	return result
}

// DetectedAgents returns only those agents whose ConfigDir exists on disk.
func DetectedAgents() []AgentInfo {
	detected := agents.Detected()
	result := make([]AgentInfo, len(detected))
	for i, a := range detected {
		result[i] = a.Info
	}
	return result
}

// FindAgent returns the AgentInfo for the given ID, or an error if unknown.
func FindAgent(id string) (AgentInfo, error) {
	a, err := agents.Find(id)
	if err != nil {
		return AgentInfo{}, err
	}
	return a.Info, nil
}
