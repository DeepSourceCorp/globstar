package javascript

import (
	sitter "github.com/smacker/go-tree-sitter"
	"globstar.dev/analysis"
)

var SQLInjection = &analysis.Analyzer{
	Name:        "sql_injection",
	Language:    analysis.LangJs,
	Description: "Using raw SQL queries with unvalidated input can lead to SQL injection vulnerabilities",
	Category:    analysis.CategorySecurity,
	Severity:    analysis.SeverityCritical,
	Run:         detectSQLInjection,
}

func detectSQLInjection(pass *analysis.Pass) (interface{}, error) {
	// Map of vulnerable function names to watch for
	vulnerableFunctions := map[string]bool{
		"query":             true,
		"raw":               true,
		"$queryRawUnsafe":   true,
		"$executeRawUnsafe": true,
	}

	// Map to track variable definitions
	varDefinitions := make(map[string]*sitter.Node)

	// First pass: collect all variable definitions
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node == nil || node.Type() != "variable_declarator" {
			return
		}

		nameNode := node.ChildByFieldName("name")
		valueNode := node.ChildByFieldName("value")

		// Ensure that the variable definition is valid
		if nameNode != nil && nameNode.Type() == "identifier" && valueNode != nil {
			varName := nameNode.Content(pass.FileContext.Source)
			if varName != "" {
				varDefinitions[varName] = valueNode
			}
		}
	})

	// Second pass: detect SQL injection vulnerabilities
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node == nil || node.Type() != "call_expression" {
			return
		}

		funcNode := node.ChildByFieldName("function")
		if funcNode == nil || funcNode.Type() != "member_expression" {
			return
		}

		propertyNode := funcNode.ChildByFieldName("property")
		if propertyNode == nil {
			return
		}

		// Extract the function name
		funcName := propertyNode.Content(pass.FileContext.Source)

		// Check if this is a function that executes raw SQL
		if !vulnerableFunctions[funcName] {
			return
		}

		// Get the arguments of the function
		args := node.ChildByFieldName("arguments")
		if args == nil || args.NamedChildCount() == 0 {
			return
		}

		// Get the first argument
		firstArg := args.NamedChild(0)
		if firstArg == nil {
			return
		}

		// Check if the argument is vulnerable
		if isSQLInjectionVulnerable(firstArg, pass.FileContext.Source, varDefinitions) {
			pass.Report(pass, node, "Potential SQL injection vulnerability detected, use parameterized queries instead")
		}
	})

	return nil, nil
}

func isSQLInjectionVulnerable(node *sitter.Node, sourceCode []byte, varDefs map[string]*sitter.Node) bool {
	if node == nil {
		return false
	}

	switch node.Type() {
	case "binary_expression":
		// Check for string concatenation
		left := node.ChildByFieldName("left")
		right := node.ChildByFieldName("right")

		// If either side is an identifier, this could be user input
		if (left != nil && left.Type() == "identifier") ||
			(right != nil && right.Type() == "identifier") {
			return true
		}

		// Recursively check both sides
		return isSQLInjectionVulnerable(left, sourceCode, varDefs) ||
			isSQLInjectionVulnerable(right, sourceCode, varDefs)

	case "template_string":
		// Check for template strings with interpolation
		for i := range int(node.NamedChildCount()) {
			child := node.NamedChild(i)
			if child != nil && child.Type() == "template_substitution" {
				return true
			}
		}

	case "identifier":
		// If it's a variable, check its definition
		varName := node.Content(sourceCode)
		if defNode, exists := varDefs[varName]; exists {
			return isSQLInjectionVulnerable(defNode, sourceCode, varDefs)
		} else { // If definition is not found, assume it to be vulnerable
			return true
		}
	}

	return false
}
