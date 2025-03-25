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

func printDirectoryContents(tempDir string, t *testing.T) {
	t.Log("Directory structure after commits:")
	err := filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path from tempDir for cleaner output
		relPath, err := filepath.Rel(tempDir, path)
		if err != nil {
			relPath = path // Fallback to full path if relative fails
		}

		// Skip the root directory itself
		if relPath == "." {
			return nil
		}

		// Log each file/directory with its type
		if info.IsDir() {
			t.Logf("DIR: %s", relPath)
		} else {
			t.Logf("FILE: %s (%d bytes)", relPath, info.Size())
		}

		return nil
	})
	require.NoError(t, err, "Failed to walk directory structure")
}

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

	t.Run("Checking the recognition of an added file between two commits", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "git-test-*")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		cli := &Cli{RootDirectory: tempDir}
		err = copyDirectory(testData, tempDir)
		require.NoError(t, err, "Failed to copy test data directory")

		repo, err := git.PlainInit(tempDir, false)
		require.NoError(t, err)

		worktree, err := repo.Worktree()
		require.NoError(t, err)

		_, err = worktree.Add(".")
		require.NoError(t, err)

		commit, err := worktree.Commit("Initial commit", &git.CommitOptions{
			Author: &object.Signature{
				Name:  "Test User",
				Email: "test@example.com",
				When:  time.Now(),
			},
		})

		filePath := filepath.Join(tempDir, "test3.txt")
		if err := os.WriteFile(filePath, []byte(""), 0644); err != nil {
			require.NoError(t, err)
		}

		_, err = worktree.Add("test3.txt")

		_, err = worktree.Commit("Add a new file", &git.CommitOptions{
			Author: &object.Signature{
				Name:  "Test User",
				Email: "test@example.com",
				When:  time.Now(),
			},
		})
		require.NoError(t, err)

		obj, err := repo.CommitObject(commit)
		initialHash := obj.Hash.String()
		require.NoError(t, err)

		changedFiles, err := cli.GetChangedFiles(initialHash)
		assert.Equal(t, 1, len(changedFiles))
	})

	t.Run("Checking the recognition of updated files", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "git-test-*")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		cli := &Cli{RootDirectory: tempDir}
		err = copyDirectory(testData, tempDir)
		require.NoError(t, err)

		repo, err := git.PlainInit(tempDir, false)
		require.NoError(t, err)

		worktree, err := repo.Worktree()
		require.NoError(t, err)

		_, err = worktree.Add(".")
		require.NoError(t, err)

		commit, err := worktree.Commit("Initial commit", &git.CommitOptions{
			Author: &object.Signature{
				Name:  "Test User",
				Email: "test@example.com",
				When:  time.Now(),
			},
		})

		obj, err := repo.CommitObject(commit)
		require.NoError(t, err)

		hash := obj.Hash.String()

		fp := filepath.Join(tempDir, "test1.txt")

		file, err := os.OpenFile(fp, os.O_APPEND|os.O_WRONLY, 0644)
		require.NoError(t, err, "could not open the file")
		defer file.Close()

		_, err = file.WriteString("Globstar peak.")
		require.NoError(t, err)

		_, err = worktree.Add(".")
		require.NoError(t, err)

		_, err = worktree.Commit("Edited file", &git.CommitOptions{
			Author: &object.Signature{
				Name:  "Test User",
				Email: "test@example.com",
				When:  time.Now(),
			},
		})

		require.NoError(t, err)

		changedFiles, err := cli.GetChangedFiles(hash)
		require.NoError(t, err)

		assert.Equal(t, 1, len(changedFiles))
		assert.Contains(t, changedFiles, fp, "Expected to contain test1.txt in the changed files :/")

	})

	t.Run("Checking the recognition of renamed files", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "git-test-*")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		cli := &Cli{RootDirectory: tempDir}
		err = copyDirectory(testData, tempDir)
		require.NoError(t, err)
		// Create and initialize repo
		repo, err := git.PlainInit(tempDir, false)
		require.NoError(t, err)

		worktree, err := repo.Worktree()
		require.NoError(t, err)

		_, err = worktree.Add(".")
		require.NoError(t, err)

		commit, err := worktree.Commit("Initial commit ", &git.CommitOptions{
			Author: &object.Signature{
				Name:  "Test User",
				Email: "test@example.com",
				When:  time.Now(),
			},
		})
		require.NoError(t, err)

		// Get initial commit hash
		obj, err := repo.CommitObject(commit)
		require.NoError(t, err)
		initialHash := obj.Hash.String()
		fp := filepath.Join(tempDir, "test2.txt")
		newFp := filepath.Join(tempDir, "test2_rename.txt")
		os.Rename(fp, filepath.Join(tempDir, "test2_rename.txt"))

		_, err = worktree.Add(".")

		_, err = worktree.Commit("Rename file", &git.CommitOptions{
			Author: &object.Signature{
				Name:  "Test User",
				Email: "test@example.com",
				When:  time.Now(),
			},
		})
		require.NoError(t, err)

		// Check changes
		changedFiles, err := cli.GetChangedFiles(initialHash)
		require.NoError(t, err)
		assert.Equal(t, 1, len(changedFiles))

		assert.Contains(t, changedFiles, newFp)
	})

}
