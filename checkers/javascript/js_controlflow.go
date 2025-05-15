//globstar:registry-exclude

package javascript

import (
	"fmt"
	"slices"
	"strings"

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

type CfgBlockType int

const (
	CfgBlockTypeStatement CfgBlockType = iota
	CfgBlockTypeEntry                  // Marking entry point of a Basic Block
	CfgBlockTypeExit                   // Marking exit point of a Basic Block
	CfgBlockTypeFunction               // Marking the start of a function node
	CfgBlockTypeBasic                  // Marking a basic block
)

var BlockDeclNodes = []string{
	"function_declaration",
	"if_statement",
	"for_statement",
}

// CFGNode is a node represeting a basic block in the control flow graph.
type CfgBlock struct {
	ID           int
	EnterNode    *sitter.Node
	Nodes        []*sitter.Node
	ExitNode     *sitter.Node
	Type         CfgBlockType
	Successors   []*CfgBlock
	Predecessors []*CfgBlock
	FunctionCtx  *FunctionCfgBlock // If the current node is a function declaration, link it to the individual Control Flow Node of the function.
}

// FunctionCFGBlock is a control flow graph for a function.
type FunctionCfgBlock struct {
	DeclarationNode *sitter.Node
	Name            string
	EntryNode       *sitter.Node
	ExitNodes       []*sitter.Node
	Nodes           []*sitter.Node // All CFG nodes in this function
}

type ControlFlowGraph struct {
	FileEntryBlock *CfgBlock
	FileExitBlock  *CfgBlock
	Functions      map[string]*FunctionCfgBlock
	AllBlocks      map[int]*CfgBlock
	nextNodeID     int
}

func NewControlFlowGraph() *ControlFlowGraph {
	return &ControlFlowGraph{
		Functions:  make(map[string]*FunctionCfgBlock),
		AllBlocks:  make(map[int]*CfgBlock),
		nextNodeID: 0,
	}
}

func (cfg *ControlFlowGraph) CreateBlock(node *sitter.Node, nodeType CfgBlockType, functionCtx *FunctionCfgBlock) (int, *CfgBlock) {
	cfgNode := &CfgBlock{
		ID:          cfg.nextNodeID,
		EnterNode:   node,
		Type:        nodeType,
		FunctionCtx: functionCtx,
	}
	cfg.AllBlocks[cfg.nextNodeID] = cfgNode
	cfg.nextNodeID++

	if functionCtx != nil {
		functionCtx.Nodes = append(functionCtx.Nodes, node)
	}

	return cfg.nextNodeID - 1, cfgNode
}

func AddEdge(from *CfgBlock, to *CfgBlock) {
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
// Add functionality to handle function calls and hoisting.

func createControlFlowGraph(pass *analysis.Pass) (interface{}, error) {
	cfg := NewControlFlowGraph()
	err := cfg.collectFunctions(pass)

	if err != nil {
		return nil, err
	}

	analysis.Preorder(pass, func(node *sitter.Node) {
		if node == nil {
			return
		}

		if node.Type() == "program" {
			if node.ChildCount() < 0 {
				return
			}
			cfg.CreateBlock(node, CfgBlockTypeEntry, nil)

			children := int(node.ChildCount())
			for i := 0; i < children; i++ {
				child := node.Child(i)
				fmt.Println("child", child.Type())
				if slices.Contains(BlockDeclNodes, child.Type()) {
					continue
					// Need a way to only attach the function decl block, when it is called in the actual code.
					// Right now it's part of the CFG even if it's not called.
					// if child.Type() == "function_declaration" {
					// 	block := cfg.Functions[child.ChildByFieldName("name").Content(pass.FileContext.Source)]
					// 	funcBlockIndex, _ := cfg.CreateBlock(child, CfgBlockTypeFunction, block)
					// 	AddEdge(cfg.AllBlocks[funcBlockIndex-1], cfg.AllBlocks[funcBlockIndex])
					// }
				}
				if child.Type() == "expression_statement" {
					callExp := child.Child(0)
					if callExp == nil {
						continue
					}
					fmt.Println("callExp", child.Child(0).Child(0).Type())
					funcNameStr := callExp.ChildByFieldName("function").Content(pass.FileContext.Source)
					block := cfg.Functions[funcNameStr]
					fmt.Println("block", block)
					if block == nil {
						continue
					}
					funcBlockIndex, _ := cfg.CreateBlock(block.DeclarationNode, CfgBlockTypeFunction, block)
					AddEdge(cfg.AllBlocks[funcBlockIndex-1], cfg.AllBlocks[funcBlockIndex])
					continue
				}
				basicBlockID, basicBlock := cfg.CreateBlock(child, CfgBlockTypeBasic, nil)
				AddEdge(cfg.AllBlocks[basicBlockID-1], basicBlock)

				j := i
				for j < children && !slices.Contains(BlockDeclNodes, node.Child(j).Type()) {
					basicBlock.Nodes = append(basicBlock.Nodes, node.Child(j))
					j++
				}
				i = j - 1 // -1 because the outer loop will increment i

			}
			exitBlockID, exitBlock := cfg.CreateBlock(nil, CfgBlockTypeExit, nil)
			AddEdge(cfg.AllBlocks[exitBlockID-1], exitBlock)
		} else {
			return
		}

	})

	// For debugging, you can uncomment this line to see the CFG immediately:
	fmt.Println(cfg.GenerateDOTWithSource(pass.FileContext.Source))

	return cfg, nil
}

func (cfg *ControlFlowGraph) collectFunctions(pass *analysis.Pass) error {
	functions := make(map[string]*FunctionCfgBlock)

	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() == "function_declaration" {
			funcName := node.ChildByFieldName("name")
			cfgNode, err := cfg.processFunctionDeclaration(pass, node)
			if err != nil {
				return
			}
			functions[funcName.Content(pass.FileContext.Source)] = cfgNode
		}
	})
	cfg.Functions = functions

	return nil
}

