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

	client, err := ethclient.Dial(cfg.Ethereum.SepoliaRPCURL)
	if err != nil {
		log.Fatal("Dial err:", err)
	}

	// Transaction hash bạn muốn theo dõi
	txHash := common.HexToHash("0x560b097798851995b70c50e3f859bc44779990b2683cee3698531442480ec982")

	fmt.Println("🔍 THEO DÕI TRANSACTION TRÊN SEPOLIA TESTNET")
	fmt.Println("============================================")
	fmt.Printf("📋 Transaction Hash: %s\n", txHash.Hex())
	fmt.Printf("🔗 Etherscan URL: https://sepolia.etherscan.io/tx/%s\n", txHash.Hex())
	fmt.Println()

	// Lấy thông tin transaction
	tx, isPending, err := client.TransactionByHash(context.Background(), txHash)
	if err != nil {
		log.Fatal("TransactionByHash err:", err)
	}

	if isPending {
		fmt.Println("⏳ Transaction đang pending...")
		return
	}

	fmt.Println("✅ THÔNG TIN TRANSACTION")
	fmt.Println("------------------------")

	// Lấy chainID để decode sender
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatal("NetworkID err:", err)
	}

	// Lấy sender address
	// Sử dụng LatestSignerForChainID để hỗ trợ cả legacy tx và EIP-1559 (dynamic fee) tx
	signer := types.LatestSignerForChainID(chainID)
	from, err := types.Sender(signer, tx)
	if err != nil {
		log.Fatal("Sender err:", err, chainID)
	}

	fmt.Printf("📤 From (Sender): %s\n", from.Hex())
	if tx.To() != nil {
		fmt.Printf("📥 To (Receiver): %s\n", tx.To().Hex())
	} else {
		fmt.Println("📥 To: Contract Creation")
	}

	// Format value từ Wei sang ETH
	ethValue := new(big.Float).SetInt(tx.Value())
	ethValue.Quo(ethValue, big.NewFloat(1e18))
	fmt.Printf("💰 Value: %s ETH (%s Wei)\n", ethValue.String(), tx.Value().String())

	fmt.Printf("🔢 Nonce: %d\n", tx.Nonce())
	fmt.Printf("⛽ Gas Limit: %d\n", tx.Gas())
	fmt.Printf("💸 Gas Price: %s Gwei\n", weiToGwei(tx.GasPrice()))

	if tx.Type() == types.DynamicFeeTxType {
		fmt.Printf("🚀 Max Fee Per Gas: %s Gwei\n", weiToGwei(tx.GasFeeCap()))
		fmt.Printf("💡 Max Priority Fee: %s Gwei\n", weiToGwei(tx.GasTipCap()))
	}

	// Lấy receipt để xem kết quả
	receipt, err := client.TransactionReceipt(context.Background(), txHash)
	if err != nil {
		log.Fatal("TransactionReceipt err:", err)
	}

	fmt.Println()
	fmt.Println("📊 KẾT QUẢ THỰC THI")
	fmt.Println("-------------------")

	if receipt.Status == 1 {
		fmt.Println("✅ Status: SUCCESS")
	} else {
		fmt.Println("❌ Status: FAILED")
	}

	fmt.Printf("📦 Block Number: %d\n", receipt.BlockNumber.Uint64())
	fmt.Printf("🔗 Block Hash: %s\n", receipt.BlockHash.Hex())
	fmt.Printf("⛽ Gas Used: %d (%.2f%%)\n", receipt.GasUsed, float64(receipt.GasUsed)/float64(tx.Gas())*100)

	// Tính transaction fee
	gasUsed := new(big.Int).SetUint64(receipt.GasUsed)
	txFee := new(big.Int).Mul(gasUsed, tx.GasPrice())
	txFeeETH := new(big.Float).SetInt(txFee)
	txFeeETH.Quo(txFeeETH, big.NewFloat(1e18))
	fmt.Printf("💸 Transaction Fee: %s ETH\n", txFeeETH.String())

	// Hiển thị logs (events)
	if len(receipt.Logs) > 0 {
		fmt.Println()
		fmt.Printf("📝 EVENTS/LOGS (%d events)\n", len(receipt.Logs))
		fmt.Println("-------------------------")
		for i, log := range receipt.Logs {
			fmt.Printf("\n🔸 Event #%d:\n", i+1)
			fmt.Printf("   Contract: %s\n", log.Address.Hex())
			fmt.Printf("   Topics:\n")
			for j, topic := range log.Topics {
				fmt.Printf("      [%d] %s\n", j, topic.Hex())
			}
			if len(log.Data) > 0 {
				fmt.Printf("   Data: %x\n", log.Data)
			}
			fmt.Printf("   Etherscan: https://sepolia.etherscan.io/tx/%s#eventlog\n", txHash.Hex())
		}
	}

	// Lấy thông tin block
	block, err := client.BlockByHash(context.Background(), receipt.BlockHash)
	if err == nil {
		fmt.Println()
		fmt.Println("📦 THÔNG TIN BLOCK")
		fmt.Println("-----------------")
		fmt.Printf("Block Number: %d\n", block.NumberU64())
		fmt.Printf("Block Hash: %s\n", block.Hash().Hex())
		fmt.Printf("Timestamp: %d\n", block.Time())
		fmt.Printf("Total Transactions: %d\n", len(block.Transactions()))
		fmt.Printf("Etherscan: https://sepolia.etherscan.io/block/%d\n", block.NumberU64())
	}

	// Kiểm tra nếu là token transfer (ERC-20)
	if len(receipt.Logs) > 0 {
		fmt.Println()
		fmt.Println("🪙 TOKEN TRANSFER DETECTION")
		fmt.Println("--------------------------")

		// ERC-20 Transfer event signature
		transferSig := common.HexToHash("0x560b097798851995b70c50e3f859bc44779990b2683cee3698531442480ec982")

		for _, log := range receipt.Logs {
			if len(log.Topics) > 0 && log.Topics[0] == transferSig {
				fmt.Println("✅ Đây là ERC-20 Token Transfer!")
				fmt.Printf("   Token Contract: %s\n", log.Address.Hex())

				if len(log.Topics) >= 3 {
					fromAddr := common.BytesToAddress(log.Topics[1].Bytes())
					toAddr := common.BytesToAddress(log.Topics[2].Bytes())

					fmt.Printf("   From: %s\n", fromAddr.Hex())
					fmt.Printf("   To: %s\n", toAddr.Hex())

					if len(log.Data) > 0 {
						value := new(big.Int).SetBytes(log.Data)
						fmt.Printf("   Value (Wei): %s\n", value.String())

						// Giả sử token có 18 decimals (phổ biến nhất)
						tokenValue := new(big.Float).SetInt(value)
						tokenValue.Quo(tokenValue, big.NewFloat(1e18))
						fmt.Printf("   Value (Tokens): %s\n", tokenValue.String())
					}
				}

				fmt.Printf("   View on Etherscan: https://sepolia.etherscan.io/token/%s\n", log.Address.Hex())
			}
		}
	}

	fmt.Println()
	fmt.Println("============================================")
	fmt.Println("✅ Hoàn thành theo dõi transaction!")
}

// weiToGwei converts Wei to Gwei
func weiToGwei(wei *big.Int) string {
	gwei := new(big.Float).SetInt(wei)
	gwei.Quo(gwei, big.NewFloat(1e9))
	return gwei.String()
}
