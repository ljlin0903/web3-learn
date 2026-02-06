package dex

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// DEXType represents the type of decentralized exchange
type DEXType string

const (
	UniswapV2 DEXType = "uniswap_v2"
	SushiSwap DEXType = "sushiswap"
	UniswapV3 DEXType = "uniswap_v3"
)

// Pool represents a liquidity pool on a DEX
type Pool struct {
	Address     common.Address // Pool contract address
	DEX         DEXType        // DEX type
	Token0      common.Address // First token address
	Token1      common.Address // Second token address
	Reserve0    *big.Int       // Reserve of token0
	Reserve1    *big.Int       // Reserve of token1
	Fee         int            // Fee in basis points (30 = 0.3%)
	LastUpdated int64          // Unix timestamp of last update
}

// Token represents an ERC20 token
type Token struct {
	Address  common.Address
	Symbol   string
	Name     string
	Decimals uint8
}

// Pair represents a trading pair
type Pair struct {
	Token0  *Token
	Token1  *Token
	Address common.Address // Pair contract address
	DEX     DEXType
}

// SwapEvent represents a Swap event from a DEX
type SwapEvent struct {
	TxHash      common.Hash
	BlockNumber uint64
	PairAddress common.Address
	Sender      common.Address
	To          common.Address
	Amount0In   *big.Int
	Amount1In   *big.Int
	Amount0Out  *big.Int
	Amount1Out  *big.Int
	Timestamp   int64
}

// Route represents a swap route through multiple pools
type Route struct {
	Pools       []*Pool
	Tokens      []common.Address
	Path        []common.Address // Token addresses in order
	AmountIn    *big.Int
	AmountOut   *big.Int
	PriceImpact *big.Float
}

// QuoteResult represents the result of a price quote
type QuoteResult struct {
	AmountOut    *big.Int
	PriceImpact  *big.Float
	Fee          *big.Int
	MinAmountOut *big.Int // After slippage
	Route        *Route
}

// DEXAdapter interface defines methods that all DEX adapters must implement
type DEXAdapter interface {
	// GetName returns the name of the DEX
	GetName() string

	// GetType returns the type of DEX
	GetType() DEXType

	// GetPool fetches pool information
	GetPool(token0, token1 common.Address) (*Pool, error)

	// GetReserves fetches current reserves of a pool
	GetReserves(pairAddress common.Address) (*big.Int, *big.Int, error)

	// GetAmountOut calculates output amount for a given input
	GetAmountOut(amountIn *big.Int, reserveIn, reserveOut *big.Int) *big.Int

	// GetAmountIn calculates required input for desired output
	GetAmountIn(amountOut *big.Int, reserveIn, reserveOut *big.Int) *big.Int

	// Quote provides a price quote for a swap
	Quote(amountIn *big.Int, tokenIn, tokenOut common.Address) (*QuoteResult, error)

	// GetRouterAddress returns the router contract address
	GetRouterAddress() common.Address

	// GetFactoryAddress returns the factory contract address
	GetFactoryAddress() common.Address
}
