package strategy

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"github.com/ljlin/mev-arbitrage-bot/pkg/config"
	"github.com/ljlin/mev-arbitrage-bot/pkg/dex"
	"github.com/ljlin/mev-arbitrage-bot/pkg/utils"
)

// ArbitrageFinder finds arbitrage opportunities across multiple DEXs
type ArbitrageFinder struct {
	poolMonitor    *dex.PoolMonitor
	minProfitBps   int
	maxTradeAmount *big.Int
	minTradeAmount *big.Int
}

// NewArbitrageFinder creates a new arbitrage finder
func NewArbitrageFinder(poolMonitor *dex.PoolMonitor, cfg *config.Config) *ArbitrageFinder {
	return &ArbitrageFinder{
		poolMonitor:    poolMonitor,
		minProfitBps:   cfg.MinProfitBps,
		maxTradeAmount: utils.EtherToWei(cfg.MaxTradeAmountETH),
		minTradeAmount: utils.EtherToWei(cfg.MinTradeAmountETH),
	}
}

// FindTriangleArbitrage finds triangle arbitrage opportunities
// Example: WETH -> USDC -> DAI -> WETH
func (af *ArbitrageFinder) FindTriangleArbitrage(startToken common.Address) ([]*ArbitragePath, error) {
	pools := af.poolMonitor.GetAllPools()
	if len(pools) < 3 {
		return nil, fmt.Errorf("insufficient pools: need at least 3, got %d", len(pools))
	}

	opportunities := make([]*ArbitragePath, 0)

	// Try different start amounts to find optimal trade size
	startAmounts := af.generateStartAmounts()

	for _, startAmount := range startAmounts {
		paths := af.searchTrianglePaths(pools, startToken, startAmount)
		opportunities = append(opportunities, paths...)
	}

	// Filter by minimum profit
	filtered := make([]*ArbitragePath, 0)
	for _, opp := range opportunities {
		if opp.ProfitBps >= af.minProfitBps {
			filtered = append(filtered, opp)
		}
	}

	log.Debugf("Found %d arbitrage opportunities (filtered from %d)",
		len(filtered), len(opportunities))

	return filtered, nil
}

// searchTrianglePaths searches for 3-hop arbitrage paths
func (af *ArbitrageFinder) searchTrianglePaths(pools []*dex.Pool, startToken common.Address, startAmount *big.Int) []*ArbitragePath {
	paths := make([]*ArbitragePath, 0)

	// First hop: startToken -> intermediateToken1
	for i, pool1 := range pools {
		var token1 common.Address
		var amount1 *big.Int

		if pool1.Token0 == startToken {
			token1 = pool1.Token1
			amount1 = utils.CalculateAmountOut(startAmount, pool1.Reserve0, pool1.Reserve1, pool1.Fee)
		} else if pool1.Token1 == startToken {
			token1 = pool1.Token0
			amount1 = utils.CalculateAmountOut(startAmount, pool1.Reserve1, pool1.Reserve0, pool1.Fee)
		} else {
			continue
		}

		if amount1.Cmp(config.BigInt0) <= 0 {
			continue
		}

		// Second hop: intermediateToken1 -> intermediateToken2
		for j, pool2 := range pools {
			if i == j {
				continue
			}

			var token2 common.Address
			var amount2 *big.Int

			if pool2.Token0 == token1 {
				token2 = pool2.Token1
				amount2 = utils.CalculateAmountOut(amount1, pool2.Reserve0, pool2.Reserve1, pool2.Fee)
			} else if pool2.Token1 == token1 {
				token2 = pool2.Token0
				amount2 = utils.CalculateAmountOut(amount1, pool2.Reserve1, pool2.Reserve0, pool2.Fee)
			} else {
				continue
			}

			if amount2.Cmp(config.BigInt0) <= 0 {
				continue
			}

			// Third hop: intermediateToken2 -> startToken (complete the loop)
			for k, pool3 := range pools {
				if k == i || k == j {
					continue
				}

				var finalToken common.Address
				var finalAmount *big.Int

				if pool3.Token0 == token2 && pool3.Token1 == startToken {
					finalToken = pool3.Token1
					finalAmount = utils.CalculateAmountOut(amount2, pool3.Reserve0, pool3.Reserve1, pool3.Fee)
				} else if pool3.Token1 == token2 && pool3.Token0 == startToken {
					finalToken = pool3.Token0
					finalAmount = utils.CalculateAmountOut(amount2, pool3.Reserve1, pool3.Reserve0, pool3.Fee)
				} else {
					continue
				}

				if finalToken != startToken {
					continue
				}

				// Calculate profit
				profit := new(big.Int).Sub(finalAmount, startAmount)
				if profit.Cmp(config.BigInt0) <= 0 {
					continue
				}

				profitBps := utils.CalculateProfit(startAmount, finalAmount)

				// Create arbitrage path
				path := &ArbitragePath{
					ID:          uuid.New().String(),
					Pools:       []*dex.Pool{pool1, pool2, pool3},
					Tokens:      []common.Address{startToken, token1, token2, startToken},
					StartToken:  startToken,
					StartAmount: new(big.Int).Set(startAmount),
					EndAmount:   finalAmount,
					Profit:      profit,
					ProfitBps:   profitBps,
					ProfitETH:   utils.WeiToEther(profit),
					Timestamp:   time.Now().Unix(),
				}

				paths = append(paths, path)
			}
		}
	}

	return paths
}

