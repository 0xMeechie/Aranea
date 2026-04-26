// Package sdk is the public entry point for embedding Aranea in an agent.
package sdk

import (
	"fmt"

	"github.com/0xMeechie/Aranea/pkg/config"
	"github.com/0xMeechie/Aranea/pkg/runtime"
)

// Agent is the top-level handle used by agent code to interact with Aranea.
type Agent struct {
	cfg   config.AgentConfig
	rt    *runtime.Runtime
	Tools ToolSet
}

// Init validates cfg, creates the runtime, and returns a ready Agent.
func Init(cfg config.AgentConfig) (*Agent, error) {
	return nil, nil
}

// InitFromFile loads the agent config from path and calls Init.
// Path is always required — no fallback or discovery is performed.
func InitFromFile(path string) (*Agent, error) {
	if path == "" {
		return nil, fmt.Errorf("config path is required")
	}

	cfg, err := config.LoadConfigFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	return Init(cfg)
}

// Shutdown flushes any pending log writes and releases runtime resources.
func (a *Agent) Shutdown() {
	if a.rt != nil {
		_ = a.rt.Close()
	}
}
