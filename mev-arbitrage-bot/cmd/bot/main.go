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

	// 加载配置
	log.Info("📋 正在加载配置...")
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("❌ 配置加载失败: %v", err)
	}

	cfg.PrintConfig()

	// 初始化区块链客户端
	log.Info("🔗 正在初始化区块链连接...")
	client, err := blockchain.NewClient(cfg)
	if err != nil {
		log.Fatalf("❌ 区块链客户端初始化失败: %v", err)
	}
	defer client.Close()

	// 验证连接
	log.Info("🔍 正在验证区块链连接...")
	if err := verifyConnection(client, cfg); err != nil {
		log.Fatalf("❌ 连接验证失败: %v", err)
	}

	log.Info("✅ 所有系统就绪！")

	if cfg.DryRun {
		log.Warn("🧪 当前运行模式: 模拟模式 - 不会发送真实交易")
	} else {
		log.Warn("⚠️  当前运行模式: 真实模式 - 将发送真实交易！")
	}

	// 设置优雅关闭
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 初始化模块
	log.Info("⚙️  正在初始化套利模块...")
	modules, err := initializeModules(client, cfg)
	if err != nil {
		log.Fatalf("❌ 模块初始化失败: %v", err)
	}

	// 开始监控
	log.Info("👀 正在启动池子监控...")
	modules.poolMonitor.Start()

	log.Info("🚀 套利机器人已启动！")
	log.Info("   正在监控 DEX 池子，寻找套利机会")
	log.Info("   按 Ctrl+C 可停止运行")
	log.Info("")

	// Main arbitrage loop
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go runArbitrageLoop(ctx, cfg, modules)

	// 等待关闭信号
	<-sigChan
	cancel()

	// 停止监控
	log.Info("🛑 正在停止池子监控...")
	modules.poolMonitor.Stop()

	log.Info("\n👋 正在优雅关闭...")
	log.Info("✅ 机器人已成功停止")
}

func verifyConnection(client *blockchain.Client, cfg *config.Config) error {
	// 获取链 ID
	chainID, err := client.GetChainID()
	if err != nil {
		return fmt.Errorf("获取链 ID 失败: %w", err)
	}
	log.Infof("✅ 已连接到网络，链 ID: %s", chainID.String())

	// 获取最新区块
	blockNumber, err := client.GetBlockNumber()
	if err != nil {
		return fmt.Errorf("获取区块高度失败: %w", err)
	}
	log.Infof("📦 最新区块高度: %d", blockNumber)

	// 获取账户余额
	balance, err := client.GetBalance(cfg.PublicAddress)
	if err != nil {
		return fmt.Errorf("获取余额失败: %w", err)
	}

	ethBalance := utils.WeiToEther(balance)
	log.Infof("💰 账户余额: %s ETH", ethBalance.Text('f', 6))

	// 验证余额是否充足
	if balance.Cmp(config.BigInt0) == 0 {
		log.Warn("⚠️  账户余额为零！")
	}

	// 获取 Gas 价格
	gasPrice, err := client.GetGasPrice()
	if err != nil {
		return fmt.Errorf("获取 Gas 价格失败: %w", err)
	}
	gasPriceGwei := utils.WeiToGwei(gasPrice)
	log.Infof("⛽ 当前 Gas 价格: %d Gwei", gasPriceGwei)

	return nil
}

