package container

import (
	"fmt"
	"os"
	"os/exec"
)

// Runtime represents the container runtime
type Runtime struct {
	Command string
	Args    []string
}

// NewRuntime creates a new container runtime
func NewRuntime(command string, args []string) *Runtime {
	return &Runtime{
		Command: command,
		Args:    args,
	}
}

// Run runs the application in container mode
func (r *Runtime) Run() error {
	// Create command
	cmd := exec.Command(r.Command, r.Args...)

	// Set stdout and stderr to the same as the parent process
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// Set environment variables
	cmd.Env = os.Environ()

	// Run the command
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run command: %v", err)
	}

	return nil
}
