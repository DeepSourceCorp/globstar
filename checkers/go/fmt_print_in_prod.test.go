// <expect-error>
fmt.Print()

// <expect-error>
fmt.Println()


package main

import (
	"fmt"
	"logger"
)

func main() {
	// <expect-error>
    fmt.Print("Hello, ")
	// <expect-error>
    fmt.Print("World!")
	// <expect-error>
	fmt.Println("Hello, World!") 
	// <expect-error>
    fmt.Println("Line 1")
	// <expect-error>
	fmt.Printf("Name: %s, Age: %d\n", name, age)

	// Safe
	logger.Info("Hello, World!")
	//Safe
	logger.Debug("Hello, World!")
}
