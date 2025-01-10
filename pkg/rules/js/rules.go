package js_rules

import "github.com/srijan-paul/deepgrep/pkg/one"

// CreateJsRules returns a list of all python rules
func CreateJsRules() []one.Rule {
	return []one.Rule{
		NoDoubleEq(),
	}
}
