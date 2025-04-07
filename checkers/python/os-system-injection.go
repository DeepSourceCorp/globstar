package python

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"globstar.dev/analysis"
)

var OsSystemInjection *analysis.Analyzer = &analysis.Analyzer{
	Name: "os-system-injection",
	Language: analysis.LangPy,
	Description: "Command injection vulnerability detected where user-supplied data is passed directly to os.system. This allows attackers to execute arbitrary system commands by injecting shell metacharacters into the input. Replace with subprocess module and pass arguments as a list to properly separate command from parameters.",
	Category: analysis.CategorySecurity,
	Severity: analysis.SeverityError,
	Run: checkOsSystemInjection,
}

func checkOsSystemInjection(pass *analysis.Pass) (interface{}, error) {
	userDataVarMap := make(map[string]bool)

	// get the variable name from the Flask decorated route function
	analysis.Preorder(pass,func(node *sitter.Node) {
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
		if defNode.Type() != "function_definition" {
			return
		}

		paramsNode := defNode.ChildByFieldName("parameters")
		if paramsNode.NamedChildCount() == 0 {
			return
		}
		if paramsNode.Type() != "parameters" {
			return
		}
		allparamNodes := getNamedChildren(paramsNode, 0)
		for _, p := range allparamNodes {
			userDataVarMap[p.Content(pass.FileContext.Source)] = true
		}
	})

	// pass for tainted assignment
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() != "assignment" {
			return
		}

		leftNode := node.ChildByFieldName("left")
		rightNode := node.ChildByFieldName("right")
		if rightNode == nil {
			return
		}

		if isUSerData(rightNode, pass.FileContext.Source, userDataVarMap) {
			userDataVarMap[leftNode.Content(pass.FileContext.Source)] = true
		}
	})

	
	// detect the vulnerabilities
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() != "call" {
			return
		}
		funcNode := node.ChildByFieldName("function")
		if funcNode.Type() != "attribute" {
			return
		}
		funcObj := funcNode.ChildByFieldName("object")
		funcAttr := funcNode.ChildByFieldName("attribute")
		
		if funcObj.Type() != "identifier" || funcAttr.Type() != "identifier" {
			return
		}
		
		if funcObj.Content(pass.FileContext.Source) != "os" || funcAttr.Content(pass.FileContext.Source) != "system" {
			return
		}
		
		arglistnode := node.ChildByFieldName("arguments")
		if arglistnode.Type() != "argument_list" {
			return
		}

		argnodes := getNamedChildren(arglistnode, 0)
		for _, arg := range argnodes {
			if isUSerData(arg, pass.FileContext.Source, userDataVarMap) {
				pass.Report(pass, node, "Command injection risk in os.system; use subprocess with argument lists instead.")
			}
		}
	})


	return nil, nil
}