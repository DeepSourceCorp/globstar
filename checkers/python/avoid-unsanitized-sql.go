package python

import (
	"fmt"
	"path/filepath"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"globstar.dev/analysis"
)

var AvoidUnsanitizedSQL = &analysis.Analyzer{
	Name:        "avoid-unsanitized-sql",
	Language:    analysis.LangPy,
	Description: "Check if SQL query is sanitized",
	Category:    analysis.CategorySecurity,
	Severity:    analysis.SeverityCritical,
	Run:         checkSQLInjection,
}

// checkSQLInjection is the rule callback that inspects each call node.
func checkSQLInjection(pass *analysis.Pass) (interface{}, error) {
	analysis.Preorder(pass, func(node *sitter.Node) {
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
			pass.Report(pass, node, "Concatenated string in SQL query is an SQL injection threat")
			return
		}

		// If the argument is an identifier, trace its origin.
		if firstArg.Type() == "identifier" {
			varName := firstArg.Content(source)
			traceVariableOrigin(pass, varName, node, make(map[string]bool), make(map[string]bool), source)
		}
	})

	return nil, nil
}

// --- Helper Functions ---

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

func traceLocalAssignments(pass *analysis.Pass, varName string, originalNode *sitter.Node,
	visitedVars map[string]bool, visitedFiles map[string]bool, source []byte) bool {

	query := `(assignment left: (identifier) @var right: (_) @value)`
	q, err := sitter.NewQuery([]byte(query), pass.Analyzer.Language.Grammar())
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
		for idx, capture := range match.Captures {
			switch q.CaptureNameForId(uint32(idx)) {
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

func traceCrossFileImports(pass *analysis.Pass, varName string, originalNode *sitter.Node,
	visitedVars map[string]bool, visitedFiles map[string]bool, source []byte) {

	query := `(
		(import_from_statement
			module_name: (dotted_name) @module
			name: (dotted_name) @imported_var
		) @import
	)`
	q, err := sitter.NewQuery([]byte(query), pass.Analyzer.Language.Grammar())
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
		for idx, capture := range match.Captures {
			switch q.CaptureNameForId(uint32(idx)) {
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
					// Create a temporary analyzer context for the imported file.
					tempPass := &analysis.Pass{
						Analyzer:    pass.Analyzer,
						FileContext: file,
						Files:       pass.Files,
						Report:      pass.Report, // Reuse the report function.
					}
					traceVariableOrigin(tempPass, varName, originalNode, visitedVars, visitedFiles, file.Source)
				}
			}
		}
	}
}

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

func getNthChild(node *sitter.Node, n int) *sitter.Node {
	if n < int(node.ChildCount()) {
		return node.Child(n)
	}
	return nil
}

func convertImportToPath(importStr string) string {
	return strings.ReplaceAll(importStr, ".", string(filepath.Separator)) + ".py"
}
