package checkers

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"globstar.dev/analysis"
)

//go:embed **/*.y*ml
var builtinCheckers embed.FS

func findYamlCheckers(checkersMap map[analysis.Language][]analysis.Analyzer) func(path string, d fs.DirEntry, err error) error {
	return func(path string, d fs.DirEntry, err error) error {
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

		patternChecker, _, err := analysis.ReadFromBytes(fileContent)
		if err != nil {
			return fmt.Errorf("invalid checker '%s': %s", d.Name(), err.Error())
		}

		lang := patternChecker.Language
		checkersMap[lang] = append(checkersMap[lang], patternChecker)
		return nil
	}
}

func LoadBuiltinYamlCheckers() (map[analysis.Language][]analysis.Analyzer, error) {
	checkersMap := make(map[analysis.Language][]analysis.Analyzer)
	err := fs.WalkDir(builtinCheckers, ".", findYamlCheckers(checkersMap))
	return checkersMap, err
}

func LoadCustomYamlCheckers(dir string) (map[analysis.Language][]analysis.Analyzer, error) {
	checkersMap := make(map[analysis.Language][]analysis.Analyzer)
	err := fs.WalkDir(os.DirFS(dir), ".", findYamlCheckers(checkersMap))
	return checkersMap, err
}

type Analyzer struct {
	TestDir   string
	Analyzers []*analysis.Analyzer
}

func LoadGoCheckers() []*analysis.Analyzer {
	analyzers := []*analysis.Analyzer{}

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

		diff, log, isPassed, err := analysis.RunAnalyzerTests(testDir, analyzerReg.Analyzers)
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
