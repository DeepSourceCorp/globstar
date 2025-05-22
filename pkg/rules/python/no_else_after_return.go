package python_rules

import (
	"github.com/DeepSourceCorp/globstar/pkg/analysis"
	sitter "github.com/smacker/go-tree-sitter"
)

// isReturnStatement checks if a node is a return statement.
func isReturnStatement(node *sitter.Node) bool {
	return node.Type() == "return_statement"
}

// blockAlwaysReturns checks if a block of statements always ends in a return.
// This is a simplified check: it looks for a return statement as one of
// the direct children of the block. More complex analysis (e.g. if/else within
// the block where both branches return) is not handled by this initial version.
func blockAlwaysReturns(blockNode *sitter.Node) bool {
	if blockNode == nil || blockNode.Type() != "block" {
		return false
	}

	// Iterate over statements in the block from last to first relevant one
	// Skipping non-code nodes like comments if any.
	// We are looking for a return statement.
	// A more sophisticated check would analyze all control flow paths.
	childCount := int(blockNode.NamedChildCount())
	if childCount == 0 {
		return false
	}

    // Check the last statement(s) of the block.
    // For now, we simplify and check if the *last* named child is a return statement.
    // TODO: Handle cases like `if cond: if inner_cond: return; else: return;`
    // TODO: Handle `try/except/finally` where all paths return.
	lastStatement := blockNode.NamedChild(childCount - 1)
	if lastStatement != nil && isReturnStatement(lastStatement) {
		return true
	}
    // A slightly more robust check: if the block is just one statement and it's a return
    if childCount == 1 && isReturnStatement(blockNode.NamedChild(0)) {
        return true
    }


	return false
}

func checkElseAfterReturn(r analysis.Rule, ana *analysis.Analyzer, node *sitter.Node) {
	if node.Type() != "if_statement" {
		return
	}

	consequenceNode := node.ChildByFieldName("consequence") // This is the 'if' block
	alternativeNode := node.ChildByFieldName("alternative") // This could be an 'elif' or 'else' block

	if alternativeNode == nil {
		return // No 'else' or 'elif', so rule doesn't apply
	}

	// We are interested in 'else' blocks, not 'elif'
	// An 'else' clause is an "else_clause" node containing a "block".
	// An 'elif' is another "if_statement" node.
	firstChildOfAlternative := alternativeNode.NamedChild(0)
	if firstChildOfAlternative == nil || firstChildOfAlternative.Type() != "block" {
		// This means it's likely an 'elif' (which is an if_statement itself)
		// or some other structure we are not targeting with this specific 'else' rule.
		return
	}


	if blockAlwaysReturns(consequenceNode) {
		ruleID := "py-no-else-after-return" // Define the rule ID string
		ana.Report(&analysis.Issue{
			Message: "Unnecessary 'else' after 'return'. The 'if' block always returns.",
			Range:   alternativeNode.Range(), // Highlight the 'else' keyword or the 'else' block
			Id:      &ruleID,                 // Pass the address of the rule ID string
		})
	}
}

func NoElseAfterReturn() analysis.Rule {
	var entry analysis.VisitFn = checkElseAfterReturn
	// The "if_statement" node type is the correct first argument for CreateRule.
	// The rule ID "py-no-else-after-return" is used when reporting an issue.
	return analysis.CreateRule("if_statement", analysis.LangPy, &entry, nil)
}
