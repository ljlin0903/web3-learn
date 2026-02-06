package config

import (
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

// Config holds all configuration for the arbitrage bot
type Config struct {
	// Network Configuration
	Network             string
	RPCHTTPSUrl         string
	RPCWSSUrl           string
	FlashbotsRelay      string
	FlashbotsSigningKey string

	// Wallet Configuration
	PrivateKey    string
	PublicAddress common.Address

	// Contract Addresses
	ArbitrageContract common.Address
	AavePoolProvider  common.Address

	// DEX Router Addresses
	UniswapV2Router common.Address
	SushiswapRouter common.Address

	// Token Addresses
	WETHAddress common.Address
	USDCAddress common.Address
	DAIAddress  common.Address

	// Strategy Parameters
	MinProfitBps       int
	MaxTradeAmountETH  *big.Float
	MinTradeAmountETH  *big.Float
	GasPriceMultiplier float64
	MaxGasPriceGwei    uint64

	// Monitoring Configuration
	LogLevel         string
	TelegramBotToken string
	TelegramChatID   string

	// Advanced Settings
	EnableFlashbots     bool
	DryRun              bool
	BlockConfirmations  uint64
	ConnectionTimeout   int
	MaxRetryAttempts    int
	PoolMonitorInterval int
}

var globalConfig *Config

// LoadConfig loads configuration from .env file
func LoadConfig() (*Config, error) {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Warn("No .env file found, using environment variables")
	}

	cfg := &Config{}

	// Network Configuration
	cfg.Network = getEnv("NETWORK", "sepolia")
	cfg.RPCHTTPSUrl = getEnv("RPC_HTTPS_URL", "")
	cfg.RPCWSSUrl = getEnv("RPC_WSS_URL", "")
	cfg.FlashbotsRelay = getEnv("FLASHBOTS_RELAY_URL", "https://relay.flashbots.net")
	cfg.FlashbotsSigningKey = getEnv("FLASHBOTS_RELAY_SIGNING_KEY", "")

	// Validate required fields
	if cfg.RPCHTTPSUrl == "" || cfg.RPCWSSUrl == "" {
		return nil, fmt.Errorf("RPC_HTTPS_URL and RPC_WSS_URL are required")
	}

	// Wallet Configuration
	cfg.PrivateKey = getEnv("PRIVATE_KEY", "")
	if cfg.PrivateKey == "" {
		return nil, fmt.Errorf("PRIVATE_KEY is required")
	}

	publicAddressStr := getEnv("PUBLIC_ADDRESS", "")
	if publicAddressStr == "" {
		return nil, fmt.Errorf("PUBLIC_ADDRESS is required")
	}
	cfg.PublicAddress = common.HexToAddress(publicAddressStr)

	// Contract Addresses
	cfg.ArbitrageContract = common.HexToAddress(getEnv("ARBITRAGE_CONTRACT_ADDRESS", ""))
	cfg.AavePoolProvider = common.HexToAddress(getEnv("AAVE_POOL_PROVIDER", ""))

	// DEX Router Addresses
	cfg.UniswapV2Router = common.HexToAddress(getEnv("UNISWAP_V2_ROUTER", ""))
	cfg.SushiswapRouter = common.HexToAddress(getEnv("SUSHISWAP_ROUTER", ""))

	// Token Addresses
	cfg.WETHAddress = common.HexToAddress(getEnv("WETH_ADDRESS", ""))
	cfg.USDCAddress = common.HexToAddress(getEnv("USDC_ADDRESS", ""))
	cfg.DAIAddress = common.HexToAddress(getEnv("DAI_ADDRESS", ""))

	// Strategy Parameters
	cfg.MinProfitBps = getEnvAsInt("MIN_PROFIT_BPS", 50)
	cfg.MaxTradeAmountETH = parseEther(getEnv("MAX_TRADE_AMOUNT_ETH", "10"))
	cfg.MinTradeAmountETH = parseEther(getEnv("MIN_TRADE_AMOUNT_ETH", "0.1"))
	cfg.GasPriceMultiplier = getEnvAsFloat64("GAS_PRICE_MULTIPLIER", 1.2)
	cfg.MaxGasPriceGwei = uint64(getEnvAsInt("MAX_GAS_PRICE_GWEI", 100))

	// Monitoring Configuration
	cfg.LogLevel = getEnv("LOG_LEVEL", "info")
	cfg.TelegramBotToken = getEnv("TELEGRAM_BOT_TOKEN", "")
	cfg.TelegramChatID = getEnv("TELEGRAM_CHAT_ID", "")

	// Advanced Settings
	cfg.EnableFlashbots = getEnvAsBool("ENABLE_FLASHBOTS", false)
	cfg.DryRun = getEnvAsBool("DRY_RUN", true)
	cfg.BlockConfirmations = uint64(getEnvAsInt("BLOCK_CONFIRMATIONS", 1))
	cfg.ConnectionTimeout = getEnvAsInt("CONNECTION_TIMEOUT", 10)
	cfg.MaxRetryAttempts = getEnvAsInt("MAX_RETRY_ATTEMPTS", 3)
	cfg.PoolMonitorInterval = getEnvAsInt("POOL_MONITOR_INTERVAL", 12)

	// Setup logging
	setupLogging(cfg.LogLevel)

	globalConfig = cfg
	return cfg, nil
}

// GetConfig returns the global configuration instance
func GetConfig() *Config {
	if globalConfig == nil {
		log.Fatal("Configuration not loaded. Call LoadConfig() first")
	}
	return globalConfig
}

// Helper functions

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Warnf("Invalid integer value for %s, using default: %d", key, defaultValue)
		return defaultValue
	}
	return value
}

func getEnvAsFloat64(key string, defaultValue float64) float64 {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		log.Warnf("Invalid float value for %s, using default: %f", key, defaultValue)
		return defaultValue
	}
	return value
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := strings.ToLower(os.Getenv(key))
	if valueStr == "" {
		return defaultValue
	}
	return valueStr == "true" || valueStr == "1" || valueStr == "yes"
}

func parseEther(value string) *big.Float {
	amount, ok := new(big.Float).SetString(value)
	if !ok {
		log.Warnf("Failed to parse ether amount: %s, using 0", value)
		return big.NewFloat(0)
	}
	return amount
}

func setupLogging(level string) {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	log.SetOutput(os.Stdout)

	switch strings.ToLower(level) {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn", "warning":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}
}

// PrintConfig prints the current configuration (sanitized)
func (c *Config) PrintConfig() {
	log.Info("==================== Configuration ====================")
	log.Infof("Network: %s", c.Network)
	log.Infof("RPC HTTPS: %s", maskURL(c.RPCHTTPSUrl))
	log.Infof("RPC WSS: %s", maskURL(c.RPCWSSUrl))
	log.Infof("Public Address: %s", c.PublicAddress.Hex())
	log.Infof("Arbitrage Contract: %s", c.ArbitrageContract.Hex())
	log.Infof("Min Profit BPS: %d (%.2f%%)", c.MinProfitBps, float64(c.MinProfitBps)/100)
	log.Infof("Max Trade Amount: %s ETH", c.MaxTradeAmountETH.Text('f', 2))
	log.Infof("Enable Flashbots: %v", c.EnableFlashbots)
	log.Infof("Dry Run Mode: %v", c.DryRun)
	log.Info("======================================================")
}

// maskURL masks sensitive parts of a URL
func maskURL(url string) string {
	if len(url) < 10 {
		return "***"
	}
	return url[:10] + "***" + url[len(url)-10:]
}
