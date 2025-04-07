package python

import (
	// "fmt"
	// "fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"globstar.dev/analysis"
)

var PathTraversalOpen *analysis.Analyzer = &analysis.Analyzer{
	Name:        "path-traversal-open",
	Language:    analysis.LangPy,
	Description: "This vulnerability involves a path traversal risk in the 'open' function call where request data is used without proper validation or sanitization. Attackers could manipulate the file path to access sensitive files outside the intended directory. To mitigate this issue, developers should implement proper path validation using os.path.abspath, os.path.realpath, or the pathlib library.",
	Category:    analysis.CategorySecurity,
	Severity:    analysis.SeverityError,
	Run:         checkPathTraversalOpen,
}

func checkPathTraversalOpen(pass *analysis.Pass) (interface{}, error) {
	userDataVarMap := make(map[string]bool)

	// get the variable name from the Flask decorated route function
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() != "decorated_definition" {
			return
		}
		decNode := node.NamedChild(0)
		if decNode.Type() != "decorator" {
			return
		}
		callNode := decNode.NamedChild(0)
		if callNode.Type() != "call" {
			return
		}
		funcNode := callNode.ChildByFieldName("function")
		if funcNode.Type() != "attribute" {
			return
		}
		if !strings.HasSuffix(funcNode.Content(pass.FileContext.Source), ".route") {
			return
		}
		defNode := node.ChildByFieldName("definition")
		paramsNode := defNode.ChildByFieldName("parameters")
		if paramsNode.Type() != "parameters" {
			return
		}
		allparamNodes := getNamedChildren(paramsNode, 0)
		for _, p := range allparamNodes {
			userDataVarMap[p.Content(pass.FileContext.Source)] = true
		}
	})

	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() != "assignment" {
			return
		}

		rightNode := node.ChildByFieldName("right")
		leftNode := node.ChildByFieldName("left")

		if rightNode == nil {
			return
		}

		if isUSerData(rightNode, pass.FileContext.Source, userDataVarMap) {
			userDataVarMap[leftNode.Content(pass.FileContext.Source)] = true
		}
	})

	// for normal open() calls
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() != "call" {
			return
		}
		funcNode := node.ChildByFieldName("function")
		if !strings.HasPrefix(funcNode.Content(pass.FileContext.Source), "open") {
			return
		}

		argListNode := node.ChildByFieldName("arguments")
		if argListNode.Type() != "argument_list" {
			return
		}
		argNodes := getNamedChildren(argListNode, 0)
		for _, arg := range argNodes {
			if isUSerData(arg, pass.FileContext.Source, userDataVarMap) {
				pass.Report(pass, node, "Unsanitized request data in 'open' function could allow path traversal attacks and unauthorized access to sensitive files.")
			}
		}
	})
	return nil, nil
}
