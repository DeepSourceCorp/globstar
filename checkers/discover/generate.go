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

// for file checkers/registry.go
// modify this template when adding checkers for a new language
var builtinCheckersGo = `// AUTOMATICALLY GENERATED: DO NOT EDIT

package checkers

import (
	"globstar.dev/checkers/javascript"
	"globstar.dev/checkers/python"
	goAnalysis "globstar.dev/analysis"
)

type Analyzer struct {
	TestDir   string
	Analyzers []*goAnalysis.Analyzer
}

var AnalyzerRegistry = []Analyzer{
	{
		TestDir:   "checkers/javascript/testdata", // relative to the repository root
		Analyzers: []*goAnalysis.Analyzer{%[1]s
		},
	},
	{
		TestDir: "checkers/python/testdata",
		Analyzers: []*goAnalysis.Analyzer{%[2]s
		},
	},
}
`

func generateBuiltinCheckerRegistry(builtinCheckerMap map[string][]string) string {
	// a map with the language name as key and the checker name string as the value
	checkerListMap := make(map[string]string)
	for lang, builtinCheckerList := range builtinCheckerMap {
		builtinChecker := "\n"
		if len(builtinCheckerList) == 0 {
			checkerListMap[lang] = ""
		} else {
			for _, checker := range builtinCheckerList {
				builtinChecker += fmt.Sprintf("\t\t\t%s,\n", checker)
			}
		}
		checkerListMap[lang] = builtinChecker
	}
	return fmt.Sprintf(builtinCheckersGo, checkerListMap["javascript"], checkerListMap["python"])
}

func GenerateBuiltinChecker(checkerDirs []string) error {
	dest := "./checkers"
	goCheckers := make(map[string][]string)
	for _, dir := range checkerDirs {
		var err error
		goCheckers[dir], err = DiscoverGoCheckers("./checkers/" + dir)
		if err != nil {
			return fmt.Errorf("error discovering builtin checkers: %v", err)
		}

		if len(goCheckers[dir]) == 0 {
			return fmt.Errorf("no builtin Go checkers found in the directory: %s", dir)
		}
	}

	// generate the builtin checker registry
	builtinCheckersGo := generateBuiltinCheckerRegistry(goCheckers)
	// write the builtin checkers registry to the destination directory `./checkers`
	// which assumed to be existing
	err := os.WriteFile(filepath.Join(dest, "registry.go"), []byte(builtinCheckersGo), 0644)
	if err != nil {
		return fmt.Errorf("error writing builtin checker registry: %v", err)
	}

	return nil
}
