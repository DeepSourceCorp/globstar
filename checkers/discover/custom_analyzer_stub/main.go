package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"globstar.dev/analysis"
)

var (
	path = flag.String("path", ".", "Path to the directory to analyze")
	test = flag.Bool("test", false, "Run the tests")
)

func main() {
	flag.Parse()

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	if *test {
		diff, passed, err := analysis.RunAnalyzerTests(*path, customRules)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}

		if !passed {
			fmt.Fprintln(os.Stderr, "Tests failed")
			fmt.Fprintf(os.Stderr, "Diff: %s\n", diff)
			os.Exit(1)
		}
		os.Exit(0)
	} else {
		issues, err := analysis.RunAnalyzers(*path, customRules)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}

		for _, issue := range issues {
			txt, _ := issue.AsText()
			log.Error().Msg(string(txt))
		}

		if len(issues) > 0 {
			os.Exit(1)
		} else {
			os.Exit(0)
		}
	}
}
