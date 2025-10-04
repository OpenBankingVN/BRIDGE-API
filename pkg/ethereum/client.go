package ethereum

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog"
)

// Client wraps Ethereum client with additional functionality
type Client struct {
	client  *ethclient.Client
	logger  zerolog.Logger
	rpcURL  string
	chainID *big.Int
}

// NewClient creates a new Ethereum client
func NewClient(rpcURL string, logger zerolog.Logger) (*Client, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum client: %w", err)
	}

	// Get chain ID
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get network ID: %w", err)
	}

	return &Client{
		client:  client,
		logger:  logger,
		rpcURL:  rpcURL,
		chainID: chainID,
	}, nil
}

// Close closes the Ethereum client connection
func (c *Client) Close() {
	if c.client != nil {
		c.client.Close()
	}
}

// GetBalance returns the ETH balance of an address
func (c *Client) GetBalance(ctx context.Context, address common.Address) (*big.Int, error) {
	balance, err := c.client.BalanceAt(ctx, address, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}
	return balance, nil
}

// GetLatestBlockNumber returns the latest block number
func (c *Client) GetLatestBlockNumber(ctx context.Context) (*big.Int, error) {
	header, err := c.client.HeaderByNumber(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest header: %w", err)
	}
	return header.Number, nil
}

// SubscribeToNewBlocks subscribes to new block headers
func (c *Client) SubscribeToNewBlocks(ctx context.Context) (<-chan *types.Header, ethereum.Subscription, error) {
	headers := make(chan *types.Header)
	sub, err := c.client.SubscribeNewHead(ctx, headers)
	if err != nil {
		close(headers)
		return nil, nil, fmt.Errorf("failed to subscribe to new headers: %w", err)
	}
	return headers, sub, nil
}

// GetChainID returns the chain ID
func (c *Client) GetChainID() *big.Int {
	return c.chainID
}

// GetClient returns the underlying ethclient.Client
func (c *Client) GetClient() *ethclient.Client {
	return c.client
}
