package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
)

type EncryptedBlob struct {
	Version    int    `json:"version"`
	Algorithm  string `json:"algorithm"`
	Nonce      string `json:"nonce"`
	Ciphertext string `json:"ciphertext"`
	Tag        string `json:"tag"`
	DEKWrapped string `json:"dek_wrapped"`
}

func GenerateDEK() ([]byte, error) {
	dek := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, dek); err != nil {
		return nil, fmt.Errorf("failed to generate DEK: %w", err)
	}
	return dek, nil
}

// EncryptWithDEK encrypts data using AES-256-GCM with the provided DEK
func EncryptWithDEK(plaintext []byte, dek []byte) (*EncryptedBlob, error) {
	// Create AES cipher
	block, err := aes.NewCipher(dek)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	// GCM appends the tag to the ciphertext, split them
	tagSize := 16 // GCM tag is always 16 bytes
	if len(ciphertext) < tagSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	actualCiphertext := ciphertext[:len(ciphertext)-tagSize]
	tag := ciphertext[len(ciphertext)-tagSize:]

	blob := &EncryptedBlob{
		Version:    1,
		Algorithm:  "AES-256-GCM",
		Nonce:      hex.EncodeToString(nonce),
		Ciphertext: hex.EncodeToString(actualCiphertext),
		Tag:        hex.EncodeToString(tag),
	}

	return blob, nil
}

// DecryptWithDEK decrypts data using AES-256-GCM with the provided DEK
func DecryptWithDEK(blob *EncryptedBlob, dek []byte) ([]byte, error) {
	// Decode hex strings
	nonce, err := hex.DecodeString(blob.Nonce)
	if err != nil {
		return nil, fmt.Errorf("failed to decode nonce: %w", err)
	}

	ciphertext, err := hex.DecodeString(blob.Ciphertext)
	if err != nil {
		return nil, fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	tag, err := hex.DecodeString(blob.Tag)
	if err != nil {
		return nil, fmt.Errorf("failed to decode tag: %w", err)
	}

	// Combine ciphertext and tag for GCM
	combined := append(ciphertext, tag...)

	// Create AES cipher
	block, err := aes.NewCipher(dek)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, combined, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}
	return plaintext, nil
}

// EncryptedBlob to JSON string
func EncryptedBlobToJSON(blob *EncryptedBlob) (string, error) {
	jsonBytes, err := json.Marshal(blob)
	if err != nil {
		return "", fmt.Errorf("failed to marshal blob: %w", err)
	}
	return string(jsonBytes), nil
}

// JSON string to EncryptedBlob
func JSONToEncryptedBlob(jsonStr string) (*EncryptedBlob, error) {
	var blob EncryptedBlob
	if err := json.Unmarshal([]byte(jsonStr), &blob); err != nil {
		return nil, fmt.Errorf("failed to unmarshal blob: %w", err)
	}
	return &blob, nil
}

// ZeroBytes overwrites a byte slice with zeros (for sensitive data)
func ZeroBytes(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
