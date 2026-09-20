package lms

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

// TestHelperProcess is not a real test. It is re-executed as a fake "lms" binary by
// pointing ExecRunner.binary at the current test executable (os.Args[0]) and passing
// "-test.run=TestHelperProcess" as its first argument, following the standard library's
// own pattern for testing os/exec without shelling out to a real external program. When
// run normally (GO_WANT_HELPER_PROCESS unset) it does nothing, so it shows up as an
// ordinary passing test in `go test -v` output.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	defer os.Exit(0)

	if out := os.Getenv("LMS_HELPER_STDOUT"); out != "" {
		io.WriteString(os.Stdout, out)
	}
	if out := os.Getenv("LMS_HELPER_STDERR"); out != "" {
		io.WriteString(os.Stderr, out)
	}
	if line := os.Getenv("LMS_HELPER_STREAM_LINE"); line != "" {
		io.WriteString(os.Stdout, line+"\n")
		os.Stdout.Sync()
		time.Sleep(10 * time.Second) // outlives the test; must be killed by ctx cancellation
	}
	if code := os.Getenv("LMS_HELPER_EXIT_CODE"); code != "" {
		n, err := strconv.Atoi(code)
		if err != nil {
			panic(err)
		}
		os.Exit(n)
	}
	os.Exit(0)
}

// helperRunner returns an ExecRunner whose "binary" is this test executable itself,
// re-invoked as TestHelperProcess above. This exercises the real exec.Cmd plumbing
// (pipes, exit codes, context cancellation) without depending on the real lms binary.
func helperRunner(t *testing.T) *ExecRunner {
	t.Helper()
	t.Setenv("GO_WANT_HELPER_PROCESS", "1")
	return &ExecRunner{binary: os.Args[0]}
}

func TestRunError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  *RunError
		want string
	}{
		{
			name: "includes exit code, args and trimmed output",
			err:  &RunError{Args: []string{"server", "start", "--port", "1234"}, ExitCode: 1, Output: []byte("Error: port 1234 already in use\n")},
			want: "lms server start --port 1234: exit code 1: Error: port 1234 already in use",
		},
		{
			name: "empty output still reports exit code",
			err:  &RunError{Args: []string{"ls", "--json"}, ExitCode: 2, Output: nil},
			want: "lms ls --json: exit code 2: ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExecRunner_Run_Success(t *testing.T) {
	runner := helperRunner(t)
	t.Setenv("LMS_HELPER_STDOUT", `{"running":false,"port":1234}`)

	out, err := runner.Run(context.Background(), "-test.run=TestHelperProcess")
	if err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}
	if string(out) != `{"running":false,"port":1234}` {
		t.Errorf("Run() output = %q, want the stdout JSON", out)
	}
}

func TestExecRunner_Run_NonZeroExit(t *testing.T) {
	runner := helperRunner(t)
	t.Setenv("LMS_HELPER_STDERR", "Error: port 1234 already in use\n")
	t.Setenv("LMS_HELPER_EXIT_CODE", "1")

	out, err := runner.Run(context.Background(), "-test.run=TestHelperProcess")
	if out != nil {
		t.Errorf("Run() output = %q, want nil on failure", out)
	}
	var runErr *RunError
	if !errors.As(err, &runErr) {
		t.Fatalf("Run() error = %v (%T), want *RunError", err, err)
	}
	if runErr.ExitCode != 1 {
		t.Errorf("ExitCode = %d, want 1", runErr.ExitCode)
	}
	if !strings.Contains(string(runErr.Output), "port 1234 already in use") {
		t.Errorf("Output = %q, want it to contain the CLI's own message", runErr.Output)
	}
}

func TestExecRunner_Stream_CancelTerminatesProcess(t *testing.T) {
	runner := helperRunner(t)
	t.Setenv("LMS_HELPER_STREAM_LINE", "log line one")

	ctx, cancel := context.WithCancel(context.Background())
	stdout, wait, err := runner.Stream(ctx, "-test.run=TestHelperProcess")
	if err != nil {
		t.Fatalf("Stream() error = %v", err)
	}
	defer stdout.Close()

	buf := make([]byte, len("log line one\n"))
	if _, err := io.ReadFull(stdout, buf); err != nil {
		t.Fatalf("reading first line: %v", err)
	}
	if string(buf) != "log line one\n" {
		t.Errorf("first line = %q, want %q", buf, "log line one\n")
	}

	cancel()
	// wait must return once the context cancellation kills the still-sleeping
	// child; if cancellation did not terminate it, this blocks for 10s and the
	// test times out.
	if err := wait(); err == nil {
		t.Error("wait() error = nil, want an error from the killed process")
	}
}

func TestResolveBinary(t *testing.T) {
	notFound := func(string) (string, error) { return "", errors.New("not found") }

	t.Run("found on PATH", func(t *testing.T) {
		wantPath := filepath.Join(t.TempDir(), "lms")
		lookPath := func(name string) (string, error) {
			if name != "lms" {
				t.Fatalf("lookPath called with %q, want %q", name, "lms")
			}
			return wantPath, nil
		}
		homeDir := func() (string, error) {
			t.Fatal("homeDir should not be consulted when PATH lookup succeeds")
			return "", nil
		}

		got, err := resolveBinary(lookPath, homeDir, runtime.GOOS)
		if err != nil {
			t.Fatalf("resolveBinary() error = %v", err)
		}
		if got != wantPath {
			t.Errorf("resolveBinary() = %q, want %q", got, wantPath)
		}
	})

	t.Run("found at fallback install location", func(t *testing.T) {
		home := t.TempDir()
		name := "lms"
		goos := "linux"
		if runtime.GOOS == "windows" {
			name = "lms.exe"
			goos = "windows"
		}
		fallback := filepath.Join(home, ".lmstudio", "bin", name)
		if err := os.MkdirAll(filepath.Dir(fallback), 0o755); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}
		if err := os.WriteFile(fallback, []byte("#!/bin/sh\n"), 0o755); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
		homeDir := func() (string, error) { return home, nil }

		got, err := resolveBinary(notFound, homeDir, goos)
		if err != nil {
			t.Fatalf("resolveBinary() error = %v", err)
		}
		if got != fallback {
			t.Errorf("resolveBinary() = %q, want %q", got, fallback)
		}
	})

	t.Run("not installed", func(t *testing.T) {
		home := t.TempDir() // empty: no fallback binary present
		homeDir := func() (string, error) { return home, nil }

		_, err := resolveBinary(notFound, homeDir, runtime.GOOS)
		if !errors.Is(err, ErrNotInstalled) {
			t.Fatalf("resolveBinary() error = %v, want ErrNotInstalled", err)
		}
	})

	t.Run("not installed when home directory is unavailable", func(t *testing.T) {
		homeDir := func() (string, error) { return "", errors.New("no home") }

		_, err := resolveBinary(notFound, homeDir, runtime.GOOS)
		if !errors.Is(err, ErrNotInstalled) {
			t.Fatalf("resolveBinary() error = %v, want ErrNotInstalled", err)
		}
	})
}

func TestNewExecRunner_InterfaceCompliance(t *testing.T) {
	var _ Runner = (*ExecRunner)(nil)
}
