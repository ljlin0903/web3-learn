package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/common"
	log "github.com/sirupsen/logrus"

	"github.com/ljlin/mev-arbitrage-bot/pkg/blockchain"
	"github.com/ljlin/mev-arbitrage-bot/pkg/config"
	"github.com/ljlin/mev-arbitrage-bot/pkg/dex"
	"github.com/ljlin/mev-arbitrage-bot/pkg/executor"
	"github.com/ljlin/mev-arbitrage-bot/pkg/flashbots"
	"github.com/ljlin/mev-arbitrage-bot/pkg/strategy"
	"github.com/ljlin/mev-arbitrage-bot/pkg/utils"
)

const (
	Version = "1.0.0"
	AppName = "MEV Arbitrage Bot"
)

func main() {
	printBanner()

	// Load configuration
	log.Info("Loading configuration...")
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	cfg.PrintConfig()

	// Initialize blockchain client
	log.Info("Initializing blockchain client...")
	client, err := blockchain.NewClient(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize blockchain client: %v", err)
	}
	defer client.Close()

	// Verify connection
	log.Info("Verifying blockchain connection...")
	if err := verifyConnection(client, cfg); err != nil {
		log.Fatalf("Connection verification failed: %v", err)
	}

	log.Info("✅ All systems operational!")

	if cfg.DryRun {
		log.Warn("🧪 Running in DRY RUN mode - No real transactions will be sent")
	} else {
		log.Warn("⚠️  Running in LIVE mode - Real transactions will be sent!")
	}

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Initialize modules
	log.Info("Initializing arbitrage modules...")
	modules, err := initializeModules(client, cfg)
	if err != nil {
		log.Fatalf("Failed to initialize modules: %v", err)
	}

	// Start monitoring
	log.Info("Starting pool monitoring...")
	modules.poolMonitor.Start()

	log.Info("🚀 Arbitrage bot is running...")
	log.Info("   Monitoring DEX pools for arbitrage opportunities")
	log.Info("   Press Ctrl+C to stop")
	log.Info("")

	// Main arbitrage loop
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go runArbitrageLoop(ctx, cfg, modules)

	// Wait for shutdown signal
	<-sigChan
	cancel()

	// Stop monitoring
	log.Info("Stopping pool monitoring...")
	modules.poolMonitor.Stop()

	log.Info("\n👋 Shutting down gracefully...")
	log.Info("✅ Bot stopped successfully")
}

func verifyConnection(client *blockchain.Client, cfg *config.Config) error {
	// Get chain ID
	chainID, err := client.GetChainID()
	if err != nil {
		return fmt.Errorf("failed to get chain ID: %w", err)
	}
	log.Infof("Connected to network with Chain ID: %s", chainID.String())

	// Get latest block
	blockNumber, err := client.GetBlockNumber()
	if err != nil {
		return fmt.Errorf("failed to get block number: %w", err)
	}
	log.Infof("Latest block number: %d", blockNumber)

	// Get account balance
	balance, err := client.GetBalance(cfg.PublicAddress)
	if err != nil {
		return fmt.Errorf("failed to get balance: %w", err)
	}

	ethBalance := utils.WeiToEther(balance)
	log.Infof("Account balance: %s ETH", ethBalance.Text('f', 6))

	// Verify sufficient balance
	if balance.Cmp(config.BigInt0) == 0 {
		log.Warn("⚠️  Account has zero balance!")
	}

	// Get gas price
	gasPrice, err := client.GetGasPrice()
	if err != nil {
		return fmt.Errorf("failed to get gas price: %w", err)
	}
	gasPriceGwei := utils.WeiToGwei(gasPrice)
	log.Infof("Current gas price: %d Gwei", gasPriceGwei)

	return nil
}

func printBanner() {
	banner := `
╔══════════════════════════════════════════════════════════╗
║                                                          ║
║         🤖  MEV ARBITRAGE BOT  🤖                        ║
║                                                          ║
║         Version: %-10s                              ║
║         Author: ljlin                                    ║
║         Network: Ethereum (Sepolia/Mainnet)              ║
║                                                          ║
╚══════════════════════════════════════════════════════════╝
`
	fmt.Printf(banner, Version)
	fmt.Println()
}

// BotModules holds all initialized modules
type BotModules struct {
	poolMonitor     *dex.PoolMonitor
	arbitrageFinder *strategy.ArbitrageFinder
	executor        *executor.Executor
	flashbotsClient *flashbots.FlashbotsClient
}

