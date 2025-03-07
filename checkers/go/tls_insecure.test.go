
import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
)

func test() {
	// Insecure: InsecureSkipVerify set to true (disables certificate verification)
	// <expect-error> InsecureSkipVerify is set to true
	insecureConfig := &tls.Config{
		ServerName:         "example.com",
		InsecureSkipVerify: true, // ⚠️ Dangerous: Skips server certificate verification
	}

	insecureClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: insecureConfig,
		},
	}

	_, err := insecureClient.Get("https://example.com")
	if err != nil {
		log.Printf("Insecure connection failed (as expected): %v\n", err)
	} else {
		log.Println("⚠️ Connection established without verifying the server certificate (unsafe).")
	}

	// Safe Example 1: InsecureSkipVerify not set (defaults to false)
	safeConfigDefault := &tls.Config{
		ServerName: "example.com",
	}

	safeClientDefault := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: safeConfigDefault,
		},
	}

	resp, err := safeClientDefault.Get("https://example.com")
	if err != nil {
		log.Fatalf("Secure connection failed: %v\n", err)
	}
	defer resp.Body.Close()
	fmt.Println("Secure connection established with default certificate verification.")

	// Safe Example 2: InsecureSkipVerify explicitly set to false
	safeConfigExplicit := &tls.Config{
		ServerName:         "example.com",
		InsecureSkipVerify: false, // 🔒 Explicitly ensures certificate verification
	}

	safeClientExplicit := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: safeConfigExplicit,
		},
	}

	resp2, err := safeClientExplicit.Get("https://example.com")
	if err != nil {
		log.Fatalf("Secure connection failed: %v\n", err)
	}
	defer resp2.Body.Close()
	fmt.Println("Secure connection established with explicit verification.")
}
