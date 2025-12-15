package startup

import (
	"context"
	"time"

	"github.com/loyalcoin/backend/internal/crypto"
	"github.com/loyalcoin/backend/internal/storage"
	"github.com/loyalcoin/backend/pkg/logger"
)

// EnsureAdminWalletAccessible checks if the admin wallet can be decrypted.
// If it fails (e.g. Vault is missing but wallet is Vault-encrypted), it regenerates the wallet.
func EnsureAdminWalletAccessible(userRepo *storage.UserRepository, walletService *crypto.WalletService) {
	// Give the server a moment to start
	time.Sleep(2 * time.Second)

	logger.Info("Checking Admin Wallet accessibility...", nil)

	ctx := context.Background()
	admin, err := userRepo.GetMerchantByEmail(ctx, "admin@loyalcoin.com")
	if err != nil {
		logger.Error("Could not find admin user for wallet check", err, nil)
		return
	}

	// Try to decrypt the key
	_, err = walletService.DecryptPrivateKey(admin.Wallet.EncryptedPrivateKey)
	if err == nil {
		logger.Info("✅ Admin wallet is accessible", nil)
		return
	}

	logger.Warn("⚠️ Admin wallet is inaccessible (likely missing Vault). Initiating repair...", map[string]interface{}{
		"error": err.Error(),
	})

	// Generate new wallet using current config (should use Fallback encryption if ENCRYPTION_KEY is set)
	newWallet, err := walletService.CreateWallet("testnet")
	if err != nil {
		logger.Error("Failed to generate new admin wallet", err, nil)
		return
	}

	// Update admin record
	admin.Wallet.Address = newWallet.Address
	admin.Wallet.EncryptedPrivateKey = newWallet.EncryptedPrivKey
	admin.Wallet.PubKeyHex = newWallet.PubKeyHex
	admin.Wallet.CreatedAt = time.Now().UTC()

	if err := userRepo.UpdateMerchant(ctx, admin); err != nil {
		logger.Error("Failed to save new admin wallet", err, nil)
		return
	}

	logger.Info("✅ Admin wallet repaired successfully", map[string]interface{}{
		"new_address": newWallet.Address,
		"note":        "Old wallet funds are lost. Please fund this new wallet.",
	})
}
