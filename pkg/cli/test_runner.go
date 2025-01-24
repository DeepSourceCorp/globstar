package cli

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/srijan-paul/deepgrep/pkg/one"
)

func runTests(dir string) (bool, error) {
	passed, err := runTestCases(dir)
	if err != nil {
		message := fmt.Sprintf("failed %v", err)
		fmt.Fprintln(os.Stderr, message)
		return false, err
	}

	return passed, nil
}

type testCase struct {
	ymlRulePath string
	testFile    string
}

func findTestCases(dir string) ([]testCase, error) {
	var pairs []testCase // List of rule file/test file pairs

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		fileName := d.Name()
		if strings.Contains(fileName, ".test") {
			baseName := strings.Split(fileName, ".test")[0]
			ymlFilePath := filepath.Join(dir, baseName+".yml")
			if _, err := os.Stat(ymlFilePath); err == nil {
				pairs = append(pairs, testCase{ymlRulePath: ymlFilePath, testFile: path})
			}

			ymlFilePath = baseName + ".yaml"
			if _, err := os.Stat(ymlFilePath); err == nil {
				pairs = append(pairs, testCase{ymlRulePath: ymlFilePath, testFile: path})
			}
		}

		return nil
	})

	return pairs, err
}

func runTestCases(dir string) (passed bool, err error) {
	testCases, err := findTestCases(dir)
	if err != nil {
		return false, err
	}

	passed = true
	for _, tc := range testCases {
		// Read and parse the rule definition
		rule, err := one.ReadFromFile(tc.ymlRulePath)
		if err != nil {
			return false, err
		}

		// Parse the test file
		parsed, err := one.ParseFile(tc.testFile)
		if err != nil {
			return false, err
		}

		// Run the analyzer
		analyzer := one.NewAnalyzer(parsed, nil)
		analyzer.YmlRules = append(analyzer.YmlRules, rule)
		issues := analyzer.Analyze()

		want, err := findExpectedLines(tc.testFile)
		if err != nil {
			return false, err
		}

		var got []int
		for _, issue := range issues {
			got = append(got, int(issue.Range.StartPoint.Row)+1) // 0-indexed to 1-indexed
		}

		slices.Sort(got)

		testName := filepath.Base(tc.testFile)

		if len(want) != len(got) {
			message := fmt.Sprintf(
				"(%s): expected issues on the following lines: %v\nbut issues were raised on lines: %v",
				testName,
				want,
				got,
			)

			fmt.Fprintf(os.Stderr, "%s", message)
			passed = false
			continue
		}

		for i := range want {
			if want[i] != got[i] {
				message := fmt.Sprintf(
					"(%s): expected lint on line %d, but next occurrence is on line %d",
					testName,
					want,
					got,
				)

				fmt.Fprintf(os.Stderr, "%s", message)
				passed = false
			}
		}
	}

	return passed, nil
}

// findExpectedLines reads a file and returns line numbers containing "<expect-error>"
// (incremented by 1).
func findExpectedLines(filePath string) ([]int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var expectedLines []int
	scanner := bufio.NewScanner(file)

	lineNumber := 0
	for scanner.Scan() {
		text := strings.ToLower(scanner.Text())
		lineNumber++
		if strings.Contains(text, "<expect-error>") || strings.Contains(text, "<expect error>") {
			expectedLines = append(expectedLines, lineNumber+1)
		}
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return expectedLines, nil
}
