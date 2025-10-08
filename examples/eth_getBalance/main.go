package main

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/OpenBankingVN/BRIDGE-API/config"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create logger

	// 1. Kết nối đến Sepolia RPC (dùng Infura / Alchemy / public RPC)
	client, err := ethclient.Dial(cfg.Ethereum.SepoliaRPCURL)
	//

	if err != nil {
		log.Fatal("Failed to connect to Ethereum client:", err)
	}

	// 2. Địa chỉ ví MetaMask của bạn
	address := common.HexToAddress("0x661Fa55d705bB7894c3a0D17408885c78F460f06")

	// 3. Lấy balance
	balance, err := client.BalanceAt(context.Background(), address, nil)
	if err != nil {
		log.Fatal("Failed to get balance:", err)
	}

	// 4. Convert từ Wei → ETH
	fBalance := new(big.Float)
	fBalance.SetString(balance.String())
	ethValue := new(big.Float).Quo(fBalance, big.NewFloat(1e18))

	fmt.Printf("Address: %s\n", address.Hex())
	fmt.Printf("Balance: %f ETH\n", ethValue)

	pendingBalance, err := client.PendingBalanceAt(context.Background(), address)
	if err != nil {
		log.Fatal("Failed to get balance:", err)
	}

	// 4. Convert từ Wei → ETH
	pfBalance := new(big.Float)
	pfBalance.SetString(pendingBalance.String())
	pethValue := new(big.Float).Quo(pfBalance, big.NewFloat(1e18))
	fmt.Printf("Pending Balance: %f ETH\n", pethValue)
}
