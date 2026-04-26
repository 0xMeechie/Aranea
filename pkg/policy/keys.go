package policy

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
)

type KeyPair struct {
	PrivateKey *ecdsa.PrivateKey
	PublicKey  *ecdsa.PublicKey
}

func (kp *KeyPair) PrivateKeyString() (string, error) {
	pbytes, err := kp.PrivateKey.Bytes()
	if err != nil {
		return "", fmt.Errorf("error getting private key string %s", err)
	}

	return string(pbytes), nil
}

func (kp *KeyPair) PrivateKeyBytes() ([]byte, error) {
	return kp.PrivateKey.Bytes()
}

func (kp *KeyPair) PublicKeyBytes() ([]byte, error) {
	return kp.PublicKey.Bytes()
}

func (kp *KeyPair) PublicKeyString() (string, error) {
	pbytes, err := kp.PublicKey.Bytes()
	if err != nil {
		return "", fmt.Errorf("error getting public key string %s", err)
	}

	return string(pbytes), nil
}

func GenerateKeyPair(keyDir string) (*KeyPair, error) {
	if err := os.MkdirAll(keyDir, 0o700); err != nil {
		return nil, fmt.Errorf("create key directory: %w", err)
	}
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate key: %w", err)
	}
	privDER, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return nil, fmt.Errorf("marshal private key: %w", err)
	}

	privPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: privDER})

	keyPath := filepath.Join(keyDir, "node.key")
	if err := os.WriteFile(keyPath, privPEM, 0o600); err != nil {
		return nil, fmt.Errorf("write key file: %w", err)
	}

	return &KeyPair{PrivateKey: priv, PublicKey: &priv.PublicKey}, nil
}

func LoadKeyPair(keyDir string) (*KeyPair, error) {
	privPath := filepath.Join(keyDir, "node.key")

	privPEM, err := os.ReadFile(privPath)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", privPath, err)
	}

	block, _ := pem.Decode(privPEM)
	if block == nil {
		return nil, fmt.Errorf("no PEM block in %s", privPath)
	}

	priv, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}

	return &KeyPair{PrivateKey: priv, PublicKey: &priv.PublicKey}, nil
}

func LoadOrGenerate(keyDir string) (*KeyPair, bool, error) {
	kp, err := LoadKeyPair(keyDir)
	if err == nil {
		return kp, false, nil
	}

	kp, err = GenerateKeyPair(keyDir)
	if err != nil {
		return nil, false, err
	}
	return kp, true, nil
}

func (kp *KeyPair) Sign(payload any) (string, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal payload: %w", err)
	}

	hash := sha256.Sum256(data)

	r, s, err := ecdsa.Sign(rand.Reader, kp.PrivateKey, hash[:])
	if err != nil {
		return "", fmt.Errorf("sign: %w", err)
	}

	sig, err := asn1.Marshal(struct{ R, S *big.Int }{r, s})
	if err != nil {
		return "", fmt.Errorf("encode signature: %w", err)
	}

	return fmt.Sprintf("%s:%s", hex.EncodeToString(hash[:]), hex.EncodeToString(sig)), nil
}
