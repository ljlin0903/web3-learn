package main

import (
	"fmt"
	"math/big"
)

// ArbitrageFinder - Triangle Arbitrage Path Finder
// 套利路径搜索器 - 三角套利路径查找

// Pool represents a liquidity pool
// Pool 表示一个流动性池子
type Pool struct {
	Name      string   // Pool name / 池子名称
	Token0    string   // Token0 symbol / Token0 符号
	Token1    string   // Token1 symbol / Token1 符号
	Reserve0  *big.Int // Token0 reserve / Token0 储备量
	Reserve1  *big.Int // Token1 reserve / Token1 储备量
	Fee       int      // Fee in basis points (30 = 0.3%) / 手续费（基点，30 = 0.3%）
}

// ArbitragePath represents a potential arbitrage opportunity
// ArbitragePath 表示一个潜在的套利机会
type ArbitragePath struct {
	Pools      []*Pool   // Sequence of pools / 池子序列
	Tokens     []string  // Token path / 代币路径
	StartAmount *big.Int // Initial amount / 初始金额
	EndAmount   *big.Int // Final amount after arbitrage / 套利后的最终金额
	Profit      *big.Int // Profit amount / 利润金额
	ProfitPct   float64  // Profit percentage / 利润百分比
}

// getAmountOut - Calculate output amount with fee
// getAmountOut - 计算含手续费的输出金额
func getAmountOut(amountIn, reserveIn, reserveOut *big.Int, feeBasisPoints int) *big.Int {
	if amountIn.Cmp(big.NewInt(0)) <= 0 {
		return big.NewInt(0)
	}
	
	// Calculate fee multiplier (e.g., 30 basis points = 997/1000)
	// 计算手续费乘数（例如：30个基点 = 997/1000）
	feeMultiplier := 10000 - feeBasisPoints
	
	// amountInWithFee = amountIn * feeMultiplier
	amountInWithFee := new(big.Int).Mul(amountIn, big.NewInt(int64(feeMultiplier)))
	
	// numerator = amountInWithFee * reserveOut
	numerator := new(big.Int).Mul(amountInWithFee, reserveOut)
	
	// denominator = reserveIn * 10000 + amountInWithFee
	denominator := new(big.Int).Mul(reserveIn, big.NewInt(10000))
	denominator.Add(denominator, amountInWithFee)
	
	// amountOut = numerator / denominator
	amountOut := new(big.Int).Div(numerator, denominator)
	return amountOut
}

// findTriangleArbitrage - Find triangle arbitrage opportunities
// findTriangleArbitrage - 查找三角套利机会
// Example: ETH -> USDC -> DAI -> ETH
// 示例：ETH -> USDC -> DAI -> ETH
func findTriangleArbitrage(pools []*Pool, startToken string, startAmount *big.Int) []*ArbitragePath {
	var opportunities []*ArbitragePath
	
	// Try all possible 3-hop paths
	// 尝试所有可能的3跳路径
	for i, pool1 := range pools {
		// First hop: startToken -> intermediateToken1
		// 第一跳：起始代币 -> 中间代币1
		var intermediateToken1 string
		var amount1 *big.Int
		
		if pool1.Token0 == startToken {
			intermediateToken1 = pool1.Token1
			amount1 = getAmountOut(startAmount, pool1.Reserve0, pool1.Reserve1, pool1.Fee)
		} else if pool1.Token1 == startToken {
			intermediateToken1 = pool1.Token0
			amount1 = getAmountOut(startAmount, pool1.Reserve1, pool1.Reserve0, pool1.Fee)
		} else {
			continue // Pool doesn't contain start token
		}
		
		if amount1.Cmp(big.NewInt(0)) <= 0 {
			continue
		}
		
		// Second hop: intermediateToken1 -> intermediateToken2
		// 第二跳：中间代币1 -> 中间代币2
		for j, pool2 := range pools {
			if i == j {
				continue // Skip same pool
			}
			
			var intermediateToken2 string
			var amount2 *big.Int
			
			if pool2.Token0 == intermediateToken1 {
				intermediateToken2 = pool2.Token1
				amount2 = getAmountOut(amount1, pool2.Reserve0, pool2.Reserve1, pool2.Fee)
			} else if pool2.Token1 == intermediateToken1 {
				intermediateToken2 = pool2.Token0
				amount2 = getAmountOut(amount1, pool2.Reserve1, pool2.Reserve0, pool2.Fee)
			} else {
				continue
			}
			
			if amount2.Cmp(big.NewInt(0)) <= 0 {
				continue
			}
			
			// Third hop: intermediateToken2 -> startToken (complete the loop)
			// 第三跳：中间代币2 -> 起始代币（闭环）
			for k, pool3 := range pools {
				if k == i || k == j {
					continue // Skip used pools
				}
				
				var finalToken string
				var finalAmount *big.Int
				
				if pool3.Token0 == intermediateToken2 && pool3.Token1 == startToken {
					finalToken = pool3.Token1
					finalAmount = getAmountOut(amount2, pool3.Reserve0, pool3.Reserve1, pool3.Fee)
				} else if pool3.Token1 == intermediateToken2 && pool3.Token0 == startToken {
					finalToken = pool3.Token0
					finalAmount = getAmountOut(amount2, pool3.Reserve1, pool3.Reserve0, pool3.Fee)
				} else {
					continue
				}
				
				if finalToken != startToken {
					continue // Path doesn't loop back
				}
				
				// Calculate profit
				// 计算利润
				profit := new(big.Int).Sub(finalAmount, startAmount)
				
				// Only record if profitable
				// 只记录盈利的机会
				if profit.Cmp(big.NewInt(0)) > 0 {
					profitFloat := new(big.Float).SetInt(profit)
					startFloat := new(big.Float).SetInt(startAmount)
					profitPct, _ := new(big.Float).Quo(profitFloat, startFloat).Float64()
					profitPct *= 100
					
					path := &ArbitragePath{
						Pools:      []*Pool{pool1, pool2, pool3},
						Tokens:     []string{startToken, intermediateToken1, intermediateToken2, startToken},
						StartAmount: new(big.Int).Set(startAmount),
						EndAmount:   finalAmount,
						Profit:      profit,
						ProfitPct:   profitPct,
					}
					
					opportunities = append(opportunities, path)
				}
			}
		}
	}
	
	return opportunities
}

