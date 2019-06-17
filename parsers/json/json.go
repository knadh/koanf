// Package json implements a koanf.Parser that parses JSON bytes as conf maps.
package json

import (
	"encoding/json"
)

// JSON implements a JSON parser.
type JSON struct{}

// Parser returns a JSON Parser.
func Parser() *JSON {
	return &JSON{}
}

// Parse parses the given JSON bytes.
func (p *JSON) Parse(b []byte) (map[string]interface{}, error) {
	var out map[string]interface{}
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	return out, nil
}
