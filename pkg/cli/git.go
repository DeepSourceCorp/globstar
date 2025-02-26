package cli

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// GetChangedFiles returns the list of files that have been changed in the codebase
func (c *Cli) GetChangedFiles() ([]string, error) {

	cmd := exec.Command("git", "status", "--porcelain=v1", "-z")

	// Set working directory
	cmd.Dir = c.RootDirectory

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to run git status in directory %s: %v", cmd.Dir, err)
	}

	var files []string

	// Parse the output
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
