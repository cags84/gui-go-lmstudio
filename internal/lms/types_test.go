package lms

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func readTestdata(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("reading testdata/%s: %v", name, err)
	}
	return data
}

func TestServerStatus_ParsesFixtures(t *testing.T) {
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
			var got ServerStatus
			if err := json.Unmarshal(readTestdata(t, tt.file), &got); err != nil {
				t.Fatalf("Unmarshal: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestModelWire_ToModel_LsAll(t *testing.T) {
	var wire []modelWire
	if err := json.Unmarshal(readTestdata(t, "ls_all.json"), &wire); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if len(wire) != 11 {
		t.Fatalf("len(wire) = %d, want 11", len(wire))
	}

	var llmCount, embeddingCount, remoteCount, localCount int
	for _, w := range wire {
		m := w.toModel()
		switch m.Type {
		case ModelTypeLLM:
			llmCount++
			if m.ParamsString == nil {
				t.Errorf("llm model %s: ParamsString = nil, want set", m.ModelKey)
			}
			if m.Vision == nil {
				t.Errorf("llm model %s: Vision = nil, want set", m.ModelKey)
			}
		case ModelTypeEmbedding:
			embeddingCount++
			if m.ParamsString != nil {
				t.Errorf("embedding model %s: ParamsString = %v, want nil (field absent in JSON)", m.ModelKey, *m.ParamsString)
			}
			if m.Vision != nil {
				t.Errorf("embedding model %s: Vision = %v, want nil (field absent in JSON)", m.ModelKey, *m.Vision)
			}
			if m.TrainedForToolUse != nil {
				t.Errorf("embedding model %s: TrainedForToolUse = %v, want nil", m.ModelKey, *m.TrainedForToolUse)
			}
			if m.Variants != nil {
				t.Errorf("embedding model %s: Variants = %v, want nil", m.ModelKey, m.Variants)
			}
			if m.SelectedVariant != nil {
				t.Errorf("embedding model %s: SelectedVariant = %v, want nil", m.ModelKey, *m.SelectedVariant)
			}
		default:
			t.Errorf("unexpected model type %q", m.Type)
		}
		if m.IsRemote() {
			remoteCount++
			if m.DeviceIdentifier == nil {
				t.Errorf("model %s: IsRemote() true but DeviceIdentifier is nil", m.ModelKey)
			}
		} else {
			localCount++
			if m.DeviceIdentifier != nil {
				t.Errorf("model %s: IsRemote() false but DeviceIdentifier = %v", m.ModelKey, *m.DeviceIdentifier)
			}
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
	if localCount != 4 {
		t.Errorf("localCount = %d, want 4", localCount)
	}
}

func TestModelWire_ToModel_LocalModelIsNotRemote(t *testing.T) {
	var wire []modelWire
	if err := json.Unmarshal(readTestdata(t, "ls_embedding.json"), &wire); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if len(wire) != 2 {
		t.Fatalf("len(wire) = %d, want 2", len(wire))
	}

	local := wire[0].toModel()
	if local.IsRemote() {
		t.Errorf("first ls_embedding.json entry: IsRemote() = true, want false")
	}
	if local.Path != "nomic-ai/nomic-embed-text-v1.5-GGUF/nomic-embed-text-v1.5.Q4_K_M.gguf" {
		t.Errorf("local Path = %q, unexpected", local.Path)
	}

	remote := wire[1].toModel()
	if !remote.IsRemote() {
		t.Errorf("second ls_embedding.json entry: IsRemote() = false, want true")
	}
	if remote.Path != "eb427d97b4c0dfc803433710bef6c718:nomic-ai/nomic-embed-text-v1.5-GGUF/nomic-embed-text-v1.5.Q4_K_M.gguf" {
		t.Errorf("remote Path = %q, unexpected prefix", remote.Path)
	}
}

func TestLoadedModelWire_ToLoadedModel(t *testing.T) {
	t.Run("one llm loaded", func(t *testing.T) {
		var wire []loadedModelWire
		if err := json.Unmarshal(readTestdata(t, "ps_one_llm.json"), &wire); err != nil {
			t.Fatalf("Unmarshal: %v", err)
		}
		if len(wire) != 1 {
			t.Fatalf("len(wire) = %d, want 1", len(wire))
		}
		got := wire[0].toLoadedModel()

		if got.Identifier != "qwen/qwen3.8-27b" {
			t.Errorf("Identifier = %q, want %q", got.Identifier, "qwen/qwen3.8-27b")
		}
		if got.ContextLength != 123648 {
			t.Errorf("ContextLength = %d, want 123648", got.ContextLength)
		}
		if got.MaxContextLength != 262144 {
			t.Errorf("MaxContextLength = %d, want 262144", got.MaxContextLength)
		}
		if got.ContextLength == got.MaxContextLength {
			t.Error("ContextLength and MaxContextLength must be distinguishable, both came out equal")
		}
		wantTime := time.UnixMilli(1789915288259)
		if !got.LastUsedTime.Equal(wantTime) {
			t.Errorf("LastUsedTime = %v, want %v", got.LastUsedTime, wantTime)
		}
		if got.TTL != nil {
			t.Errorf("TTL = %v, want nil (ttlMs is null in the fixture)", *got.TTL)
		}
		if got.Status != "idle" {
			t.Errorf("Status = %q, want %q", got.Status, "idle")
		}
		if got.Queued != 0 {
			t.Errorf("Queued = %d, want 0", got.Queued)
		}
		if got.Parallel != 4 {
			t.Errorf("Parallel = %d, want 4", got.Parallel)
		}
		if !got.IsRemote() {
			t.Error("IsRemote() = false, want true (deviceIdentifier is set)")
		}
	})

	t.Run("none loaded", func(t *testing.T) {
		var wire []loadedModelWire
		if err := json.Unmarshal(readTestdata(t, "ps_empty.json"), &wire); err != nil {
			t.Fatalf("Unmarshal: %v", err)
		}
		if wire == nil {
			t.Fatal("Unmarshal of an empty JSON array produced a nil slice")
		}
		if len(wire) != 0 {
			t.Fatalf("len(wire) = %d, want 0", len(wire))
		}
	})
}

func TestLoadedModelWire_TTLConversion(t *testing.T) {
	raw := `{"modelKey":"m","format":"gguf","displayName":"M","publisher":"p","path":"p/m","sizeBytes":1,"indexedModelIdentifier":"p/m","deviceIdentifier":null,"type":"llm","architecture":"a","quantization":{"name":"Q4","bits":4},"maxContextLength":100,"identifier":"p/m","ttlMs":60000,"lastUsedTime":0,"contextLength":50,"status":"idle","queued":0,"parallel":1}`

	var w loadedModelWire
	if err := json.Unmarshal([]byte(raw), &w); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	got := w.toLoadedModel()
	if got.TTL == nil {
		t.Fatal("TTL = nil, want 60s")
	}
	if *got.TTL != 60*time.Second {
		t.Errorf("TTL = %v, want 60s", *got.TTL)
	}
}
