// Package toml implements a koanf.Parser that parses TOML bytes as conf maps.
package toml

import (
	"bytes"

	"github.com/pelletier/go-toml"
)

// TOML implements a TOML parser.
type TOML struct{}

// Parser returns a TOML Parser.
func Parser() *TOML {
	return &TOML{}
}

// Unmarshal parses the given TOML bytes.
func (p *TOML) Unmarshal(b []byte) (map[string]interface{}, error) {
	r, err := toml.LoadReader(bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}
	return r.ToMap(), err
}

// Marshal marshals the given config map to TOML bytes.
func (p *TOML) Marshal(o map[string]interface{}) ([]byte, error) {
	out, err := toml.TreeFromMap(o)
	if err != nil {
		return nil, err
	}
	return out.Marshal()
}
