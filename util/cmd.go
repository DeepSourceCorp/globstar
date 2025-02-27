package util

import (
	"bytes"
	"os/exec"
)

func RunCmd(command string, args []string, cmdDir string) (string, string, error) {
	cmd := exec.Command(command, args...)
	cmd.Dir = cmdDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return stdout.String(), stderr.String(), err
	}

	outStr, errStr := stdout.String(), stderr.String()

	return outStr, errStr, nil
}
