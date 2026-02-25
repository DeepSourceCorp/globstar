package checkers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestYamlAnalyzers(t *testing.T) {
	path := "./testdata"

	passed, err := RunYamlAnalyzers(path)
	assert.NoError(t, err)
	assert.True(t, passed)
}
