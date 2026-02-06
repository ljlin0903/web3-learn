package dex

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	log "github.com/sirupsen/logrus"

	"github.com/ljlin/mev-arbitrage-bot/pkg/config"
	"github.com/ljlin/mev-arbitrage-bot/pkg/utils"
)

// Uniswap V2 Router ABI (simplified, only needed functions)
const UniswapV2RouterABI = `[
	{
		"constant":true,
		"inputs":[{"internalType":"uint256","name":"amountIn","type":"uint256"},{"internalType":"address[]","name":"path","type":"address[]"}],
		"name":"getAmountsOut",
		"outputs":[{"internalType":"uint256[]","name":"amounts","type":"uint256[]"}],
		"stateMutability":"view",
		"type":"function"
	},
	{
		"constant":true,
		"inputs":[],
		"name":"factory",
		"outputs":[{"internalType":"address","name":"","type":"address"}],
		"stateMutability":"view",
		"type":"function"
	}
]`

// Uniswap V2 Factory ABI (simplified)
const UniswapV2FactoryABI = `[
	{
		"constant":true,
		"inputs":[{"internalType":"address","name":"","type":"address"},{"internalType":"address","name":"","type":"address"}],
		"name":"getPair",
		"outputs":[{"internalType":"address","name":"","type":"address"}],
		"stateMutability":"view",
		"type":"function"
	}
]`

// Uniswap V2 Pair ABI (simplified)
const UniswapV2PairABI = `[
	{
		"constant":true,
		"inputs":[],
		"name":"getReserves",
		"outputs":[{"internalType":"uint112","name":"_reserve0","type":"uint112"},{"internalType":"uint112","name":"_reserve1","type":"uint112"},{"internalType":"uint32","name":"_blockTimestampLast","type":"uint32"}],
		"stateMutability":"view",
		"type":"function"
	},
	{
		"constant":true,
		"inputs":[],
		"name":"token0",
		"outputs":[{"internalType":"address","name":"","type":"address"}],
		"stateMutability":"view",
		"type":"function"
	},
	{
		"constant":true,
		"inputs":[],
		"name":"token1",
		"outputs":[{"internalType":"address","name":"","type":"address"}],
		"stateMutability":"view",
		"type":"function"
	}
]`

// UniswapV2Adapter implements DEXAdapter for Uniswap V2
type UniswapV2Adapter struct {
	client         *ethclient.Client
	routerAddress  common.Address
	factoryAddress common.Address
	routerABI      abi.ABI
	factoryABI     abi.ABI
	pairABI        abi.ABI
	fee            int // 30 basis points (0.3%)
}

