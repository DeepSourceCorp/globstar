import (
	"fmt"
	"os"
)

func test() {
	fileName := "example.txt"

	// <expect-error> os.Create is called (which uses default file permissions)
	file, err := os.Create(fileName)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()
	fmt.Println("File created with default permissions (potential security risk).")

	// Safe alternative using os.OpenFile with explicit permissions (0600)
	file, err = os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Println("Error creating file securely:", err)
		return
	}
	defer file.Close()
	fmt.Println("File created securely with 0600 permissions.")
}
