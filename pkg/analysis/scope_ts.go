// scope resolution implementation for JS and TS files
package analysis

import (
	"slices"

	sitter "github.com/smacker/go-tree-sitter"
)

type UnresolvedRef struct {
	id               *sitter.Node
	surroundingScope *Scope
}

type TsScopeBuilder struct {
	ast    *sitter.Node
	source []byte
	// unresolvedRefs is the list of references that could not be resolved thus far in the traversal
	unresolvedRefs []UnresolvedRef
}

func (*TsScopeBuilder) GetLanguage() Language {
	return LangJs
}

var ScopeNodes = []string{
	"statement_block",
	"function_declaration",
	"function_expression",
	"for_statement",
	"for_in_statement",
	"for_of_statement",
	"program",
}

func (*TsScopeBuilder) NodeCreatesScope(node *sitter.Node) bool {
	return slices.Contains(ScopeNodes, node.Type())
}

func (*TsScopeBuilder) DeclaresVariable(node *sitter.Node) bool {
	typ := node.Type()
	// addition of function_declaration and formal_parameters necessary for functional scope handling.
	return typ == "variable_declarator" || typ == "import_clause" || typ == "import_specifier" || typ == "formal_parameters" || typ == "function_declaration"
}

func (ts *TsScopeBuilder) scanDecl(idOrPattern, declarator *sitter.Node, decls []*Variable) []*Variable {
	switch idOrPattern.Type() {
	case "identifier":
		// <name> = ...
		nameStr := idOrPattern.Content(ts.source)
		decls = append(decls, &Variable{
			Kind:     VarKindVariable,
			Name:     nameStr,
			DeclNode: declarator,
		})

	case "object_pattern":
		// { <properties> } = ...
		props := ChildrenOfType(idOrPattern, "shorthand_property_identifier_pattern")
		for _, prop := range props {
			decls = append(decls, &Variable{
				Kind:     VarKindVariable,
				Name:     prop.Content(ts.source),
				DeclNode: declarator,
			})
		}

		pairs := ChildrenOfType(idOrPattern, "pair_pattern")
		for _, pair := range pairs {
			decls = ts.scanDecl(pair, declarator, decls)
		}

		// { realName : <alias> } = ...
		// alias can be an identifier or nested object pattern.
	case "pair_pattern":
		binding := idOrPattern.ChildByFieldName("value")
		decls = ts.scanDecl(binding, declarator, decls)

	case "array_pattern":
		// [ <elements> ] = foo
		childrenIds := ChildrenOfType(idOrPattern, "identifier")
		childrenObjPatterns := ChildrenOfType(idOrPattern, "object_pattern")
		childrenArrayPatterns := ChildrenOfType(idOrPattern, "array_pattern")
		for _, id := range childrenIds {
			decls = append(decls, &Variable{
				Kind:     VarKindVariable,
				Name:     id.Content(ts.source),
				DeclNode: declarator,
			})
		}

		for _, objPattern := range childrenObjPatterns {
			decls = ts.scanDecl(objPattern, declarator, decls)
		}

		for _, arrayPattern := range childrenArrayPatterns {
			decls = ts.scanDecl(arrayPattern, declarator, decls)
		}

		for _, objectPattern := range childrenObjPatterns {
			decls = ts.scanDecl(objectPattern, declarator, decls)
		}
	}

	return decls
}

func (ts *TsScopeBuilder) variableFromImportSpecifier(specifier *sitter.Node) *Variable {
	name := specifier.ChildByFieldName("name")
	if name == nil {
		// skipcq: TCV-001
		return nil
	}

	var Name string
	if specifier.Child(2) != nil {
		// alias (<imported> as <local>)
		local := specifier.Child(2)
		Name = local.Content(ts.source)
	} else {
		// no alias
		Name = name.Content(ts.source)
	}

	return &Variable{
		Kind:     VarKindImport,
		Name:     Name,
		DeclNode: specifier,
	}
}

