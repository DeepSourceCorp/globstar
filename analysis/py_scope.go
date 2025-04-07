package analysis

import (
	"slices"

	sitter "github.com/smacker/go-tree-sitter"
)

// NOTE: should this struct type be moved to another file?
/*
type UnresolvedRef struct {
	id               *sitter.Node
	surroundingScope *Scope
}
*/

type PyScopeBuilder struct {
	ast    *sitter.Node
	source []byte
	// list of references that could not be resolved thus far
	unresolvedRefs []UnresolvedRef
}

func (py *PyScopeBuilder) GetLanguage() Language {
	return LangPy
}

var PyScopeNodes = []string{
	"module",
	"function_definition",
	"class_definition",
	"for_statement",
	"while_statement",
	"if_statement",
	"elif_clause",
	"else_clause",
	"with_statement",
	"try_statement",
	"except_clause",
	"list_comprehension",
	"dictionary_comprehension",
	"lambda",
}

func (py *PyScopeBuilder) NodeCreatesScope(node *sitter.Node) bool {
	return slices.Contains(PyScopeNodes, node.Type())
}

func (py *PyScopeBuilder) DeclaresVariable(node *sitter.Node) bool {
	typ := node.Type()
	return typ == "assignment" || typ == "dotted_name" || typ == "aliased_import" || typ == "with_statement" || typ == "parameters" || typ == "function_definition"
}

func (py *PyScopeBuilder) scanDecl(idOrPattern, declarator *sitter.Node, decls []*Variable) []*Variable {
	switch idOrPattern.Type() {
	case "identifier":
		// TODO: implement for <name1> = <name2> = ...
		// <name> = ...
		nameStr := idOrPattern.Content(py.source)
		decls = append(decls, &Variable{
			Kind:     VarKindVariable,
			Name:     nameStr,
			DeclNode: declarator,
		})

	case "pattern_list", "tuple_pattern", "list_pattern":
		// <name1>, <name2> = ..., ...
		// (<name1>, <name2>) = ..., ...
		// [<name1>, <name2>] = ..., ...
		ids := ChildrenOfType(idOrPattern, "identifier")
		for _, id := range ids {
			decls = append(decls, &Variable{
				Kind:     VarKindVariable,
				Name:     id.Content(py.source),
				DeclNode: declarator,
			})
		}

		// <name1>, *<name2> = ..., ..., ...
		// also applicable to tuple_pattern & list_pattern
		splats := ChildrenOfType(idOrPattern, "list_splat_pattern")
		for _, splat := range splats {
			splatIdNode := splat.Child(0)
			if splatIdNode.Type() == "identifier" {
				decls = append(decls, &Variable{
					Kind:     VarKindVariable,
					Name:     splatIdNode.Content(py.source),
					DeclNode: declarator,
				})
			}
		}
	}

	return decls
}

func (py *PyScopeBuilder) CollectVariables(node *sitter.Node) []*Variable {
	var declaredVars []*Variable
	switch node.Type() {
	case "assignment":
		lhs := node.ChildByFieldName("left")
		return py.scanDecl(lhs, node, declaredVars)

	case "function_definition":
		name := node.ChildByFieldName("name")
		// skipcq: TCV-001
		if name == nil {
			break
		}

		declaredVars = append(declaredVars, &Variable{
			Kind:     VarKindFunction,
			Name:     name.Content(py.source),
			DeclNode: node,
		})

	case "parameters":
		declaredVars = py.variableFromFunctionParams(node, declaredVars)

	case "aliased_import":
		// import <name> as <alias>
		aliasName := node.ChildByFieldName("name")
		if aliasName != nil {
			declaredVars = append(declaredVars, &Variable{
				Kind:     VarKindImport,
				Name:     aliasName.Content(py.source),
				DeclNode: aliasName,
			})
		}

	case "dotted_name":
		// import <name>
		defaultImport := FirstChildOfType(node, "identifier")
		if defaultImport != nil {
			declaredVars = append(declaredVars, &Variable{
				Kind:     VarKindImport,
				Name:     defaultImport.Content(py.source),
				DeclNode: defaultImport,
			})
		}

	case "with_statement":
		clause := FirstChildOfType(node, "with_clause")
		item := FirstChildOfType(clause, "with_item")

		value := item.ChildByFieldName("value")
		alias := value.ChildByFieldName("alias")
		if alias != nil {
			id := FirstChildOfType(alias, "identifier")
			if id != nil {
				declaredVars = append(declaredVars, &Variable{
					Kind:     VarKindVariable,
					Name:     id.Content(py.source),
					DeclNode: node,
				})
			}
		}
	}

	return declaredVars
}

