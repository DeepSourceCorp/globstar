package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"globstar.dev/pkg/config"
)

// TestRunCheckers_ExecutesCustomYamlCheckers is a regression test ensuring
// that YAML pattern checkers loaded from a user's .globstar directory are
// actually executed against project files. Previously, RunCheckers loaded
// the checkers into a map but never passed them to the analyzer runner, so
// custom YAML rules silently produced zero findings.
func TestRunCheckers_ExecutesCustomYamlCheckers(t *testing.T) {
	tmpDir := t.TempDir()

	checkerDir := filepath.Join(tmpDir, ".globstar")
	if err := os.MkdirAll(checkerDir, 0o755); err != nil {
		t.Fatalf("mkdir .globstar: %v", err)
	}

	const yamlChecker = `language: go
name: go_filepath_clean_test
message: "Found filepath.Clean"
category: security
severity: critical
pattern: >
  (call_expression
    function: (selector_expression
      operand: (identifier) @pkg
      field: (field_identifier) @func
    )
    (#eq? @pkg "filepath")
    (#eq? @func "Clean")
  ) @go_filepath_clean_test
`
	if err := os.WriteFile(filepath.Join(checkerDir, "my_check.yml"), []byte(yamlChecker), 0o644); err != nil {
		t.Fatalf("write yaml: %v", err)
	}

	const goSource = `package main

import "path/filepath"

func main() {
	_ = filepath.Clean("x")
}
`
	if err := os.WriteFile(filepath.Join(tmpDir, "test.go"), []byte(goSource), 0o644); err != nil {
		t.Fatalf("write go file: %v", err)
	}

	conf := &config.Config{}
	conf.PopulateDefaults()
	conf.CheckerDir = checkerDir

	c := &Cli{
		RootDirectory: tmpDir,
		Config:        conf,
	}

	err := c.RunCheckers(false, true)
	// A critical issue should be raised, which by default causes RunCheckers
	// to return a non-nil error ("found N issues") via FailWhen. If the YAML
	// execution bug returns, no issues are produced and err would be nil.
	require.Error(t, err, "expected RunCheckers to report a finding from the custom YAML checker")
	require.Contains(t, err.Error(), "found 1 issues")
}
