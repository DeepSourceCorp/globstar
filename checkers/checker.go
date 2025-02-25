package checkers

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/go-cmp/cmp"
	sitter "github.com/smacker/go-tree-sitter"
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
		Analyzers: []*goAnalysis.Analyzer{&javascript.NoDoubleEq},
	},
}

func LoadGoRules() []*goAnalysis.Analyzer {
	analyzers := []*goAnalysis.Analyzer{}

	for _, analyzer := range AnalyzerRegistry {
		analyzers = append(analyzers, analyzer.Analyzers...)
	}
	return analyzers
}

func RunAnalyzerTests(analyzerRegistry []Analyzer) {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("error getting current working directory: %v\n", err)
	}
	for _, analyzerReg := range analyzerRegistry {
		testDir := filepath.Join(cwd, analyzerReg.TestDir)
		expectedIssues, err := getExpectedIssuesInDir(testDir)
		if err != nil {
			fmt.Printf("error getting expected issues in dir %s: %v\n", testDir, err)
			return
		}

		raisedIssues, err := goAnalysis.RunAnalyzers(testDir, analyzerReg.Analyzers)
		if err != nil {
			fmt.Printf("error running tests on dir %s: %v\n", testDir, err)
			return
		}

		raisedIssuesMap := make(map[string]map[int][]string)
		for _, issue := range raisedIssues {
			if _, ok := raisedIssuesMap[issue.Filepath]; !ok {
				raisedIssuesMap[issue.Filepath] = make(map[int][]string)
			}

			line := int(issue.Node.Range().StartPoint.Row + 1)
			if _, ok := raisedIssuesMap[issue.Filepath][line]; !ok {
				raisedIssuesMap[issue.Filepath][line] = []string{}
			}

			raisedIssuesMap[issue.Filepath][line] = append(raisedIssuesMap[issue.Filepath][line], issue.Message)
		}

		// verify issues raised are as expected from the test files
		diff := verifyIssues(&expectedIssues, &raisedIssuesMap)
		if diff != "" {
			fmt.Printf("Issues raised in %s are not as expected:\n%s\n", testDir, diff)
		} else {
			fmt.Printf("All tests passed in %s\n", testDir)
		}
	}
}

func verifyIssues(expectedIssues, raisedIssues *map[string]map[int][]string) string {
	// TODO: return a human readable string of the diff
	return cmp.Diff(expectedIssues, raisedIssues)
}

func getExpectedIssuesInDir(testDir string) (map[string]map[int][]string, error) {
	// map of test file path to map of line number to issue message
	// {"file.test.ext": {1: {"issue1 message"}, {"issue2 message"}}}
	expectedIssues := make(map[string]map[int][]string)
	err := filepath.Walk(testDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, fmt.Sprintf(".test%s", filepath.Ext(path))) {
			return nil
		}

		// load the pragmas (<commentIdentifier> <expect-error>) from the test file
		file, err := goAnalysis.ParseFile(path)
		if err != nil {
			// skip the file if it can't be parsed
			return nil
		}

		query, err := sitter.NewQuery([]byte("(comment) @pragma"), file.Language.Grammar())
		if err != nil {
			return nil
		}

		expectedIssues[path] = getExpectedIssuesInFile(file, query)

		return nil
	})
	if err != nil {
		return expectedIssues, err
	}

	return expectedIssues, nil
}

func getExpectedIssuesInFile(file *goAnalysis.ParseResult, query *sitter.Query) map[int][]string {
	commentIdentifier := goAnalysis.CommentIdentifierFromPath(file.FilePath)
	commentIdentifier = strings.ReplaceAll(commentIdentifier, "(*", "\\(\\*")
	commentIdentifier = strings.ReplaceAll(commentIdentifier, "//", "\\/\\/")
	commentIdentifier = strings.ReplaceAll(commentIdentifier, "<!--", "<\\!--")

	pattern := fmt.Sprintf(`^%s\s+<expect-error>\s+(?P<message>.+)$`, commentIdentifier)
	pragmaRegexp := regexp.MustCompile(pattern)

	expectedIssues := map[int][]string{}
	cursor := sitter.NewQueryCursor()
	cursor.Exec(query, file.Ast)
	for {
		m, ok := cursor.NextMatch()

		if !ok {
			break
		}

		for _, capture := range m.Captures {
			captureName := query.CaptureNameForId(capture.Index)
			if captureName != "pragma" {
				continue
			}
			expectedLine := -1
			pragma := capture.Node.Content(file.Source)
			prevNode := capture.Node.PrevSibling()
			if prevNode != nil && (prevNode.EndPoint().Row == capture.Node.StartPoint().Row) {
				// if the comment is on the same line as the troublesome code,
				// the line number of the issue is the same as the line number of the comment
				expectedLine = int(prevNode.StartPoint().Row) + 1
			} else {
				// +2 because the pragma is on the line above the expected issue,
				// and the line number is 0-indexed
				expectedLine = int(capture.Node.StartPoint().Row) + 2
			}
			// TODO match using regex, use groups to find the expected issue message as well
			matches := pragmaRegexp.FindAllStringSubmatch(pragma, -1)
			if matches == nil {
				continue
			}

			message := ""
			for _, match := range matches {
				for i, group := range pragmaRegexp.SubexpNames() {
					if i == 0 || group == "" {
						continue
					}

					if group == "message" {
						message = match[i]
					}
				}

				if _, ok := expectedIssues[expectedLine]; !ok {
					expectedIssues[expectedLine] = []string{}
				}

				expectedIssues[expectedLine] = append(expectedIssues[expectedLine], message)
			}
		}
	}
	return expectedIssues
}
