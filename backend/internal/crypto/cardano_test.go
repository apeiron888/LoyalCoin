package crypto

import (
	"crypto/ed25519"
	"testing"
)

func TestGenerateCardanoWallet(t *testing.T) {
	// Test testnet wallet generation
	wallet, err := GenerateCardanoWallet(0x00)
	if err != nil {
		t.Fatalf("Failed to generate wallet: %v", err)
	}

	// Check private key
	if len(wallet.PrivateKey) != ed25519.PrivateKeySize {
		t.Errorf("Expected private key length %d, got %d", ed25519.PrivateKeySize, len(wallet.PrivateKey))
	}

	// Check public key
	if len(wallet.PublicKey) != ed25519.PublicKeySize {
		t.Errorf("Expected public key length %d, got %d", ed25519.PublicKeySize, len(wallet.PublicKey))
	}

	// Check address format
	if wallet.Address == "" {
		t.Error("Address should not be empty")
	}

	// Testnet address should start with "addr_test"
	if len(wallet.Address) < 9 || wallet.Address[:9] != "addr_test" {
		t.Errorf("Testnet address should start with 'addr_test', got: %s", wallet.Address)
	}
}

func TestPrivateKeyHexConversion(t *testing.T) {
	wallet, _ := GenerateCardanoWallet(0x00)

	// Convert to hex
	hexStr := PrivateKeyToHex(wallet.PrivateKey)
	if hexStr == "" {
		t.Error("Hex string should not be empty")
	}

	// Convert back
	privateKey, err := HexToPrivateKey(hexStr)
	if err != nil {
		t.Fatalf("Failed to convert hex to private key: %v", err)
	}

	// Verify match
	if !SecureCompare(wallet.PrivateKey, privateKey) {
		t.Error("Private key mismatch after hex conversion")
	}
}

func TestPublicKeyHexConversion(t *testing.T) {
	wallet, _ := GenerateCardanoWallet(0x00)

	hexStr := PublicKeyToHex(wallet.PublicKey)
	if hexStr == "" {
		t.Error("Hex string should not be empty")
	}

	publicKey, err := HexToPublicKey(hexStr)
	if err != nil {
		t.Fatalf("Failed to convert hex to public key: %v", err)
	}

	if !SecureCompare(wallet.PublicKey, publicKey) {
		t.Error("Public key mismatch after hex conversion")
	}
}

func TestSignAndVerify(t *testing.T) {
	wallet, _ := GenerateCardanoWallet(0x00)
	message := []byte("Hello, Cardano!")

	// Sign
	signature := SignMessage(wallet.PrivateKey, message)
	if len(signature) != ed25519.SignatureSize {
		t.Errorf("Expected signature length %d, got %d", ed25519.SignatureSize, len(signature))
	}

	// Verify with correct public key
	if !VerifySignature(wallet.PublicKey, message, signature) {
		t.Error("Signature verification failed with correct public key")
	}

	// Verify with wrong message
	wrongMessage := []byte("Wrong message")
	if VerifySignature(wallet.PublicKey, wrongMessage, signature) {
		t.Error("Signature should not verify with wrong message")
	}

	// Generate another wallet and try to verify
	wallet2, _ := GenerateCardanoWallet(0x00)
	if VerifySignature(wallet2.PublicKey, message, signature) {
		t.Error("Signature should not verify with wrong public key")
	}
}

func TestDerivePublicKey(t *testing.T) {
	wallet, _ := GenerateCardanoWallet(0x00)

	derivedPublic := DerivePublicKey(wallet.PrivateKey)

	if !SecureCompare(wallet.PublicKey, derivedPublic) {
		t.Error("Derived public key should match wallet public key")
	}
}

func TestMainnetAddress(t *testing.T) {
	wallet, err := GenerateCardanoWallet(0x01) // Mainnet
	if err != nil {
		t.Fatalf("Failed to generate mainnet wallet: %v", err)
	}

	// Mainnet address should start with "addr1"
	if len(wallet.Address) < 4 || wallet.Address[:4] != "addr" {
		t.Errorf("Mainnet address should start with 'addr', got: %s", wallet.Address)
	}
}
