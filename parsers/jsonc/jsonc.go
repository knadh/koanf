// Package jsonc implements a koanf.Parser that parses JSONC bytes
// (JSON with comments and trailing commas) as conf maps.
package jsonc

import (
	"bytes"
	"encoding/json"

	"github.com/tailscale/hujson"
)

// JSONC implements a JSONC parser.
type JSONC struct{}

// Parser returns a JSONC Parser.
func Parser() *JSONC {
	return &JSONC{}
}

// Unmarshal parses the given JSONC bytes. Comments and trailing commas are
// stripped before the result is parsed as standard JSON.
func (p *JSONC) Unmarshal(b []byte) (map[string]any, error) {
	// hujson.Standardize rewrites comments in place, so work on a copy to
	// leave the caller's bytes untouched.
	std, err := hujson.Standardize(bytes.Clone(b))
	if err != nil {
		return nil, err
	}

	var out map[string]any
	if err := json.Unmarshal(std, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Marshal marshals the given config map to JSON bytes. Comments are not preserved.
func (p *JSONC) Marshal(o map[string]any) ([]byte, error) {
	return json.Marshal(o)
}
