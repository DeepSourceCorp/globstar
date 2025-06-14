package analysis

import (
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
}

func TestNodeFilters(t *testing.T) {
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

func TestAnalysisFunction(t *testing.T) {
	path := "./testdata/mock-analysis-function.yml"
	ana, yamlAna, err := ReadFromFile(path)
	require.NoError(t, err, "Failed to read YAML data")
	assert.Equal(t, ana.Name, "run_taint_analysis")
	assert.Equal(t, yamlAna.AnalysisFunction.Name, "taint")
	assert.Equal(t, len(yamlAna.AnalysisFunction.Parameters["sources"]), 1)
	assert.Equal(t, len(yamlAna.AnalysisFunction.Parameters["sinks"]), 1)
	assert.Equal(t, yamlAna.AnalysisFunction.Name, "taint")

}
