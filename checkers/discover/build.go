package discover

import (
	"fmt"
	"path/filepath"

	"globstar.dev/util"
)

// Build the analyzer binary from the given directory to the destination directory
func BuildAnalyzer(analyzerDir, dest string) error {
	// install all the dependencies from go.mod in the analyzerDir
	_, _, err := util.RunCmd("go", []string{"mod", "download"}, analyzerDir)
	if err != nil {
		return fmt.Errorf("error installing dependencies: %v", err)
	}

	// build the analyzer binary
	customAnalyzerName := "custom-analyzer"
	stdout, stderr, err := util.RunCmd("go", []string{"build", "-o", filepath.Join(dest, customAnalyzerName), "."}, analyzerDir)
	if err != nil {
		err = fmt.Errorf("error building analyzer: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
		return fmt.Errorf("error building analyzer: %v", err)
	}

	return nil
}
