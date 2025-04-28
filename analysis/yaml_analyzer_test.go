package analysis

import (
	"testing"
	"path/filepath"


	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)


func TestCreateYamlChecker(t *testing.T) {
	t.Run("test is yaml checker created", func(t *testing.T) {
		checker, err := ReadFromFile(filepath.Join("fixtures", "checkers", "test_checker.yml"))
		require.Nil(t, err)
		assert.Equal(t, checker.Name(), "test_checker")
		assert.Equal(t, checker.Language(), LangPy)
		require.Nil(t, checker.NodeFilters())
		require.Nil(t, checker.PathFilter())
	})
}