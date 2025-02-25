package golang

import (
	"strconv"

	sitter "github.com/smacker/go-tree-sitter"
	"globstar.dev/analysis"
)

var WeakPbkdf2Iterations analysis.Analyzer = analysis.Analyzer{
	Name:        "weak-pbkdf2-iterations",
	Language:    analysis.LangGo,
	Description: "This rule checks for PBKDF2 usage with fewer than 310,000 iterations.",
	Category:    analysis.CategorySecurity,
	Severity:    analysis.SeverityWarning,
	Run:         weakPbkdf2Iterations,
}

func weakPbkdf2Iterations(pass *analysis.Pass) (interface{}, error) {
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() != "call_expression" {
			return
		}

		function := node.ChildByFieldName("function")
		if function == nil || function.Type() != "selector_expression" {
			return
		}

		operand := function.ChildByFieldName("operand")
		field := function.ChildByFieldName("field")
		if operand == nil || field == nil {
			return
		}

		source := pass.FileContext.Source
		if operand.StartByte() >= uint32(len(source)) || operand.EndByte() > uint32(len(source)) {
			return
		}
		if field.StartByte() >= uint32(len(source)) || field.EndByte() > uint32(len(source)) {
			return
		}

		operandContent := string(source[operand.StartByte():operand.EndByte()])
		fieldContent := string(source[field.StartByte():field.EndByte()])

		if operandContent == "pbkdf2" && fieldContent == "Key" {
			arguments := node.ChildByFieldName("arguments")
			if arguments == nil || arguments.Type() != "argument_list" || arguments.NamedChildCount() < 5 {
				return
			}

			// Iterations is the 3rd argument (index 2)
			iterNode := arguments.NamedChild(2)
			if iterNode == nil || iterNode.StartByte() >= uint32(len(source)) || iterNode.EndByte() > uint32(len(source)) {
				return
			}

			iterContent := string(source[iterNode.StartByte():iterNode.EndByte()])
			iterations, err := strconv.ParseInt(iterContent, 10, 64)
			if err != nil {
				return
			}

			if iterations < 310000 {
				pass.Report(pass, node, "PBKDF2: Use at least 310,000 iterations for security (OWASP 2023 recommendation)")
			}
		}
	})
	return nil, nil
}
