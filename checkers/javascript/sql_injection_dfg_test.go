package javascript

import (
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ana "globstar.dev/analysis"
)

func parseJsCode(t *testing.T, source []byte) *ana.ParseResult {
	pass, err := ana.Parse("", source, ana.LangJs, ana.LangJs.Grammar())
	require.NoError(t, err)

	return pass
}

func TestSQLInjection(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		wantVuln bool
	}{

		// Vanilla implementation gives them as vulnerable
		{
			name: "hard_coded_inputs",
			code: `
				const limit = 10;
				db.query(` + "`" + "SELECT * FROM users LIMIT ${limit}" + "`" + `);
			`,
			wantVuln: false,
		},
		{
			name: "non_sql_values",
			code: `
				const message = "Hello," + username;
				db.query("SELECT * FROM logs WHERE message = ?", [message]);
			`,
			wantVuln: false,
		},
		// Falsely Detected by non-DFA implementation of sql_injection
		{
			name: "direct_user_input",
			code: `
				const userId = req.body.id;
				db.query("SELECT * FROM users WHERE id = " + userId);
			`,
			wantVuln: true,
		},
		// Falselly flagged by non-DFA implmentation
		{
			name: "sanitized_input",
			code: `
				const userId = escape(req.body.id);
				db.query("SELECT * FROM users WHERE id = " + userId);
			`,
			wantVuln: false,
		},
		{
			name: "template_literal_injection",
			code: `
				const name = req.query.name;
				connection.query(` + "`" + "SELECT * FROM users WHERE name = ${name}" + "`" + `);
			`,
			wantVuln: true,
		},
		{
			name: "indirect_taint_flow",
			code: `
				const input = req.params.id;
				const temp = input;
				const query = "SELECT * FROM users WHERE id = " + temp;
				db.query(query);
			`,
			wantVuln: true,
		},
		{
			name: "safe_parameterized_query",
			code: `
				const input = req.params.id;
				const sanitized = parameterize(input);
				db.query("SELECT * FROM users WHERE id = ?", sanitized);
			`,
			wantVuln: false,
		},
		// DFG Implementation detects only one of them as vulnerable, whereas the basic implementation flags both as vulnerable.
		{
			name: "multiple_sanitization_paths",
			code: `
				const input = req.body.data;
				const path1 = sanitize(input);
				const path2 = input;
				db.query("SELECT * FROM users WHERE data = " + path1);
				db.query("SELECT * FROM users WHERE data = " + path2);
			`,
			wantVuln: true,
		},
		{
			name: "binary_expression_chaining",
			code: `
				const a = req.query.a;
				const b = "static";
				const c = a + b + "more";
				db.query("SELECT * FROM table WHERE col = " + c);
			`,
			wantVuln: true,
		},
		{
			name: "nested_template_strings",
			code: `
				const user = req.body.user;
				const role = req.body.role;
				db.query(` + "`" + "SELECT * FROM users WHERE name = ${user} AND role = ${escape(role)}" + "`" + `);
			`,
			wantVuln: true, // Should detect unsanitized user variable
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parseResult := parseJsCode(t, []byte(tt.code))
			var reportedVulns []struct {
				Pass *ana.Pass
				Node *sitter.Node
				Msg  string
			}
			pass := &ana.Pass{
				Analyzer:    SQLInjection,
				FileContext: parseResult,
				Report: func(p *ana.Pass, n *sitter.Node, msg string) {
					reportedVulns = append(reportedVulns, struct {
						Pass *ana.Pass
						Node *sitter.Node
						Msg  string
					}{p, n, msg})
				},
			}

			_, err := detectSQLInjection(pass)

			require.NoError(t, err)

			t.Logf("detectedVulnerabilities: %v\n", len(reportedVulns))
			if tt.wantVuln {
				assert.Greater(t, len(reportedVulns), 0, "Expected to detect sql injection vulnerability")
			} else {
				assert.Equal(t, 0, len(reportedVulns), "Unexpected Sql injection vuln. reported")
			}

		})

		t.Run(tt.name+" DFG", func(t *testing.T) {
			parseResult := parseJsCode(t, []byte(tt.code))
			var reportedVulns []struct {
				Pass *ana.Pass
				Node *sitter.Node
				Msg  string
			}
			pass := &ana.Pass{
				Analyzer:    SQLInjectionDFG,
				FileContext: parseResult,
				Report: func(p *ana.Pass, n *sitter.Node, msg string) {
					reportedVulns = append(reportedVulns, struct {
						Pass *ana.Pass
						Node *sitter.Node
						Msg  string
					}{p, n, msg})
				},
			}

			_, err := detectSQLInjectionDFG(pass)

			require.NoError(t, err)
			t.Logf("Detected Vuln: %v\n", len(reportedVulns))

			if tt.wantVuln {
				assert.Greater(t, len(reportedVulns), 0, "Expected to detect sql injection vulnerability")
			} else {
				assert.Equal(t, 0, len(reportedVulns), "Unexpected Sql injection vuln. reported")
			}

		})
	}
}
