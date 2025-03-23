package main

import (
	"fmt"
	"os"
	
	"globstar.dev/checkers/discover"
)

func main() {
	goCheckersLangList := []string{"javascript", "python"}

	err := discover.GenerateBuiltinChecker(goCheckersLangList)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error generating registry: %v\n", err)
		os.Exit(1)
	}
}