// Package toml implements a koanf.Parser that parses TOML bytes as conf maps.
package toml

import (
	"github.com/pelletier/go-toml/v2"
)

// TOML implements a TOML parser.
type TOML struct{}

// Parser returns a TOML Parser.
func Parser() *TOML {
	return &TOML{}
}

// Unmarshal parses the given TOML bytes.
func (p *TOML) Unmarshal(b []byte) (map[string]any, error) {
	var outMap map[string]any

	if err := toml.Unmarshal(b, &outMap); err != nil {
		return nil, err
	}

	if len(outMap) == 0 {
		return nil, nil
	}

	return outMap, nil
}

// Marshal marshals the given config map to TOML bytes.
func (p *TOML) Marshal(o map[string]any) ([]byte, error) {
	if len(o) == 0 {
		return nil, nil
	}

	out, err := toml.Marshal(&o)
	if err != nil {
		return nil, err
	}

	return out, nil
}
