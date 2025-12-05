package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/loyalcoin/backend/internal/api"
	"github.com/loyalcoin/backend/internal/auth"
	"github.com/loyalcoin/backend/internal/cardano"
	"github.com/loyalcoin/backend/internal/config"
	"github.com/loyalcoin/backend/internal/crypto"
	"github.com/loyalcoin/backend/internal/indexer"
	"github.com/loyalcoin/backend/internal/models"
	"github.com/loyalcoin/backend/internal/storage"
	"github.com/loyalcoin/backend/pkg/logger"
	middleware "github.com/loyalcoin/backend/pkg/middleware"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize logger
	logger.Init(cfg.LogLevel, cfg.LogFormat)
	logger.Info("Starting LoyalCoin Auth Service", map[string]interface{}{
		"env": cfg.Env,
	})

	// Generate JWT keys if they don't exist
	if _, err := os.Stat(cfg.JWTPrivateKeyPath); os.IsNotExist(err) {
		logger.Info("Generating new JWT RSA keys", nil)
		if err := os.MkdirAll("./keys", 0755); err != nil {
			logger.Error("Failed to create keys directory", err, nil)
			os.Exit(1)
		}
		if err := auth.GenerateRSAKeys(cfg.JWTPrivateKeyPath, cfg.JWTPublicKeyPath); err != nil {
			logger.Error("Failed to generate RSA keys", err, nil)
			os.Exit(1)
		}
		logger.Info("JWT keys generated successfully", nil)
	}

	// Connect to MongoDB
	db, err := storage.Connect(cfg)
	if err != nil {
		logger.Error("Failed to connect to MongoDB", err, nil)
		os.Exit(1)
	}
	defer db.Close()

	// Initialize JWT service
	jwtService, err := auth.NewJWTService(
		cfg.JWTPrivateKeyPath,
		cfg.JWTPublicKeyPath,
		cfg.JWTExpirationHours,
	)
	if err != nil {
		logger.Error("Failed to initialize JWT service", err, nil)
		os.Exit(1)
	}

	// Initialize Vault client
	vaultClient := crypto.NewVaultClient(
		cfg.VaultAddr,
		cfg.VaultToken,
		cfg.VaultTransitKey,
	)

	// Check Vault connectivity
	if err := vaultClient.Health(); err != nil {
		logger.Warn("Vault health check failed - wallet creation may fail", map[string]interface{}{
			"error": err.Error(),
		})
	} else {
		logger.Info("Successfully connected to Vault", nil)
	}

	// Initialize Wallet Service (for generating Cardano wallets)
	walletService := crypto.NewWalletService(vaultClient)

	utxoRepo := storage.NewUTXORepository(db)
	txLogRepo := storage.NewTxLogRepository(db)
	settlementRepo := storage.NewSettlementRepository(db)
	allocationRepo := storage.NewAllocationRepository(db)

	cardanoService := cardano.NewCardanoService(
		cfg,
		utxoRepo,
		txLogRepo,
		walletService,
	)

	// Initialize repositories
	userRepo := storage.NewUserRepository(db)

	// Initialize handlers
	authHandler := api.NewAuthHandler(userRepo, jwtService, cfg)
	walletHandler := api.NewWalletHandler(cardanoService, userRepo, txLogRepo)
	settlementHandler := api.NewSettlementHandler(settlementRepo, userRepo, cfg.ExchangeRateLCNETB)
	allocationHandler := api.NewAllocationHandler(allocationRepo, userRepo, cfg.ExchangeRateLCNETB)
	adminHandler := api.NewAdminHandler(
		settlementRepo,
		allocationRepo,
		userRepo,
		txLogRepo,
		cardanoService,
		cfg.GovernanceWalletAddress,
	)

	// Set Gin mode
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create router
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.RequestLoggerMiddleware())

	// Inject wallet service into context
	router.Use(func(c *gin.Context) {
		c.Set("wallet_service", walletService)
		c.Next()
	})

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "auth-service",
		})
	})

	// API v1 routes
	authGroup := router.Group("/api/v1/auth")
	authGroup.POST("/signup", authHandler.Signup)
	authGroup.POST("/login", authHandler.Login)

	authMiddleware := middleware.AuthMiddleware(jwtService)

	walletGroup := router.Group("/api/v1/wallet")
	walletGroup.Use(authMiddleware)
	walletGroup.GET("/balance", walletHandler.GetBalance)
	walletGroup.GET("/transactions", walletHandler.GetTransactions)

	lcnGroup := router.Group("/api/v1/lcn")
	lcnGroup.Use(authMiddleware)
	lcnGroup.POST("/issue", walletHandler.IssueLCN)
	lcnGroup.POST("/redeem", walletHandler.RedeemLCN)

	// Merchant settlement routes (MERCHANT role required)
	merchantGroup := router.Group("/api/v1/merchant")
	merchantGroup.Use(authMiddleware)
	merchantGroup.Use(middleware.RequireRole(models.RoleMerchant))
	merchantGroup.POST("/settlement/request", settlementHandler.RequestSettlement)
	merchantGroup.GET("/settlement/history", settlementHandler.GetSettlementHistory)
	merchantGroup.POST("/allocation/purchase", allocationHandler.RequestAllocation)
	merchantGroup.GET("/allocation/history", allocationHandler.GetAllocationHistory)

	// Admin routes (ADMIN role required)
	adminGroup := router.Group("/api/v1/admin")
	adminGroup.Use(authMiddleware)
	adminGroup.Use(middleware.RequireRole(models.RoleAdmin))
	adminGroup.POST("/allocation/approve", adminHandler.ApproveAllocation)
	adminGroup.POST("/settlement/approve", adminHandler.ApproveSettlement)
	adminGroup.GET("/allocation/pending", adminHandler.GetPendingAllocations)
	adminGroup.GET("/settlement/pending", adminHandler.GetPendingSettlements)
	adminGroup.GET("/reserve/status", adminHandler.GetReserveStatus)

	// Initialize and start indexer service
	indexerConfig := indexer.DefaultConfig()
	// Get blockfrost client using same config as cardanoService
	blockfrostClient := cardano.NewBlockfrostClient(cfg.BlockfrostProjectID, cfg.BlockfrostAPIURL)
	indexerService := indexer.NewService(
		indexerConfig,
		blockfrostClient,
		txLogRepo,
		userRepo,
	)
	indexerService.Start()
	defer indexerService.Stop()

	// Setup graceful shutdown
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Handler: router,
	}

	// Start server in background
	go func() {
		logger.Info("Server starting", map[string]interface{}{
			"address": srv.Addr,
		})
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Failed to start server", err, nil)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...", nil)

	// Graceful shutdown with 10 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", err, nil)
	}

	logger.Info("Server exited", nil)
}
