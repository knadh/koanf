// Package kdl implements a koanf.Parser that parses KDL bytes as conf maps.
package kdl

import kdl "github.com/sblinch/kdl-go"

// KDL implements a KDL parser.
type KDL struct{}

// Parser returns a KDL Parser.
func Parser() *KDL {
	return &KDL{}
}

// Unmarshal parses the given KDL bytes.
//
// In case of KDL, nodes are parsed as-so to allow access to nested keys and use lists.
// alternative representations which directly use kdl nodes should be possible,
// using options in the struct to choose between each and also set any kdl-go options.
//
// - all documents become string-maps
//
// - nodes with the same name as previous nodes in a document will replace them in the map.
//
// - all nodes are parsed as string-maps, lists, strings, or numbers.
//
// - a single argument without properties or children becomes a value.
//
// - multiple arguments without properties or children become a list.
//
// - nodes with properties will be parsed as string-maps.
//
// - nodes with children and arguments will be parsed as string-maps.
//
// - nodes with only children will be parsed as lists.
//
// - nodes with children or properties and any arguments will replace the key "" for the node with a list of all arguments.
//
// - string-map key priority: children > arguments-in-the-""-key > keyprops
//
// - children nodes parsed as string-maps with the same name as any properties or previous children nodes will replace them in the map.
func (p *KDL) Unmarshal(b []byte) (map[string]interface{}, error) {
	var out map[string]interface{}
	if err := kdl.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Marshal marshals the given config map to KDL bytes.
func (p *KDL) Marshal(o map[string]interface{}) ([]byte, error) {
	return kdl.Marshal(o)
}
