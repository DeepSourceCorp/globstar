package cli

import (
	"fmt"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// GetChangedFiles returns all changes between latest commit and the previous one
func (c *Cli) GetChangedFiles() ([]string, error) {

	// Open the git repository at the root directory
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
		return nil, fmt.Errorf("Couldnt get revision hash: %v", err)
	}

	// Get the commit object for the latest commit
	latestCommit, err := repo.CommitObject(*latestCommitHash)
	if err != nil {
		return nil, fmt.Errorf("Could not get revision commit: %v\n", err)
	}

	// Resolve the hash for the previous commit
	prevCommitHash, err := repo.ResolveRevision(plumbing.Revision(prev))
	if err != nil {
		return nil, fmt.Errorf("Couldnt get revision hash: %v", err)
	}

	// Get the commit object for the previous commit
	prevCommit, err := repo.CommitObject(*prevCommitHash)
	if err != nil {
		return nil, fmt.Errorf("Could not get revision commit: %v\n", err)
	}

	// GEnerate a patch between the previous and latest commits
	patch, err := prevCommit.Patch(latestCommit)
	if err != nil {
		return nil, fmt.Errorf("Could not create a patch")
	}

	// Extract the changed files from the patch
	filePatches := patch.FilePatches()
	var changedFiles []string;
	for _, filePatch := range filePatches {
		_, file := filePatch.Files()
		// Construct absolute paths, to be used by the analyzer.
		filePath, err := filepath.Abs(filepath.Join(c.RootDirectory, file.Path()))
		if err != nil {
			return changedFiles, fmt.Errorf("could not resolve filepath")
		}

		changedFiles = append(changedFiles, filePath)
	}

	

	return changedFiles, nil

}