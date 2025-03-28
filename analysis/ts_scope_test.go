package analysis

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func parseFile(t *testing.T, source string) *ParseResult {
	parsed, err := Parse("file.ts", []byte(source), LangJs, LangJs.Grammar())
	require.NoError(t, err)
	require.NotNil(t, parsed)
	return parsed
}

func Test_BuildScopeTree(t *testing.T) {
	t.Run("is able to resolve references", func(t *testing.T) {
		source := `
			let x = 1
			{
				let y = x
			}`
		parsed := parseFile(t, source)

		scopeTree := MakeScopeTree(parsed.Language, parsed.Ast, parsed.Source)
		require.NotNil(t, scopeTree)

		varX, exists := scopeTree.Root.Variables["x"]
		require.True(t, exists)
		require.NotNil(t, varX)

		varY, exists := scopeTree.Root.Children[0].Variables["y"]
		require.True(t, exists)
		require.NotNil(t, varY)
		require.Equal(t, VarKindVariable, varY.Kind)

		assert.Equal(t, 1, len(varX.Refs))
		xRef := varX.Refs[0]
		assert.Equal(t, "x", xRef.Variable.Name)
		require.Equal(t, VarKindVariable, varY.Kind)
	})

	t.Run("supports import statements", func(t *testing.T) {
		source := `
			import { extname } from 'path'
			{
					let { extname = 1 } = null //  does NOT count as a reference
			}

			let { x = extname } = null // counts as a reference

			{
				extname('file.txt') // counts as a reference
				let { extname } = null //  does NOT count as a reference
			}

			import { readFile as r } from 'file'
			r('file.txt')
			function f(r = x) {} // NOT a reference
		  `
		parsed := parseFile(t, source)

		scopeTree := MakeScopeTree(parsed.Language, parsed.Ast, parsed.Source)
		require.NotNil(t, scopeTree)

		{
			varR, exists := scopeTree.Root.Variables["r"]
			require.True(t, exists)
			require.NotNil(t, varR)

			assert.Equal(t, VarKindImport, varR.Kind)

			rRefs := varR.Refs
			require.Equal(t, 1, len(rRefs))
			assert.Equal(t, "call_expression", rRefs[0].Node.Parent().Type())
		}

		{
			varExtname, exists := scopeTree.Root.Variables["extname"]
			require.True(t, exists)
			require.NotNil(t, varExtname)

			assert.Equal(t, VarKindImport, varExtname.Kind)

			extnameRefs := varExtname.Refs
			require.Equal(t, 2, len(extnameRefs))
			assert.Equal(t, "object_assignment_pattern", extnameRefs[0].Node.Parent().Type())
			assert.Equal(t, "call_expression", extnameRefs[1].Node.Parent().Type())
		}
	})
}
