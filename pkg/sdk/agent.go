// Package sdk is the public entry point for embedding Aranea in an agent.
package sdk

import (
	"fmt"

	"github.com/0xMeechie/Aranea/pkg/config"
)

const (
	defaultAraneaEndpoint = "http://127.0.0.1:8080"
)

// Agent is the top-level handle used by agent code to interact with Aranea.
type AraneaClient struct {
	client *HTTPClient
}

type ClientConfig struct {
	// Endpoint will default to the managed Aranea servers. If self hosted
	// enter the endpoint to where the Aranea runtime is running
	Endpoint string
}

type AraneaAgent struct{}

// Init validates cfg, and returns a ready Aranea Agent.
func Init(cfg ClientConfig) (*AraneaClient, error) {
	var endpoint string
	if cfg.Endpoint == "" {
		endpoint = defaultAraneaEndpoint
	} else {
		endpoint = cfg.Endpoint
	}

	sc, _, err := client.Fetch("GET", endpoint+"/healthz", "")
	if err != nil {
		return &AraneaClient{}, fmt.Errorf("error initializing connection %w", err)
	}

	if sc != 200 {
		return &AraneaClient{}, fmt.Errorf("unable to reach aranea endpoint")
	}
	return &AraneaClient{
		client: client,
	}, nil
}

// InitFromFile loads the agent config from path and calls Init.
// Path is always required — no fallback or discovery is performed.
func InitFromFile(path string) (*AraneaClient, error) {
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
func (a *AraneaClient) Shutdown() {
}
