//go:build integration

package cardano

import (
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func TestBlockfrostIntegration(t *testing.T) {
	// Load .env from project root
	// Since tests run in the package directory, we need to go up two levels
	err := godotenv.Load("../../.env")
	if err != nil {
		t.Log("Warning: .env file not found, relying on environment variables")
	}

	projectID := os.Getenv("BLOCKFROST_PROJECT_ID")
	apiURL := os.Getenv("BLOCKFROST_API_URL")

	if projectID == "" {
		t.Skip("Skipping integration test: BLOCKFROST_PROJECT_ID not set")
	}
	if apiURL == "" {
		apiURL = "https://cardano-preprod.blockfrost.io/api/v0"
	}

	client := NewBlockfrostClient(projectID, apiURL)

	t.Run("Health Check", func(t *testing.T) {
		err := client.Health()
		assert.NoError(t, err)
	})

	t.Run("Get Protocol Parameters", func(t *testing.T) {
		params, err := client.GetProtocolParameters()
		assert.NoError(t, err)
		assert.NotEmpty(t, params)
	})

	t.Run("Get Address Info", func(t *testing.T) {
		// Use the merchant address we funded
		address := "addr_test1vqnj7lgsnepa94vtfwkx2utdw5urdeah9u6uk2udz0vgn5qdu9ant"
		info, err := client.GetAddressInfo(address)
		assert.NoError(t, err)
		assert.NotNil(t, info)
		assert.Equal(t, address, info.Address)

		// Verify balance (should have ADA and LCN)
		hasADA := false
		hasLCN := false
		for _, asset := range info.Amount {
			if asset.Unit == "lovelace" {
				hasADA = true
			} else if len(asset.Unit) > 56 { // Native asset
				hasLCN = true
			}
		}
		assert.True(t, hasADA, "Address should have ADA")
		assert.True(t, hasLCN, "Address should have LCN")
	})
}
