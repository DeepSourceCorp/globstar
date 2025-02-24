package python

import (
	"fmt"
	"path/filepath"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"globstar.dev/analysis"
)

// SQLInjection creates an analyzer that detects unsafe SQL query construction.
func SQLInjection() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:        "sql-injection",
		Description: "Detects unsafe SQL query construction",
		Category:    analysis.CategorySecurity,
		Severity:    analysis.SeverityCritical,
		Language:    analysis.LangPy, // assuming LangPy is defined elsewhere
		Run: func(pass *analysis.Pass) (interface{}, error) {
			// Define an inner function that captures pass.
			fn := func(n *sitter.Node) {
				visitCall(pass, n)
			}
			analysis.Preorder(pass, fn)
			return nil, nil
		},
	}
}

// visitCall checks if a call node is a SQL execution call and if its arguments are unsafe.
func visitCall(pass *analysis.Pass, node *sitter.Node) {
	source := pass.FileContext.Source

	// Only process call nodes.
	if node.Type() != "call" {
		return
	}

	// Extract the function part (e.g. cursor.execute).
	functionNode := node.ChildByFieldName("function")
	if functionNode == nil {
		return
	}

	// Proceed only if the function is one of our recognized SQL methods.
	if !isSQLExecuteMethod(functionNode, source) {
		return
	}

	// Check the first argument.
	argsNode := node.ChildByFieldName("arguments")
	if argsNode == nil {
		return
	}
	firstArg := getNthChild(argsNode, 0)
	if firstArg == nil {
		return
	}

	// If the query string is built unsafely, report an issue.
	if isUnsafeString(firstArg, source) {
		pass.Report(pass, node, "Direct use of unsafe string in SQL query")
		return
	}

	// If the argument is an identifier, trace its origin.
	if firstArg.Type() == "identifier" {
		varName := firstArg.Content(source)
		traceVariableOrigin(pass, varName, node, make(map[string]bool), make(map[string]bool), source)
	}
}

// isSQLExecuteMethod returns true if the function node is one of the SQL execution methods.
func isSQLExecuteMethod(node *sitter.Node, source []byte) bool {
	var funcName string
	switch node.Type() {
	case "identifier":
		funcName = node.Content(source)
	case "attribute":
		attr := node.ChildByFieldName("attribute")
		if attr != nil {
			funcName = attr.Content(source)
		}
	}

	sqlMethods := map[string]bool{
		"execute":       true,
		"executemany":   true,
		"executescript": true,
	}
	return sqlMethods[funcName]
}

// isUnsafeString returns true if the node represents an unsafely built string (e.g. f-string interpolation or unsafe concatenation).
func isUnsafeString(node *sitter.Node, source []byte) bool {
	// Check for f-strings with interpolation.
	if node.Type() == "fstring" {
		for i := 0; i < int(node.ChildCount()); i++ {
			if node.Child(i).Type() == "interpolation" {
				return true
			}
		}
	}

	// Check for unsafe binary concatenation.
	if node.Type() == "binary_operator" {
		op := node.ChildByFieldName("operator")
		if op != nil && op.Content(source) == "+" {
			return containsVariable(node.ChildByFieldName("left"), source) ||
				containsVariable(node.ChildByFieldName("right"), source)
		}
	}

	return false
}

// traceVariableOrigin recursively traces the origin of a variable through local assignments and cross‐file imports.
func traceVariableOrigin(pass *analysis.Pass, varName string, originalNode *sitter.Node,
	visitedVars map[string]bool, visitedFiles map[string]bool, source []byte) {

	if visitedVars[varName] {
		return
	}
	visitedVars[varName] = true

	if traceLocalAssignments(pass, varName, originalNode, visitedVars, visitedFiles, source) {
		return
	}

	traceCrossFileImports(pass, varName, originalNode, visitedVars, visitedFiles, source)
}

