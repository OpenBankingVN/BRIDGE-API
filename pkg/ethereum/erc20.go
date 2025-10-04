package ethereum

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// ERC20TransferEvent represents a Transfer event from ERC-20 contract
type ERC20TransferEvent struct {
	From        common.Address `json:"from"`
	To          common.Address `json:"to"`
	Value       *big.Int       `json:"value"`
	TxHash      common.Hash    `json:"txHash"`
	BlockNumber uint64         `json:"blockNumber"`
	LogIndex    uint           `json:"logIndex"`
}

// ERC20Listener handles listening to ERC-20 Transfer events
type ERC20Listener struct {
	client      *Client
	contractABI abi.ABI
}

// NewERC20Listener creates a new ERC-20 event listener
func NewERC20Listener(client *Client) (*ERC20Listener, error) {
	// ERC-20 Transfer event ABI
	contractABI, err := abi.JSON(strings.NewReader(`[
		{
			"anonymous": false,
			"inputs": [
				{
					"indexed": true,
					"name": "from",
					"type": "address"
				},
				{
					"indexed": true,
					"name": "to",
					"type": "address"
				},
				{
					"indexed": false,
					"name": "value",
					"type": "uint256"
				}
			],
			"name": "Transfer",
			"type": "event"
		}
	]`))
	if err != nil {
		return nil, fmt.Errorf("failed to parse contract ABI: %w", err)
	}

	return &ERC20Listener{
		client:      client,
		contractABI: contractABI,
	}, nil
}

// ListenToTransfers listens to Transfer events for a specific contract and address
func (l *ERC20Listener) ListenToTransfers(
	ctx context.Context,
	contractAddress common.Address,
	targetAddress *common.Address, // nil means listen to all transfers
	fromBlock *big.Int,
) (<-chan ERC20TransferEvent, ethereum.Subscription, error) {

	// Create filter query - start with basic contract and event filter
	query := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddress},
		FromBlock: fromBlock,
		Topics:    [][]common.Hash{{l.contractABI.Events["Transfer"].ID}},
	}

	// For address filtering, we'll use a different approach:
	// Instead of filtering at the subscription level (which can be buggy),
	// we'll subscribe to all Transfer events and filter in the processing goroutine

	// Subscribe to logs
	logs := make(chan types.Log)
	sub, err := l.client.GetClient().SubscribeFilterLogs(ctx, query, logs)
	if err != nil {
		close(logs)
		return nil, nil, fmt.Errorf("failed to subscribe to logs: %w", err)
	}

	// Create channel for processed events
	events := make(chan ERC20TransferEvent, 100)

	// Start goroutine to process logs
	go func() {
		defer close(events)
		for {
			select {
			case <-ctx.Done():
				return
			case log := <-logs:
				event, err := l.parseTransferEvent(log)
				if err != nil {
					l.client.logger.Error().Err(err).Msg("Failed to parse transfer event")
					continue
				}

				// Apply address filtering in the processing goroutine
				if targetAddress != nil {
					// Check if this transfer involves our target address (either from or to)
					if event.To != *targetAddress && event.From != *targetAddress {
						continue
					}
				}

				select {
				case events <- event:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return events, sub, nil
}

// GetHistoricalTransfers gets historical Transfer events
func (l *ERC20Listener) GetHistoricalTransfers(
	ctx context.Context,
	contractAddress common.Address,
	targetAddress *common.Address,
	fromBlock, toBlock *big.Int,
) ([]ERC20TransferEvent, error) {

	// Create filter query
	query := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddress},
		FromBlock: fromBlock,
		ToBlock:   toBlock,
		Topics:    [][]common.Hash{{l.contractABI.Events["Transfer"].ID}},
	}

	// Get logs (no address filtering at query level)
	logs, err := l.client.GetClient().FilterLogs(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to filter logs: %w", err)
	}

	// Parse events and apply address filtering
	events := make([]ERC20TransferEvent, 0, len(logs))
	for _, log := range logs {
		event, err := l.parseTransferEvent(log)
		if err != nil {
			l.client.logger.Error().Err(err).Msg("Failed to parse transfer event")
			continue
		}

		// Apply address filtering if specified
		if targetAddress != nil {
			// Check if this transfer involves our target address (either from or to)
			if event.To != *targetAddress && event.From != *targetAddress {
				continue
			}
		}

		events = append(events, event)
	}

	return events, nil
}

// parseTransferEvent parses a log into a Transfer event
func (l *ERC20Listener) parseTransferEvent(log types.Log) (ERC20TransferEvent, error) {
	// Unpack the event data
	var transferEvent struct {
		From  common.Address
		To    common.Address
		Value *big.Int
	}

	// Parse the log data
	err := l.contractABI.UnpackIntoInterface(&transferEvent, "Transfer", log.Data)
	if err != nil {
		return ERC20TransferEvent{}, fmt.Errorf("failed to unpack event data: %w", err)
	}

	// Extract indexed parameters from topics
	if len(log.Topics) < 3 {
		return ERC20TransferEvent{}, fmt.Errorf("invalid number of topics in log")
	}

	transferEvent.From = common.BytesToAddress(log.Topics[1].Bytes())
	transferEvent.To = common.BytesToAddress(log.Topics[2].Bytes())

	return ERC20TransferEvent{
		From:        transferEvent.From,
		To:          transferEvent.To,
		Value:       transferEvent.Value,
		TxHash:      log.TxHash,
		BlockNumber: log.BlockNumber,
		LogIndex:    log.Index,
	}, nil
}

// GetTokenBalance returns the token balance of an address for a specific ERC-20 contract
func (l *ERC20Listener) GetTokenBalance(
	ctx context.Context,
	contractAddress common.Address,
	address common.Address,
) (*big.Int, error) {
	// Create a simple ABI for balanceOf function
	balanceOfABI := `[{"inputs":[{"name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"balance","type":"uint256"}],"stateMutability":"view","type":"function"}]`

	contractABI, err := abi.JSON(strings.NewReader(balanceOfABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse balanceOf ABI: %w", err)
	}

	// Encode the function call
	data, err := contractABI.Pack("balanceOf", address)
	if err != nil {
		return nil, fmt.Errorf("failed to pack balanceOf call: %w", err)
	}

	// Make the call
	result, err := l.client.GetClient().CallContract(ctx, ethereum.CallMsg{
		To:   &contractAddress,
		Data: data,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to call balanceOf: %w", err)
	}

	// Decode the result
	var balance *big.Int
	err = contractABI.UnpackIntoInterface(&balance, "balanceOf", result)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack balanceOf result: %w", err)
	}

	return balance, nil
}
