package models

import "time"

type Role string

const (
	RoleAdmin    Role = "ADMIN"
	RoleMerchant Role = "MERCHANT"
	RoleCustomer Role = "CUSTOMER"
)

type UserStatus string

const (
	StatusActive              UserStatus = "ACTIVE"
	StatusSuspended           UserStatus = "SUSPENDED"
	StatusPendingVerification UserStatus = "PENDING_VERIFICATION"
)

type TxType string

const (
	TxTypeIssuance   TxType = "ISSUANCE"
	TxTypeRedemption TxType = "REDEMPTION"
	TxTypeAllocation TxType = "ALLOCATION"
	TxTypeMint       TxType = "MINT"
	TxTypeSettlement TxType = "SETTLEMENT"
)

type TxStatus string

const (
	TxStatusPending   TxStatus = "PENDING"
	TxStatusConfirmed TxStatus = "CONFIRMED"
	TxStatusFailed    TxStatus = "FAILED"
)

type SettlementStatus string

const (
	SettlementPending    SettlementStatus = "PENDING"
	SettlementApproved   SettlementStatus = "APPROVED"
	SettlementProcessing SettlementStatus = "PROCESSING"
	SettlementCompleted  SettlementStatus = "COMPLETED"
	SettlementRejected   SettlementStatus = "REJECTED"
)

// Cardano wallet
type Wallet struct {
	Address             string    `bson:"address" json:"address"`
	EncryptedPrivateKey string    `bson:"encrypted_private_key" json:"-"`
	PubKeyHex           string    `bson:"pub_key_hex,omitempty" json:"pub_key_hex,omitempty"`
	CreatedAt           time.Time `bson:"created_at" json:"created_at"`
}

// Merchant Bank Account Details
type BankAccount struct {
	AccountNumber string `bson:"account_number" json:"account_number"`
	BankName      string `bson:"bank_name" json:"bank_name"`
	AccountHolder string `bson:"account_holder" json:"account_holder"`
	Verified      bool   `bson:"verified" json:"verified"`
}

