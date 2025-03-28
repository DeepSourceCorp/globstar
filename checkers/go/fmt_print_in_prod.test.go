package golang

import (
	"fmt"
	"log"
)

func testFmtInProd() {
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
	log.Println("Hello, World!")
	//Safe
	log.Println("Hello, World!")
}
