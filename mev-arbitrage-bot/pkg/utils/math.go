package utils

import (
	"math/big"

	"github.com/ljlin/mev-arbitrage-bot/pkg/config"
)

// CalculateAmountOut calculates the output amount for a given input using constant product formula
// Formula: amountOut = (amountIn * feeMultiplier * reserveOut) / (reserveIn * 10000 + amountIn * feeMultiplier)
func CalculateAmountOut(amountIn, reserveIn, reserveOut *big.Int, feeBps int) *big.Int {
	if amountIn.Cmp(config.BigInt0) <= 0 {
		return big.NewInt(0)
	}
	if reserveIn.Cmp(config.BigInt0) <= 0 || reserveOut.Cmp(config.BigInt0) <= 0 {
		return big.NewInt(0)
	}

	// Calculate fee multiplier (e.g., 30 basis points = 9970/10000)
	feeMultiplier := new(big.Int).Sub(config.BigInt10000, big.NewInt(int64(feeBps)))

	// amountInWithFee = amountIn * feeMultiplier
	amountInWithFee := new(big.Int).Mul(amountIn, feeMultiplier)

	// numerator = amountInWithFee * reserveOut
	numerator := new(big.Int).Mul(amountInWithFee, reserveOut)

	// denominator = reserveIn * 10000 + amountInWithFee
	denominator := new(big.Int).Mul(reserveIn, config.BigInt10000)
	denominator.Add(denominator, amountInWithFee)

	// amountOut = numerator / denominator
	amountOut := new(big.Int).Div(numerator, denominator)
	return amountOut
}

// CalculateAmountIn calculates the required input for a desired output
// Formula: amountIn = (reserveIn * amountOut * 10000) / ((reserveOut - amountOut) * feeMultiplier) + 1
func CalculateAmountIn(amountOut, reserveIn, reserveOut *big.Int, feeBps int) *big.Int {
	if amountOut.Cmp(config.BigInt0) <= 0 {
		return big.NewInt(0)
	}
	if reserveIn.Cmp(config.BigInt0) <= 0 || reserveOut.Cmp(config.BigInt0) <= 0 {
		return big.NewInt(0)
	}
	if amountOut.Cmp(reserveOut) >= 0 {
		return big.NewInt(0) // Cannot output more than reserve
	}

	feeMultiplier := new(big.Int).Sub(config.BigInt10000, big.NewInt(int64(feeBps)))

	// numerator = reserveIn * amountOut * 10000
	numerator := new(big.Int).Mul(reserveIn, amountOut)
	numerator.Mul(numerator, config.BigInt10000)

	// denominator = (reserveOut - amountOut) * feeMultiplier
	denominator := new(big.Int).Sub(reserveOut, amountOut)
	denominator.Mul(denominator, feeMultiplier)

	// amountIn = numerator / denominator + 1
	amountIn := new(big.Int).Div(numerator, denominator)
	amountIn.Add(amountIn, config.BigInt1)

	return amountIn
}

// CalculatePriceImpact calculates the price impact percentage for a trade
func CalculatePriceImpact(amountIn, reserveIn, reserveOut *big.Int, feeBps int) *big.Float {
	// Initial price: reserveOut / reserveIn
	initialPrice := new(big.Float).Quo(
		new(big.Float).SetInt(reserveOut),
		new(big.Float).SetInt(reserveIn),
	)

	// Calculate amount out
	amountOut := CalculateAmountOut(amountIn, reserveIn, reserveOut, feeBps)

	// Execution price: amountOut / amountIn
	executionPrice := new(big.Float).Quo(
		new(big.Float).SetInt(amountOut),
		new(big.Float).SetInt(amountIn),
	)

	// Price impact = (initialPrice - executionPrice) / initialPrice * 100
	diff := new(big.Float).Sub(initialPrice, executionPrice)
	impact := new(big.Float).Quo(diff, initialPrice)
	impact.Mul(impact, big.NewFloat(100))

	return impact
}

// CalculateProfit calculates profit in basis points
// Returns profit in BPS (1 BPS = 0.01%)
func CalculateProfit(startAmount, endAmount *big.Int) int {
	if startAmount.Cmp(config.BigInt0) <= 0 {
		return 0
	}

	diff := new(big.Int).Sub(endAmount, startAmount)
	if diff.Cmp(config.BigInt0) <= 0 {
		return 0
	}

	// profit = (diff * 10000) / startAmount
	profit := new(big.Int).Mul(diff, config.BigInt10000)
	profit.Div(profit, startAmount)

	return int(profit.Int64())
}

// MinBigInt returns the minimum of two big.Int values
func MinBigInt(a, b *big.Int) *big.Int {
	if a.Cmp(b) < 0 {
		return new(big.Int).Set(a)
	}
	return new(big.Int).Set(b)
}

// MaxBigInt returns the maximum of two big.Int values
func MaxBigInt(a, b *big.Int) *big.Int {
	if a.Cmp(b) > 0 {
		return new(big.Int).Set(a)
	}
	return new(big.Int).Set(b)
}

// PercentageToBps converts a percentage to basis points
func PercentageToBps(percentage float64) int {
	return int(percentage * 100)
}

// BpsToPercentage converts basis points to percentage
func BpsToPercentage(bps int) float64 {
	return float64(bps) / 100.0
}

// WeiToEther converts Wei to Ether
func WeiToEther(wei *big.Int) *big.Float {
	fbalance := new(big.Float).SetInt(wei)
	ethValue := new(big.Float).Quo(fbalance, big.NewFloat(config.WeiPerEther))
	return ethValue
}

// EtherToWei converts Ether to Wei
func EtherToWei(ether *big.Float) *big.Int {
	wei := new(big.Float).Mul(ether, big.NewFloat(config.WeiPerEther))
	result, _ := wei.Int(nil)
	return result
}

// GweiToWei converts Gwei to Wei
func GweiToWei(gwei uint64) *big.Int {
	return new(big.Int).Mul(big.NewInt(int64(gwei)), big.NewInt(config.GweiPerEther))
}

// WeiToGwei converts Wei to Gwei
func WeiToGwei(wei *big.Int) uint64 {
	gwei := new(big.Int).Div(wei, big.NewInt(config.GweiPerEther))
	return gwei.Uint64()
}