func printBanner() {
	banner := `
╔══════════════════════════════════════════════════════════╗
║                                                          ║
║         🤖  MEV 套利机器人  🤖                        ║
║                                                          ║
║         版本: %-10s                              ║
║         作者: ljlin                                    ║
║         网络: 以太坊 (Sepolia/主网)                  ║
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

// initializeModules 初始化所有机器人模块
func initializeModules(client *blockchain.Client, cfg *config.Config) (*BotModules, error) {
	modules := &BotModules{}

	// 获取 HTTP 客户端用于合约交互
	httpClient := client.GetHTTPClient()

	// 初始化 DEX 适配器
	log.Info("🔌 正在初始化 DEX 适配器...")
	uniswapAdapter, err := dex.NewUniswapV2Adapter(httpClient, cfg.UniswapV2Router)
	if err != nil {
		return nil, fmt.Errorf("创建 Uniswap 适配器失败: %w", err)
	}

	// 初始化池子监控器
	log.Info("📊 正在初始化池子监控器...")
	modules.poolMonitor = dex.NewPoolMonitor(httpClient, cfg)
	modules.poolMonitor.RegisterAdapter(uniswapAdapter)

	// 添加要监控的池子
	if err := addMonitoredPools(modules.poolMonitor, cfg); err != nil {
		return nil, fmt.Errorf("添加监控池子失败: %w", err)
	}

	// 初始化套利查找器
	log.Info("🎯 正在初始化套利策略...")
	modules.arbitrageFinder = strategy.NewArbitrageFinder(
		modules.poolMonitor,
		cfg,
	)

	// 初始化 Flashbots（如果启用）
	if cfg.EnableFlashbots {
		log.Info("🛡️  正在初始化 Flashbots 客户端...")
		fbClient, err := flashbots.NewFlashbotsClient(httpClient, cfg)
		if err != nil {
			log.Warnf("Flashbots 初始化失败: %v", err)
		} else {
			modules.flashbotsClient = fbClient
			log.Info("✅ Flashbots 客户端初始化成功")
		}
	}

	// 初始化执行器
	log.Info("⚙️  正在初始化交易执行器...")
	modules.executor, err = executor.NewExecutor(httpClient, modules.flashbotsClient, cfg)
	if err != nil {
		return nil, fmt.Errorf("创建执行器失败: %w", err)
	}

	log.Info("✅ 所有模块初始化成功")
	return modules, nil
}

// addMonitoredPools 添加要监控的池子
func addMonitoredPools(monitor *dex.PoolMonitor, cfg *config.Config) error {
	log.Info("👀 正在添加监控池子...")

	// 定义要监控的代币对
	pairs := []struct {
		token0  common.Address
		token1  common.Address
		dexType dex.DEXType
	}{
		// WETH/USDC 交易对
		{cfg.WETHAddress, cfg.USDCAddress, dex.UniswapV2},
		// WETH/DAI 交易对
		{cfg.WETHAddress, cfg.DAIAddress, dex.UniswapV2},
		// USDC/DAI 交易对
		{cfg.USDCAddress, cfg.DAIAddress, dex.UniswapV2},
	}

	for _, pair := range pairs {
		// 跳过零地址
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
		log.Infof("✅ 正在监控 %s 池子: %s", pair.dexType, pool.Address.Hex()[:10]+"...")
	}

	allPools := monitor.GetAllPools()
	log.Infof("📊 当前监控 %d 个池子", len(allPools))

	if len(allPools) == 0 {
		return fmt.Errorf("没有可监控的池子")
	}

	return nil
}

// runArbitrageLoop 运行主套利检测和执行循环
func runArbitrageLoop(ctx context.Context, cfg *config.Config, modules *BotModules) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	log.Info("🔄 套利检测循环已启动...")

	for {
		select {
		case <-ctx.Done():
			log.Info("🛑 套利循环已停止")
			return

		case <-ticker.C:
			// 搜索套利机会
			opportunities := searchArbitrageOpportunities(cfg, modules)

			if len(opportunities) == 0 {
				log.Debug("🔍 未找到盈利套利机会")
				continue
			}

			// 执行最佳机会
			best := opportunities[0]
			log.Infof("🎯 发现套利机会！利润: %s ETH (%.2f%%)",
				best.Path.ProfitETH.Text('f', 6),
				float64(best.Path.NetProfitBps)/100)

			if err := modules.executor.ExecuteArbitrage(ctx, best); err != nil {
				log.Errorf("❌ 执行套利失败: %v", err)
			}
		}
	}
}

// searchArbitrageOpportunities 搜索盈利套利路径
func searchArbitrageOpportunities(cfg *config.Config, modules *BotModules) []*strategy.ArbitrageOpportunity {
	// 使用 WETH 作为三角套利的基础代币
	if cfg.WETHAddress == (common.Address{}) {
		return nil
	}

	paths, err := modules.arbitrageFinder.FindTriangleArbitrage(cfg.WETHAddress)
	if err != nil {
		log.Debugf("套利搜索错误: %v", err)
		return nil
	}

	if len(paths) == 0 {
		return nil
	}

	// 转换为套利机会
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
