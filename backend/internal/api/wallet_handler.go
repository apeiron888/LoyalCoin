package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/loyalcoin/backend/internal/cardano"
	"github.com/loyalcoin/backend/internal/models"
	"github.com/loyalcoin/backend/internal/storage"
	"github.com/loyalcoin/backend/pkg/logger"
)

type WalletHandler struct {
	cardanoService *cardano.CardanoService
	userRepo       *storage.UserRepository
	txLogRepo      *storage.TxLogRepository
}

func NewWalletHandler(
	cardanoService *cardano.CardanoService,
	userRepo *storage.UserRepository,
	txLogRepo *storage.TxLogRepository,
) *WalletHandler {
	return &WalletHandler{
		cardanoService: cardanoService,
		userRepo:       userRepo,
		txLogRepo:      txLogRepo,
	}
}

// GET /api/v1/wallet/balance
func (h *WalletHandler) GetBalance(c *gin.Context) {
	// Get wallet address from JWT claims
	walletAddress := c.GetString("wallet_address")
	userID := c.GetString("user_id")

	logger.Info("Fetching wallet balance", map[string]interface{}{
		"user_id": userID,
		"address": walletAddress,
	})

	// Get balance from Cardano service
	balance, err := h.cardanoService.GetBalance(walletAddress)
	if err != nil {
		logger.Error("Failed to get balance", err, map[string]interface{}{
			"user_id": userID,
		})
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"code":    "500_BALANCE_FETCH_FAILED",
			"message": "Failed to fetch balance",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data": gin.H{
			"address":      balance.Address,
			"ada":          balance.ADA,
			"lovelace":     balance.Lovelace,
			"lcn":          balance.LCN,
			"lcn_atomic":   balance.LCNAtomic,
			"other_assets": balance.OtherAssets,
		},
	})
}

// GET /api/v1/wallet/transactions
func (h *WalletHandler) GetTransactions(c *gin.Context) {
	walletAddress := c.GetString("wallet_address")
	userID := c.GetString("user_id")
	limit := 20
	offset := 0

	if limitParam := c.Query("limit"); limitParam != "" {
		if l, err := strconv.Atoi(limitParam); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	if offsetParam := c.Query("offset"); offsetParam != "" {
		if o, err := strconv.Atoi(offsetParam); err == nil && o >= 0 {
			offset = o
		}
	}
	logger.Info("Fetching transactions", map[string]interface{}{
		"user_id": userID,
		"limit":   limit,
		"offset":  offset,
	})
	txLogs, err := h.txLogRepo.GetTxLogsByAddress(c.Request.Context(), walletAddress, limit, offset)
	if err != nil {
		logger.Error("Failed to get transactions", err, map[string]interface{}{
			"user_id": userID,
		})
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"code":    "500_TX_FETCH_FAILED",
			"message": "Failed to fetch transactions",
		})
		return
	}

	transactions := make([]gin.H, 0, len(txLogs))
	for _, tx := range txLogs {
		direction := "sent"
		if tx.ToAddress == walletAddress {
			direction = "received"
		}
		txData := gin.H{
			"tx_hash":      tx.TxHash,
			"type":         tx.Type,
			"direction":    direction,
			"from_address": tx.FromAddress,
			"to_address":   tx.ToAddress,
			"amount_lcn":   float64(tx.AmountLCN) / 1000, // Convert to LCN
			"status":       tx.Status,
			"submitted_at": tx.SubmittedAt,
		}
		if tx.ConfirmedAt != nil {
			txData["confirmed_at"] = *tx.ConfirmedAt
		}
		if tx.BlockHeight > 0 {
			txData["block_height"] = tx.BlockHeight
		}
		transactions = append(transactions, txData)
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data": gin.H{
			"transactions": transactions,
			"count":        len(transactions),
			"limit":        limit,
			"offset":       offset,
		},
	})
}

