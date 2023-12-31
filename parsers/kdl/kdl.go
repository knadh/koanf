// Package kdl implements a koanf.Parser that parses KDL bytes as conf maps.
package kdl

import (
	kdl "github.com/sblinch/kdl-go"
)

// KDL implements a KDL parser.
type KDL struct{}

// Parser returns a KDL Parser.
func Parser() *KDL {
	return &KDL{}
}

// Unmarshal parses the given KDL bytes.
func (p *KDL) Unmarshal(b []byte) (map[string]interface{}, error) {
	var o map[string]interface{}
	err := kdl.Unmarshal(b, &o)
	return o, err
}

// Marshal marshals the given config map to KDL bytes.
func (p *KDL) Marshal(o map[string]interface{}) ([]byte, error) {
	return kdl.Marshal(o)
}
