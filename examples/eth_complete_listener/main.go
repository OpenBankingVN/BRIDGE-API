package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/OpenBankingVN/BRIDGE-API/config"
	"github.com/OpenBankingVN/BRIDGE-API/pkg/ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
)

func main() {
	fmt.Println("🚀 Complete Ethereum Transfer Listener")
	fmt.Println("=====================================")
	fmt.Println("✅ Detects both ETH and ERC-20 token transfers!")
	fmt.Println()

	// Load configuration
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create logger
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	// Create Ethereum client
	ethClient, err := ethereum.NewClient(cfg.Ethereum.SepoliaRPCURL, logger)
	if err != nil {
		log.Fatalf("Failed to create Ethereum client: %v", err)
	}
	defer ethClient.Close()

	// Create services
	pollingListener, err := ethereum.NewPollingListener(ethClient, logger)
	if err != nil {
		log.Fatalf("Failed to create polling listener: %v", err)
	}

	ethTransferListener, err := ethereum.NewETHTransferListener(ethClient, logger)
	if err != nil {
		log.Fatalf("Failed to create ETH transfer listener: %v", err)
	}

	// Your target address
	targetAddressStr := "0xA7cF451F98b565dfd15274A90B502a74F59008f0"
	targetAddr := common.HexToAddress(targetAddressStr)

	fmt.Printf("🎯 Target Address: %s\n", targetAddr.Hex())
	fmt.Printf("🔗 Chain ID: %s\n", ethClient.GetChainID().String())

	// Check current balances
	fmt.Println("\n💰 Current balances:")
	checkBalances(ethClient, targetAddr)

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\n🛑 Shutting down...")
		cancel()
	}()

	fmt.Println("\n🎧 Starting complete monitoring...")
	fmt.Println("   Monitoring ETH transfers AND ERC-20 token transfers")
	fmt.Println("   Send ETH or tokens to your address to see events!")
	fmt.Println("   Press Ctrl+C to stop")

	// Start ETH transfer monitoring
	go func() {
		ethCallback := func(event ethereum.ETHTransferEvent) {
			fmt.Printf("\n🔥 ETH TRANSFER DETECTED! 🔥\n")
			fmt.Printf("  📤 From: %s\n", event.From.Hex())
			fmt.Printf("  📥 To: %s\n", event.To.Hex())
			fmt.Printf("  💎 Value: %s ETH\n", formatETH(event.Value))
			fmt.Printf("  🔗 Tx Hash: %s\n", event.TxHash.Hex())
			fmt.Printf("  📦 Block: %d\n", event.BlockNumber)
			fmt.Printf("  ⏰ Time: %s\n", time.Now().Format(time.RFC3339))
			fmt.Printf("  ⛽ Gas Used: %d\n", event.GasUsed)

			// Check direction
			if event.To == targetAddr {
				fmt.Printf("  ✅ INCOMING ETH TRANSFER TO YOUR ADDRESS!\n")
			} else if event.From == targetAddr {
				fmt.Printf("  📤 OUTGOING ETH TRANSFER FROM YOUR ADDRESS!\n")
			}

			fmt.Printf("================================\n")
		}

		ethTransferListener.StartPollingETH(ctx, []common.Address{targetAddr}, 2*time.Second, ethCallback)
	}()

	// Start ERC-20 token monitoring
	go func() {
		contractAddresses := []common.Address{
			common.HexToAddress("0x1c7D4B196Cb0C7B01d743Fbc6116a902379C7238"), // USDC
			common.HexToAddress("0x7169D38820dfd117C3FA1f22a697dBA58d90BA06"), // USDT
			common.HexToAddress("0xFF34B3d4Aee8ddCd6F9AFFFB6Fe49bD371b8a357"), // DAI
		}

		tokenCallback := func(event ethereum.ERC20TransferEvent) {
			fmt.Printf("\n🔥 TOKEN TRANSFER DETECTED! 🔥\n")
			fmt.Printf("  📤 From: %s\n", event.From.Hex())
			fmt.Printf("  📥 To: %s\n", event.To.Hex())
			fmt.Printf("  💎 Value: %s tokens\n", event.Value.String())
			fmt.Printf("  🔗 Tx Hash: %s\n", event.TxHash.Hex())
			fmt.Printf("  📦 Block: %d\n", event.BlockNumber)
			fmt.Printf("  ⏰ Time: %s\n", time.Now().Format(time.RFC3339))

			// Check direction
			if event.To == targetAddr {
				fmt.Printf("  ✅ INCOMING TOKEN TRANSFER TO YOUR ADDRESS!\n")
			} else if event.From == targetAddr {
				fmt.Printf("  📤 OUTGOING TOKEN TRANSFER FROM YOUR ADDRESS!\n")
			}

			fmt.Printf("================================\n")
		}

		pollingListener.StartPolling(ctx, contractAddresses, &targetAddr, 2*time.Second, tokenCallback)
	}()

	// Keep the program running
	<-ctx.Done()
	fmt.Println("👋 Monitoring stopped. Goodbye!")
}

func checkBalances(client *ethereum.Client, address common.Address) {
	ctx := context.Background()

	// Check ETH balance
	ethBalance, err := client.GetBalance(ctx, address)
	if err != nil {
		fmt.Printf("❌ Failed to get ETH balance: %v\n", err)
	} else {
		fmt.Printf("   ETH: %s ETH\n", formatETH(ethBalance))
	}

	// Create a simple service for token balance checks
	logger := zerolog.New(os.Stdout)
	ethService, err := ethereum.NewService(client, logger)
	if err != nil {
		fmt.Printf("❌ Failed to create service: %v\n", err)
		return
	}
	defer ethService.Close()

	// Check USDC balance (6 decimals)
	usdcAddr := common.HexToAddress("0x1c7D4B196Cb0C7B01d743Fbc6116a902379C7238")
	usdcBalance, err := ethService.GetTokenBalance(ctx, usdcAddr, address)
	if err != nil {
		fmt.Printf("   USDC: Failed to get balance (%v)\n", err)
	} else {
		usdcFloat := new(big.Float).SetInt(usdcBalance)
		usdcFloat.Quo(usdcFloat, big.NewFloat(1e6)) // USDC has 6 decimals
		fmt.Printf("   USDC: %s USDC\n", usdcFloat.String())
	}

	// Check USDT balance (6 decimals)
	usdtAddr := common.HexToAddress("0x7169D38820dfd117C3FA1f22a697dBA58d90BA06")
	usdtBalance, err := ethService.GetTokenBalance(ctx, usdtAddr, address)
	if err != nil {
		fmt.Printf("   USDT: Failed to get balance (%v)\n", err)
	} else {
		usdtFloat := new(big.Float).SetInt(usdtBalance)
		usdtFloat.Quo(usdtFloat, big.NewFloat(1e6)) // USDT has 6 decimals
		fmt.Printf("   USDT: %s USDT\n", usdtFloat.String())
	}
}

// formatETH formats Wei to ETH
func formatETH(wei *big.Int) string {
	ethFloat := new(big.Float).SetInt(wei)
	ethFloat.Quo(ethFloat, big.NewFloat(1e18))
	return ethFloat.String()
}
