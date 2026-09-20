package lms

import (
	"context"
	"fmt"
	"io"
)

// FakeRunner is a Runner test double. Each test configures RunFunc/StreamFunc to
// script the bytes or error a given call should produce, and reads Calls afterwards
// to assert on the exact argv Client built — building the wrong argv is a real
// defect this package exists to prevent, so tests check it explicitly rather than
// only checking the parsed result.
type FakeRunner struct {
	RunFunc    func(ctx context.Context, args ...string) ([]byte, error)
	StreamFunc func(ctx context.Context, args ...string) (io.ReadCloser, func() error, error)
	Calls      [][]string
}

func (f *FakeRunner) Run(ctx context.Context, args ...string) ([]byte, error) {
	f.Calls = append(f.Calls, append([]string(nil), args...))
	if f.RunFunc == nil {
		return nil, fmt.Errorf("FakeRunner.Run: no RunFunc configured for args %v", args)
	}
	return f.RunFunc(ctx, args...)
}

func (f *FakeRunner) Stream(ctx context.Context, args ...string) (io.ReadCloser, func() error, error) {
	f.Calls = append(f.Calls, append([]string(nil), args...))
	if f.StreamFunc == nil {
		return nil, nil, fmt.Errorf("FakeRunner.Stream: no StreamFunc configured for args %v", args)
	}
	return f.StreamFunc(ctx, args...)
}

var _ Runner = (*FakeRunner)(nil)
