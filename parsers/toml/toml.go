// Package toml implements a koanf.Parser that parses TOML bytes as conf maps.
package toml

import (
	"fmt"
	"github.com/pelletier/go-toml/v2"
)

// TOML implements a TOML parser.
type TOML struct{}

// Parser returns a TOML Parser.
func Parser() *TOML {
	return &TOML{}
}

// Unmarshal parses the given TOML bytes.
func (p *TOML) Unmarshal(b []byte) (map[string]interface{}, error) {
	var test map[string]interface{}

	err := toml.Unmarshal(b, &test)
	if err != nil {
		return nil, err
	}

	return test, nil
}

// Marshal marshals the given config map to TOML bytes.
func (p *TOML) Marshal(o map[string]interface{}) ([]byte, error) {
	out, err := toml.Marshal(&o)
	if err != nil {
		return nil, err
	}

	fmt.Println(string(out))
	return out, nil
}
