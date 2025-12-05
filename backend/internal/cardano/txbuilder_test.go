package cardano

import (
	"testing"

	"github.com/loyalcoin/backend/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestConvertBlockfrostUTXOs(t *testing.T) {
	tests := []struct {
		name     string
		input    []AddressUTXO
		expected []models.UTXO
		wantErr  bool
	}{
		{
			name: "Lovelace only",
			input: []AddressUTXO{
				{
					TxHash:      "hash1",
					OutputIndex: 0,
					Amount: []Asset{
						{Unit: "lovelace", Quantity: "1000000"},
					},
				},
			},
			expected: []models.UTXO{
				{
					TxHash: "hash1",
					Index:  0,
					Value: models.UTXOValue{
						Lovelace: 1000000,
						Assets:   []models.UTXOAsset{},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Lovelace and Asset",
			input: []AddressUTXO{
				{
					TxHash:      "hash2",
					OutputIndex: 1,
					Amount: []Asset{
						{Unit: "lovelace", Quantity: "2000000"},
						{Unit: "394bbb17f90e895ab320b25ac9f999ffc1759acc79e8e64a67ef7a7f4c434e", Quantity: "100"},
					},
				},
			},
			expected: []models.UTXO{
				{
					TxHash: "hash2",
					Index:  1,
					Value: models.UTXOValue{
						Lovelace: 2000000,
						Assets: []models.UTXOAsset{
							{
								PolicyID:  "394bbb17f90e895ab320b25ac9f999ffc1759acc79e8e64a67ef7a7f",
								AssetName: "4c434e",
								Quantity:  100,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Asset with empty name (PolicyID only)",
			input: []AddressUTXO{
				{
					TxHash:      "hash3",
					OutputIndex: 2,
					Amount: []Asset{
						{Unit: "lovelace", Quantity: "3000000"},
						{Unit: "394bbb17f90e895ab320b25ac9f999ffc1759acc79e8e64a67ef7a7f", Quantity: "50"},
					},
				},
			},
			expected: []models.UTXO{
				{
					TxHash: "hash3",
					Index:  2,
					Value: models.UTXOValue{
						Lovelace: 3000000,
						Assets: []models.UTXOAsset{
							{
								PolicyID:  "394bbb17f90e895ab320b25ac9f999ffc1759acc79e8e64a67ef7a7f",
								AssetName: "",
								Quantity:  50,
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertBlockfrostUTXOs(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tt.expected), len(got))
				for i := range got {
					assert.Equal(t, tt.expected[i].TxHash, got[i].TxHash)
					assert.Equal(t, tt.expected[i].Index, got[i].Index)
					assert.Equal(t, tt.expected[i].Value.Lovelace, got[i].Value.Lovelace)
					assert.Equal(t, len(tt.expected[i].Value.Assets), len(got[i].Value.Assets))
					for j := range got[i].Value.Assets {
						assert.Equal(t, tt.expected[i].Value.Assets[j], got[i].Value.Assets[j])
					}
				}
			}
		})
	}
}