// Merchant Details
type Merchant struct {
	ID            string      `bson:"_id,omitempty" json:"id"`
	BusinessName  string      `bson:"business_name" json:"business_name"`
	Email         string      `bson:"email" json:"email"`
	PasswordHash  string      `bson:"password_hash" json:"-"`
	Role          Role        `bson:"role" json:"role"`
	Wallet        Wallet      `bson:"wallet" json:"wallet"`
	AllocationLCN uint64      `bson:"allocation_lcn" json:"allocation_lcn"`
	BalanceLCN    uint64      `bson:"balance_lcn" json:"balance_lcn"`
	BankAccount   BankAccount `bson:"bank_account" json:"bank_account"`
	Status        UserStatus  `bson:"status" json:"status"`
	CreatedAt     time.Time   `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time   `bson:"updated_at" json:"updated_at"`
}

// Customer Details
type Customer struct {
	ID           string    `bson:"_id,omitempty" json:"id"`
	Username     string    `bson:"username" json:"username"`
	Email        string    `bson:"email" json:"email"`
	Phone        string    `bson:"phone,omitempty" json:"phone,omitempty"`
	PasswordHash string    `bson:"password_hash" json:"-"`
	Wallet       Wallet    `bson:"wallet" json:"wallet"`
	CreatedAt    time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time `bson:"updated_at" json:"updated_at"`
}

// Blockchain transaction log
type TxLog struct {
	ID            string                 `bson:"_id,omitempty" json:"id"`
	TxHash        string                 `bson:"tx_hash" json:"tx_hash"`
	FromAddress   string                 `bson:"from_address" json:"from_address"`
	ToAddress     string                 `bson:"to_address" json:"to_address"`
	AmountLCN     uint64                 `bson:"amount_lcn" json:"amount_lcn"`
	AssetPolicyID string                 `bson:"asset_policy_id" json:"asset_policy_id"`
	AssetName     string                 `bson:"asset_name" json:"asset_name"`
	Type          TxType                 `bson:"type" json:"type"`
	Status        TxStatus               `bson:"status" json:"status"`
	BlockHeight   int64                  `bson:"block_height,omitempty" json:"block_height,omitempty"`
	SubmittedAt   time.Time              `bson:"submitted_at" json:"submitted_at"`
	ConfirmedAt   *time.Time             `bson:"confirmed_at,omitempty" json:"confirmed_at,omitempty"`
	Meta          map[string]interface{} `bson:"meta,omitempty" json:"meta,omitempty"`
}

// Merchant's LCN → ETB settlement request
type SettlementRequest struct {
	ID               string           `bson:"_id,omitempty" json:"id"`
	MerchantID       string           `bson:"merchant_id" json:"merchant_id"`
	AmountLCN        uint64           `bson:"amount_lcn" json:"amount_lcn"`
	AmountETB        float64          `bson:"amount_etb" json:"amount_etb"`
	ExchangeRate     float64          `bson:"exchange_rate" json:"exchange_rate"`
	Status           SettlementStatus `bson:"status" json:"status"`
	BankAccount      BankAccount      `bson:"bank_account" json:"bank_account"`
	RequestedAt      time.Time        `bson:"requested_at" json:"requested_at"`
	ApprovedAt       *time.Time       `bson:"approved_at,omitempty" json:"approved_at,omitempty"`
	ProcessedAt      *time.Time       `bson:"processed_at,omitempty" json:"processed_at,omitempty"`
	TxHash           string           `bson:"tx_hash,omitempty" json:"tx_hash,omitempty"`
	PaymentReference string           `bson:"payment_reference,omitempty" json:"payment_reference,omitempty"`
	AdminID          string           `bson:"admin_id,omitempty" json:"admin_id,omitempty"`
	AdminNotes       string           `bson:"admin_notes,omitempty" json:"admin_notes,omitempty"`
}

// Merchant's request to buy more LCN
type AllocationPurchase struct {
	ID                string     `bson:"_id,omitempty" json:"id"`
	MerchantID        string     `bson:"merchant_id" json:"merchant_id"`
	AmountLCN         uint64     `bson:"amount_lcn" json:"amount_lcn"`
	AmountETBPaid     float64    `bson:"amount_etb_paid" json:"amount_etb_paid"`
	PaymentMethod     string     `bson:"payment_method" json:"payment_method"`
	PaymentReference  string     `bson:"payment_reference" json:"payment_reference"`
	PaymentProofURL   string     `bson:"payment_proof_url,omitempty" json:"payment_proof_url,omitempty"`
	Status            string     `bson:"status" json:"status"` // PENDING, VERIFIED, CONFIRMED, REJECTED
	PurchasedAt       time.Time  `bson:"purchased_at" json:"purchased_at"`
	VerifiedAt        *time.Time `bson:"verified_at,omitempty" json:"verified_at,omitempty"`
	LCNTransferTxHash string     `bson:"lcn_transfer_tx_hash,omitempty" json:"lcn_transfer_tx_hash,omitempty"`
	AdminID           string     `bson:"admin_id,omitempty" json:"admin_id,omitempty"`
	AdminNotes        string     `bson:"admin_notes,omitempty" json:"admin_notes,omitempty"`
}

// UTXOAsset
type UTXOAsset struct {
	PolicyID  string `bson:"policy_id" json:"policy_id"`
	AssetName string `bson:"asset_name" json:"asset_name"`
	Quantity  uint64 `bson:"quantity" json:"quantity"`
}

// UTXOValue
type UTXOValue struct {
	Lovelace uint64      `bson:"lovelace" json:"lovelace"`
	Assets   []UTXOAsset `bson:"assets,omitempty" json:"assets,omitempty"`
}

// Unspent transaction output
type UTXO struct {
	TxHash    string    `bson:"tx_hash" json:"tx_hash"`
	Index     int       `bson:"index" json:"index"`
	Value     UTXOValue `bson:"value" json:"value"`
	FetchedAt time.Time `bson:"fetched_at" json:"fetched_at"`
}

// Cached UTXOs for an address
type UTXOCache struct {
	ID          string    `bson:"_id,omitempty" json:"id"`
	Address     string    `bson:"address" json:"address"`
	UTXOs       []UTXO    `bson:"utxos" json:"utxos"`
	LastFetched time.Time `bson:"last_fetched" json:"last_fetched"`
}

// Snapshot of the governance reserve
type GovernanceReserve struct {
	ID                      string    `bson:"_id,omitempty" json:"id"`
	SnapshotDate            time.Time `bson:"snapshot_date" json:"snapshot_date"`
	LCNTotalMinted          uint64    `bson:"lcn_total_minted" json:"lcn_total_minted"`
	LCNInGovernanceWallet   uint64    `bson:"lcn_in_governance_wallet" json:"lcn_in_governance_wallet"`
	LCNAllocatedToMerchants uint64    `bson:"lcn_allocated_to_merchants" json:"lcn_allocated_to_merchants"`
	LCNHeldByCustomers      uint64    `bson:"lcn_held_by_customers" json:"lcn_held_by_customers"`
	ETBReserveRequired      float64   `bson:"etb_reserve_required" json:"etb_reserve_required"`
	ETBReserveActual        float64   `bson:"etb_reserve_actual" json:"etb_reserve_actual"`
	ReserveRatio            float64   `bson:"reserve_ratio" json:"reserve_ratio"`
	TotalSettlementsPending float64   `bson:"total_settlements_pending" json:"total_settlements_pending"`
	CreatedAt               time.Time `bson:"created_at" json:"created_at"`
}
