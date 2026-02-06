package flashbots

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// FlashbotsBundle represents a bundle of transactions to be sent via Flashbots
// FlashbotsBundle 表示通过 Flashbots 发送的交易捆绑包
type FlashbotsBundle struct {
	Transactions    []*types.Transaction // 要发送的交易列表
	BlockNumber     uint64               // 目标区块号
	MinTimestamp    uint64               // 最小时间戳
	MaxTimestamp    uint64               // 最大时间戳
	RevertingHashes []common.Hash        // 允许失败的交易哈希
}

// BundleResponse represents the response from Flashbots relay
// BundleResponse 表示 Flashbots 中继的响应
type BundleResponse struct {
	BundleHash common.Hash // Bundle 哈希
	Success    bool        // 是否成功
	Error      string      // 错误信息
}

// SimulationResult represents the result of a bundle simulation
// SimulationResult 表示 Bundle 模拟的结果
type SimulationResult struct {
	Success          bool     // 模拟是否成功
	GasUsed          uint64   // 使用的 Gas
	GasPrice         *big.Int // Gas 价格
	CoinbaseDiff     *big.Int // 矿工收益差异
	TotalGasFees     *big.Int // 总 Gas 费用
	StateBlockNumber uint64   // 状态区块号
}
