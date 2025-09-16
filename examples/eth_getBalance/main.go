package main

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	// 1. Kết nối đến Sepolia RPC (dùng Infura / Alchemy / public RPC)
	client, err := ethclient.Dial("https://eth-sepolia.g.alchemy.com/v2/xxxx")
	//

	if err != nil {
		log.Fatal("Failed to connect to Ethereum client:", err)
	}

	// 2. Địa chỉ ví MetaMask của bạn
	address := common.HexToAddress("0xADDRESS")

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
}
