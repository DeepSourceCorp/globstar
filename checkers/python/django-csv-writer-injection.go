package python

import (
	"regexp"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"globstar.dev/analysis"
)

var CsvWriterInjection *analysis.Analyzer = &analysis.Analyzer{
	Name:        "csv-writer-injection",
	Language:    analysis.LangPy,
	Description: "Using the built-in csv module with user input can allow attackers to inject malicious formulas into CSV files. When opened in a spreadsheet, these formulas may execute scripts that steal data or install malware. To mitigate this risk, use defusedcsv, a drop-in replacement that helps prevent formula injection.",
	Category:    analysis.CategorySecurity,
	Severity:    analysis.SeverityWarning,
	Run:         checkCsvWriterInjection,
}

func checkCsvWriterInjection(pass *analysis.Pass) (interface{}, error) {
	writerVarMap := make(map[string]bool)
	userDataVarMap := make(map[string]bool)

	// get the variable names of the csv writer instance
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() != "assignment" {
			return
		}

		leftNode := node.ChildByFieldName("left")
		rightNode := node.ChildByFieldName("right")

		if rightNode == nil {
			return
		}

		if isCsvWriter(rightNode, pass.FileContext.Source) {
			writerVarMap[leftNode.Content(pass.FileContext.Source)] = true
		}
	})

	// get the variable name from the Flask decorated route function
	analysis.Preorder(pass,func(node *sitter.Node) {
		if node.Type() != "decorated_definition" {
			return
		}
		decNode := node.NamedChild(0)
		if decNode.Type() != "decorator" {
			return
		}
		callNode := node.NamedChild(0)
		if callNode.Type() != "call" {
			return
		}
		funcNode := callNode.ChildByFieldName("function")
		if funcNode.Type() != "attribute" {
			return
		}
		if strings.HasSuffix(funcNode.Content(pass.FileContext.Source), ".route") {
			return
		}
		defNode := node.ChildByFieldName("function_definition")
		paramsNode := defNode.ChildByFieldName("parameters")
		if paramsNode.Type() != "parameters" {
			return
		}
		allparamNodes := getNamedChildren(paramsNode, 0)
		for _, p := range allparamNodes {
			userDataVarMap[p.Content(pass.FileContext.Source)] = true
		}


	})

	// get the variable names of user data or user formatted data
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

	// find the occurrences of this vulnerability
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() != "call" {
			return
		}

		funcNode := node.ChildByFieldName("function")

		if funcNode.Type() != "attribute" {
			return
		}

		funcObjNode := funcNode.ChildByFieldName("object")
		funcAttNode := funcNode.ChildByFieldName("attribute")

		if funcObjNode.Type() != "identifier" || funcAttNode.Type() != "identifier" {
			return
		}

		if !writerVarMap[funcObjNode.Content(pass.FileContext.Source)] {
			return
		}

		csvWriteMethodPat := `^(writerow|writerows|writeheader)$`
		re := regexp.MustCompile(csvWriteMethodPat)

		if !re.MatchString(funcAttNode.Content(pass.FileContext.Source)) {
			return
		}

		argListNode := node.ChildByFieldName("arguments")
		if argListNode.Type() != "argument_list" {
			return
		}

		csvWriterArgListNode := getNamedChildren(argListNode, 0)

		for _, argNode := range csvWriterArgListNode {
			if isUSerData(argNode, pass.FileContext.Source, userDataVarMap) {
				pass.Report(pass, node, "User input detected in CSV write method - potential formula injection vulnerability")
			}
		}

	})
	return nil, nil
}

func isUSerData(node *sitter.Node, source []byte, userDataVarMap map[string]bool) bool {
	switch node.Type() {
	case "call":
		if isRequestCall(node, source) && !strings.Contains(node.Content(source), "build_absolute_url") {
			return true
		}

		funcNode := node.ChildByFieldName("function")
		if funcNode.Type() != "attribute" {
			return false
		}

		if !strings.HasSuffix(funcNode.Content(source), ".format") {
			return false
		}

		argListNode := node.ChildByFieldName("arguments")
		if argListNode.Type() != "argument_list" {
			return false
		}

		allArgNodes := getNamedChildren(argListNode, 0)

		for _, argNode := range allArgNodes {
			if argNode.Type() == "identifier" && userDataVarMap[argNode.Content(source)] {
				return true
			}
		}

	case "string":
		if node.Content(source)[0] != 'f' {
			return false
		}

		stringChildrenNodes := getNamedChildren(node, 0)

		for _, strNode := range stringChildrenNodes {
			if strNode.Type() == "interpolation" {
				exprNode := strNode.ChildByFieldName("expression")
				if exprNode.Type() == "identifier" && userDataVarMap[exprNode.Content(source)] {
					return true
				}
			}
		}

	case "binary_operator":
		binaryOpStr := node.Content(source)

		for key := range userDataVarMap {
			pattern := `\b` + key + `\b`
			re := regexp.MustCompile(pattern)

			if re.MatchString(binaryOpStr) {
				return true
			}
		}

	case "identifier":
		return isRequestCall(node, source) || userDataVarMap[node.Content(source)]
	}

	return false
}

func isCsvWriter(node *sitter.Node, source []byte) bool {
	if node.Type() != "call" {
		return false
	}

	funcNode := node.ChildByFieldName("function")
	if funcNode.Type() != "attribute" {
		return false
	}

	funcObjNode := funcNode.ChildByFieldName("object")
	funcAttNode := funcNode.ChildByFieldName("attribute")

	if funcObjNode.Type() != "identifier" || funcAttNode.Type() != "identifier" {
		return false
	}

	if funcObjNode.Content(source) != "csv" || funcAttNode.Content(source) != "writer" {
		return false
	}

	return true
}
