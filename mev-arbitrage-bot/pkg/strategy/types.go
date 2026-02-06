package strategy

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ljlin/mev-arbitrage-bot/pkg/dex"
)

// ArbitragePath represents a profitable arbitrage opportunity
type ArbitragePath struct {
	ID           string           // Unique identifier
	Pools        []*dex.Pool      // Sequence of pools to trade through
	Tokens       []common.Address // Token addresses in order
	StartToken   common.Address   // Starting token
	StartAmount  *big.Int         // Initial amount
	EndAmount    *big.Int         // Final amount after arbitrage
	Profit       *big.Int         // Profit amount (EndAmount - StartAmount)
	ProfitBps    int              // Profit in basis points
	ProfitETH    *big.Float       // Profit in ETH (for display)
	GasCostEst   *big.Int         // Estimated gas cost
	NetProfit    *big.Int         // Profit after gas
	NetProfitBps int              // Net profit in basis points
	PriceImpact  *big.Float       // Total price impact
	Timestamp    int64            // Discovery timestamp
}

// ArbitrageOpportunity represents a validated arbitrage opportunity
type ArbitrageOpportunity struct {
	Path         *ArbitragePath
	IsExecutable bool
	Reason       string // Reason if not executable
	Priority     int    // Execution priority (higher = more urgent)
}
