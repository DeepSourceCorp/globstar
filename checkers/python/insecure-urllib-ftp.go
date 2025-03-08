package python

import (
	"regexp"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"globstar.dev/analysis"
)

var InsecureOpenerDirectorFtpOpen *analysis.Analyzer = &analysis.Analyzer{
	Name:        "insecure-urllib-ftp",
	Language:    analysis.LangPy,
	Description: "An unsecured FTP connection was detected where ftp:// is being used. Data transmitted over this channel is unencrypted, posing a security risk. It is recommended to use SFTP instead. Since urllib does not support SFTP, consider using a library that provides secure file transfer capabilities.",
	Category:    analysis.CategorySecurity,
	Severity:    analysis.SeverityError,
	Run:         checkInsecureOpenerDirectorFtpOpen,
}

func checkInsecureOpenerDirectorFtpOpen(pass *analysis.Pass) (interface{}, error) {
	urlVarMap := make(map[string]bool)
	odVarMap := make(map[string]bool)

	// check if url exists as function parameters
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() != "function_definition" {
			return
		}

		// check default parameter variable
		funcParams := node.ChildByFieldName("parameters")
		if funcParams.Type() != "parameters" {
			return
		}

		if funcParams.NamedChildCount() > 0 {
			parameterNodes := getNamedChildren(funcParams, 0)

			for _, paramNode := range parameterNodes {
				ok, name := isInsecureUrl(paramNode, pass.FileContext.Source)
				if ok {
					urlVarMap[name] = true
				}
			}
		}
	})

	// check for insecure assignments
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() == "assignment" {
			leftNode := node.ChildByFieldName("left")
			rightNode := node.ChildByFieldName("right")

			varContent := rightNode.Content(pass.FileContext.Source)

			// check if ftp url is stored in a variable
			if isFtpUrl(trimQuotes(varContent)) {
				urlVarMap[leftNode.Content(pass.FileContext.Source)] = true
			}

			// check if OpenerDirector instance is created
			if strings.Contains(varContent, "OpenerDirector") || strings.Contains(varContent, "Request") || strings.Contains(varContent, "URLopener") {
				odVarMap[leftNode.Content(pass.FileContext.Source)] = true
			}
		}
	})

	// direct passing of ftp urls
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() != "call" {
			return
		}

		functionNode := node.ChildByFieldName("function")

		if !isOpenDirectorCall(functionNode, pass.FileContext.Source, odVarMap) {
			return
		}

		argNodes := node.ChildByFieldName("arguments")
		if argNodes != nil && argNodes.NamedChildCount() > 0 {
			urlNode := argNodes.NamedChild(0)

			if urlNode.Type() == "string" {
				urlString := trimQuotes(urlNode.Content(pass.FileContext.Source))
				if isFtpUrl(urlString) {
					pass.Report(pass, node, "Unsecured FTP connection detected. Data is unencrypted—use SFTP instead.")
				}
			} else if urlNode.Type() == "identifier" {
				urlVarName := urlNode.Content(pass.FileContext.Source)
				if urlVarMap[urlVarName] {
					pass.Report(pass, node, "Unsecured FTP connection detected. Data is unencrypted—use SFTP instead.")
				}
			}

		}
	})

	return nil, nil
}

func isOpenDirectorCall(node *sitter.Node, source []byte, odVarMap map[string]bool) bool {
	funcContent := node.Content(source)
	if strings.Contains(funcContent, "OpenerDirector") || strings.Contains(funcContent, "Request") || strings.Contains(funcContent, "urlopen") || strings.Contains(funcContent, "URLopener") || strings.Contains(funcContent, "urlretrieve") {
		return true
	}

	for odVar := range odVarMap {
		if strings.Contains(funcContent, odVar) {
			return true
		}
	}

	return false
}

// get all the children nodes of a node
func getNamedChildren(node *sitter.Node, startIdx int) []*sitter.Node {
	var namedChildren []*sitter.Node
	childrenCount := node.NamedChildCount()

	for i := startIdx; i < int(childrenCount); i++ {
		namedChildren = append(namedChildren, node.NamedChild(i))
	}

	return namedChildren
}

// check if url is an insecure fpt link
func isInsecureUrl(node *sitter.Node, source []byte) (bool, string) {
	if node.Type() == "default_parameter" {
		paramNameNode := node.ChildByFieldName("name")
		paramValueNode := node.ChildByFieldName("value")

		paramName := paramNameNode.Content(source)
		urlString := trimQuotes(paramValueNode.Content(source))

		if isFtpUrl(urlString) {
			return true, paramName
		}
		return false, ""

	}

	return false, ""
}

func isFtpUrl(s string) bool {
	pattern := `(?i)^ftp://.*`
	return regexp.MustCompile(pattern).MatchString(s)
}

func trimQuotes(s string) string {
	return strings.Trim(s, "\"")
}
