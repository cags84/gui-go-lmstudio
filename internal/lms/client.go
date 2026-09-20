package lms

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// Client is the typed API over a Runner. Besides Runner implementations
// themselves, it is the only place that knows the lms command-line vocabulary
// (subcommands and flags); everything else in the app talks to Client.
type Client struct {
	runner Runner
}

// New builds a Client backed by r.
func New(r Runner) *Client {
	return &Client{runner: r}
}

// ModelFilter selects which model types ListModels returns. It is a small
// closed enum rather than a bare string so a caller cannot accidentally build
// an lms flag that does not exist.
type ModelFilter int

const (
	AllModels ModelFilter = iota
	LLMModelsOnly
	EmbeddingModelsOnly
)

// ServerStatus runs `lms server status --json` and parses its output.
func (c *Client) ServerStatus(ctx context.Context) (ServerStatus, error) {
	args := []string{"server", "status", "--json"}
	out, err := c.runner.Run(ctx, args...)
	if err != nil {
		return ServerStatus{}, err
	}
	var status ServerStatus
	if err := json.Unmarshal(out, &status); err != nil {
		return ServerStatus{}, fmt.Errorf("parse output of %q: %w", "lms "+strings.Join(args, " "), err)
	}
	return status, nil
}

// ListModels runs `lms ls --json`, adding --llm or --embedding when filter
// asks for one model type, and parses the result.
func (c *Client) ListModels(ctx context.Context, filter ModelFilter) ([]Model, error) {
	args := []string{"ls", "--json"}
	switch filter {
	case LLMModelsOnly:
		args = append(args, "--llm")
	case EmbeddingModelsOnly:
		args = append(args, "--embedding")
	}

	out, err := c.runner.Run(ctx, args...)
	if err != nil {
		return nil, err
	}
	var wire []modelWire
	if err := json.Unmarshal(out, &wire); err != nil {
		return nil, fmt.Errorf("parse output of %q: %w", "lms "+strings.Join(args, " "), err)
	}

	models := make([]Model, 0, len(wire))
	for _, w := range wire {
		models = append(models, w.toModel())
	}
	return models, nil
}

// LoadedModels runs `lms ps --json` and parses the result. Note that this
// command wakes the lms daemon as a side effect, so an empty result is not, by
// itself, proof that the daemon was never running.
func (c *Client) LoadedModels(ctx context.Context) ([]LoadedModel, error) {
	args := []string{"ps", "--json"}
	out, err := c.runner.Run(ctx, args...)
	if err != nil {
		return nil, err
	}
	var wire []loadedModelWire
	if err := json.Unmarshal(out, &wire); err != nil {
		return nil, fmt.Errorf("parse output of %q: %w", "lms "+strings.Join(args, " "), err)
	}

	loaded := make([]LoadedModel, 0, len(wire))
	for _, w := range wire {
		loaded = append(loaded, w.toLoadedModel())
	}
	return loaded, nil
}
