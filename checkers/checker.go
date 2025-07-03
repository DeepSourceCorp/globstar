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

func RunYamlAnalyzers(dir string) (passed bool, err error) {
	issues, err := analysis.RunYamlTests(dir)
	if err != nil {
		return false, fmt.Errorf("error running yaml tests: %w", err)
	}

	passed = true

	for test, yaml := range issues {

		if yaml.YamlAnalyzer.AnalysisFunction != nil {
			name := yaml.YamlAnalyzer.AnalysisFunction.Name
			lang := yaml.YamlAnalyzer.Analyzer.Language
			InitializeAnalysisFunctionDirectory(name, lang)
			analysisFuncAnalyzer := yaml.YamlAnalyzer.AnalysisFunction.Analyzer
			if analysisFuncAnalyzer == nil {
				return false, fmt.Errorf("no analysis function found for %s in %v", name, lang)
			}
			funcIssues, err := analysis.RunAnalyzers(test.TestFile, []*analysis.Analyzer{analysisFuncAnalyzer}, nil)
			if err != nil {
				return false, fmt.Errorf("error running analysis function for %s: %w", name, err)
			}
			for _, issue := range funcIssues {
				yaml.Got = append(yaml.Got, int(issue.Node.Range().StartPoint.Row)+1)
			}
		}

		if len(yaml.Want) != len(yaml.Got) {
			fmt.Println("Hmm... the number of issues raised is not as expected.")
			testName := filepath.Base(test.YamlCheckerPath)
			message := fmt.Sprintf(
				"(%s): expected issues on the following lines: %v\nbut issues were raised on lines: %v\n",
				testName,
				yaml.Want,
				yaml.Got,
			)
			fmt.Fprintf(os.Stderr, "%s", message)
			passed = false
			continue
		}
		for j := 0; j < len(yaml.Want); j++ {
			if yaml.Want[j] != yaml.Got[j] {
				testName := filepath.Base(test.YamlCheckerPath)
				message := fmt.Sprintf(
					"(%s): expected issue on line %d, but next occurrence is on line %d\n",
					testName,
					yaml.Want[j],
					yaml.Got[j],
				)
				fmt.Fprintf(os.Stderr, "%s\n", message)
				passed = false
			}

		}
	}

	return passed, nil
}
