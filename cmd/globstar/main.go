package main

import (
	"flag"
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

	outputFormat := flag.String("output-format", "text", "Specify the output format (text, json, sarif)")
	flag.Parse()

	cli := cli.Cli{
		RootDirectory: cwd,
		Rules:         nil, // no custom rule set
		OutputFormat:  *outputFormat,
	}

	err = cli.Run()
	if err != nil {
		os.Exit(1)
	}
}
