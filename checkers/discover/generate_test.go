package discover

import "testing"

func TestGenerateAnalyzerRegistry(t *testing.T) {
	tests := []struct {
		name    string
		goRules []string
		want    string
	}{
		{
			name:    "empty rules",
			goRules: []string{},
			want: `package main

import (
	"globstar.dev/analysis"
)

var customRules []*analysis.Analyzer = []*analysis.Analyzer{}`,
		},
		{
			name:    "single rule",
			goRules: []string{"checkers.NoDoubleEq"},
			want: `package main

import (
	"globstar.dev/analysis"
)

var customRules []*analysis.Analyzer = []*analysis.Analyzer{
	checkers.NoDoubleEq,
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generateAnalyzerRegistry(tt.goRules); got != tt.want {
				t.Errorf("generateAnalyzerRegistry() = %v, want %v", got, tt.want)
			}
		})
	}
}