// NewUniswapV2Adapter creates a new Uniswap V2 adapter
func NewUniswapV2Adapter(client *ethclient.Client, routerAddress common.Address) (*UniswapV2Adapter, error) {
	// Parse ABIs
	routerABI, err := abi.JSON(strings.NewReader(UniswapV2RouterABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse router ABI: %w", err)
	}

	factoryABI, err := abi.JSON(strings.NewReader(UniswapV2FactoryABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse factory ABI: %w", err)
	}

	pairABI, err := abi.JSON(strings.NewReader(UniswapV2PairABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse pair ABI: %w", err)
	}

	adapter := &UniswapV2Adapter{
		client:        client,
		routerAddress: routerAddress,
		routerABI:     routerABI,
		factoryABI:    factoryABI,
		pairABI:       pairABI,
		fee:           config.UniswapV2FeeBps,
	}

	// Get factory address from router
	factoryAddr, err := adapter.getFactoryFromRouter()
	if err != nil {
		return nil, fmt.Errorf("failed to get factory address: %w", err)
	}
	adapter.factoryAddress = factoryAddr

	log.Infof("Uniswap V2 adapter initialized (Router: %s, Factory: %s)",
		routerAddress.Hex(), factoryAddr.Hex())

	return adapter, nil
}

// GetName returns the name of the DEX
func (u *UniswapV2Adapter) GetName() string {
	return "Uniswap V2"
}

// GetType returns the type of DEX
func (u *UniswapV2Adapter) GetType() DEXType {
	return UniswapV2
}

// GetRouterAddress returns the router contract address
func (u *UniswapV2Adapter) GetRouterAddress() common.Address {
	return u.routerAddress
}

// GetFactoryAddress returns the factory contract address
func (u *UniswapV2Adapter) GetFactoryAddress() common.Address {
	return u.factoryAddress
}

// getFactoryFromRouter fetches factory address from router contract
func (u *UniswapV2Adapter) getFactoryFromRouter() (common.Address, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create bound contract
	contract := bind.NewBoundContract(u.routerAddress, u.routerABI, u.client, u.client, u.client)

	// Call factory() method
	var result []interface{}
	err := contract.Call(&bind.CallOpts{Context: ctx}, &result, "factory")
	if err != nil {
		return common.Address{}, fmt.Errorf("factory call failed: %w", err)
	}

	if len(result) == 0 {
		return common.Address{}, fmt.Errorf("factory call returned no results")
	}

	factoryAddr, ok := result[0].(common.Address)
	if !ok {
		return common.Address{}, fmt.Errorf("invalid factory address type")
	}

	return factoryAddr, nil
}

// GetPairAddress fetches the pair address for two tokens
func (u *UniswapV2Adapter) GetPairAddress(token0, token1 common.Address) (common.Address, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create bound contract
	contract := bind.NewBoundContract(u.factoryAddress, u.factoryABI, u.client, u.client, u.client)

	// Call getPair(token0, token1)
	var result []interface{}
	err := contract.Call(&bind.CallOpts{Context: ctx}, &result, "getPair", token0, token1)
	if err != nil {
		return common.Address{}, fmt.Errorf("getPair call failed: %w", err)
	}

	if len(result) == 0 {
		return common.Address{}, fmt.Errorf("getPair returned no results")
	}

	pairAddr, ok := result[0].(common.Address)
	if !ok {
		return common.Address{}, fmt.Errorf("invalid pair address type")
	}

	// Check if pair exists (zero address means no pair)
	if pairAddr == (common.Address{}) {
		return common.Address{}, fmt.Errorf("pair does not exist")
	}

	return pairAddr, nil
}

// GetReserves fetches current reserves of a pair
func (u *UniswapV2Adapter) GetReserves(pairAddress common.Address) (*big.Int, *big.Int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create bound contract
	contract := bind.NewBoundContract(pairAddress, u.pairABI, u.client, u.client, u.client)

	// Call getReserves()
	var result []interface{}
	err := contract.Call(&bind.CallOpts{Context: ctx}, &result, "getReserves")
	if err != nil {
		return nil, nil, fmt.Errorf("getReserves call failed: %w", err)
	}

	if len(result) < 2 {
		return nil, nil, fmt.Errorf("getReserves returned insufficient results")
	}

	reserve0, ok := result[0].(*big.Int)
	if !ok {
		return nil, nil, fmt.Errorf("invalid reserve0 type")
	}

	reserve1, ok := result[1].(*big.Int)
	if !ok {
		return nil, nil, fmt.Errorf("invalid reserve1 type")
	}

	return reserve0, reserve1, nil
}

// GetPool fetches pool information
func (u *UniswapV2Adapter) GetPool(token0, token1 common.Address) (*Pool, error) {
	// Get pair address
	pairAddr, err := u.GetPairAddress(token0, token1)
	if err != nil {
		return nil, fmt.Errorf("failed to get pair address: %w", err)
	}

	// Get reserves
	reserve0, reserve1, err := u.GetReserves(pairAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to get reserves: %w", err)
	}

	pool := &Pool{
		Address:     pairAddr,
		DEX:         UniswapV2,
		Token0:      token0,
		Token1:      token1,
		Reserve0:    reserve0,
		Reserve1:    reserve1,
		Fee:         u.fee,
		LastUpdated: time.Now().Unix(),
	}

	return pool, nil
}

// GetAmountOut calculates output amount for a given input
func (u *UniswapV2Adapter) GetAmountOut(amountIn *big.Int, reserveIn, reserveOut *big.Int) *big.Int {
	return utils.CalculateAmountOut(amountIn, reserveIn, reserveOut, u.fee)
}

// GetAmountIn calculates required input for desired output
func (u *UniswapV2Adapter) GetAmountIn(amountOut *big.Int, reserveIn, reserveOut *big.Int) *big.Int {
	return utils.CalculateAmountIn(amountOut, reserveIn, reserveOut, u.fee)
}

// Quote provides a price quote for a swap
func (u *UniswapV2Adapter) Quote(amountIn *big.Int, tokenIn, tokenOut common.Address) (*QuoteResult, error) {
	// Get pool
	pool, err := u.GetPool(tokenIn, tokenOut)
	if err != nil {
		return nil, fmt.Errorf("failed to get pool: %w", err)
	}

	// Determine direction
	var reserveIn, reserveOut *big.Int
	if pool.Token0 == tokenIn {
		reserveIn = pool.Reserve0
		reserveOut = pool.Reserve1
	} else {
		reserveIn = pool.Reserve1
		reserveOut = pool.Reserve0
	}

	// Calculate amount out
	amountOut := u.GetAmountOut(amountIn, reserveIn, reserveOut)

	// Calculate price impact
	priceImpact := utils.CalculatePriceImpact(amountIn, reserveIn, reserveOut, u.fee)

	// Calculate fee
	fee := new(big.Int).Mul(amountIn, big.NewInt(int64(u.fee)))
	fee.Div(fee, config.BigInt10000)

	// Calculate min amount out (with 0.5% slippage tolerance)
	slippage := big.NewInt(50) // 0.5%
	minAmountOut := new(big.Int).Mul(amountOut, new(big.Int).Sub(config.BigInt10000, slippage))
	minAmountOut.Div(minAmountOut, config.BigInt10000)

	result := &QuoteResult{
		AmountOut:    amountOut,
		PriceImpact:  priceImpact,
		Fee:          fee,
		MinAmountOut: minAmountOut,
		Route: &Route{
			Pools:       []*Pool{pool},
			Tokens:      []common.Address{tokenIn, tokenOut},
			Path:        []common.Address{tokenIn, tokenOut},
			AmountIn:    amountIn,
			AmountOut:   amountOut,
			PriceImpact: priceImpact,
		},
	}

	return result, nil
}
