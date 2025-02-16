package rules

import (
	"globstar.dev/globstar/pkg/analysis"
	js_rules "globstar.dev/globstar/pkg/rules/js"
	python_rules "globstar.dev/globstar/pkg/rules/python"
)

// CreateBaseRuleMap creates a base ruleset for each supported language
func CreateBaseRuleMap() map[analysis.Language][]analysis.Rule {
	jsRules := js_rules.CreateJsRules()
	return map[analysis.Language][]analysis.Rule{
		analysis.LangPy:  python_rules.CreatePyRules(),
		analysis.LangJs:  jsRules,
		analysis.LangTsx: jsRules,
		analysis.LangTs:  jsRules,
	}
}
