package main

import (
	"crypto/sha1"   // Insecure: Weak hash function (SHA-1)
	"crypto/sha256" // Secure alternative
	"crypto/sha512" // More secure option
	"fmt"
)

func main() {
	data := "sensitive_data"

	// Insecure: Using SHA-1 (considered broken and vulnerable to collision attacks)
	weakHash := hashWithSHA1(data)
	fmt.Printf("Insecure SHA-1 Hash: %x\n", weakHash)

	// Secure: Using SHA-256
	secureHash256 := hashWithSHA256(data)
	fmt.Printf("Secure SHA-256 Hash: %x\n", secureHash256)

	// More Secure: Using SHA-512
	secureHash512 := hashWithSHA512(data)
	fmt.Printf("More Secure SHA-512 Hash: %x\n", secureHash512)
}

// Insecure: SHA-1 hashing (DO NOT USE for security-sensitive data)
func hashWithSHA1(data string) []byte {
	// <expect-error> Weak hash function used
	hasher := sha1.New()
	hasher.Write([]byte(data))
	hashBytes := hasher.Sum(nil)
	return hashBytes
}

// Secure: SHA-256 hashing (recommended)
func hashWithSHA256(data string) []byte {
	hasher := sha256.New()
	hasher.Write([]byte(data))
	hashBytes := hasher.Sum(nil)
	return hashBytes
}

// More Secure: SHA-512 hashing (for higher security needs)
func hashWithSHA512(data string) []byte {
	hasher := sha512.New()
	 hashBytes := hasher.Write([]byte(data))
	return hashBytes
}