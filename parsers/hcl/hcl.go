// Package hcl implements a koanf.Parser that parses Hashicorp
// HCL bytes as conf maps.
package hcl

import (
	"bytes"
	"encoding/json"

	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/printer"
)

// HCL implements a Hashicorp HCL parser.
type HCL struct{ flattenSlices bool }

// Parser returns an HCL Parser.
// flattenSlices flattens HCL structures where maps turn into
// lists of maps. Read more here: https://github.com/hashicorp/hcl/issues/162
// It's recommended to turn this setting on.
func Parser(flattenSlices bool) *HCL {
	return &HCL{flattenSlices: flattenSlices}
}

// Unmarshal parses the given HCL bytes.
func (p *HCL) Unmarshal(b []byte) (map[string]interface{}, error) {
	o, err := hcl.Parse(string(b))
	if err != nil {
		return nil, err
	}

	var out map[string]interface{}
	if err := hcl.DecodeObject(&out, o); err != nil {
		return nil, err
	}
	if p.flattenSlices {
		flattenHCL(out)
	}
	return out, nil
}

// Marshal marshals the given config map to HCL bytes.
func (p *HCL) Marshal(o map[string]interface{}) ([]byte, error) {
	j, err := json.Marshal(o)
	if err != nil {
		return nil, err
	}

	tree, err := hcl.Parse(string(j))
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	if err = printer.Fprint(buf, tree); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// flattenHCL flattens an unmarshalled HCL structure where maps
// turn into slices -- https://github.com/hashicorp/hcl/issues/162.
func flattenHCL(mp map[string]interface{}) {
	for k, val := range mp {
		if v, ok := val.([]map[string]interface{}); ok {
			if len(v) == 1 {
				mp[k] = v[0]
			}
		}
	}
	for _, val := range mp {
		if v, ok := val.(map[string]interface{}); ok {
			flattenHCL(v)
		}
	}
}
