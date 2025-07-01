package analysis

import (
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestReadFile(t *testing.T) {
	path := "./testdata/mock-checker.yml"
	ana, anaYaml, err := ReadFromFile(path)

	require.Nil(t, err)
	name := ana.Name
	assert.Equal(t, name, "mock-checker")
	language := ana.Language
	assert.Equal(t, language, LangJs)
	category := ana.Category
	assert.Equal(t, category, CategoryStyle)
	severity := ana.Severity
	assert.Equal(t, severity, SeverityInfo)
	assert.Equal(t, anaYaml.Message, "This is just a mock checker")
	assert.Equal(t, len(anaYaml.Patterns), 1)

	path = "./testdata/mock-wrong-checker.yml"
	_, _, err = ReadFromFile(path)

	assert.Error(t, err)

	path = "wrong_path.yml"
	_, _, err = ReadFromFile(path)

	assert.Error(t, err)
}

func TestNodeFiltersInside(t *testing.T) {
	jsData := `
	var globalVar = 1; // shouldn't match
	function test() {
		var localVar = 2; // Should match
		let anotherVar = 3; // should match
	}
	`
	path := "./testdata/node-filter-checker.yml"
	ana, _, err := ReadFromFile(path)
	require.NoError(t, err, "Failed to read YAML data")

	parsedJs, err := Parse("", []byte(jsData), LangJs, LangJs.Grammar())
	require.NoError(t, err, "Failed to parse JS data")

	var matchCount int
	var matches []string

	reportFunc := func(pass *Pass, node *sitter.Node, message string) {
		matchCount++
		t.Log(node.Content(pass.FileContext.Source))
		matches = append(matches, message)
	}

	pass := &Pass{
		Analyzer:    &ana,
		FileContext: parsedJs,
		Report:      reportFunc,
		Files:       []*ParseResult{parsedJs},
	}

	_, err = ana.Run(pass)
	require.NoError(t, err, "Failed to run YAML analyzer")
	assert.Equal(t, matchCount, 2, "Expected 2 matches")
}

func TestNodeFilterPatternNotInside(t *testing.T) {
	jsData := `
    var globalVar = 1; // Should match - not inside catch block
    try {
        var tryVar = 2; // Should match - not inside catch block
    } catch (e) {
        var catchVar = 3; // Should NOT match - inside catch block
        let anotherCatchVar = 4; // Should NOT match - inside catch block
    }
    var anotherGlobal = 5; // Should match - not inside catch block
    `

	yamlContent := `
language: javascript
name: no-vars-in-catch
message: "Variable declared outside catch block"
category: style
severity: info
pattern: (variable_declarator) @no-vars-in-catch
filters:
  - pattern-not-inside: (catch_clause)
description: "Test checker for pattern-not-inside"
`

	ana, _, err := ReadFromBytes([]byte(yamlContent))
	require.NoError(t, err, "Failed to read YAML data")

	parsedJs, err := Parse("", []byte(jsData), LangJs, LangJs.Grammar())
	require.NoError(t, err, "Failed to parse JS data")

	var matchCount int
	var matches []string

	reportFunc := func(pass *Pass, node *sitter.Node, message string) {
		matchCount++
		matches = append(matches, node.Content(pass.FileContext.Source))
		t.Logf("Match: %s", node.Content(pass.FileContext.Source))
	}

	pass := &Pass{
		Analyzer:    &ana,
		FileContext: parsedJs,
		Report:      reportFunc,
		Files:       []*ParseResult{parsedJs},
	}

	_, err = ana.Run(pass)
	require.NoError(t, err, "Failed to run YAML analyzer")

	// Should match 3 variables (globalVar, tryVar, anotherGlobal)
	// but NOT the 2 variables inside catch block (catchVar, anotherCatchVar)
	assert.Equal(t, 3, matchCount, "Expected 3 matches - variables not inside catch blocks")

	for _, match := range matches {
		assert.NotContains(t, match, "catchVar", "Should not match variables inside catch block")
		assert.NotContains(t, match, "anotherCatchVar", "Should not match variables inside catch block")
	}
}

func TestInvalidNodeFilters(t *testing.T) {
	inside :=
		`language: javascript
name: test
message: "Variable declared outside catch block"
category: style
severity: info
pattern: (variable_declarator) 
filters:
  - pattern-inside: "hello world"
description: "Test checker for pattern-inside"
`

	outside :=
		`language: javascript
name: test
message: "Variable declared outside catch block"
category: style
severity: info
pattern: (variable_declarator) 
filters:
  - pattern-not-inside: "hello"
description: "Test checker for pattern-inside"
`

	_, _, err := ReadFromBytes([]byte(inside))
	assert.EqualError(t, err, "invalid tree-sitter pattern inside 'pattern-inside' field")

	_, _, err = ReadFromBytes([]byte(outside))
	assert.EqualError(t, err, "invalid tree-sitter pattern inside 'pattern-not-inside' field")

}

func TestNodeFilterWithTests(t *testing.T) {
	path := "./testdata/node-filter-test-checker.yml"
	ana, yamlAna, err := ReadFromFile(path)
	require.NoError(t, err, "Failed to read YAML data")
	assert.Equal(t, ana.Name, "node-filter-test-checker")
	assert.Len(t, yamlAna.NodeFilter, 2)

	diff, log, passed, err := RunAnalyzerTests("./testdata", []*Analyzer{&ana})
	require.NoError(t, err, "Failed to run analyzer tests")
	t.Logf("Diff: %s", diff)
	t.Logf("Log: %s", log)
	assert.True(t, passed)
}

func TestPatterns(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		expectError    bool
		expectedErrMsg string
		expectEmpty    bool
		expectedCount  int
	}{
		{
			name:           "PatternAbsent",
			path:           "./testdata/yaml_tests/patterns/no-pattern.yml",
			expectError:    true,
			expectedErrMsg: "no pattern provided in checker 'no-pattern'",
			expectEmpty:    true,
		},
		{
			name:          "PatternMultiple",
			path:          "./testdata/yaml_tests/patterns/multi-pattern.yml",
			expectError:   false,
			expectEmpty:   false,
			expectedCount: 2,
		},
		{
			name:        "FaultyPattern",
			path:        "./testdata/yaml_tests/patterns/wrong-pattern.yml",
			expectError: true,
			expectEmpty: true,
		},
		{
			name:           "SingleAndMultiplePattern",
			path:           "./testdata/yaml_tests/patterns/single-multiple.yml",
			expectedErrMsg: "only one of 'pattern' or 'patterns' can be provided in a checker definition",
			expectError:    true,
			expectEmpty:    true,
		},
		{
			name:           "InvalidPatterns",
			path:           "./testdata/yaml_tests/patterns/invalid-patterns.yml",
			expectError:    true,
			expectedErrMsg: "invalid tree-sitter query in one of the patterns",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ana, yamlAna, err := ReadFromFile(tt.path)

			if tt.expectError {
				assert.Error(t, err)
				if tt.expectedErrMsg != "" {
					assert.Contains(t, err.Error(), tt.expectedErrMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			if tt.expectEmpty {
				assert.Equal(t, ana, Analyzer{})
				assert.Equal(t, yamlAna, YamlAnalyzer{})
			} else {
				assert.NotNil(t, ana)
				assert.NotNil(t, yamlAna)
				if tt.expectedCount > 0 {
					assert.Equal(t, len(yamlAna.Patterns), tt.expectedCount)
				}
			}
		})
	}
}

func TestVerifyChecker(t *testing.T) {
	tests := []struct {
		name           string
		yamlContent    string
		expectError    bool
		expectedErrMsg string
		expectedLang   Language
		expectedCode   string
		expectedMsg    string
	}{
		{
			name: "UnknownLanguage",
			yamlContent: `language: javascri	
name: wrong-pattern
message: "Checking wrong pattern presence"
category: style
severity: info
pattern: 
description: "Test checker no-pattern"`,
			expectError:    true,
			expectedErrMsg: "unknown language code",
		},
		{
			name: "MissingNameField",
			yamlContent: `language: javascript
message: "Checking wrong pattern presence"
category: style
severity: info
pattern: 
description: "Test checker no-pattern"`,
			expectError:    true,
			expectedErrMsg: "missing necessary field",
		},
		{
			name: "MissingMessageField",
			yamlContent: `language: javascript
name: test-checker
category: style
severity: info
pattern: 
description: "Test checker no-message"`,
			expectError:    true,
			expectedErrMsg: "missing necessary field",
		},
		{
			name: "ValidChecker",
			yamlContent: `language: javascript
name: test-checker
message: "Test message"
category: style
severity: info
pattern: some-pattern
description: "Valid test checker"`,
			expectError:  false,
			expectedCode: "test-checker",
			expectedMsg:  "Test message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var checker Yaml
			err := yaml.Unmarshal([]byte(tt.yamlContent), &checker)
			assert.NoError(t, err)

			lang, code, msg, err := verifyChecker(checker)

			if tt.expectError {
				assert.Error(t, err)
				if tt.expectedErrMsg != "" {
					assert.Contains(t, err.Error(), tt.expectedErrMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotEqual(t, LangUnknown, lang)
				assert.Equal(t, tt.expectedCode, code)
				assert.Equal(t, tt.expectedMsg, msg)
			}
		})
	}
}

func TestInvalidPath(t *testing.T) {
	path := "./testdata/yaml_tests/path_filters/malformed_path.yml"
	_, _, err := ReadFromFile(path)

	assert.EqualError(t, err, "invalid exclude pattern in yaml checker")

	path = "./testdata/yaml_tests/path_filters/malformed_path_include.yml"
	_, _, err = ReadFromFile(path)

	assert.EqualError(t, err, "invalid include pattern in yaml checker")

}

func TestPathFilters(t *testing.T) {
	path := "./testdata/yaml_tests/path_filters/valid_path.yml"

	_, yamlAna, err := ReadFromFile(path)

	assert.NoError(t, err)
	assert.Equal(t, len(yamlAna.PathFilter.ExcludeGlobs), 4)
	assert.Equal(t, len(yamlAna.PathFilter.IncludeGlobs), 1)
}

func TestReadFromBytes(t *testing.T) {
	src :=
		`Name: unmarshal-error`

	_, _, err := ReadFromBytes([]byte(src))

	assert.Error(t, err)

	src = `language: javascript
		name:test
	`
	_, _, err = ReadFromBytes([]byte(src))

	assert.Error(t, err)

}