// initializeModules initializes all bot modules
func initializeModules(client *blockchain.Client, cfg *config.Config) (*BotModules, error) {
	modules := &BotModules{}

	// Get HTTP client for contract interactions
	httpClient := client.GetHTTPClient()

	// Initialize DEX adapters
	log.Info("Initializing DEX adapters...")
	uniswapAdapter, err := dex.NewUniswapV2Adapter(httpClient, cfg.UniswapV2Router)
	if err != nil {
		return nil, fmt.Errorf("failed to create Uniswap adapter: %w", err)
	}

	// Initialize pool monitor
	log.Info("Initializing pool monitor...")
	modules.poolMonitor = dex.NewPoolMonitor(httpClient, cfg)
	modules.poolMonitor.RegisterAdapter(uniswapAdapter)

	// Add pools to monitor
	if err := addMonitoredPools(modules.poolMonitor, cfg); err != nil {
		return nil, fmt.Errorf("failed to add monitored pools: %w", err)
	}

	// Initialize arbitrage finder
	log.Info("Initializing arbitrage strategy...")
	modules.arbitrageFinder = strategy.NewArbitrageFinder(
		modules.poolMonitor,
		cfg,
	)

	// Initialize Flashbots (if enabled)
	if cfg.EnableFlashbots {
		log.Info("Initializing Flashbots client...")
		fbClient, err := flashbots.NewFlashbotsClient(httpClient, cfg)
		if err != nil {
			log.Warnf("Failed to initialize Flashbots: %v", err)
		} else {
			modules.flashbotsClient = fbClient
			log.Info("✅ Flashbots client initialized")
		}
	}

	// Initialize executor
	log.Info("Initializing transaction executor...")
	modules.executor, err = executor.NewExecutor(httpClient, modules.flashbotsClient, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create executor: %w", err)
	}

	log.Info("✅ All modules initialized successfully")
	return modules, nil
}

// addMonitoredPools adds pools to the monitor
func addMonitoredPools(monitor *dex.PoolMonitor, cfg *config.Config) error {
	log.Info("Adding pools to monitor...")

	// Define token pairs to monitor
	pairs := []struct {
		token0  common.Address
		token1  common.Address
		dexType dex.DEXType
	}{
		// WETH/USDC pairs
		{cfg.WETHAddress, cfg.USDCAddress, dex.UniswapV2},
		// WETH/DAI pairs
		{cfg.WETHAddress, cfg.DAIAddress, dex.UniswapV2},
		// USDC/DAI pairs
		{cfg.USDCAddress, cfg.DAIAddress, dex.UniswapV2},
	}

	for _, pair := range pairs {
		// Skip if addresses are zero
		if pair.token0 == (common.Address{}) || pair.token1 == (common.Address{}) {
			continue
		}

		// Try to get pool for this token pair
		pool, err := monitor.GetPoolForTokens(pair.token0, pair.token1, pair.dexType)
		if err != nil {
			log.Warnf("Failed to get pool for %s/%s: %v", pair.token0.Hex()[:8], pair.token1.Hex()[:8], err)
			continue
		}

		// Pool already exists in monitor
		log.Infof("✅ Monitoring %s pool: %s", pair.dexType, pool.Address.Hex()[:10]+"...")
	}

	allPools := monitor.GetAllPools()
	log.Infof("Monitoring %d pools", len(allPools))

	if len(allPools) == 0 {
		return fmt.Errorf("no pools available to monitor")
	}

	return nil
}

// runArbitrageLoop runs the main arbitrage detection and execution loop
func runArbitrageLoop(ctx context.Context, cfg *config.Config, modules *BotModules) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	log.Info("Starting arbitrage detection loop...")

	for {
		select {
		case <-ctx.Done():
			log.Info("Arbitrage loop stopped")
			return

		case <-ticker.C:
			// Search for arbitrage opportunities
			opportunities := searchArbitrageOpportunities(cfg, modules)

			if len(opportunities) == 0 {
				log.Debug("No profitable arbitrage opportunities found")
				continue
			}

			// Execute best opportunity
			best := opportunities[0]
			log.Infof("🎯 Found opportunity! Profit: %s ETH (%.2f%%)",
				best.Path.ProfitETH.Text('f', 6),
				float64(best.Path.NetProfitBps)/100)

			if err := modules.executor.ExecuteArbitrage(ctx, best); err != nil {
				log.Errorf("Failed to execute arbitrage: %v", err)
			}
		}
	}
}

// searchArbitrageOpportunities searches for profitable arbitrage paths
func searchArbitrageOpportunities(cfg *config.Config, modules *BotModules) []*strategy.ArbitrageOpportunity {
	// Use WETH as the base token for triangle arbitrage
	if cfg.WETHAddress == (common.Address{}) {
		return nil
	}

	paths, err := modules.arbitrageFinder.FindTriangleArbitrage(cfg.WETHAddress)
	if err != nil {
		log.Debugf("Arbitrage search error: %v", err)
		return nil
	}

	if len(paths) == 0 {
		return nil
	}

	// Convert to opportunities
	opportunities := make([]*strategy.ArbitrageOpportunity, 0, len(paths))
	for _, path := range paths {
		opp := &strategy.ArbitrageOpportunity{
			Path:         path,
			IsExecutable: true,
			Priority:     path.ProfitBps,
		}
		opportunities = append(opportunities, opp)
	}

	return opportunities
}
