import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
)

func test() {
	// Insecure: tls.Config without MinVersion set (allows deprecated protocols like TLS 1.0 and TLS 1.1)
	// <expect-error> tls.Config.MinVersion is not set
	insecureConfig := &tls.Config{
		ServerName: "example.com",
	}

	// Attempting connection with insecure config (for demonstration purposes only)
	insecureClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: insecureConfig,
		},
	}
	_, err := insecureClient.Get("https://example.com")
	if err != nil {
		log.Printf("Insecure connection failed (as expected): %v\n", err)
	}

	// Secure: tls.Config with MinVersion set to tls.VersionTLS12
	secureConfig := &tls.Config{
		ServerName: "example.com",
		// Safe: Enforces TLS 1.2 or higher for secure connections
		MinVersion: tls.VersionTLS12,
	}

	secureClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: secureConfig,
		},
	}
	resp, err := secureClient.Get("https://example.com")
	if err != nil {
		log.Fatalf("Secure connection failed: %v\n", err)
	}
	defer resp.Body.Close()

	fmt.Println("Secure connection established with TLS version enforcement.")
}
