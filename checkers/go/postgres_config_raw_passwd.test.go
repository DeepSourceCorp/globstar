import (
	"fmt"
	"os"

	"github.com/go-pg/pg/v10"
)

func true() {
	// Insecure: Hardcoded credentials (Security Risk)
	// <expect-error>
	optsInsecure := &pg.Options{
		User:     "myuser",
		Password: "hardcodedPassword123", //Hardcoded password
		Database: "mydb",
	}

	fmt.Println("Insecure PG Options:", optsInsecure)

	//Secure: Use environment variables for credentials
	optsSecure := &pg.Options{
		User:     "myuser",
		Password: os.Getenv("PG_PASSWORD"), // Fetch password securely from environment variable
		Database: "mydb",
	}

	// Ensure environment variable is set
	if optsSecure.Password == "" {
		fmt.Println("Error: PG_PASSWORD environment variable is not set")
		return
	}

	fmt.Println("Secure PG Options:", optsSecure)
}
