// globstar:registry-exclude

package javascript

import (
	sitter "github.com/smacker/go-tree-sitter"
	"globstar.dev/analysis"
)

var ControlFlowAnalyzer = &analysis.Analyzer{
	Name:        "control_flow_analyzer",
	Language:    analysis.LangJs,
	Description: "Create a Control Flow Graph for a javascript file",
	Run:         createControlFlowGraph,
	Requires:    []*analysis.Analyzer{ScopeAnalyzer},
}

type CFGNodeType int

const (
	NodeTypeStatement    CFGNodeType = iota
	NodeTypeEntry                    // Marking entry point of a Basic Block
	NodeTypeExit                     // Marking exit point of a Basic Block
	NodeTypeFunction                 // Marking the start of a function node
	NodeTypeFunctionCall             // Marking a function call
	NodeTypeReturn                   // Marking the return statement, or the end of a function
)

// CFGNode is a node represeting a basic block in the control flow graph.
type CFGNode struct {
	ID           int
	AstNode      *sitter.Node
	Type         CFGNodeType
	Successors   []*CFGNode
	Predecessors []*CFGNode
	Scope        *analysis.Scope
	FunctionCtx  *FunctionCFG // If the current node is a function declaration, link it to the individual Control Flow Node of the function.
}

// FunctionCFG is a control flow graph for a function.
type FunctionCFG struct {
	DeclarationNode *sitter.Node
	Name            string
	EntryNode       *CFGNode
	ExitNodes       []*CFGNode
	Nodes           []*CFGNode // All CFG nodes in this function
}

type ControlFlowGraph struct {
	FileEntryNode *CFGNode
	FileExitNode  *CFGNode
	Functions     map[*sitter.Node]*FunctionCFG
	AllNodes      map[int]*CFGNode
	nextNodeID    int
}

func (cfg *ControlFlowGraph) CreateNode(node *sitter.Node, nodeType CFGNodeType, scope *analysis.Scope, functionCtx *FunctionCFG) *CFGNode {
	cfgNode := &CFGNode{
		ID:          cfg.nextNodeID,
		AstNode:     node,
		Type:        NodeTypeStatement,
		Scope:       scope,
		FunctionCtx: functionCtx,
	}
	cfg.AllNodes[cfg.nextNodeID] = cfgNode
	cfg.nextNodeID++

	if functionCtx != nil {
		functionCtx.Nodes = append(functionCtx.Nodes, cfgNode)
	}

	return cfgNode
}

func AddEdge(from *CFGNode, to *CFGNode) {
	if from == nil || to == nil {
		return
	}
	from.Successors = append(from.Successors, to)
	to.Predecessors = append(to.Predecessors, from)
}

// TODO:
// Focus on a simple control flow implementation first.
// Each function call creates a new control flow node
// Hoisting needs to be handled for function calls (ref: https://developer.mozilla.org/en-US/docs/Web/JavaScript/Guide/Functions#function_hoisting)
// Hoisting only happens for function declarations, not function expressions.
// Gather all the function nodes prior to linking them. (A builder method?)
// Connecting the nodes will happen after the graph is fully created

func createControlFlowGraph(pass *analysis.Pass) (interface{}, error) {
	return nil, nil
}

func collectFunctions(pass *analysis.Pass) (map[*sitter.Node]*FunctionCFG, error) {
	return nil, nil
}

func processFunctionDeclaration(cfg *ControlFlowGraph, node *sitter.Node) (*FunctionCFG, error) {
	return nil, nil
}
