package cardano

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/loyalcoin/backend/internal/config"
	"github.com/loyalcoin/backend/internal/crypto"
	"github.com/loyalcoin/backend/internal/models"
	"github.com/loyalcoin/backend/internal/storage"
	"github.com/loyalcoin/backend/pkg/logger"
)

type CardanoService struct {
	blockfrost    *BlockfrostClient
	txBuilder     *TxBuilder
	utxoRepo      *storage.UTXORepository
	txLogRepo     *storage.TxLogRepository
	walletService *crypto.WalletService // Added WalletService
	policyID      string
	assetName     string
}

func NewCardanoService(
	cfg *config.Config,
	utxoRepo *storage.UTXORepository,
	txLogRepo *storage.TxLogRepository,
	walletService *crypto.WalletService,
) *CardanoService {
	blockfrost := NewBlockfrostClient(cfg.BlockfrostProjectID, cfg.BlockfrostAPIURL)

	txBuilder := NewTxBuilder(
		cfg.MinADAOutput,
		cfg.FeeA,
		cfg.FeeB,
		cfg.FeeBufferMultiplier,
	)
	return &CardanoService{
		blockfrost:    blockfrost,
		txBuilder:     txBuilder,
		utxoRepo:      utxoRepo,
		txLogRepo:     txLogRepo,
		walletService: walletService,
		policyID:      cfg.LCNPolicyID,
		assetName:     cfg.LCNAssetName,
	}
}

type Balance struct {
	Address     string
	Lovelace    uint64
	ADA         float64
	LCNAtomic   uint64
	LCN         float64
	OtherAssets map[string]uint64
}

// Retrieves wallet balance (ADA-backed LCN)
func (s *CardanoService) GetBalance(address string) (*Balance, error) {
	ctx := context.Background()
	cachedUTXOs, err := s.utxoRepo.GetUTXOsByAddress(ctx, address)

	var utxos []models.UTXO

	if err != nil || time.Since(cachedUTXOs.LastFetched) > 1*time.Minute {
		logger.Debug("Fetching UTXOs from Blockfrost", map[string]interface{}{
			"address": address,
		})

		bfUTXOs, err := s.blockfrost.GetAddressUTXOs(address)
		if err != nil {
			// Check if it's a "not found" error (unfunded address)
			errMsg := err.Error()
			if strings.Contains(errMsg, "404") || strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "Not Found") {
				logger.Debug("Address has no transactions (unfunded)", map[string]interface{}{
					"address": address,
				})
				// Return zero balance for unfunded addresses
				utxos = []models.UTXO{}
			} else {
				return nil, fmt.Errorf("failed to get UTXOs: %w", err)
			}
		} else {
			utxos, err = ConvertBlockfrostUTXOs(bfUTXOs)
			if err != nil {
				return nil, fmt.Errorf("failed to convert UTXOs: %w", err)
			}

			// Update cache
			if err := s.utxoRepo.UpdateUTXOs(ctx, address, utxos); err != nil {
				logger.Warn("Failed to update UTXO cache", map[string]interface{}{
					"error": err.Error(),
				})
			}
		}
	} else {
		// Use cached UTXOs
		utxos = cachedUTXOs.UTXOs
	}

	// Calculate ADA balance only
	balance := &Balance{
		Address:     address,
		OtherAssets: make(map[string]uint64),
	}

	for _, utxo := range utxos {
		balance.Lovelace += utxo.Value.Lovelace
	}

	balance.ADA = float64(balance.Lovelace) / 1_000_000

	// LCN is backed by ADA at 1 ADA = 100 LCN ratio
	balance.LCNAtomic = balance.Lovelace / 10000 // Lovelace / 10,000
	balance.LCN = balance.ADA * 100              // ADA × 100
	return balance, nil
}

