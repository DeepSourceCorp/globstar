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

// TODO: figure out it does not work
// func TestYamlAnalyzerRunner(t *testing.T) {
// 	t.Run("test if the yaml analyzer works", func(t *testing.T) {
// 		checker, err := ReadFromFile(filepath.Join("fixtures", "checkers", "test_checker.yml"))
// 		require.Nil(t, err)
// 		require.NotNil(t, checker)

// 		fmt.Printf("Checker loaded: %s\n", checker.Name())
// 		fmt.Println("Checker patterns: ", checker.Patterns())
// 		analyzer, err := FromFile(filepath.Join("fixtures", "checkers", "test_checker.test.py"), []Checker{})
// 		require.Nil(t, err)
		
// 		cwd, err := os.Getwd()
// 		require.Nil(t, err)
// 		analyzer.WorkDir = filepath.Join(cwd, "fixtures", "checkers")
// 		analyzer.YamlCheckers = append(analyzer.YamlCheckers, checker)
// 		fmt.Println(analyzer)
// 		issues := analyzer.Analyze()
// 		fmt.Println(analyzer)

// 		fmt.Println(issues)

// 		assert.Equal(t, 2, len(issues))
// 		issue := issues[0]
// 		require.Nil(t, err)
// 		require.NotNil(t, issue)
// 		assert.Equal(t, issue.Category, CategorySecurity)
// 		assert.Equal(t, issue.Severity, SeverityWarning)

// 	})
// }