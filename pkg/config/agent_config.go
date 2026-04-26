// Package agent handles all the config data for agents
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

type AgentConfig struct {
	AgentConfig   Agent   `yaml:"agent"`
	RuntimeConfig Runtime `yaml:"runtime"`
	PolicyConfig  Policy  `yaml:"policy"`
	AuditConfig   Audit   `yaml:"audit"`
}

type Agent struct {
	ID      string `yaml:"id"`
	Version string `yaml:"version"`
}

type Runtime struct {
	Endpoint     string `yaml:"endpoint"`
	FailBehavior string `yaml:"failBehavior"`
}

type Policy struct {
	DefaultBehavior string                 `yaml:"defaultBehavior,omitempty"`
	Rules           []Rule                 `yaml:"rules,omitempty"`
	RuleMap         map[string]ParamConfig `yaml:"-"`
}

type Audit struct {
	Level        string `yaml:"level,omitempty"`
	Sign         bool   `yaml:"sign,omitempty"`
	SyncLocation string `yaml:"syncTo,omitempty"`
}

type Rule struct {
	Tool       string      `yaml:"tool,omitempty"`
	Parameters ParamConfig `yaml:"parameters,omitempty"`
}

type ParamConfig struct {
	Domains        []string `yaml:"domains,omitempty"`
	Paths          []string `yaml:"paths,omitempty"`
	RateLimit      int      `yaml:"rateLimit,omitempty"`
	AllowedActions []string `yaml:"allowedActions,omitempty"`
	Tables         []string `yaml:"tables,omitempty"`
}

// simple validate for now
func (c *AgentConfig) validate() error {
	if c.AgentConfig.ID == "" {
		return fmt.Errorf("Agent ID is required")
	}

	if c.PolicyConfig.DefaultBehavior == "" {
		c.PolicyConfig.DefaultBehavior = "deny"
	}

	c.PolicyConfig.RuleMap = make(map[string]ParamConfig, len(c.PolicyConfig.Rules))
	for _, rule := range c.PolicyConfig.Rules {
		c.PolicyConfig.RuleMap[rule.Tool] = rule.Parameters
	}

	return nil
}

// LoadConfigFromFile takes in a file path and returns the agent config
func LoadConfigFromFile(filePath string) (AgentConfig, error) {
	if filePath == "" {
		return AgentConfig{}, fmt.Errorf("valid config path is required")
	}
	// validate file is the correct type
	extType := filepath.Ext(filePath)
	switch extType {
	case ".yaml", ".yml":
		fmt.Println("correct file type")
	default:
		return AgentConfig{}, fmt.Errorf("agent config file must be a yaml file. Got a %s instead", extType)
	}
	var config AgentConfig
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return AgentConfig{}, fmt.Errorf("could not open config file: %s", err)
	}
	err = yaml.Unmarshal(fileBytes, &config)
	if err != nil {
		return AgentConfig{}, fmt.Errorf("could not unmarshall config %s", err)
	}

	err = config.validate()
	if err != nil {
		return AgentConfig{}, err
	}
	return config, nil
}
