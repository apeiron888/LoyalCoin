package cardano

import (
	"fmt"
	"strconv"

	"github.com/loyalcoin/backend/internal/models"
)

type TxOutput struct {
	Address  string
	Lovelace uint64
	Assets   map[string]uint64
}

type TxBuilder struct {
	minADAOutput uint64
	feeA         uint64
	feeB         uint64
	feeBuffer    float64
}

func NewTxBuilder(minADAOutput, feeA, feeB uint64, feeBuffer float64) *TxBuilder {
	return &TxBuilder{
		minADAOutput: minADAOutput,
		feeA:         feeA,
		feeB:         feeB,
		feeBuffer:    feeBuffer,
	}
}

type UTXOSelectionResult struct {
	SelectedUTXOs  []models.UTXO
	TotalLovelace  uint64
	TotalAssets    map[string]uint64
	ChangeLovelace uint64
	ChangeAssets   map[string]uint64
	EstimatedFee   uint64
}

func getLovelace(utxo models.UTXO) uint64 {
	return utxo.Value.Lovelace
}

func getAssets(utxo models.UTXO) map[string]uint64 {
	assets := make(map[string]uint64)
	for _, asset := range utxo.Value.Assets {
		key := fmt.Sprintf("%s.%s", asset.PolicyID, asset.AssetName)
		assets[key] = asset.Quantity
	}
	return assets
}

// Greedy UTXO selection for given outputs
func (b *TxBuilder) SelectUTXOs(availableUTXOs []models.UTXO, outputs []TxOutput) (*UTXOSelectionResult, error) {
	requiredLovelace := uint64(0)
	requiredAssets := make(map[string]uint64)

	for _, output := range outputs {
		requiredLovelace += output.Lovelace
		for assetID, quantity := range output.Assets {
			requiredAssets[assetID] += quantity
		}
	}
	sortedUTXOs := make([]models.UTXO, len(availableUTXOs))
	copy(sortedUTXOs, availableUTXOs)

	// Simple bubble sort (sufficient for small UTXO sets)
	for i := 0; i < len(sortedUTXOs); i++ {
		for j := i + 1; j < len(sortedUTXOs); j++ {
			if getLovelace(sortedUTXOs[j]) > getLovelace(sortedUTXOs[i]) {
				sortedUTXOs[i], sortedUTXOs[j] = sortedUTXOs[j], sortedUTXOs[i]
			}
		}
	}
	selectedUTXOs := []models.UTXO{}
	accumulatedLovelace := uint64(0)
	accumulatedAssets := make(map[string]uint64)

	for _, utxo := range sortedUTXOs {
		selectedUTXOs = append(selectedUTXOs, utxo)
		accumulatedLovelace += getLovelace(utxo)

		for assetID, quantity := range getAssets(utxo) {
			accumulatedAssets[assetID] += quantity
		}
		estimatedFee := b.estimateFee(len(selectedUTXOs), len(outputs)+1)
		if accumulatedLovelace >= requiredLovelace+estimatedFee {
			assetsOK := true
			for assetID, required := range requiredAssets {
				if accumulatedAssets[assetID] < required {
					assetsOK = false
					break
				}
			}
			if assetsOK {
				changeLovelace := accumulatedLovelace - requiredLovelace - estimatedFee
				changeAssets := make(map[string]uint64)

				for assetID, accumulated := range accumulatedAssets {
					required := requiredAssets[assetID]
					if accumulated > required {
						changeAssets[assetID] = accumulated - required
					}
				}
				if len(changeAssets) > 0 && changeLovelace < b.minADAOutput {
					continue
				}

				return &UTXOSelectionResult{
					SelectedUTXOs:  selectedUTXOs,
					TotalLovelace:  accumulatedLovelace,
					TotalAssets:    accumulatedAssets,
					ChangeLovelace: changeLovelace,
					ChangeAssets:   changeAssets,
					EstimatedFee:   estimatedFee,
				}, nil
			}
		}
		if len(selectedUTXOs) > 50 {
			return nil, fmt.Errorf("transaction too large: exceeded 50 inputs")
		}
	}

	return nil, fmt.Errorf("insufficient funds: need %d lovelace, have %d", requiredLovelace, accumulatedLovelace)
}

func (b *TxBuilder) estimateFee(numInputs, numOutputs int) uint64 {
	estimatedSize := 10 + (numInputs * 150) + (numOutputs * 200)
	// Fee Calc
	baseFee := b.feeA + (b.feeB * uint64(estimatedSize))

	// Apply buffer (20% by default)
	bufferedFee := uint64(float64(baseFee) * b.feeBuffer)

	return bufferedFee
}

// Calculates minimum ADA required for an output
func (b *TxBuilder) CalculateMinADA(numAssets int) uint64 {
	if numAssets == 0 {
		return 1_000_000
	}
	return b.minADAOutput + (uint64(numAssets) * 200_000)
}

// Validates that all outputs meet minimum ADA requirements
func (b *TxBuilder) ValidateOutputs(outputs []TxOutput) error {
	for i, output := range outputs {
		requiredMinADA := b.CalculateMinADA(len(output.Assets))

		if output.Lovelace < requiredMinADA {
			return fmt.Errorf("output %d has insufficient ADA: has %d, needs %d",
				i, output.Lovelace, requiredMinADA)
		}
	}
	return nil
}

// Converts Blockfrost UTXOs to models.UTXO format
func ConvertBlockfrostUTXOs(bfUTXOs []AddressUTXO) ([]models.UTXO, error) {
	utxos := make([]models.UTXO, 0, len(bfUTXOs))

	for _, bfUTXO := range bfUTXOs {
		utxo := models.UTXO{
			TxHash: bfUTXO.TxHash,
			Index:  bfUTXO.OutputIndex,
		}

		utxoAssets := []models.UTXOAsset{}
		for _, asset := range bfUTXO.Amount {
			if asset.Unit == "lovelace" {
				lovelace, err := strconv.ParseUint(asset.Quantity, 10, 64)
				if err != nil {
					return nil, fmt.Errorf("invalid lovelace amount: %w", err)
				}
				utxo.Value.Lovelace = lovelace
			} else {
				policyID := asset.Unit
				assetName := ""

				if len(asset.Unit) > 56 {
					policyID = asset.Unit[:56]
					assetName = asset.Unit[56:]
				}

				quantity, err := strconv.ParseUint(asset.Quantity, 10, 64)
				if err != nil {
					return nil, fmt.Errorf("invalid asset quantity: %w", err)
				}
				utxoAssets = append(utxoAssets, models.UTXOAsset{
					PolicyID:  policyID,
					AssetName: assetName,
					Quantity:  quantity,
				})
			}
		}
		utxo.Value.Assets = utxoAssets
		utxos = append(utxos, utxo)
	}
	return utxos, nil
}
