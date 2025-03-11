package discover

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// for file analyzers.go
var analyzersGo = `package main

import (
	"globstar.dev/analysis"
	"globstar.dev/customanalyzer/checkers"
)

var customCheckers []*analysis.Analyzer = []*analysis.Analyzer{%s}`

func generateAnalyzerRegistry(goCheckers []string) string {
	if len(goCheckers) == 0 {
		return fmt.Sprintf(analyzersGo, "")
	}
	customCheckers := "\n"
	for _, checker := range goCheckers {
		customCheckers += fmt.Sprintf("\t%s,\n", checker)
	}

	return fmt.Sprintf(analyzersGo, customCheckers)
}

// Generate the complete buildable analyzer from the .globstar directory
func GenerateAnalyzer(checkerDir, dest string) error {
	// Discover the custom checkers from the checker directory
	goCheckers, err := DiscoverGoCheckers(checkerDir)
	if err != nil {
		return fmt.Errorf("error discovering custom checkers: %v", err)
	}

	if len(goCheckers) == 0 {
		return fmt.Errorf("no custom Go checkers found in the directory: %s", checkerDir)
	}

	// Copy the custom analyzer stub files to the destination directory
	err = CopyEmbeddedFiles(AnalyzerStubs, dest)
	if err != nil {
		return fmt.Errorf("error copying custom analyzer stub files: %v", err)
	}

	// Generate the custom analyzer registry
	analyzersGo := generateAnalyzerRegistry(goCheckers)
	// write the custom analyzer registry to the destination directory (the directory is assumed to be created, since we copied the stub files)
	err = os.WriteFile(filepath.Join(dest, "analyzers.go"), []byte(analyzersGo), 0644)
	if err != nil {
		return fmt.Errorf("error writing custom analyzer registry: %v", err)
	}

	// copy the custom checkers to dest/checkers directory
	err = Copy(checkerDir, filepath.Join(dest, "checkers"), func(path string) bool {
		return strings.HasSuffix(path, ".go")
	})
	if err != nil {
		return fmt.Errorf("error copying custom checkers: %v", err)
	}

	return nil
}
