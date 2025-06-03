package analysis

// import (
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// )

// func parseFile(t *testing.T, source string) *ParseResult {
// 	parsed, err := Parse("file.ts", []byte(source), LangJs, LangJs.Grammar())
// 	require.NoError(t, err)
// 	require.NotNil(t, parsed)
// 	return parsed
// }

// func Test_BuildScopeTree(t *testing.T) {
// 	t.Run("is able to resolve references", func(t *testing.T) {
// 		source := `
// 			let x = 1
// 			{
// 				let y = x
// 			}`
// 		parsed := parseFile(t, source)

// 		scopeTree := MakeScopeTree(parsed.Language, parsed.Ast, parsed.Source)
// 		require.NotNil(t, scopeTree)
// 		globalScope := scopeTree.Root.Children[0]
// 		varX, exists := globalScope.Variables["x"]
// 		require.True(t, exists)
// 		require.NotNil(t, varX)

// 		varY, exists := globalScope.Children[0].Variables["y"]
// 		require.True(t, exists)
// 		require.NotNil(t, varY)
// 		require.Equal(t, VarKindVariable, varY.Kind)

// 		assert.Equal(t, 1, len(varX.Refs))
// 		xRef := varX.Refs[0]
// 		assert.Equal(t, "x", xRef.Variable.Name)
// 		require.Equal(t, VarKindVariable, varY.Kind)
// 	})

// 	t.Run("supports import statements", func(t *testing.T) {
// 		source := `
// 			import { extname } from 'path'
// 			{
// 					let { extname = 1 } = null //  does NOT count as a reference
// 			}

// 			let { x = extname } = null // counts as a reference

// 			{
// 				extname('file.txt') // counts as a reference
// 				let { extname } = null //  does NOT count as a reference
// 			}

// 			import { readFile as r } from 'file'
// 			r('file.txt')
// 			function f(r = x) {} // NOT a reference
// 		  `
// 		parsed := parseFile(t, source)

// 		scopeTree := MakeScopeTree(parsed.Language, parsed.Ast, parsed.Source)
// 		require.NotNil(t, scopeTree)
// 		globalScope := scopeTree.Root.Children[0]
// 		{
// 			varR, exists := globalScope.Variables["r"]
// 			require.True(t, exists)
// 			require.NotNil(t, varR)

// 			assert.Equal(t, VarKindImport, varR.Kind)

// 			rRefs := varR.Refs
// 			require.Equal(t, 1, len(rRefs))
// 			assert.Equal(t, "call_expression", rRefs[0].Node.Parent().Type())
// 		}

// 		{
// 			varExtname, exists := globalScope.Variables["extname"]
// 			require.True(t, exists)
// 			require.NotNil(t, varExtname)

// 			assert.Equal(t, VarKindImport, varExtname.Kind)

// 			extnameRefs := varExtname.Refs
// 			require.Equal(t, 2, len(extnameRefs))
// 			assert.Equal(t, "object_assignment_pattern", extnameRefs[0].Node.Parent().Type())
// 			assert.Equal(t, "call_expression", extnameRefs[1].Node.Parent().Type())
// 		}
// 	})

// 	t.Run("handles function declaration with parameters", func(t *testing.T) {
// 		source := `
// 		function greet(name, age = 18) {
// 			let greeting = "Hello";
// 			return greeting + " " + name;
// 		}
// 		greet("Alice")
// 		`

// 		parsed := parseFile(t, source)
// 		require.NotNil(t, parsed)
// 		scopeTree := MakeScopeTree(parsed.Language, parsed.Ast, parsed.Source)
// 		globalScope := scopeTree.Root.Children[0]
// 		// Checking function declaration
// 		funcVar := globalScope.Lookup("greet")
// 		require.NotNil(t, funcVar)
// 		funcVariable, exists := globalScope.Variables["greet"] // tagged as an Identifier
// 		require.True(t, exists)
// 		require.NotNil(t, funcVariable)

// 		funcScope := scopeTree.GetScope(funcVar.DeclNode)
// 		require.NotNil(t, funcScope)

// 		nameVar, exists := funcScope.Variables["name"]
// 		require.True(t, exists)
// 		require.Equal(t, VarKindParameter, nameVar.Kind)

// 		ageVar, exists := funcScope.Variables["age"]
// 		require.True(t, exists)
// 		require.Equal(t, VarKindParameter, ageVar.Kind)

// 		// existence of function body

// 		bodyScope := funcScope.Children[0]
// 		require.NotNil(t, bodyScope)

// 		greetingVar, exists := bodyScope.Variables["greeting"]
// 		require.True(t, exists)
// 		require.Equal(t, VarKindVariable, greetingVar.Kind)
// 	})
// }
