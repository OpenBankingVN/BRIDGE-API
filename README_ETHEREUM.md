# Ethereum Token Transfer Listener for Sepolia Testnet

This implementation provides a comprehensive Go solution for listening to token transfer events on the Sepolia testnet using the Ethereum go-ethereum library.

## Features

- ✅ **Real-time Transfer Monitoring**: Listen to ERC-20 Transfer events in real-time
- ✅ **Address Filtering**: Monitor specific addresses for incoming/outgoing transfers
- ✅ **Multiple Contract Support**: Monitor multiple token contracts simultaneously
- ✅ **Historical Data**: Retrieve historical transfer events
- ✅ **Balance Queries**: Get ETH and token balances
- ✅ **Sepolia Testnet Ready**: Pre-configured for Sepolia testnet
- ✅ **Graceful Shutdown**: Proper cleanup and error handling

## Quick Start

### 1. Configuration

Update your `config/config.yaml` with your Ethereum RPC endpoint:

```yaml
ETHEREUM:
  SEPOLIA_RPC_URL: "https://eth-sepolia.g.alchemy.com/v2/YOUR_API_KEY"
  CONTRACTS:
    USDC: "0x1c7D4B196Cb0C7B01d743Fbc6116a902379C7238"
    USDT: "0x7169D38820dfd117C3FA1f22a697dBA58d90BA06"
    DAI: "0xFF34B3d4Aee8ddCd6F9AFFFB6Fe49bD371b8a357"
```

### 2. Run the Example

```bash
# Run the token transfer listener example
go run examples/eth_listenerBurn/main.go
```

### 3. Basic Usage

```go
package main

import (
    "context"
    "log"
    
    "github.com/OpenBankingVN/BRIDGE-API/config"
    "github.com/OpenBankingVN/BRIDGE-API/pkg/ethereum"
    "github.com/ethereum/go-ethereum/common"
)

func main() {
    // Load config
    cfg, _ := config.NewConfig()
    
    // Create Ethereum client
    client, _ := ethereum.NewClient(cfg.Ethereum.SepoliaRPCURL, logger)
    defer client.Close()
    
    // Create service
    service, _ := ethereum.NewService(client, logger)
    defer service.Close()
    
    // Monitor transfers to your address
    targetAddr := common.HexToAddress("0x742d35Cc6634C0532925a3b8D6Ac6E2f596C2c8d")
    contractAddr := common.HexToAddress(cfg.Ethereum.Contracts["USDC"])
    
    ctx := context.Background()
    
    // Start monitoring
    service.MonitorAddress(ctx, targetAddr, []common.Address{contractAddr}, 
        func(event ethereum.ERC20TransferEvent) {
            log.Printf("Transfer: %s -> %s, Value: %s", 
                event.From.Hex(), event.To.Hex(), event.Value.String())
        })
}
```

## API Reference

### Ethereum Client (`pkg/ethereum/client.go`)

```go
// Create new client
client, err := ethereum.NewClient(rpcURL, logger)

// Get ETH balance
balance, err := client.GetBalance(ctx, address)

// Get latest block number
blockNumber, err := client.GetLatestBlockNumber(ctx)

// Subscribe to new blocks
headers, sub, err := client.SubscribeToNewBlocks(ctx)
```

### ERC20 Listener (`pkg/ethereum/erc20.go`)

```go
// Create listener
listener, err := ethereum.NewERC20Listener(client)

// Listen to transfers
events, sub, err := listener.ListenToTransfers(ctx, contractAddr, targetAddr, fromBlock)

// Get historical transfers
transfers, err := listener.GetHistoricalTransfers(ctx, contractAddr, targetAddr, fromBlock, toBlock)

// Get token balance
balance, err := listener.GetTokenBalance(ctx, contractAddr, address)
```

### Ethereum Service (`pkg/ethereum/service.go`)

```go
// Create service
service, err := ethereum.NewService(client, logger)

// Monitor specific address
err := service.MonitorAddress(ctx, targetAddr, contractAddrs, callback)

// Monitor all transfers
err := service.MonitorAllTransfers(ctx, contractAddrs, callback)

// Get historical data
transfers, err := service.GetHistoricalTransfers(ctx, contractAddr, targetAddr, fromBlock, toBlock)

// Check balances
ethBalance, err := service.GetETHBalance(ctx, address)
tokenBalance, err := service.GetTokenBalance(ctx, contractAddr, address)
```

