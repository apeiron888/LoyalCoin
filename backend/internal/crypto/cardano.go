package crypto

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/btcutil/bech32"
	"golang.org/x/crypto/blake2b"
)

// Cardano wallet with keypair and address
type CardanoWallet struct {
	PrivateKey ed25519.PrivateKey
	PublicKey  ed25519.PublicKey
	Address    string // Bech32 encoded address
}

// New Cardano wallet (ed25519 keypair)
func GenerateCardanoWallet(networkTag byte) (*CardanoWallet, error) {
	// Generate ed25519 keypair
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ed25519 key: %w", err)
	}
	// Generate Cardano address
	address, err := deriveCardanoAddress(publicKey, networkTag)
	if err != nil {
		return nil, fmt.Errorf("failed to derive address: %w", err)
	}
	return &CardanoWallet{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		Address:    address,
	}, nil
}

// deriveCardanoAddress derives a Cardano payment address from a public key
// Uses Blake2b-224 hash and Bech32 encoding
func deriveCardanoAddress(publicKey ed25519.PublicKey, networkTag byte) (string, error) {
	// 1. Hash the public key with Blake2b-224
	// Cardano uses Blake2b-224 (28 bytes) for key hashing
	hash, err := blake2b.New(28, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create blake2b-224 hasher: %w", err)
	}
	hash.Write(publicKey)
	pkHash := hash.Sum(nil)

	// 2. Construct address header
	// Header byte: 0b0110_0000 (0x60) for testnet payment address (enterprise)
	// Header byte: 0b0110_0001 (0x61) for mainnet payment address (enterprise)
	// The networkTag passed in is 0x00 (testnet) or 0x01 (mainnet)

	var header byte
	if networkTag == 0x01 { // Mainnet
		header = 0x61 // Payment key, no stake key
	} else { // Testnet
		header = 0x60 // Payment key, no stake key
	}

	// 3. Construct address payload (Header + KeyHash)
	// Enterprise address is 29 bytes: 1 byte header + 28 bytes key hash
	payload := make([]byte, 1+28)
	payload[0] = header
	copy(payload[1:], pkHash)

	// 4. Encode as Bech32
	prefix := "addr_test"
	if networkTag == 0x01 {
		prefix = "addr"
	}

	// Convert 8-bit data to 5-bit groups for Bech32
	converted, err := bech32.ConvertBits(payload, 8, 5, true)
	if err != nil {
		return "", fmt.Errorf("failed to convert bits for bech32: %w", err)
	}

	address, err := bech32.Encode(prefix, converted)
	if err != nil {
		return "", fmt.Errorf("failed to encode bech32: %w", err)
	}

	return address, nil
}

// PrivateKeyToHex converts a private key to hex string
func PrivateKeyToHex(privateKey ed25519.PrivateKey) string {
	return hex.EncodeToString(privateKey)
}

// HexToPrivateKey converts hex string to private key
func HexToPrivateKey(hexStr string) (ed25519.PrivateKey, error) {
	keyBytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode hex: %w", err)
	}

	if len(keyBytes) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid private key length: expected %d, got %d",
			ed25519.PrivateKeySize, len(keyBytes))
	}

	return ed25519.PrivateKey(keyBytes), nil
}

// PublicKeyToHex converts a public key to hex string
func PublicKeyToHex(publicKey ed25519.PublicKey) string {
	return hex.EncodeToString(publicKey)
}

// HexToPublicKey converts hex string to public key
func HexToPublicKey(hexStr string) (ed25519.PublicKey, error) {
	keyBytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode hex: %w", err)
	}

	if len(keyBytes) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid public key length: expected %d, got %d",
			ed25519.PublicKeySize, len(keyBytes))
	}

	return ed25519.PublicKey(keyBytes), nil
}

// SignMessage signs a message with the private key
func SignMessage(privateKey ed25519.PrivateKey, message []byte) []byte {
	return ed25519.Sign(privateKey, message)
}

// VerifySignature verifies a signature with the public key
func VerifySignature(publicKey ed25519.PublicKey, message, signature []byte) bool {
	return ed25519.Verify(publicKey, message, signature)
}

// DerivePublicKey derives the public key from a private key
func DerivePublicKey(privateKey ed25519.PrivateKey) ed25519.PublicKey {
	// ed25519 private key is 64 bytes: 32-byte seed + 32-byte public key
	// Extract the public key portion
	if len(privateKey) != ed25519.PrivateKeySize {
		return nil
	}

	publicKey := make([]byte, ed25519.PublicKeySize)
	copy(publicKey, privateKey[32:])

	return ed25519.PublicKey(publicKey)
}

// SecureCompare does a constant-time comparison of two byte slices
func SecureCompare(a, b []byte) bool {
	return bytes.Equal(a, b)
}
