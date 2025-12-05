package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/loyalcoin/backend/internal/config"
	"github.com/loyalcoin/backend/pkg/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DB struct {
	Client   *mongo.Client
	Database *mongo.Database
}

func Connect(cfg *config.Config) (*DB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(cfg.MongoDBURI)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}
	logger.Info("Successfully connected to MongoDB", map[string]interface{}{
		"database": cfg.MongoDBDatabase,
	})
	db := &DB{
		Client:   client,
		Database: client.Database(cfg.MongoDBDatabase),
	}

	// Create indexes
	if err := db.createIndexes(ctx); err != nil {
		logger.Warn("Failed to create indexes", map[string]interface{}{
			"error": err.Error(),
		})
	}

	return db, nil
}

func (db *DB) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return db.Client.Disconnect(ctx)
}

func (db *DB) createIndexes(ctx context.Context) error {
	// Merchant indexes
	merchantCollection := db.Database.Collection("merchants")
	merchantIndexes := []mongo.IndexModel{
		{
			Keys:    map[string]interface{}{"email": 1},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    map[string]interface{}{"wallet.address": 1},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: map[string]interface{}{"status": 1},
		},
	}
	if _, err := merchantCollection.Indexes().CreateMany(ctx, merchantIndexes); err != nil {
		return fmt.Errorf("failed to create merchant indexes: %w", err)
	}

	// Customer indexes
	customerCollection := db.Database.Collection("customers")
	customerIndexes := []mongo.IndexModel{
		{
			Keys:    map[string]interface{}{"email": 1},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    map[string]interface{}{"wallet.address": 1},
			Options: options.Index().SetUnique(true),
		},
	}
	if _, err := customerCollection.Indexes().CreateMany(ctx, customerIndexes); err != nil {
		return fmt.Errorf("failed to create customer indexes: %w", err)
	}

	// Transaction log indexes
	txLogCollection := db.Database.Collection("transaction_logs")
	txLogIndexes := []mongo.IndexModel{
		{
			Keys:    map[string]interface{}{"tx_hash": 1},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: map[string]interface{}{"from_address": 1},
		},
		{
			Keys: map[string]interface{}{"to_address": 1},
		},
		{
			Keys: map[string]interface{}{"status": 1},
		},
		{
			Keys: map[string]interface{}{"type": 1},
		},
		{
			Keys: map[string]interface{}{"submitted_at": -1},
		},
	}
	if _, err := txLogCollection.Indexes().CreateMany(ctx, txLogIndexes); err != nil {
		return fmt.Errorf("failed to create transaction log indexes: %w", err)
	}

	// Settlement request indexes
	settlementCollection := db.Database.Collection("settlement_requests")
	settlementIndexes := []mongo.IndexModel{
		{
			Keys: map[string]interface{}{"merchant_id": 1},
		},
		{
			Keys: map[string]interface{}{"status": 1},
		},
		{
			Keys: map[string]interface{}{"requested_at": -1},
		},
	}
	if _, err := settlementCollection.Indexes().CreateMany(ctx, settlementIndexes); err != nil {
		return fmt.Errorf("failed to create settlement indexes: %w", err)
	}

	// UTXO cache indexes
	utxoCacheCollection := db.Database.Collection("utxo_cache")
	utxoIndexes := []mongo.IndexModel{
		{
			Keys:    map[string]interface{}{"address": 1},
			Options: options.Index().SetUnique(true),
		},
	}
	if _, err := utxoCacheCollection.Indexes().CreateMany(ctx, utxoIndexes); err != nil {
		return fmt.Errorf("failed to create UTXO cache indexes: %w", err)
	}

	logger.Info("Successfully created database indexes", nil)
	return nil
}

func (db *DB) GetCollection(name string) *mongo.Collection {
	return db.Database.Collection(name)
}
