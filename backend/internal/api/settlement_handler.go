package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/loyalcoin/backend/internal/models"
	"github.com/loyalcoin/backend/internal/storage"
	"github.com/loyalcoin/backend/pkg/logger"
)

type SettlementHandler struct {
	settlementRepo *storage.SettlementRepository
	userRepo       *storage.UserRepository
	exchangeRate   float64 // LCN to ETB exchange rate
}

func NewSettlementHandler(
	settlementRepo *storage.SettlementRepository,
	userRepo *storage.UserRepository,
	exchangeRate float64,
) *SettlementHandler {
	return &SettlementHandler{
		settlementRepo: settlementRepo,
		userRepo:       userRepo,
		exchangeRate:   exchangeRate,
	}
}

// POST /api/v1/merchant/settlement/request
func (h *SettlementHandler) RequestSettlement(c *gin.Context) {
	// Get user ID from JWT claims
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"code":    "401_UNAUTHORIZED",
			"message": "User not authenticated",
		})
		return
	}
	// Verify user is a merchant
	merchantID := userID.(string)
	merchant, err := h.userRepo.GetMerchantByID(c.Request.Context(), merchantID)
	if err != nil {
		logger.Error("Failed to get merchant", err, map[string]interface{}{
			"merchant_id": merchantID,
		})
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"code":    "404_MERCHANT_NOT_FOUND",
			"message": "Merchant not found",
		})
		return
	}
	var req struct {
		AmountLCN   uint64             `json:"amount_lcn" binding:"required,min=1"`
		BankAccount models.BankAccount `json:"bank_account" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"code":    "400_INVALID_REQUEST",
			"message": "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Verify merchant has sufficient balance
	if merchant.BalanceLCN < req.AmountLCN {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"code":    "400_INSUFFICIENT_BALANCE",
			"message": "Insufficient LCN balance",
			"data": gin.H{
				"available": merchant.BalanceLCN,
				"requested": req.AmountLCN,
			},
		})
		return
	}

	// Calculate ETB amount
	amountLCNFloat := float64(req.AmountLCN) / 1000.0
	amountETB := amountLCNFloat * h.exchangeRate

	// Create settlement request
	settlement := &models.SettlementRequest{
		MerchantID:   merchantID,
		AmountLCN:    req.AmountLCN,
		AmountETB:    amountETB,
		ExchangeRate: h.exchangeRate,
		BankAccount:  req.BankAccount,
	}

	if err := h.settlementRepo.CreateSettlement(c.Request.Context(), settlement); err != nil {
		logger.Error("Failed to create settlement", err, map[string]interface{}{
			"merchant_id": merchantID,
		})
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"code":    "500_SETTLEMENT_FAILED",
			"message": "Failed to create settlement request",
		})
		return
	}
	logger.Audit("SETTLEMENT_REQUESTED", merchantID, map[string]interface{}{
		"settlement_id": settlement.ID,
		"amount_lcn":    req.AmountLCN,
		"amount_etb":    amountETB,
	})

	c.JSON(http.StatusCreated, gin.H{
		"status": "ok",
		"data": gin.H{
			"settlement_id":             settlement.ID,
			"amount_lcn":                req.AmountLCN,
			"amount_etb":                amountETB,
			"exchange_rate":             h.exchangeRate,
			"status":                    settlement.Status,
			"estimated_processing_time": "24-48 hours",
		},
	})
}

// GET /api/v1/merchant/settlement/history
func (h *SettlementHandler) GetSettlementHistory(c *gin.Context) {
	// Get user ID from JWT claims
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"code":    "401_UNAUTHORIZED",
			"message": "User not authenticated",
		})
		return
	}

	merchantID := userID.(string)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	statusFilter := c.Query("status")

	if limit > 100 {
		limit = 100
	}
	if limit < 1 {
		limit = 20
	}

	var status *models.SettlementStatus
	if statusFilter != "" {
		s := models.SettlementStatus(statusFilter)
		status = &s
	}

	settlements, total, err := h.settlementRepo.GetSettlementsByMerchant(
		c.Request.Context(),
		merchantID,
		limit,
		offset,
		status,
	)
	if err != nil {
		logger.Error("Failed to get settlement history", err, map[string]interface{}{
			"merchant_id": merchantID,
		})
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"code":    "500_QUERY_FAILED",
			"message": "Failed to retrieve settlement history",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data": gin.H{
			"settlements": settlements,
			"total":       total,
			"limit":       limit,
			"offset":      offset,
		},
	})
}
