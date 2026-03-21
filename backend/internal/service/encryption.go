package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
)

// EncryptionService provides AES-256-GCM encryption for sensitive fields.
type EncryptionService struct {
	gcm cipher.AEAD
}

// NewEncryptionService creates an EncryptionService from a 32-byte AES-256 key.
func NewEncryptionService(key []byte) (*EncryptionService, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create AES cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM: %w", err)
	}
	return &EncryptionService{gcm: gcm}, nil
}

// Encrypt returns the plaintext encrypted with AES-256-GCM, prefixed with "enc:".
func (s *EncryptionService) Encrypt(plaintext string) (string, error) {
	nonce := make([]byte, s.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}
	ciphertext := s.gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return "enc:" + base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt handles both encrypted ("enc:" prefix) and plaintext values.
// Unencrypted values are returned as-is for backward compatibility.
func (s *EncryptionService) Decrypt(value string) (string, error) {
	if !strings.HasPrefix(value, "enc:") {
		return value, nil
	}
	data, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(value, "enc:"))
	if err != nil {
		return "", fmt.Errorf("decode ciphertext: %w", err)
	}
	nonceSize := s.gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	plaintext, err := s.gcm.Open(nil, data[:nonceSize], data[nonceSize:], nil)
	if err != nil {
		return "", fmt.Errorf("decrypt: %w", err)
	}
	return string(plaintext), nil
}
