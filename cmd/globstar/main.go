package main

import (
	"fmt"
	"os"

	"globstar.dev/pkg/cli"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	cli := cli.Cli{
		RootDirectory: cwd,
		Checkers:      nil, // no custom checker set
	}

	err = cli.Run()
	if err != nil {
		os.Exit(1)
	}
}
