package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type VaultClient struct {
	addr        string
	token       string
	transitKey  string
	httpClient  *http.Client
	fallbackKey []byte // For non-Vault encryption
	useVault    bool
}

func NewVaultClient(addr, token, transitKey string) *VaultClient {
	// Check if Vault is disabled
	useVault := addr != "" && addr != "disabled" && !strings.HasPrefix(addr, "http://localhost")

	// Get fallback encryption key from environment
	var fallbackKey []byte
	if envKey := os.Getenv("ENCRYPTION_KEY"); envKey != "" {
		decodedKey, err := hex.DecodeString(envKey)
		if err == nil && len(decodedKey) == 32 {
			fallbackKey = decodedKey
		}
	}

	return &VaultClient{
		addr:        addr,
		token:       token,
		transitKey:  transitKey,
		fallbackKey: fallbackKey,
		useVault:    useVault,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// WrapDEK wraps a DEK using Vault's transit encryption or fallback encryption
func (v *VaultClient) WrapDEK(dek []byte) (string, error) {
	// Use fallback encryption if Vault is disabled
	if !v.useVault || v.fallbackKey != nil {
		return v.wrapDEKFallback(dek)
	}

	// Base64 encode the DEK for Vault
	dekBase64 := base64.StdEncoding.EncodeToString(dek)

	payload := map[string]interface{}{
		"plaintext": dekBase64,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Vault API path needs /v1/ prefix
	url := fmt.Sprintf("%s/v1/transit/encrypt/%s", v.addr, v.transitKey)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Vault-Token", v.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call Vault: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("vault returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Data struct {
			Ciphertext string `json:"ciphertext"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}
	return result.Data.Ciphertext, nil
}

// UnwrapDEK unwraps a DEK using Vault's transit decryption or fallback decryption
func (v *VaultClient) UnwrapDEK(ciphertext string) ([]byte, error) {
	// Use fallback decryption if Vault is disabled or if ciphertext is fallback format
	if !v.useVault || v.fallbackKey != nil || strings.HasPrefix(ciphertext, "fb:") {
		return v.unwrapDEKFallback(ciphertext)
	}

	payload := map[string]interface{}{
		"ciphertext": ciphertext,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Vault API path needs /v1/ prefix
	url := fmt.Sprintf("%s/v1/transit/decrypt/%s", v.addr, v.transitKey)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Vault-Token", v.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Vault: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("vault returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Data struct {
			Plaintext string `json:"plaintext"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Decode base64
	dek, err := base64.StdEncoding.DecodeString(result.Data.Plaintext)
	if err != nil {
		return nil, fmt.Errorf("failed to decode DEK: %w", err)
	}
	return dek, nil
}

// --- Fallback encryption methods (no Vault required) ---

// wrapDEKFallback encrypts DEK using AES-256-GCM with environment key
func (v *VaultClient) wrapDEKFallback(dek []byte) (string, error) {
	if len(v.fallbackKey) != 32 {
		return "", fmt.Errorf("fallback encryption key not configured (set ENCRYPTION_KEY environment variable)")
	}

	// Create AES cipher
	block, err := aes.NewCipher(v.fallbackKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt
	ciphertext := gcm.Seal(nonce, nonce, dek, nil)

	// Return with "fb:" prefix to indicate fallback encryption
	return "fb:" + base64.StdEncoding.EncodeToString(ciphertext), nil
}

// unwrapDEKFallback decrypts DEK using AES-256-GCM with environment key
func (v *VaultClient) unwrapDEKFallback(ciphertext string) ([]byte, error) {
	if len(v.fallbackKey) != 32 {
		return nil, fmt.Errorf("fallback encryption key not configured (set ENCRYPTION_KEY environment variable)")
	}

	// Remove "fb:" prefix if present
	ciphertext = strings.TrimPrefix(ciphertext, "fb:")

	// Decode base64
	encryptedData, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	// Create AES cipher
	block, err := aes.NewCipher(v.fallbackKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Extract nonce
	nonceSize := gcm.NonceSize()
	if len(encryptedData) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertextBytes := encryptedData[:nonceSize], encryptedData[nonceSize:]

	// Decrypt
	dek, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return dek, nil
}

func (v *VaultClient) Health() error {
	url := fmt.Sprintf("%s/v1/sys/health", v.addr)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call Vault: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("vault health check failed with status: %d", resp.StatusCode)
	}
	return nil
}
