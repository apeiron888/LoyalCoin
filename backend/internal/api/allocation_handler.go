package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/loyalcoin/backend/internal/models"
	"github.com/loyalcoin/backend/internal/storage"
	"github.com/loyalcoin/backend/pkg/logger"
)

type AllocationHandler struct {
	allocationRepo *storage.AllocationRepository
	userRepo       *storage.UserRepository
	exchangeRate   float64 // LCN to ETB exchange rate
}

func NewAllocationHandler(
	allocationRepo *storage.AllocationRepository,
	userRepo *storage.UserRepository,
	exchangeRate float64,
) *AllocationHandler {
	return &AllocationHandler{
		allocationRepo: allocationRepo,
		userRepo:       userRepo,
		exchangeRate:   exchangeRate,
	}
}

// POST /api/v1/merchant/allocation/purchase
func (h *AllocationHandler) RequestAllocation(c *gin.Context) {
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

	var req struct {
		AmountLCN        uint64 `json:"amount_lcn" binding:"required,min=1"`
		PaymentMethod    string `json:"payment_method" binding:"required"`
		PaymentReference string `json:"payment_reference" binding:"required"`
		PaymentProofURL  string `json:"payment_proof_url"`
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

	// Calculate ETB cost
	amountETB := float64(req.AmountLCN) * h.exchangeRate

	allocation := &models.AllocationPurchase{
		MerchantID:       merchantID,
		AmountLCN:        req.AmountLCN,
		AmountETBPaid:    amountETB,
		PaymentMethod:    req.PaymentMethod,
		PaymentReference: req.PaymentReference,
		PaymentProofURL:  req.PaymentProofURL,
	}

	if err := h.allocationRepo.CreateAllocation(c.Request.Context(), allocation); err != nil {
		logger.Error("Failed to create allocation purchase", err, map[string]interface{}{
			"merchant_id": merchantID,
		})
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"code":    "500_ALLOCATION_FAILED",
			"message": "Failed to create allocation purchase request",
		})
		return
	}
	logger.Audit("ALLOCATION_REQUESTED", merchantID, map[string]interface{}{
		"allocation_id": allocation.ID,
		"amount_lcn":    req.AmountLCN,
		"amount_etb":    amountETB,
	})

	c.JSON(http.StatusCreated, gin.H{
		"status": "ok",
		"data": gin.H{
			"purchase_id": allocation.ID,
			"amount_lcn":  req.AmountLCN,
			"amount_etb":  amountETB,
			"status":      allocation.Status,
			"message":     "Payment verification required. Admin will review within 24 hours.",
		},
	})
}

// GET /api/v1/merchant/allocation/history
func (h *AllocationHandler) GetAllocationHistory(c *gin.Context) {
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
	var status *string
	if statusFilter != "" {
		status = &statusFilter
	}

	allocations, total, err := h.allocationRepo.GetAllocationsByMerchant(
		c.Request.Context(),
		merchantID,
		limit,
		offset,
		status,
	)
	if err != nil {
		logger.Error("Failed to get allocation history", err, map[string]interface{}{
			"merchant_id": merchantID,
		})
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"code":    "500_QUERY_FAILED",
			"message": "Failed to retrieve allocation history",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data": gin.H{
			"allocations": allocations,
			"total":       total,
			"limit":       limit,
			"offset":      offset,
		},
	})
}
