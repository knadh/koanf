// Package hjson implements a koanf.Parser that parses HJSON bytes as conf maps.
// Very similar to json.
package hjson

import (
	"github.com/hjson/hjson-go/v4"
)

// HJSON implements a HJSON parser.
type HJSON struct{}

// Parser returns a HJSON parser.
func Parser() *HJSON {
	return &HJSON{}
}

// Unmarshal parses the given HJSON bytes.
func (p *HJSON) Unmarshal(b []byte) (map[string]any, error) {
	var out map[string]any
	if err := hjson.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Marshal marshals the given config map to HJSON bytes.
func (p *HJSON) Marshal(o map[string]any) ([]byte, error) {
	return hjson.Marshal(o)
}
