// Package hcl implements a koanf.Parser that parses Hashicorp
// HCL bytes as conf maps.
package hcl

import (
	"github.com/hashicorp/hcl"
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

// Parse parses the given HCL bytes.
func (p *HCL) Parse(b []byte) (map[string]interface{}, error) {
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
