package cli

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// GetChangedFiles executes a git command to collect files that have changed in the repository
// and returns their absolute paths. It skips files that have been deleted or have unknown status.
// Returns a slice of absolute file paths and any error encountered during execution.

func (c *Cli) GetChangedFiles() ([]string, error) {

	cmd := exec.Command("git", "status", "--porcelain=v1", "-z")

	cmd.Dir = c.RootDirectory
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to run git status in directory %s:%v", cmd.Dir, err)
	}

	var files []string

	// Parse the output from git status command
	// Split by NUL character which separates entries in -z format
	for _, line := range strings.Split(strings.TrimRight(out.String(), "\x00"), "\x00") {
		if len(line) < 3 {
			continue
		}

		status := parseGitStatus(line[:2])
		path := strings.TrimSpace(line[3:])

		if status == "renamed" {
			parts := strings.Split(path, " -> ")
			if len(parts) == 2 {
				path = parts[1]
			}
		}

		if status == "deleted" || status == "unknown" {
			continue
		}

		if path != "" {
			// Convert path to absolute path
			absPath, err := filepath.Abs(filepath.Join(cmd.Dir, path))
			if err != nil {
				return nil, fmt.Errorf("failed to get absolute path for %s: %v", path, err)
			}

			files = append(files, absPath)
		}
	}

	return files, nil

}

// parseGitStatus converts a git status code to a human-readable string.
// It takes the status code from `git status --porcelain` output and
// returns a string describing the status.
func parseGitStatus(code string) string {
	code = strings.TrimSpace(code)

	switch code {
	case "M", "MM":
		return "modified"
	case "A":
		return "added"
	case "D":
		return "deleted"
	case "R":
		return "renamed"
	case "??":
		return "untracked"
	default:
		return "unknown"
	}
}
