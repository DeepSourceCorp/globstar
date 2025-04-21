package analysis

import (
	// "fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func parsePyFile(t *testing.T, source string) *ParseResult {
	parsed, err := Parse("file.py", []byte(source), LangPy, LangPy.Grammar())
	require.NoError(t, err)
	require.NotNil(t, parsed)
	return parsed
}

func Test_PyBuildScopeTree(t *testing.T) {
	t.Run("is able to resolve references", func(t *testing.T) {
		source := `
			x = 1
			if True:
				y = x
			z = x`
		parsed := parsePyFile(t, source)

		scopeTree := MakeScopeTree(parsed.Language, parsed.Ast, parsed.Source)
		require.NotNil(t, scopeTree)

		globalScope := scopeTree.Root.Children[0]
		require.NotNil(t, globalScope)

		varX, exists := globalScope.Variables["x"]
		require.True(t, exists)
		require.NotNil(t, varX)

		varY, exists := globalScope.Children[0].Variables["y"]
		require.True(t, exists)
		require.NotNil(t, varY)
		require.Equal(t, VarKindVariable, varY.Kind)

		assert.Equal(t, 2, len(varX.Refs))
		xRef := varX.Refs[0]
		assert.Equal(t, "x", xRef.Variable.Name)
		require.Equal(t, VarKindVariable, varY.Kind)

	})

	t.Run("supports import statements", func(t *testing.T) {
		source := `
			import os

			os.system("cat file.txt")

			from csv import read

			if True:
				f = read(file.csv)
			`
		parsed := parsePyFile(t, source)

		scopeTree := MakeScopeTree(parsed.Language, parsed.Ast, parsed.Source)
		require.NotNil(t, scopeTree)

		globalScope := scopeTree.Root.Children[0]
		require.NotNil(t, globalScope)

		{
			varOs, exists := globalScope.Variables["os"]
			require.NotNil(t, varOs)
			require.True(t, exists)

			assert.Equal(t, VarKindImport, varOs.Kind)

			osRefs := varOs.Refs
			require.Equal(t, 1, len(osRefs))
			assert.Equal(t, "attribute", osRefs[0].Node.Parent().Type())
		}

		{
			varRead, exists := globalScope.Variables["read"]
			require.True(t, exists)
			require.NotNil(t, varRead)
			assert.Equal(t, VarKindImport, varRead.Kind)

			varF, exists := globalScope.Children[0].Variables["f"]
			require.True(t, exists)
			require.NotNil(t, varF)
			assert.Equal(t, VarKindVariable, varF.Kind)

			readRefs := varRead.Refs
			require.Equal(t, 1, len(readRefs))
			assert.Equal(t, "call", readRefs[0].Node.Parent().Type())
		}

	})
}
