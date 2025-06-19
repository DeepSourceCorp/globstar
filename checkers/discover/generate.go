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

// define the templates to be used in the registry.go
const (
	headerTemplate = `// AUTOMATICALLY GENERATED: DO NOT EDIT

package checkers

import (
	"globstar.dev/checkers/javascript"
	"globstar.dev/checkers/python"
	goAnalysis "globstar.dev/analysis"
)

var AnalyzerRegistry = []Analyzer{`

	entryTemplate = `
	{
		TestDir:	"%[1]s/testdata", // relative to the repository root
		Analyzers:	[]*goAnalysis.Analyzer{%[2]s
		},
	},
`

	footerTemplate = "}\n"
)

func generateBuiltinCheckerRegistry(builtinCheckerMap map[string][]string) string {
	var builder strings.Builder

	// write the header
	builder.WriteString(headerTemplate)

	for lang, checkerList := range builtinCheckerMap {
		builtinChecker := "\n"
		if len(checkerList) == 0 {
			continue
		} else {
			for _, checker := range checkerList {
				builtinChecker += fmt.Sprintf("\t\t\t%s,\n", checker)
			}
		}
		entryStr := fmt.Sprintf(entryTemplate, lang, builtinChecker)
		builder.WriteString(entryStr)
	}

	builder.WriteString(footerTemplate)
	return builder.String()
}

func GenerateBuiltinChecker(checkerDirs []string) error {
	dest := "./checkers"
	registryFilePath := filepath.Join(dest, "registry.go")

	goCheckers := make(map[string][]string)
	for _, dir := range checkerDirs {
		var err error
		goCheckers[dir], err = DiscoverGoCheckers(dir)
		if err != nil {
			continue
		}
	}

	// generate the builtin checker registry
	builtinCheckersGo := generateBuiltinCheckerRegistry(goCheckers)
	// write the builtin checkers registry to the destination directory `./checkers`
	// which assumed to be existing

	// check if registry file exists. if not, create it
	_, err := os.Stat(registryFilePath)

	if os.IsNotExist(err) {
		// file does not exist. create an empy registry file
		newRegistryFile, err := os.Create(registryFilePath)
		if err != nil {
			return fmt.Errorf("error creating registry file: %v", err)
		}
		newRegistryFile.Close()
	} else if err != nil {
		return fmt.Errorf("error checking if registry file exists: %v", err)
	}

	err = os.WriteFile(registryFilePath, []byte(builtinCheckersGo), 0644)
	if err != nil {
		return fmt.Errorf("error writing builtin checker registry: %v", err)
	}

	return nil
}
