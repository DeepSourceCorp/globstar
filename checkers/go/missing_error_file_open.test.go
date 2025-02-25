import (
	"fmt"
	"os"
)

func badFileHandling() error {
	// <expect-error> error not captured
	f, _ := os.Open("data.txt")
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	data := make([]byte, 100)
	_, err = f.Read(data)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	return nil
}

func goodFileHandling() error {
	// Safe err is captured
	f, err := os.Open("data.txt") 
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close() 
	data := make([]byte, 100)
	_, err = f.Read(data)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	return nil
}
