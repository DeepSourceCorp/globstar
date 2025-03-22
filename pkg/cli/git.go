package cli

import (
	"fmt"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/utils/merkletrie"
)

// GetChangedFiles returns all changes between compareHash and the current
// state of the working directory
func (c *Cli) GetChangedFiles(compareHash string) ([]string, error) {
	changedFiles := []string{}
	repo, err := git.PlainOpen(c.RootDirectory)
	if err !=  nil {
		return nil, fmt.Errorf("Could not open the Directory.")
	}

	// Define references for the current and previous commits
	head := "HEAD"
	prev:= "HEAD~1"

	// Resolve the hash for the latest commit
	latestCommitHash, err := repo.ResolveRevision(plumbing.Revision(head))
	if err != nil {
		return changedFiles, fmt.Errorf("failed to open git repository: %w", err)
	}

	// Get worktree
	worktree, err := repo.Worktree()
	if err != nil {
		return changedFiles, fmt.Errorf("failed to get worktree: %w", err)
	}

	// Get status between commit and working directory
	status, err := worktree.Status()
	if err != nil {
		return changedFiles, fmt.Errorf("failed to get status: %w", err)
	}

	// Get compare commit
	if compareHash == "" {
		return changedFiles, fmt.Errorf("compare hash is required")
	}

	hash := plumbing.NewHash(compareHash)
	compareCommit, err := repo.CommitObject(hash)
	if err != nil {
		return changedFiles, fmt.Errorf("failed to get compare commit: %w", err)
	}

	// Get commit tree for comparison
	compareTree, err := compareCommit.Tree()
	if err != nil {
		return changedFiles, fmt.Errorf("failed to get compare tree: %w", err)
	}

	// Get the status between the HEAD commit and the compare hash
	// Get the HEAD commit tree for comparison
	headRef, err := repo.Head()
	if err != nil {
		return changedFiles, fmt.Errorf("failed to get HEAD ref: %w", err)
	}

	headCommit, err := repo.CommitObject(headRef.Hash())
	if err != nil {
		return changedFiles, fmt.Errorf("failed to get HEAD commit: %w", err)
	}

	headTree, err := headCommit.Tree()
	if err != nil {
		return changedFiles, fmt.Errorf("failed to get HEAD tree: %w", err)
	}

	// Get the changes between the two trees
	changes, err := compareTree.Diff(headTree)
	if err != nil {
		return changedFiles, fmt.Errorf("failed to get changes between HEAD and compare commit: %w", err)
	}

	changedFilesMap := map[string]struct{}{}
	// Get the files that have changed
	for _, change := range changes {
		action, err := change.Action()
		if err != nil {
			continue
		}

		if action == merkletrie.Delete {
			continue
		}

		changedFilesMap[change.To.Name] = struct{}{}
	}

	for file, fileStatus := range status {
		if fileStatus.Worktree == git.Deleted || fileStatus.Staging == git.Deleted {
			delete(changedFilesMap, file)
			continue
		}

		if fileStatus.Worktree == git.Unmodified && fileStatus.Staging == git.Unmodified {
			continue
		}

		changedFilesMap[file] = struct{}{}
	}

	for file := range changedFilesMap {
		changedFiles = append(changedFiles, filepath.Join(c.RootDirectory, file))
	}

	

	return changedFiles, nil
}
