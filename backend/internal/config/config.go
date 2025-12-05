package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	// Server
	Env  string
	Port string
	Host string

	// Database
	MongoDBURI      string
	MongoDBDatabase string

	// Cardano
	CardanoNetwork      string
	BlockfrostProjectID string
	BlockfrostAPIURL    string

	// Token
	LCNPolicyID             string
	LCNAssetName            string
	GovernanceWalletAddress string

	// Vault
	VaultAddr       string
	VaultToken      string
	VaultTransitKey string

	// JWT
	JWTPrivateKeyPath  string
	JWTPublicKeyPath   string
	JWTExpirationHours int

	// Security
	BcryptCost int
	AESKeySize int

	// Rate Limiting
	RateLimitPerIP   int
	RateLimitPerUser int

	// Transaction Settings
	MinADAOutput          uint64
	FeeA                  uint64
	FeeB                  uint64
	FeeBufferMultiplier   float64
	ConfirmationsRequired int
	WalletSeedADA         uint64

	// Settlement
	ExchangeRateLCNETB            float64
	SettlementProcessingTimeHours int

	// Monitoring
	PrometheusPort string
	SentryDSN      string

	// Redis
	RedisURL string

	// Logging
	LogLevel  string
	LogFormat string
}

// Load loads configuration from environment variables
func Load() *Config {
	// Load .env file if it exists (development)
	_ = godotenv.Load()

	cfg := &Config{
		// Server
		Env:  getEnv("ENV", "development"),
		Port: getEnv("PORT", "8080"),
		Host: getEnv("HOST", "0.0.0.0"),

		// Database
		MongoDBURI:      getEnv("MONGODB_URI", "mongodb://localhost:27017"),
		MongoDBDatabase: getEnv("MONGODB_DATABASE", "loyalcoin_dev"),

		// Cardano
		CardanoNetwork:      getEnv("CARDANO_NETWORK", "testnet"),
		BlockfrostProjectID: getEnv("BLOCKFROST_PROJECT_ID", ""),
		BlockfrostAPIURL:    getEnv("BLOCKFROST_API_URL", "https://cardano-preprod.blockfrost.io/api/v0"),

		// Token
		LCNPolicyID:             getEnv("LCN_POLICY_ID", ""),
		LCNAssetName:            getEnv("LCN_ASSET_NAME", "4c434e"),
		GovernanceWalletAddress: getEnv("GOVERNANCE_WALLET_ADDRESS", ""),

		// Vault
		VaultAddr:       getEnv("VAULT_ADDR", "http://localhost:8200"),
		VaultToken:      getEnv("VAULT_TOKEN", ""),
		VaultTransitKey: getEnv("VAULT_TRANSIT_KEY", "lcn-keys"),

		// JWT
		JWTPrivateKeyPath:  getEnv("JWT_PRIVATE_KEY_PATH", "./keys/jwt_private.pem"),
		JWTPublicKeyPath:   getEnv("JWT_PUBLIC_KEY_PATH", "./keys/jwt_public.pem"),
		JWTExpirationHours: getEnvAsInt("JWT_EXPIRATION_HOURS", 24),

		// Security
		BcryptCost: getEnvAsInt("BCRYPT_COST", 12),
		AESKeySize: getEnvAsInt("AES_KEY_SIZE", 32),

		// Rate Limiting
		RateLimitPerIP:   getEnvAsInt("RATE_LIMIT_PER_IP", 100),
		RateLimitPerUser: getEnvAsInt("RATE_LIMIT_PER_USER", 30),

		// Transaction Settings
		MinADAOutput:          getEnvAsUint64("MIN_ADA_OUTPUT", 1200000),
		FeeA:                  getEnvAsUint64("FEE_A", 155381),
		FeeB:                  getEnvAsUint64("FEE_B", 44),
		FeeBufferMultiplier:   getEnvAsFloat64("FEE_BUFFER_MULTIPLIER", 1.2),
		ConfirmationsRequired: getEnvAsInt("CONFIRMATIONS_REQUIRED", 3),
		WalletSeedADA:         getEnvAsUint64("WALLET_SEED_ADA", 5000000),

		// Settlement
		ExchangeRateLCNETB:            getEnvAsFloat64("EXCHANGE_RATE_LCN_ETB", 1.0),
		SettlementProcessingTimeHours: getEnvAsInt("SETTLEMENT_PROCESSING_TIME_HOURS", 48),

		// Monitoring
		PrometheusPort: getEnv("PROMETHEUS_PORT", "9090"),
		SentryDSN:      getEnv("SENTRY_DSN", ""),

		// Redis
		RedisURL: getEnv("REDIS_URL", "redis://localhost:6379"),

		// Logging
		LogLevel:  getEnv("LOG_LEVEL", "info"),
		LogFormat: getEnv("LOG_FORMAT", "json"),
	}

	// Validate required fields
	if cfg.BlockfrostProjectID == "" {
		log.Fatal("BLOCKFROST_PROJECT_ID is required")
	}

	return cfg
}

// Helper functions
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return fallback
}

func getEnvAsUint64(key string, fallback uint64) uint64 {
	if value := os.Getenv(key); value != "" {
		if uint64Val, err := strconv.ParseUint(value, 10, 64); err == nil {
			return uint64Val
		}
	}
	return fallback
}

func getEnvAsFloat64(key string, fallback float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			return floatVal
		}
	}
	return fallback
}
