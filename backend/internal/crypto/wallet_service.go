package crypto

import (
	"crypto/ed25519"
	"fmt"
)

type WalletService struct {
	vaultClient *VaultClient
}

func NewWalletService(vaultClient *VaultClient) *WalletService {
	return &WalletService{
		vaultClient: vaultClient,
	}
}

// CreateWallet generates a new Cardano wallet with encrypted private key
func (s *WalletService) CreateWallet(network string) (*WalletResult, error) {
	// Determine network tag
	var networkTag byte
	switch network {
	case "mainnet":
		networkTag = 0x01
	default: // testnet, preprod, etc
		networkTag = 0x00
	}

	// 1. Generate Cardano wallet
	wallet, err := GenerateCardanoWallet(networkTag)
	if err != nil {
		return nil, fmt.Errorf("failed to generate wallet: %w", err)
	}

	// 2. Generate DEK (Data Encryption Key)
	dek, err := GenerateDEK()
	if err != nil {
		return nil, fmt.Errorf("failed to generate DEK: %w", err)
	}
	defer ZeroBytes(dek)

	// 3. Encrypt private key with DEK
	privKeyHex := PrivateKeyToHex(wallet.PrivateKey)
	encryptedBlob, err := EncryptWithDEK([]byte(privKeyHex), dek)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt private key: %w", err)
	}

	// 4. Wrap DEK with Vault
	wrappedDEK, err := s.vaultClient.WrapDEK(dek)
	if err != nil {
		return nil, fmt.Errorf("failed to wrap DEK: %w", err)
	}
	encryptedBlob.DEKWrapped = wrappedDEK

	// 5. Convert to JSON
	encryptedKeyJSON, err := EncryptedBlobToJSON(encryptedBlob)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize encrypted key: %w", err)
	}

	return &WalletResult{
		Address:          wallet.Address,
		EncryptedPrivKey: encryptedKeyJSON,
		WrappedDEK:       wrappedDEK,
		PubKeyHex:        PublicKeyToHex(wallet.PublicKey),
	}, nil
}

// DecryptPrivateKey decrypts an encrypted private key
func (s *WalletService) DecryptPrivateKey(encryptedKeyJSON string) (ed25519.PrivateKey, error) {
	// 1. Parse encrypted blob
	blob, err := JSONToEncryptedBlob(encryptedKeyJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to parse encrypted key: %w", err)
	}

	// 2. Unwrap DEK from Vault
	dek, err := s.vaultClient.UnwrapDEK(blob.DEKWrapped)
	if err != nil {
		return nil, fmt.Errorf("failed to unwrap DEK: %w", err)
	}
	defer ZeroBytes(dek) // Clear DEK from memory

	// 3. Decrypt private key with DEK
	privateKeyHexBytes, err := DecryptWithDEK(blob, dek)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt private key: %w", err)
	}
	defer ZeroBytes(privateKeyHexBytes) // Clear plaintext from memory

	// 4. Convert hex to private key
	privateKey, err := HexToPrivateKey(string(privateKeyHexBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return privateKey, nil
}

// SignTransaction signs a transaction with the decrypted private key
func (s *WalletService) SignTransaction(encryptedKeyJSON string, txBody []byte) ([]byte, error) {
	// 1. Decrypt private key
	privateKey, err := s.DecryptPrivateKey(encryptedKeyJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt key: %w", err)
	}
	defer ZeroBytes(privateKey) // Clear from memory

	// 2. Sign transaction
	signature := SignMessage(privateKey, txBody)

	return signature, nil
}
