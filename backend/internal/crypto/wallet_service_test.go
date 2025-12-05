package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWalletService_CreateWallet(t *testing.T) {
	// This test requires a running Vault instance
	// Skip if Vault is not available
	vaultAddr := "http://localhost:8200"
	vaultToken := "dev-only-token"
	transitKey := "lcn-transit-key"

	vaultClient := NewVaultClient(vaultAddr, vaultToken, transitKey)

	walletService := NewWalletService(vaultClient)

	result, err := walletService.CreateWallet("testnet")
	require.NoError(t, err)
	assert.NotEmpty(t, result.Address)
	assert.NotEmpty(t, result.EncryptedPrivKey)
	assert.NotEmpty(t, result.PubKeyHex)

	// Verify address format (testnet should start with addr_test)
	assert.Contains(t, result.Address, "addr_test")
}

func TestWalletService_DecryptPrivateKey(t *testing.T) {
	vaultAddr := "http://localhost:8200"
	vaultToken := "dev-only-token"
	transitKey := "lcn-transit-key"

	vaultClient := NewVaultClient(vaultAddr, vaultToken, transitKey)

	walletService := NewWalletService(vaultClient)

	// Create a wallet first
	result, err := walletService.CreateWallet("testnet")
	require.NoError(t, err)

	// Decrypt the private key
	privKey, err := walletService.DecryptPrivateKey(result.EncryptedPrivKey)
	require.NoError(t, err)
	assert.NotEmpty(t, privKey)

	// Verify it's a valid hex string (128 chars for Ed25519)
	assert.Equal(t, 128, len(privKey))
}
