package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/loyalcoin/backend/internal/cardano"
	"github.com/loyalcoin/backend/internal/models"
	"github.com/loyalcoin/backend/internal/storage"
	"github.com/loyalcoin/backend/pkg/logger"
)

// Admin-related requests
type AdminHandler struct {
	settlementRepo *storage.SettlementRepository
	allocationRepo *storage.AllocationRepository
	userRepo       *storage.UserRepository
	txLogRepo      *storage.TxLogRepository
	cardanoService *cardano.CardanoService
	governanceAddr string
}

func NewAdminHandler(
	settlementRepo *storage.SettlementRepository,
	allocationRepo *storage.AllocationRepository,
	userRepo *storage.UserRepository,
	txLogRepo *storage.TxLogRepository,
	cardanoService *cardano.CardanoService,
	governanceAddr string,
) *AdminHandler {
	return &AdminHandler{
		settlementRepo: settlementRepo,
		allocationRepo: allocationRepo,
		userRepo:       userRepo,
		txLogRepo:      txLogRepo,
		cardanoService: cardanoService,
		governanceAddr: governanceAddr,
	}
}

// POST /api/v1/admin/allocation/approve
func (h *AdminHandler) ApproveAllocation(c *gin.Context) {
	// Get admin ID from JWT claims
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"code":    "401_UNAUTHORIZED",
			"message": "User not authenticated",
		})
		return
	}

	var req struct {
		PurchaseID string `json:"purchase_id" binding:"required"`
		Action     string `json:"action" binding:"required,oneof=APPROVE REJECT"`
		Notes      string `json:"notes"`
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

	// Get allocation purchase
	allocation, err := h.allocationRepo.GetAllocationByID(c.Request.Context(), req.PurchaseID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"code":    "404_ALLOCATION_NOT_FOUND",
			"message": "Allocation purchase not found",
		})
		return
	}

	// Verify status is PENDING
	if allocation.Status != "PENDING" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"code":    "400_INVALID_STATUS",
			"message": "Allocation already processed",
			"data": gin.H{
				"current_status": allocation.Status,
			},
		})
		return
	}

	// REJECT:
	if req.Action == "REJECT" {
		allocation.Status = "REJECTED"
		allocation.AdminID = adminID.(string)
		allocation.AdminNotes = req.Notes
		now := time.Now().UTC()
		allocation.VerifiedAt = &now

		if err := h.allocationRepo.UpdateAllocation(c.Request.Context(), allocation); err != nil {
			logger.Error("Failed to update allocation", err, nil)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"code":    "500_UPDATE_FAILED",
				"message": "Failed to reject allocation",
			})
			return
		}

		logger.Audit("ALLOCATION_REJECTED", adminID.(string), map[string]interface{}{
			"allocation_id": allocation.ID,
			"merchant_id":   allocation.MerchantID,
		})

		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"data": gin.H{
				"purchase_id": allocation.ID,
				"status":      "REJECTED",
			},
		})
		return
	}

	// APPROVE: Transfer LCN from governance to merchant
	merchant, err := h.userRepo.GetMerchantByID(c.Request.Context(), allocation.MerchantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"code":    "404_MERCHANT_NOT_FOUND",
			"message": "Merchant not found",
		})
		return
	}

	// Get governance wallet details (assuming it's stored as a special user or config)
	// For this implementation, we'll use the governance address from config and assume
	// we have access to its private key via the wallet service or environment
	// In a real production system, this would be a multi-sig or hardware wallet interaction

	// Retrieve governance wallet
	// This part depends on how the governance wallet is stored.
	// If it's a merchant/user with role ADMIN, we can find it.
	// Let's try to find the user associated with the governance address
	govUser, err := h.userRepo.GetMerchantByEmail(c.Request.Context(), "admin@loyalcoin.com")
	if err != nil {
		logger.Error("Failed to retrieve governance wallet owner", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"code":    "500_GOVERNANCE_WALLET_ERROR",
			"message": "Failed to access governance wallet",
		})
		return
	}

	// Perform the transfer (ADA-backed LCN)
	// allocation.AmountLCN is in whole LCN units
	// TransferADA will convert to lovelace (LCN × 10,000)

	txHash, err := h.cardanoService.TransferADA(
		govUser.Wallet.Address,
		merchant.Wallet.Address,
		allocation.AmountLCN, // whole LCN units
		govUser.Wallet.EncryptedPrivateKey,
	)
	if err != nil {
		logger.Error("Failed to transfer ADA (LCN)", err, map[string]interface{}{
			"from":   govUser.Wallet.Address,
			"to":     merchant.Wallet.Address,
			"amount": allocation.AmountLCN,
		})
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"code":    "500_TRANSFER_FAILED",
			"message": "Failed to transfer LCN: " + err.Error(),
		})
		return
	}

	// Update merchant allocation
	merchant.AllocationLCN += allocation.AmountLCN
	merchant.BalanceLCN += allocation.AmountLCN

	if err := h.userRepo.UpdateMerchant(c.Request.Context(), merchant); err != nil {
		logger.Error("Failed to update merchant allocation", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"code":    "500_UPDATE_FAILED",
			"message": "Failed to update merchant allocation",
		})
		return
	}

	// Update allocation status
	allocation.Status = "CONFIRMED"
	allocation.AdminID = adminID.(string)
	allocation.AdminNotes = req.Notes
	allocation.LCNTransferTxHash = txHash
	now := time.Now().UTC()
	allocation.VerifiedAt = &now

	if err := h.allocationRepo.UpdateAllocation(c.Request.Context(), allocation); err != nil {
		logger.Error("Failed to update allocation", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"code":    "500_UPDATE_FAILED",
			"message": "Failed to approve allocation",
		})
		return
	}

	logger.Audit("ALLOCATION_APPROVED", adminID.(string), map[string]interface{}{
		"allocation_id": allocation.ID,
		"merchant_id":   allocation.MerchantID,
		"amount_lcn":    allocation.AmountLCN,
	})

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data": gin.H{
			"purchase_id": allocation.ID,
			"status":      "CONFIRMED",
			"message":     "LCN allocated to merchant wallet",
		},
	})
}

