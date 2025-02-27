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

var customRules []*analysis.Analyzer = []*analysis.Analyzer{%s}`

func generateAnalyzerRegistry(goRules []string) string {
	if len(goRules) == 0 {
		return fmt.Sprintf(analyzersGo, "")
	}
	customRules := "\n"
	for _, rule := range goRules {
		customRules += fmt.Sprintf("\t%s,\n", rule)
	}

	return fmt.Sprintf(analyzersGo, customRules)
}

// Generate the complete buildable analyzer from the .globstar directory
func GenerateAnalyzer(ruleDir, dest string) error {
	// Discover the custom rules from the rule directory
	goRules, err := DiscoverGoRules(ruleDir)
	if err != nil {
		return fmt.Errorf("error discovering custom rules: %v", err)
	}

	if len(goRules) == 0 {
		return fmt.Errorf("no custom Go rules found in the directory: %s", ruleDir)
	}

	// Copy the custom analyzer stub files to the destination directory
	err = CopyEmbeddedFiles(AnalyzerStubs, dest)
	if err != nil {
		return fmt.Errorf("error copying custom analyzer stub files: %v", err)
	}

	// Generate the custom analyzer registry
	analyzersGo := generateAnalyzerRegistry(goRules)
	// write the custom analyzer registry to the destination directory (the directory is assumed to be created, since we copied the stub files)
	err = os.WriteFile(filepath.Join(dest, "analyzers.go"), []byte(analyzersGo), 0644)
	if err != nil {
		return fmt.Errorf("error writing custom analyzer registry: %v", err)
	}

	// copy the custom rules to dest/checkers directory
	err = Copy(ruleDir, filepath.Join(dest, "checkers"), func(path string) bool {
		return strings.HasSuffix(path, ".go")
	})
	if err != nil {
		return fmt.Errorf("error copying custom rules: %v", err)
	}

	return nil
}
