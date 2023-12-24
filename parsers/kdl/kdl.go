// Package kdl implements a koanf.Parser that parses KDL bytes as conf maps.
package kdl

import (
	"fmt"
	"reflect"

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
	var input interface{}
	if err := kdl.Unmarshal(b, &input); err != nil {
		return nil, err
	}
	if input == nil {
		return nil, nil
	}

	inputType := reflect.TypeOf(input)

	switch {
	case inputType == reflect.TypeOf(map[string]interface{}{}):
		return input.(map[string]interface{}), nil

	default:
		return nil, fmt.Errorf("unimplemented input type: %v", inputType)
	}

	return nil, fmt.Errorf("unimplemented")
}

// Marshal marshals the given config map to KDL bytes.
func (p *KDL) Marshal(o map[string]interface{}) ([]byte, error) {
	wrapper := map[string]interface{}{
		"root": o,
	}
	return kdl.Marshal(wrapper)
}
