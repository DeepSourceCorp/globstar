package python

import (
	"regexp"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"globstar.dev/analysis"
)

var DjangoRequestDataWrite *analysis.Analyzer = &analysis.Analyzer{
	Name:        "django-request-data-write",
	Language:    analysis.LangPy,
	Description: "User-controlled request data is directly written to a file, which can lead to security risks such as unauthorized file modification, forced log rotation, or denial-of-service by exhausting disk space. Ensure proper input sanitization or escaping to mitigate these threats.",
	Category:    analysis.CategorySecurity,
	Severity:    analysis.SeverityWarning,
	Run:         checkDjangoRequestDataWrite,
}

func checkDjangoRequestDataWrite(pass *analysis.Pass) (interface{}, error) {
	reqVarMap := make(map[string]bool)
	intermVarMap := make(map[string]bool)

	// get var names for data received from `request` calls
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() != "assignment" {
			return
		}

		leftNode := node.ChildByFieldName("left")
		rightNode := node.ChildByFieldName("right")

		if rightNode == nil {
			return
		}

		if isRequestCall(rightNode, pass.FileContext.Source) {
			reqVarMap[leftNode.Content(pass.FileContext.Source)] = true
		}
	})

	// get var names for data formatted by user data
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() != "assignment" {
			return
		}

		leftNode := node.ChildByFieldName("left")
		rightNode := node.ChildByFieldName("right")

		if rightNode == nil {
			return
		}

		if isUserTainted(rightNode, pass.FileContext.Source, intermVarMap, reqVarMap) {
			intermVarMap[leftNode.Content(pass.FileContext.Source)] = true
		}
	})

	// catch insecure file write methods
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() != "call" {
			return
		}

		funcNode := node.ChildByFieldName("function")
		if funcNode.Type() != "attribute" && funcNode.Type() != "identifier" {
			return
		}

		if !strings.HasSuffix(funcNode.Content(pass.FileContext.Source), "write") {
			return
		}

		argListNode := node.ChildByFieldName("arguments")
		if argListNode.Type() != "argument_list" {
			return
		}

		argNodes := getNamedChildren(argListNode, 0)
		for _, arg := range argNodes {
			if isUserTaintedDataWrite(arg, pass.FileContext.Source, intermVarMap, reqVarMap) {
				pass.Report(pass, node, "User-controlled data written to a file may enable log tampering, forced rotation, or disk exhaustion—sanitize input before writing")
			}
		}

	})

	return nil, nil
}

func isUserTaintedDataWrite(node *sitter.Node, source []byte, intermVarMap, reqVarMap map[string]bool) bool {
	switch node.Type() {
	case "call":
		if isInFunc(node, source, intermVarMap, reqVarMap) {
			return true
		}
		functionNode := node.ChildByFieldName("function")
		if functionNode.Type() != "attribute" {
			return false
		}

		if !strings.HasSuffix(functionNode.Content(source), ".format") {
			return false
		}

		argListNode := node.ChildByFieldName("arguments")
		if argListNode.Type() != "argument_list" {
			return false
		}

		argsNode := getNamedChildren(argListNode, 0)
		for _, arg := range argsNode {
			if arg.Type() == "identifier" && reqVarMap[arg.Content(source)] {
				return true
			} else if arg.Type() == "call" && isRequestCall(arg, source) {
				return true
			}
		}

	case "string":
		if node.Content(source)[0] != 'f' {
			return false
		}
		stringChildrenNodes := getNamedChildren(node, 0)
		for _, strnode := range stringChildrenNodes {
			if strnode.Type() == "interpolation" {
				exprnode := strnode.ChildByFieldName("expression")
				if exprnode.Type() == "identifier" && reqVarMap[exprnode.Content(source)] {
					return true
				} else if exprnode.Type() == "call" && isRequestCall(exprnode, source) {
					return true
				}
			}
		}

	case "binary_operator":
		binOpStr := node.Content(source)

		for reqvar := range reqVarMap {
			pattern := `\b` + reqvar + `\b`
			re := regexp.MustCompile(pattern)

			if re.MatchString(binOpStr) {
				return true
			}
		}

		rightNode := node.ChildByFieldName("right")
		if rightNode.Type() == "call" && isRequestCall(rightNode, source) {
			return true
		} else if rightNode.Type() == "tuple" {
			targsNode := getNamedChildren(rightNode, 0)
			for _, targ := range targsNode {
				if targ.Type() == "identifier" && reqVarMap[targ.Content(source)] {
					return true
				} else if targ.Type() == "call" && isRequestCall(targ, source) {
					return true
				}
			}
		}

	case "identifier":
		return reqVarMap[node.Content(source)] || intermVarMap[node.Content(source)]

	case "subscript":
		return isRequestCall(node, source)
	}

	return false
}

func isInFunc(node *sitter.Node, source []byte, intermVarMap, reqvarmap map[string]bool) bool {
	if node.Type() != "call" {
		return false
	}
	if strings.HasSuffix(node.Content(source), ".format") {
		return false
	}
	argListNode := node.ChildByFieldName("arguments")
	if argListNode.Type() != "argument_list" {
		return false
	}
	argNodes := getNamedChildren(argListNode, 0)

	for _, arg := range argNodes {
		if isUserTainted(arg, source, intermVarMap, reqvarmap) {
			return true
		}
	}
	return false
}
