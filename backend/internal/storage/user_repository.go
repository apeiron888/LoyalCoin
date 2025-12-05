package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/loyalcoin/backend/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepository struct {
	db *DB
}

func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateMerchant(ctx context.Context, merchant *models.Merchant) error {
	merchant.CreatedAt = time.Now().UTC()
	merchant.UpdatedAt = time.Now().UTC()
	merchant.Role = models.RoleMerchant

	collection := r.db.GetCollection("merchants")
	result, err := collection.InsertOne(ctx, merchant)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("merchant with this email already exists")
		}
		return fmt.Errorf("failed to create merchant: %w", err)
	}

	merchant.ID = result.InsertedID.(primitive.ObjectID).Hex()
	return nil
}

// Creates a new customer
func (r *UserRepository) CreateCustomer(ctx context.Context, customer *models.Customer) error {
	customer.CreatedAt = time.Now().UTC()
	customer.UpdatedAt = time.Now().UTC()

	collection := r.db.GetCollection("customers")
	result, err := collection.InsertOne(ctx, customer)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("customer with this email already exists")
		}
		return fmt.Errorf("failed to create customer: %w", err)
	}

	customer.ID = result.InsertedID.(primitive.ObjectID).Hex()
	return nil
}

// Retrieves a merchant by email
func (r *UserRepository) GetMerchantByEmail(ctx context.Context, email string) (*models.Merchant, error) {
	collection := r.db.GetCollection("merchants")

	var merchant models.Merchant
	err := collection.FindOne(ctx, bson.M{"email": email}).Decode(&merchant)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("merchant not found")
		}
		return nil, fmt.Errorf("failed to get merchant: %w", err)
	}

	return &merchant, nil
}

// Retrieves a customer by email
func (r *UserRepository) GetCustomerByEmail(ctx context.Context, email string) (*models.Customer, error) {
	collection := r.db.GetCollection("customers")

	var customer models.Customer
	err := collection.FindOne(ctx, bson.M{"email": email}).Decode(&customer)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("customer not found")
		}
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	return &customer, nil
}

// Retrieves a merchant by ID
func (r *UserRepository) GetMerchantByID(ctx context.Context, id string) (*models.Merchant, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid merchant ID: %w", err)
	}
	collection := r.db.GetCollection("merchants")

	var merchant models.Merchant
	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&merchant)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("merchant not found")
		}
		return nil, fmt.Errorf("failed to get merchant: %w", err)
	}
	return &merchant, nil
}

// Retrieves a customer by ID
func (r *UserRepository) GetCustomerByID(ctx context.Context, id string) (*models.Customer, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid customer ID: %w", err)
	}

	collection := r.db.GetCollection("customers")

	var customer models.Customer
	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&customer)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("customer not found")
		}
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	return &customer, nil
}

// UpdateMerchant updates a merchant
func (r *UserRepository) UpdateMerchant(ctx context.Context, merchant *models.Merchant) error {
	objectID, err := primitive.ObjectIDFromHex(merchant.ID)
	if err != nil {
		return fmt.Errorf("invalid merchant ID: %w", err)
	}

	merchant.UpdatedAt = time.Now().UTC()

	collection := r.db.GetCollection("merchants")
	update := bson.M{
		"$set": bson.M{
			"business_name":  merchant.BusinessName,
			"email":          merchant.Email,
			"password_hash":  merchant.PasswordHash,
			"role":           merchant.Role,
			"wallet":         merchant.Wallet,
			"allocation_lcn": merchant.AllocationLCN,
			"balance_lcn":    merchant.BalanceLCN,
			"bank_account":   merchant.BankAccount,
			"status":         merchant.Status,
			"updated_at":     merchant.UpdatedAt,
		},
	}
	_, err = collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		update,
	)
	if err != nil {
		return fmt.Errorf("failed to update merchant: %w", err)
	}
	return nil
}

// UpdateCustomer updates a customer
func (r *UserRepository) UpdateCustomer(ctx context.Context, customer *models.Customer) error {
	objectID, err := primitive.ObjectIDFromHex(customer.ID)
	if err != nil {
		return fmt.Errorf("invalid customer ID: %w", err)
	}
	customer.UpdatedAt = time.Now().UTC()

	collection := r.db.GetCollection("customers")
	update := bson.M{
		"$set": bson.M{
			"username":      customer.Username,
			"email":         customer.Email,
			"phone":         customer.Phone,
			"password_hash": customer.PasswordHash,
			"wallet":        customer.Wallet,
			"updated_at":    customer.UpdatedAt,
		},
	}
	_, err = collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		update,
	)
	if err != nil {
		return fmt.Errorf("failed to update customer: %w", err)
	}
	return nil
}
