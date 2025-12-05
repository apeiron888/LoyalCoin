package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/loyalcoin/backend/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TxLogRepository struct {
	db *DB
}

func NewTxLogRepository(db *DB) *TxLogRepository {
	return &TxLogRepository{db: db}
}

func (r *TxLogRepository) CreateTxLog(ctx context.Context, txLog *models.TxLog) error {
	txLog.SubmittedAt = time.Now().UTC()

	collection := r.db.GetCollection("transaction_logs")
	result, err := collection.InsertOne(ctx, txLog)
	if err != nil {
		return fmt.Errorf("failed to create transaction log: %w", err)
	}
	txLog.ID = result.InsertedID.(primitive.ObjectID).Hex()
	return nil
}

func (r *TxLogRepository) GetTxLogByHash(ctx context.Context, txHash string) (*models.TxLog, error) {
	collection := r.db.GetCollection("transaction_logs")

	var txLog models.TxLog
	err := collection.FindOne(ctx, bson.M{"tx_hash": txHash}).Decode(&txLog)
	if err != nil {
		return nil, fmt.Errorf("transaction not found: %w", err)
	}

	return &txLog, nil
}

// Retrieves transaction logs for an address
func (r *TxLogRepository) GetTxLogsByAddress(ctx context.Context, address string, limit, offset int) ([]*models.TxLog, error) {
	collection := r.db.GetCollection("transaction_logs")
	// Query for transactions where address is sender or receiver
	filter := bson.M{
		"$or": []bson.M{
			{"from_address": address},
			{"to_address": address},
		},
	}
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "submitted_at", Value: -1}})
	findOptions.SetLimit(int64(limit))
	findOptions.SetSkip(int64(offset))

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to query transactions: %w", err)
	}
	defer cursor.Close(ctx)

	var txLogs []*models.TxLog
	if err := cursor.All(ctx, &txLogs); err != nil {
		return nil, fmt.Errorf("failed to decode transactions: %w", err)
	}
	return txLogs, nil
}

// Updates the status of a transaction
func (r *TxLogRepository) UpdateTxStatus(ctx context.Context, txHash string, status models.TxStatus, blockHeight int64) error {
	collection := r.db.GetCollection("transaction_logs")

	update := bson.M{
		"$set": bson.M{
			"status":       status,
			"block_height": blockHeight,
		},
	}
	if status == models.TxStatusConfirmed {
		update["$set"].(bson.M)["confirmed_at"] = time.Now().UTC()
	}
	_, err := collection.UpdateOne(ctx, bson.M{"tx_hash": txHash}, update)
	if err != nil {
		return fmt.Errorf("failed to update transaction status: %w", err)
	}
	return nil
}

// Retrieves pending transactions for processing
func (r *TxLogRepository) GetPendingTransactions(ctx context.Context, limit int) ([]models.TxLog, error) {
	collection := r.db.GetCollection("transaction_logs")

	filter := bson.M{"status": models.TxStatusPending}
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "submitted_at", Value: 1}})
	findOptions.SetLimit(int64(limit))

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending transactions: %w", err)
	}
	defer cursor.Close(ctx)

	var txLogs []models.TxLog
	if err := cursor.All(ctx, &txLogs); err != nil {
		return nil, fmt.Errorf("failed to decode pending transactions: %w", err)
	}
	return txLogs, nil
}

// Updates a transaction log entry
func (r *TxLogRepository) UpdateTransaction(ctx context.Context, txLog *models.TxLog) error {
	collection := r.db.GetCollection("transaction_logs")
	update := bson.M{
		"$set": bson.M{
			"tx_hash":         txLog.TxHash,
			"from_address":    txLog.FromAddress,
			"to_address":      txLog.ToAddress,
			"amount_lcn":      txLog.AmountLCN,
			"asset_policy_id": txLog.AssetPolicyID,
			"asset_name":      txLog.AssetName,
			"type":            txLog.Type,
			"status":          txLog.Status,
			"block_height":    txLog.BlockHeight,
			"confirmed_at":    txLog.ConfirmedAt,
			"meta":            txLog.Meta,
		},
	}
	objID, err := primitive.ObjectIDFromHex(txLog.ID)
	if err != nil {
		_, err := collection.UpdateOne(
			ctx,
			bson.M{"tx_hash": txLog.TxHash},
			update,
		)
		return err
	}
	_, err = collection.UpdateOne(
		ctx,
		bson.M{"_id": objID},
		update,
	)
	if err != nil {
		return fmt.Errorf("failed to update transaction: %w", err)
	}
	return nil
}