## Data Structures

### ERC20TransferEvent

```go
type ERC20TransferEvent struct {
    From       common.Address `json:"from"`
    To         common.Address `json:"to"`
    Value      *big.Int       `json:"value"`
    TxHash     common.Hash    `json:"txHash"`
    BlockNumber uint64        `json:"blockNumber"`
    LogIndex   uint          `json:"logIndex"`
}
```

## Integration with Bridge API

The Ethereum service integrates with your existing bridge architecture:

```go
// In internal/app/app.go
func Run(cfg *config.Config) {
    // ... existing setup ...
    
    // Create Ethereum client
    ethClient, err := ethereum.NewClient(cfg.Ethereum.SepoliaRPCURL, l.Zerolog())
    if err != nil {
        l.Fatal(fmt.Errorf("app - Run - ethereum.NewClient: %w", err))
    }
    defer ethClient.Close()
    
    // Create Ethereum service
    ethService, err := ethereum.NewService(ethClient, l.Zerolog())
    if err != nil {
        l.Fatal(fmt.Errorf("app - Run - ethereum.NewService: %w", err))
    }
    defer ethService.Close()
    
    // Create Ethereum use case
    ethUseCase := ethereumuc.New(ethService, l.Zerolog())
    
    // Start monitoring (example addresses)
    targetAddresses := []common.Address{
        common.HexToAddress("0x742d35Cc6634C0532925a3b8D6Ac6E2f596C2c8d"),
    }
    
    contractAddresses := []common.Address{
        common.HexToAddress(cfg.Ethereum.Contracts["USDC"]),
        common.HexToAddress(cfg.Ethereum.Contracts["USDT"]),
    }
    
    ctx := context.Background()
    ethUseCase.StartMonitoring(ctx, targetAddresses, contractAddresses)
    
    // ... rest of your app setup ...
}
```

## Sepolia Testnet Token Addresses

Common token addresses on Sepolia testnet:

- **USDC**: `0x1c7D4B196Cb0C7B01d743Fbc6116a902379C7238`
- **USDT**: `0x7169D38820dfd117C3FA1f22a697dBA58d90BA06`
- **DAI**: `0xFF34B3d4Aee8ddCd6F9AFFFB6Fe49bD371b8a357`
- **WETH**: `0xfFf9976782d46CC05630D1f6eBAb18b2324d6B14`

## Getting Sepolia ETH

You can get free Sepolia ETH from:
- [Sepolia Faucet](https://sepoliafaucet.com/)
- [Alchemy Faucet](https://sepoliafaucet.com/)
- [Chainlink Faucet](https://faucets.chain.link/sepolia)

## Error Handling

The implementation includes comprehensive error handling:

- Connection failures to Ethereum RPC
- Invalid contract addresses
- Subscription errors
- Block parsing errors
- Balance query failures

## Performance Considerations

- **Connection Pooling**: The client reuses connections efficiently
- **Event Batching**: Events are processed in batches for better performance
- **Memory Management**: Proper cleanup of subscriptions and connections
- **Rate Limiting**: Built-in retry logic for RPC calls

## Security Notes

- **Private Keys**: Never hardcode private keys in your application
- **RPC Endpoints**: Use secure, authenticated RPC endpoints in production
- **Address Validation**: Always validate addresses before processing
- **Event Verification**: Verify events against multiple sources when critical

## Testing

```bash
# Run tests
go test ./pkg/ethereum/...

# Run with race detection
go test -race ./pkg/ethereum/...
```

## Troubleshooting

### Common Issues

1. **"connection refused"**: Check your RPC URL and network connectivity
2. **"invalid contract address"**: Verify the contract address is correct for Sepolia
3. **"subscription failed"**: Check if the contract emits Transfer events
4. **"balance query failed"**: Ensure the address and contract are valid

### Debug Mode

Enable debug logging in your config:

```yaml
LOG:
  LEVEL: debug
```

This will provide detailed logs of all Ethereum operations.

## Contributing

When adding new features:

1. Follow the existing code structure
2. Add comprehensive error handling
3. Include unit tests
4. Update documentation
5. Follow Go best practices

## License

This implementation is part of the BRIDGE-API project and follows the same license terms.
