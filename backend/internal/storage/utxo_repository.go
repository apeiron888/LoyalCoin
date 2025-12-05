package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/loyalcoin/backend/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UTXORepository struct {
	db *DB
}

func NewUTXORepository(db *DB) *UTXORepository {
	return &UTXORepository{db: db}
}

// Cached UTXOs for an address
type CachedUTXOs struct {
	Address     string
	UTXOs       []models.UTXO
	LastFetched time.Time
}

// Retrieves cached UTXOs for an address
func (r *UTXORepository) GetUTXOsByAddress(ctx context.Context, address string) (*CachedUTXOs, error) {
	collection := r.db.GetCollection("utxo_cache")

	var result struct {
		Address     string    `bson:"address"`
		UTXOs       []bson.M  `bson:"utxos"`
		LastFetched time.Time `bson:"last_fetched"`
	}
	err := collection.FindOne(ctx, bson.M{"address": address}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &CachedUTXOs{
				Address:     address,
				UTXOs:       []models.UTXO{},
				LastFetched: time.Time{},
			}, nil
		}
		return nil, fmt.Errorf("failed to get UTXOs: %w", err)
	}

	utxos := make([]models.UTXO, 0, len(result.UTXOs))
	for _, utxoData := range result.UTXOs {
		utxo := models.UTXO{
			TxHash: utxoData["tx_hash"].(string),
			Index:  int(utxoData["output_index"].(int32)),
		}
		utxo.Value.Lovelace = uint64(utxoData["lovelace"].(int64))
		if assetsData, ok := utxoData["assets"].(bson.M); ok {
			utxo.Value.Assets = []models.UTXOAsset{}
			for assetID, quantity := range assetsData {
				utxo.Value.Assets = append(utxo.Value.Assets, models.UTXOAsset{
					PolicyID: assetID,
					Quantity: uint64(quantity.(int64)),
				})
			}
		}

		utxos = append(utxos, utxo)
	}

	return &CachedUTXOs{
		Address:     result.Address,
		UTXOs:       utxos,
		LastFetched: result.LastFetched,
	}, nil
}

// Updates the UTXO cache for an address
func (r *UTXORepository) UpdateUTXOs(ctx context.Context, address string, utxos []models.UTXO) error {
	collection := r.db.GetCollection("utxo_cache")

	utxosData := make([]bson.M, 0, len(utxos))
	for _, utxo := range utxos {
		assetsData := make(map[string]int64)
		for _, asset := range utxo.Value.Assets {
			assetsData[asset.PolicyID] = int64(asset.Quantity)
		}
		utxoData := bson.M{
			"tx_hash":      utxo.TxHash,
			"output_index": utxo.Index,
			"lovelace":     int64(utxo.Value.Lovelace),
			"assets":       assetsData,
		}
		utxosData = append(utxosData, utxoData)
	}

	update := bson.M{
		"$set": bson.M{
			"address":      address,
			"utxos":        utxosData,
			"last_fetched": time.Now().UTC(),
		},
	}
	upsert := true
	updateOptions := options.Update().SetUpsert(upsert)

	_, err := collection.UpdateOne(
		ctx,
		bson.M{"address": address},
		update,
		updateOptions,
	)

	if err != nil {
		return fmt.Errorf("failed to update UTXOs: %w", err)
	}

	return nil
}

// ClearCache clears UTXO cache for an address
func (r *UTXORepository) ClearCache(ctx context.Context, address string) error {
	collection := r.db.GetCollection("utxo_cache")

	_, err := collection.DeleteOne(ctx, bson.M{"address": address})
	if err != nil {
		return fmt.Errorf("failed to clear cache: %w", err)
	}

	return nil
}
