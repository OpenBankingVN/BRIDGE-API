package ethereum

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
)

// PollingListener uses polling instead of subscriptions for free tier compatibility
type PollingListener struct {
	client    *Client
	listener  *ERC20Listener
	logger    zerolog.Logger
	lastBlock *big.Int
	seenTxs   map[string]bool
}

// NewPollingListener creates a new polling-based listener
func NewPollingListener(client *Client, logger zerolog.Logger) (*PollingListener, error) {
	listener, err := NewERC20Listener(client)
	if err != nil {
		return nil, fmt.Errorf("failed to create ERC20 listener: %w", err)
	}

	// Get current block number
	ctx := context.Background()
	latestBlock, err := client.GetLatestBlockNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest block: %w", err)
	}

	return &PollingListener{
		client:    client,
		listener:  listener,
		logger:    logger,
		lastBlock: latestBlock,
		seenTxs:   make(map[string]bool),
	}, nil
}

// StartPolling starts polling for new transfers
func (p *PollingListener) StartPolling(
	ctx context.Context,
	contractAddresses []common.Address,
	targetAddress *common.Address,
	pollInterval time.Duration,
	onTransfer func(event ERC20TransferEvent),
) error {

	p.logger.Info().
		Int("contracts_count", len(contractAddresses)).
		Dur("poll_interval", pollInterval).
		Msg("Starting polling for transfers")

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			p.logger.Info().Msg("Polling stopped")
			return nil
		case <-ticker.C:
			p.pollForNewTransfers(ctx, contractAddresses, targetAddress, onTransfer)
		}
	}
}

// pollForNewTransfers polls for new transfers since last check
func (p *PollingListener) pollForNewTransfers(
	ctx context.Context,
	contractAddresses []common.Address,
	targetAddress *common.Address,
	onTransfer func(event ERC20TransferEvent),
) {

	// Get current block number
	currentBlock, err := p.client.GetLatestBlockNumber(ctx)
	if err != nil {
		p.logger.Error().Err(err).Msg("Failed to get current block number")
		return
	}

	// If no new blocks, skip
	if currentBlock.Cmp(p.lastBlock) <= 0 {
		return
	}

	p.logger.Debug().
		Str("from_block", p.lastBlock.String()).
		Str("to_block", currentBlock.String()).
		Int("contracts_count", len(contractAddresses)).
		Msg("Checking for new transfers")

	// Check each contract for new transfers
	for _, contractAddr := range contractAddresses {
		transfers, err := p.listener.GetHistoricalTransfers(
			ctx, contractAddr, targetAddress, p.lastBlock, currentBlock)
		if err != nil {
			p.logger.Error().Err(err).
				Str("contract", contractAddr.Hex()).
				Msg("Failed to get historical transfers")
			continue
		}

		// Process new transfers
		for _, transfer := range transfers {
			txHash := transfer.TxHash.Hex()

			// Skip if we've already seen this transaction
			if p.seenTxs[txHash] {
				continue
			}

			// Mark as seen
			p.seenTxs[txHash] = true

			// Call the callback
			p.logger.Info().
				Str("tx_hash", txHash).
				Str("from", transfer.From.Hex()).
				Str("to", transfer.To.Hex()).
				Str("value", transfer.Value.String()).
				Uint64("block", transfer.BlockNumber).
				Msg("New transfer found")

			onTransfer(transfer)
		}
	}

	// Update last block
	p.lastBlock = currentBlock
}

// GetLastBlock returns the last processed block
func (p *PollingListener) GetLastBlock() *big.Int {
	return p.lastBlock
}
