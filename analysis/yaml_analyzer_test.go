package analysis

import (
	"path/filepath"
	"testing"

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

func TestParseTargetFile(t *testing.T) {
	t.Run("test if target file is parsed", func(t *testing.T) {
		checker, err := ReadFromFile(filepath.Join("fixtures", "checkers", "test_checker.yml"))
		require.Nil(t, err)
		require.NotNil(t, checker)
		analyzer, err := FromFile(filepath.Join("fixtures", "checkers", "test_checker.test.py"), []Checker{})
		analyzer.YamlCheckers = append(analyzer.YamlCheckers, checker)
		require.Nil(t, err)
	})
}

func TestYamlAnalyzerRunner(t *testing.T) {
	t.Run("test if the yaml analyzer works", func(t *testing.T) {
		testCheckerPath := filepath.Join("fixtures", "checkers", "test_checker.yml")
		testFilePath := filepath.Join("fixtures", "checkers", "test_checker.test.py")

		checker, err := ReadFromFile(testCheckerPath)
		require.Nil(t, err, "should create checker")
		require.NotNil(t, checker, "checker should exist")

		analyzer, err := FromFile(testFilePath, []Checker{})
		require.Nil(t, err, "should create analyzer from test file")
		require.NotNil(t, analyzer, "analyzer should not be nil")

		analyzer.WorkDir = filepath.Join("fixtures", "checkers")
		analyzer.YamlCheckers = append(analyzer.YamlCheckers, checker)

		issues := analyzer.Analyze()
		require.NotNil(t, issues, "issues should be raised")

	})
}