// Transfers ADA (representing LCN at 1 ADA = 100 LCN ratio)
func (s *CardanoService) TransferADA(
	fromAddress string,
	toAddress string,
	amountLCN uint64, // in whole LCN units
	encryptedPrivateKey string,
) (string, error) {
	// Convert LCN to Lovelace
	// 100 LCN = 1 ADA = 1,000,000 Lovelace
	// So: 1 LCN = 0.01 ADA = 10,000 Lovelace
	lovelaceAmount := amountLCN * 10000 // LCN to Lovelace

	// 1. Decrypt private key
	privateKey, err := s.walletService.DecryptPrivateKey(encryptedPrivateKey)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt private key: %w", err)
	}
	privateKeyHex := crypto.PrivateKeyToHex(privateKey)

	// 2. Prepare input for Node.js simple transfer script
	inputData := map[string]interface{}{
		"privateKey": privateKeyHex,
		"toAddress":  toAddress,
		"lovelace":   lovelaceAmount,
	}

	inputJSON, err := json.Marshal(inputData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal input data: %w", err)
	}

	// 3. Execute simple ADA transfer script
	cmd := exec.Command("node", "scripts/transfer/transfer-ada.mjs")
	cmd.Env = append(os.Environ(), "NODE_OPTIONS=--dns-result-order=ipv4first")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start transfer script: %w", err)
	}

	_, err = stdin.Write(inputJSON)
	if err != nil {
		return "", fmt.Errorf("failed to write to stdin: %w", err)
	}
	stdin.Close()

	if err := cmd.Wait(); err != nil {
		return "", fmt.Errorf("transfer script failed: %s (stderr: %s)", err, stderr.String())
	}

	// 4. Parse output
	var result struct {
		Status  string `json:"status"`
		TxHash  string `json:"txHash"`
		Message string `json:"message"`
	}

	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return "", fmt.Errorf("failed to parse script output: %w (output: %s)", err, stdout.String())
	}

	if result.Status != "ok" {
		return "", fmt.Errorf("transfer failed: %s", result.Message)
	}

	txHash := result.TxHash

	// 5. Record transaction
	ctx := context.Background()
	txLog := &models.TxLog{
		TxHash:        txHash,
		FromAddress:   fromAddress,
		ToAddress:     toAddress,
		AmountLCN:     amountLCN,
		AssetPolicyID: "ADA", // Mark as ADA-backed
		AssetName:     "LCN",
		Type:          models.TxTypeIssuance,
		Status:        models.TxStatusPending,
		SubmittedAt:   time.Now().UTC(),
	}

	if err := s.txLogRepo.CreateTxLog(ctx, txLog); err != nil {
		logger.Warn("Failed to record transaction", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// 6. Clear UTXO cache for both addresses to ensure fresh UTXOs are fetched
	// This prevents "BadInputsUTxO" errors on subsequent transactions
	if err := s.utxoRepo.ClearCache(ctx, fromAddress); err != nil {
		logger.Warn("Failed to clear UTXO cache for sender", map[string]interface{}{
			"address": fromAddress,
			"error":   err.Error(),
		})
	}
	if err := s.utxoRepo.ClearCache(ctx, toAddress); err != nil {
		logger.Warn("Failed to clear UTXO cache for receiver", map[string]interface{}{
			"address": toAddress,
			"error":   err.Error(),
		})
	}

	return txHash, nil
}

// Transfers LCN tokens from one address to another
func (s *CardanoService) TransferLCN(
	fromAddress string,
	toAddress string,
	amountLCN uint64, // in atomic units
	encryptedPrivateKey string,
) (string, error) {
	// 1. Decrypt private key
	privateKey, err := s.walletService.DecryptPrivateKey(encryptedPrivateKey)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt private key: %w", err)
	}
	// Convert to hex for the script
	privateKeyHex := crypto.PrivateKeyToHex(privateKey)

	// 2. Prepare input for Node.js script
	lcnAssetID := fmt.Sprintf("%s%s", s.policyID, s.assetName)

	inputData := map[string]interface{}{
		"privateKey": privateKeyHex,
		"toAddress":  toAddress,
		"amount":     amountLCN,
		"assetId":    lcnAssetID,
	}

	inputJSON, err := json.Marshal(inputData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal input data: %w", err)
	}

	// 3. Execute Node.js script
	// Use the transfer.mjs script we created
	cmd := exec.Command("node", "scripts/transfer/transfer.mjs")
	// Set NODE_OPTIONS to force IPv4 if needed (based on previous experience)
	cmd.Env = append(os.Environ(), "NODE_OPTIONS=--dns-result-order=ipv4first")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start transfer script: %w", err)
	}

	// Write input to stdin
	_, err = stdin.Write(inputJSON)
	if err != nil {
		return "", fmt.Errorf("failed to write to stdin: %w", err)
	}
	stdin.Close()

	if err := cmd.Wait(); err != nil {
		return "", fmt.Errorf("transfer script failed: %s (stderr: %s)", err, stderr.String())
	}

	// 4. Parse output
	var result struct {
		Status  string `json:"status"`
		TxHash  string `json:"txHash"`
		Message string `json:"message"`
	}

	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		// Try to read raw stdout if JSON parse fails
		return "", fmt.Errorf("failed to parse script output: %w (output: %s)", err, stdout.String())
	}

	if result.Status != "ok" {
		return "", fmt.Errorf("transfer failed: %s", result.Message)
	}

	txHash := result.TxHash

	// 5. Record transaction
	ctx := context.Background()
	txLog := &models.TxLog{
		TxHash:        txHash,
		FromAddress:   fromAddress,
		ToAddress:     toAddress,
		AmountLCN:     amountLCN,
		AssetPolicyID: s.policyID,
		AssetName:     s.assetName,
		Type:          models.TxTypeIssuance,
		Status:        models.TxStatusPending,
		SubmittedAt:   time.Now().UTC(),
	}

	if err := s.txLogRepo.CreateTxLog(ctx, txLog); err != nil {
		logger.Warn("Failed to record transaction", map[string]interface{}{
			"error": err.Error(),
		})
	}

	return txHash, nil
}

func (s *CardanoService) Health() error {
	return s.blockfrost.Health()
}
