package ethereum

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
)

// ETHTransferEvent represents an ETH transfer
type ETHTransferEvent struct {
	From        common.Address `json:"from"`
	To          common.Address `json:"to"`
	Value       *big.Int       `json:"value"`
	TxHash      common.Hash    `json:"txHash"`
	BlockNumber uint64         `json:"blockNumber"`
	GasUsed     uint64         `json:"gasUsed"`
	GasPrice    *big.Int       `json:"gasPrice"`
}

// ETHTransferListener handles listening to ETH transfers
type ETHTransferListener struct {
	client    *Client
	logger    zerolog.Logger
	lastBlock *big.Int
	seenTxs   map[string]bool
}

// NewETHTransferListener creates a new ETH transfer listener
func NewETHTransferListener(client *Client, logger zerolog.Logger) (*ETHTransferListener, error) {
	// Get current block number
	ctx := context.Background()
	latestBlock, err := client.GetLatestBlockNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest block: %w", err)
	}

	return &ETHTransferListener{
		client:    client,
		logger:    logger,
		lastBlock: latestBlock,
		seenTxs:   make(map[string]bool),
	}, nil
}

// StartPollingETH starts polling for ETH transfers
func (e *ETHTransferListener) StartPollingETH(
	ctx context.Context,
	targetAddresses []common.Address,
	pollInterval time.Duration,
	onTransfer func(event ETHTransferEvent),
) error {

	e.logger.Info().
		Int("target_addresses_count", len(targetAddresses)).
		Dur("poll_interval", pollInterval).
		Msg("Starting polling for ETH transfers")

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			e.logger.Info().Msg("ETH transfer polling stopped")
			return nil
		case <-ticker.C:
			e.pollForNewETHTransfers(ctx, targetAddresses, onTransfer)
		}
	}
}

// pollForNewETHTransfers polls for new ETH transfers since last check
func (e *ETHTransferListener) pollForNewETHTransfers(
	ctx context.Context,
	targetAddresses []common.Address,
	onTransfer func(event ETHTransferEvent),
) {

	// Get current block number
	currentBlock, err := e.client.GetLatestBlockNumber(ctx)
	if err != nil {
		e.logger.Error().Err(err).Msg("Failed to get current block number")
		return
	}

	// If no new blocks, skip
	if currentBlock.Cmp(e.lastBlock) <= 0 {
		return
	}

	e.logger.Debug().
		Str("from_block", e.lastBlock.String()).
		Str("to_block", currentBlock.String()).
		Int("target_addresses_count", len(targetAddresses)).
		Msg("Checking for new ETH transfers")

	// Check each block for ETH transfers
	for blockNum := new(big.Int).Add(e.lastBlock, big.NewInt(1)); blockNum.Cmp(currentBlock) <= 0; blockNum.Add(blockNum, big.NewInt(1)) {
		e.checkBlockForETHTransfers(ctx, blockNum, targetAddresses, onTransfer)
	}

	// Update last block
	e.lastBlock = currentBlock
}

// checkBlockForETHTransfers checks a specific block for ETH transfers
func (e *ETHTransferListener) checkBlockForETHTransfers(
	ctx context.Context,
	blockNumber *big.Int,
	targetAddresses []common.Address,
	onTransfer func(event ETHTransferEvent),
) {

	// Get block by number
	block, err := e.client.GetClient().BlockByNumber(ctx, blockNumber)
	if err != nil {
		e.logger.Error().Err(err).
			Str("block_number", blockNumber.String()).
			Msg("Failed to get block")
		return
	}

	// Check each transaction in the block
	for _, tx := range block.Transactions() {
		// Skip if we've already seen this transaction
		txHash := tx.Hash().Hex()
		if e.seenTxs[txHash] {
			continue
		}

		// Skip contract creation transactions
		if tx.To() == nil {
			continue
		}

		// Check if this transaction involves any of our target addresses
		from, err := e.client.GetClient().TransactionSender(ctx, tx, block.Hash(), 0)
		if err != nil {
			continue
		}

		to := *tx.To()

		// Check if transaction involves our target addresses
		involvesTarget := false
		for _, targetAddr := range targetAddresses {
			if from == targetAddr || to == targetAddr {
				involvesTarget = true
				break
			}
		}

		if !involvesTarget {
			continue
		}

		// Get transaction receipt for gas information
		receipt, err := e.client.GetClient().TransactionReceipt(ctx, tx.Hash())
		if err != nil {
			continue
		}

		// Mark as seen
		e.seenTxs[txHash] = true

		// Create ETH transfer event
		event := ETHTransferEvent{
			From:        from,
			To:          to,
			Value:       tx.Value(),
			TxHash:      tx.Hash(),
			BlockNumber: block.NumberU64(),
			GasUsed:     receipt.GasUsed,
			GasPrice:    tx.GasPrice(),
		}

		// Call the callback
		e.logger.Info().
			Str("tx_hash", txHash).
			Str("from", from.Hex()).
			Str("to", to.Hex()).
			Str("value", tx.Value().String()).
			Uint64("block", block.NumberU64()).
			Msg("New ETH transfer found")

		onTransfer(event)
	}
}

// GetLastBlock returns the last processed block
func (e *ETHTransferListener) GetLastBlock() *big.Int {
	return e.lastBlock
}
