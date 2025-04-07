package python

import (
	"regexp"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"globstar.dev/analysis"
)

var DangerousCreateExec *analysis.Analyzer = &analysis.Analyzer{
	Name:        "dangerous-create-exec",
	Language:    analysis.LangPy,
	Description: "This checker detects when `create_subprocess_exec` has tainted data passed into it. This can cause a Command Injection vulnerability",
	Category:    analysis.CategorySecurity,
	Severity:    analysis.SeverityError,
	Run:         dangerousCreateExec,
}

func dangerousCreateExec(pass *analysis.Pass) (interface{}, error) {
	eventVars := make(map[string]bool)

	// first pass to get variables tainted by event
	analysis.Preorder(pass, func(node *sitter.Node) {
		// check if this is an assignment
		if node.Type() != "assignment" {
			return
		}

		left := node.ChildByFieldName("left")
		right := node.ChildByFieldName("right")

		if left.Type() == "identifier" && isEventSource(right, pass.FileContext.Source) {
			eventVars[left.Content(pass.FileContext.Source)] = true
		}

	})

	// second pass to find where the unsafe variables have been used
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() == "call" && isFunctionNamed(node, pass.FileContext.Source) {
			argNode := node.ChildByFieldName("arguments")

			if argNode != nil && argNode.NamedChildCount() >= 2 {
				listSplat := argNode.NamedChild(1)

				if listSplat != nil && listSplat.NamedChildCount() > 0 {
					listSplatName := listSplat.NamedChild(0)

					if listSplatName.Type() == "identifier" {
						listSplatNameContent := listSplatName.Content(pass.FileContext.Source)

						if eventVars[listSplatNameContent] {
							pass.Report(pass, node, "Detected passing tainted data from `event` parameter directly to an exec method which can cause command injection vulnerabilities")
						}
					}
				}
			}
		}
	})

	// find where tainted event parameter has been directly used
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() == "call" && isFunctionNamed(node, pass.FileContext.Source) {
			argNode := node.ChildByFieldName("arguments")

			if argNode != nil && argNode.NamedChildCount() >= 2 {
				eventvarNode := argNode.NamedChild(1)

				if eventvarNode.Type() != "subscript" {
					return
				}

				eventVarSubscript := eventvarNode.ChildByFieldName("value")

				if eventVarSubscript.Content(pass.FileContext.Source) == "event" {
					pass.Report(pass, node, "Detected passing tainted data from `event` parameter directly to an exec method which can cause command injection vulnerabilities")
				}
			}
		}
	})

	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() == "call" && isFunctionNamed(node, pass.FileContext.Source) {
			argNode := node.ChildByFieldName("arguments")

			if argNode != nil && argNode.NamedChildCount() >= 4 {
				shellNode := argNode.NamedChild(1)
				eventNode := argNode.NamedChild(3)

				if shellNode.Type() != "string" || eventNode.Type() != "identifier" {
					return
				}

				shellString := shellNode.Content(pass.FileContext.Source)
				eventVar := eventNode.Content(pass.FileContext.Source)

				if isShell(shellString) && eventVars[eventVar] {
					pass.Report(pass, node, "Detected passing tainted data from `event` parameter directly to an exec method which can cause command injection vulnerabilities")
				}
			}
		}
	})

	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() == "call" && isFunctionNamed(node, pass.FileContext.Source) {
			argNode := node.ChildByFieldName("arguments")

			if argNode != nil && argNode.NamedChildCount() >= 4 {
				shellNode := argNode.NamedChild(1)
				eventNode := argNode.NamedChild(3)

				if shellNode.Type() != "string" || eventNode.Type() != "subscript" {
					return
				}

				shellString := shellNode.Content(pass.FileContext.Source)
				eventVarSubscript := eventNode.ChildByFieldName("value")

				if isShell(shellString) && eventVarSubscript.Content(pass.FileContext.Source) == "event" {
					pass.Report(pass, node, "Detected passing tainted data from `event` parameter directly to an exec method which can cause command injection vulnerabilities")
				}
			}
		}
	})

	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() == "call" && isFunctionNamed(node, pass.FileContext.Source) {
			argNode := node.ChildByFieldName("arguments")

			if argNode != nil && argNode.NamedChildCount() >= 2 {
				listNode := argNode.NamedChild(1)

				if listNode.Type() != "list" {
					return
				}

				shellNode := listNode.NamedChild(0)
				eventNode := listNode.NamedChild(2)

				if shellNode.Type() != "string" || eventNode.Type() != "identifier" {
					return
				}

				shellString := shellNode.Content(pass.FileContext.Source)
				eventVar := eventNode.Content(pass.FileContext.Source)

				if isShell(shellString) && eventVars[eventVar] {
					pass.Report(pass, node, "Detected passing tainted data from `event` parameter directly to an exec method which can cause command injection vulnerabilities")
				}
			}
		}
	})

	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() == "call" && isFunctionNamed(node, pass.FileContext.Source) {
			argNode := node.ChildByFieldName("arguments")

			if argNode != nil && argNode.NamedChildCount() >= 2 {
				listNode := argNode.NamedChild(1)

				if listNode.Type() != "list" {
					return
				}

				shellNode := listNode.NamedChild(0)
				eventNode := listNode.NamedChild(2)

				if shellNode.Type() != "string" || eventNode.Type() != "subscript" {
					return
				}

				shellString := shellNode.Content(pass.FileContext.Source)
				eventVarSubscript := eventNode.ChildByFieldName("value")

				if isShell(shellString) && eventVarSubscript.Content(pass.FileContext.Source) == "event" {
					pass.Report(pass, node, "Detected passing tainted data from `event` parameter directly to an exec method which can cause command injection vulnerabilities")
				}
			}
		}
	})

	return nil, nil
}

func isEventSource(node *sitter.Node, source []byte) bool {
	if node.Type() == "subscript" && node.ChildByFieldName("value").Content(source) == "event" {
		return true
	}
	return false
}

func isFunctionNamed(node *sitter.Node, source []byte) bool {
	funcName := node.ChildByFieldName("function").Content(source)

	if funcName == "asyncio.subprocess.create_subprocess_exec" || funcName == "subprocess.create_subprocess_exec" || funcName == "create_subprocess_exec" {

		return true
	}
	return false
}

func isShell(shellString string) bool {
	shellString = strings.Trim(shellString, `"'`)
	shellPattern := regexp.MustCompile(`(^|/)?(sh|bash|ksh|csh|tcsh|zsh)$`)
	return shellPattern.MatchString(shellString)
}
