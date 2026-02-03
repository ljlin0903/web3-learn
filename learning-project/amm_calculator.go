package main

import (
	"fmt"
	"math/big"
)

// AMM (Automated Market Maker) Price Calculator - Uniswap V2 Style
// AMM（自动做市商）价格计算器 - Uniswap V2 风格

// calculatePrice - Calculate the price of token1 in terms of token0
// calculatePrice - 计算 token1 相对于 token0 的价格
// Formula: price = reserve1 / reserve0
// 公式：价格 = 储备量1 / 储备量0
func calculatePrice(reserve0, reserve1 *big.Int, decimals0, decimals1 int) *big.Float {
	// Convert reserves to float with proper decimals
	// 将储备量转换为带正确小数位的浮点数
	r0 := new(big.Float).SetInt(reserve0)
	r1 := new(big.Float).SetInt(reserve1)
	
	// Adjust for decimals difference
	// 调整小数位差异
	decimalsDiff := decimals1 - decimals0
	if decimalsDiff != 0 {
		adjustment := new(big.Float).SetFloat64(float64(1))
		for i := 0; i < abs(decimalsDiff); i++ {
			adjustment.Mul(adjustment, big.NewFloat(10))
		}
		if decimalsDiff > 0 {
			r1.Quo(r1, adjustment)
		} else {
			r1.Mul(r1, adjustment)
		}
	}
	
	// Calculate price
	// 计算价格
	price := new(big.Float).Quo(r1, r0)
	return price
}

// getAmountOut - Calculate output amount for a given input (with 0.3% fee)
// getAmountOut - 根据输入计算输出金额（含 0.3% 手续费）
// Formula: amountOut = (amountIn * 997 * reserveOut) / (reserveIn * 1000 + amountIn * 997)
// 公式：输出金额 = (输入金额 * 997 * 输出储备) / (输入储备 * 1000 + 输入金额 * 997)
func getAmountOut(amountIn, reserveIn, reserveOut *big.Int) *big.Int {
	if amountIn.Cmp(big.NewInt(0)) <= 0 {
		return big.NewInt(0)
	}
	if reserveIn.Cmp(big.NewInt(0)) <= 0 || reserveOut.Cmp(big.NewInt(0)) <= 0 {
		return big.NewInt(0)
	}
	
	// amountInWithFee = amountIn * 997
	// 含手续费的输入金额 = 输入金额 * 997
	amountInWithFee := new(big.Int).Mul(amountIn, big.NewInt(997))
	
	// numerator = amountInWithFee * reserveOut
	// 分子 = 含手续费的输入金额 * 输出储备
	numerator := new(big.Int).Mul(amountInWithFee, reserveOut)
	
	// denominator = reserveIn * 1000 + amountInWithFee
	// 分母 = 输入储备 * 1000 + 含手续费的输入金额
	denominator := new(big.Int).Mul(reserveIn, big.NewInt(1000))
	denominator.Add(denominator, amountInWithFee)
	
	// amountOut = numerator / denominator
	// 输出金额 = 分子 / 分母
	amountOut := new(big.Int).Div(numerator, denominator)
	return amountOut
}

// getAmountIn - Calculate required input for a desired output (with 0.3% fee)
// getAmountIn - 根据期望输出计算所需输入（含 0.3% 手续费）
// Formula: amountIn = (reserveIn * amountOut * 1000) / ((reserveOut - amountOut) * 997) + 1
// 公式：输入金额 = (输入储备 * 输出金额 * 1000) / ((输出储备 - 输出金额) * 997) + 1
func getAmountIn(amountOut, reserveIn, reserveOut *big.Int) *big.Int {
	if amountOut.Cmp(big.NewInt(0)) <= 0 {
		return big.NewInt(0)
	}
	if reserveIn.Cmp(big.NewInt(0)) <= 0 || reserveOut.Cmp(big.NewInt(0)) <= 0 {
		return big.NewInt(0)
	}
	if amountOut.Cmp(reserveOut) >= 0 {
		return big.NewInt(0) // Cannot output more than reserve
	}
	
	// numerator = reserveIn * amountOut * 1000
	// 分子 = 输入储备 * 输出金额 * 1000
	numerator := new(big.Int).Mul(reserveIn, amountOut)
	numerator.Mul(numerator, big.NewInt(1000))
	
	// denominator = (reserveOut - amountOut) * 997
	// 分母 = (输出储备 - 输出金额) * 997
	denominator := new(big.Int).Sub(reserveOut, amountOut)
	denominator.Mul(denominator, big.NewInt(997))
	
	// amountIn = numerator / denominator + 1
	// 输入金额 = 分子 / 分母 + 1
	amountIn := new(big.Int).Div(numerator, denominator)
	amountIn.Add(amountIn, big.NewInt(1)) // Add 1 to account for rounding
	
	return amountIn
}

