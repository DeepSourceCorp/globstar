package analysis

import (
	"fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

type goExtractor struct {
	filePath string
	source []byte
	functions []Function
	calls []Call
	packageName string
	currentFunc *BasicFunction
}

func ParseGoFile(filePath string) ([]Function, []Call, error) {
	if !strings.HasSuffix(filePath, ".go") {
		return []Function{}, []Call{}, nil
	}
	parseResult, err := ParseFile(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse Go file %s: %w", filePath, err)
	}

	extractor := &goExtractor{
		filePath: filePath,
		source: parseResult.Source,
		functions: make([]Function, 0),
		calls: make([]Call, 0),
	}

	// first pass
	extractor.extractFunctionsOnly(parseResult.Ast)

	// second pass
	extractor.extractCallsOnly(parseResult.Ast)

	return extractor.functions, extractor.calls, nil
}

func (e *goExtractor) extractFunctionsOnly(node *sitter.Node) {
	switch node.Type() {
	case "package_clause":
		e.extractPackage(node)
	case "function_declaration":
		e.extractFunction(node)
	case "method_declaration":
		e.extractMethod(node)
	}

	for i := 0; i < int(node.ChildCount()); i++ {
		e.extractFunctionsOnly(node.Child(i))
	}
}

func (e *goExtractor) extractCallsOnly(node *sitter.Node) {
	switch node.Type() {
	case "function_declaration", "method_declaration":
		e.setCurrentFunction(node)
		
		// extract calls from the function body
		for i := 0; i < int(node.ChildCount()); i++ {
			e.extractCallsFromBody(node.Child(i))
		}
		e.currentFunc = nil
	}

	for i := 0; i < int(node.ChildCount()); i++ {
		e.extractCallsOnly(node.Child(i))
	}
}

func (e *goExtractor) extractCallsFromBody(node *sitter.Node) {
	if node.Type() == "call_expression" {
		e.extractCall(node)
	}

	for i := 0; i < int(node.ChildCount()); i++ {
		e.extractCallsFromBody(node.Child(i))
	}
}

func (e *goExtractor) setCurrentFunction(node *sitter.Node) {
	funcName := ""
	if node.Type() == "function_declaration" {
		if nameNode := node.ChildByFieldName("name"); nameNode != nil {
			funcName = nameNode.Content(e.source)
		}
	} else if node.Type() == "method_declaration" {
		if nameNode := node.ChildByFieldName("name"); nameNode != nil {
			funcName = nameNode.Content(e.source)
		}
	}

	for _, fn := range e.functions {
		if fn.Name() == funcName {
			if bf, ok := fn.(*BasicFunction); ok {
				e.currentFunc = bf
				break
			}
		}
	}
}

func (e *goExtractor) extractPackage(node *sitter.Node) {
	nameNode := node.Child(0)
	if nameNode != nil && nameNode.Type() == "package_identifier" {
		e.packageName = nameNode.Content(e.source)
	}
}


/*
for functions defined as

func a() {}
*/
func (e *goExtractor) extractFunction(node *sitter.Node) {
	funcNameNode := node.ChildByFieldName("name")
	if funcNameNode == nil {
		return
	}
	funcName := funcNameNode.Content(e.source)
	qualifiedName := e.packageName + "." + funcName
	funcLineNo := int(node.StartPoint().Row) + 1

	function := &BasicFunction{
		Id: fmt.Sprintf("go:%s:%s:%d", e.filePath, qualifiedName, funcLineNo),
		Name_: funcName,
		QualifiedName_: qualifiedName,
		Filepath_: e.filePath,
		LineNumber_: funcLineNo,
		IsThirdParty_: e.isThirdPartyFile(),
	}

	e.functions = append(e.functions, function)
	e.currentFunc = function
}


/*
for interface or struct methods
func (s *someStruct) myFunc() {}
*/
func (e *goExtractor) extractMethod(node *sitter.Node) {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return
	}

	methodName := nameNode.Content(e.source)
	receiverType := e.extractReceiverType(node)
	line := int(node.StartPoint().Row) + 1
	var qualifiedName string
	if receiverType != "" {
		qualifiedName = fmt.Sprintf("%s.(%s).%s", e.packageName, receiverType, methodName)
	} else {
		qualifiedName = e.packageName + "." + methodName
	}

	function := &BasicFunction{
		Id: fmt.Sprintf("go:%s:%s:%d", e.filePath, qualifiedName, line),
		Name_: methodName,
		QualifiedName_: qualifiedName,
		Filepath_: e.filePath,
		LineNumber_: line,
		IsThirdParty_: e.isThirdPartyFile(),
	}

	e.functions = append(e.functions, function)
	e.currentFunc = function
}

func (e *goExtractor) extractReceiverType(methodNode *sitter.Node) string {
	receiverNode := methodNode.ChildByFieldName("receiver")
	if receiverNode == nil {
		return ""
	}

	for i := 0; i < int(receiverNode.ChildCount()); i++ {
		child := receiverNode.Child(i)
		if child.Type() == "parameter_list" {
			for j := 0; j < int(child.ChildCount()); j++ {
				param := child.Child(j)
				if param.Type() == "parameter_declaration" {
					typeNode := param.ChildByFieldName("type")
					if typeNode != nil {
						type_id := typeNode.Child(0)
						return type_id.Content(e.source)
					}
				}
			}
		}
	}

	return ""
}

func (e *goExtractor) extractCall(node *sitter.Node) {
	if e.currentFunc == nil {
		return
	}

	functionNode := node.ChildByFieldName("function")
	if functionNode == nil {
		return
	}


	calleeName := e.extractCallName(functionNode)
	if calleeName == "" {
		return
	}


	line := int(node.StartPoint().Row) + 1
	var callee Function
	found := false

	for _, existingFunc := range e.functions {
		if existingFunc.Name() == calleeName {
			callee = existingFunc
			found = true
			break
		}
	}

	if !found {
		callee = &BasicFunction{
			Id: fmt.Sprintf("go:unknown:%s:-1", calleeName),
			Name_: calleeName,
			QualifiedName_: calleeName,
			Filepath_: "unknown",
			LineNumber_: -1,
			IsThirdParty_: e.isThirdPartyFile(),
		}

		e.functions = append(e.functions, callee)
	}

	call := &BasicCall{
		CallerFunc: e.currentFunc,
		CalleeFunc: callee,
		SourceFile: e.filePath,
		SourceLine: fmt.Sprintf("%d", line),
	}

	e.calls = append(e.calls, call)
}

func (e *goExtractor) extractCallName(node *sitter.Node) string {
	switch node.Type() {
	case "identifier":
		return node.Content(e.source)
	case "selector_expression":
		var parts []string
		e.collectSelectorParts(node, &parts)
		result := strings.Join(parts, ".")
		return result

	default:
		content := node.Content(e.source)
		return content

	}
}

func (e *goExtractor) collectSelectorParts(node *sitter.Node, parts *[]string) {
	if node.Type() == "selector_expression" {
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			childType := child.Type()
			childContent := child.Content(e.source)

			switch childType {
			case "identifier":
				*parts = append(*parts, childContent)
			case "selector_expression":
				e.collectSelectorParts(child, parts)
			case "field_identifier":
				*parts = append(*parts, childContent)
			}
		}
	} else if node.Type() == "identifier" {
		*parts = append(*parts, node.Content(e.source))
	}
}

func (e *goExtractor) isThirdPartyFile() bool {
	return strings.Contains(e.filePath, "/vendor/") || strings.Contains(e.filePath, "/pkg/mod/")
}