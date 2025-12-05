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

type SettlementRepository struct {
	db *DB
}

func NewSettlementRepository(db *DB) *SettlementRepository {
	return &SettlementRepository{db: db}
}

func (r *SettlementRepository) CreateSettlement(ctx context.Context, settlement *models.SettlementRequest) error {
	settlement.RequestedAt = time.Now().UTC()
	settlement.Status = models.SettlementPending

	collection := r.db.GetCollection("settlement_requests")
	result, err := collection.InsertOne(ctx, settlement)
	if err != nil {
		return fmt.Errorf("failed to create settlement request: %w", err)
	}

	settlement.ID = result.InsertedID.(primitive.ObjectID).Hex()
	return nil
}

// Retrieves a settlement by ID
func (r *SettlementRepository) GetSettlementByID(ctx context.Context, id string) (*models.SettlementRequest, error) {
	collection := r.db.GetCollection("settlement_requests")

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid settlement ID: %w", err)
	}

	var settlement models.SettlementRequest
	err = collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&settlement)
	if err != nil {
		return nil, fmt.Errorf("settlement not found: %w", err)
	}

	return &settlement, nil
}

// Retrieves settlements for a merchant
func (r *SettlementRepository) GetSettlementsByMerchant(ctx context.Context, merchantID string, limit, offset int, status *models.SettlementStatus) ([]*models.SettlementRequest, int64, error) {
	collection := r.db.GetCollection("settlement_requests")

	filter := bson.M{"merchant_id": merchantID}
	if status != nil {
		filter["status"] = *status
	}
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count settlements: %w", err)
	}

	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "requested_at", Value: -1}})
	findOptions.SetLimit(int64(limit))
	findOptions.SetSkip(int64(offset))

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query settlements: %w", err)
	}
	defer cursor.Close(ctx)

	var settlements []*models.SettlementRequest
	if err := cursor.All(ctx, &settlements); err != nil {
		return nil, 0, fmt.Errorf("failed to decode settlements: %w", err)
	}

	return settlements, total, nil
}

// Updates a settlement request
func (r *SettlementRepository) UpdateSettlement(ctx context.Context, settlement *models.SettlementRequest) error {
	collection := r.db.GetCollection("settlement_requests")

	objID, err := primitive.ObjectIDFromHex(settlement.ID)
	if err != nil {
		return fmt.Errorf("invalid settlement ID: %w", err)
	}
	update := bson.M{
		"$set": bson.M{
			"merchant_id":       settlement.MerchantID,
			"amount_lcn":        settlement.AmountLCN,
			"amount_etb":        settlement.AmountETB,
			"exchange_rate":     settlement.ExchangeRate,
			"status":            settlement.Status,
			"bank_account":      settlement.BankAccount,
			"approved_at":       settlement.ApprovedAt,
			"processed_at":      settlement.ProcessedAt,
			"tx_hash":           settlement.TxHash,
			"payment_reference": settlement.PaymentReference,
			"admin_id":          settlement.AdminID,
			"admin_notes":       settlement.AdminNotes,
		},
	}
	_, err = collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return fmt.Errorf("failed to update settlement: %w", err)
	}
	return nil
}

// Retrieves all pending settlements (admin view)
func (r *SettlementRepository) GetAllPendingSettlements(ctx context.Context, limit, offset int) ([]*models.SettlementRequest, int64, error) {
	collection := r.db.GetCollection("settlement_requests")

	filter := bson.M{"status": models.SettlementPending}
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count pending settlements: %w", err)
	}

	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "requested_at", Value: 1}})
	findOptions.SetLimit(int64(limit))
	findOptions.SetSkip(int64(offset))

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query pending settlements: %w", err)
	}
	defer cursor.Close(ctx)

	var settlements []*models.SettlementRequest
	if err := cursor.All(ctx, &settlements); err != nil {
		return nil, 0, fmt.Errorf("failed to decode settlements: %w", err)
	}

	return settlements, total, nil
}
