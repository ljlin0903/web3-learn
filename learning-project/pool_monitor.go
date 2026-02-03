package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
)

// Uniswap V2 Pair Contract - Swap Event
// Uniswap V2 交易对合约 - Swap 事件
// event Swap(address indexed sender, uint amount0In, uint amount1In, uint amount0Out, uint amount1Out, address indexed to);

// Uniswap V2 Pair ABI (only Swap event)
// Uniswap V2 交易对 ABI（仅 Swap 事件）
const pairABI = `[
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "sender", "type": "address"},
			{"indexed": false, "name": "amount0In", "type": "uint256"},
			{"indexed": false, "name": "amount1In", "type": "uint256"},
			{"indexed": false, "name": "amount0Out", "type": "uint256"},
			{"indexed": false, "name": "amount1Out", "type": "uint256"},
			{"indexed": true, "name": "to", "type": "address"}
		],
		"name": "Swap",
		"type": "event"
	}
]`

// SwapEvent represents the Swap event data
// SwapEvent 表示 Swap 事件数据
type SwapEvent struct {
	Sender    common.Address
	Amount0In *big.Int
	Amount1In *big.Int
	Amount0Out *big.Int
	Amount1Out *big.Int
	To        common.Address
}

func main() {
	// 1. Load configuration
	// 1. 加载配置
	fmt.Println("[初始化] 正在加载 .env 配置...")
	err := godotenv.Load()
	if err != nil {
		log.Fatal("❌ 加载 .env 文件失败")
	}

	wssURL := os.Getenv("RPC_WSS_URL")
	if wssURL == "" {
		log.Fatal("❌ RPC_WSS_URL 未配置")
	}
	fmt.Printf("✓ 配置加载成功\n[配置] WSS 节点: %s\n", wssURL)

	// 2. Connect to Ethereum node
	// 2. 连接以太坊节点
	fmt.Println("\n[连接] 正在建立 WebSocket 连接...")
	client, err := ethclient.Dial(wssURL)
	if err != nil {
		log.Fatalf("❌ 连接失败: %v", err)
	}
	defer client.Close()
	fmt.Println("✓ WebSocket 连接成功！")

	// 3. Example: WETH/USDC Pair on Sepolia (replace with actual address)
	// 3. 示例：Sepolia 测试网上的 WETH/USDC 交易对（需替换为实际地址）
	// 【测试专用，正式项目需删除】Note: Sepolia may not have active Uniswap pools, use mainnet fork or local test
	// 【测试专用，正式项目需删除】注意：Sepolia 可能没有活跃的 Uniswap 池子，建议使用主网分叉或本地测试
	pairAddress := common.HexToAddress("0x0000000000000000000000000000000000000000") // Placeholder
	
	fmt.Printf("\n[监听] 交易对地址: %s\n", pairAddress.Hex())
	fmt.Println("⚠️  当前为占位地址，Sepolia 测试网可能没有活跃的 Uniswap V2 池子")
	fmt.Println("💡 建议：使用主网地址配合 Alchemy 主网 API 或本地 Anvil 分叉测试\n")

	// 4. Parse ABI
	// 4. 解析 ABI
	contractABI, err := abi.JSON(strings.NewReader(pairABI))
	if err != nil {
		log.Fatalf("❌ ABI 解析失败: %v", err)
	}

	// 5. Create filter query for Swap events
	// 5. 创建 Swap 事件过滤查询
	query := ethereum.FilterQuery{
		Addresses: []common.Address{pairAddress},
		Topics:    [][]common.Hash{{contractABI.Events["Swap"].ID}}, // Swap event signature
	}

	fmt.Println("[订阅] 正在订阅 Swap 事件...")
	logs := make(chan types.Log)
	sub, err := client.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		log.Fatalf("❌ 订阅失败: %v", err)
	}
	defer sub.Unsubscribe()

	fmt.Println("✓ 订阅成功！")
	fmt.Println("\n========================================")
	fmt.Println("🎯 监听 Swap 事件中... (按 Ctrl+C 退出)")
	fmt.Println("========================================\n")

	// 6. Listen for events
	// 6. 监听事件
	swapCount := 0
	for {
		select {
		case err := <-sub.Err():
			log.Fatalf("❌ 订阅错误: %v", err)

		case vLog := <-logs:
			swapCount++
			
			// Parse event data
			// 解析事件数据
			var swapEvent SwapEvent
			err := contractABI.UnpackIntoInterface(&swapEvent, "Swap", vLog.Data)
			if err != nil {
				log.Printf("❌ 解析事件失败: %v", err)
				continue
			}

			// Extract indexed parameters (sender, to)
			// 提取索引参数（sender, to）
			swapEvent.Sender = common.HexToAddress(vLog.Topics[1].Hex())
			swapEvent.To = common.HexToAddress(vLog.Topics[2].Hex())

			// Display swap information
			// 显示交易信息
			fmt.Printf("💱 Swap #%d 事件\n", swapCount)
			fmt.Printf("   📍 交易对: %s\n", pairAddress.Hex())
			fmt.Printf("   👤 发送者: %s\n", swapEvent.Sender.Hex())
			fmt.Printf("   📥 Token0 输入: %s\n", swapEvent.Amount0In.String())
			fmt.Printf("   📥 Token1 输入: %s\n", swapEvent.Amount1In.String())
			fmt.Printf("   📤 Token0 输出: %s\n", swapEvent.Amount0Out.String())
			fmt.Printf("   📤 Token1 输出: %s\n", swapEvent.Amount1Out.String())
			fmt.Printf("   🎯 接收者: %s\n", swapEvent.To.Hex())
			fmt.Printf("   🔗 区块号: %d\n", vLog.BlockNumber)
			fmt.Printf("   📝 交易哈希: %s\n", vLog.TxHash.Hex())
			
			// Calculate effective price (simple version)
			// 计算有效价格（简化版）
			if swapEvent.Amount0In.Cmp(big.NewInt(0)) > 0 && swapEvent.Amount1Out.Cmp(big.NewInt(0)) > 0 {
				// Token0 -> Token1
				price := new(big.Float).Quo(
					new(big.Float).SetInt(swapEvent.Amount1Out),
					new(big.Float).SetInt(swapEvent.Amount0In),
				)
				fmt.Printf("   💰 价格 (Token1/Token0): %s\n", price.String())
			} else if swapEvent.Amount1In.Cmp(big.NewInt(0)) > 0 && swapEvent.Amount0Out.Cmp(big.NewInt(0)) > 0 {
				// Token1 -> Token0
				price := new(big.Float).Quo(
					new(big.Float).SetInt(swapEvent.Amount0Out),
					new(big.Float).SetInt(swapEvent.Amount1In),
				)
				fmt.Printf("   💰 价格 (Token0/Token1): %s\n", price.String())
			}
			fmt.Println("   ----------------------------------------")
		}
	}
}
