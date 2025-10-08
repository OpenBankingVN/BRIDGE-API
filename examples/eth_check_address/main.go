package main

import (
	"context"
	"fmt"
	"log"
	"regexp"

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
	if err != nil {
		log.Fatal(err)
	}
	//
	re := regexp.MustCompile("^0x[0-9a-fA-F]{40}$")

	fmt.Printf("is valid: %v\n", re.MatchString("0x323b5d4c32345ced77393b3530b1eed0f346429d")) // is valid: true
	fmt.Printf("is valid: %v\n", re.MatchString("0xZYXb5d4c32345ced77393b3530b1eed0f346429d")) // is valid: false

	// 0x Protocol Token (ZRX) smart contract address
	addressStr := common.HexToAddress("0xe41d2489571d322189246dafa5ebde1f4699f498")
	address := common.HexToAddress(addressStr.Hex())
	bytecode, err := client.CodeAt(context.Background(), address, nil) // nil is latest block
	if err != nil {
		log.Fatal(err)
	}

	isContract := len(bytecode) > 0

	fmt.Printf("Address %s is contract: %v\n", address.Hex(), isContract) // is contract: true

	// a random user account address
	addressSt1r := "0x661Fa55d705bB7894c3a0D17408885c78F460f06"
	address1 := common.HexToAddress(addressSt1r)
	bytecode, err = client.CodeAt(context.Background(), address1, nil) // nil is latest block
	if err != nil {
		log.Fatal(err)
	}

	isContract = len(bytecode) > 0

	fmt.Printf("Address %s is contract: %v\n", address1.Hex(), isContract) // is contract: false
}
