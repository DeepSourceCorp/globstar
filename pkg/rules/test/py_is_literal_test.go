package rules

import (
	"testing"
	py_rules "github.com/srijan-paul/deepgrep/pkg/rules/python"
)

func TestIsLiteral(t *testing.T) {
	rule := py_rules.IsLiteral()
	expected := []ExpectedIssue{{Message: "Do not use 'is' to compare literals. Use '==' instead"}}

	testCase := &TestCase{
		Name: "is-literal",
		Rule: rule,
		Raise: []ShouldRaise{
			{
				Code: `
				def foo():
					if x is 42:
						pass`,

				Expected: expected,
			},

			{
				Code: "x = bar is 'baz'",
				Expected: expected,
			},

			{
				Code: "x = 4.0 is x",
				Expected: expected,
			},
		},
		Pass: []string{
			`
			def foo():
			  if x is y:
					pass
			`,

			`
			def foo():
				if x == 42:
					pass
			`,
		},
	}

	testCase.Run(t)
}

