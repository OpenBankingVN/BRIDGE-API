package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/OpenBankingVN/BRIDGE-API/config"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	// 1. Kết nối tới Sepolia RPC
	client, err := ethclient.Dial(cfg.Ethereum.SepoliaRPCURL)
	if err != nil {
		log.Fatalf("failed to connect to Sepolia RPC: %v", err)
	}

	// 2. Import private key của ví MetaMask (đang có 0.2 ETH testnet)
	privateKey, err := crypto.HexToECDSA(cfg.Ethereum.PrivateKey)
	if err != nil {
		log.Fatalf("failed to parse private key: %v", err)
	}

	// 3. Lấy public key → địa chỉ ví gửi
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	// 4. Lấy nonce của account (số tx gửi đi)
	ctx := context.Background()
	nonce, err := client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		log.Fatalf("failed to fetch pending nonce: %v", err)
	}

	// 5. Lấy thông tin phí EIP-1559 (dynamic fee)
	gasTipCap, err := client.SuggestGasTipCap(ctx)
	if err != nil {
		log.Fatalf("failed to suggest gas tip cap: %v", err)
	}
	header, err := client.HeaderByNumber(ctx, nil)
	if err != nil {
		log.Fatalf("failed to get latest header: %v", err)
	}
	baseFee := header.BaseFee
	// maxFeePerGas = baseFee*2 + tip (buffer to avoid underpricing during spikes)
	maxFeePerGas := new(big.Int).Mul(baseFee, big.NewInt(2))
	maxFeePerGas.Add(maxFeePerGas, gasTipCap)

	// 6. Thông tin giao dịch
	toEnv := cfg.Ethereum.ToAddress
	if toEnv == "" {
		log.Fatal("TO_ADDRESS is not set. Export your EOA recipient address (hex)")
	}
	toAddress := common.HexToAddress(toEnv) // địa chỉ nhận ETH
	value := big.NewInt(10000000000000000)  // 0.0001 ETH (đơn vị Wei)
	var data []byte                         // tx ETH thường không cần data
	// gasLimit := uint64(21000)            // Gas chuẩn cho transfer ETH

	// Cảnh báo nếu đích là contract (có thể không nhận ETH trực tiếp)
	code, err := client.CodeAt(ctx, toAddress, nil)
	if err != nil {
		log.Printf("warning: failed to check if recipient is contract: %v", err)
	} else if len(code) > 0 {
		log.Fatal("recipient is a contract. Please set TO_ADDRESS to an EOA (normal wallet)")
	}

	// In số dư trước khi gửi (để so sánh)
	fromBalBefore, err := client.BalanceAt(ctx, fromAddress, nil)
	if err != nil {
		log.Printf("warning: failed to get sender balance before: %v", err)
	}
	toBalBefore, err := client.BalanceAt(ctx, toAddress, nil)
	if err != nil {
		log.Printf("warning: failed to get recipient balance before: %v", err)
	}
	fmt.Printf("From balance BEFORE: %s Wei\n", fromBalBefore.String())
	fmt.Printf("To   balance BEFORE: %s Wei\n", toBalBefore.String())

	// 7. Ước lượng gas động theo đích đến (EOA/contract) để tránh out of gas
	estimatedGas, err := client.EstimateGas(ctx, ethereum.CallMsg{
		From:  fromAddress,
		To:    &toAddress,
		Value: value,
		Data:  data,
	})
	if err != nil {
		log.Fatalf("failed to estimate gas: %v", err)
	}

	// Tạo transaction kiểu EIP-1559 (DynamicFeeTx)
	// Thêm buffer 20% cho gas estimate để tránh thiếu gas do dao động
	bufferedGas := estimatedGas + estimatedGas/5
	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   big.NewInt(11155111),
		Nonce:     nonce,
		To:        &toAddress,
		Value:     value,
		Gas:       bufferedGas,
		GasTipCap: gasTipCap,
		GasFeeCap: maxFeePerGas,
		Data:      data,
	})

	// 8. Ký transaction
	chainID := big.NewInt(11155111) // Sepolia chainID
	signer := types.NewLondonSigner(chainID)
	signedTx, err := types.SignTx(tx, signer, privateKey)
	if err != nil {
		log.Fatalf("failed to sign transaction: %v", err)
	}

	// 9. Gửi transaction
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatalf("failed to send transaction: %v", err)
	}

	fmt.Printf("Transaction sent! TX Hash: %s\n", signedTx.Hash().Hex())
	fmt.Printf("View on Etherscan: https://sepolia.etherscan.io/tx/%s\n", signedTx.Hash().Hex())

	// 10. Chờ receipt để biết contract execution có thành công không
	receiptCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	var receipt *types.Receipt
	for {
		receipt, err = client.TransactionReceipt(receiptCtx, signedTx.Hash())
		if err != nil {
			if err == ethereum.NotFound {
				time.Sleep(1 * time.Second)
				continue
			}
			log.Fatalf("failed to fetch receipt: %v", err)
		}
		// In thông tin receipt
		status := "FAILED"
		if receipt.Status == types.ReceiptStatusSuccessful {
			status = "SUCCESS"
		}
		fmt.Printf("Receipt status: %s | Block: %d | GasUsed: %d\n", status, receipt.BlockNumber.Uint64(), receipt.GasUsed)
		break
	}

	// In số dư sau tại block của receipt
	fromBalAfter, err := client.BalanceAt(ctx, fromAddress, receipt.BlockNumber)
	if err != nil {
		log.Printf("warning: failed to get sender balance after: %v", err)
	}
	toBalAfter, err := client.BalanceAt(ctx, toAddress, receipt.BlockNumber)
	if err != nil {
		log.Printf("warning: failed to get recipient balance after: %v", err)
	}
	fmt.Printf("From balance AFTER: %s Wei\n", fromBalAfter.String())
	fmt.Printf("To   balance AFTER: %s Wei\n", toBalAfter.String())

	// Chênh lệch
	if fromBalBefore != nil && fromBalAfter != nil {
		fromDelta := new(big.Int).Sub(fromBalAfter, fromBalBefore)
		fmt.Printf("From delta: %s Wei\n", fromDelta.String())
	}
	if toBalBefore != nil && toBalAfter != nil {
		toDelta := new(big.Int).Sub(toBalAfter, toBalBefore)
		fmt.Printf("To   delta: %s Wei\n", toDelta.String())
	}
}
