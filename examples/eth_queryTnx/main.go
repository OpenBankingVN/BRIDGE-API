package main

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/OpenBankingVN/BRIDGE-API/config"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
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
		log.Fatal("Dial err:", err)
	}

	blockNumber := big.NewInt(9376134)
	block, err := client.BlockByNumber(context.Background(), blockNumber)
	if err != nil {
		log.Fatal("BlockByNumber err:", err)
	}

	fmt.Println("blockNumber:", blockNumber.Uint64())
	fmt.Println("count:", len(block.Transactions()))
	fmt.Println("++++++++++++++++++++++++++++++++++++++++++++++++")
	for _, tx := range block.Transactions() {
		fmt.Println("tx.Hash().Hex():", tx.Hash().Hex())               // 0x5d49fcaa394c97ec8a9c3e7bd9e8388d420fb050a52083ca52ff24b3b65bc9c2
		fmt.Println("tx.Value().String():", tx.Value().String())       // 10000000000000000
		fmt.Println("tx.Gas():", tx.Gas())                             // 105000
		fmt.Println("tx.GasPrice().Uint64():", tx.GasPrice().Uint64()) // 102000000000
		fmt.Println("tx.Nonce():", tx.Nonce())                         // 110644
		fmt.Println("tx.Data():", tx.Data())                           // []
		fmt.Println("tx.To().Hex():", tx.To().Hex())                   // 0x55fE59D8Ad77035154dDd0AD0388D09Dd4047A8e

		chainID, err := client.NetworkID(context.Background())
		if err != nil {
			log.Fatal("NetworkID err:", err)
		}
		fmt.Println("chainID:", chainID.Uint64())
		if from, err := types.Sender(types.NewEIP155Signer(chainID), tx); err == nil {
			fmt.Println("from.Hex():", from.Hex()) // 0x0fD081e3Bb178dc45c0cb23202069ddA57064258
		}

		receipt, err := client.TransactionReceipt(context.Background(), tx.Hash())
		if err != nil {
			log.Fatal("TransactionReceipt err:", err)
		}

		fmt.Println("receipt.Status:", receipt.Status) // 1
		fmt.Println("receipt.Logs:", receipt.Logs)
		fmt.Println("--------------------------------")
	}
	fmt.Println("++++++++++++++++++++++++++++++++++++++++++++++++")

	// Use the block's hash directly from the block we already have
	count := uint(len(block.Transactions()))
	fmt.Println("Transaction count from block:", count)

	for idx := uint(0); idx < count; idx++ {
		tx, err := client.TransactionInBlock(context.Background(), block.Hash(), idx)
		if err != nil {
			log.Fatal("TransactionInBlock err:", err)
		}

		fmt.Println("tx.Hash().Hex():", tx.Hash().Hex())
	}
	fmt.Println("++++++++++++++++++++++++++++++++++++++++++++++++")

	txHash := common.HexToHash("0xabfc3db93edbb4e37e11f9166cfd7916e5a4073a76d7b2f60c2f964e31073e47")
	tx, isPending, err := client.TransactionByHash(context.Background(), txHash)
	if err != nil {
		log.Fatal("TransactionByHash err:", err)
	}

	fmt.Println(tx.Hash().Hex()) //
	fmt.Println(isPending)       // false
}
