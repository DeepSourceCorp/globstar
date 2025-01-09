package rules

import (
	"fmt"
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/srijan-paul/deepgrep/pkg/one"
	"github.com/stretchr/testify/require"
)

type ExpectedIssue struct {
	Message string
	Start   *sitter.Point
	End     *sitter.Point
}

type ShouldRaise struct {
	Code     string
	Expected []ExpectedIssue
}

type TestCase struct {
	Name  string
	Rule  one.Rule
	Raise []ShouldRaise
	Pass  []string
}

func serializeIssues(issues []*one.Issue) []string {
	issues_str := []string{}
	for _, issue := range issues {
		issues_str = append(issues_str, fmt.Sprintf("[%d:%d-%d:%d] %s",
			issue.Range.StartPoint.Row,
			issue.Range.StartPoint.Column,
			issue.Range.EndPoint.Row,
			issue.Range.EndPoint.Column,
			issue.Message,
		))
	}
	
	return issues_str 
}

func (testCase *TestCase) Run(t *testing.T) {
	for _, shouldRaise := range testCase.Raise {
		lang := testCase.Rule.GetLanguage()
		grammar := lang.Grammar()

		parseResult, err := one.Parse(testCase.Name, []byte(shouldRaise.Code), lang, grammar)
		require.NoError(t, err)
		require.NotNil(t, parseResult)

		analyzer := one.NewAnalyzer(parseResult, []one.Rule{testCase.Rule})
		require.NotNil(t, analyzer)
		analyzer.AddRule(testCase.Rule)

		got := analyzer.Analyze()
		if len(got) != len(shouldRaise.Expected) {
			t.Errorf("expected %d issues, but got %d. Snippet:\n%s\n", len(shouldRaise.Expected), len(got), shouldRaise.Code)
		}

		for _, want := range shouldRaise.Expected {
			found := false
			for _, issue := range got {
				messagesMatch := issue.Message == want.Message
				startsMatch := want.Start == nil || issue.Range.StartPoint == *want.Start
				endsMatch := want.End == nil || issue.Range.EndPoint == *want.End

				if messagesMatch && startsMatch && endsMatch {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("expected issue not found: %v.\nIssues raised: %v", want, serializeIssues(got))
			}
		}
	}

	for _, pass := range testCase.Pass {
		lang := testCase.Rule.GetLanguage()
		grammar := lang.Grammar()

		parseResult, err := one.Parse(testCase.Name, []byte(pass), lang, grammar)
		require.NoError(t, err)
		require.NotNil(t, parseResult)

		analyzer := one.NewAnalyzer(parseResult, []one.Rule{testCase.Rule})
		require.NotNil(t, analyzer)
		analyzer.AddRule(testCase.Rule)

		got := analyzer.Analyze()
		if len(got) > 0 {
			t.Errorf("expected no issues, but got: %v", serializeIssues(got))
		}
	}
}