// POST /api/v1/admin/settlement/approve
func (h *AdminHandler) ApproveSettlement(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"code":    "401_UNAUTHORIZED",
			"message": "User not authenticated",
		})
		return
	}

	var req struct {
		SettlementID     string `json:"settlement_id" binding:"required"`
		Action           string `json:"action" binding:"required,oneof=APPROVE REJECT"`
		PaymentReference string `json:"payment_reference"`
		Notes            string `json:"notes"`
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

	settlement, err := h.settlementRepo.GetSettlementByID(c.Request.Context(), req.SettlementID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"code":    "404_SETTLEMENT_NOT_FOUND",
			"message": "Settlement request not found",
		})
		return
	}

	if settlement.Status != models.SettlementPending && settlement.Status != models.SettlementApproved {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"code":    "400_INVALID_STATUS",
			"message": "Settlement already processed",
			"data": gin.H{
				"current_status": settlement.Status,
			},
		})
		return
	}

	if req.Action == "REJECT" {
		settlement.Status = models.SettlementRejected
		settlement.AdminID = adminID.(string)
		settlement.AdminNotes = req.Notes
		now := time.Now().UTC()
		settlement.ApprovedAt = &now

		if err := h.settlementRepo.UpdateSettlement(c.Request.Context(), settlement); err != nil {
			logger.Error("Failed to update settlement", err, nil)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"code":    "500_UPDATE_FAILED",
				"message": "Failed to reject settlement",
			})
			return
		}
		logger.Audit("SETTLEMENT_REJECTED", adminID.(string), map[string]interface{}{
			"settlement_id": settlement.ID,
			"merchant_id":   settlement.MerchantID,
		})
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"data": gin.H{
				"settlement_id": settlement.ID,
				"status":        "REJECTED",
			},
		})
		return
	}

	// APPROVE: Transfer tADA from merchant to admin (governance wallet)
	merchant, err := h.userRepo.GetMerchantByID(c.Request.Context(), settlement.MerchantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"code":    "404_MERCHANT_NOT_FOUND",
			"message": "Merchant not found",
		})
		return
	}

	// Get governance wallet (admin wallet to receive funds)
	govUser, err := h.userRepo.GetMerchantByEmail(c.Request.Context(), "admin@loyalcoin.com")
	if err != nil {
		logger.Error("Failed to retrieve governance wallet", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"code":    "500_GOVERNANCE_WALLET_ERROR",
			"message": "Failed to access governance wallet",
		})
		return
	}

	// Transfer tADA from merchant to admin (governance) wallet
	txHash, err := h.cardanoService.TransferADA(
		merchant.Wallet.Address,
		govUser.Wallet.Address,
		settlement.AmountLCN,
		merchant.Wallet.EncryptedPrivateKey,
	)
	if err != nil {
		logger.Error("Failed to transfer tADA for settlement", err, map[string]interface{}{
			"from":       merchant.Wallet.Address,
			"to":         govUser.Wallet.Address,
			"amount_lcn": settlement.AmountLCN,
		})
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"code":    "500_TRANSFER_FAILED",
			"message": "Failed to transfer ADA: " + err.Error(),
		})
		return
	}

	settlement.Status = models.SettlementCompleted
	settlement.AdminID = adminID.(string)
	settlement.AdminNotes = req.Notes
	settlement.PaymentReference = req.PaymentReference
	settlement.TxHash = txHash
	now := time.Now().UTC()
	settlement.ApprovedAt = &now
	settlement.ProcessedAt = &now

	if err := h.settlementRepo.UpdateSettlement(c.Request.Context(), settlement); err != nil {
		logger.Error("Failed to update settlement", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"code":    "500_UPDATE_FAILED",
			"message": "Failed to complete settlement",
		})
		return
	}
	logger.Audit("SETTLEMENT_APPROVED", adminID.(string), map[string]interface{}{
		"settlement_id":     settlement.ID,
		"merchant_id":       settlement.MerchantID,
		"amount_lcn":        settlement.AmountLCN,
		"amount_etb":        settlement.AmountETB,
		"tx_hash":           txHash,
		"payment_reference": req.PaymentReference,
	})
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data": gin.H{
			"settlement_id":     settlement.ID,
			"status":            "COMPLETED",
			"tx_hash":           txHash,
			"payment_reference": req.PaymentReference,
			"message":           "Settlement completed. tADA transferred from merchant to admin wallet.",
		},
	})
}

