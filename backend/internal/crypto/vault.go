package crypto

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type VaultClient struct {
	addr       string
	token      string
	transitKey string
	httpClient *http.Client
}

func NewVaultClient(addr, token, transitKey string) *VaultClient {
	return &VaultClient{
		addr:       addr,
		token:      token,
		transitKey: transitKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// WrapDEK wraps a DEK using Vault's transit encryption
func (v *VaultClient) WrapDEK(dek []byte) (string, error) {
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

// UnwrapDEK unwraps a DEK using Vault's transit decryption
func (v *VaultClient) UnwrapDEK(ciphertext string) ([]byte, error) {
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