// displayArbitragePath - Display arbitrage path details
// displayArbitragePath - 显示套利路径详情
func displayArbitragePath(path *ArbitragePath, index int) {
	fmt.Printf("\n💰 套利机会 #%d (Arbitrage Opportunity #%d)\n", index, index)
	fmt.Println("   ========================================")
	fmt.Printf("   📊 利润率 (Profit): %s (%0.4f%%)\n", path.Profit.String(), path.ProfitPct)
	fmt.Printf("   💵 初始金额 (Start): %s %s\n", path.StartAmount.String(), path.Tokens[0])
	fmt.Printf("   💰 最终金额 (End): %s %s\n", path.EndAmount.String(), path.Tokens[len(path.Tokens)-1])
	fmt.Println("\n   🔄 交易路径 (Trade Path):")
	
	for i, pool := range path.Pools {
		fmt.Printf("      Step %d: %s (%s/%s)\n", i+1, pool.Name, pool.Token0, pool.Token1)
		fmt.Printf("              %s -> %s\n", path.Tokens[i], path.Tokens[i+1])
	}
	fmt.Println("   ========================================")
}

func main() {
	fmt.Println("========================================")
	fmt.Println("🔍 三角套利路径搜索器")
	fmt.Println("🔍 Triangle Arbitrage Finder")
	fmt.Println("========================================\n")
	
	// Example pools - Simulated liquidity pools
	// 示例池子 - 模拟的流动性池子
	// 【测试专用，正式项目需删除】Note: These are simulated values for demonstration
	// 【测试专用，正式项目需删除】注意：这些是用于演示的模拟值
	
	// Helper to create big.Int from string
	// 辅助函数：从字符串创建 big.Int
	bigInt := func(s string) *big.Int {
		n := new(big.Int)
		n.SetString(s, 10)
		return n
	}
	
	pools := []*Pool{
		{
			Name:     "Uniswap ETH/USDC",
			Token0:   "ETH",
			Token1:   "USDC",
			Reserve0: bigInt("100000000000000000000"),   // 100 ETH
			Reserve1: bigInt("200000000000"),           // 200,000 USDC (1 ETH = 2000 USDC)
			Fee:      30,                               // 0.3%
		},
		{
			Name:     "Sushiswap USDC/DAI",
			Token0:   "USDC",
			Token1:   "DAI",
			Reserve0: bigInt("500000000000"),           // 500,000 USDC
			Reserve1: bigInt("520000000000000000000000"), // 520,000 DAI (4% premium, 1 USDC = 1.04 DAI)
			Fee:      30,
		},
		{
			Name:     "Curve DAI/ETH",
			Token0:   "DAI",
			Token1:   "ETH",
			Reserve0: bigInt("190000000000000000000000"), // 190,000 DAI
			Reserve1: bigInt("100000000000000000000"),   // 100 ETH (1 ETH = 1900 DAI, creates arbitrage opportunity)
			Fee:      30,
		},
	}
	
	fmt.Println("📊 流动性池子信息 (Liquidity Pools):")
	for i, pool := range pools {
		fmt.Printf("   %d. %s\n", i+1, pool.Name)
		fmt.Printf("      Reserve: %s %s / %s %s\n", 
			pool.Reserve0.String(), pool.Token0,
			pool.Reserve1.String(), pool.Token1)
	}
	
	// Search for arbitrage with 1 ETH
	// 用 1 ETH 搜索套利机会
	startToken := "ETH"
	startAmount := bigInt("1000000000000000000") // 1 ETH
	
	fmt.Printf("\n🔍 搜索套利路径...\n")
	fmt.Printf("   起始代币 (Start Token): %s\n", startToken)
	fmt.Printf("   起始金额 (Start Amount): %s wei (1 ETH)\n\n", startAmount.String())
	
	opportunities := findTriangleArbitrage(pools, startToken, startAmount)
	
	if len(opportunities) == 0 {
		fmt.Println("❌ 未发现盈利套利机会")
		fmt.Println("❌ No profitable arbitrage opportunities found")
	} else {
		fmt.Printf("✅ 发现 %d 个套利机会！\n", len(opportunities))
		fmt.Printf("✅ Found %d arbitrage opportunities!\n", len(opportunities))
		
		for i, opp := range opportunities {
			displayArbitragePath(opp, i+1)
		}
		
		// Find best opportunity
		// 找到最佳机会
		var bestOpp *ArbitragePath
		for _, opp := range opportunities {
			if bestOpp == nil || opp.ProfitPct > bestOpp.ProfitPct {
				bestOpp = opp
			}
		}
		
		fmt.Println("\n🏆 最佳套利机会 (Best Opportunity):")
		fmt.Printf("   利润率: %0.4f%%\n", bestOpp.ProfitPct)
		fmt.Printf("   利润: %s wei\n", bestOpp.Profit.String())
	}
	
	fmt.Println("\n✅ 搜索完成！")
}
