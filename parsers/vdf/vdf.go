// Package VDF implements a koanf.Parser that parses VDF bytes as conf maps.
package vdf

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/andygrunwald/vdf"
)

// JSON implements a VDF parser.
type VDF struct{}

// Parser returns a VDF Parser.
func Parser() *VDF {
	return &VDF{}
}

// Unmarshal parses the given VDF bytes.
func (p *VDF) Unmarshal(b []byte) (map[string]interface{}, error) {
	reader := bytes.NewReader(b)
	vdfp := vdf.NewParser(reader)

	m, err := vdfp.Parse()
	if err != nil {
		return nil, err
	}
	return m, nil
}

// Marshal marshals the given config map to VDF bytes.
func (p *VDF) Marshal(o map[string]interface{}) ([]byte, error) {
	d, err := newDumper(o)

	if err != nil {
		return nil, err
	}
	return []byte(d), nil
}

// I'll remove this when
// https://github.com/andygrunwald/vdf/pull/55 gets merged.
func newDumper(vdfMap map[string]interface{}) (string, error) {
	var outBuilder strings.Builder
	err := recursiveMap(vdfMap, 0, &outBuilder)
	if err != nil {
		return "", err
	}
	return outBuilder.String(), nil
}

func recursiveMap(m map[string]interface{}, depth int, outBuilder *strings.Builder) error {
	for key, value := range m {
		switch valueType := value.(type) {
		case map[string]interface{}:
			outBuilder.WriteString(fmt.Sprintf("%s\"%s\"\n%s{\n", strings.Repeat("\t", depth), key, strings.Repeat("\t", depth)))
			err := recursiveMap(valueType, depth+1, outBuilder)
			if err != nil {
				return err
			}
			outBuilder.WriteString(fmt.Sprintf("%s}\n", strings.Repeat("\t", depth)))
		case string:
			outBuilder.WriteString(fmt.Sprintf("%s\"%s\"\t\t\"%s\"\n", strings.Repeat("\t", depth), key, value))
		default:
			return errors.New("unsupported value type")
		}
	}
	return nil
}
