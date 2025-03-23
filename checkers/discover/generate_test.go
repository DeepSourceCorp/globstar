package discover

import (
	"testing"
)

func TestGenerateAnalyzerRegistry(t *testing.T) {
	tests := []struct {
		name       string
		goCheckers []string
		want       string
	}{
		{
			name:       "empty checkers",
			goCheckers: []string{},
			want: `package main

import (
	"globstar.dev/analysis"
)

var customCheckers []*analysis.Analyzer = []*analysis.Analyzer{}`,
		},
		{
			name:       "single checker",
			goCheckers: []string{"checkers.NoDoubleEq"},
			want: `package main

import (
	"globstar.dev/analysis"
)

var customCheckers []*analysis.Analyzer = []*analysis.Analyzer{
	checkers.NoDoubleEq,
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generateAnalyzerRegistry(tt.goCheckers); got != tt.want {
				t.Errorf("generateAnalyzerRegistry() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateBuiltinChecker(t *testing.T) {
	tests := []struct {
		name       string
		goCheckers map[string][]string
		want       string
	}{
		{
			name:       "empty checkers",
			goCheckers: map[string][]string{},
			want: `// AUTOMATICALLY GENERATED: DO NOT EDIT

package checkers

import (
	"globstar.dev/checkers/javascript"
	"globstar.dev/checkers/python"
	goAnalysis "globstar.dev/analysis"
)

type Analyzer struct {
	TestDir   string
	Analyzers []*goAnalysis.Analyzer
}

var AnalyzerRegistry = []Analyzer{
	{
		TestDir:   "checkers/javascript/testdata", // relative to the repository root
		Analyzers: []*goAnalysis.Analyzer{
		},
	},
	{
		TestDir: "checkers/python/testdata",
		Analyzers: []*goAnalysis.Analyzer{
		},
	},
}
`,
		},
		{
			name: "only javascript",
			goCheckers: map[string][]string{
				"javascript": {
					"javascript.NoDoubleEq",
					"javascript.SQLInjection",
				},
			},
			want: `// AUTOMATICALLY GENERATED: DO NOT EDIT

package checkers

import (
	"globstar.dev/checkers/javascript"
	"globstar.dev/checkers/python"
	goAnalysis "globstar.dev/analysis"
)

type Analyzer struct {
	TestDir   string
	Analyzers []*goAnalysis.Analyzer
}

var AnalyzerRegistry = []Analyzer{
	{
		TestDir:   "checkers/javascript/testdata", // relative to the repository root
		Analyzers: []*goAnalysis.Analyzer{
			javascript.NoDoubleEq,
			javascript.SQLInjection,

		},
	},
	{
		TestDir: "checkers/python/testdata",
		Analyzers: []*goAnalysis.Analyzer{
		},
	},
}
`,
		},
		{
			name: "only python",
			goCheckers: map[string][]string{
				"python": {
					"python.DjangoSQLInjection",
					"python.DjangoCSVWriterInjection",
				},
			},
			want: `// AUTOMATICALLY GENERATED: DO NOT EDIT

package checkers

import (
	"globstar.dev/checkers/javascript"
	"globstar.dev/checkers/python"
	goAnalysis "globstar.dev/analysis"
)

type Analyzer struct {
	TestDir   string
	Analyzers []*goAnalysis.Analyzer
}

var AnalyzerRegistry = []Analyzer{
	{
		TestDir:   "checkers/javascript/testdata", // relative to the repository root
		Analyzers: []*goAnalysis.Analyzer{
		},
	},
	{
		TestDir: "checkers/python/testdata",
		Analyzers: []*goAnalysis.Analyzer{
			python.DjangoSQLInjection,
			python.DjangoCSVWriterInjection,

		},
	},
}
`,
		},
		{
			name: "both language",
			goCheckers: map[string][]string{
				"javascript": {
					"javascript.SQLInjection",
					"javascript.NoDoubleEq",
				},
				"python": {
					"python.DjangoSQLInjection",
					"python.DjangoCSVWriterInjection",
					"python.InsecureUrllibFtp",
				},
			},
			want: `// AUTOMATICALLY GENERATED: DO NOT EDIT

package checkers

import (
	"globstar.dev/checkers/javascript"
	"globstar.dev/checkers/python"
	goAnalysis "globstar.dev/analysis"
)

type Analyzer struct {
	TestDir   string
	Analyzers []*goAnalysis.Analyzer
}

var AnalyzerRegistry = []Analyzer{
	{
		TestDir:   "checkers/javascript/testdata", // relative to the repository root
		Analyzers: []*goAnalysis.Analyzer{
			javascript.SQLInjection,
			javascript.NoDoubleEq,

		},
	},
	{
		TestDir: "checkers/python/testdata",
		Analyzers: []*goAnalysis.Analyzer{
			python.DjangoSQLInjection,
			python.DjangoCSVWriterInjection,
			python.InsecureUrllibFtp,

		},
	},
}
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generateBuiltinCheckerRegistry(tt.goCheckers); got != tt.want {
				t.Errorf("generateBuiltinAnalyzerRegistry() = \n%v, want \n%v", got, tt.want)
			}
		})
	}

}
