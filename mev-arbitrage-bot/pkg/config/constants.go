package config

import "math/big"

// Protocol Constants
const (
	// Wei per Ether
	WeiPerEther = 1e18

	// Gwei per Ether
	GweiPerEther = 1e9

	// Uniswap V2 Fee (0.3%)
	UniswapV2FeeBps = 30

	// SushiSwap Fee (0.3%)
	SushiSwapFeeBps = 30

	// Aave Flash Loan Fee (0.09%)
	AaveFlashLoanFeeBps = 9

	// Basis Points per 100%
	BasisPointsPerPercent    = 100
	BasisPointsPer100Percent = 10000

	// Gas limits
	DefaultGasLimit   = 300000
	FlashLoanGasLimit = 500000

	// Block time (seconds)
	AverageBlockTime = 12
)

// Network Chain IDs
const (
	MainnetChainID = 1
	SepoliaChainID = 11155111
	GoerliChainID  = 5
)

// Common token decimals
const (
	ETHDecimals  = 18
	WETHDecimals = 18
	USDCDecimals = 6
	DAIDecimals  = 18
	USDTDecimals = 6
)

// Pre-calculated big.Int values for performance
var (
	BigInt0     = big.NewInt(0)
	BigInt1     = big.NewInt(1)
	BigInt10    = big.NewInt(10)
	BigInt100   = big.NewInt(100)
	BigInt1000  = big.NewInt(1000)
	BigInt10000 = big.NewInt(10000)

	// Fee multipliers for AMM calculations
	FeeMultiplier997  = big.NewInt(997) // 10000 - 30 (0.3% fee)
	FeeMultiplier1000 = big.NewInt(1000)

	// Wei conversion
	WeiPerEtherBigInt  = new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	GweiPerEtherBigInt = new(big.Int).Exp(big.NewInt(10), big.NewInt(9), nil)
)

// Error messages
const (
	ErrInsufficientLiquidity = "insufficient liquidity"
	ErrExcessiveSlippage     = "slippage exceeds maximum"
	ErrUnprofitableTrade     = "trade is unprofitable"
	ErrGasPriceTooHigh       = "gas price exceeds maximum"
	ErrInvalidPath           = "invalid arbitrage path"
	ErrContractReverted      = "contract execution reverted"
	ErrTransactionFailed     = "transaction failed"
)