// calculatePriceImpact - Calculate price impact percentage
// calculatePriceImpact - 计算价格影响百分比
func calculatePriceImpact(amountIn, reserveIn, reserveOut *big.Int) *big.Float {
	// Initial price: reserveOut / reserveIn
	// 初始价格：输出储备 / 输入储备
	initialPrice := new(big.Float).Quo(
		new(big.Float).SetInt(reserveOut),
		new(big.Float).SetInt(reserveIn),
	)
	
	// Calculate amount out
	// 计算输出金额
	amountOut := getAmountOut(amountIn, reserveIn, reserveOut)
	
	// Execution price: amountOut / amountIn
	// 执行价格：输出金额 / 输入金额
	executionPrice := new(big.Float).Quo(
		new(big.Float).SetInt(amountOut),
		new(big.Float).SetInt(amountIn),
	)
	
	// Price impact = (initialPrice - executionPrice) / initialPrice * 100
	// 价格影响 = (初始价格 - 执行价格) / 初始价格 * 100
	diff := new(big.Float).Sub(initialPrice, executionPrice)
	impact := new(big.Float).Quo(diff, initialPrice)
	impact.Mul(impact, big.NewFloat(100))
	
	return impact
}

// abs - Helper function to get absolute value
// abs - 辅助函数获取绝对值
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Example usage / 示例用法
func main() {
	fmt.Println("========================================")
	fmt.Println("AMM 价格计算器 - Uniswap V2")
	fmt.Println("AMM Price Calculator - Uniswap V2")
	fmt.Println("========================================\n")
	
	// Example: WETH/USDC pool
	// 示例：WETH/USDC 池子
	// Reserve0 (WETH): 100 ETH = 100 * 10^18 wei
	// Reserve1 (USDC): 200,000 USDC = 200,000 * 10^6 (USDC has 6 decimals)
	reserve0 := new(big.Int)
	reserve0.SetString("100000000000000000000", 10) // 100 ETH
	
	reserve1 := new(big.Int)
	reserve1.SetString("200000000000", 10) // 200,000 USDC
	
	decimals0 := 18 // WETH decimals
	decimals1 := 6  // USDC decimals
	
	fmt.Println("📊 池子状态 (Pool State):")
	fmt.Printf("   Token0 (WETH) 储备: %s wei (100 ETH)\n", reserve0.String())
	fmt.Printf("   Token1 (USDC) 储备: %s (200,000 USDC)\n", reserve1.String())
	fmt.Println()
	
	// 1. Calculate current price
	// 1. 计算当前价格
	price := calculatePrice(reserve0, reserve1, decimals0, decimals1)
	fmt.Printf("💰 当前价格 (Current Price): 1 WETH = %s USDC\n\n", price.Text('f', 2))
	
	// 2. Calculate output for 1 ETH input
	// 2. 计算输入 1 ETH 的输出
	oneETH := new(big.Int)
	oneETH.SetString("1000000000000000000", 10) // 1 ETH
	
	amountOut := getAmountOut(oneETH, reserve0, reserve1)
	amountOutFloat := new(big.Float).Quo(
		new(big.Float).SetInt(amountOut),
		big.NewFloat(1000000), // USDC decimals
	)
	
	fmt.Println("🔄 交易模拟 (Trade Simulation):")
	fmt.Printf("   输入 (Input): 1 ETH\n")
	fmt.Printf("   输出 (Output): %s USDC\n", amountOutFloat.Text('f', 2))
	
	// 3. Calculate price impact
	// 3. 计算价格影响
	priceImpact := calculatePriceImpact(oneETH, reserve0, reserve1)
	fmt.Printf("   价格影响 (Price Impact): %s%%\n\n", priceImpact.Text('f', 4))
	
	// 4. Calculate required input for 1000 USDC output
	// 4. 计算获得 1000 USDC 需要的输入
	targetOutput := new(big.Int)
	targetOutput.SetString("1000000000", 10) // 1000 USDC
	
	requiredInput := getAmountIn(targetOutput, reserve0, reserve1)
	requiredInputFloat := new(big.Float).Quo(
		new(big.Float).SetInt(requiredInput),
		big.NewFloat(1e18), // ETH decimals
	)
	
	fmt.Println("🎯 反向计算 (Reverse Calculation):")
	fmt.Printf("   期望输出 (Target Output): 1000 USDC\n")
	fmt.Printf("   所需输入 (Required Input): %s ETH\n", requiredInputFloat.Text('f', 6))
	
	fmt.Println("\n✅ 计算完成！")
}
