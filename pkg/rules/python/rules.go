package python_rules

import "github.com/DeepSourceCorp/globstar/pkg/analysis"

// CreatePyRules returns a list of all python rules
func CreatePyRules() []analysis.Rule {
	return []analysis.Rule{
		IsLiteral(),
		IfTuple(),
		NoElseAfterReturn(),
	}
}
