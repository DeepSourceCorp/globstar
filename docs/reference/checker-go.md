# Checker Go Interface

While YAML checkers are great for simple pattern matching, sometimes you need more sophisticated analysis. Globstar allows you to write checkers in Go that have full access to the tree-sitter AST and can perform complex validations.

::: warning NOTE
We have some buitlin checkers written in Go, but we're working on making the Go API more user-friendly. This is why it's not yet documented in detail, and not available in the current release. Please stay tuned for updates!
:::

## Example Checker

Here's a simple checker that prevents using tuples in if conditions in Python (which always evaluate to True):

```go
package python_rules

import (
    "github.com/DeepSourceCorp/globstar/pkg/analysis"
    sitter "github.com/smacker/go-tree-sitter"
)

func checkIfTuple(r analysis.Rule, ana *analysis.Analyzer, node *sitter.Node) {
    cond := node.ChildByFieldName("condition")
    if cond.Type() == "tuple" {
        ana.Report(&analysis.Issue{
            Message: "Do not use 'if' with a tuple argument, as tuples always evaluate to true",
            Range:   node.Range(),
        })
    }
}

func IfTuple() analysis.Rule {
    var entry analysis.VisitFn = checkIfTuple
    return analysis.CreateRule("if_statement", analysis.LangPy, &entry, nil)
}
```

This checker demonstrates the key components of writing Go checkers:
1. Node type checking (`if_statement`)
2. Field access (`ChildByFieldName("condition")`)
3. Issue reporting with a descriptive message