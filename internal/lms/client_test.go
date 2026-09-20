package lms

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestClient_ServerStatus(t *testing.T) {
	tests := []struct {
		name string
		file string
		want ServerStatus
	}{
		{name: "stopped", file: "server_status_stopped.json", want: ServerStatus{Running: false, Port: 1234}},
		{name: "running", file: "server_status_running.json", want: ServerStatus{Running: true, Port: 1234}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := readTestdata(t, tt.file)
			fake := &FakeRunner{
				RunFunc: func(ctx context.Context, args ...string) ([]byte, error) {
					return data, nil
				},
			}
			client := New(fake)

			got, err := client.ServerStatus(context.Background())
			if err != nil {
				t.Fatalf("ServerStatus() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("ServerStatus() = %+v, want %+v", got, tt.want)
			}
			if len(fake.Calls) != 1 {
				t.Fatalf("Calls = %v, want exactly one call", fake.Calls)
			}
			wantArgv := []string{"server", "status", "--json"}
			if !equalArgv(fake.Calls[0], wantArgv) {
				t.Errorf("argv = %v, want %v", fake.Calls[0], wantArgv)
			}
		})
	}
}

func TestClient_ListModels_ParsesLsAll(t *testing.T) {
	fake := &FakeRunner{
		RunFunc: func(ctx context.Context, args ...string) ([]byte, error) {
			return readTestdata(t, "ls_all.json"), nil
		},
	}
	client := New(fake)

	models, err := client.ListModels(context.Background(), AllModels)
	if err != nil {
		t.Fatalf("ListModels() error = %v", err)
	}
	if len(models) != 11 {
		t.Fatalf("len(models) = %d, want 11", len(models))
	}

	var llmCount, embeddingCount, remoteCount int
	for _, m := range models {
		switch m.Type {
		case ModelTypeLLM:
			llmCount++
		case ModelTypeEmbedding:
			embeddingCount++
			if m.ParamsString != nil {
				t.Errorf("embedding model %s: ParamsString set, want nil", m.ModelKey)
			}
			if m.Vision != nil {
				t.Errorf("embedding model %s: Vision set, want nil", m.ModelKey)
			}
		}
		if m.IsRemote() {
			remoteCount++
		}
	}
	if llmCount != 9 {
		t.Errorf("llmCount = %d, want 9", llmCount)
	}
	if embeddingCount != 2 {
		t.Errorf("embeddingCount = %d, want 2", embeddingCount)
	}
	if remoteCount != 7 {
		t.Errorf("remoteCount = %d, want 7", remoteCount)
	}
}

func TestClient_ListModels_BuildsArgvPerFilter(t *testing.T) {
	tests := []struct {
		name     string
		filter   ModelFilter
		wantArgv []string
	}{
		{name: "all models", filter: AllModels, wantArgv: []string{"ls", "--json"}},
		{name: "llm only", filter: LLMModelsOnly, wantArgv: []string{"ls", "--json", "--llm"}},
		{name: "embedding only", filter: EmbeddingModelsOnly, wantArgv: []string{"ls", "--json", "--embedding"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &FakeRunner{
				RunFunc: func(ctx context.Context, args ...string) ([]byte, error) {
					return []byte("[]"), nil
				},
			}
			client := New(fake)

			if _, err := client.ListModels(context.Background(), tt.filter); err != nil {
				t.Fatalf("ListModels() error = %v", err)
			}
			if len(fake.Calls) != 1 {
				t.Fatalf("Calls = %v, want exactly one call", fake.Calls)
			}
			if !equalArgv(fake.Calls[0], tt.wantArgv) {
				t.Errorf("argv = %v, want %v", fake.Calls[0], tt.wantArgv)
			}
		})
	}
}

func TestClient_LoadedModels(t *testing.T) {
	t.Run("one llm loaded", func(t *testing.T) {
		fake := &FakeRunner{
			RunFunc: func(ctx context.Context, args ...string) ([]byte, error) {
				return readTestdata(t, "ps_one_llm.json"), nil
			},
		}
		client := New(fake)

		loaded, err := client.LoadedModels(context.Background())
		if err != nil {
			t.Fatalf("LoadedModels() error = %v", err)
		}
		if len(loaded) != 1 {
			t.Fatalf("len(loaded) = %d, want 1", len(loaded))
		}
		if loaded[0].Identifier != "qwen/qwen3.8-27b" {
			t.Errorf("Identifier = %q, want %q", loaded[0].Identifier, "qwen/qwen3.8-27b")
		}
		if loaded[0].ContextLength != 123648 {
			t.Errorf("ContextLength = %d, want 123648", loaded[0].ContextLength)
		}
		if loaded[0].TTL != nil {
			t.Errorf("TTL = %v, want nil", *loaded[0].TTL)
		}
		wantArgv := []string{"ps", "--json"}
		if !equalArgv(fake.Calls[0], wantArgv) {
			t.Errorf("argv = %v, want %v", fake.Calls[0], wantArgv)
		}
	})

	t.Run("none loaded returns empty slice, not an error", func(t *testing.T) {
		fake := &FakeRunner{
			RunFunc: func(ctx context.Context, args ...string) ([]byte, error) {
				return readTestdata(t, "ps_empty.json"), nil
			},
		}
		client := New(fake)

		loaded, err := client.LoadedModels(context.Background())
		if err != nil {
			t.Fatalf("LoadedModels() error = %v, want nil", err)
		}
		if len(loaded) != 0 {
			t.Errorf("len(loaded) = %d, want 0", len(loaded))
		}
	})
}

func TestClient_MalformedJSON_NamesTheCommand(t *testing.T) {
	tests := []struct {
		name       string
		call       func(c *Client) error
		wantInName string
	}{
		{
			name: "server status",
			call: func(c *Client) error {
				_, err := c.ServerStatus(context.Background())
				return err
			},
			wantInName: "server status",
		},
		{
			name: "list models",
			call: func(c *Client) error {
				_, err := c.ListModels(context.Background(), AllModels)
				return err
			},
			wantInName: "ls",
		},
		{
			name: "loaded models",
			call: func(c *Client) error {
				_, err := c.LoadedModels(context.Background())
				return err
			},
			wantInName: "ps",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &FakeRunner{
				RunFunc: func(ctx context.Context, args ...string) ([]byte, error) {
					return []byte("not json"), nil
				},
			}
			client := New(fake)

			err := tt.call(client)
			if err == nil {
				t.Fatal("error = nil, want a parse error")
			}
			if !strings.Contains(err.Error(), tt.wantInName) {
				t.Errorf("error = %q, want it to name the command %q", err.Error(), tt.wantInName)
			}
		})
	}
}

func TestClient_NonZeroExit_PropagatesRunError(t *testing.T) {
	wantErr := &RunError{Args: []string{"server", "status", "--json"}, ExitCode: 1, Output: []byte("Error: something went wrong")}
	fake := &FakeRunner{
		RunFunc: func(ctx context.Context, args ...string) ([]byte, error) {
			return nil, wantErr
		},
	}
	client := New(fake)

	_, err := client.ServerStatus(context.Background())
	var runErr *RunError
	if !errors.As(err, &runErr) {
		t.Fatalf("error = %v (%T), want *RunError", err, err)
	}
	if runErr.ExitCode != 1 {
		t.Errorf("ExitCode = %d, want 1", runErr.ExitCode)
	}
	if !strings.Contains(string(runErr.Output), "something went wrong") {
		t.Errorf("Output = %q, want the CLI's own message", runErr.Output)
	}
}

func equalArgv(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
