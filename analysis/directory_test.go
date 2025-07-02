package analysis

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDirectoryPopulation(t *testing.T) {
	checker := `name: "run_taint_analysis"
language: javascript
category: security
severity: high
message: "This is just a mock checker"
analysisFunction:
  name: taint
  parameters:
    sources: 
        - |
         (call_expression
            function: (identifier) @sourceName 
              (#eq? @sourceName "getUserInput"))
    sinks: 
        - |
         (call_expression
            function: (identifier) @sinkName
              (#eq? @sinkName "perform_db_operation"))
pattern: |
  (call_expression)
description: "Runs a taint analysis on the provided function and its parameters."`

	_, _, err := ReadFromBytes([]byte(checker))
	assert.NoError(t, err)
	assert.NotNil(t, AnalysisFuncDirectory.Pool["taint"])
}
