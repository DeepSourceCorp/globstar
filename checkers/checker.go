package checkers

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	goAnalysis "globstar.dev/analysis"
	"globstar.dev/checkers/javascript"
	"globstar.dev/pkg/analysis"
)

//go:embed **/*.y*ml
var builtinCheckers embed.FS

func LoadYamlRules() (map[analysis.Language][]analysis.YmlRule, error) {
	rulesMap := make(map[analysis.Language][]analysis.YmlRule)
	err := fs.WalkDir(builtinCheckers, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if d.IsDir() {
			return nil
		}

		fileExt := filepath.Ext(path)
		isYamlFile := fileExt == ".yaml" || fileExt == ".yml"
		if !isYamlFile {
			return nil
		}

		fileContent, err := builtinCheckers.ReadFile(path)
		if err != nil {
			return nil
		}

		patternRule, err := analysis.ReadFromBytes(fileContent)
		if err != nil {
			return fmt.Errorf("invalid rule '%s': %s", d.Name(), err.Error())
		}

		lang := patternRule.Language()
		rulesMap[lang] = append(rulesMap[lang], patternRule)
		return nil
	})
	return rulesMap, err
}

type Analyzer struct {
	TestDir   string
	Analyzers []*goAnalysis.Analyzer
}

var AnalyzerRegistry = []Analyzer{
	{
		TestDir:   "checkers/javascript/testdata", // relative to the repository root
		Analyzers: []*goAnalysis.Analyzer{javascript.NoDoubleEq},
	},
}

func LoadGoRules() []*goAnalysis.Analyzer {
	analyzers := []*goAnalysis.Analyzer{}

	for _, analyzer := range AnalyzerRegistry {
		analyzers = append(analyzers, analyzer.Analyzers...)
	}
	return analyzers
}

func RunAnalyzerTests(analyzerRegistry []Analyzer) (bool, []error) {
	passed := true
	errors := []error{}
	cwd, err := os.Getwd()
	if err != nil {
		errors = append(errors, err)
		return false, errors
	}

	for _, analyzerReg := range analyzerRegistry {

		fmt.Printf("Running tests in %s for analyzers:\n", analyzerReg.TestDir)
		testDir := filepath.Join(cwd, analyzerReg.TestDir)

		diff, log, isPassed, err := goAnalysis.RunAnalyzerTests(testDir, analyzerReg.Analyzers)
		if err != nil {
			errors = append(errors, err)
		}

		fmt.Println(log)

		if !isPassed {
			fmt.Printf("Issues raised are not as expected:\n%s\n", diff)
			passed = false
		}
	}

	return passed, errors
}
