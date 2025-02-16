package js_rules

import "globstar.dev/globstar/pkg/analysis"

// CreateJsRules returns a list of all python rules
func CreateJsRules() []analysis.Rule {
	return []analysis.Rule{
		NoDoubleEq(),
		UnusedImport(),
	}
}
