package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Load the .env file from the current directory
	// 1. 从当前目录加载 .env 文件
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
		// 加载 .env 文件出错
	}

	// 2. Get the HTTPS RPC URL from environment variables
	// 2. 从环境变量中获取 HTTPS RPC 链接
	rpcURL := os.Getenv("RPC_HTTPS_URL")
	// 3. Get the Public Wallet Address from environment variables
	// 3. 从环境变量中获取钱包公钥地址
	walletAddress := os.Getenv("PUBLIC_ADDRESS")

	// 4. Connect to the Ethereum node (Sepolia)
	// 4. 连接到以太坊节点 (Sepolia 测试网)
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Successfully connected to Ethereum network!")
	// 成功连接到以太坊网络！

	// 5. Convert the string address to a common.Address type
	// 5. 将字符串地址转换为以太坊通用的 Address 类型
	account := common.HexToAddress(walletAddress)

	// 6. Query the balance of the account (nil means get the latest block balance)
	// 6. 查询账户余额 (nil 表示获取最新区块的余额)
	balance, err := client.BalanceAt(context.Background(), account, nil)
	if err != nil {
		log.Fatal(err)
	}

	// 7. Convert the balance from Wei to ETH (1 ETH = 10^18 Wei)
	// 7. 将余额从 Wei 转换为 ETH 单位 (1 ETH = 10^18 Wei)
	fbalance := new(big.Float)
	fbalance.SetString(balance.String())
	ethValue := new(big.Float).Quo(fbalance, big.NewFloat(math.Pow10(18)))

	fmt.Printf("Wallet Address: %s\n", walletAddress)
	// 钱包地址: %s
	fmt.Printf("Account Balance: %f ETH\n", ethValue)
	// 账户余额: %f ETH
}
