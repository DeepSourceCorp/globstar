package discover

import "testing"

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
