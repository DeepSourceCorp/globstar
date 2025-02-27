package discover

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverGoRules(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		dir  string
		want struct {
			goRules []string
			err     error
		}
	}{
		{
			dir: filepath.Join(cwd, "fixtures", "checkers"),
			want: struct {
				goRules []string
				err     error
			}{
				goRules: []string{"checkers.NoDoubleEq"},
				err:     nil,
			},
		},
		{
			dir: "fixtures/empty",
			want: struct {
				goRules []string
				err     error
			}{
				goRules: []string{},
				err:     nil,
			},
		},
		{
			dir: "fixtures/invalid",
			want: struct {
				goRules []string
				err     error
			}{
				goRules: []string{},
				err:     fmt.Errorf("open fixtures/invalid: no such file or directory"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.dir, func(t *testing.T) {
			got, err := DiscoverGoRules(tt.dir)
			if err != tt.want.err && err.Error() != tt.want.err.Error() {
				t.Errorf("DiscoverGoRules() error = %v, want %v", err, tt.want.err)
			}
			if len(got) != len(tt.want.goRules) {
				t.Errorf("DiscoverGoRules() = %v, want %v", got, tt.want.goRules)
			}
		})
	}
}
