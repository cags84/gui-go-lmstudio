package lms

import (
	"context"
	"testing"
	"time"
)

// TestServerStatus_Integration exercises the real lms binary. ServerStatus is
// read-only (it never starts, stops, loads or unloads anything), so it is safe
// to run against whatever LM Studio installation happens to be on the host.
// It is skipped under -short, and skipped (not failed) when lms is not
// installed, so CI machines without LM Studio stay green.
func TestServerStatus_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	runner, err := NewExecRunner()
	if err != nil {
		t.Skipf("lms binary not available: %v", err)
	}
	client := New(runner)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	status, err := client.ServerStatus(ctx)
	if err != nil {
		t.Fatalf("ServerStatus() error = %v", err)
	}
	t.Logf("real lms server status: running=%v port=%d", status.Running, status.Port)
}
