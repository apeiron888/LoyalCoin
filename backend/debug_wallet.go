package main

import (
	"fmt"

	"github.com/loyalcoin/backend/internal/crypto"
)

func main() {
	fmt.Println("Testing wallet generation...")
	wallet, err := crypto.GenerateCardanoWallet(0x00) // Testnet
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Success! Address: %s\n", wallet.Address)
}
