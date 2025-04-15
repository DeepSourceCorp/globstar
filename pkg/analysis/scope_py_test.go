package analysis

import (
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

	t.Run("supports function parameters", func(t *testing.T) {
		source := `
			def myFunc(a, b=2, c:int, d:str="Hello"):
				A = otherFunc(a)
				C = b + c
				print(d)
			`
		parsed := parsePyFile(t, source)

		scopeTree := MakeScopeTree(parsed.Language, parsed.Ast, parsed.Source)
		require.NotNil(t, scopeTree)

		globalScope := scopeTree.Root.Children[0]
		require.NotNil(t, globalScope)

		{
			varMyFunc, exists := globalScope.Variables["myFunc"]
			require.NotNil(t, varMyFunc)
			require.True(t, exists)

			assert.Equal(t, VarKindFunction, varMyFunc.Kind)
			myFuncRefs := varMyFunc.Refs
			require.Equal(t, 0, len(myFuncRefs))
		}

		{
			varA, exists := globalScope.Children[0].Variables["a"]
			require.NotNil(t, varA)
			require.True(t, exists)
			assert.Equal(t, VarKindParameter, varA.Kind)
			
			aRefs := varA.Refs
			require.Equal(t, 1, len(aRefs))
			assert.Equal(t, "argument_list", aRefs[0].Node.Parent().Type())
		}

		{
			varB, exists := globalScope.Children[0].Variables["b"]
			require.NotNil(t, varB)
			require.True(t, exists)
			assert.Equal(t, VarKindParameter, varB.Kind)

			bRefs := varB.Refs
			require.Equal(t, 1, len(bRefs))
			assert.Equal(t, "binary_operator", bRefs[0].Node.Parent().Type())
		}
	})

	t.Run("supports with statements", func(t *testing.T) {
		source := `
			with open("file.txt", 'r') as f:
				print(f.read(5))
			`
		parsed := parsePyFile(t, source)

		scopeTree := MakeScopeTree(parsed.Language, parsed.Ast, parsed.Source)
		require.NotNil(t, scopeTree)

		globalScope := scopeTree.Root.Children[0]
		require.NotNil(t, globalScope)

		{
			varF, exists := globalScope.Variables["f"]
			require.NotNil(t, varF)
			require.True(t, exists)

			assert.Equal(t, VarKindVariable, varF.Kind)
			fRefs := varF.Refs
			require.Equal(t, 1, len(fRefs))
			assert.Equal(t, "call", fRefs[0].Node.Parent().Parent().Type())
		}
	})

	t.Run("supports walrus operator", func(t *testing.T) {
		source := `
if (n := random.randint(1, 100)) > 50:
	print("Greater than 50")
		`
		parsed := parsePyFile(t, source)

		scopeTree := MakeScopeTree(parsed.Language, parsed.Ast, parsed.Source)
		require.NotNil(t, scopeTree)

		globalScope := scopeTree.Root.Children[0]
		require.NotNil(t, globalScope)

		varN, exists := globalScope.Children[0].Variables["n"]
		require.NotNil(t, varN)
		require.True(t, exists)
	})

	// for `list_comprehension`, `dictionary_comprehension`, `generator_comprehension`, `set_comprehension`
	t.Run("supports comprehension statements", func(t *testing.T) {
		source := `
a = [x for x in range(10) if x % 2 == 0]

b = {x: x**2 for x in myList if x == 10}
		`
		parsed := parsePyFile(t, source)

		scopeTree := MakeScopeTree(parsed.Language, parsed.Ast, parsed.Source)
		require.NotNil(t, scopeTree)

		globalScope := scopeTree.Root.Children[0]
		require.NotNil(t, globalScope)

		{
			lcScope := globalScope.Children[0]
			require.NotNil(t, lcScope)
			varX, exists := lcScope.Variables["x"]
			require.NotNil(t, varX)
			require.True(t, exists)

			xRefs := varX.Refs
			assert.Equal(t, 2, len(xRefs)) // first in the expression, second in the if-else statement
		}
		
		{
			dcScope := globalScope.Children[1]
			require.NotNil(t, dcScope)
			varX, exists := dcScope.Variables["x"]
			require.NotNil(t, varX)
			require.True(t, exists)

			xRefs := varX.Refs
			assert.Equal(t, 3, len(xRefs))
		}
	})

	t.Run("supports loop statements", func(t *testing.T) {
		source := `
for id, value in enumerate(someList):
	print(id, value)
		`
		parsed := parsePyFile(t, source)

		scopeTree := MakeScopeTree(parsed.Language, parsed.Ast, parsed.Source)
		require.NotNil(t, scopeTree)

		globalScope := scopeTree.Root.Children[0]
		require.NotNil(t, globalScope)

		forLoopScope := globalScope.Children[0]
		require.NotNil(t, forLoopScope)
		{
			varId, exists := globalScope.Variables["id"]
			require.NotNil(t, varId)
			require.True(t, exists)

			idRefs := varId.Refs
			assert.Equal(t, 1, len(idRefs))
		}

		{
			varValue, exists := globalScope.Variables["value"]
			require.NotNil(t, varValue)
			require.True(t, exists)

			valueRefs := varValue.Refs
			assert.Equal(t, 1, len(valueRefs))
		}

	})

	t.Run("supports exception statements", func(t *testing.T) {
		source := `
			try:
				result = 10 / 2
			except ZeroDivisionError as e:
				print(e)
			`
		parsed := parsePyFile(t, source)

		scopeTree := MakeScopeTree(parsed.Language, parsed.Ast, parsed.Source)
		require.NotNil(t, scopeTree)

		globalScope := scopeTree.Root.Children[0]
		require.NotNil(t, globalScope)

		varE, exists := globalScope.Variables["e"]
		require.NotNil(t, varE)
		require.True(t, exists)

		assert.Equal(t, VarKindError, varE.Kind)
		eRefs := varE.Refs
		require.Equal(t, 1, len(eRefs))
		assert.Equal(t, "call", eRefs[0].Node.Parent().Parent().Type())
	})

	t.Run("supports classes", func(t *testing.T) {
		source := `
			class MyClass:
				def __init__(self, name):
					self.name = name

				def print_name(self):
					print(self.name)
			`
		parsed := parsePyFile(t, source)

		scopeTree := MakeScopeTree(parsed.Language, parsed.Ast, parsed.Source)
		require.NotNil(t, scopeTree)

		globalScope := scopeTree.Root.Children[0]
		require.NotNil(t, globalScope)

		{
			varClass, exists := globalScope.Variables["MyClass"]
			require.NotNil(t, varClass)
			require.True(t, exists)
			assert.Equal(t, VarKindClass, varClass.Kind)
		}

	})

}
