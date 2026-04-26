package node

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/0xMeechie/Aranea/pkg/policy"
)

const (
	DefaultVersion = 1
)

type NodeIdentity struct {
	Version   uint      `json:"version"`
	PublicKey string    `json:"publicKey"`
	CreatedAt time.Time `json:"cratedAt"`
	NodeName  string    `json:"nodeName"`
	keyPairs  policy.KeyPair
}

type NodeIdentityFactory struct {
	Version   uint      `json:"version"`
	PublicKey string    `json:"publicKey"`
	CreatedAt time.Time `json:"cratedAt"`
	NodeName  string    `json:"nodeName"`
}

func (nif NodeIdentityFactory) newNodeIdentity(kp policy.KeyPair) NodeIdentity {
	return NodeIdentity{
		Version:   nif.Version,
		PublicKey: nif.PublicKey,
		CreatedAt: nif.CreatedAt,
		NodeName:  nif.NodeName,
		keyPairs:  kp,
	}
}

func ReadIdentityFile(path string) (*NodeIdentity, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading identity file: %w", err)
	}
	var id NodeIdentity
	if err := json.Unmarshal(data, &id); err != nil {
		return nil, fmt.Errorf("error parsing identity file: %w", err)
	}
	return &id, nil
}

func (ni *NodeIdentity) WriteFile(path string) error {
	idData, err := json.MarshalIndent(ni, "", " ")
	if err != nil {
		return fmt.Errorf("error marshalling id file %s", err)
	}

	err = os.WriteFile(path, idData, 0o666)
	if err != nil {
		return fmt.Errorf("error writing id file %s", err)
	}

	return nil
}

type NodeConfig struct {
	Version      uint   `json:"version"`
	ConfigFile   string `json:"configFile"`
	IdentityFile string `json:"identityFile"`
	LogLevel     string `json:"logLevel"`
	keyDir       string
}

func (nc NodeConfig) KeyDir() string {
	return nc.keyDir
}

func ReadConfigFile(path string) (*NodeConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}
	var cfg NodeConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}
	return &cfg, nil
}

func (ni *NodeConfig) WriteFile() error {
	cfgData, err := json.MarshalIndent(ni, "", " ")
	if err != nil {
		return fmt.Errorf("error marshalling config file %s", err)
	}

	err = os.WriteFile(ni.ConfigFile, cfgData, 0o666)
	if err != nil {
		return fmt.Errorf("error writing config file %s", err)
	}

	return nil
}
