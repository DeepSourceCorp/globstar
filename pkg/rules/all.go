package rules

import (
	"github.com/srijan-paul/deepgrep/pkg/one"
	js_rules "github.com/srijan-paul/deepgrep/pkg/rules/js"
	python_rules "github.com/srijan-paul/deepgrep/pkg/rules/python"
)

// CreateBaseRuleMap creates a base ruleset for each supported language
func CreateBaseRuleMap() map[one.Language][]one.Rule {
	jsRules := js_rules.CreateJsRules()
	return map[one.Language][]one.Rule{
		one.LangPy:  python_rules.CreatePyRules(),
		one.LangJs:  jsRules,
		one.LangTsx: jsRules,
		one.LangTs:  jsRules,
	}
}
