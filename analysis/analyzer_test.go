package analysis

import (
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func parseTestFile(t *testing.T, filename string, source string, language Language) *ParseResult {
	parsed, err := Parse(filename, []byte(source), language, language.Grammar())
	require.NoError(t, err)
	require.NotNil(t, parsed)
	return parsed
}

// mock checker function for unit testing
func mockChecker(pass *Pass) (interface{}, error) {
	Preorder(pass, func(node *sitter.Node) {
		if node.Type() == "assert_statement" {
			pass.Report(pass, node, "Assert not allowed")
		}
	})

	return nil, nil
}

func TestSkipCqComment(t *testing.T) {
	tests := []struct {
		name      string
		checkerId string
		filename  string
		source    string
		language  Language
		checker   func(*Pass) (any, error)
		want      bool
	}{
		{
			name:      "skipcq on same line",
			checkerId: "no-assert",
			filename:  "no-assert.test.py",
			source: `
				def someFunc(a, b):
					assert a == b # skipcq
			`,
			language: LangPy,
			checker:  mockChecker,
			want:     true,
		},
		{
			name:      "skipcq on previous line",
			checkerId: "no-assert",
			filename:  "no-assert.test.py",
			source: `
				if True:
					# skipcq
					assert a == 10
			`,
			language: LangPy,
			checker:  mockChecker,
			want:     true,
		},
		{
			name:      "skipcq with target checker name",
			checkerId: "no-assert",
			filename:  "no-assert.test.py",
			language:  LangPy,
			source: `
				if a >= 20:
					# skipcq: no-assert
					assert a < 30
			`,
			want: true,
		},
		{
			name:      "skipcq with mismatched target name",
			checkerId: "no-assert",
			filename:  "no-assert.test.py",
			language:  LangPy,
			source: `
				try:
					# skipcq: sql-injection
					assert a == b
				except AssertionError as e:
					print(e)
			`,
			want: false,
		},
		{
			name:      "skipcq comment not present",
			checkerId: "no-assert",
			filename:  "no-assert.test.py",
			source: `
				assert 1 == 2
			`,
			language: LangPy,
			checker:  mockChecker,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed := parseTestFile(t, tt.filename, tt.source, tt.language)
			analyzer := &Analyzer{
				Name:        tt.checkerId,
				Description: "analyzer for testing",
				Category:    CategorySecurity,
				Severity:    SeverityWarning,
				Language:    tt.language,
				Requires:    []*Analyzer{},
				Run:         tt.checker,
			}

			var reportNode *sitter.Node

			pass := &Pass{
				Analyzer:    analyzer,
				FileContext: parsed,
				Files:       []*ParseResult{parsed},
				Report: func(p *Pass, n *sitter.Node, msg string) {
					reportNode = n
				},
			}

			_, err := mockChecker(pass)
			require.NoError(t, err)
			require.NotNil(t, reportNode, "checker should return a non-nil node for assert")

			issue := &Issue{
				Id:       &tt.checkerId,
				Filepath: "no-assert.test.py",
				Node:     reportNode,
			}

			skipLines := GatherSkipInfo(parsed)

			result := ContainsSkipcq(skipLines, issue)
			assert.Equal(t, tt.want, result)
		})
	}
}
