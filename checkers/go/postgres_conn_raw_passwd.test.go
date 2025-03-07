import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq" // PostgreSQL driver
)

func test() {
	// Insecure: Hardcoded password - Vulnerable to credential leaks
	// <expect-error>
	dbInsecure, err := sql.Open("postgres", "postgres://myuser:hardcodedPassword456@localhost:5432/mydb")
	if err != nil {
		log.Fatalf("Failed to connect (insecure): %v", err)
	}
	defer dbInsecure.Close()

	fmt.Println("Insecure connection established (not recommended).")

	// Secure: Use environment variables to avoid hardcoding sensitive data
	user := os.Getenv("PG_USER")
	passwd := os.Getenv("PG_PASSWORD")
	host := os.Getenv("PG_HOST")
	port := os.Getenv("PG_PORT")
	name := os.Getenv("PG_DBNAME")

	// Check if essential environment variables are set
	if user == "" || passwd == "" || host == "" || port == "" || name == "" {
		log.Fatal("One or more environment variables (PG_USER, PG_PASSWORD, PG_HOST, PG_PORT, PG_DBNAME) are missing.")
	}

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, passwd, host, port, name)
	dbSecure, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect (secure): %v", err)
	}
	defer dbSecure.Close()

	fmt.Println("Secure connection established using environment variables.")
}