func (py *PyScopeBuilder) OnNodeEnter(node *sitter.Node, scope *Scope) {
	// collected identifier references if found
	if node.Type() == "identifier" || node.Type() == "list_splat_pattern" {
		parent := node.Parent()
		if parent == nil {
			return
		}

		parentType := parent.Type()

		if parentType == "assignment" && parent.ChildByFieldName("left") == node {
			return
		}

		if parentType == "parameters" {
			return
		}

		if parentType == "default_parameter" && parent.ChildByFieldName("name") == node {
			return
		}

		if parentType == "pattern_list" || parentType == "tuple_pattern" || parentType == "list_pattern" {
			return
		}

		// module names in from <module_name> import ... are not references
		// names in import <name> as <alias> are not references
		if parentType == "dotted_name" && !isModuleName(parent) && parent.Parent().Type() != "aliased_import" {
			return
		}

		if parentType == "aliased_import" {
			return
		}

		if parentType == "function_definition" {
			return
		}

		if parentType == "parameters" || parentType == "default_parameter" || parentType == "typed_default_parameter" {
			return
		}

		if parentType == "as_pattern_target" {
			return
		}

		// resolve this reference
		variable := scope.Lookup(node.Content(py.source))
		if variable == nil {
			unresolved := UnresolvedRef{
				id:               node,
				surroundingScope: scope,
			}

			py.unresolvedRefs = append(py.unresolvedRefs, unresolved)
			return
		}

		ref := &Reference{
			Variable: variable,
			Node:     node,
		}

		variable.Refs = append(variable.Refs, ref)

	}
}

func (py *PyScopeBuilder) OnNodeExit(node *sitter.Node, scope *Scope) {
	if node.Type() == "module" {
		for _, unresolved := range py.unresolvedRefs {
			variable := unresolved.surroundingScope.Lookup(unresolved.id.Content(py.source))

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

func isModuleName(dottedNameNode *sitter.Node) bool {
	if dottedNameNode.Type() != "dotted_name" {
		return false
	}

	importNode := dottedNameNode.Parent()
	if importNode.Type() != "import_from_statement" || importNode == nil {
		return false
	}

	moduleNameChildren := ChildrenWithFieldName(importNode, "module_name")

	return slices.Contains(moduleNameChildren, dottedNameNode)
}

func (py *PyScopeBuilder) variableFromFunctionParams(node *sitter.Node, decls []*Variable) []*Variable {
	childrenCount := node.NamedChildCount()
	for i := 0; i < int(childrenCount); i++ {
		param := node.NamedChild(i)

		if param == nil {
			continue
		}

		// handle the parameter types:
		// identifier, typed_parameter, default_parameter, typed_default_parameter
		if param.Type() == "identifier" {
			decls = append(decls, &Variable{
				Kind:     VarKindParameter,
				Name:     param.Content(py.source),
				DeclNode: param,
			})
		} else if param.Type() == "typed_parameter" || param.Type() == "list_splat_pattern" || param.Type() == "dictionary_splat_pattern" {
			idNode := FirstChildOfType(param, "identifier")
			if idNode != nil {
				decls = append(decls, &Variable{
					Kind:     VarKindParameter,
					Name:     idNode.Content(py.source),
					DeclNode: param,
				})
			}
		} else if param.Type() == "default_parameter" || param.Type() == "typed_default_parameter" {
			name := ChildWithFieldName(param, "name")
			if name != nil {
				if name.Type() == "identifier" {
					decls = append(decls, &Variable{
						Kind:     VarKindParameter,
						Name:     name.Content(py.source),
						DeclNode: param,
					})
				} else if name.Type() == "tuple_pattern" {
					childrenIds := ChildrenOfType(name, "identifier")
					childrenListSplat := ChildrenOfType(name, "list_splat_pattern")

					for _, id := range childrenIds {
						decls = append(decls, &Variable{
							Kind:     VarKindParameter,
							Name:     id.Content(py.source),
							DeclNode: param,
						})
					}

					for _, listSplatPat := range childrenListSplat {
						splatId := FirstChildOfType(listSplatPat, "identifier")
						if splatId != nil {
							decls = append(decls, &Variable{
								Kind:     VarKindParameter,
								Name:     listSplatPat.Content(py.source),
								DeclNode: param,
							})
						}
					}
				}
			}
		} else if param.Type() == "tuple_pattern" {
			childrenIds := ChildrenOfType(param, "identifier")
			childrenListSplat := ChildrenOfType(param, "list_splat_pattern")

			for _, id := range childrenIds {
				decls = append(decls, &Variable{
					Kind:     VarKindParameter,
					Name:     id.Content(py.source),
					DeclNode: param,
				})
			}

			for _, listSplatPat := range childrenListSplat {
				splatId := FirstChildOfType(listSplatPat, "identifier")
				if splatId != nil {
					decls = append(decls, &Variable{
						Kind:     VarKindParameter,
						Name:     listSplatPat.Content(py.source),
						DeclNode: param,
					})
				}
			}
		}
	}

	return decls
}
