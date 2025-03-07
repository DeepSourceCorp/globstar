import (
	// <expect-error> Weak hash function used
	"crypto/md5"    // UNSAFE: Weak hash function
	"crypto/sha256" // SAFE: Strong hash function
	"encoding/hex"
	"fmt"
)

func testWeakHash(data string) {
	hasher := md5.New() // MD5 is considered cryptographically broken

	_, err := hasher.Write([]byte(data))
	if err != nil {
		fmt.Println("Error writing data to MD5 hasher:", err)
		return
	}

	hashBytes := hasher.Sum(nil)
	fmt.Printf("MD5 Hash (Weak): %s\n", hex.EncodeToString(hashBytes))
}

func testSafeHash(data string) {
	// Safe - using SHA256
	hasher := sha256.New()

	_, err := hasher.Write([]byte(data))
	if err != nil {
		fmt.Println("Error writing data to SHA256 hasher:", err)
		return
	}

	hashBytes := hasher.Sum(nil)
	fmt.Printf("SHA256 Hash (Secure): %s\n", hex.EncodeToString(hashBytes))
}
