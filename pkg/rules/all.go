package rules

import (
	"github.com/srijan-paul/deepgrep/pkg/one"
	python_rules "github.com/srijan-paul/deepgrep/pkg/rules/python"
)

// CreateRules creates a base ruleset for each supported language
func CreateRules() map[one.Language][]one.Rule {
	return map[one.Language][]one.Rule{
		one.LangPy: python_rules.CreatePyRules(),
	}
}
