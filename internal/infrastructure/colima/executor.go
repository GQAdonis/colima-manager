package colima

import (
	"os/exec"
)

// Command defines the interface for command execution
type Command interface {
	Output() ([]byte, error)
	CombinedOutput() ([]byte, error)
	Run() error
}

// Executor defines the interface for executing commands
type Executor interface {
	Command(name string, args ...string) Command
}

// RealExecutor implements Executor using real system commands
type RealExecutor struct{}

// RealCommand wraps exec.Cmd to implement Command interface
type RealCommand struct {
	*exec.Cmd
}

func (c *RealCommand) Output() ([]byte, error) {
	return c.Cmd.Output()
}

func (c *RealCommand) CombinedOutput() ([]byte, error) {
	return c.Cmd.CombinedOutput()
}

func (c *RealCommand) Run() error {
	return c.Cmd.Run()
}

func (e *RealExecutor) Command(name string, args ...string) Command {
	return &RealCommand{Cmd: exec.Command(name, args...)}
}

// NewRealExecutor creates a new RealExecutor
func NewRealExecutor() Executor {
	return &RealExecutor{}
}
