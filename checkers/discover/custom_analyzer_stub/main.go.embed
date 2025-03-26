package main

import (
	"flag"
	"fmt"
	"os"

	"globstar.dev/analysis"
)

var (
	path = flag.String("path", ".", "Path to the directory to analyze")
	test = flag.Bool("test", false, "Run the tests")
)

func main() {
	flag.Parse()

	if *test {
		fmt.Fprintf(os.Stderr, "Running tests in %s for analyzers\n", *path)
		diff, log, passed, err := analysis.RunAnalyzerTests(*path, customCheckers)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error running tests: %s", err.Error())
			os.Exit(1)
		}

		fmt.Fprintln(os.Stderr, log)

		if !passed {
			fmt.Fprintln(os.Stderr, "Tests failed")
			fmt.Fprintf(os.Stderr, "Diff: %s\n", diff)
			os.Exit(1)
		} else {
			fmt.Fprintln(os.Stderr, "Tests passed")
		}
		os.Exit(0)
	} else {
		issues, err := analysis.RunAnalyzers(*path, customCheckers, nil)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}

		for _, issue := range issues {
			txt, _ := issue.AsJson()
			fmt.Fprintln(os.Stderr, string(txt))
		}

		if len(issues) > 0 {
			os.Exit(1)
		} else {
			os.Exit(0)
		}
	}
}
