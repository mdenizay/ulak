// Package server handles SSH key generation and remote execution.
package server

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/crypto/ssh"
)

// GenerateDeployKey creates an Ed25519 SSH key pair for a project.
// Keys are written to dir/id_ed25519 and dir/id_ed25519.pub with 0600 permissions.
// Returns the public key string suitable for adding to GitHub deploy keys.
func GenerateDeployKey(dir string, comment string) (pubKey string, err error) {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("keygen: mkdir: %w", err)
	}

	pubEd, privEd, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", fmt.Errorf("keygen: generate: %w", err)
	}

	// Encode private key as OpenSSH PEM
	privPEM, err := ssh.MarshalPrivateKey(privEd, comment)
	if err != nil {
		return "", fmt.Errorf("keygen: marshal private: %w", err)
	}

	privPath := filepath.Join(dir, "id_ed25519")
	if err := os.WriteFile(privPath, pem.EncodeToMemory(privPEM), 0600); err != nil {
		return "", fmt.Errorf("keygen: write private key: %w", err)
	}

	sshPub, err := ssh.NewPublicKey(pubEd)
	if err != nil {
		return "", fmt.Errorf("keygen: marshal public: %w", err)
	}
	pubStr := string(ssh.MarshalAuthorizedKey(sshPub))

	pubPath := filepath.Join(dir, "id_ed25519.pub")
	if err := os.WriteFile(pubPath, []byte(pubStr), 0600); err != nil {
		return "", fmt.Errorf("keygen: write public key: %w", err)
	}

	return pubStr, nil
}
