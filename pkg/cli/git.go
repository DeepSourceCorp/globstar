package cli

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// GetChangedFiles returns all changes (including working directory changes)
func (c *Cli) GetChangedFiles() ([]string, error) {

	repo, err := git.PlainOpen(c.RootDirectory)
	if err !=  nil {
		return nil, fmt.Errorf("Could not open the Directory.")
	}

	head, err := repo.Head()

	if err != nil {
		return nil, fmt.Errorf("failed to get the HEAD: %v", err)
	}


	currentCommitHash := head.Hash()

	currCommit, err := repo.CommitObject(currentCommitHash)

	if err != nil {
		return nil, fmt.Errorf("Could not get commit from the hash")
	}

	commitIter, err := repo.Log(&git.LogOptions{From: currentCommitHash})
	if err != nil {
		return nil, fmt.Errorf("Unable to fetch details of the commits.")
	}

	var hashHistory []plumbing.Hash

	err = commitIter.ForEach(func(c *object.Commit) error {
		hashHistory	= append(hashHistory, c.Hash)
		return nil
	})

	prevCommit, err:= repo.CommitObject(hashHistory[1])

	if err != nil {
		return nil, fmt.Errorf("could not get previous commitObject from hash")
	}

	patch, err := currCommit.Patch(prevCommit)
	if err != nil {
		return nil, fmt.Errorf("Could not create patch")
	}

	filePatches := patch.FilePatches()

	for _, filePatch := range filePatches {
		_, file := filePatch.Files()
		fmt.Printf("%v\n", file.Path())
	}

	return nil, nil

}