// GET /api/v1/admin/reserve/status
func (h *AdminHandler) GetReserveStatus(c *gin.Context) {
	govUser, err := h.userRepo.GetMerchantByEmail(c.Request.Context(), "admin@loyalcoin.com")
	if err != nil {
		logger.Error("Failed to retrieve governance wallet owner", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"code":    "500_GOVERNANCE_WALLET_ERROR",
			"message": "Failed to access governance wallet",
		})
		return
	}

	govBalance, err := h.cardanoService.GetBalance(govUser.Wallet.Address)
	if err != nil {
		logger.Error("Failed to get governance balance", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"code":    "500_BALANCE_QUERY_FAILED",
			"message": "Failed to query governance balance",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data": gin.H{
			"lcn_balance":               govBalance.LCN,
			"lcn_balance_atomic":        govBalance.LCNAtomic,
			"governance_wallet_address": govBalance.Address,
			"health":                    "ACTIVE",
		},
	})
}

// GET /api/v1/admin/allocation/pending
func (h *AdminHandler) GetPendingAllocations(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	allocations, total, err := h.allocationRepo.GetAllPendingAllocations(c.Request.Context(), limit, offset)
	if err != nil {
		logger.Error("Failed to get pending allocations", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"code":    "500_QUERY_FAILED",
			"message": "Failed to retrieve pending allocations",
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

// GET /api/v1/admin/settlement/pending
func (h *AdminHandler) GetPendingSettlements(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	settlements, total, err := h.settlementRepo.GetAllPendingSettlements(c.Request.Context(), limit, offset)
	if err != nil {
		logger.Error("Failed to get pending settlements", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"code":    "500_QUERY_FAILED",
			"message": "Failed to retrieve pending settlements",
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
