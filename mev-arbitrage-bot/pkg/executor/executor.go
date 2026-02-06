package executor

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	log "github.com/sirupsen/logrus"

	"github.com/ljlin/mev-arbitrage-bot/pkg/config"
	"github.com/ljlin/mev-arbitrage-bot/pkg/flashbots"
	"github.com/ljlin/mev-arbitrage-bot/pkg/strategy"
)

// Executor handles transaction execution
// Executor 处理交易执行
//
// 主要职责:
// 1. 构建交易
// 2. 签名交易
// 3. 通过 Flashbots 或普通方式发送
// 4. 跟踪交易状态
type Executor struct {
	ethClient       *ethclient.Client
	flashbotsClient *flashbots.FlashbotsClient
	privateKey      *ecdsa.PrivateKey
	publicAddress   common.Address
	config          *config.Config
	nonce           uint64
	gasPrice        *big.Int
}

// NewExecutor creates a new executor
// NewExecutor 创建新的执行器
func NewExecutor(
	ethClient *ethclient.Client,
	flashbotsClient *flashbots.FlashbotsClient,
	cfg *config.Config,
) (*Executor, error) {
	// 解析私钥
	privateKey, err := crypto.HexToECDSA(cfg.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	executor := &Executor{
		ethClient:       ethClient,
		flashbotsClient: flashbotsClient,
		privateKey:      privateKey,
		publicAddress:   cfg.PublicAddress,
		config:          cfg,
	}

	// 初始化 nonce
	if err := executor.updateNonce(); err != nil {
		return nil, fmt.Errorf("failed to initialize nonce: %w", err)
	}

	// 初始化 gas price
	if err := executor.updateGasPrice(); err != nil {
		return nil, fmt.Errorf("failed to initialize gas price: %w", err)
	}

	log.Info("Transaction executor initialized")
	return executor, nil
}

// ExecuteArbitrage executes an arbitrage opportunity
// ExecuteArbitrage 执行套利机会
//
// 执行流程:
// 1. 验证机会是否仍然有效
// 2. 构建交易
// 3. 如果启用 Flashbots，通过 Flashbots 发送
// 4. 否则通过普通方式发送
// 5. 等待交易确认
// 6. 返回执行结果
func (e *Executor) ExecuteArbitrage(ctx context.Context, opportunity *strategy.ArbitrageOpportunity) error {
	log.Infof("Executing arbitrage opportunity: %s", opportunity.Path.ID[:8])

	// 检查是否为 Dry Run 模式
	if e.config.DryRun {
		log.Warn("🧪 DRY RUN MODE - Transaction not sent")
		e.logArbitrageDetails(opportunity)
		return nil
	}

	// 更新 Gas 价格
	if err := e.updateGasPrice(); err != nil {
		return fmt.Errorf("failed to update gas price: %w", err)
	}

	// 检查 Gas 价格是否超过最大值
	maxGasPrice := new(big.Int).Mul(
		big.NewInt(int64(e.config.MaxGasPriceGwei)),
		big.NewInt(1e9),
	)
	if e.gasPrice.Cmp(maxGasPrice) > 0 {
		return fmt.Errorf("gas price %s exceeds maximum %s",
			e.gasPrice.String(), maxGasPrice.String())
	}

	// 构建交易
	tx, err := e.buildArbitrageTx(opportunity)
	if err != nil {
		return fmt.Errorf("failed to build transaction: %w", err)
	}

	// 选择发送方式
	if e.config.EnableFlashbots && e.flashbotsClient != nil {
		return e.sendViaFlashbots(ctx, tx, opportunity)
	}

	return e.sendViaMempool(ctx, tx, opportunity)
}

// buildArbitrageTx builds an arbitrage transaction
// buildArbitrageTx 构建套利交易
//
// 交易内容:
// - To: 套利合约地址
// - Data: 编码的函数调用（包含交易路径、金额等）
// - Value: 0 (使用闪电贷，不需要自有资金)
// - Gas: 估算的 Gas 限制
// - GasPrice: 当前 Gas 价格
func (e *Executor) buildArbitrageTx(opportunity *strategy.ArbitrageOpportunity) (*types.Transaction, error) {
	log.Debug("Building arbitrage transaction")

	// TODO: 实现完整的交易构建
	// 实际需要:
	// 1. 调用智能合约的套利函数
	// 2. 编码函数参数（路径、金额等）
	// 3. 估算 Gas
	// 4. 签名交易

	// 当前为示例实现
	nonce := e.nonce
	gasLimit := uint64(500000) // 套利交易通常需要较高 Gas

	// 构建交易
	tx := types.NewTransaction(
		nonce,
		e.config.ArbitrageContract,
		big.NewInt(0), // Value = 0 (使用闪电贷)
		gasLimit,
		e.gasPrice,
		[]byte{}, // TODO: 编码实际的函数调用
	)

	// 签名交易
	chainID, err := e.ethClient.ChainID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), e.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	// 增加 nonce
	e.nonce++

	log.Debugf("Transaction built: hash=%s, nonce=%d", signedTx.Hash().Hex(), nonce)
	return signedTx, nil
}

