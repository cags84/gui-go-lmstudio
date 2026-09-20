package lms

import "time"

// modelWire mirrors exactly the JSON object `lms ls --json` emits for one
// model, including the fields the CLI omits entirely for embedding entries.
// Client converts this into the domain-level Model before returning it, so the
// rest of the app never sees the raw wire shape.
type modelWire struct {
	Type                   ModelType    `json:"type"`
	ModelKey               string       `json:"modelKey"`
	Format                 string       `json:"format"`
	DisplayName            string       `json:"displayName"`
	Publisher              string       `json:"publisher"`
	Path                   string       `json:"path"`
	SizeBytes              int64        `json:"sizeBytes"`
	IndexedModelIdentifier string       `json:"indexedModelIdentifier"`
	DeviceIdentifier       *string      `json:"deviceIdentifier"`
	Architecture           string       `json:"architecture"`
	Quantization           Quantization `json:"quantization"`
	MaxContextLength       int64        `json:"maxContextLength"`
	ParamsString           *string      `json:"paramsString,omitempty"`
	Variants               []string     `json:"variants,omitempty"`
	SelectedVariant        *string      `json:"selectedVariant,omitempty"`
	Vision                 *bool        `json:"vision,omitempty"`
	TrainedForToolUse      *bool        `json:"trainedForToolUse,omitempty"`
}

func (w modelWire) toModel() Model {
	return Model{
		Type:                   w.Type,
		ModelKey:               w.ModelKey,
		Format:                 w.Format,
		DisplayName:            w.DisplayName,
		Publisher:              w.Publisher,
		Path:                   w.Path,
		SizeBytes:              w.SizeBytes,
		IndexedModelIdentifier: w.IndexedModelIdentifier,
		DeviceIdentifier:       w.DeviceIdentifier,
		Architecture:           w.Architecture,
		Quantization:           w.Quantization,
		MaxContextLength:       w.MaxContextLength,
		ParamsString:           w.ParamsString,
		Variants:               w.Variants,
		SelectedVariant:        w.SelectedVariant,
		Vision:                 w.Vision,
		TrainedForToolUse:      w.TrainedForToolUse,
	}
}

// loadedModelWire mirrors `lms ps --json`: every modelWire field plus the
// runtime fields that only exist while a model is loaded. lastUsedTime and
// ttlMs stay as raw epoch milliseconds here; toLoadedModel converts them to
// time.Time/time.Duration at this boundary so the rest of the app never
// handles raw epoch milliseconds.
type loadedModelWire struct {
	modelWire
	Identifier    string `json:"identifier"`
	TTLMs         *int64 `json:"ttlMs"`
	LastUsedTime  int64  `json:"lastUsedTime"`
	ContextLength int64  `json:"contextLength"`
	Status        string `json:"status"`
	Queued        int    `json:"queued"`
	Parallel      int    `json:"parallel"`
}

func (w loadedModelWire) toLoadedModel() LoadedModel {
	var ttl *time.Duration
	if w.TTLMs != nil {
		d := time.Duration(*w.TTLMs) * time.Millisecond
		ttl = &d
	}
	return LoadedModel{
		Model:         w.modelWire.toModel(),
		Identifier:    w.Identifier,
		TTL:           ttl,
		LastUsedTime:  time.UnixMilli(w.LastUsedTime),
		ContextLength: w.ContextLength,
		Status:        w.Status,
		Queued:        w.Queued,
		Parallel:      w.Parallel,
	}
}
