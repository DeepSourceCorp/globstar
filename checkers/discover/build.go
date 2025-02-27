package discover

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Build the analyzer binary from the given directory to the destination directory
func BuildAnalyzer(analyzerDir, dest string) error {
	// install all the dependencies from go.mod in the analyzerDir
	_, _, err := runCmd("go", []string{"mod", "download"}, analyzerDir)
	if err != nil {
		return fmt.Errorf("error installing dependencies: %v", err)
	}

	// build the analyzer binary
	customAnalyzerName := "custom-analyzer"
	_, _, err = runCmd("go", []string{"build", "-o", filepath.Join(dest, customAnalyzerName), "."}, analyzerDir)
	if err != nil {
		return fmt.Errorf("error building analyzer: %v", err)
	}

	return nil
}

func runCmd(command string, args []string, cmdDir string) (string, string, error) {
	cmd := exec.Command(command, args...)
	cmd.Dir = cmdDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	startTime := time.Now()

	err := cmd.Run()
	if err != nil {
		log.Printf("cmd.Run() failed with %s\n", err)
		return stdout.String(), stderr.String(), err
	}

	log.Println("==> EXEC COMMAND " + command + " " + strings.Join(args, " ") + " DIR - " + cmdDir)

	outStr, errStr := stdout.String(), stderr.String()

	log.Println("==> EXEC COMMAND ENDED. TIME TAKEN: ", time.Since(startTime))

	return outStr, errStr, nil
}
