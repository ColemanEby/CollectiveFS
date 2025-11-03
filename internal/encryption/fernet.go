package encryption

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"

	"github.com/fernet/fernet-go"
)

// FernetEncryption handles Fernet encryption/decryption, matching Python's cryptography.fernet behavior.
type FernetEncryption struct {
	key *fernet.Key
}

// NewFernetEncryption creates a new Fernet encryption instance from a key.
func NewFernetEncryption(key []byte) (*FernetEncryption, error) {
	// The Python implementation stores raw 32 bytes
	// Fernet-go expects a base64 URL-encoded key string
	// Convert raw bytes to base64 URL-encoded string
	keyBase64 := base64.URLEncoding.EncodeToString(key)

	fKey, err := fernet.DecodeKey(keyBase64)
	if err != nil {
		return nil, fmt.Errorf("invalid Fernet key: %w", err)
	}

	return &FernetEncryption{
		key: fKey,
	}, nil
}

// EncryptChunk encrypts a chunk file in place, matching Python's encryptChunk behavior.
func (f *FernetEncryption) EncryptChunk(filePath string) error {
	// Read original file
	original, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Encrypt using fernet-go
	encrypted, err := fernet.EncryptAndSign(original, f.key)
	if err != nil {
		return fmt.Errorf("failed to encrypt: %w", err)
	}

	// Write encrypted file back (in place)
	if err := os.WriteFile(filePath, encrypted, 0644); err != nil {
		return fmt.Errorf("failed to write encrypted file: %w", err)
	}

	return nil
}

// DecryptChunk decrypts a chunk file in place, matching Python's decryptChunk behavior.
func (f *FernetEncryption) DecryptChunk(filePath string) error {
	// Read encrypted file
	encrypted, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read encrypted file: %w", err)
	}

	// Decrypt (TTL = 0 means no expiration check, matching Python)
	decrypted := fernet.VerifyAndDecrypt(encrypted, 0, []*fernet.Key{f.key})
	if decrypted == nil {
		return fmt.Errorf("decryption failed")
	}

	// Write decrypted file back (in place)
	if err := os.WriteFile(filePath, decrypted, 0644); err != nil {
		return fmt.Errorf("failed to write decrypted file: %w", err)
	}

	return nil
}

// GenerateKey generates a new Fernet key, matching Python's Fernet.generate_key().
func GenerateKey() ([]byte, error) {
	// Generate 32 random bytes
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}

	// Return as raw bytes (matching Python's binary key format)
	return key, nil
}
