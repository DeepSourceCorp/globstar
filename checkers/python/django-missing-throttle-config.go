package python

import (
	sitter "github.com/smacker/go-tree-sitter"
	"globstar.dev/analysis"
)

var DjangoMissingThrottleConfig *analysis.Analyzer = &analysis.Analyzer{
	Name:        "django-missing-throttle-config",
	Language:    analysis.LangPy,
	Description: "Django REST framework is missing rate-limiting configurations, which could lead to resource starvation or DoS attacks. Add 'DEFAULT_THROTTLE_CLASSES' and 'DEFAULT_THROTTLE_RATES' to enforce rate limits.",
	Category:    analysis.CategorySecurity,
	Severity:    analysis.SeverityWarning,
	Run:         checkDjangoMissingThrottleConfig,
}

func checkDjangoMissingThrottleConfig(pass *analysis.Pass) (interface{}, error) {
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() != "assignment" {
			return
		}

		isThrottleConfigPresent := false

		leftNode := node.ChildByFieldName("left")
		rightNode := node.ChildByFieldName("right")

		if rightNode == nil {
			return
		}

		if leftNode.Type() != "identifier" {
			return
		}

		if leftNode.Content(pass.FileContext.Source) != "REST_FRAMEWORK" {
			return
		}

		if rightNode.Type() != "dictionary" {
			return
		}

		dictPairNodeList := getNamedChildren(rightNode, 0)

		for _, pairNode := range dictPairNodeList {
			if pairNode.Type() == "pair" {
				keyNode := pairNode.ChildByFieldName("key")
				if keyNode.Type() == "string" && keyNode.Content(pass.FileContext.Source)[0] != 'f' {
					stringContentNode := keyNode.NamedChild(1)
					if stringContentNode.Type() == "string_content" {
						if stringContentNode.Content(pass.FileContext.Source) == "DEFAULT_THROTTLE_RATES" {
							isThrottleConfigPresent = true
						}
					}
				}
			}
		}

		if !isThrottleConfigPresent {
			pass.Report(pass, node, "Missing rate-limiting in Django REST framework may allow DoS attacks")
		}

	})

	return nil, nil
}
