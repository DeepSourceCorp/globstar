package cli

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

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

		if d.IsDir(){
			return os.MkdirAll(dstPath, os.ModePerm)
		}

		return copyFile(path, dstPath)

	})
}


func copyFile(src, dst string) error{
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

 func TestGetChangedFiles(t *testing.T){
	tempDir, err := os.MkdirTemp("", "git-test-*")
	require.NoError(t,err)
	defer os.RemoveAll(tempDir)

	testData := "testdata"

	err = copyDirectory(testData, tempDir)
	require.NoError(t, err, "Failed to copy test data directory")
 }