func (cfg *ControlFlowGraph) processFunctionDeclaration(pass *analysis.Pass, node *sitter.Node) (*FunctionCfgBlock, error) {
	if node.Type() != "function_declaration" {
		return nil, nil
	}

	funcName := node.ChildByFieldName("name")
	if funcName == nil {
		return nil, fmt.Errorf("function declaration has no name")
	}

	cfgNode := &FunctionCfgBlock{
		DeclarationNode: node,
		Name:            funcName.Content(pass.FileContext.Source),
		EntryNode:       node,
		ExitNodes:       make([]*sitter.Node, 0),
	}
	body := node.ChildByFieldName("body")
	bodyChildCount := body.NamedChildCount()
	for i := 0; i < int(bodyChildCount); i++ {
		child := body.NamedChild(i)
		if child.Type() == "return_statement" {
			cfgNode.ExitNodes = append(cfgNode.ExitNodes, child)
		} else {
			cfgNode.Nodes = append(cfgNode.ExitNodes, child)
		}
	}
	// cfg.CreateBlock(node, CfgBlockTypeFunction, cfgNode)

	return cfgNode, nil
}

// Print outputs the entire Control Flow Graph structure in a readable format
func (cfg *ControlFlowGraph) Print() string {
	var result string

	result += "Control Flow Graph:\n"
	result += fmt.Sprintf("Total Blocks: %d\n\n", len(cfg.AllBlocks))

	// Print all blocks in order
	for i := 0; i < len(cfg.AllBlocks); i++ {
		block := cfg.AllBlocks[i]
		if block == nil {
			continue
		}

		// Block header
		result += fmt.Sprintf("Block ID: %d (Type: %s)\n", block.ID, blockTypeToString(block.Type))

		// Show node type
		if block.EnterNode != nil {
			result += fmt.Sprintf("  Enter Node: %s\n", block.EnterNode.Type())
		}

		// List child nodes
		if len(block.Nodes) > 0 {
			result += "  Nodes:\n"
			for _, node := range block.Nodes {
				result += fmt.Sprintf("    - %s\n", node.Type())
			}
		}

		// Show function context if applicable
		if block.FunctionCtx != nil {
			result += fmt.Sprintf("  Function: %s\n", block.FunctionCtx.Name)
		}

		// Show edges (connections between blocks)
		if len(block.Predecessors) > 0 {
			result += "  Predecessors:"
			for _, pred := range block.Predecessors {
				result += fmt.Sprintf(" %d", pred.ID)
			}
			result += "\n"
		}

		if len(block.Successors) > 0 {
			result += "  Successors:"
			for _, succ := range block.Successors {
				result += fmt.Sprintf(" %d", succ.ID)
			}
			result += "\n"
		}

		result += "\n"
	}

	// Print functions
	if len(cfg.Functions) > 0 {
		result += "Functions:\n"
		for _, function := range cfg.Functions {
			result += fmt.Sprintf("  %s:\n", function.Name)
			result += fmt.Sprintf("    Entry: %s\n", function.EntryNode.Type())

			if len(function.ExitNodes) > 0 {
				result += "    Exit Nodes:\n"
				for _, exitNode := range function.ExitNodes {
					result += fmt.Sprintf("      - %s\n", exitNode.Type())
				}
			}

			result += "\n"
		}
	}

	return result
}

