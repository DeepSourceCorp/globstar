package analysis

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitializeAnalysisFunction(t *testing.T) {

	fn := InitializeAnalysisFunction(AnalysisFunction{
		Name: "taint",
		Parameters: map[string][]string{
			"sources": {"string"},
			"sinks":   {"string"},
		},
	})

	inbuiltTaintAnalyzer, err := fn.Run([]string{"source1", "source2"}, []string{"sink1", "sink2"})
	assert.NoError(t, err)

	assert.Equal(t, inbuiltTaintAnalyzer.Name, "taint_analyzer")

}

func TestPopulationOfAnalysisFuncitonRegistry(t *testing.T) {
	_ = InitializeAnalysisFunction(AnalysisFunction{
		Name: "taint",
		Parameters: map[string][]string{
			"sources": {"string"},
			"sinks":   {"string"},
		},
	})

	assert.Equal(t, len(RegisteredAnalysisFunctions), 1)

}
