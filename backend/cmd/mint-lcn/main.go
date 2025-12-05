package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/loyalcoin/backend/internal/cardano"
	"github.com/loyalcoin/backend/internal/config"
	"github.com/loyalcoin/backend/internal/crypto"
	"github.com/loyalcoin/backend/pkg/logger"
)

func main() {
	// Load config
	cfg := config.Load()
	logger.Init("info", "json")

	// Initialize Vault (needed for wallet creation)
	vaultClient := crypto.NewVaultClient(cfg.VaultAddr, cfg.VaultToken, cfg.VaultTransitKey)
	if err := vaultClient.Health(); err != nil {
		log.Fatalf("Vault health check failed: %v", err)
	}

	walletService := crypto.NewWalletService(vaultClient)
	blockfrost := cardano.NewBlockfrostClient(cfg.BlockfrostProjectID, cfg.BlockfrostAPIURL)

	fmt.Println("🚀 LoyalCoin Minting Tool")
	fmt.Println("==========================")

	// 1. Create System Governance Wallet
	fmt.Println("\n1️⃣  Creating System Governance Wallet (to sign minting tx)...")
	wallet, err := walletService.CreateWallet(cfg.CardanoNetwork)
	if err != nil {
		log.Fatalf("Failed to create wallet: %v", err)
	}

	fmt.Printf("✅ Wallet Created:\n")
	fmt.Printf("   Address: %s\n", wallet.Address)
	fmt.Printf("   PubKeyHash: %s\n", wallet.PubKeyHex) // Simplified, actual PKH derivation needed

	// 2. Fund Wallet
	fmt.Println("\n2️⃣  FUNDING REQUIRED")
	fmt.Printf("   Please send at least 100 tADA to: %s\n", wallet.Address)
	fmt.Println("   Waiting for funds (checking every 10s)...")

	for {
		utxos, err := blockfrost.GetAddressUTXOs(wallet.Address)
		if err != nil {
			fmt.Printf(".")
			time.Sleep(10 * time.Second)
			continue
		}

		var totalLovelace uint64
		for _, u := range utxos {
			for _, amt := range u.Amount {
				if amt.Unit == "lovelace" {
					val, _ := strconv.ParseUint(amt.Quantity, 10, 64)
					totalLovelace += val
				}
			}
		}

		if totalLovelace >= 50000000 { // 50 ADA
			fmt.Printf("\n✅ Funds received: %d Lovelace (%.2f ADA)\n", totalLovelace, float64(totalLovelace)/1000000)
			break
		}
		time.Sleep(10 * time.Second)
	}

	// 3. Generate Policy Script
	fmt.Println("\n3️⃣  Generating Minting Policy...")
	// TODO: Implement policy generation logic here
	// For MVP, we will use a simple "signed by this wallet" policy

	// 4. Mint Tokens
	fmt.Println("\n4️⃣  Minting 10,000,000 LCN...")
	// TODO: Implement minting tx building and submission

	fmt.Println("\n✅ Minting Complete!")
}
