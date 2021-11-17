// Package nestedtext implements a koanf Parser that parses NestedText bytes as conf maps.
package nestedtext

import (
	"bytes"
	"errors"

	"github.com/npillmayer/nestext"
	"github.com/npillmayer/nestext/ntenc"
)

// NT implements a NestedText parser.
type NT struct{}

// Parser returns a NestedText Parser.
func Parser() *NT {
	return &NT{}
}

// Unmarshal parses the given NestedText bytes.
//
// If the NT content does not reflect a dict (NT allows for top-level lists or strings as well),
// the content will be wrapped into a dict with a single key named "nestedtext".
func (p *NT) Unmarshal(b []byte) (map[string]interface{}, error) {
	r := bytes.NewReader(b)
	result, err := nestext.Parse(r, nestext.TopLevel("dict"))
	if err != nil {
		return nil, err
	}
	// Given option-parameter TopLevel, nestext.Parse is guaranteed to wrap the return value
	// in an appropriate type (dict = map[string]interface{} in this case), if necessary.
	// However, guard against type conversion failure anyway.
	rmap, ok := result.(map[string]interface{})
	if !ok {
		return nil, errors.New("NestedText configuration expected to be a dict at top-level")
	}
	return rmap, nil
}

// Marshal marshals the given config map to NestedText bytes.
func (p *NT) Marshal(m map[string]interface{}) ([]byte, error) {
	var buf bytes.Buffer
	_, err := ntenc.Encode(m, &buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
