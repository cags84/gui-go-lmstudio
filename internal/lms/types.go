package lms

import "time"

// ModelType is the "type" field of an lms model entry: the values LM Studio's
// CLI emits for on-disk and loaded models.
type ModelType string

const (
	ModelTypeLLM       ModelType = "llm"
	ModelTypeEmbedding ModelType = "embedding"
)

// ServerStatus mirrors `lms server status --json` exactly; no conversion is
// needed so it doubles as its own wire format.
type ServerStatus struct {
	Running bool `json:"running"`
	Port    int  `json:"port"`
}

// Quantization is the {"name","bits"} object nested in every model entry.
type Quantization struct {
	Name string `json:"name"`
	Bits int    `json:"bits"`
}

// Model is one entry from `lms ls --json`, translated into Go-native types.
//
// Several fields are pointers because the CLI omits them entirely on embedding
// entries (ParamsString, SelectedVariant, Vision, TrainedForToolUse) or emits
// them as an explicit JSON null (DeviceIdentifier). A plain zero value (empty
// string, false) would be indistinguishable from "the CLI told us this is
// empty/false", so a caller that needs the difference (for example a UI badge
// for "no tool-use support" vs "unknown, this is an embedding model") needs the
// pointer. Variants is left as a plain nil-or-populated slice: JSON arrays have
// no separate "false" state to confuse with "absent".
type Model struct {
	Type                   ModelType
	ModelKey               string
	Format                 string
	DisplayName            string
	Publisher              string
	Path                   string
	SizeBytes              int64
	IndexedModelIdentifier string
	// DeviceIdentifier is non-nil only for a model served by a remote LM Link
	// device. Prefer IsRemote over re-deriving this from a ":" prefix on Path:
	// that prefix is present on Path for `ls` entries but was observed absent
	// on the same remote model's Path in a `ps` entry, so the prefix is not a
	// reliable signal by itself.
	DeviceIdentifier  *string
	Architecture      string
	Quantization      Quantization
	MaxContextLength  int64
	ParamsString      *string
	Variants          []string
	SelectedVariant   *string
	Vision            *bool
	TrainedForToolUse *bool
}

// IsRemote reports whether this model is served by a remote LM Link device
// rather than the local lms daemon. Remote is the common case in real usage,
// not an edge case: a captured `lms ls --json` on a machine with one LM Link
// peer returned 7 remote entries out of 11.
func (m Model) IsRemote() bool {
	return m.DeviceIdentifier != nil
}

// LoadedModel is one entry from `lms ps --json`: everything from Model plus the
// runtime fields that exist only while the model is actually loaded.
type LoadedModel struct {
	Model
	Identifier string
	// TTL is nil when the model has no idle-unload timeout configured
	// (ttlMs was JSON null).
	TTL           *time.Duration
	LastUsedTime  time.Time
	ContextLength int64
	Status        string
	Queued        int
	Parallel      int
}
