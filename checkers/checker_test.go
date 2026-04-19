package checkers

import (
	"os"
	"path/filepath"
	"testing"

	"globstar.dev/analysis"
)

// TestLoadCustomYamlCheckers_ReadsFromDisk is a regression test ensuring that
// custom YAML checkers are loaded from the local filesystem and not silently
// dropped because the loader was reading from the embedded FS.
func TestLoadCustomYamlCheckers_ReadsFromDisk(t *testing.T) {
	dir := t.TempDir()
	const yml = `language: go
name: go_custom_test
message: "custom checker fired"
category: security
severity: critical
pattern: >
  [
    (import_spec
      path: (interpreted_string_literal) @import
      (#eq? @import "\"crypto/md5\""))
  ] @go_custom_test
`
	if err := os.WriteFile(filepath.Join(dir, "custom.yml"), []byte(yml), 0o644); err != nil {
		t.Fatalf("write yaml: %v", err)
	}

	checkersMap, err := LoadCustomYamlCheckers(dir)
	if err != nil {
		t.Fatalf("LoadCustomYamlCheckers: %v", err)
	}

	goCheckers, ok := checkersMap[analysis.LangGo]
	if !ok || len(goCheckers) == 0 {
		t.Fatalf("expected at least one Go checker loaded from disk, got map=%v", checkersMap)
	}

	found := false
	for _, c := range goCheckers {
		if c.Name == "go_custom_test" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected 'go_custom_test' to be loaded; got %+v", goCheckers)
	}
}
