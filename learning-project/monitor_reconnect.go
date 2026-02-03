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

// Block Monitor with Auto-Reconnection
// 带自动重连的区块监听器
func main() {
	// Load configuration
	// 加载配置
	fmt.Println("[初始化] 正在加载 .env 配置...")
	err := godotenv.Load()
	if err != nil {
		log.Fatal("❌ 加载 .env 文件失败")
	}

	wssURL := os.Getenv("RPC_WSS_URL")
	if wssURL == "" {
		log.Fatal("❌ RPC_WSS_URL 未配置")
	}
	fmt.Printf("✓ 配置加载成功\n[配置] WSS 节点: %s\n\n", wssURL)

	// Setup graceful shutdown
	// 设置优雅退出
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	blockCount := 0
	reconnectDelay := 3 * time.Second // Reconnection delay / 重连延迟
	maxReconnectDelay := 30 * time.Second // Max reconnection delay / 最大重连延迟

	// Main reconnection loop
	// 主重连循环
	for {
		select {
		case <-sigChan:
			fmt.Println("\n\n🛑 收到退出信号")
			fmt.Printf("📊 共监听到 %d 个区块\n", blockCount)
			fmt.Println("👋 程序已安全退出")
			return

		default:
			// Try to connect and subscribe
			// 尝试连接并订阅
			err := runMonitor(wssURL, &blockCount, sigChan)
			if err != nil {
				fmt.Printf("\n⚠️  连接断开: %v\n", err)
				fmt.Printf("⏳ %d 秒后自动重连...\n\n", int(reconnectDelay.Seconds()))
				
				// Wait before reconnecting (with interrupt check)
				// 等待重连（可被中断）
				timer := time.NewTimer(reconnectDelay)
				select {
				case <-timer.C:
					// Increase delay for next reconnection (exponential backoff)
					// 增加下次重连延迟（指数退避）
					reconnectDelay *= 2
					if reconnectDelay > maxReconnectDelay {
						reconnectDelay = maxReconnectDelay
					}
				case <-sigChan:
					timer.Stop()
					fmt.Println("\n🛑 收到退出信号")
					fmt.Printf("📊 共监听到 %d 个区块\n", blockCount)
					fmt.Println("👋 程序已安全退出")
					return
				}
			} else {
				// Reset reconnection delay on successful connection
				// 连接成功后重置重连延迟
				reconnectDelay = 3 * time.Second
			}
		}
	}
}

// runMonitor - Establish connection, subscribe, and listen for blocks
// runMonitor - 建立连接、订阅并监听区块
func runMonitor(wssURL string, blockCount *int, sigChan chan os.Signal) error {
	fmt.Println("[连接] 正在建立 WebSocket 连接...")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	client, err := ethclient.DialContext(ctx, wssURL)
	if err != nil {
		return fmt.Errorf("连接失败: %w", err)
	}
	defer client.Close()
	fmt.Println("✓ WebSocket 连接成功！")

	// Subscribe to new block headers
	// 订阅新区块头
	fmt.Println("[订阅] 正在订阅新区块事件 (newHeads)...")
	headers := make(chan *types.Header)
	sub, err := client.SubscribeNewHead(context.Background(), headers)
	if err != nil {
		return fmt.Errorf("订阅失败: %w", err)
	}
	defer sub.Unsubscribe()
	fmt.Println("✓ 订阅成功！开始监听新区块...\n")
	fmt.Println("========================================")
	fmt.Println("🎯 监听中... (按 Ctrl+C 退出)")
	fmt.Println("========================================\n")

	// Heartbeat - Check connection health every 30 seconds
	// 心跳检测 - 每30秒检查连接健康状态
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Main event loop
	// 主事件循环
	for {
		select {
		case err := <-sub.Err():
			return fmt.Errorf("订阅错误: %w", err)

		case header := <-headers:
			*blockCount++
			timestamp := time.Unix(int64(header.Time), 0)

			fmt.Printf("📦 新区块 #%d | Block Number: %d\n", *blockCount, header.Number.Uint64())
			fmt.Printf("   ⏰ 时间: %s\n", timestamp.Format("2006-01-02 15:04:05"))
			fmt.Printf("   🔗 区块哈希: %s\n", header.Hash().Hex())
			fmt.Printf("   ⛽ Gas 使用: %d / %d\n", header.GasUsed, header.GasLimit)
			fmt.Printf("   📊 难度: %s\n", header.Difficulty.String())
			fmt.Println("   ----------------------------------------")

		case <-ticker.C:
			// Heartbeat check - Verify connection is alive
			// 心跳检查 - 验证连接是否存活
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			_, err := client.BlockNumber(ctx)
			cancel()
			if err != nil {
				return fmt.Errorf("心跳检测失败: %w", err)
			}
			fmt.Println("💓 心跳正常")

		case <-sigChan:
			// User interrupt - Exit gracefully
			// 用户中断 - 优雅退出
			return nil
		}
	}
}
