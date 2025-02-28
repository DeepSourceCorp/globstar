// A language agnostic interface for scope handling which
// also handles forward declarations and references (e.g: hoisting).
// BUT, references aren't tracked across files in a language like Golang or C++ (macros/extern/using namespace)

package analysis

import sitter "github.com/smacker/go-tree-sitter"

// Reference represents a variable reference inside a source file
// Cross-file references like those in Golang and C++ (macros/extern) are NOT supported,
// so this shouldn't be used for checkers like "unused-variable", but is safe to use for checkers like
// "unused-import"
type Reference struct {
	// IsWriteRef determines if this reference is a write reference.
	// For write refs, only the expression being assigned is stored.
	// i.e: for `a = 3`, this list will store the `3` node, not the assignment node
	IsWriteRef bool
	// Variable stores the variable being referenced
	Variable *Variable
	// Node stores the node that references the variable
	Node *sitter.Node
}

type VarKind int32

const (
	VarKindError VarKind = iota
	VarKindImport
	VarKindFunction
	VarKindVariable
	VarKindParameter
)

type Variable struct {
	Kind VarKind
	// Stores the name of the variable
	Name string
	// DeclNode is the AST node that declares this variable
	DeclNode *sitter.Node
	// Refs is a list of references to this variable throughout the file
	Refs []*Reference
}

// ScopeBuilder is an interface that has to be implemented
// once for every supported language.
// Languages that don't implement a `ScopeBuilder` can still have checkers, just
// not any that require scope resolution.
type ScopeBuilder interface {
	GetLanguage() Language
	// NodeCreatesScope returns true if the node introduces a new scope
	// into the scope tree
	NodeCreatesScope(node *sitter.Node) bool
	// DeclaresVariable determines if we can extract new variables out of this AST node
	DeclaresVariable(node *sitter.Node) bool
	// CollectVariables extracts variables from the node and adds them to the scope
	CollectVariables(node *sitter.Node) []*Variable
	// OnNodeEnter is called when the scope builder enters a node
	// for the first time, and hasn't scanned its children decls just yet
	// can be used to handle language specific scoping rules, if any
	// If `node` is smth like a block statement, `currentScope` corresponds
	// to the scope introduced by the block statement.
	OnNodeEnter(node *sitter.Node, currentScope *Scope)
	// OnNodeExit is called when the scope builder exits a node
	// can be used to handle language specific scoping rules, if any
	// If `node` is smth like a block statement, `currentScope` corresponds
	// to the scope introduced by the block statement.
	OnNodeExit(node *sitter.Node, currentScope *Scope)
}

type Scope struct {
	// AstNode is the AST node that introduces this scope into the scope tree
	AstNode *sitter.Node
	// Variables is a map of variable name to an object representing it
	Variables map[string]*Variable
	// Upper is the parent scope of this scope
	Upper *Scope
	// Children is a list of scopes that are children of this scope
	Children []*Scope
}

func NewScope(upper *Scope) *Scope {
	return &Scope{
		Variables: map[string]*Variable{},
		Upper:     upper,
	}
}

// Lookup searches for a variable in the current scope and its parents
func (s *Scope) Lookup(name string) *Variable {
	if v, exists := s.Variables[name]; exists {
		return v
	}

	if s.Upper != nil {
		return s.Upper.Lookup(name)
	}

	return nil
}

type ScopeTree struct {
	Language Language
	// ScopeOfNode maps every scope-having node to its corresponding scope.
	// E.g: a block statement is mapped to the scope it introduces.
	ScopeOfNode map[*sitter.Node]*Scope
	// Root is the top-level scope in the program,
	// usually associated with the `program` or `module` node
	Root *Scope
}

// BuildScopeTree constructs a scope tree from the AST for a program
func BuildScopeTree(builder ScopeBuilder, ast *sitter.Node, source []byte) *ScopeTree {
	root := NewScope(nil)
	root.AstNode = ast

	scopeOfNode := make(map[*sitter.Node]*Scope)
	buildScopeTree(builder, source, ast, root, scopeOfNode)

	return &ScopeTree{
		Language:    builder.GetLanguage(),
		ScopeOfNode: scopeOfNode,
		Root:        root,
	}
}

func buildScopeTree(
	builder ScopeBuilder,
	source []byte,
	node *sitter.Node,
	scope *Scope,
	scopeOfNode map[*sitter.Node]*Scope,
) *Scope {
	builder.OnNodeEnter(node, scope)
	defer builder.OnNodeExit(node, scope)

	if builder.DeclaresVariable(node) {
		decls := builder.CollectVariables(node)
		for _, decl := range decls {
			scope.Variables[decl.Name] = decl
		}
	}

	nextScope := scope
	if builder.NodeCreatesScope(node) {
		nextScope = NewScope(scope)
		scopeOfNode[node] = nextScope

		if scope != nil {
			scope.Children = append(scope.Children, nextScope)
		} else {
			scope = nextScope // root
		}
	}

	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		buildScopeTree(builder, source, child, nextScope, scopeOfNode)
	}

	return scope
}

// GetScope finds the nearest surrounding scope of an AST node
func (st *ScopeTree) GetScope(node *sitter.Node) *Scope {
	if scope, exists := st.ScopeOfNode[node]; exists {
		return scope
	}

	if parent := node.Parent(); parent != nil {
		return st.GetScope(parent)
	}

	return nil
}

func MakeScopeTree(lang Language, ast *sitter.Node, source []byte) *ScopeTree {
	switch lang {
	case LangPy:
		return nil
	case LangTs, LangJs, LangTsx:
		builder := &TsScopeBuilder{
			ast:    ast,
			source: source,
		}
		return BuildScopeTree(builder, ast, source)
	default:
		return nil
	}
}
