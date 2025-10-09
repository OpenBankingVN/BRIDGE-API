// https://goethereumbook.org/block-query/
package main

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/OpenBankingVN/BRIDGE-API/config"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	client, err := ethclient.Dial(cfg.Ethereum.SepoliaRPCURL)
	if err != nil {
		log.Fatal(err)
	}

	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("headerNumber:", header.Number.String()) // 5671744

	blockNumber := big.NewInt(5671744)
	block, err := client.BlockByNumber(context.Background(), blockNumber)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("blockNumber:", block.Number().Uint64())    // 5671744
	fmt.Println("time:", block.Time())                      // 1527211625
	fmt.Println("difficulty:", block.Difficulty().Uint64()) // 3217000136609065
	fmt.Println("hash:", block.Hash().Hex())                // 0x9e8751ebb5069389b855bba72d94902cc385042661498a415979b7b6ee9ba4b9
	fmt.Println("transactions:", len(block.Transactions())) // 144

	count, err := client.TransactionCount(context.Background(), block.Hash())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(
		"count:", count) // 144
}
