package blockchain

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	log "github.com/sirupsen/logrus"

	"github.com/ljlin/mev-arbitrage-bot/pkg/config"
)

// Client wraps Ethereum RPC client with reconnection logic
type Client struct {
	httpClient *ethclient.Client
	wssClient  *ethclient.Client
	config     *config.Config
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewClient creates a new blockchain client
func NewClient(cfg *config.Config) (*Client, error) {
	ctx, cancel := context.WithCancel(context.Background())

	client := &Client{
		config: cfg,
		ctx:    ctx,
		cancel: cancel,
	}

	if err := client.connectHTTP(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to connect HTTP client: %w", err)
	}

	if err := client.connectWSS(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to connect WSS client: %w", err)
	}

	log.Info("Blockchain client initialized successfully")
	return client, nil
}

// connectHTTP establishes HTTPS RPC connection
func (c *Client) connectHTTP() error {
	ctx, cancel := context.WithTimeout(c.ctx, time.Duration(c.config.ConnectionTimeout)*time.Second)
	defer cancel()

	client, err := ethclient.DialContext(ctx, c.config.RPCHTTPSUrl)
	if err != nil {
		return fmt.Errorf("HTTPS dial failed: %w", err)
	}

	// Test connection
	if _, err := client.ChainID(ctx); err != nil {
		client.Close()
		return fmt.Errorf("HTTPS connection test failed: %w", err)
	}

	c.httpClient = client
	log.Info("HTTPS RPC client connected")
	return nil
}

// connectWSS establishes WebSocket connection
func (c *Client) connectWSS() error {
	ctx, cancel := context.WithTimeout(c.ctx, time.Duration(c.config.ConnectionTimeout)*time.Second)
	defer cancel()

	client, err := ethclient.DialContext(ctx, c.config.RPCWSSUrl)
	if err != nil {
		return fmt.Errorf("WSS dial failed: %w", err)
	}

	// Test connection
	if _, err := client.ChainID(ctx); err != nil {
		client.Close()
		return fmt.Errorf("WSS connection test failed: %w", err)
	}

	c.wssClient = client
	log.Info("WebSocket RPC client connected")
	return nil
}

// GetHTTPClient returns the HTTP client
func (c *Client) GetHTTPClient() *ethclient.Client {
	return c.httpClient
}

// GetWSSClient returns the WebSocket client
func (c *Client) GetWSSClient() *ethclient.Client {
	return c.wssClient
}

// GetBalance returns the balance of an address
func (c *Client) GetBalance(address common.Address) (*big.Int, error) {
	ctx, cancel := context.WithTimeout(c.ctx, 5*time.Second)
	defer cancel()

	balance, err := c.httpClient.BalanceAt(ctx, address, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	return balance, nil
}

// GetBlockNumber returns the latest block number
func (c *Client) GetBlockNumber() (uint64, error) {
	ctx, cancel := context.WithTimeout(c.ctx, 5*time.Second)
	defer cancel()

	blockNumber, err := c.httpClient.BlockNumber(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get block number: %w", err)
	}

	return blockNumber, nil
}

// GetBlock returns a block by number
func (c *Client) GetBlock(number uint64) (*types.Block, error) {
	ctx, cancel := context.WithTimeout(c.ctx, 5*time.Second)
	defer cancel()

	block, err := c.httpClient.BlockByNumber(ctx, big.NewInt(int64(number)))
	if err != nil {
		return nil, fmt.Errorf("failed to get block: %w", err)
	}

	return block, nil
}

// GetGasPrice returns the suggested gas price
func (c *Client) GetGasPrice() (*big.Int, error) {
	ctx, cancel := context.WithTimeout(c.ctx, 5*time.Second)
	defer cancel()

	gasPrice, err := c.httpClient.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %w", err)
	}

	return gasPrice, nil
}

// GetChainID returns the chain ID
func (c *Client) GetChainID() (*big.Int, error) {
	ctx, cancel := context.WithTimeout(c.ctx, 5*time.Second)
	defer cancel()

	chainID, err := c.httpClient.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}

	return chainID, nil
}

// Close closes all client connections
func (c *Client) Close() {
	c.cancel()
	if c.httpClient != nil {
		c.httpClient.Close()
		log.Info("HTTPS client closed")
	}
	if c.wssClient != nil {
		c.wssClient.Close()
		log.Info("WSS client closed")
	}
}

// HealthCheck performs a health check on the connection
func (c *Client) HealthCheck() error {
	ctx, cancel := context.WithTimeout(c.ctx, 5*time.Second)
	defer cancel()

	if _, err := c.httpClient.BlockNumber(ctx); err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	return nil
}
