package js_rules

import "github.com/DeepSourceCorp/globstar/pkg/analysis"

// CreateJsRules returns a list of all python rules
func CreateJsRules() []analysis.Rule {
	return []analysis.Rule{
		NoDoubleEq(),
		UnusedImport(),
	}
}
