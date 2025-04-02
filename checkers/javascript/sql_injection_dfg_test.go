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
		// {
		// 	name: "hard_coded_inputs",
		// 	code: `
		// 			const limit = 10;
		// 			db.query(` + "`" + "SELECT * FROM users LIMIT ${limit}" + "`" + `);
		// 		`,
		// 	wantVuln: false,
		// },
		// {
		// 	name: "non_sql_values",
		// 	code: `
		// 			const message = "Hello," + username;
		// 			db.query("SELECT * FROM logs WHERE message = ?", [message]);
		// 		`,
		// 	wantVuln: false,
		// },
		// // Falsely Detected by non-DFA implementation of sql_injection
		// {
		// 	name: "direct_user_input",
		// 	code: `
		// 			const userId = req.body.id;
		// 			db.query("SELECT * FROM users WHERE id = " + userId);
		// 		`,
		// 	wantVuln: true,
		// },
		// // Falselly flagged by non-DFA implmentation
		// {
		// 	name: "sanitized_input",
		// 	code: `
		// 			const userId = escape(req.body.id);
		// 			db.query("SELECT * FROM users WHERE id = " + userId);
		// 		`,
		// 	wantVuln: false,
		// },
		// {
		// 	name: "template_literal_injection",
		// 	code: `
		// 			const name = req.query.name;
		// 			connection.query(` + "`" + "SELECT * FROM users WHERE name = ${name}" + "`" + `);
		// 		`,
		// 	wantVuln: true,
		// },
		// {
		// 	name: "indirect_taint_flow",
		// 	code: `
		// 			const input = req.params.id;
		// 			const temp = input;
		// 			const query = "SELECT * FROM users WHERE id = " + temp;
		// 			db.query(query);
		// 		`,
		// 	wantVuln: true,
		// },
		// {
		// 	name: "safe_parameterized_query",
		// 	code: `
		// 			const input = req.params.id;
		// 			const sanitized = parameterize(input);
		// 			db.query("SELECT * FROM users WHERE id = ?", sanitized);
		// 		`,
		// 	wantVuln: false,
		// },
		// // DFG Implementation detects only one of them as vulnerable, whereas the basic implementation flags both as vulnerable.
		// {
		// 	name: "multiple_sanitization_paths",
		// 	code: `
		// 			const input = req.body.data;
		// 			const path1 = sanitize(input);
		// 			const path2 = input;
		// 			db.query("SELECT * FROM users WHERE data = " + path1);
		// 			db.query("SELECT * FROM users WHERE data = " + path2);
		// 		`,
		// 	wantVuln: true,
		// },
		// {
		// 	name: "binary_expression_chaining",
		// 	code: `
		// 			const a = req.query.a;
		// 			const b = "static";
		// 			const c = a + b + "more";
		// 			db.query("SELECT * FROM table WHERE col = " + c);
		// 		`,
		// 	wantVuln: true,
		// },
		// {
		// 	name: "nested_template_strings",
		// 	code: `
		// 			const user = req.body.user;
		// 			const role = req.body.role;
		// 			db.query(` + "`" + "SELECT * FROM users WHERE name = ${user} AND role = ${escape(role)}" + "`" + `);
		// 		`,
		// 	wantVuln: true, // Should detect unsanitized user variable
		// },
		// {
		// 	name: "global_scope_taint",
		// 	code: `
		//     let userInput = req.query.id; // Tainted global variable
		//     function getUser() {
		//         db.query("SELECT * FROM users WHERE id = " + userInput); // Vulnerable use of global
		//     }
		//     getUser();
		// `,
		// 	wantVuln: true,
		// },

		{
			// Doesn't Pass for dfg
			name: "function_parameter_taint",
			code: `
        function findUser(userId) {
            // userId parameter is tainted when called
            db.query("SELECT * FROM users WHERE id = " + y);
        }
        const taintedId = req.params.id;
        findUser(sanitize(taintedId)); // Call with tainted data
    `,
			wantVuln: false,
		},
		// {
		// 	name: "variable_shadowing_safe",
		// 	code: `
		//     let id = req.query.id; // Tainted global 'id'
		//     function getUserData() {
		//         let id = 123; // Safe local 'id' shadows the global one
		//         db.query("SELECT * FROM users WHERE id = " + id); // Uses safe local 'id'
		//     }
		//     getUserData();
		// `,
		// 	wantVuln: false,
		// },
		// {
		// 	name: "variable_shadowing_unsafe",
		// 	code: `
		//     let id = 123; // Safe global 'id'
		//     function getUserData(req) {
		//         let id = req.params.userId; // Unsafe local 'id' shadows the global one
		//         db.query("SELECT * FROM users WHERE id = " + id); // Uses unsafe local 'id'
		//     }
		//     getUserData(req);
		// `,
		// 	wantVuln: true,
		// },
		// {
		// 	name: "closure_taint",
		// 	code: `
		//     function setupHandler(req) {
		//         const userId = req.body.id; // Tainted variable in outer scope
		//         return function executeQuery() {
		//             // Inner function (closure) uses the tainted userId
		//             db.query("SELECT * FROM data WHERE user_id = " + userId);
		//         };
		//     }
		//     const handler = setupHandler(req);
		//     handler();
		// `,
		// 	wantVuln: true,
		// },
		// {
		// 	name: "reassignment_safe_to_unsafe",
		// 	code: `
		//     let queryParam = "default"; // Initially safe
		//     if (req.query.user) {
		//         queryParam = req.query.user; // Reassigned with potentially tainted data
		//     }
		//     db.query("SELECT * FROM logs WHERE user = " + queryParam); // Potentially vulnerable
		// `,
		// 	wantVuln: true,
		// },
		// {
		// 	name: "reassignment_unsafe_to_safe",
		// 	code: `
		//     let userId = req.params.id; // Initially tainted
		//     if (!isValidIdFormat(userId)) {
		//          userId = "invalid"; // Reassigned to a safe, static value
		//     } else {
		//          userId = escape(userId); // Reassigned to a sanitized value
		//     }
		//     db.query("SELECT * FROM users WHERE id = " + userId); // Now safe
		// `,
		// 	wantVuln: false,
		// },
		{
			// Doesn't pass for DFG
			name: "block_scope_taint_let",
			code: `
        let userId = "guest";
        if (req.user) {
            let userId = req.user.id; // Tainted variable limited to block scope
            db.query("SELECT data FROM user_data WHERE id = " + userId); // Vulnerable within the block
        }
        // Outside the block, userId refers to "guest"
        db.query("SELECT * FROM logs WHERE user = " + userId); // This would be safe if uncommented
    `,
			wantVuln: true, // Vulnerability exists within the 'if' block
		},
		// {
		// 	name: "block_scope_no_taint_const",
		// 	code: `
		//     const SAFE_ID = 1;
		//     if (req.params.id) {
		//         const inputId = req.params.id; // Tainted, but only used locally if needed
		//         // Assume some logic here, but inputId doesn't reach query
		//     }
		//     // SAFE_ID is still 1
		//     db.query("SELECT * FROM config WHERE id = " + SAFE_ID);
		// `,
		// 	wantVuln: false,
		// },
	}
	for _, tt := range tests {
		tt := tt // Capture range variable

		// --- Run Basic Test ---

		// t.Run(tt.name+" Basic", func(t *testing.T) {
		// 	parseResult := parseJsCode(t, []byte(tt.code))
		// 	var reportedVulns []struct {
		// 		Pass *ana.Pass
		// 		Node *sitter.Node
		// 		Msg  string
		// 	}
		// 	pass := &ana.Pass{
		// 		Analyzer:    SQLInjection,
		// 		FileContext: parseResult,
		// 		Report: func(p *ana.Pass, n *sitter.Node, msg string) {
		// 			reportedVulns = append(reportedVulns, struct {
		// 				Pass *ana.Pass
		// 				Node *sitter.Node
		// 				Msg  string
		// 			}{p, n, msg})
		// 		},
		// 	}

		// 	_, err := detectSQLInjection(pass)
		// 	require.NoError(t, err)

		// 	t.Logf("Basic: detected vulnerabilities: %d\n", len(reportedVulns))

		// })

		// --- Run DFG Test (always runs if dfgOnly is true, or runs after basic if dfgOnly is false) ---
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

			t.Logf("DFG: detected vulnerabilities: %d\n", len(reportedVulns))
			// Assert based on the DFG expectation
			if tt.wantVuln {
				assert.Greater(
					t,
					len(reportedVulns),
					0,
					"DFG: Expected to detect sql injection vulnerability",
				)
			} else {
				assert.Equal(
					t,
					0,
					len(reportedVulns),
					"DFG: Unexpected Sql injection vuln. reported",
				)
			}
		})
	}

}
