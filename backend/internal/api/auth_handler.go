package api

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/loyalcoin/backend/internal/auth"
	"github.com/loyalcoin/backend/internal/config"
	"github.com/loyalcoin/backend/internal/crypto"
	"github.com/loyalcoin/backend/internal/models"
	"github.com/loyalcoin/backend/internal/storage"
	"github.com/loyalcoin/backend/pkg/logger"
)

type AuthHandler struct {
	userRepo   *storage.UserRepository
	jwtService *auth.JWTService
	config     *config.Config
}

func NewAuthHandler(userRepo *storage.UserRepository, jwtService *auth.JWTService, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		userRepo:   userRepo,
		jwtService: jwtService,
		config:     cfg,
	}
}

type SignupRequest struct {
	Email        string      `json:"email" binding:"required,email"`
	Password     string      `json:"password" binding:"required"`
	Role         models.Role `json:"role" binding:"required"`
	BusinessName string      `json:"business_name"` // Required for MERCHANT
	Username     string      `json:"username"`      // Required for CUSTOMER
	Phone        string      `json:"phone"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Signup(c *gin.Context) {
	var req SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"code":    "400_INVALID_REQUEST",
			"message": "Invalid request body",
			"data":    err.Error(),
		})
		return
	}

	if err := auth.ValidatePassword(req.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"code":    "400_WEAK_PASSWORD",
			"message": err.Error(),
		})
		return
	}

	passwordHash, err := auth.HashPassword(req.Password, h.config.BcryptCost)
	if err != nil {
		logger.Error("Failed to hash password", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"code":    "500_INTERNAL_ERROR",
			"message": "Failed to process request",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Generate real Cardano wallet with envelope encryption
	walletService := c.MustGet("wallet_service").(*crypto.WalletService)
	walletResult, err := walletService.CreateWallet(h.config.CardanoNetwork)
	if err != nil {
		logger.Error("Failed to create wallet", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"code":    "500_WALLET_CREATION_FAILED",
			"message": "Failed to create wallet",
		})
		return
	}

	wallet := models.Wallet{
		Address:             walletResult.Address,
		EncryptedPrivateKey: walletResult.EncryptedPrivKey,
		PubKeyHex:           walletResult.PubKeyHex,
		CreatedAt:           time.Now().UTC(),
	}

	switch req.Role {
	case models.RoleMerchant:
		if req.BusinessName == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"code":    "400_BUSINESS_NAME_REQUIRED",
				"message": "Business name is required for merchants",
			})
			return
		}

		merchant := &models.Merchant{
			BusinessName:  req.BusinessName,
			Email:         req.Email,
			PasswordHash:  passwordHash,
			Role:          models.RoleMerchant,
			Wallet:        wallet,
			Status:        models.StatusPendingVerification,
			AllocationLCN: 0,
			BalanceLCN:    0,
		}

		if err := h.userRepo.CreateMerchant(ctx, merchant); err != nil {
			logger.Error("Failed to create merchant", err, map[string]interface{}{
				"email": req.Email,
			})
			c.JSON(http.StatusConflict, gin.H{
				"status":  "error",
				"code":    "409_CONFLICT",
				"message": err.Error(),
			})
			return
		}
		logger.Info("Merchant created", map[string]interface{}{
			"merchant_id": merchant.ID,
			"email":       merchant.Email,
		})

		// Generate JWT token for immediate login
		token, err := h.jwtService.GenerateToken(merchant.ID, merchant.Role, merchant.Wallet.Address)
		if err != nil {
			logger.Error("Failed to generate JWT after signup", err, nil)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"code":    "500_TOKEN_GENERATION_FAILED",
				"message": "Account created but login failed",
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"status": "ok",
			"data": gin.H{
				"token":      token,
				"expires_at": time.Now().Add(time.Duration(h.config.JWTExpirationHours) * time.Hour).Format(time.RFC3339),
				"user": gin.H{
					"id":             merchant.ID,
					"email":          merchant.Email,
					"business_name":  merchant.BusinessName,
					"role":           merchant.Role,
					"wallet_address": merchant.Wallet.Address,
					"status":         merchant.Status,
				},
			},
		})

	case models.RoleCustomer:
		if req.Username == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"code":    "400_USERNAME_REQUIRED",
				"message": "Username is required for customers",
			})
			return
		}

		customer := &models.Customer{
			Username:     req.Username,
			Email:        req.Email,
			Phone:        req.Phone,
			PasswordHash: passwordHash,
			Wallet:       wallet,
		}

		if err := h.userRepo.CreateCustomer(ctx, customer); err != nil {
			logger.Error("Failed to create customer", err, map[string]interface{}{
				"email": req.Email,
			})
			c.JSON(http.StatusConflict, gin.H{
				"status":  "error",
				"code":    "409_CONFLICT",
				"message": err.Error(),
			})
			return
		}
		logger.Info("Customer created", map[string]interface{}{
			"customer_id": customer.ID,
			"email":       customer.Email,
		})

		c.JSON(http.StatusCreated, gin.H{
			"status": "ok",
			"data": gin.H{
				"user_id":        customer.ID,
				"wallet_address": customer.Wallet.Address,
				"role":           models.RoleCustomer,
			},
		})

	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"code":    "400_INVALID_ROLE",
			"message": "Role must be MERCHANT or CUSTOMER",
		})
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"code":    "400_INVALID_REQUEST",
			"message": "Invalid request body",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// Merchant login
	merchant, merchantErr := h.userRepo.GetMerchantByEmail(ctx, req.Email)
	if merchantErr == nil {
		if !auth.CheckPasswordHash(req.Password, merchant.PasswordHash) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"code":    "401_INVALID_CREDENTIALS",
				"message": "Invalid email or password",
			})
			return
		}
		token, err := h.jwtService.GenerateToken(merchant.ID, merchant.Role, merchant.Wallet.Address)
		if err != nil {
			logger.Error("Failed to generate JWT", err, nil)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"code":    "500_INTERNAL_ERROR",
				"message": "Failed to generate token",
			})
			return
		}
		logger.Audit("USER_LOGIN", merchant.ID, map[string]interface{}{
			"role":  merchant.Role,
			"email": merchant.Email,
		})
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"data": gin.H{
				"token":      token,
				"expires_at": time.Now().Add(time.Duration(h.config.JWTExpirationHours) * time.Hour).Format(time.RFC3339),
				"user": gin.H{
					"id":             merchant.ID,
					"email":          merchant.Email,
					"role":           merchant.Role,
					"wallet_address": merchant.Wallet.Address,
				},
			},
		})
		return
	}

	// Customer login
	customer, customerErr := h.userRepo.GetCustomerByEmail(ctx, req.Email)
	if customerErr == nil {
		if !auth.CheckPasswordHash(req.Password, customer.PasswordHash) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"code":    "401_INVALID_CREDENTIALS",
				"message": "Invalid email or password",
			})
			return
		}
		token, err := h.jwtService.GenerateToken(customer.ID, models.RoleCustomer, customer.Wallet.Address)
		if err != nil {
			logger.Error("Failed to generate JWT", err, nil)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"code":    "500_INTERNAL_ERROR",
				"message": "Failed to generate token",
			})
			return
		}
		logger.Audit("USER_LOGIN", customer.ID, map[string]interface{}{
			"role":  models.RoleCustomer,
			"email": customer.Email,
		})
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"data": gin.H{
				"token":      token,
				"expires_at": time.Now().Add(time.Duration(h.config.JWTExpirationHours) * time.Hour).Format(time.RFC3339),
				"user": gin.H{
					"id":             customer.ID,
					"email":          customer.Email,
					"role":           models.RoleCustomer,
					"wallet_address": customer.Wallet.Address,
				},
			},
		})
		return
	}

	// User not found
	c.JSON(http.StatusUnauthorized, gin.H{
		"status":  "error",
		"code":    "401_INVALID_CREDENTIALS",
		"message": "Invalid email or password",
	})
}
