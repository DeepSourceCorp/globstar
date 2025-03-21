package cli

import (
	"fmt"
	"path/filepath"

	"github.com/go-git/go-git/v5"
)

// GetChangedFiles returns the absolute paths of files that have changed since the last commit.
// This includes both staged and unstaged changes in the working directory.
func (c *Cli) GetChangedFiles() ([]string, error) {
	// Open the repository
	repo, err := git.PlainOpen(c.RootDirectory)
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository: %v", err)
	}
	
	// Get worktree
	worktree, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree: %v", err)
	}
	
	// Get status of working directory
	status, err := worktree.Status()
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %v", err)
	}
	
	// Create a slice to store the absolute paths
	changedFiles := make([]string, 0, len(status))
	
	// Iterate through all changes
	for path, fileStatus := range status {
		// Skip deleted files
		if fileStatus.Worktree == git.Deleted && fileStatus.Staging == git.Unmodified {
			continue
		}
		
		// Include files with changes in either worktree or staging area
		if fileStatus.Worktree != git.Unmodified || fileStatus.Staging != git.Unmodified {
			absPath, err := filepath.Abs(filepath.Join(c.RootDirectory, path))
			if err != nil {
				return nil, fmt.Errorf("failed to get absolute path for %s: %v", path, err)
			}
			changedFiles = append(changedFiles, absPath)
		}
	}
	
	return changedFiles, nil
}