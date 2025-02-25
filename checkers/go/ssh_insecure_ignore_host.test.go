import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"log"
)

func test() {
	user := "your-username"
	password := "your-password"
	host := "your-server:22"

	// Insecure configuration (not recommended)
	// <expect-error> usage of InsecureIgnoreHostKey
	insecureConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), 
	}

	fmt.Println("Attempting insecure SSH connection...")
	client, err := ssh.Dial("tcp", host, insecureConfig)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()
	fmt.Println("Insecure SSH connection established.")

	// Safe configuration (recommended)
	safeConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: nil, // Safe - forces proper host key verification
	}

	fmt.Println("Attempting safe SSH connection...")
	clientSafe, err := ssh.Dial("tcp", host, safeConfig)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer clientSafe.Close()
	fmt.Println("Safe SSH connection established.")
}
