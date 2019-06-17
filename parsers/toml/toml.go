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

// Parse parses the given TOML bytes.
func (p *TOML) Parse(b []byte) (map[string]interface{}, error) {
	r, err := toml.LoadReader(bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}
	return r.ToMap(), err
}
