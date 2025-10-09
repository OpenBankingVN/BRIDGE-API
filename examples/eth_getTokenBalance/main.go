// Mục đích: Đọc số dư token ERC20 (ví dụ: BANKVN) từ blockchain Ethereum (Sepolia)
package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Token Service - communicate w blockchain
type TokenService struct {
	client    *ethclient.Client
	tokenAddr common.Address
	parsedABI abi.ABI
}

func NewTokenService(rpcURL, tokenAddr string) (*TokenService, error) {
	// Connect to RPC node (Alchemhy / Infura / Ankr ...)
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RPC: %v", err)
	}

	// Chuẩn API ERC20 cơ bản (balanceOf + decimals)
	// ABI (Application Binary Interface) là bản mô tả “ngôn ngữ chung”
	// giữa ứng dụng (ở đây là Go) và smart contract (Solidity).
	const erc20ABI = `[{
		"constant": true,
		"inputs": [{"name": "owner", "type": "address"}],
		"name": "balanceOf",
		"outputs": [{"name": "", "type": "uint256"}],
		"type": "function"
	}, {
		"constant": true,
		"inputs": [],
		"name": "decimals",
		"outputs": [{"name": "", "type": "uint8"}],
		"type": "function"
	}]`

	parsedABI, err := abi.JSON(strings.NewReader(erc20ABI)) // abi.JSON() → Parse chuỗi JSON thành đối tượng Go có thể dùng được.
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI: %v", err)
	}

	return &TokenService{
		client:    client,
		tokenAddr: common.HexToAddress(tokenAddr),
		parsedABI: parsedABI,
	}, nil
}

// Get token balance (đơn vị chuẩn, ví dụ: BANKVN = 18 decimals)
func (t *TokenService) GetTokenBalance(walletAddress string) (*big.Float, error) {
	account := common.HexToAddress(walletAddress)

	// Pack dữ liệu gọi hàm balanceOf(address) -> Tạo ra payload nhị phân: balanceOf(0xc3C2D3C831C4907131A1ECe81C2cc3b376F25341)
	callData, err := t.parsedABI.Pack("balanceOf", account)
	if err != nil {
		return nil, fmt.Errorf("failed to pack balanceOf call: %v", err)
	}

	// Gửi truy vấn eth_call (chỉ đọc, kh ghi data lên chain, nên không tốn gas)
	result, err := t.client.CallContract(context.Background(), ethereum.CallMsg{
		To:   &t.tokenAddr,
		Data: callData,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("contract call failed: %v", err)
	}

	// Giải mã dữ liệu trả về từ EVM - uint256
	out, err := t.parsedABI.Unpack("balanceOf", result)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack result: %v", err)
	}

	// Convert sang kiểu big.Int để xử lý trong Go.
	balance := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	// Hiển thị balance raw (đơn vị Wei — 1 token = 10^18 Wei nếu decimals = 18)
	// fmt.Printf("Token balance (raw Wei): %s\n", balance.String())

	// Convert Wei -> Token thực tế.
	// Giả sử token có 18 chữ số thập phân (chuẩn ERC20)
	decimals := big.NewFloat(1e18)
	fBalance := new(big.Float).SetInt(balance)
	tokenValue := new(big.Float).Quo(fBalance, decimals)

	return tokenValue, nil
}

func main() {
	// Config RPC and contract
	rpcURL := "https://eth-sepolia.g.alchemy.com/v2/n2..."    // Sepolia RPC - Dùng Alchemy or ...
	userAddr := "0xc3C2D3C831C4907131A1ECe81C2cc3b376F25341"  // Địa chỉ ví muốn check
	tokenAddr := "0x4D2c45779c89103c76Ae9cF5FE5596007a6cf815" // BANKVN Token (Token muốn check trong ví)

	// Init service token
	service, err := NewTokenService(rpcURL, tokenAddr)
	if err != nil {
		log.Fatal("Init TokenService faild:", err)
	}

	// Get token balance
	balance, err := service.GetTokenBalance(userAddr)
	if err != nil {
		log.Fatal("Failed to get token balance:", err)
	}

	fmt.Printf("==== Wallet: %s\n", userAddr)
	fmt.Printf("==== Balance: %f %s\n", balance, "BANKVN")
}
