package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	// TODO: Uncomment when you have generated the contract binding
	// Replace "ceth" with your actual contract binding package path
	// "github.com/OpenBankingVN/BRIDGE-API/pkg/contracts/ceth"
)

// SupplyETH supplies ETH to the cETH contract by calling the Mint() function
//
// To use contract bindings (recommended):
//  1. Generate contract bindings using abigen:
//     abigen --abi ceth.abi --pkg ceth --type Ceth --out pkg/contracts/ceth/ceth.go
//  2. Uncomment the contract binding code below and comment out the manual transaction code
//
// Parameters:
//   - client: Ethereum client connection
//   - privateKey: Private key for signing the transaction
//   - cETH: cETH contract address
//   - amountWei: Amount of ETH to supply (in Wei)
//
// Returns:
//   - Transaction hash on success
//   - Error if transaction fails
func SupplyETH(client *ethclient.Client, privateKey *ecdsa.PrivateKey, cETH common.Address, amountWei *big.Int) (common.Hash, error) {
	ctx := context.Background()

	// Get chain ID from network (supports local fork which may use mainnet chainID)
	chainID, err := client.NetworkID(ctx)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to get network ID: %w", err)
	}

	// Create transactor with chain ID
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to create transactor: %w", err)
	}

	// Set the amount of ETH to send with the transaction
	auth.Value = amountWei

	// Get the suggested gas tip cap
	gasTipCap, err := client.SuggestGasTipCap(ctx)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to suggest gas tip cap: %w", err)
	}

	// Get base fee for EIP-1559
	header, err := client.HeaderByNumber(ctx, nil)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to get latest header: %w", err)
	}
	baseFee := header.BaseFee
	if baseFee == nil {
		baseFee = big.NewInt(0)
	}

	// Calculate max fee per gas (baseFee * 2 + tip)
	maxFeePerGas := new(big.Int).Mul(baseFee, big.NewInt(2))
	maxFeePerGas.Add(maxFeePerGas, gasTipCap)

	// Set gas options
	auth.GasTipCap = gasTipCap
	auth.GasFeeCap = maxFeePerGas

	// Get nonce
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return common.Hash{}, fmt.Errorf("failed to assert public key type")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to get nonce: %w", err)
	}
	auth.Nonce = big.NewInt(int64(nonce))

	// OPTION 1: Use contract bindings (recommended - uncomment when you have generated bindings)
	/*
		// Import the contract binding package:
		// import "github.com/OpenBankingVN/BRIDGE-API/pkg/contracts/ceth"

		contract, err := ceth.NewCeth(cETH, client)
		if err != nil {
			return common.Hash{}, fmt.Errorf("failed to instantiate contract: %w", err)
		}

		// Call the Mint function (payable function that accepts ETH)
		tx, err := contract.Mint(auth)
		if err != nil {
			return common.Hash{}, fmt.Errorf("failed to call Mint: %w", err)
		}

		return tx.Hash(), nil
	*/

	// OPTION 2: Manual transaction (requires function selector)
	// Calculate function selector for mint() - first 4 bytes of keccak256("mint()")
	// If your function has a different signature, update this
	mintSelector := crypto.Keccak256([]byte("mint()"))[:4]

	// Estimate gas
	gasLimit, err := client.EstimateGas(ctx, ethereum.CallMsg{
		From:  fromAddress,
		To:    &cETH,
		Value: amountWei,
		Data:  mintSelector,
	})
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to estimate gas: %w", err)
	}

	// Add 20% buffer to gas estimate
	bufferedGas := gasLimit + gasLimit/5

	// Create transaction with function selector
	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     nonce,
		To:        &cETH,
		Value:     amountWei,
		Gas:       bufferedGas,
		GasTipCap: gasTipCap,
		GasFeeCap: maxFeePerGas,
		Data:      mintSelector,
	})

	// Sign the transaction
	signer := types.NewLondonSigner(chainID)
	signedTx, err := types.SignTx(tx, signer, privateKey)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Send the transaction
	err = client.SendTransaction(ctx, signedTx)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return signedTx.Hash(), nil
}

