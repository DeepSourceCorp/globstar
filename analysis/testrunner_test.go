package analysis

import (
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
)

func TestVerifyIssues(t *testing.T) {
	tests := []struct {
		name     string
		expected map[string]map[int][]string
		raised   map[string]map[int][]string
		wantDiff string
	}{
		{
			name: "matching issues",
			expected: map[string]map[int][]string{
				"test.js": {
					1: {"TEST001: Test error"},
				},
			},
			raised: map[string]map[int][]string{
				"test.js": {
					1: {"TEST001: Test error"},
				},
			},
			wantDiff: "",
		},
		{
			name: "missing issue",
			expected: map[string]map[int][]string{
				"test.js": {
					1: {"TEST001: Test error"},
				},
			},
			raised: map[string]map[int][]string{
				"test.js": {},
			},
			wantDiff: "\nFile: test.js, Line: 1\n  Expected:\n    - TEST001: Test error\n  Got: no issues\n",
		},
		{
			name: "unexpected issue",
			expected: map[string]map[int][]string{
				"test.js": {},
			},
			raised: map[string]map[int][]string{
				"test.js": {
					1: {"TEST001: Test error"},
				},
			},
			wantDiff: "\nFile: test.js, Line: 1\n  Expected: no issues\n  Got:\n    - TEST001: Test error\n",
		},
		{
			name: "unexpected file",
			expected: map[string]map[int][]string{
				"test.js": {},
			},
			raised: map[string]map[int][]string{
				"test.js": {},
				"new.js": {
					1: {"TEST001: Test error"},
				},
			},
			wantDiff: "\nUnexpected file with issues: new.js\n  Line 1:\n    - TEST001: Test error\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := verifyIssues(&tt.expected, &tt.raised)
			if got != tt.wantDiff {
				t.Errorf("verifyIssues() diff = %v, want %v", got, tt.wantDiff)
			}
		})
	}
}

func TestMessagesEqual(t *testing.T) {
	tests := []struct {
		name     string
		expected []string
		actual   []string
		want     bool
	}{
		{
			name:     "identical messages",
			expected: []string{"TEST001: Error 1", "TEST002: Error 2"},
			actual:   []string{"TEST001: Error 1", "TEST002: Error 2"},
			want:     true,
		},
		{
			name:     "different order",
			expected: []string{"TEST002: Error 2", "TEST001: Error 1"},
			actual:   []string{"TEST001: Error 1", "TEST002: Error 2"},
			want:     true,
		},
		{
			name:     "different messages",
			expected: []string{"TEST001: Error 1"},
			actual:   []string{"TEST001: Different error"},
			want:     false,
		},
		{
			name:     "different lengths",
			expected: []string{"TEST001: Error 1"},
			actual:   []string{"TEST001: Error 1", "TEST002: Error 2"},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := messagesEqual(tt.expected, tt.actual); got != tt.want {
				t.Errorf("messagesEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetExpectedIssuesInFile(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		want     map[int][]string
		language Language
	}{
		{
			name: "javascript.js",
			content: `function test() {
    // <expect-error> TEST001: Error message
    var x = 1 == 1;
}`,
			want: map[int][]string{
				3: {"TEST001: Error message"},
			},
			language: LangJs,
		},
		{
			name: "multiple-errors.js",
			content: `function test() {
    // <expect-error> TEST001: First error
    var x = 1 == 1;
    // <expect-error> TEST002: Second error
    var y = 2 == 2;
}`,
			want: map[int][]string{
				3: {"TEST001: First error"},
				5: {"TEST002: Second error"},
			},
			language: LangJs,
		},
		{
			name: "no-expect-error.js",
			content: `function test() {
	var x = 1 == 1;
}`,
			want:     map[int][]string{},
			language: LangJs,
		},
		{
			name: "expect-error-without-message.js",
			content: `function test() {
	// <expect-error>
	var x = 1 == 1;
}`,
			want: map[int][]string{
				3: {""},
			},
			language: LangJs,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, err := Parse(tt.name, []byte(tt.content), tt.language, tt.language.Grammar())
			if err != nil {
				t.Fatalf("Failed to parse file: %v", err)
			}

			query, err := sitter.NewQuery([]byte("(comment) @pragma"), file.Language.Grammar())
			if err != nil {
				t.Fatalf("Failed to create query: %v", err)
			}

			got := getExpectedIssuesInFile(file, query)
			if !mapsEqual(got, tt.want) {
				t.Errorf("getExpectedIssuesInFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper function to compare maps
func mapsEqual(got, want map[int][]string) bool {
	if len(got) != len(want) {
		return false
	}
	for k, v := range got {
		if !messagesEqual(v, want[k]) {
			return false
		}
	}
	return true
}
