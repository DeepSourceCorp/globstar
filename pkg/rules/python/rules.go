package python_rules

import "github.com/srijan-paul/deepgrep/pkg/one"

// CreatePyRules returns a list of all python rules
func CreatePyRules() []one.Rule {
	return []one.Rule{
		IsLiteral(),
		IfTuple(),
	}
}

