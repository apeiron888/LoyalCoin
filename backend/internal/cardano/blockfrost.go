package cardano

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type BlockfrostClient struct {
	projectID  string
	baseURL    string
	httpClient *http.Client
}

func NewBlockfrostClient(projectID, baseURL string) *BlockfrostClient {
	return &BlockfrostClient{
		projectID: projectID,
		baseURL:   baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// UTXO for an address
type AddressUTXO struct {
	TxHash      string  `json:"tx_hash"`
	OutputIndex int     `json:"output_index"`
	Amount      []Asset `json:"amount"`
	Block       string  `json:"block"`
	DataHash    string  `json:"data_hash"`
}

// Cardano asset
type Asset struct {
	Unit     string `json:"unit"`
	Quantity string `json:"quantity"`
}

// Address information
type AddressInfo struct {
	Address string  `json:"address"`
	Amount  []Asset `json:"amount"`
	Type    string  `json:"type"`
}

type TxSubmitResponse struct {
	TxHash string `json:"tx_hash"`
}

func (c *BlockfrostClient) GetAddressUTXOs(address string) ([]AddressUTXO, error) {
	url := fmt.Sprintf("%s/addresses/%s/utxos", c.baseURL, address)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("project_id", c.projectID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Blockfrost: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("blockfrost returned status %d: %s", resp.StatusCode, string(body))
	}

	var utxos []AddressUTXO
	if err := json.NewDecoder(resp.Body).Decode(&utxos); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return utxos, nil
}

// Retrieves address information including balance
func (c *BlockfrostClient) GetAddressInfo(address string) (*AddressInfo, error) {
	url := fmt.Sprintf("%s/addresses/%s", c.baseURL, address)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("project_id", c.projectID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Blockfrost: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("blockfrost returned status %d: %s", resp.StatusCode, string(body))
	}

	var info AddressInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &info, nil
}

// Submits a signed transaction to the blockchain
func (c *BlockfrostClient) SubmitTransaction(signedTxCBOR []byte) (string, error) {
	url := fmt.Sprintf("%s/tx/submit", c.baseURL)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(signedTxCBOR))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("project_id", c.projectID)
	req.Header.Set("Content-Type", "application/cbor")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call Blockfrost: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("blockfrost returned status %d: %s", resp.StatusCode, string(body))
	}

	var result TxSubmitResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.TxHash, nil
}

// Retrieves current protocol parameters
func (c *BlockfrostClient) GetProtocolParameters() (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/epochs/latest/parameters", c.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("project_id", c.projectID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Blockfrost: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("blockfrost returned status %d: %s", resp.StatusCode, string(body))
	}

	var params map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&params); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return params, nil
}

// Health checks Blockfrost connectivity
func (c *BlockfrostClient) Health() error {
	url := fmt.Sprintf("%s/health", c.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call Blockfrost: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("blockfrost health check failed with status: %d", resp.StatusCode)
	}

	return nil
}

type TransactionDetails struct {
	TxHash      string `json:"hash"`
	Block       string `json:"block"`
	BlockHeight int64  `json:"block_height"`
	BlockTime   int64  `json:"block_time"`
	Slot        int64  `json:"slot"`
	Index       int    `json:"index"`
	Confirmed   bool
}

type BlockInfo struct {
	Height int64  `json:"height"`
	Hash   string `json:"hash"`
	Time   int64  `json:"time"`
}

// Retrieves transaction details by hash
func (c *BlockfrostClient) GetTransactionDetails(txHash string) (*TransactionDetails, error) {
	url := fmt.Sprintf("%s/txs/%s", c.baseURL, txHash)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("project_id", c.projectID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Blockfrost: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("transaction not found: %s", txHash)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("blockfrost returned status %d: %s", resp.StatusCode, string(body))
	}

	var details TransactionDetails
	if err := json.NewDecoder(resp.Body).Decode(&details); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	details.Confirmed = details.Block != ""

	return &details, nil
}

// Retrieves the latest block information
func (c *BlockfrostClient) GetLatestBlock() (*BlockInfo, error) {
	url := fmt.Sprintf("%s/blocks/latest", c.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("project_id", c.projectID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Blockfrost: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("blockfrost returned status %d: %s", resp.StatusCode, string(body))
	}

	var block BlockInfo
	if err := json.NewDecoder(resp.Body).Decode(&block); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &block, nil
}