// traceLocalAssignments looks for local assignments to the variable and reports if it originates from an unsafe string.
func traceLocalAssignments(pass *analysis.Pass, varName string, originalNode *sitter.Node,
	visitedVars map[string]bool, visitedFiles map[string]bool, source []byte) bool {

	query := `(assignment left: (identifier) @var right: (_) @value)`
	q, err := sitter.NewQuery([]byte(query), pass.Analyzer.Language.Parser())
	if err != nil {
		return false
	}
	defer q.Close()

	cursor := sitter.NewQueryCursor()
	defer cursor.Close()
	cursor.Exec(q, pass.FileContext.Ast)

	for {
		match, ok := cursor.NextMatch()
		if !ok {
			break
		}

		var varNode, valueNode *sitter.Node
		for _, capture := range match.Captures {
			switch capture.Name {
			case "var":
				varNode = capture.Node
			case "value":
				valueNode = capture.Node
			}
		}

		if varNode != nil && varNode.Content(source) == varName {
			if isUnsafeString(valueNode, source) {
				pass.Report(pass, originalNode, fmt.Sprintf("Variable '%s' originates from an unsafe string", varName))
				return true
			}

			if valueNode.Type() == "identifier" {
				newVar := valueNode.Content(source)
				traceVariableOrigin(pass, newVar, originalNode, visitedVars, visitedFiles, source)
				return true
			}
		}
	}
	return false
}

// traceCrossFileImports looks for import-from statements to trace the variable across files.
func traceCrossFileImports(pass *analysis.Pass, varName string, originalNode *sitter.Node,
	visitedVars map[string]bool, visitedFiles map[string]bool, source []byte) {

	query := `(
		(import_from_statement
			module_name: (dotted_name) @module
			name: (dotted_name) @imported_var
		) @import
	)`
	q, err := sitter.NewQuery([]byte(query), pass.Analyzer.Language.Parser())
	if err != nil {
		return
	}
	defer q.Close()

	cursor := sitter.NewQueryCursor()
	defer cursor.Close()
	cursor.Exec(q, pass.FileContext.Ast)

	for {
		match, ok := cursor.NextMatch()
		if !ok {
			break
		}

		var moduleNode, varNode *sitter.Node
		for _, capture := range match.Captures {
			switch capture.Name {
			case "module":
				moduleNode = capture.Node
			case "imported_var":
				varNode = capture.Node
			}
		}

		if varNode != nil && varNode.Content(source) == varName && moduleNode != nil {
			modulePath := convertImportToPath(moduleNode.Content(source))
			if visitedFiles[modulePath] {
				continue
			}
			visitedFiles[modulePath] = true

			for _, file := range pass.Files {
				if strings.HasSuffix(file.FilePath, modulePath) {
					newPass := &analysis.Pass{
						Analyzer:    pass.Analyzer,
						FileContext: file,
						Files:       pass.Files,
						Report:      pass.Report,
					}
					traceVariableOrigin(newPass, varName, originalNode, visitedVars, visitedFiles, file.Source)
				}
			}
		}
	}
}

// containsVariable returns true if the node (or any of its subnodes) is an identifier or attribute.
func containsVariable(node *sitter.Node, source []byte) bool {
	if node == nil {
		return false
	}
	switch node.Type() {
	case "identifier", "attribute":
		return true
	case "binary_operator":
		return containsVariable(node.ChildByFieldName("left"), source) ||
			containsVariable(node.ChildByFieldName("right"), source)
	case "parenthesized_expression":
		return containsVariable(node.NamedChild(0), source)
	default:
		return false
	}
}

// getNthChild returns the nth child of a node or nil if out of bounds.
func getNthChild(node *sitter.Node, n int) *sitter.Node {
	if n < int(node.ChildCount()) {
		return node.Child(n)
	}
	return nil
}

// convertImportToPath converts a dotted module name to a file path (e.g. "a.b" -> "a/b.py").
func convertImportToPath(importStr string) string {
	return strings.ReplaceAll(importStr, ".", string(filepath.Separator)) + ".py"
}
