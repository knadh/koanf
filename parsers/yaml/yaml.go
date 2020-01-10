// Package yaml implements a koanf.Parser that parses YAML bytes as conf maps.
package yaml

import (
	"gopkg.in/yaml.v2"
)

// YAML implements a YAML parser.
type YAML struct{}

// Parser returns a YAML Parser.
func Parser() *YAML {
	return &YAML{}
}

// Parse parses the given YAML bytes.
func (p *YAML) Parse(b []byte) (map[string]interface{}, error) {
	var out map[string]interface{}
	if err := yaml.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	return out, nil
}
