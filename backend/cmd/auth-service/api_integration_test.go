//go:build integration

package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const baseURL = "http://localhost:8080"

type SignupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
	Username string `json:"username,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type IssueRequest struct {
	CustomerAddress string `json:"customer_address"`
	AmountLCN       uint64 `json:"amount_lcn"`
	Reference       string `json:"reference"`
}

func TestEndToEndFlow(t *testing.T) {
	// Wait for server to be ready
	time.Sleep(1 * time.Second)

	// Generate unique email for this test run
	timestamp := time.Now().UnixNano()

	// Test 1: Create Merchant
	t.Run("Create Merchant", func(t *testing.T) {
		req := SignupRequest{
			Email:    "test_merchant_" + string(rune(timestamp)) + "@test.com",
			Password: "Test123!",
			Role:     "MERCHANT",
		}

		body, _ := json.Marshal(req)
		resp, err := http.Post(baseURL+"/api/v1/auth/signup", "application/json", bytes.NewBuffer(body))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, "ok", result["status"])
	})

	// Test 2: Create Customer
	customerEmail := "test_customer_" + string(rune(timestamp)) + "@test.com"
	var customerAddress string

	t.Run("Create Customer", func(t *testing.T) {
		req := SignupRequest{
			Email:    customerEmail,
			Password: "Test123!",
			Role:     "CUSTOMER",
			Username: "testcustomer" + string(rune(timestamp)),
		}

		body, _ := json.Marshal(req)
		resp, err := http.Post(baseURL+"/api/v1/auth/signup", "application/json", bytes.NewBuffer(body))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		data := result["data"].(map[string]interface{})
		customerAddress = data["wallet_address"].(string)
		assert.NotEmpty(t, customerAddress)
	})

	// Test 3: Login and Get Balance
	t.Run("Login and Check Balance", func(t *testing.T) {
		req := LoginRequest{
			Email:    customerEmail,
			Password: "Test123!",
		}

		body, _ := json.Marshal(req)
		resp, err := http.Post(baseURL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(body))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		data := result["data"].(map[string]interface{})
		token := data["token"].(string)
		assert.NotEmpty(t, token)

		// Check balance
		balanceReq, _ := http.NewRequest("GET", baseURL+"/api/v1/wallet/balance", nil)
		balanceReq.Header.Set("Authorization", "Bearer "+token)

		balanceResp, err := http.DefaultClient.Do(balanceReq)
		require.NoError(t, err)
		defer balanceResp.Body.Close()

		assert.Equal(t, http.StatusOK, balanceResp.StatusCode)

		bodyBytes, _ := io.ReadAll(balanceResp.Body)
		t.Logf("Balance response: %s", string(bodyBytes))
	})
}
