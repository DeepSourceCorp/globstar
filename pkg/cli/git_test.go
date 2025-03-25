package cli

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This function uses the copyFile function to copy an entire `src` directory into the `dst` path
func copyDirectory(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if d.IsDir() {
			return os.MkdirAll(dstPath, os.ModePerm)
		}

		return copyFile(path, dstPath)

	})
}

// This functions copies a given file from the provided `src` directory to `dst` directory
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}

	defer srcFile.Close()

	dstFile, err := os.Create(dst)

	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)

	return nil
}

func TestGetChangedFiles(t *testing.T) {

	testData := "testdata"

	t.Run("Checking the recognition of current changed state of the work directory, without commiting.", func(t *testing.T) {
		// Creating a temporary directory to run the tests on.
		tempDir, err := os.MkdirTemp("", "git-test-*")
		require.NoError(t, err, "Could Not create a temporary directory")
		defer os.RemoveAll(tempDir)

		cli := &Cli{RootDirectory: tempDir}
		// Copying the test-data directory containing assets for the tests to be run on.
		err = copyDirectory(testData, tempDir)
		// Initializing the temporary directory as a git repo.
		repo, err := git.PlainInit(tempDir, false)
		require.NoError(t, err)

		worktree, err := repo.Worktree()
		require.NoError(t, err)

		_, err = worktree.Add(".")
		require.NoError(t, err)

		_, err = worktree.Commit("Initial commit", &git.CommitOptions{
			Author: &object.Signature{
				Name:  "Test User",
				Email: "test@example.com",
				When:  time.Now(),
			},
		})

		fp := filepath.Join(tempDir, "test2.txt")

		file2, err := os.OpenFile(fp, os.O_APPEND|os.O_WRONLY, 0644)
		require.NoError(t, err, "could not open file for editing")

		_, err = file2.WriteString("Use Globstar. It's the best out there.")
		require.NoError(t, err, "could not open file for editing")

		changedFiles, err := cli.GetChangedFiles("")
		require.NoError(t, err, "Couldn't fetch the changed files")
		assert.Contains(t, changedFiles, fp)

	})

	t.Run("Checking the recognition of an added file between two commits", func(t *testing.T) {
		// Creating a temporary directory to run the tests on.
		tempDir, err := os.MkdirTemp("", "git-test-*")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		cli := &Cli{RootDirectory: tempDir}
		// Copying the test-data directory containing assets for the tests to be run on.
		err = copyDirectory(testData, tempDir)
		require.NoError(t, err, "Failed to copy test data directory")

		// Initializing the temporary directory as a git repo.
		repo, err := git.PlainInit(tempDir, false)
		require.NoError(t, err)

		worktree, err := repo.Worktree()
		require.NoError(t, err)

		// Adding all files to git staging area.
		_, err = worktree.Add(".")
		require.NoError(t, err)

		// Creating initial commit with all files.
		commit, err := worktree.Commit("Initial commit", &git.CommitOptions{
			Author: &object.Signature{
				Name:  "Test User",
				Email: "test@example.com",
				When:  time.Now(),
			},
		})

		// Creating a new file to test file addition detection.
		filePath := filepath.Join(tempDir, "test3.txt")
		if err := os.WriteFile(filePath, []byte(""), 0644); err != nil {
			require.NoError(t, err)
		}

		// Adding the new file to git staging area.
		_, err = worktree.Add("test3.txt")

		// Creating a second commit with the new file.
		_, err = worktree.Commit("Add a new file", &git.CommitOptions{
			Author: &object.Signature{
				Name:  "Test User",
				Email: "test@example.com",
				When:  time.Now(),
			},
		})
		require.NoError(t, err)

		// Getting the hash of the initial commit for comparison.
		obj, err := repo.CommitObject(commit)
		initialHash := obj.Hash.String()
		require.NoError(t, err)

		// Checking if the new file is detected in the changes.
		changedFiles, err := cli.GetChangedFiles(initialHash)
		assert.Equal(t, 1, len(changedFiles))
	})

	t.Run("Checking the recognition of updated files", func(t *testing.T) {
		// Creating a temporary directory to run the tests on.
		tempDir, err := os.MkdirTemp("", "git-test-*")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		cli := &Cli{RootDirectory: tempDir}
		// Copying the test-data directory containing assets for the tests to be run on.
		err = copyDirectory(testData, tempDir)
		require.NoError(t, err)

		// Initializing the temporary directory as a git repo.
		repo, err := git.PlainInit(tempDir, false)
		require.NoError(t, err)

		worktree, err := repo.Worktree()
		require.NoError(t, err)

		// Adding all files to git staging area.
		_, err = worktree.Add(".")
		require.NoError(t, err)

		// Creating initial commit with all files.
		commit, err := worktree.Commit("Initial commit", &git.CommitOptions{
			Author: &object.Signature{
				Name:  "Test User",
				Email: "test@example.com",
				When:  time.Now(),
			},
		})

		// Getting the hash of the initial commit for comparison.
		obj, err := repo.CommitObject(commit)
		require.NoError(t, err)

		hash := obj.Hash.String()

		// Modifying an existing file to test file update detection.
		fp := filepath.Join(tempDir, "test1.txt")

		file, err := os.OpenFile(fp, os.O_APPEND|os.O_WRONLY, 0644)
		require.NoError(t, err, "could not open the file")
		defer file.Close()

		// Adding content to the file.
		_, err = file.WriteString("Globstar peak.")
		require.NoError(t, err)

		// Adding the modified file to git staging area.
		_, err = worktree.Add(".")
		require.NoError(t, err)

		// Creating a second commit with the modified file.
		_, err = worktree.Commit("Edited file", &git.CommitOptions{
			Author: &object.Signature{
				Name:  "Test User",
				Email: "test@example.com",
				When:  time.Now(),
			},
		})

		require.NoError(t, err)

		// Checking if the modified file is detected in the changes.
		changedFiles, err := cli.GetChangedFiles(hash)
		require.NoError(t, err)

		assert.Equal(t, 1, len(changedFiles))
		assert.Contains(t, changedFiles, fp, "Expected to contain test1.txt in the changed files :/")
	})

	t.Run("Checking the recognition of renamed files", func(t *testing.T) {
		// Creating a temporary directory to run the tests on.
		tempDir, err := os.MkdirTemp("", "git-test-*")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		cli := &Cli{RootDirectory: tempDir}
		// Copying the test-data directory containing assets for the tests to be run on.
		err = copyDirectory(testData, tempDir)
		require.NoError(t, err)

		// Initializing the temporary directory as a git repo.
		repo, err := git.PlainInit(tempDir, false)
		require.NoError(t, err)

		worktree, err := repo.Worktree()
		require.NoError(t, err)

		// Adding all files to git staging area.
		_, err = worktree.Add(".")
		require.NoError(t, err)

		// Creating initial commit with all files.
		commit, err := worktree.Commit("Initial commit ", &git.CommitOptions{
			Author: &object.Signature{
				Name:  "Test User",
				Email: "test@example.com",
				When:  time.Now(),
			},
		})
		require.NoError(t, err)

		// Getting the hash of the initial commit for comparison.
		obj, err := repo.CommitObject(commit)
		require.NoError(t, err)
		initialHash := obj.Hash.String()

		// Renaming a file to test file rename detection.
		fp := filepath.Join(tempDir, "test2.txt")
		newFp := filepath.Join(tempDir, "test2_rename.txt")
		os.Rename(fp, filepath.Join(tempDir, "test2_rename.txt"))

		// Adding the renamed file to git staging area.
		_, err = worktree.Add(".")

		// Creating a second commit with the renamed file.
		_, err = worktree.Commit("Rename file", &git.CommitOptions{
			Author: &object.Signature{
				Name:  "Test User",
				Email: "test@example.com",
				When:  time.Now(),
			},
		})
		require.NoError(t, err)

		// Checking if the renamed file is detected in the changes.
		changedFiles, err := cli.GetChangedFiles(initialHash)
		require.NoError(t, err)
		assert.Equal(t, 1, len(changedFiles))

		assert.Contains(t, changedFiles, newFp)
	})

	t.Run("Recognition of changed file in the working directory along with the commit related to the hash provided.", func(t *testing.T) {
		// Creating a temporary directory to run the tests on.
		tempDir, err := os.MkdirTemp("", "git-test-*")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		cli := &Cli{RootDirectory: tempDir}
		// Copying the test-data directory containing assets for the tests to be run on.
		err = copyDirectory(testData, tempDir)
		require.NoError(t, err, "Failed to copy test data directory")

		// Initializing the temporary directory as a git repo.
		repo, err := git.PlainInit(tempDir, false)
		require.NoError(t, err)

		worktree, err := repo.Worktree()
		require.NoError(t, err)

		// Adding all files to git staging area.
		_, err = worktree.Add(".")
		require.NoError(t, err)

		// Creating initial commit with all files.
		commit, err := worktree.Commit("Initial commit", &git.CommitOptions{
			Author: &object.Signature{
				Name:  "Test User",
				Email: "test@example.com",
				When:  time.Now(),
			},
		})

		// Creating a new file to test file addition detection.
		filePath := filepath.Join(tempDir, "test3.txt")
		if err := os.WriteFile(filePath, []byte(""), 0644); err != nil {
			require.NoError(t, err)
		}

		// Adding the new file to git staging area.
		_, err = worktree.Add("test3.txt")

		// Creating a second commit with the new file.
		_, err = worktree.Commit("Add a new file", &git.CommitOptions{
			Author: &object.Signature{
				Name:  "Test User",
				Email: "test@example.com",
				When:  time.Now(),
			},
		})
		require.NoError(t, err)

		// Getting the hash of the initial commit for comparison.
		obj, err := repo.CommitObject(commit)
		initialHash := obj.Hash.String()
		require.NoError(t, err)

		// Creating an uncommitted file to test working directory changes.
		createdFileName := filepath.Join(cli.RootDirectory, "test4.txt")
		_, err = os.Create(createdFileName)
		require.NoError(t, err)

		// Checking if both committed and uncommitted changes are detected.
		changedFiles, err := cli.GetChangedFiles(initialHash)

		assert.Contains(t, changedFiles, filePath)
		assert.Contains(t, changedFiles, createdFileName)
	})

}
