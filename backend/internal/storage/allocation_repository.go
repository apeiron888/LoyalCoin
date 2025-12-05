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

type AllocationRepository struct {
	db *DB
}

func NewAllocationRepository(db *DB) *AllocationRepository {
	return &AllocationRepository{db: db}
}

// New allocation purchase request
func (r *AllocationRepository) CreateAllocation(ctx context.Context, allocation *models.AllocationPurchase) error {
	allocation.PurchasedAt = time.Now().UTC()
	allocation.Status = "PENDING"

	collection := r.db.GetCollection("allocation_purchases")
	result, err := collection.InsertOne(ctx, allocation)
	if err != nil {
		return fmt.Errorf("failed to create allocation purchase: %w", err)
	}

	allocation.ID = result.InsertedID.(primitive.ObjectID).Hex()
	return nil
}

// Get allocation by ID
func (r *AllocationRepository) GetAllocationByID(ctx context.Context, id string) (*models.AllocationPurchase, error) {
	collection := r.db.GetCollection("allocation_purchases")

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid allocation ID: %w", err)
	}

	var allocation models.AllocationPurchase
	err = collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&allocation)
	if err != nil {
		return nil, fmt.Errorf("allocation not found: %w", err)
	}

	return &allocation, nil
}

// Get allocations by merchant id
func (r *AllocationRepository) GetAllocationsByMerchant(ctx context.Context, merchantID string, limit, offset int, status *string) ([]*models.AllocationPurchase, int64, error) {
	collection := r.db.GetCollection("allocation_purchases")

	filter := bson.M{"merchant_id": merchantID}
	if status != nil {
		filter["status"] = *status
	}
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count allocations: %w", err)
	}

	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "purchased_at", Value: -1}})
	findOptions.SetLimit(int64(limit))
	findOptions.SetSkip(int64(offset))

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query allocations: %w", err)
	}
	defer cursor.Close(ctx)

	var allocations []*models.AllocationPurchase
	if err := cursor.All(ctx, &allocations); err != nil {
		return nil, 0, fmt.Errorf("failed to decode allocations: %w", err)
	}

	return allocations, total, nil
}

// Update allocation purchase
func (r *AllocationRepository) UpdateAllocation(ctx context.Context, allocation *models.AllocationPurchase) error {
	collection := r.db.GetCollection("allocation_purchases")

	objID, err := primitive.ObjectIDFromHex(allocation.ID)
	if err != nil {
		return fmt.Errorf("invalid allocation ID: %w", err)
	}

	update := bson.M{
		"$set": bson.M{
			"merchant_id":          allocation.MerchantID,
			"amount_lcn":           allocation.AmountLCN,
			"amount_etb_paid":      allocation.AmountETBPaid,
			"payment_method":       allocation.PaymentMethod,
			"payment_reference":    allocation.PaymentReference,
			"payment_proof_url":    allocation.PaymentProofURL,
			"status":               allocation.Status,
			"verified_at":          allocation.VerifiedAt,
			"lcn_transfer_tx_hash": allocation.LCNTransferTxHash,
			"admin_id":             allocation.AdminID,
			"admin_notes":          allocation.AdminNotes,
		},
	}

	_, err = collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return fmt.Errorf("failed to update allocation: %w", err)
	}

	return nil
}

// Get all pending allocations (admin view)
func (r *AllocationRepository) GetAllPendingAllocations(ctx context.Context, limit, offset int) ([]*models.AllocationPurchase, int64, error) {
	collection := r.db.GetCollection("allocation_purchases")

	filter := bson.M{"status": "PENDING"}

	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count pending allocations: %w", err)
	}

	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "purchased_at", Value: 1}})
	findOptions.SetLimit(int64(limit))
	findOptions.SetSkip(int64(offset))

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query pending allocations: %w", err)
	}
	defer cursor.Close(ctx)

	var allocations []*models.AllocationPurchase
	if err := cursor.All(ctx, &allocations); err != nil {
		return nil, 0, fmt.Errorf("failed to decode allocations: %w", err)
	}

	return allocations, total, nil
}
