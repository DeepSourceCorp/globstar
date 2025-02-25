package golang

import (
	"strconv"

	sitter "github.com/smacker/go-tree-sitter"
	"globstar.dev/analysis"
)

var WeakBcryptCost analysis.Analyzer = analysis.Analyzer{
	Name:        "weak-bcrypt-cost",
	Language:    analysis.LangGo,
	Description: "This rule checks for bcrypt usage with cost factor less than 10.",
	Category:    analysis.CategorySecurity,
	Severity:    analysis.SeverityWarning,
	Run:         weakBcryptCost,
}

func weakBcryptCost(pass *analysis.Pass) (interface{}, error) {
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

		if operandContent == "bcrypt" && fieldContent == "GenerateFromPassword" {
			arguments := node.ChildByFieldName("arguments")
			if arguments == nil || arguments.Type() != "argument_list" || arguments.NamedChildCount() < 2 {
				return
			}

			// Cost is the 2nd argument (index 1)
			costNode := arguments.NamedChild(1)
			if costNode == nil || costNode.StartByte() >= uint32(len(source)) || costNode.EndByte() > uint32(len(source)) {
				return
			}

			costContent := string(source[costNode.StartByte():costNode.EndByte()])
			cost, err := strconv.ParseInt(costContent, 10, 64)
			if err != nil {
				return
			}

			if cost < 10 {
				pass.Report(pass, node, "bcrypt: Cost factor should be at least 10 (detected "+costContent+")")
			}
		}
	})
	return nil, nil
}
