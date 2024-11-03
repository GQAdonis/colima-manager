package colima

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/gqadonis/colima-manager/internal/domain"
	"github.com/gqadonis/colima-manager/internal/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockExecutor implements the Executor interface for testing
type mockExecutor struct {
	commands map[string]mockOutput
}

type mockOutput struct {
	output []byte
	err    error
}

// mockCommand implements the Command interface
type mockCommand struct {
	mockOutput mockOutput
}

func (c *mockCommand) Output() ([]byte, error) {
	return c.mockOutput.output, c.mockOutput.err
}

func (c *mockCommand) CombinedOutput() ([]byte, error) {
	return c.mockOutput.output, c.mockOutput.err
}

func (c *mockCommand) Run() error {
	return c.mockOutput.err
}

// Command returns a new mockCommand that implements the Command interface
func (m *mockExecutor) Command(name string, args ...string) Command {
	// Build the command string to match exactly what's being requested
	cmdStr := name
	if len(args) > 0 {
		cmdStr = name + " " + strings.Join(args, " ")
	}

	// Get mock output if it exists, otherwise use empty output
	output, ok := m.commands[cmdStr]
	if !ok {
		output = mockOutput{
			output: []byte(""),
			err:    nil,
		}
	}

	return &mockCommand{
		mockOutput: output,
	}
}

func TestCheckDependencies(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)

	tests := []struct {
		name           string
		commands       map[string]mockOutput
		expectedStatus *domain.DependencyStatus
		expectError    bool
	}{
		{
			name: "all dependencies present",
			commands: map[string]mockOutput{
				"brew --prefix": {
					output: []byte("/usr/local/bin"),
					err:    nil,
				},
				"which colima": {
					output: []byte("/usr/local/bin/colima"),
					err:    nil,
				},
				"colima version": {
					output: []byte("0.6.0"),
					err:    nil,
				},
				"brew list --versions lima": {
					output: []byte("lima 0.6.0"),
					err:    nil,
				},
			},
			expectedStatus: &domain.DependencyStatus{
				Homebrew:      true,
				HomebrewPath:  "/usr/local/bin",
				Colima:        true,
				ColimaVersion: "0.6.0",
				Lima:          true,
				LimaVersion:   "0.6.0",
			},
			expectError: false,
		},
		{
			name: "homebrew missing",
			commands: map[string]mockOutput{
				"brew --prefix": {
					output: nil,
					err:    os.ErrNotExist,
				},
			},
			expectedStatus: &domain.DependencyStatus{
				Homebrew: false,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := &mockExecutor{commands: tt.commands}
			repo := &ColimaRepository{
				homeDir: homeDir,
				log:     logger.GetLogger(),
				exec:    mockExec,
			}

			status, err := repo.CheckDependencies(context.Background())
			if tt.expectError {
				assert.Error(t, err)
				assert.IsType(t, &domain.DependencyError{}, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatus.Homebrew, status.Homebrew)
				assert.Equal(t, tt.expectedStatus.Colima, status.Colima)
				assert.Equal(t, tt.expectedStatus.Lima, status.Lima)
				if tt.expectedStatus.Lima {
					assert.Equal(t, tt.expectedStatus.LimaVersion, status.LimaVersion)
				}
			}
		})
	}
}
