package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/loyalcoin/backend/internal/auth"
	"github.com/loyalcoin/backend/internal/config"
	"github.com/loyalcoin/backend/internal/crypto"
	"github.com/loyalcoin/backend/internal/models"
	"github.com/loyalcoin/backend/internal/storage"
	"github.com/loyalcoin/backend/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func main() {
	email := flag.String("email", "", "Admin email")
	password := flag.String("password", "", "Admin password")
	flag.Parse()

	if *email == "" || *password == "" {
		fmt.Println("Usage: go run cmd/create-admin/main.go -email <email> -password <password>")
		os.Exit(1)
	}

	// Load config
	cfg := config.Load()

	// Initialize logger FIRST
	logger.Init(cfg.LogLevel, cfg.LogFormat)

	// Connect to MongoDB
	db, err := storage.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer db.Close()

	userRepo := storage.NewUserRepository(db)

	// Check if user exists
	existing, _ := userRepo.GetMerchantByEmail(context.Background(), *email)

	// Initialize Vault Client
	vaultClient := crypto.NewVaultClient(cfg.VaultAddr, cfg.VaultToken, cfg.VaultTransitKey)

	// Initialize Wallet Service
	walletService := crypto.NewWalletService(vaultClient)

	// Generate a real wallet for the admin
	walletResult, err := walletService.CreateWallet("testnet") // Use testnet for dev
	if err != nil {
		log.Fatalf("Failed to generate admin wallet: %v", err)
	}

	// Hash password (cost 10 is standard)
	hashedPassword, err := auth.HashPassword(*password, 10)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// Create admin user (stored as a Merchant with ADMIN role)
	admin := &models.Merchant{
		ID:            primitive.NewObjectID().Hex(),
		BusinessName:  "System Admin",
		Email:         *email,
		PasswordHash:  hashedPassword,
		Role:          models.RoleAdmin,
		Status:        models.StatusActive,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
		AllocationLCN: 0,
		BalanceLCN:    0,
		Wallet: models.Wallet{
			Address:             walletResult.Address,
			EncryptedPrivateKey: walletResult.EncryptedPrivKey,
			PubKeyHex:           walletResult.PubKeyHex,
			CreatedAt:           time.Now().UTC(),
		},
	}

	collection := db.GetCollection("merchants")

	// Check if user exists first to avoid duplicate key error if we re-run
	// If exists, update it
	if existing != nil {
		fmt.Printf("User %s exists, updating wallet...\n", *email)
		_, err = collection.UpdateOne(
			context.Background(),
			bson.M{"email": *email},
			bson.M{"$set": bson.M{
				"wallet": admin.Wallet,
				"role":   models.RoleAdmin,
			}},
		)
	} else {
		_, err = collection.InsertOne(context.Background(), admin)
	}

	if err != nil {
		log.Fatalf("Failed to create/update admin user: %v", err)
	}

	fmt.Printf("✅ Successfully created/updated admin user: %s\n", *email)
	fmt.Printf("   Wallet Address: %s\n", walletResult.Address)
	fmt.Println("You can now login to the Admin Portal with these credentials.")
}
