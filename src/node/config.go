package node

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// LogLevel controls verbosity of node logging.
type LogLevel string

const (
	LogLevelInfo    LogLevel = "info"
	LogLevelDebug   LogLevel = "debug"
	LogLevelVerbose LogLevel = "verbose"
)

// NodeConfig is the on-disk config stored at $HOME/.config/aranea/config/config.json.
type NodeConfig struct {
	Version      int      `json:"version"`
	ConfigDir    string   `json:"config_dir"`
	ConfigFile   string   `json:"config_file"`
	IdentityFile string   `json:"identity_file"`
	AuditDir     string   `json:"audit_dir"`
	LogLevel     LogLevel `json:"log_level"`
}

func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("HOME not set: %w", err)
	}
	return filepath.Join(home, DefaultConfigLocation), nil
}

// LoadConfig reads the node config from disk.
func LoadConfig() (*NodeConfig, error) {
	dir, err := configDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(dir, "config", ConfigFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read config: %w", err)
	}
	var cfg NodeConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	return &cfg, nil
}

// InitNode sets up the node directory structure and writes default config and identity.
// If force is false and files already exist, it is a no-op.
func InitNode(force bool) error {
	dir, err := configDir()
	if err != nil {
		return err
	}

	dirs := []string{
		filepath.Join(dir, "identity"),
		filepath.Join(dir, "config"),
		filepath.Join(dir, "audit"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, DefaultFolderPermissions); err != nil {
			return fmt.Errorf("failed to create dir %s: %w", d, err)
		}
	}

	cfgPath := filepath.Join(dir, "config", ConfigFile)
	if !force {
		if _, err := os.Stat(cfgPath); err == nil {
			return nil // already initialised
		}
	}

	idPath := filepath.Join(dir, "identity", IdentityFile)
	cfg := NodeConfig{
		Version:      ConfigVersion,
		ConfigDir:    dir,
		ConfigFile:   cfgPath,
		IdentityFile: idPath,
		AuditDir:     filepath.Join(dir, "audit"),
		LogLevel:     LogLevelInfo,
	}
	data, err := json.MarshalIndent(cfg, "", "\t")
	if err != nil {
		return err
	}
	if err := os.WriteFile(cfgPath, data, DefaultFilePermissions); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	id, err := GenerateIdentity()
	if err != nil {
		return fmt.Errorf("failed to generate identity: %w", err)
	}
	return id.Write(force)
}
