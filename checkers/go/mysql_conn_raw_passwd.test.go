import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql" // MySQL driver
)

func unsafe_test() {
	// <expect-error> Hardcoded connection string - unsafe practice
	db, err := sql.Open("mysql", "user:password@tcp(localhost:3306)/dbname")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("Successfully connected to the database!")
}

func safe_test() {
	// Retrieve configuration from environment variables
	user := os.Getenv("DB_USER")
	if user == "" {
		log.Fatal("Environment variable DB_USER is not set")
	}
	passwd := os.Getenv("DB_PASSWORD")
	if passwd == "" {
		log.Fatal("Environment variable DB_PASSWORD is not set")
	}
	db := os.Getenv("DB_NAME")
	if db == "" {
		log.Fatal("Environment variable DB_NAME is not set")
	}
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost" // Default value
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "3306" // Default MySQL port
	}

	// Safe: Construct connection string dynamically
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", user, passwd, host, port, db)
	dbConn, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer dbConn.Close()

	// Test the connection
	err = dbConn.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("Successfully connected to the database!")
}