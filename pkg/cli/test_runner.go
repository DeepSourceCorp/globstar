package cli

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"globstar.dev/pkg/analysis"
)

func runTests(dir string) (bool, error) {
	passed, err := runTestCases(dir)
	if err != nil {
		return false, err
	}

	return passed, nil
}

type testCase struct {
	yamlCheckerPath string
	testFile        string
}

func findTestCases(dir string) ([]testCase, error) {
	var pairs []testCase // List of checker file/test file pairs

	err := filepath.Walk(dir, func(path string, d fs.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if d.IsDir() {
			return nil
		}

		if d.Mode()&fs.ModeSymlink != 0 {
			// skip symlinks
			return nil
		}

		fileExt := filepath.Ext(path)
		isYamlFile := fileExt == ".yaml" || fileExt == ".yml"
		if !isYamlFile {
			return nil
		}

		patternChecker, err := analysis.ReadFromFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid checker '%s': %s\n", d.Name(), err.Error())
			return nil
		}

		testFile := strings.TrimSuffix(path, fileExt) + ".test" + analysis.GetExtFromLanguage(patternChecker.Language())

		if _, err := os.Stat(testFile); os.IsNotExist(err) {
			testFile = ""
		}

		pairs = append(pairs, testCase{
			yamlCheckerPath: path,
			testFile:        testFile,
		})

		return nil
	})

	return pairs, err
}

func runTestCases(dir string) (passed bool, err error) {
	testCases, err := findTestCases(dir)
	if err != nil {
		return false, err
	}

	if len(testCases) == 0 {
		return false, fmt.Errorf("no test cases found")
	}

	passed = true
	for _, tc := range testCases {
		if tc.testFile == "" {
			fmt.Fprintf(os.Stderr, "No test cases found for test: %s\n", filepath.Base(tc.yamlCheckerPath))
			continue
		}

		fmt.Fprintf(os.Stderr, "Running test case: %s\n", filepath.Base(tc.yamlCheckerPath))
		// Read and parse the checker definition
		checker, err := analysis.ReadFromFile(tc.yamlCheckerPath)
		if err != nil {
			return false, err
		}

		// Parse the test file
		analyzer, err := analysis.FromFile(tc.testFile, []analysis.Checker{})
		if err != nil {
			return false, err
		}
		analyzer.WorkDir = dir
		analyzer.YamlCheckers = append(analyzer.YamlCheckers, checker)
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
				"(%s): expected issues on the following lines: %v\nbut issues were raised on lines: %v\n",
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
					"(%s): expected issue on line %d, but next occurrence is on line %d\n",
					testName,
					want,
					got,
				)

				fmt.Fprintf(os.Stderr, "%s\n", message)
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
