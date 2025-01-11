package one

import sitter "github.com/smacker/go-tree-sitter"

type CgNode struct {
	FunDecl *sitter.Node
	// Calls is a list of nodes that correspond to functions that are
	// called *within* this function's body
	Calls []*CgEdge
}

type ParamToArg struct {
	ParamIndex int
	ArgIndex   int
}

type CgEdge struct {
	ParamToArgMap []ParamToArg
	Callee        *CgNode
}

type CgBuilder interface {
	IsCallExpr(node *sitter.Node) bool
	GetCallee(callExprNode *sitter.Node) *sitter.Node
}

