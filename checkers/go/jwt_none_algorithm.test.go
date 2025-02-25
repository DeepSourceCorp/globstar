package main

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// CustomClaims defines the JWT claims structure.
type CustomClaims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func main() {
	// Create claims with username and expiration time
	claims := CustomClaims{
		Username: "example_user",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)), // Token expires in 1 hour
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "your-app",
		},
	}

	// -------------------------
	// Unsafe token creation (SigningMethodNone) - SHOULD NOT BE USED
	// -------------------------
	// <expect-error>
	tokenNone := jwt.NewWithClaims(jwt.SigningMethodNone, claims)

	// Attempting to sign with 'None' signature type is unsafe and should be avoided
	// <expect-error>
	ssNone, err := tokenNone.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		fmt.Println("Expected error (None Signature):", err)
	} else {
		fmt.Println("Unsafe Token (None Signature):", ssNone)
	}

	// -------------------------
	// Safe token creation (HS256) - Recommended approach
	// -------------------------
	// Make sure the SECRET environment variable is set
	secret := os.Getenv("SECRET")
	if secret == "" {
		fmt.Println("SECRET environment variable is not set.")
		return
	}

	// Create a token using the HS256 signing method
	tokenHS256 := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the provided secret
	ssHS256, err := tokenHS256.SignedString([]byte(secret))
	if err != nil {
		fmt.Println("Error signing token with HS256:", err)
		return
	}

	// Safe signed token output
	fmt.Println("Safe Token (HS256):", ssHS256)
}