// POST /api/v1/lcn/issue (Merchant only)
func (h *WalletHandler) IssueLCN(c *gin.Context) {
	userID := c.GetString("user_id")
	roleValue, _ := c.Get("role")
	role := roleValue.(models.Role)

	if role != models.RoleMerchant {
		c.JSON(http.StatusForbidden, gin.H{
			"status":  "error",
			"code":    "403_MERCHANT_ONLY",
			"message": "Only merchants can issue LCN",
		})
		return
	}
	var req struct {
		CustomerAddress string  `json:"customer_address" binding:"required"`
		AmountLCN       float64 `json:"amount_lcn" binding:"required,gt=0"`
		Reference       string  `json:"reference"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"code":    "400_INVALID_REQUEST",
			"message": "Invalid request body",
		})
		return
	}
	logger.Audit("LCN_ISSUANCE_INITIATED", userID, map[string]interface{}{
		"customer_address": req.CustomerAddress,
		"amount_lcn":       req.AmountLCN,
	})

	ctx := c.Request.Context()
	merchant, err := h.userRepo.GetMerchantByID(ctx, userID)
	if err != nil {
		logger.Error("Merchant not found", err, map[string]interface{}{
			"user_id": userID,
		})
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"code":    "404_MERCHANT_NOT_FOUND",
			"message": "Merchant not found",
		})
		return
	}

	// Check actual wallet balance instead of database allocation
	balance, err := h.cardanoService.GetBalance(merchant.Wallet.Address)
	if err != nil {
		logger.Error("Failed to get merchant balance", err, map[string]interface{}{
			"user_id": userID,
		})
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"code":    "500_BALANCE_CHECK_FAILED",
			"message": "Failed to verify merchant balance",
		})
		return
	}

	// Check if merchant has enough LCN (based on actual blockchain balance)
	if balance.LCN < req.AmountLCN {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"code":    "400_INSUFFICIENT_BALANCE",
			"message": "Insufficient LCN balance",
			"data": gin.H{
				"requested": req.AmountLCN,
				"available": balance.LCN,
			},
		})
		return
	}

	// Transfer LCN (TransferADA expects whole LCN and converts to lovelace internally)
	txHash, err := h.cardanoService.TransferADA(
		merchant.Wallet.Address,
		req.CustomerAddress,
		uint64(req.AmountLCN),
		merchant.Wallet.EncryptedPrivateKey,
	)
	if err != nil {
		logger.Error("Failed to issue LCN", err, map[string]interface{}{
			"merchant_id": userID,
		})
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"code":    "500_ISSUANCE_FAILED",
			"message": "Failed to issue LCN: " + err.Error(),
		})
		return
	}
	logger.Info("LCN issued successfully", map[string]interface{}{
		"merchant_id": userID,
		"tx_hash":     txHash,
		"amount_lcn":  req.AmountLCN,
	})
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data": gin.H{
			"tx_hash":           txHash,
			"amount_lcn":        req.AmountLCN,
			"remaining_balance": balance.LCN - req.AmountLCN,
		},
	})
}

// POST /api/v1/lcn/redeem (Customer only)
func (h *WalletHandler) RedeemLCN(c *gin.Context) {
	userID := c.GetString("user_id")
	roleValue, _ := c.Get("role")
	role := roleValue.(models.Role)

	// Verify customer role
	if role != models.RoleCustomer {
		c.JSON(http.StatusForbidden, gin.H{
			"status":  "error",
			"code":    "403_CUSTOMER_ONLY",
			"message": "Only customers can redeem LCN",
		})
		return
	}
	var req struct {
		MerchantAddress string  `json:"merchant_address" binding:"required"`
		AmountLCN       float64 `json:"amount_lcn" binding:"required,gt=0"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"code":    "400_INVALID_REQUEST",
			"message": "Invalid request body",
		})
		return
	}
	logger.Audit("LCN_REDEMPTION_INITIATED", userID, map[string]interface{}{
		"merchant_address": req.MerchantAddress,
		"amount_lcn":       req.AmountLCN,
	})

	// Get customer
	ctx := c.Request.Context()
	customer, err := h.userRepo.GetCustomerByID(ctx, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"code":    "404_CUSTOMER_NOT_FOUND",
			"message": "Customer not found",
		})
		return
	}
	// Check customer balance
	balance, err := h.cardanoService.GetBalance(customer.Wallet.Address)
	if err != nil {
		logger.Error("Failed to get customer balance", err, map[string]interface{}{
			"customer_id": userID,
		})
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"code":    "500_BALANCE_CHECK_FAILED",
			"message": "Failed to verify balance",
		})
		return
	}
	// Check if customer has enough LCN
	if balance.LCN < req.AmountLCN {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"code":    "400_INSUFFICIENT_BALANCE",
			"message": "Insufficient LCN balance",
			"data": gin.H{
				"requested": req.AmountLCN,
				"available": balance.LCN,
			},
		})
		return
	}
	// Transfer LCN (TransferADA expects whole LCN and converts to lovelace internally)
	txHash, err := h.cardanoService.TransferADA(
		customer.Wallet.Address,
		req.MerchantAddress,
		uint64(req.AmountLCN), // Pass whole LCN, not atomic units
		customer.Wallet.EncryptedPrivateKey,
	)
	if err != nil {
		logger.Error("Failed to redeem LCN", err, map[string]interface{}{
			"customer_id": userID,
		})
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"code":    "500_REDEMPTION_FAILED",
			"message": "Failed to redeem LCN: " + err.Error(),
		})
		return
	}
	logger.Info("LCN redeemed successfully", map[string]interface{}{
		"customer_id": userID,
		"tx_hash":     txHash,
		"amount_lcn":  req.AmountLCN,
	})
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data": gin.H{
			"tx_hash":    txHash,
			"amount_lcn": req.AmountLCN,
		},
	})
}
