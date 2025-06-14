package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	ana "globstar.dev/analysis"
)

func TestGetAnalysisFunction(t *testing.T) {
	path := "testdata/mock-analysis-function.yml"
	_, yamlAnalyzer, err := ana.ReadFromFile(path)
	analysisFunction := yamlAnalyzer.AnalysisFunction
	assert.NotNil(t, analysisFunction)
	assert.Len(t, analysisFunction.Parameters, 2)
	assert.Len(t, analysisFunction.Parameters["sources"], 1)
	assert.NoError(t, err)

	analysisFuncAnalyzer := GetAnalysisFunction(&yamlAnalyzer)
	assert.Equal(t, analysisFuncAnalyzer.Name, "taint_detector")
}

func TestAnalysisFunction(t *testing.T) {
	path := "testdata/mock-analysis-function.yml"
	passed, err := runTests(path)
	assert.NoError(t, err)
	assert.True(t, passed)
}
