package python

import (
	"strings"
	"regexp"

	sitter "github.com/smacker/go-tree-sitter"
	"globstar.dev/analysis"
)

var DjangoRequestHttpResponse *analysis.Analyzer = &analysis.Analyzer{
	Name:        "django-request-httpresponse",
	Language:    analysis.LangPy,
	Description: "User-controlled data in `HttpResponse` may enable XSS, allowing attackers to steal cookies or sensitive data. Escape or sanitize input to prevent script injection. Use secure templating or built-in encoding to mitigate risks.",
	Category:    analysis.CategorySecurity,
	Severity:    analysis.SeverityWarning,
	Run:         checkDjangoRequestHttpResponse,
}

func checkDjangoRequestHttpResponse(pass *analysis.Pass) (interface{}, error) {
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

		if isUserTaintedHttpResp(rightNode, pass.FileContext.Source, intermVarMap, reqVarMap) {
			intermVarMap[leftNode.Content(pass.FileContext.Source)] = true
		}
	})

	// catch insecure HttpResponse and HttpResponseBadRequest calls
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() != "call" {
			return
		}

		funcNode := node.ChildByFieldName("function")
		if funcNode.Type() != "attribute" && funcNode.Type() != "identifier" {
			return
		}
		
		if !strings.HasSuffix(funcNode.Content(pass.FileContext.Source), "HttpResponse") && !strings.HasSuffix(funcNode.Content(pass.FileContext.Source), "HttpResponseBadRequest") {
			return
		}
		
		argListNode := node.ChildByFieldName("arguments")
		if argListNode.Type() != "argument_list" {
			return
		}

		argNodes := getNamedChildren(argListNode, 0)
		for _, arg := range argNodes {
			if isUserTaintedHttpResp(arg, pass.FileContext.Source, intermVarMap, reqVarMap) {
				pass.Report(pass, node, "User-controlled data in `HttpResponse`/`HttpResponseBadRequest` may lead to XSS—escape or sanitize input to prevent script injection and data exposure.")
			}
		}
	})

	return nil, nil
}

func isUserTaintedHttpResp(node *sitter.Node, source []byte, intermVarMap, reqVarMap map[string]bool) bool {
	switch node.Type() {
	case "call":
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