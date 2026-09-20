// Package lms is the sole boundary between this application and the lms CLI
// (LM Studio's command-line tool). No other package may call os/exec directly:
// everything the app does with LM Studio goes through Runner and Client here, so
// the rest of the codebase can be tested without spawning a real process.
package lms

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// ErrNotInstalled is returned (wrapped) by NewExecRunner when the lms binary
// cannot be found on PATH or at any known fallback install location. It is a
// sentinel, not a string, so the UI can show a distinct "please install LM
// Studio" message instead of a generic exec failure.
var ErrNotInstalled = errors.New("lms: binary not found on PATH or in known install locations")

// RunError is returned by Runner.Run when the child process exits with a
// non-zero status. lms explains most failures (for example a port already in
// use) as plain text on stdout/stderr rather than through a distinctive exit
// code, so callers need the captured output text, not just success/failure, to
// react to it or show it to the user. Fields are exported so callers can
// inspect the exit code and output directly instead of string-matching
// Error()'s formatted message.
type RunError struct {
	Args     []string
	ExitCode int
	Output   []byte
}

func (e *RunError) Error() string {
	return fmt.Sprintf("lms %s: exit code %d: %s", strings.Join(e.Args, " "), e.ExitCode, strings.TrimSpace(string(e.Output)))
}

// Runner executes lms commands. Implementations must not interpret output:
// parsing JSON is Client's job, so Runner stays a thin, swappable process
// boundary (ExecRunner for real use, FakeRunner in tests).
type Runner interface {
	// Run executes a one-shot lms command and returns its stdout. On a
	// non-zero exit it returns a *RunError instead of the output.
	Run(ctx context.Context, args ...string) ([]byte, error)

	// Stream starts a long-running lms command (such as `lms log stream`) and
	// returns its stdout pipe plus a wait function the caller must call after
	// closing the reader. Cancelling ctx terminates the child process.
	Stream(ctx context.Context, args ...string) (io.ReadCloser, func() error, error)
}

// ExecRunner is the real Runner, backed by the lms binary resolved once at
// construction time.
type ExecRunner struct {
	binary string
}

// NewExecRunner resolves the lms binary (PATH first, then the known fallback
// install locations) and returns a ready-to-use ExecRunner. It returns an error
// wrapping ErrNotInstalled when the binary cannot be found anywhere, so callers
// can distinguish "not installed" from a later command failure.
func NewExecRunner() (*ExecRunner, error) {
	binary, err := resolveBinary(exec.LookPath, os.UserHomeDir, runtime.GOOS)
	if err != nil {
		return nil, err
	}
	return &ExecRunner{binary: binary}, nil
}

// resolveBinary implements the lookup rule with injectable dependencies so it
// can be tested without touching the real PATH or the real home directory.
func resolveBinary(lookPath func(string) (string, error), homeDir func() (string, error), goos string) (string, error) {
	if path, err := lookPath("lms"); err == nil {
		return path, nil
	}

	home, err := homeDir()
	if err != nil {
		return "", fmt.Errorf("resolve lms binary: %w", ErrNotInstalled)
	}

	name := "lms"
	if goos == "windows" {
		name = "lms.exe"
	}
	fallback := filepath.Join(home, ".lmstudio", "bin", name)
	if _, err := os.Stat(fallback); err == nil {
		return fallback, nil
	}

	return "", fmt.Errorf("resolve lms binary: %w", ErrNotInstalled)
}

// Run implements Runner.
func (r *ExecRunner) Run(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, r.binary, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			// Concatenate both streams: lms sometimes explains a failure on
			// stdout, sometimes on stderr, and the caller should not have to
			// guess which.
			output := append([]byte(nil), stdout.Bytes()...)
			if stderr.Len() > 0 {
				if len(output) > 0 {
					output = append(output, '\n')
				}
				output = append(output, stderr.Bytes()...)
			}
			return nil, &RunError{Args: args, ExitCode: exitErr.ExitCode(), Output: output}
		}
		return nil, fmt.Errorf("run lms %s: %w", strings.Join(args, " "), err)
	}

	// Only stdout is returned on success: mixing in stderr here would risk
	// corrupting the JSON that Client parses from this output.
	return stdout.Bytes(), nil
}

// Stream implements Runner.
func (r *ExecRunner) Stream(ctx context.Context, args ...string) (io.ReadCloser, func() error, error) {
	cmd := exec.CommandContext(ctx, r.binary, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("open stdout pipe for lms %s: %w", strings.Join(args, " "), err)
	}
	if err := cmd.Start(); err != nil {
		return nil, nil, fmt.Errorf("start lms %s: %w", strings.Join(args, " "), err)
	}

	// exec.CommandContext kills the process when ctx is cancelled, so the
	// caller cancelling the context is enough to stop a long-running stream
	// such as `lms log stream`; no extra plumbing is needed here.
	return stdout, cmd.Wait, nil
}