// sendViaFlashbots sends transaction via Flashbots
// sendViaFlashbots 通过 Flashbots 发送交易
//
// 优势:
// - 防止被抢跑
// - 失败不消耗 Gas
// - 可以获得 MEV 收益的一部分
func (e *Executor) sendViaFlashbots(
	ctx context.Context,
	tx *types.Transaction,
	opportunity *strategy.ArbitrageOpportunity,
) error {
	log.Info("📡 Sending transaction via Flashbots")

	// 获取当前区块号
	blockNumber, err := e.ethClient.BlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("failed to get block number: %w", err)
	}

	// 目标下一个区块
	targetBlock := blockNumber + 1

	// 构建 Bundle
	bundle := e.flashbotsClient.BuildBundle([]*types.Transaction{tx}, targetBlock)

	// 先模拟
	simResult, err := e.flashbotsClient.SimulateBundle(ctx, bundle)
	if err != nil {
		return fmt.Errorf("bundle simulation failed: %w", err)
	}

	if !simResult.Success {
		return fmt.Errorf("bundle simulation failed: not profitable")
	}

	log.Infof("Simulation successful: gas=%d, profit=%s ETH",
		simResult.GasUsed, simResult.CoinbaseDiff.String())

	// 发送 Bundle
	response, err := e.flashbotsClient.SendBundle(ctx, bundle)
	if err != nil {
		return fmt.Errorf("failed to send bundle: %w", err)
	}

	if !response.Success {
		return fmt.Errorf("bundle rejected: %s", response.Error)
	}

	log.Infof("✅ Bundle sent successfully: hash=%s", response.BundleHash.Hex())
	return nil
}

// sendViaMempool sends transaction via normal mempool
// sendViaMempool 通过普通交易池发送交易
//
// 注意: 可能被 MEV 机器人抢跑！
func (e *Executor) sendViaMempool(
	ctx context.Context,
	tx *types.Transaction,
	opportunity *strategy.ArbitrageOpportunity,
) error {
	log.Warn("⚠️  Sending transaction via public mempool (may be front-run)")

	// 发送交易
	err := e.ethClient.SendTransaction(ctx, tx)
	if err != nil {
		return fmt.Errorf("failed to send transaction: %w", err)
	}

	log.Infof("Transaction sent: %s", tx.Hash().Hex())

	// 等待确认
	receipt, err := e.waitForReceipt(ctx, tx.Hash())
	if err != nil {
		return fmt.Errorf("transaction failed: %w", err)
	}

	if receipt.Status == 1 {
		log.Infof("✅ Transaction confirmed: block=%d, gas=%d",
			receipt.BlockNumber.Uint64(), receipt.GasUsed)
	} else {
		log.Errorf("❌ Transaction reverted: block=%d", receipt.BlockNumber.Uint64())
		return fmt.Errorf("transaction reverted")
	}

	return nil
}

// waitForReceipt waits for transaction receipt
// waitForReceipt 等待交易收据
func (e *Executor) waitForReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	timeout := time.After(2 * time.Minute)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return nil, fmt.Errorf("timeout waiting for transaction")
		case <-ticker.C:
			receipt, err := e.ethClient.TransactionReceipt(ctx, txHash)
			if err == nil {
				return receipt, nil
			}
		}
	}
}

// updateNonce updates the account nonce
// updateNonce 更新账户 nonce
func (e *Executor) updateNonce() error {
	nonce, err := e.ethClient.PendingNonceAt(context.Background(), e.publicAddress)
	if err != nil {
		return err
	}
	e.nonce = nonce
	return nil
}

// updateGasPrice updates the gas price
// updateGasPrice 更新 Gas 价格
func (e *Executor) updateGasPrice() error {
	gasPrice, err := e.ethClient.SuggestGasPrice(context.Background())
	if err != nil {
		return err
	}

	// 应用倍数
	multiplier := big.NewFloat(e.config.GasPriceMultiplier)
	gasPriceFloat := new(big.Float).SetInt(gasPrice)
	gasPriceFloat.Mul(gasPriceFloat, multiplier)

	adjustedGasPrice, _ := gasPriceFloat.Int(nil)
	e.gasPrice = adjustedGasPrice

	return nil
}

// logArbitrageDetails logs arbitrage details for dry run
// logArbitrageDetails 记录套利详情（用于模拟模式）
func (e *Executor) logArbitrageDetails(opportunity *strategy.ArbitrageOpportunity) {
	path := opportunity.Path
	log.Info("========================================")
	log.Info("Arbitrage Opportunity Details")
	log.Info("========================================")
	log.Infof("ID: %s", path.ID)
	log.Infof("Profit: %s ETH (%.2f%%)",
		path.ProfitETH.Text('f', 6),
		float64(path.ProfitBps)/100)
	log.Infof("Net Profit: %s ETH (%.2f%%)",
		path.ProfitETH.Text('f', 6),
		float64(path.NetProfitBps)/100)
	log.Infof("Start Amount: %s", path.StartAmount.String())
	log.Infof("End Amount: %s", path.EndAmount.String())
	log.Info("Path:")
	for i, token := range path.Tokens {
		log.Infof("  %d. %s", i+1, token.Hex())
	}
	log.Info("========================================")
}
