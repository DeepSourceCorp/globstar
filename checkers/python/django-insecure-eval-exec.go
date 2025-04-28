package python

import (
	"regexp"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"globstar.dev/analysis"
)

var DjangoInsecureEvalExec *analysis.Analyzer = &analysis.Analyzer{
	Name:        "django-insecure-eval-exec",
	Language:    analysis.LangPy,
	Description: "Using `eval`/`exec` or `os.system` with user data creates a severe security vulnerability that allows attackers to execute arbitrary code on your system. This dangerous practice can lead to complete system compromise, data theft, or service disruption. Instead, replace `eval` with dedicated libraries or methods specifically designed for your required functionality.",
	Category:    analysis.CategorySecurity,
	Severity:    analysis.SeverityWarning,
	Run:         checkDjangoInsecureEvalExec,
}

func checkDjangoInsecureEvalExec(pass *analysis.Pass) (interface{}, error) {
	requestVarMap := make(map[string]bool)
	userFmtStrVarMap := make(map[string]bool)

	// first pass: check for assignment of `request` data stored in variables
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() != "assignment" {
			return
		}

		leftNode := node.ChildByFieldName("left")
		rightNode := node.ChildByFieldName("right")

		if rightNode.Type() != "call" && rightNode.Type() != "subscript" && rightNode.Type() != "binary_operator" {
			return
		}

		if isRequestCall(rightNode, pass.FileContext.Source) {
			varName := leftNode.Content(pass.FileContext.Source)
			requestVarMap[varName] = true
		}

	})

	// second pass: get variable names for string formatting
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() != "assignment" {
			return
		}

		leftNode := node.ChildByFieldName("left")
		rightNode := node.ChildByFieldName("right")

		if isStringFormatted(rightNode, pass.FileContext.Source, requestVarMap, userFmtStrVarMap) {
			userFmtStrVarMap[leftNode.Content(pass.FileContext.Source)] = true
		}

	})

	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() != "call" {
			return

		}

		funcNode := node.ChildByFieldName("function")
		funcPattern := `\b(eval|exec)\b`
		re := regexp.MustCompile(funcPattern)
		if !re.MatchString(funcNode.Content(pass.FileContext.Source)) && !strings.Contains(funcNode.Content(pass.FileContext.Source), "os.system") {
			return
		}

		argNode := node.ChildByFieldName("arguments")

		argumentList := getNamedChildren(argNode, 0)

		for _, arg := range argumentList {
			if arg.Type() == "identifier" {
				if hasUserDataVar(arg, pass.FileContext.Source, requestVarMap, userFmtStrVarMap) {
					pass.Report(pass, node, "Detected user data in code-execution call which can cause remote code execution")
				}
			} else if isRequestCall(arg, pass.FileContext.Source) {
				pass.Report(pass, node, "Detected user data in code-execution call which can cause remote code execution")
			} else if isStringFormatted(arg, pass.FileContext.Source, requestVarMap, userFmtStrVarMap) {
				pass.Report(pass, node, "Detected user data in code-execution call which can cause remote code execution")
			} else if isBase64Decoded(arg, pass.FileContext.Source, requestVarMap, userFmtStrVarMap) {
				pass.Report(pass, node, "Detected user data in code-execution call which can cause remote code execution")
			}
		}

	})
	return nil, nil
}

func isBase64Decoded(node *sitter.Node, source []byte, reqVarMap map[string]bool, userFmtStringVarMap map[string]bool) bool {
	if node.Type() != "call" {
		return false
	}

	funcNode := node.ChildByFieldName("function")
	if funcNode.Type() != "attribute" {
		return false
	}

	funcAttribute := funcNode.Content(source)
	if funcAttribute != "base64.decodestring" && funcAttribute != "decodestring" {
		return false
	}

	argsListNode := node.ChildByFieldName("arguments")
	argsList := getNamedChildren(argsListNode, 0)
	for _, argsNode := range argsList {
		switch argsNode.Type() {
		case "call":
			if isRequestCall(argsNode, source) || isStringFormatted(argsNode, source, reqVarMap, userFmtStringVarMap) {
				return true
			}
			byteFuncName := argsNode.ChildByFieldName("function")
			if byteFuncName.Type() != "identifier" && byteFuncName.Content(source) != "bytes" {
				return false
			}
			byteArgsNode := argsNode.ChildByFieldName("arguments")
			if byteArgsNode.NamedChildCount() == 0 {
				return false
			}

			byteArgs := getNamedChildren(byteArgsNode, 0)
			for _, arg := range byteArgs {
				if isStringFormatted(arg, source, reqVarMap, userFmtStringVarMap) || isRequestCall(arg, source) || hasUserDataVar(arg, source, reqVarMap, userFmtStringVarMap) {
					return true
				}
			}

			return false

		case "identifier":
			if hasUserDataVar(argsNode, source, reqVarMap, userFmtStringVarMap) {
				return true
			}
			return false

		case "string":
			if isStringFormatted(node, source, reqVarMap, userFmtStringVarMap) {
				return true
			}
			return false
		}
	}
	return false
}

func isStringFormatted(node *sitter.Node, source []byte, reqVarMap map[string]bool, userFmtStringVarMap map[string]bool) bool {
	switch node.Type() {
	case "call":
		funcNode := node.ChildByFieldName("function")
		if funcNode.Type() != "attribute" {
			return false
		}
		strObjectNode := funcNode.ChildByFieldName("object")
		funcAttribute := funcNode.Content(source)
		if !strings.HasSuffix(funcAttribute, ".format") || strObjectNode.Type() != "string" {
			return false
		}

		argNode := node.ChildByFieldName("arguments")
		if argNode.Type() != "argument_list" {
			return false
		}

		reqArgNode := argNode.NamedChild(0)
		if !isRequestCall(reqArgNode, source) && !hasUserDataVar(reqArgNode, source, reqVarMap, userFmtStringVarMap) {
			return false
		}

		return true

	case "binary_operator":
		binaryOpLeftNode := node.ChildByFieldName("left")
		binaryOpRightNode := node.ChildByFieldName("right")
		if binaryOpLeftNode.Type() != "string" {
			return false
		}

		if !isRequestCall(binaryOpRightNode, source) && !hasUserDataVar(binaryOpRightNode, source, reqVarMap, userFmtStringVarMap) {
			return false
		}

		return true

	case "string":
		strContent := node.Content(source)
		// check if f-string
		if strContent[0] != 'f' {
			return false
		}

		allChildren := getNamedChildren(node, 0)

		// check if user data is present in f-string interpolation
		for _, child := range allChildren {
			if child.Type() == "interpolation" {
				if isRequestCall(child.NamedChild(0), source) || hasUserDataVar(child.NamedChild(0), source, reqVarMap, userFmtStringVarMap) {
					return true
				}
			}
		}

	}

	return false
}

func hasUserDataVar(node *sitter.Node, source []byte, reqVarMap map[string]bool, userFmtStringVarMap map[string]bool) bool {
	if node.Type() != "identifier" {
		return false
	}

	argName := node.Content(source)

	for key := range reqVarMap {
		if argName == key {
			return true
		}
	}

	for key := range userFmtStringVarMap {
		if argName == key {
			return true
		}
	}

	return false
}
