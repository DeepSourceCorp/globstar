package golang

import (
	"strconv"

	sitter "github.com/smacker/go-tree-sitter"
	"globstar.dev/analysis"
)

var WeakScryptCost analysis.Analyzer = analysis.Analyzer{
	Name:        "weak-scrypt-cost",
	Language:    analysis.LangGo,
	Description: "This rule checks for scrypt usage with N parameter less than 32768.",
	Category:    analysis.CategorySecurity,
	Severity:    analysis.SeverityWarning,
	Run:         weakScryptCost,
}

func weakScryptCost(pass *analysis.Pass) (interface{}, error) {
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

		if operandContent == "scrypt" && fieldContent == "Key" {
			arguments := node.ChildByFieldName("arguments")
			if arguments == nil || arguments.Type() != "argument_list" || arguments.NamedChildCount() < 6 {
				return
			}

			// N parameter is the 3rd argument (index 2)
			nNode := arguments.NamedChild(2)
			if nNode == nil || nNode.StartByte() >= uint32(len(source)) || nNode.EndByte() > uint32(len(source)) {
				return
			}

			nContent := string(source[nNode.StartByte():nNode.EndByte()])
			n, err := strconv.ParseInt(nContent, 10, 64)
			if err != nil {
				return
			}

			if n < 32768 {
				pass.Report(pass, node, "scrypt: N parameter should be ≥ 32768 (detected "+nContent+")")
			}
		}
	})
	return nil, nil
}
