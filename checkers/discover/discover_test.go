package discover

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverGoCheckers(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		dir  string
		want struct {
			goCheckers []string
			err        error
		}
	}{
		{
			dir: filepath.Join(cwd, "fixtures", "checkers"),
			want: struct {
				goCheckers []string
				err        error
			}{
				goCheckers: []string{"checkers.NoDoubleEq"},
				err:        nil,
			},
		},
		{
			dir: "fixtures/empty",
			want: struct {
				goCheckers []string
				err        error
			}{
				goCheckers: []string{},
				err:        nil,
			},
		},
		{
			dir: "fixtures/invalid",
			want: struct {
				goCheckers []string
				err        error
			}{
				goCheckers: []string{},
				err:        fmt.Errorf("open fixtures/invalid: no such file or directory"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.dir, func(t *testing.T) {
			got, err := DiscoverGoCheckers(tt.dir)
			if err != tt.want.err && err.Error() != tt.want.err.Error() {
				t.Errorf("DiscoverGoCheckers() error = %v, want %v", err, tt.want.err)
			}
			if len(got) != len(tt.want.goCheckers) {
				t.Errorf("DiscoverGoCheckers() = %v, want %v", got, tt.want.goCheckers)
			}
		})
	}
}
