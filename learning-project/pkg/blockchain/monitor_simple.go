package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
)

// Block Monitor - Subscribe to new block headers via WebSocket
// 区块监听器 - 通过 WebSocket 订阅新区块头
func main() {
	// 1. Load environment configuration
	// 1. 加载环境配置
	fmt.Println("[初始化] 正在加载 .env 配置...")
	err := godotenv.Load()
	if err != nil {
		log.Fatal("❌ 加载 .env 文件失败 (Failed to load .env file)")
	}
	fmt.Println("✓ 配置加载成功")

	// 2. Get WebSocket RPC URL
	// 2. 获取 WebSocket RPC 链接
	wssURL := os.Getenv("RPC_WSS_URL")
	if wssURL == "" {
		log.Fatal("❌ RPC_WSS_URL 未配置 (RPC_WSS_URL not configured)")
	}
	fmt.Printf("[配置] WSS 节点: %s\n", wssURL)

	// 3. Connect to Ethereum node via WebSocket with timeout
	// 3. 通过 WebSocket 连接以太坊节点（带超时）
	fmt.Println("\n[连接] 正在建立 WebSocket 连接...")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	client, err := ethclient.DialContext(ctx, wssURL)
	if err != nil {
		log.Fatalf("❌ WebSocket 连接失败 (Connection failed): %v", err)
	}
	defer client.Close()
	fmt.Println("✓ WebSocket 连接成功！")

	// 4. Subscribe to new block headers
	// 4. 订阅新区块头
	fmt.Println("\n[订阅] 正在订阅新区块事件 (newHeads)...")
	headers := make(chan *types.Header)
	sub, err := client.SubscribeNewHead(context.Background(), headers)
	if err != nil {
		log.Fatalf("❌ 订阅失败 (Subscription failed): %v", err)
	}
	defer sub.Unsubscribe()
	fmt.Println("✓ 订阅成功！开始监听新区块...\n")

	// 5. Setup graceful shutdown - Listen for SIGINT (Ctrl+C) or SIGTERM
	// 5. 设置优雅退出 - 监听 SIGINT (Ctrl+C) 或 SIGTERM 信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 6. Main event loop - Process new blocks
	// 6. 主事件循环 - 处理新区块
	fmt.Println("========================================")
	fmt.Println("🎯 监听中... (按 Ctrl+C 退出)")
	fmt.Println("========================================\n")

	blockCount := 0

	for {
		select {
		case err := <-sub.Err():
			// WebSocket connection error or disconnection
			// WebSocket 连接错误或断开
			log.Fatalf("❌ 订阅错误 (Subscription error): %v", err)

		case header := <-headers:
			// New block received
			// 收到新区块
			blockCount++
			timestamp := time.Unix(int64(header.Time), 0)

			fmt.Printf("📦 新区块 #%d | Block Number: %d\n", blockCount, header.Number.Uint64())
			fmt.Printf("   ⏰ 时间 (Time): %s\n", timestamp.Format("2006-01-02 15:04:05"))
			fmt.Printf("   🔗 区块哈希 (Hash): %s\n", header.Hash().Hex())
			fmt.Printf("   ⛽ Gas 使用量 (Gas Used): %d\n", header.GasUsed)
			fmt.Printf("   🎯 Gas 上限 (Gas Limit): %d\n", header.GasLimit)
			fmt.Printf("   📊 难度 (Difficulty): %s\n", header.Difficulty.String())
			fmt.Println("   ----------------------------------------")

		case <-sigChan:
			// Graceful shutdown on Ctrl+C
			// 收到退出信号，优雅关闭
			fmt.Println("\n\n🛑 收到退出信号，正在关闭...")
			fmt.Printf("📊 共监听到 %d 个区块\n", blockCount)
			fmt.Println("👋 程序已安全退出")
			return
		}
	}
}
