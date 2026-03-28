// Package security provides AES-256-GCM encryption for secrets at rest.
// The vault key is derived from /etc/machine-id (Linux) or a local fallback.
package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
)

// Vault encrypts and decrypts secrets using AES-256-GCM.
type Vault struct {
	key []byte // 32 bytes
}

// NewVaultFromMachineID derives a vault key from the host's machine-id.
// On Linux this is /etc/machine-id; on macOS it uses the hardware UUID via ioreg.
func NewVaultFromMachineID() (*Vault, error) {
	id, err := machineID()
	if err != nil {
		return nil, fmt.Errorf("vault: get machine id: %w", err)
	}
	// Stretch to 32 bytes via SHA-256
	hash := sha256.Sum256([]byte("ulak-vault-v1:" + id))
	return &Vault{key: hash[:]}, nil
}

// NewVaultFromKey creates a Vault from a raw base64-encoded 32-byte key.
func NewVaultFromKey(b64 string) (*Vault, error) {
	key, err := base64.StdEncoding.DecodeString(b64)
	if err != nil || len(key) != 32 {
		return nil, errors.New("vault: invalid key (must be base64-encoded 32 bytes)")
	}
	return &Vault{key: key}, nil
}

// Encrypt encrypts plaintext and returns a base64-encoded ciphertext (nonce+ciphertext).
func (v *Vault) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(v.key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ct := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ct), nil
}

// Decrypt decrypts a base64-encoded ciphertext produced by Encrypt.
func (v *Vault) Decrypt(b64ct string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(b64ct)
	if err != nil {
		return "", fmt.Errorf("vault: decode: %w", err)
	}
	block, err := aes.NewCipher(v.key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(data) < gcm.NonceSize() {
		return "", errors.New("vault: ciphertext too short")
	}
	nonce, ct := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return "", fmt.Errorf("vault: decrypt: %w", err)
	}
	return string(plain), nil
}

func machineID() (string, error) {
	// Linux
	if data, err := os.ReadFile("/etc/machine-id"); err == nil {
		return string(data), nil
	}
	// Fallback: ~/.ulak/.machine-id (created once)
	home, _ := os.UserHomeDir()
	path := home + "/.ulak/.machine-id"
	if data, err := os.ReadFile(path); err == nil {
		return string(data), nil
	}
	// Generate and persist a new one
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	id := base64.URLEncoding.EncodeToString(b)
	_ = os.MkdirAll(home+"/.ulak", 0700)
	_ = os.WriteFile(path, []byte(id), 0600)
	return id, nil
}
