package dex

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	log "github.com/sirupsen/logrus"

	"github.com/ljlin/mev-arbitrage-bot/pkg/config"
)

// PoolMonitor monitors multiple liquidity pools for reserve changes
type PoolMonitor struct {
	client   *ethclient.Client
	adapters map[DEXType]DEXAdapter
	pools    map[common.Address]*Pool
	mu       sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
	interval time.Duration
}

// NewPoolMonitor creates a new pool monitor
func NewPoolMonitor(client *ethclient.Client, cfg *config.Config) *PoolMonitor {
	ctx, cancel := context.WithCancel(context.Background())

	monitor := &PoolMonitor{
		client:   client,
		adapters: make(map[DEXType]DEXAdapter),
		pools:    make(map[common.Address]*Pool),
		ctx:      ctx,
		cancel:   cancel,
		interval: time.Duration(cfg.PoolMonitorInterval) * time.Second,
	}

	return monitor
}

// RegisterAdapter registers a DEX adapter
func (pm *PoolMonitor) RegisterAdapter(adapter DEXAdapter) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.adapters[adapter.GetType()] = adapter
	log.Infof("Registered DEX adapter: %s", adapter.GetName())
}

// AddPool adds a pool to monitor
func (pm *PoolMonitor) AddPool(pool *Pool) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Verify adapter exists for this DEX type
	if _, exists := pm.adapters[pool.DEX]; !exists {
		return fmt.Errorf("no adapter registered for DEX type: %s", pool.DEX)
	}

	pm.pools[pool.Address] = pool
	log.Infof("Added pool to monitor: %s (%s)", pool.Address.Hex(), pool.DEX)

	return nil
}

// GetPool retrieves a monitored pool
func (pm *PoolMonitor) GetPool(address common.Address) (*Pool, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	pool, exists := pm.pools[address]
	if !exists {
		return nil, fmt.Errorf("pool not found: %s", address.Hex())
	}

	// Return a copy to prevent race conditions
	poolCopy := *pool
	return &poolCopy, nil
}

// GetAllPools returns all monitored pools
func (pm *PoolMonitor) GetAllPools() []*Pool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	pools := make([]*Pool, 0, len(pm.pools))
	for _, pool := range pm.pools {
		poolCopy := *pool
		pools = append(pools, &poolCopy)
	}

	return pools
}

// UpdatePool updates a pool's reserves
func (pm *PoolMonitor) UpdatePool(address common.Address) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pool, exists := pm.pools[address]
	if !exists {
		return fmt.Errorf("pool not found: %s", address.Hex())
	}

	adapter, exists := pm.adapters[pool.DEX]
	if !exists {
		return fmt.Errorf("adapter not found for DEX: %s", pool.DEX)
	}

	// Fetch new reserves
	reserve0, reserve1, err := adapter.GetReserves(address)
	if err != nil {
		return fmt.Errorf("failed to fetch reserves: %w", err)
	}

	// Update pool
	pool.Reserve0 = reserve0
	pool.Reserve1 = reserve1
	pool.LastUpdated = time.Now().Unix()

	log.Debugf("Updated pool %s: Reserve0=%s, Reserve1=%s",
		address.Hex(), reserve0.String(), reserve1.String())

	return nil
}

// Start begins monitoring pools
func (pm *PoolMonitor) Start() {
	log.Info("Starting pool monitor...")

	go pm.monitorLoop()

	log.Infof("Pool monitor started (interval: %v)", pm.interval)
}

// Stop stops the pool monitor
func (pm *PoolMonitor) Stop() {
	log.Info("Stopping pool monitor...")
	pm.cancel()
}

// monitorLoop periodically updates all pools
func (pm *PoolMonitor) monitorLoop() {
	ticker := time.NewTicker(pm.interval)
	defer ticker.Stop()

	for {
		select {
		case <-pm.ctx.Done():
			log.Info("Pool monitor stopped")
			return

		case <-ticker.C:
			pm.updateAllPools()
		}
	}
}

// updateAllPools updates reserves for all monitored pools
func (pm *PoolMonitor) updateAllPools() {
	pm.mu.RLock()
	addresses := make([]common.Address, 0, len(pm.pools))
	for addr := range pm.pools {
		addresses = append(addresses, addr)
	}
	pm.mu.RUnlock()

	var wg sync.WaitGroup
	for _, addr := range addresses {
		wg.Add(1)
		go func(address common.Address) {
			defer wg.Done()

			if err := pm.UpdatePool(address); err != nil {
				log.Warnf("Failed to update pool %s: %v", address.Hex(), err)
			}
		}(addr)
	}

	wg.Wait()
	log.Debugf("Updated %d pools", len(addresses))
}

// SubscribeNewBlocks subscribes to new blocks and updates pools
func (pm *PoolMonitor) SubscribeNewBlocks(wssClient *ethclient.Client) error {
	headers := make(chan *types.Header)

	sub, err := wssClient.SubscribeNewHead(pm.ctx, headers)
	if err != nil {
		return fmt.Errorf("failed to subscribe to new heads: %w", err)
	}

	log.Info("Subscribed to new blocks for pool monitoring")

	go func() {
		defer sub.Unsubscribe()

		for {
			select {
			case <-pm.ctx.Done():
				return

			case err := <-sub.Err():
				log.Errorf("Block subscription error: %v", err)
				return

			case header := <-headers:
				log.Debugf("New block %d, updating pools...", header.Number.Uint64())
				pm.updateAllPools()
			}
		}
	}()

	return nil
}

// GetPoolsByDEX returns all pools for a specific DEX
func (pm *PoolMonitor) GetPoolsByDEX(dexType DEXType) []*Pool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	pools := make([]*Pool, 0)
	for _, pool := range pm.pools {
		if pool.DEX == dexType {
			poolCopy := *pool
			pools = append(pools, &poolCopy)
		}
	}

	return pools
}

// GetPoolForTokens finds a pool for the given token pair
func (pm *PoolMonitor) GetPoolForTokens(token0, token1 common.Address, dexType DEXType) (*Pool, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	for _, pool := range pm.pools {
		if pool.DEX != dexType {
			continue
		}

		if (pool.Token0 == token0 && pool.Token1 == token1) ||
			(pool.Token0 == token1 && pool.Token1 == token0) {
			poolCopy := *pool
			return &poolCopy, nil
		}
	}

	return nil, fmt.Errorf("pool not found for tokens %s/%s on %s",
		token0.Hex(), token1.Hex(), dexType)
}
