package cli

import (
	"fmt"
	"path/filepath"
	"slices"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func getCurrentWorkDirState(repo *git.Repository, rootDir string) ([]string, error) {
	worktree, err := repo.Worktree()

	if err != nil {
		return nil, fmt.Errorf("Could not get the current directory worktree: %v\n", worktree)
	}

	var changedFiles []string

	status, err := worktree.Status()
	if err != nil {
		return nil, fmt.Errorf("Could not get worktree status: %v", err)
	}

	for file, fileStatus := range status {
		if fileStatus.Worktree != git.Unmodified || fileStatus.Staging != git.Unmodified {
			fp, err := filepath.Abs(filepath.Join(rootDir, file))
			if err != nil {
				return changedFiles, fmt.Errorf("could not resolve filepath for %s: %v", file, err)
			}
			changedFiles = append(changedFiles, fp)
		}

	}

	return changedFiles, nil
}

// GetChangedFiles returns all changes between latest commit and the previous one
func (c *Cli) GetChangedFiles(compareCommitHash string) ([]string, error) {
	// Open the git repository at the root directory
	repo, err := git.PlainOpen(c.RootDirectory)
	if err != nil {
		return nil, fmt.Errorf("Could not open the Directory.")
	}

	headRef, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("Unable to fetch HEAD of the repo, try commiting first :P. Err: %v\n", err)
	}

	headCommit, err := repo.CommitObject(headRef.Hash())
	if err != nil {
		return nil, fmt.Errorf("Could not get HEAD commit: %v\n", err)
	}

	var changedFiles []string
	switch len(compareCommitHash) {
	case 0:
		currDirState, err := getCurrentWorkDirState(repo, c.RootDirectory)
		if err != nil {
			return nil, fmt.Errorf("%v\n", err)
		}

		changedFiles = currDirState

	default:
		prev := compareCommitHash
		prevCommitHash, err := repo.ResolveRevision(plumbing.Revision(prev))
		if err != nil {
			return nil, fmt.Errorf("Could not get revision hash: %v\n", err)
		}

		prevCommit, err := repo.CommitObject(*prevCommitHash)

		if err != nil {
			return nil, fmt.Errorf("Could not get the commit object for provided commit-hash: %v\n", prevCommit)
		}

		allChanges, err := getCurrentWorkDirState(repo, c.RootDirectory)

		if err != nil {
			return nil, fmt.Errorf("Problem getting current directory state: %v\n", err)
		}

		patch, err := prevCommit.Patch(headCommit)
		if err != nil {
			return nil, fmt.Errorf("Could not create a patch")
		}
		changedFiles = allChanges
		// Extract the changed files from the patch
		filePatches := patch.FilePatches()
		for _, filePatch := range filePatches {
			_, file := filePatch.Files()
			// Construct absolute paths, to be used by the analyzer.
			filePath, err := filepath.Abs(filepath.Join(c.RootDirectory, file.Path()))
			if err != nil {
				return changedFiles, fmt.Errorf("could not resolve filepath")
			}
			if slices.Contains(allChanges, filePath) {
				continue
			}
			changedFiles = append(changedFiles, filePath)
		}
	}

	return changedFiles, nil

}
