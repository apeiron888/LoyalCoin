// TODO 2 Notifications not done yet
package indexer

import (
	"context"
	"time"

	"github.com/loyalcoin/backend/internal/cardano"
	"github.com/loyalcoin/backend/internal/models"
	"github.com/loyalcoin/backend/internal/storage"
	"github.com/loyalcoin/backend/pkg/logger"
)

type Config struct {
	PollInterval        time.Duration
	BatchSize           int
	ConfirmationBlocks  int64
	MaxRetries          int
	RetryBackoff        time.Duration
	EnableNotifications bool
}

func DefaultConfig() *Config {
	return &Config{
		PollInterval:        30 * time.Second,
		BatchSize:           50,
		ConfirmationBlocks:  3,
		MaxRetries:          10,
		RetryBackoff:        5 * time.Minute,
		EnableNotifications: true,
	}
}

type Service struct {
	config     *Config
	blockfrost *cardano.BlockfrostClient
	txLogRepo  *storage.TxLogRepository
	userRepo   *storage.UserRepository
	stopCh     chan struct{}
	stoppedCh  chan struct{}
}

func NewService(
	config *Config,
	blockfrost *cardano.BlockfrostClient,
	txLogRepo *storage.TxLogRepository,
	userRepo *storage.UserRepository,
) *Service {
	if config == nil {
		config = DefaultConfig()
	}

	return &Service{
		config:     config,
		blockfrost: blockfrost,
		txLogRepo:  txLogRepo,
		userRepo:   userRepo,
		stopCh:     make(chan struct{}),
		stoppedCh:  make(chan struct{}),
	}
}

// Begins the indexer background process
func (s *Service) Start() {
	logger.Info("Starting transaction indexer service", map[string]interface{}{
		"poll_interval":       s.config.PollInterval,
		"confirmation_blocks": s.config.ConfirmationBlocks,
	})

	go s.run()
}

// Gracefully stops the indexer
func (s *Service) Stop() {
	logger.Info("Stopping transaction indexer service", nil)
	close(s.stopCh)
	<-s.stoppedCh
	logger.Info("Transaction indexer service stopped", nil)
}

// Main indexer loop
func (s *Service) run() {
	defer close(s.stoppedCh)

	ticker := time.NewTicker(s.config.PollInterval)
	defer ticker.Stop()

	s.processPendingTransactions()

	for {
		select {
		case <-ticker.C:
			s.processPendingTransactions()
		case <-s.stopCh:
			return
		}
	}
}

// Fetches and processes all pending transactions
func (s *Service) processPendingTransactions() {
	ctx := context.Background()

	// Fetch pending transactions
	pendingTxs, err := s.txLogRepo.GetPendingTransactions(ctx, s.config.BatchSize)
	if err != nil {
		logger.Error("Failed to fetch pending transactions", err, nil)
		return
	}
	if len(pendingTxs) == 0 {
		logger.Debug("No pending transactions to process", nil)
		return
	}
	logger.Info("Processing pending transactions", map[string]interface{}{
		"count": len(pendingTxs),
	})
	for _, tx := range pendingTxs {
		s.processTransaction(ctx, &tx)
	}
}

// Checks and updates a single transaction
func (s *Service) processTransaction(ctx context.Context, tx *models.TxLog) {
	// Query Blockfrost for transaction details
	txDetails, err := s.blockfrost.GetTransactionDetails(tx.TxHash)
	if err != nil {
		logger.Debug("Transaction not yet on-chain", map[string]interface{}{
			"tx_hash": tx.TxHash,
			"error":   err.Error(),
		})
		// Check if we should mark as failed after max retries
		elapsed := time.Since(tx.SubmittedAt)
		if elapsed > s.config.RetryBackoff*time.Duration(s.config.MaxRetries) {
			s.markTransactionFailed(ctx, tx, "Transaction not confirmed after maximum retry period")
		}
		return
	}
	// Check if transaction is confirmed
	if !txDetails.Confirmed {
		logger.Debug("Transaction still pending confirmation", map[string]interface{}{
			"tx_hash": tx.TxHash,
		})
		return
	}
	currentBlock, err := s.blockfrost.GetLatestBlock()
	if err != nil {
		logger.Error("Failed to get latest block", err, map[string]interface{}{
			"tx_hash": tx.TxHash,
		})
		return
	}

	confirmations := currentBlock.Height - txDetails.BlockHeight
	if confirmations < s.config.ConfirmationBlocks {
		logger.Debug("Waiting for more confirmations", map[string]interface{}{
			"tx_hash":       tx.TxHash,
			"confirmations": confirmations,
			"required":      s.config.ConfirmationBlocks,
		})
		return
	}
	s.markTransactionConfirmed(ctx, tx, txDetails)
}

func (s *Service) markTransactionConfirmed(ctx context.Context, tx *models.TxLog, details *cardano.TransactionDetails) {
	now := time.Now().UTC()
	tx.Status = models.TxStatusConfirmed
	tx.BlockHeight = details.BlockHeight
	tx.ConfirmedAt = &now

	err := s.txLogRepo.UpdateTransaction(ctx, tx)
	if err != nil {
		logger.Error("Failed to update transaction status", err, map[string]interface{}{
			"tx_hash": tx.TxHash,
		})
		return
	}
	logger.Info("Transaction confirmed", map[string]interface{}{
		"tx_hash":      tx.TxHash,
		"block_height": details.BlockHeight,
		"type":         tx.Type,
	})
	if s.config.EnableNotifications {
		s.notifyTransactionConfirmed(ctx, tx)
	}
}

// Updates transaction status to FAILED
func (s *Service) markTransactionFailed(ctx context.Context, tx *models.TxLog, reason string) {
	tx.Status = models.TxStatusFailed
	if tx.Meta == nil {
		tx.Meta = make(map[string]interface{})
	}
	tx.Meta["failure_reason"] = reason

	err := s.txLogRepo.UpdateTransaction(ctx, tx)
	if err != nil {
		logger.Error("Failed to mark transaction as failed", err, map[string]interface{}{
			"tx_hash": tx.TxHash,
		})
		return
	}
	logger.Warn("Transaction marked as failed", map[string]interface{}{
		"tx_hash": tx.TxHash,
		"reason":  reason,
	})
	if s.config.EnableNotifications {
		s.notifyTransactionFailed(ctx, tx, reason)
	}
}

// Sends a confirmation notification
func (s *Service) notifyTransactionConfirmed(ctx context.Context, tx *models.TxLog) {
	// TODO: Implement notification logic
	// - WebSocket to connected clients
	// - Push notification to mobile app
	// - Email notification (optional)
	logger.Debug("Notification sent: transaction confirmed", map[string]interface{}{
		"tx_hash": tx.TxHash,
	})
}

// Sends a failure notification
func (s *Service) notifyTransactionFailed(ctx context.Context, tx *models.TxLog, reason string) {
	// TODO: Implement notification logic
	logger.Debug("Notification sent: transaction failed", map[string]interface{}{
		"tx_hash": tx.TxHash,
		"reason":  reason,
	})
}