// Helper function to convert block type to string representation
func blockTypeToString(blockType CfgBlockType) string {
	switch blockType {
	case CfgBlockTypeStatement:
		return "Statement"
	case CfgBlockTypeEntry:
		return "Entry"
	case CfgBlockTypeExit:
		return "Exit"
	case CfgBlockTypeFunction:
		return "Function"
	case CfgBlockTypeBasic:
		return "Basic"
	default:
		return "Unknown"
	}
}

// GenerateDOT creates a DOT graph representation of the CFG for visualization
func (cfg *ControlFlowGraph) GenerateDOT() string {
	return cfg.GenerateDOTWithSource(nil)
}

// GenerateDOTWithSource creates a DOT graph representation of the CFG including source code snippets
func (cfg *ControlFlowGraph) GenerateDOTWithSource(source []byte) string {
	var result string

	// Start DOT graph
	result += "digraph CFG {\n"
	result += "  node [shape=box];\n"

	// Define nodes
	for _, block := range cfg.AllBlocks {
		if block == nil {
			continue
		}

		nodeLabel := fmt.Sprintf("Block %d\\n%s", block.ID, blockTypeToString(block.Type))

		// Add node type info
		if block.EnterNode != nil {
			nodeLabel += fmt.Sprintf("\\n%s", block.EnterNode.Type())
		}

		// Add function context if applicable
		if block.FunctionCtx != nil {
			nodeLabel += fmt.Sprintf("\\nFunction: %s", block.FunctionCtx.Name)
		}

		// Add all nodes in this block with their statement representation
		if len(block.Nodes) > 0 {
			nodeLabel += "\\n\\nStatements:"
			for _, node := range block.Nodes {
				// Add node type
				nodeLabel += fmt.Sprintf("\\n- %s", node.Type())

				// Add code snippet if source code is available
				if source != nil && node.StartByte() < uint32(len(source)) && node.EndByte() <= uint32(len(source)) {
					snippet := string(source[node.StartByte():node.EndByte()])

					// Clean the snippet for DOT format
					snippet = escapeForDot(snippet)

					// Truncate if too long
					if len(snippet) > 30 {
						snippet = snippet[:27] + "..."
					}

					nodeLabel += fmt.Sprintf(": %s", snippet)
				}
			}
		}

		// Node style based on type
		nodeStyle := ""
		switch block.Type {
		case CfgBlockTypeEntry:
			nodeStyle = ", color=green"
		case CfgBlockTypeExit:
			nodeStyle = ", color=red"
		case CfgBlockTypeFunction:
			nodeStyle = ", color=blue, style=filled, fillcolor=lightblue"
		}

		result += fmt.Sprintf("  node%d [label=\"%s\"%s];\n", block.ID, nodeLabel, nodeStyle)
	}

	// Define edges
	for _, block := range cfg.AllBlocks {
		if block == nil {
			continue
		}

		for _, succ := range block.Successors {
			result += fmt.Sprintf("  node%d -> node%d;\n", block.ID, succ.ID)
		}
	}

	// End DOT graph
	result += "}\n"

	return result
}

// escapeForDot escapes special characters in strings for DOT format
func escapeForDot(s string) string {
	// Replace newlines with \n
	s = strings.ReplaceAll(s, "\n", "\\n")

	// Replace quotes with escaped quotes
	s = strings.ReplaceAll(s, "\"", "\\\"")

	// Replace backslashes with escaped backslashes
	s = strings.ReplaceAll(s, "\\", "\\\\")

	return s
}

// PrintCFG is a utility function that can be used to print a CFG from analyzer results
func PrintCFG(result interface{}) string {
	cfg, ok := result.(*ControlFlowGraph)
	if !ok {
		return "Error: Result is not a ControlFlowGraph"
	}
	return cfg.Print()
}

// PrintCFGDOT returns the DOT representation of the CFG from analyzer results
func PrintCFGDOT(result interface{}) string {
	cfg, ok := result.(*ControlFlowGraph)
	if !ok {
		return "Error: Result is not a ControlFlowGraph"
	}
	return cfg.GenerateDOT()
}

// PrintCFGDOTWithSource returns the DOT representation with source code snippets
func PrintCFGDOTWithSource(result interface{}, source []byte) string {
	cfg, ok := result.(*ControlFlowGraph)
	if !ok {
		return "Error: Result is not a ControlFlowGraph"
	}
	return cfg.GenerateDOTWithSource(source)
}
