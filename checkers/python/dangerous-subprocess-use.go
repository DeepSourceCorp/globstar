package python

import (
	"regexp"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"globstar.dev/analysis"
)

var DangerpusSubprocessUse *analysis.Analyzer = &analysis.Analyzer{
	Name:        "dangerous-subprocess-use",
	Language:    analysis.LangPy,
	Description: "Potential command injection risk: event data flows into subprocess with shell=True. This allows attackers to execute arbitrary commands through maliciously crafted input. Consider using shell=False (the default) with shlex.split() to safely separate command arguments.",
	Category:    analysis.CategorySecurity,
	Severity:    analysis.SeverityError,
	Run:         checkDangerousSubprocessUse,
}

func checkDangerousSubprocessUse(pass *analysis.Pass) (interface{}, error) {
	eventVars := make(map[string]bool)

	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() != "assignment" {
			return
		}

		left := node.ChildByFieldName("left")
		right := node.ChildByFieldName("right")

		if left.Type() == "identifier" && isEventSource(right, pass.FileContext.Source) {
			eventVars[left.Content(pass.FileContext.Source)] = true
		}
	})

	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() == "call" && isFunctionSubprocess(node, pass.FileContext.Source) {
			argNode := node.ChildByFieldName("arguments")

			if argNode != nil && argNode.NamedChildCount() >= 2 {
				taintedDataNode := argNode.NamedChild(0)
				shellBoolNode := argNode.NamedChild(1)

				if isShellTrue(shellBoolNode, pass.FileContext.Source) && isEventTaintedNode(taintedDataNode, pass.FileContext.Source) {
					pass.Report(pass, node, "Unsanitized event data in subprocess with shell=True enables command injection")
				}
			}
		}
	})

	// tainted data present in list
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() == "call" && isFunctionSubprocess(node, pass.FileContext.Source) {
			argNode := node.ChildByFieldName("arguments")

			if argNode != nil && argNode.NamedChildCount() >= 2 {
				taintedListNode := argNode.NamedChild(0)
				shellBoolNode := argNode.NamedChild(1)

				if isShellTrue(shellBoolNode, pass.FileContext.Source) && isEventTaintedListNode(taintedListNode, pass.FileContext.Source, eventVars) {
					pass.Report(pass, node, "Unsanitized event data in subprocess with shell=True enables command injection")
				}
			}
		}
	})

	// subprocess.Popen based vulns
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() == "call" && isFunctionSubprocess(node, pass.FileContext.Source) {
			argNode := node.ChildByFieldName("arguments")

			if argNode != nil && argNode.NamedChildCount() >= 4 {
				argumentsNode := getNamedChildren(argNode, 0)
				eventBool := false
				shellBool := false

				for _, arg := range argumentsNode {
					if isEventTaintedNode(arg, pass.FileContext.Source) || isEventTaintedListNode(arg, pass.FileContext.Source, eventVars) {
						eventBool = true
					}

					if isShellTrue(arg, pass.FileContext.Source) {
						shellBool = true
					}
				}

				if shellBool && eventBool {
					pass.Report(pass, node, "Unsanitized event data in subprocess with shell=True enables command injection")
				}
			}
		}
	})

	return nil, nil
}

func isShlexSanitized(code string) bool {
	pattern := `shlex\.quote\(\s*[^)]+\s*\)|shlex\.split\(\s*[^)]+\s*\)|pipes\.quote\(\s*[^)]+\s*\)`
	re := regexp.MustCompile(pattern)
	return re.MatchString(code)
}

func isEventTaintedListNode(node *sitter.Node, source []byte, eventVars map[string]bool) bool {
	if node.Type() == "list" {
		listNodeContent := node.Content(source)
		for varName := range eventVars {
			if strings.Contains(listNodeContent, varName) && !isShlexSanitized(listNodeContent) {
				return true
			}
		}
	}
	return false
}

func isShellTrue(node *sitter.Node, source []byte) bool {
	if node.Type() == "keyword_argument" {
		nameNode := node.ChildByFieldName("name")
		valNode := node.ChildByFieldName("value")
		if nameNode.Type() != "identifier" {
			return false
		}

		if nameNode.Content(source) == "shell" && valNode.Type() == "true" {
			return true
		}
	}
	return false
}

func isEventSource(node *sitter.Node, source []byte) bool {

	switch node.Type() {
	case "subscript":
		if node.ChildByFieldName("value").Content(source) == "event" {
			return true
		}
	case "call":
		funcNode := node.ChildByFieldName("function")
		if funcNode.Type() != "attribute" {
			return false
		}
		subscriptNode := funcNode.ChildByFieldName("object")
		if subscriptNode.Type() == "subscript" && subscriptNode.ChildByFieldName("value").Content(source) == "event" {
			return true
		}
	case "string":
		strContent := node.Content(source)
		// checking for f-strings
		if strContent[0] == 'f' {
			if strings.Contains(strContent, "event") {
				return true
			}
		}
	}
	return false
}

func isFunctionSubprocess(node *sitter.Node, source []byte) bool {
	funcName := node.ChildByFieldName("function").Content(source)
	return strings.Contains(funcName, "subprocess")
}
