package main

import (
	"crypto/rand"
	"fmt"
	// <expect-error> import "math/rand"
	"math/rand"
	"math/big"
	"time"
)

func unsafeRandomExample(n int) {
	// -------------------------
	// UNSAFE: Using math/rand for security-sensitive randomness
	// -------------------------
	// Weak random number generator
	mathrand.Seed(time.Now().UnixNano())
	randomNumber := rand.Intn(n)     // Not suitable for cryptographic purposes
	fmt.Printf("Unsafe random number (math/rand): %d\n", randomNumber)
}

func safeRandomExample(n int) {
	// -------------------------
	// SAFE: Using crypto/rand for secure randomness
	// -------------------------
	// crypto/rand provides cryptographically secure random numbers
	randomNumber, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if err != nil {
		fmt.Println("Error generating secure random number:", err)
		return
	}
	fmt.Printf("Safe random number (crypto/rand): %d\n", randomNumber.Int64())
}
