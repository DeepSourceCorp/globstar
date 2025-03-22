package python

import (
	// "fmt"
	"slices"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"globstar.dev/analysis"
)

var DjangoInsecurePickleDeserialize *analysis.Analyzer = &analysis.Analyzer{
	Name:        "django-insecure-pickle-deserialize",
	Language:    analysis.LangPy,
	Description: "Insecure deserialization with pickle, _pickle, cpickle, dill, shelve, or yaml can lead to remote code execution. These libraries execute arbitrary code when loading untrusted data",
	Category:    analysis.CategorySecurity,
	Severity:    analysis.SeverityWarning,
	Run:         checkDjangoInsecurePickleDeserialize,
}

func checkDjangoInsecurePickleDeserialize(pass *analysis.Pass) (interface{}, error) {
	reqvarmap := make(map[string]bool)

	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() != "assignment" {
			return
		}

		left := node.ChildByFieldName("left")
		right := node.ChildByFieldName("right")

		if right == nil {
			return
		}

		if isRequestSource(right, pass.FileContext.Source, reqvarmap) {
			reqvarmap[left.Content(pass.FileContext.Source)] = true
		}
	})

	// detect pickle and dill calls
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() != "call" {
			return
		}

		libraryNames := []string{"pickle", "_pickle", "cPickle", "shelve", "dill"}
		methodNames := []string{"dump", "dumps", "load", "loads", "dump_session", "load_session"}

		funcNode := node.ChildByFieldName("function")
		if funcNode.Type() != "attribute" {
			return
		}

		funcObj := funcNode.ChildByFieldName("object").Content(pass.FileContext.Source)
		funcAttr := funcNode.ChildByFieldName("attribute").Content(pass.FileContext.Source)

		if !slices.Contains(libraryNames, funcObj) || !slices.Contains(methodNames, funcAttr) {
			return
		}

		argListNode := node.ChildByFieldName("arguments")
		if argListNode.Type() != "argument_list" {
			return
		}

		argNodes := getNamedChildren(argListNode, 0)

		for _, arg := range argNodes {
			if arg.Type() == "identifier" {
				if reqvarmap[arg.Content(pass.FileContext.Source)] {
					pass.Report(pass, node, "Detected insecure deserialization which may lead to remote code execution—use safer alternatives or validate input")
				}
			} else if arg.Type() == "call" {
				if isRequestSource(arg, pass.FileContext.Source, reqvarmap) {
					pass.Report(pass, node, "Detected insecure deserialization which may lead to remote code execution—use safer alternatives or validate input")
				}
			}
		}
	})

	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() != "call" {
			return
		}

		yamlMethodNames := []string{"dump", "dump_all", "load", "load_all"}
		funcNode := node.ChildByFieldName("function")
		if funcNode.Type() != "attribute" {
			return
		}

		funcObj := funcNode.ChildByFieldName("object").Content(pass.FileContext.Source)
		funcAttr := funcNode.ChildByFieldName("attribute").Content(pass.FileContext.Source)

		if funcObj != "yaml" || !slices.Contains(yamlMethodNames, funcAttr) {
			return
		}

		argListNode := node.ChildByFieldName("arguments")
		if argListNode.Type() != "argument_list" {
			return
		}
		argNodes := getNamedChildren(argListNode, 0)
		containsSafeParam := false
		for _, arg := range argNodes {
			if arg.Type() == "keyword_argument" {
				key := arg.ChildByFieldName("name").Content(pass.FileContext.Source)
				val := arg.ChildByFieldName("value").Content(pass.FileContext.Source)

				if (key == "Dumper" || key == "Loader") && (strings.Contains(val, "SafeDumper") || strings.Contains(val, "SafeLoader")) {
					containsSafeParam = true
				}
			}
		}

		if containsSafeParam {
			return
		}

		for _, arg := range argNodes {
			if arg.Type() == "identifier" {
				if reqvarmap[arg.Content(pass.FileContext.Source)] {
					pass.Report(pass, node, "Detected insecure deserialization which may lead to remote code execution—use safer alternatives or validate input")
				}
			} else if arg.Type() == "call" {
				if isRequestSource(arg, pass.FileContext.Source, reqvarmap) {
					pass.Report(pass, node, "Detected insecure deserialization which may lead to remote code execution—use safer alternatives or validate input")
				}
			}
		}

	})

	return nil, nil
}

func isRequestSource(node *sitter.Node, source []byte, reqVarMap map[string]bool) bool {
	switch node.Type() {
	case "call":
		if isEncoded(node, source, reqVarMap) {
			return true
		}
		funcNode := node.ChildByFieldName("function")
		if funcNode.Type() != "attribute" {
			return false
		}
		objectNode := funcNode.ChildByFieldName("object")
		if !strings.Contains(objectNode.Content(source), "request") {
			return false
		}

		attributeNode := funcNode.ChildByFieldName("attribute")
		if attributeNode.Type() != "identifier" {
			return false
		}

		return true

	case "subscript":
		valueNode := node.ChildByFieldName("value")
		if valueNode.Type() != "attribute" {
			return false
		}

		objNode := valueNode.ChildByFieldName("object")
		if objNode.Type() != "identifier" && objNode.Content(source) != "request" {
			return false
		}

		return true
	
	case "identifier":
		return reqVarMap[node.Content(source)]
	}

	return false
}

func isEncoded(node *sitter.Node, source []byte, reqVarMap map[string]bool) bool {
	if node.Type() != "call" {
		return false
	}

	if strings.HasPrefix(node.Content(source), "request") {
		return false
	}

	argListNode := node.ChildByFieldName("arguments")
	if argListNode.Type() != "argument_list" {
		return false
	}

	argNodes := getNamedChildren(argListNode, 0)
	for _, arg := range argNodes {
		if isRequestSource(arg, source, reqVarMap) {
			return true
		}
	}

	return false
}
