package main

import (
	"fmt"
	"os"
	"os/exec"
)

// Vulnerable: Direct use of unsanitized user input in command execution.
// <expect-error> command injection
func test1(userInput string) {

	// Creating an exec.Cmd with user-provided Path (potential injection point)
	cmd := &exec.Cmd{
		Path:   userInput,
		Args:   []string{"foo", "bar"},
		Stdout: os.Stdout,
		Stderr: os.Stdout,
	}

	fmt.Println("Running potentially unsafe command...")
	err := cmd.Start()
	if err != nil {
		fmt.Printf("Error starting command: %v\n", err)
	}
}


func sanitizeInput(input string) string {
	// Remove potentially dangerous characters
	safeInput := strings.TrimSpace(input)
	safeInput = filepath.Clean(safeInput)

	// Allow only absolute paths without shell metacharacters
	if !filepath.IsAbs(safeInput) || strings.ContainsAny(safeInput, "&;|<>") {
		fmt.Println("Invalid command path detected.")
		return ""
	}

	return safeInput
}

// Safe: Sanitized input used for command execution.
func test2(userInput string) {
	// Sanitize user input to prevent command injection
	cmdPath := sanitizeInput(userInput)
	if cmdPath == "" {
		fmt.Println("⚠️ Command execution aborted due to unsafe input.")
		return
	}

	// Create a safe exec.Cmd with validated Path
	cmd := &exec.Cmd{
		Path:   cmdPath,
		Args:   []string{"foo", "bar"},
		Stdout: os.Stdout,
		Stderr: os.Stdout,
	}

	fmt.Println("✅ Running sanitized command...")
	err := cmd.Start()
	if err != nil {
		fmt.Printf("Error starting command: %v\n", err)
	}
}

