package node

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type identityFile struct {
	Version   int    `json:"version"`
	PublicKey string `json:"public_key"`
	CreatedAt int64  `json:"created_at"`
	NodeName  string `json:"node_name"`
}

// Identity holds the full keypair and node metadata in memory.
type Identity struct {
	Version    int
	PublicKey  string
	PrivateKey string
	CreatedAt  int64
	NodeName   string
}

func identityDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("HOME not set: %w", err)
	}
	return filepath.Join(home, DefaultConfigLocation, "identity"), nil
}

// GenerateIdentity creates a new Ed25519 keypair and derives a node name
// from the SHA-256 hash of the public key.
func GenerateIdentity() (*Identity, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	pubKey := base64.StdEncoding.EncodeToString(pub)
	privKey := base64.StdEncoding.EncodeToString(priv)

	h := sha256.Sum256(pub)
	nodeName := fmt.Sprintf("%x", h)

	return &Identity{
		Version:    IdentityVersion,
		PublicKey:  pubKey,
		PrivateKey: privKey,
		CreatedAt:  time.Now().Unix(),
		NodeName:   nodeName,
	}, nil
}

// Write saves the public identity file and private key to disk.
// If force is false and the files already exist, Write is a no-op.
func (id *Identity) Write(force bool) error {
	dir, err := identityDir()
	if err != nil {
		return err
	}

	idPath := filepath.Join(dir, IdentityFile)
	keyPath := filepath.Join(dir, "identity.key")

	if !force {
		if _, err := os.Stat(idPath); err == nil {
			return nil
		}
	}

	f := identityFile{
		Version:   id.Version,
		PublicKey: id.PublicKey,
		CreatedAt: id.CreatedAt,
		NodeName:  id.NodeName,
	}
	data, err := json.MarshalIndent(f, "", "\t")
	if err != nil {
		return err
	}

	if err := os.WriteFile(idPath, data, DefaultFilePermissions); err != nil {
		return err
	}
	return os.WriteFile(keyPath, []byte(id.PrivateKey), DefaultFilePermissions)
}

// LoadIdentity reads the identity file and private key from disk.
func LoadIdentity() (*Identity, error) {
	dir, err := identityDir()
	if err != nil {
		return nil, err
	}

	idPath := filepath.Join(dir, IdentityFile)
	keyPath := filepath.Join(dir, "identity.key")

	data, err := os.ReadFile(idPath)
	if err != nil {
		return nil, fmt.Errorf("could not read identity: %w", err)
	}

	var f identityFile
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("invalid identity file: %w", err)
	}

	privKey, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("could not read private key: %w", err)
	}

	return &Identity{
		Version:    f.Version,
		PublicKey:  f.PublicKey,
		PrivateKey: string(privKey),
		CreatedAt:  f.CreatedAt,
		NodeName:   f.NodeName,
	}, nil
}
