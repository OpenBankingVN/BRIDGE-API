package ethereum

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
)

// Service provides high-level Ethereum operations for the bridge
type Service struct {
	client   *Client
	listener *ERC20Listener
	logger   zerolog.Logger
}

// NewService creates a new Ethereum service
func NewService(client *Client, logger zerolog.Logger) (*Service, error) {
	listener, err := NewERC20Listener(client)
	if err != nil {
		return nil, fmt.Errorf("failed to create ERC20 listener: %w", err)
	}

	return &Service{
		client:   client,
		listener: listener,
		logger:   logger,
	}, nil
}

// MonitorAddress monitors token transfers to/from a specific address
func (s *Service) MonitorAddress(
	ctx context.Context,
	targetAddress common.Address,
	contractAddresses []common.Address,
	onTransfer func(event ERC20TransferEvent),
) error {

	s.logger.Info().
		Str("target_address", targetAddress.Hex()).
		Int("contracts_count", len(contractAddresses)).
		Msg("Starting address monitoring")

	// Start monitoring for each contract
	for _, contractAddr := range contractAddresses {
		go func(contract common.Address) {
			s.monitorContract(ctx, contract, &targetAddress, onTransfer)
		}(contractAddr)
	}

	return nil
}

// MonitorAllTransfers monitors all transfers for specified contracts
func (s *Service) MonitorAllTransfers(
	ctx context.Context,
	contractAddresses []common.Address,
	onTransfer func(event ERC20TransferEvent),
) error {

	s.logger.Info().
		Int("contracts_count", len(contractAddresses)).
		Msg("Starting all transfers monitoring")

	// Start monitoring for each contract
	for _, contractAddr := range contractAddresses {
		go func(contract common.Address) {
			s.monitorContract(ctx, contract, nil, onTransfer)
		}(contractAddr)
	}

	return nil
}

// monitorContract monitors transfers for a specific contract
func (s *Service) monitorContract(
	ctx context.Context,
	contractAddress common.Address,
	targetAddress *common.Address,
	onTransfer func(event ERC20TransferEvent),
) {

	contractStr := contractAddress.Hex()
	s.logger.Info().
		Str("contract", contractStr).
		Msg("Starting contract monitoring")

	// Get latest block to start from
	latestBlock, err := s.client.GetLatestBlockNumber(ctx)
	if err != nil {
		s.logger.Error().Err(err).Str("contract", contractStr).Msg("Failed to get latest block")
		return
	}

	// Start from a few blocks ago to catch any recent transfers
	fromBlock := new(big.Int).Sub(latestBlock, big.NewInt(10))

	// Subscribe to new transfers
	events, sub, err := s.listener.ListenToTransfers(ctx, contractAddress, targetAddress, fromBlock)
	if err != nil {
		s.logger.Error().Err(err).Str("contract", contractStr).Msg("Failed to subscribe to transfers")
		return
	}
	defer sub.Unsubscribe()

	// Process events
	for {
		select {
		case <-ctx.Done():
			s.logger.Info().Str("contract", contractStr).Msg("Stopping contract monitoring")
			return
		case err := <-sub.Err():
			if err != nil {
				s.logger.Error().Err(err).Str("contract", contractStr).Msg("Subscription error")
				// Wait before retrying
				time.Sleep(5 * time.Second)
				continue
			}
		case event := <-events:
			s.logger.Info().
				Str("contract", contractStr).
				Str("tx_hash", event.TxHash.Hex()).
				Str("from", event.From.Hex()).
				Str("to", event.To.Hex()).
				Str("value", event.Value.String()).
				Uint64("block", event.BlockNumber).
				Msg("Transfer event received")

			// Call the callback function
			onTransfer(event)
		}
	}
}

// GetHistoricalTransfers gets historical transfers for an address
func (s *Service) GetHistoricalTransfers(
	ctx context.Context,
	contractAddress common.Address,
	targetAddress *common.Address,
	fromBlock, toBlock *big.Int,
) ([]ERC20TransferEvent, error) {

	return s.listener.GetHistoricalTransfers(ctx, contractAddress, targetAddress, fromBlock, toBlock)
}

// GetTokenBalance gets the token balance for an address
func (s *Service) GetTokenBalance(
	ctx context.Context,
	contractAddress common.Address,
	address common.Address,
) (*big.Int, error) {

	return s.listener.GetTokenBalance(ctx, contractAddress, address)
}

// GetETHBalance gets the ETH balance for an address
func (s *Service) GetETHBalance(ctx context.Context, address common.Address) (*big.Int, error) {
	return s.client.GetBalance(ctx, address)
}

// ValidateAddress validates if an address is valid
func (s *Service) ValidateAddress(address string) (common.Address, error) {
	if !common.IsHexAddress(address) {
		return common.Address{}, fmt.Errorf("invalid address format: %s", address)
	}
	return common.HexToAddress(address), nil
}

// Close closes the service
func (s *Service) Close() {
	if s.client != nil {
		s.client.Close()
	}
}
