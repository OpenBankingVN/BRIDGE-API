package main

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/OpenBankingVN/BRIDGE-API/config"
	"github.com/OpenBankingVN/BRIDGE-API/pkg/ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
)

func main() {
	fmt.Println("🔍 Find Your Transfer")
	fmt.Println("====================")

	// Load configuration
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create logger
	logger := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()

	// Create Ethereum client
	ethClient, err := ethereum.NewClient(cfg.Ethereum.SepoliaRPCURL, logger)
	if err != nil {
		log.Fatalf("Failed to create Ethereum client: %v", err)
	}
	defer ethClient.Close()

	// Create Ethereum service
	ethService, err := ethereum.NewService(ethClient, logger)
	if err != nil {
		log.Fatalf("Failed to create Ethereum service: %v", err)
	}
	defer ethService.Close()

	ctx := context.Background()
	targetAddr := common.HexToAddress(cfg.Ethereum.Address)
	if err != nil {
		log.Fatalf("Failed to convert address: %v", err)
	}

	fmt.Printf("🎯 Target Address: %s\n", targetAddr.Hex())

	// Get latest block
	latestBlock, err := ethClient.GetLatestBlockNumber(ctx)
	if err != nil {
		log.Fatalf("Failed to get latest block: %v", err)
	}
	fmt.Printf("📦 Latest block: %s\n", latestBlock.String())

	// Check more token contracts (common Sepolia tokens)
	allTokens := map[string]string{
		// Major tokens
		"USDC": "0x1c7D4B196Cb0C7B01d743Fbc6116a902379C7238",
		"USDT": "0x7169D38820dfd117C3FA1f22a697dBA58d90BA06",
		"DAI":  "0xFF34B3d4Aee8ddCd6F9AFFFB6Fe49bD371b8a357",
		"WETH": "0xfFf9976782d46CC05630D1f6eBAb18b2324d6B14",

		// Other common tokens
		"LINK": "0x779877A7B0D9E8603169DdbD7836e478b4624789",
		"UNI":  "0x1f9840a85d5aF5bf1D1762F925BDADdC4201F984",
		"WBTC": "0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599",
	}

	fmt.Println("\n💰 Current token balances:")
	for tokenName, contractAddr := range allTokens {
		contractAddress := common.HexToAddress(contractAddr)
		balance, err := ethService.GetTokenBalance(ctx, contractAddress, targetAddr)
		if err != nil {
			fmt.Printf("   %s: Failed (%v)\n", tokenName, err)
		} else {
			balanceFloat := new(big.Float).SetInt(balance)
			// Most tokens use 18 decimals, but USDC/USDT use 6
			if tokenName == "USDC" || tokenName == "USDT" {
				balanceFloat.Quo(balanceFloat, big.NewFloat(1e6))
			} else {
				balanceFloat.Quo(balanceFloat, big.NewFloat(1e18))
			}

			if balance.Cmp(big.NewInt(0)) > 0 {
				fmt.Printf("   %s: %s %s ✅\n", tokenName, balanceFloat.String(), tokenName)
			} else {
				fmt.Printf("   %s: %s %s\n", tokenName, balanceFloat.String(), tokenName)
			}
		}
	}

	// Check recent transfers for each token (last 5 blocks only due to RPC limits)
	fmt.Println("\n🔍 Checking recent transfers (last 5 blocks)...")
	fromBlock := new(big.Int).Sub(latestBlock, big.NewInt(5))

	foundTransfers := false
	for tokenName, contractAddr := range allTokens {
		contractAddress := common.HexToAddress(contractAddr)
		transfers, err := ethService.GetHistoricalTransfers(ctx, contractAddress, &targetAddr, fromBlock, latestBlock)
		if err != nil {
			fmt.Printf("   %s: Failed (%v)\n", tokenName, err)
			continue
		}

		if len(transfers) > 0 {
			foundTransfers = true
			fmt.Printf("   %s: %d transfers found!\n", tokenName, len(transfers))

			for i, transfer := range transfers {
				valueFloat := new(big.Float).SetInt(transfer.Value)
				if tokenName == "USDC" || tokenName == "USDT" {
					valueFloat.Quo(valueFloat, big.NewFloat(1e6))
				} else {
					valueFloat.Quo(valueFloat, big.NewFloat(1e18))
				}

				fmt.Printf("     %d. From: %s\n", i+1, transfer.From.Hex())
				fmt.Printf("        To: %s\n", transfer.To.Hex())
				fmt.Printf("        Value: %s (%s %s)\n", transfer.Value.String(), valueFloat.String(), tokenName)
				fmt.Printf("        Block: %d\n", transfer.BlockNumber)
				fmt.Printf("        Tx: %s\n", transfer.TxHash.Hex())
				fmt.Printf("        Etherscan: https://sepolia.etherscan.io/tx/%s\n", transfer.TxHash.Hex())
				fmt.Println()
			}
		} else {
			fmt.Printf("   %s: No transfers found\n", tokenName)
		}
	}

	if !foundTransfers {
		fmt.Println("\n❌ No transfers found to your address in the last 5 blocks.")
		fmt.Println("\n💡 Possible reasons:")
		fmt.Println("   1. Your transfer was more than 5 blocks ago")
		fmt.Println("   2. You sent a different token not in our list")
		fmt.Println("   3. The transaction failed or is still pending")
		fmt.Println("   4. You sent to a different address")

		fmt.Println("\n🔧 Troubleshooting steps:")
		fmt.Println("   1. Check your wallet transaction history")
		fmt.Println("   2. Look for the transaction hash")
		fmt.Println("   3. Check the transaction on Etherscan")
		fmt.Println("   4. Verify the recipient address")

		fmt.Println("\n🌐 Check your transaction on Etherscan:")
		fmt.Printf("   https://sepolia.etherscan.io/address/%s\n", targetAddr.Hex())
	} else {
		fmt.Println("\n✅ Found transfers to your address!")
		fmt.Println("   The listener should have caught these events.")
	}
}
