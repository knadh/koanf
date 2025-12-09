// Package huml implements a koanf.Parser that parses HUML bytes as conf maps.
package huml

import (
	"github.com/huml-lang/go-huml"
)

// HUML implements a HUML parser.
type HUML struct{}

// Parser returns a HUML Parser.
func Parser() *HUML {
	return &HUML{}
}

// Unmarshal parses the given HUML bytes.
func (p *HUML) Unmarshal(b []byte) (map[string]any, error) {
	var result map[string]any
	err := huml.Unmarshal(b, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Marshal marshals the given config map to HUML bytes.
func (p *HUML) Marshal(o map[string]any) ([]byte, error) {
	return huml.Marshal(o)
}
