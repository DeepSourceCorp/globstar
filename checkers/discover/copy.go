package discover

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// We use go.mod.embed because embedding is not supported for go.mod files
//
//go:embed custom_analyzer_stub/*.go custom_analyzer_stub/go.mod.embed  custom_analyzer_stub/go.sum.embed
var AnalyzerStubs embed.FS

// Copy all the files from the source directory to the destination directory
func Copy(src, dest string, copyFile func(string) bool) error {
	err := os.MkdirAll(dest, 0755)
	if err != nil {
		return err
	}

	files, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, file := range files {
		if !copyFile(file.Name()) {
			continue
		}

		srcFile := filepath.Join(src, file.Name())
		destFile := filepath.Join(dest, file.Name())

		if file.IsDir() {
			err = Copy(srcFile, destFile, copyFile)
			if err != nil {
				return err
			}
		} else {
			data, err := os.ReadFile(srcFile)
			if err != nil {
				return err
			}

			err = os.WriteFile(destFile, data, 0644)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Copy all the files recursively from the analyzerFiles embed.FS to the given directory
func CopyEmbeddedFiles(embeddedFiles embed.FS, dest string) error {
	err := os.MkdirAll(dest, 0755)
	if err != nil {
		return err
	}

	files, err := embeddedFiles.ReadDir("custom_analyzer_stub")
	if err != nil {
		return err
	}

	for _, file := range files {
		data, err := embeddedFiles.ReadFile(fmt.Sprintf("custom_analyzer_stub/%s", file.Name()))
		if err != nil {
			return err
		}

		fileName := file.Name()
		if strings.HasSuffix(file.Name(), ".embed") {
			fileName = strings.TrimSuffix(file.Name(), ".embed")
		}

		err = os.WriteFile(filepath.Join(dest, fileName), data, 0755)
		if err != nil {
			return err
		}
	}

	return nil
}