func main() {
	// cfg, err := config.NewConfig()
	// if err != nil {
	// 	log.Fatalf("Failed to load config: %v", err)
	// }

	// 1. Connect to local mainnet fork RPC
	// Default local fork URL (change if your local node uses a different port)
	localForkURL := "http://localhost:8545"

	// Option: Use URL from config if available, otherwise use default local URL
	// if cfg.Ethereum.SepoliaRPCURL != "" {
	// 	localForkURL = cfg.Ethereum.SepoliaRPCURL
	// }

	client, err := ethclient.Dial(localForkURL)
	if err != nil {
		log.Fatalf("Failed to connect to Ethereum client: %v", err)
	}
	defer client.Close()

	fmt.Printf("✅ Connected to Ethereum client (Local Mainnet Fork: %s)\n", localForkURL)

	// 2. Load private key
	privateKey, err := crypto.HexToECDSA("ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80")
	if err != nil {
		log.Fatalf("Failed to parse private key: %v", err)
	}

	// 3. Get the sender address
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("Failed to assert public key type")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	fmt.Printf("📤 From Address: %s\n", fromAddress.Hex())

	// 4. Set cETH contract address
	// Compound cETH contract on mainnet: 0x4Ddc2D193948926D02f9B1fE9e1daa0718270ED5
	// If using a local fork, this should be the same as mainnet address (if forked from mainnet)
	// TODO: Replace with your actual cETH contract address if different
	cETHAddress := common.HexToAddress("0x4Ddc2D193948926D02f9B1fE9e1daa0718270ED5") // Compound cETH on mainnet

	// Validate that contract address is different from sender address
	if cETHAddress == fromAddress {
		log.Fatalf("❌ Error: cETH contract address cannot be the same as sender address!\n"+
			"   Sender: %s\n"+
			"   Contract: %s\n"+
			"   Please set a valid cETH contract address.", fromAddress.Hex(), cETHAddress.Hex())
	}

	if cETHAddress == (common.Address{}) {
		log.Fatal("Please set the cETH contract address")
	}

	fmt.Printf("📋 cETH Contract Address: %s\n", cETHAddress.Hex())

	// 5. Set amount to supply (in Wei) - 10 ETH = 10000000000000000000 Wei
	// Use big.Int with string to avoid int64 overflow
	amountWei, _ := new(big.Int).SetString("100000000000000000000", 10) // 100 ETH

	// Convert to ETH for display
	amountETH := new(big.Float).SetInt(amountWei)
	amountETH.Quo(amountETH, big.NewFloat(1e18))
	fmt.Printf("💰 Amount to supply: %s ETH (%s Wei)\n", amountETH.String(), amountWei.String())

	// 6. Check balance before
	ctx := context.Background()
	balanceBefore, err := client.BalanceAt(ctx, fromAddress, nil)
	if err != nil {
		log.Printf("Warning: Failed to get balance: %v", err)
	} else {
		balanceETH := new(big.Float).SetInt(balanceBefore)
		balanceETH.Quo(balanceETH, big.NewFloat(1e18))
		fmt.Printf("💳 Balance before: %s ETH\n", balanceETH.String())
	}

	// 7. Call SupplyETH function
	fmt.Println("\n🚀 Supplying ETH to cETH contract...")
	txHash, err := SupplyETH(client, privateKey, cETHAddress, amountWei)
	if err != nil {
		log.Fatalf("❌ Failed to supply ETH: %v", err)
	}

	fmt.Printf("✅ Transaction sent! TX Hash: %s\n", txHash.Hex())
	fmt.Printf("🔗 Transaction Hash: %s (Local fork - no explorer)\n", txHash.Hex())

	// 8. Wait for transaction receipt
	fmt.Println("\n⏳ Waiting for transaction confirmation...")
	receiptCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	var receipt *types.Receipt
	for {
		receipt, err = client.TransactionReceipt(receiptCtx, txHash)
		if err != nil {
			if err == ethereum.NotFound {
				time.Sleep(2 * time.Second)
				fmt.Print(".")
				continue
			}
			log.Fatalf("Failed to fetch receipt: %v", err)
		}
		break
	}

	// 9. Display receipt information
	status := "❌ FAILED"
	if receipt.Status == types.ReceiptStatusSuccessful {
		status = "✅ SUCCESS"
	}
	fmt.Printf("\n%s\n", status)
	fmt.Printf("📦 Block Number: %d\n", receipt.BlockNumber.Uint64())
	fmt.Printf("⛽ Gas Used: %d\n", receipt.GasUsed)

	// 10. Check balance after
	balanceAfter, err := client.BalanceAt(ctx, fromAddress, receipt.BlockNumber)
	if err != nil {
		log.Printf("Warning: Failed to get balance after: %v", err)
	} else {
		balanceETH := new(big.Float).SetInt(balanceAfter)
		balanceETH.Quo(balanceETH, big.NewFloat(1e18))
		fmt.Printf("💳 Balance after: %s ETH\n", balanceETH.String())
	}
}
