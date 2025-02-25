import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// CustomClaims defines custom JWT claims
type CustomClaims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func main() {
	// Define claims
	claims := CustomClaims{
		Username: "user123",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // Token expires in 24 hours
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	insecureToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	//<expect-error> Hardcoding secrets is insecure
	tokenStringInsecure, err := insecureToken.SignedString([]byte("secret"))
	if err != nil {
		log.Fatalf("Error signing token (insecure): %v", err)
	}
	fmt.Println("Insecure token:", tokenStringInsecure)

	// Secure: Signing key from environment variable
	secretKey := os.Getenv("SECRET") // 🔑 Get signing key securely
	if secretKey == "" {
		log.Fatal("Environment variable SECRET not set")
	}

	secureToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStringSecure, err := secureToken.SignedString([]byte(secretKey))
	if err != nil {
		log.Fatalf("Error signing token (secure): %v", err)
	}
	fmt.Println("Secure token:", tokenStringSecure)
}
