package node

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/0xMeechie/Aranea/pkg/policy"
)

const (
	DefaultConfigFile      = "config.json"
	DefaultIdentityFile    = "identity.json"
	DefaultConfigDirectory = ".config/aranea"
	DefaultKeyDirectory    = "keys"
)

type InitConfig struct {
	Forced bool
	Path   string
}

// Load attempts to read the node config and identity files from disk.
// Returns both structs if the files exist and are valid.
// Returns an error if either file is missing or cannot be parsed.
func Load(config InitConfig) (*NodeConfig, *NodeIdentity, error) {
	var homeLocation string
	if config.Path == "" {
		slog.Info("config path not set.")
		homeLocation = os.Getenv("HOME")
		if homeLocation == "" {
			return nil, nil, fmt.Errorf("$HOME is not currently set. Please set it or use the -Path Flag")
		}
	} else {
		homeLocation = config.Path
	}

	configLocation := filepath.Join(homeLocation, DefaultConfigDirectory)
	cfgFilePath := filepath.Join(configLocation, DefaultConfigFile)
	idFilePath := filepath.Join(configLocation, DefaultIdentityFile)

	slog.Info(fmt.Sprintf("using config path %s ", configLocation))

	keyDir := filepath.Join(configLocation, DefaultKeyDirectory)

	cfg, err := ReadConfigFile(cfgFilePath)
	if err != nil {
		return nil, nil, fmt.Errorf("config file: %w", err)
	}
	cfg.keyDir = keyDir

	id, err := ReadIdentityFile(idFilePath)
	if err != nil {
		return nil, nil, fmt.Errorf("identity file: %w", err)
	}

	return cfg, id, nil
}

var ErrAlreadyInitialized = errors.New("node is already initialized")

// Init runs the full initialization process to generate a new identity and config.
// It calls Load first; if both files already exist and Forced is false, it returns ErrAlreadyInitialized.
// If Forced is true, it skips the check and regenerates everything.
func Init(config InitConfig) error {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	slog.Info("starting the node initializing process")

	if !config.Forced {
		_, _, loadErr := Load(config)
		if loadErr == nil {
			slog.Info("node is already initialized, skipping")
			return ErrAlreadyInitialized
		}
		slog.Debug("existing node files not found or invalid, proceeding with init")
	} else {
		slog.Info("forced init enabled, reinitializing node")
	}

	var homeLocation string
	if config.Path == "" {
		slog.Debug("path not set. Using the default path of ./.config/aranea")
		homeLocation = os.Getenv("HOME")
		if homeLocation == "" {
			return fmt.Errorf("$HOME is not currently set. Please set it or use the -Path Flag")
		}
	} else {
		homeLocation = config.Path
	}

	configLocation := filepath.Join(homeLocation, DefaultConfigDirectory)

	mdErr := os.MkdirAll(configLocation, 0o666)
	if mdErr != nil {
		return fmt.Errorf("error creating directory %s", mdErr)
	}

	keyDir := filepath.Join(configLocation, DefaultKeyDirectory)
	kp, err := policy.GenerateKeyPair(keyDir)
	if err != nil {
		return fmt.Errorf("couldn't generate keypair for node %s", err)
	}

	pubBytes, err := kp.PublicKeyBytes()
	if err != nil {
		return fmt.Errorf("error getting pub key bytes %s", err)
	}

	pub := base64.StdEncoding.EncodeToString(pubBytes)
	pubSha := sha256.Sum256(pubBytes)

	nodeName := hex.EncodeToString(pubSha[:])

	idf := NodeIdentityFactory{
		Version:   DefaultVersion,
		PublicKey: pub,
		CreatedAt: time.Now(),
		NodeName:  nodeName,
	}

	id := idf.newNodeIdentity(*kp)

	idFilePath := filepath.Join(configLocation, DefaultIdentityFile)
	err = id.WriteFile(idFilePath)
	if err != nil {
		return fmt.Errorf("error writing identity file %s", err)
	}

	slog.Info(fmt.Sprintf("successful created the identity file at location %s", idFilePath))

	cfgFilePath := filepath.Join(configLocation, DefaultConfigFile)
	cfgFileData := NodeConfig{
		Version:      DefaultVersion,
		IdentityFile: idFilePath,
		ConfigFile:   cfgFilePath,
		LogLevel:     "debug",
		keyDir:       keyDir,
	}

	err = cfgFileData.WriteFile()
	if err != nil {
		return fmt.Errorf("error writing config file %s", err)
	}

	slog.Info(fmt.Sprintf("successful created the config file at location %s", cfgFilePath))
	slog.Info(fmt.Sprintf("initializing node in the following directory %s", configLocation))

	return nil
}