// generateStartAmounts generates different start amounts to test
func (af *ArbitrageFinder) generateStartAmounts() []*big.Int {
	amounts := make([]*big.Int, 0)

	// Start with min amount
	current := new(big.Int).Set(af.minTradeAmount)
	step := new(big.Int).Div(af.maxTradeAmount, big.NewInt(10)) // 10% steps

	for current.Cmp(af.maxTradeAmount) <= 0 {
		amounts = append(amounts, new(big.Int).Set(current))
		current.Add(current, step)
	}

	return amounts
}

// EstimateGasCost estimates gas cost for an arbitrage path
func (af *ArbitrageFinder) EstimateGasCost(path *ArbitragePath, gasPrice *big.Int) *big.Int {
	// Estimate gas units based on number of swaps
	// Each swap costs approximately 100,000 gas
	// Plus base transaction cost of 21,000 gas
	numSwaps := int64(len(path.Pools))
	gasUnits := big.NewInt(21000 + numSwaps*100000)

	// Total gas cost = gas units * gas price
	gasCost := new(big.Int).Mul(gasUnits, gasPrice)

	return gasCost
}

// CalculateNetProfit calculates net profit after gas costs
func (af *ArbitrageFinder) CalculateNetProfit(path *ArbitragePath, gasPrice *big.Int) {
	gasCost := af.EstimateGasCost(path, gasPrice)
	path.GasCostEst = gasCost

	// Net profit = profit - gas cost
	netProfit := new(big.Int).Sub(path.Profit, gasCost)
	path.NetProfit = netProfit

	// Calculate net profit in bps
	if netProfit.Cmp(config.BigInt0) > 0 {
		path.NetProfitBps = utils.CalculateProfit(path.StartAmount,
			new(big.Int).Add(path.StartAmount, netProfit))
	} else {
		path.NetProfitBps = 0
	}
}

// ValidateOpportunity validates if an opportunity is executable
func (af *ArbitrageFinder) ValidateOpportunity(path *ArbitragePath, gasPrice *big.Int) *ArbitrageOpportunity {
	opportunity := &ArbitrageOpportunity{
		Path:         path,
		IsExecutable: true,
		Priority:     path.ProfitBps,
	}

	// Calculate net profit
	af.CalculateNetProfit(path, gasPrice)

	// Check if still profitable after gas
	if path.NetProfit.Cmp(config.BigInt0) <= 0 {
		opportunity.IsExecutable = false
		opportunity.Reason = "unprofitable after gas costs"
		return opportunity
	}

	// Check if net profit meets minimum threshold
	if path.NetProfitBps < af.minProfitBps {
		opportunity.IsExecutable = false
		opportunity.Reason = fmt.Sprintf("net profit %d bps < minimum %d bps",
			path.NetProfitBps, af.minProfitBps)
		return opportunity
	}

	// Check trade amount bounds
	if path.StartAmount.Cmp(af.minTradeAmount) < 0 {
		opportunity.IsExecutable = false
		opportunity.Reason = "trade amount below minimum"
		return opportunity
	}

	if path.StartAmount.Cmp(af.maxTradeAmount) > 0 {
		opportunity.IsExecutable = false
		opportunity.Reason = "trade amount above maximum"
		return opportunity
	}

	// All checks passed
	log.Infof("✅ Valid arbitrage opportunity: %s (Net Profit: %.4f%%, %s ETH)",
		path.ID[:8], utils.BpsToPercentage(path.NetProfitBps),
		path.ProfitETH.Text('f', 6))

	return opportunity
}

// FindBestOpportunity finds the best arbitrage opportunity
func (af *ArbitrageFinder) FindBestOpportunity(startToken common.Address, gasPrice *big.Int) (*ArbitrageOpportunity, error) {
	// Find all opportunities
	paths, err := af.FindTriangleArbitrage(startToken)
	if err != nil {
		return nil, err
	}

	if len(paths) == 0 {
		return nil, fmt.Errorf("no arbitrage opportunities found")
	}

	// Validate and find best
	var bestOpp *ArbitrageOpportunity

	for _, path := range paths {
		opp := af.ValidateOpportunity(path, gasPrice)

		if !opp.IsExecutable {
			continue
		}

		if bestOpp == nil || opp.Path.NetProfitBps > bestOpp.Path.NetProfitBps {
			bestOpp = opp
		}
	}

	if bestOpp == nil {
		return nil, fmt.Errorf("no executable opportunities found")
	}

	return bestOpp, nil
}
