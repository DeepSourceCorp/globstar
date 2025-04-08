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
		globalScope := scopeTree.Root.Children[0]
		require.NotNil(t, globalScope)
		varX, exists := globalScope.Variables["x"]
		require.True(t, exists)
		require.NotNil(t, varX)

		varY, exists := globalScope.Children[0].Variables["y"]
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

		globalScope := scopeTree.Root.Children[0]
		{
			require.NotNil(t, globalScope)
			varR, exists := globalScope.Variables["r"]
			require.True(t, exists)
			require.NotNil(t, varR)

			assert.Equal(t, VarKindImport, varR.Kind)

			rRefs := varR.Refs
			require.Equal(t, 1, len(rRefs))
			assert.Equal(t, "call_expression", rRefs[0].Node.Parent().Type())
		}

		{
			varExtname, exists := globalScope.Variables["extname"]
			require.True(t, exists)
			require.NotNil(t, varExtname)

			assert.Equal(t, VarKindImport, varExtname.Kind)

			extnameRefs := varExtname.Refs
			require.Equal(t, 2, len(extnameRefs))
			assert.Equal(t, "object_assignment_pattern", extnameRefs[0].Node.Parent().Type())
			assert.Equal(t, "call_expression", extnameRefs[1].Node.Parent().Type())
		}
	})
	t.Run("handles function declaration with parameters", func(t *testing.T) {
		source := `
		function greet(name, age = 18) {
			let greeting = "Hello";
			return greeting + " " + name;	
		}
		greet("Alice")
		`

		parsed := parseFile(t, source)
		require.NotNil(t, parsed)
		scopeTree := MakeScopeTree(parsed.Language, parsed.Ast, parsed.Source)
		require.NotNil(t, scopeTree)
		// Checking function declaration
		globalScope := scopeTree.Root.Children[0]
		funcVar := globalScope.Lookup("greet")
		require.NotNil(t, funcVar)
		funcVariable, exists := globalScope.Variables["greet"] // tagged as an Identifier
		require.True(t, exists)
		require.NotNil(t, funcVariable)

		funcScope := scopeTree.GetScope(funcVar.DeclNode)
		require.NotNil(t, funcScope)

		nameVar, exists := funcScope.Variables["name"]
		require.True(t, exists)
		require.Equal(t, VarKindParameter, nameVar.Kind)

		ageVar, exists := funcScope.Variables["age"]
		require.True(t, exists)
		require.Equal(t, VarKindParameter, ageVar.Kind)

		// existence of function body

		bodyScope := funcScope.Children[0]
		require.NotNil(t, bodyScope)

		greetingVar, exists := bodyScope.Variables["greeting"]
		require.True(t, exists)
		require.Equal(t, VarKindVariable, greetingVar.Kind)
	})

}

func TestExportHandling(t *testing.T) {
	tests := []struct {
		name   string
		source string
		checks func(t *testing.T, scopeTree *ScopeTree)
	}{
		// {
		// 	name: "named_exports",
		// 	source: `
		//         const foo = 123;
		//         const bar = 456;
		//         export { foo, bar as baz };
		//     `,
		// 	checks: func(t *testing.T, scopeTree *ScopeTree) {
		// 		globalScope := scopeTree.Root.Children[0]

		// 		fooVar := globalScope.Lookup("foo")
		// 		require.NotNil(t, fooVar, "foo should be defined")
		// 		assert.True(t, fooVar.Exported, "foo should be marked as exported")

		// 		barVar := globalScope.Lookup("bar")
		// 		require.NotNil(t, barVar, "bar should be defined")
		// 		assert.True(t, barVar.Exported, "bar should be marked as exported")
		// 	},
		// },
		{
			name: "direct_export_declaration",
			source: `
		        export const x = 10;
		        // export function hello() {}
		        // export class MyClass {}
		    `,
			checks: func(t *testing.T, scopeTree *ScopeTree) {
				globalScope := scopeTree.Root.Children[0]

				xVar := globalScope.Lookup("x")
				require.NotNil(t, xVar, "x should be defined")
				assert.True(t, xVar.Exported, "x should be marked as exported")
				assert.Equal(t, VarKindVariable, xVar.Kind)

				helloVar := globalScope.Lookup("hello")
				require.NotNil(t, helloVar, "hello should be defined")
				assert.True(t, helloVar.Exported, "hello should be marked as exported")
				assert.Equal(t, VarKindFunction, helloVar.Kind)

				classVar := globalScope.Lookup("MyClass")
				require.NotNil(t, classVar, "MyClass should be defined")
				assert.True(t, classVar.Exported, "MyClass should be marked as exported")
				assert.Equal(t, VarKindClass, classVar.Kind)
			},
		},
		// {
		// 	name: "default_export",
		// 	source: `
		//         const value = 42;
		//         export default value;
		//     `,
		// 	checks: func(t *testing.T, scopeTree *ScopeTree) {
		// 		globalScope := scopeTree.Root.Children[0]

		// 		valueVar := globalScope.Lookup("value")
		// 		require.NotNil(t, valueVar, "value should be defined")
		// 		assert.True(t, valueVar.Exported, "value should be marked as exported")
		// 	},
		// },
		// {
		// 	name: "mixed_exports",
		// 	source: `
		//         const a = 1, b = 2;
		//         function helper() {}
		//         export { a };
		//         export const c = 3;
		//         export default helper;
		//     `,
		// 	checks: func(t *testing.T, scopeTree *ScopeTree) {
		// 		globalScope := scopeTree.Root.Children[0]

		// 		aVar := globalScope.Lookup("a")
		// 		require.NotNil(t, aVar, "a should be defined")
		// 		assert.True(t, aVar.Exported, "a should be marked as exported")

		// 		bVar := globalScope.Lookup("b")
		// 		require.NotNil(t, bVar, "b should be defined")
		// 		assert.False(t, bVar.Exported, "b should not be marked as exported")

		// 		cVar := globalScope.Lookup("c")
		// 		require.NotNil(t, cVar, "c should be defined")
		// 		assert.True(t, cVar.Exported, "c should be marked as exported")

		// 		helperVar := globalScope.Lookup("helper")
		// 		require.NotNil(t, helperVar, "helper should be defined")
		// 		assert.True(t, helperVar.Exported, "helper should be marked as exported")
		// 	},
		// },
		// {
		// 	name: "export_from_inner_scope",
		// 	source: `
		//         const outer = 1;
		//         {
		//             const inner = 2;
		//             export { inner };
		//         }
		//         export { outer };
		//     `,
		// 	checks: func(t *testing.T, scopeTree *ScopeTree) {
		// 		globalScope := scopeTree.Root.Children[0]

		// 		outerVar := globalScope.Lookup("outer")
		// 		t.Log(outerVar)
		// 		require.NotNil(t, outerVar, "outer should be defined")
		// 		assert.True(t, outerVar.Exported, "outer should be marked as exported")

		// 		innerVar := globalScope.Children[0].Lookup("inner")
		// 		require.NotNil(t, innerVar, "inner should be defined")
		// 		assert.True(t, innerVar.Exported, "inner should be marked as exported")
		// 	},
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed := parseFile(t, tt.source)
			require.NotNil(t, parsed)

			scopeTree := MakeScopeTree(parsed.Language, parsed.Ast, parsed.Source)
			require.NotNil(t, scopeTree)

			tt.checks(t, scopeTree)
		})
	}
}
