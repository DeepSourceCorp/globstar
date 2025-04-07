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

func TestSkipCq(t *testing.T) {
	tests := []struct{
		name string
		source string
		language Language
		want bool
	}{
		{
			name: "skipcq comment on same line",
			language: LangPy,
			source: `
				def someFunc(a, b):
					assert a == b # <skipcq>
			`,
			want: true,
		},
		{
			name: "skipcq comment on previous line",
			language: LangPy,
			source: `
				if True:
				# <skipcq>
					assert 1 == 2
			`,
			want: true,
		},
		{
			name: "skipcq comment not present",
			language: LangPy,
			source: `
				assert a == b
			`,
			want: false,
		},

	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed := parseTestFile(t, "no-assert.test.py", tt.source, tt.language)
			analyzer := &Analyzer{
				Language: tt.language,
				ParseResult: parsed,
			}

			query, err := sitter.NewQuery([]byte("(assert_statement) @assert"), tt.language.Grammar())
			require.NoError(t, err)

			cursor := sitter.NewQueryCursor()
			cursor.Exec(query, parsed.Ast)

			match, ok := cursor.NextMatch()
			require.True(t, ok, "failed to find assert statements")

			var assertNode *sitter.Node
			for _, captureNode := range match.Captures {
				if query.CaptureNameForId(captureNode.Index) == "assert" {
					assertNode = captureNode.Node
					break
				}
			}

			require.NotNil(t, assertNode, "failed to capture assert node")

			issue := &Issue{
				Filepath: "no-assert.test.py",
				Node: assertNode,
			}

			res := analyzer.ContainsSkipcq(issue)
			assert.Equal(t, tt.want, res)
		})
	}
}