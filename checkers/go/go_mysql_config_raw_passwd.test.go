import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql" // MySQL driver
	"github.com/go-sql-driver/mysql"
)

func test() {
	// Hardcoded credentials - unsafe practice
	// <expect-error>
	cfg := mysql.Config{
		User:   "admin", // Hardcoded username
		Passwd: "admin", // Hardcoded password
		Net:    "tcp",
		Addr:   "localhost:3306",
		DBName: "mydb",
	}

	passwdEnv := os.Getenv("MYSQL_PASSWORD")
	if passwdEnv == "" {
		log.Fatal("Environment variable MYSQL_PASSWORD is not set")
	}

	// Safe configuration using environment variable
	cfg := mysql.Config{
		User:   "admin",       // Username could also be an env var if needed
		Passwd: passwdEnv,     // Password from environment variable
		Net:    "tcp",
		Addr:   "localhost:3306",
		DBName: "mydb",
	}

	// Create DSN (Data Source Name) from config
	dsn := cfg.FormatDSN()

	// Open database connection
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Test the connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("Successfully connected to the database!")
}