func (ts *TsScopeBuilder) CollectVariables(node *sitter.Node) []*Variable {
	var declaredVars []*Variable
	switch node.Type() {
	case "variable_declarator":
		lhs := node.ChildByFieldName("name")
		return ts.scanDecl(lhs, node, declaredVars)

	case "function_declaration":
		name := node.ChildByFieldName("name")
		// skipcq: TCV-001
		if name == nil {
			break
		}

		declaredVars = append(declaredVars, &Variable{
			Kind:     VarKindFunction,
			Name:     name.Content(ts.source),
			DeclNode: node,
		})

	case "formal_parameters":
		// TODO

		for i := 0; i < int(node.NamedChildCount()); i++ {
			param := node.NamedChild(i)
			if param == nil {
				continue
			}
			// Handle different parameter types (required, optional, rest, patterns)
			// Simple identifier parameter: function foo(x)
			// Required parameter often wraps identifier: function foo(x: number)
			var identifier *sitter.Node
			if param.Type() == "identifier" {
				identifier = param
			} else if param.Type() == "required_parameter" || param.Type() == "optional_parameter" {
				// Look for pattern which might be identifier or destructuring
				pattern := param.ChildByFieldName("pattern")
				if pattern != nil && pattern.Type() == "identifier" {
					identifier = pattern
				}
				// TODO: Handle destructuring patterns within parameters if needed by calling scanDecl
			} else if param.Type() == "assignment_pattern" {
				// Parameter with default value: function foo(x = 1)
				left := param.ChildByFieldName("left")
				if left != nil && left.Type() == "identifier" {
					identifier = left
				}
				// TODO: Handle destructuring patterns within parameters if needed by calling scanDecl
			}
			// TODO: Handle rest parameter (...)+
			if identifier != nil {
				declaredVars = append(declaredVars, &Variable{
					Kind:     VarKindParameter,
					Name:     identifier.Content(ts.source),
					DeclNode: param, // Use the parameter node itself (or identifier) as DeclNode
				})
			}
			// Add handling for destructuring patterns here if necessary using scanDecl
		}

	case "import_specifier":
		// import { <name> } from ...
		variable := ts.variableFromImportSpecifier(node)
		declaredVars = append(declaredVars, variable)

	case "import_clause":
		// import <default>, { <non_default> } from ...
		defaultImport := FirstChildOfType(node, "identifier")
		if defaultImport != nil {
			declaredVars = append(declaredVars, &Variable{
				Kind:     VarKindImport,
				Name:     defaultImport.Content(ts.source),
				DeclNode: defaultImport,
			})
		}
	}

	return declaredVars
}

func (ts *TsScopeBuilder) OnNodeEnter(node *sitter.Node, scope *Scope) {
	// collect identifier references if one is found
	if node.Type() == "identifier" {
		parent := node.Parent()
		if parent == nil {
			return
		}

		parentType := parent.Type()

		if parentType == "variable_declarator" && parent.ChildByFieldName("name") == node {
			return
		}

		if parentType == "formal_parameters" {
			return
		}

		// binding identifiers in array patterns are not references.
		// e.g. in `const [a, b] = foo;`, `a` and `b` are not references.
		if parentType == "array_pattern" {
			return
		}

		if parentType == "assignment_pattern" && parent.ChildByFieldName("left") == node {
			return
		}

		if parentType == "required_parameter" && parent.ChildByFieldName("pattern") == node {
			return
		}

		// destructured property binding names are *not* references.
		// e.g. in `const { a: b } = foo;`, `a` is not a reference.
		if parentType == "pair_pattern" && parent.ChildByFieldName("key") == node {
			return
		}

		if parentType == "import_clause" || parentType == "import_specifier" {
			return
		}

		// try to resolve this reference to a target variable
		variable := scope.Lookup(node.Content(ts.source))
		if variable == nil {
			unresolved := UnresolvedRef{
				id:               node,
				surroundingScope: scope,
			}

			ts.unresolvedRefs = append(ts.unresolvedRefs, unresolved)
			return
		}

		// If a variable is found, add a reference to it
		ref := &Reference{
			Variable: variable,
			Node:     node,
		}
		variable.Refs = append(variable.Refs, ref)
	}
}

func (ts *TsScopeBuilder) OnNodeExit(node *sitter.Node, scope *Scope) {
	if node.Type() == "program" {
		// At the end, try to resolve all unresolved references
		for _, unresolved := range ts.unresolvedRefs {
			variable := unresolved.surroundingScope.Lookup(
				unresolved.id.Content(ts.source),
			)

			if variable == nil {
				continue
			}

			ref := &Reference{
				Variable: variable,
				Node:     unresolved.id,
			}

			variable.Refs = append(variable.Refs, ref)
		}
	}
}
