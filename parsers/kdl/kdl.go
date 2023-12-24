// Package kdl implements a koanf.Parser that parses KDL bytes as conf maps.
package kdl

import (
	"bytes"
	"fmt"
	"reflect"

	"log"

	kdl "github.com/sblinch/kdl-go"
	k "github.com/sblinch/kdl-go/document"
)

type ParseStrategy int

const (
	Primitive ParseStrategy = iota
	NodeMap
	StringMap
)

type MergeStrategy int

const (
	Overwrite MergeStrategy = iota
	// Overlay
	// Append
	Skip
	Strict
)

type MergeOptions struct {
	Arguments               MergeStrategy
	Properties              MergeStrategy
	Nodes                   MergeStrategy
	Strict                  bool
	ChildrenPropertyKey     string
	ArgumentsPropertyKey    string
	ArgumentsAlwaysProperty bool
	ArgumentsAlwaysArray    bool
	ArgumentsAlwaysIncluded bool
}

type KDLOptions struct {
	ParseStrategy ParseStrategy
	MergeOptions  MergeOptions
}

// KDL implements a KDL parser.
type KDL struct{ Options KDLOptions }

func (p *KDL) SetOptions(o KDLOptions) {
	p.Options = o
}

func (p *KDL) GetOptions() KDLOptions {
	return p.Options
}

func DefaultMergeOptions() MergeOptions {
	return MergeOptions{
		Arguments:  Overwrite,
		Properties: Overwrite,
		Nodes:      Overwrite,
		Strict:     false,
		// consider: arguments into args key, arguments arg=nil arguments arg=arg
		// ChildrenPropertyKey:     "", // todo: if childrenpropertykey is not null, then put all children in list under that key instead of parsing as string-map
		ArgumentsPropertyKey:    "",
		ArgumentsAlwaysProperty: false,
		ArgumentsAlwaysArray:    false,
		ArgumentsAlwaysIncluded: false,
	}
}

func DefaultPrimitiveOptions() KDLOptions {
	return KDLOptions{
		ParseStrategy: Primitive,
		MergeOptions:  DefaultMergeOptions(),
	}
}

func DefaultNodeMapOptions() KDLOptions {
	return KDLOptions{
		ParseStrategy: NodeMap,
		MergeOptions:  DefaultMergeOptions(),
	}
}

func DefaultStringMapOptions() KDLOptions {
	return KDLOptions{
		ParseStrategy: StringMap,
		MergeOptions:  DefaultMergeOptions(),
	}
}

func DefaultOptions() KDLOptions {
	return DefaultStringMapOptions()
}

// Parser returns a KDL Parser.
func Parser() *KDL {
	return &KDL{DefaultOptions()}
}

func NodeMapParser() *KDL {
	return &KDL{DefaultNodeMapOptions()}
}

func StringMapParser() *KDL {
	return &KDL{DefaultStringMapOptions()}
}

func PrimitiveParser() *KDL {
	return &KDL{DefaultPrimitiveOptions()}
}

// Unmarshal parses the given KDL bytes.
//
// In case of KDL, nodes are parsed as-so to allow access to nested keys and use lists.
// alternative representations which directly use kdl nodes should be possible,
// using Options in the struct to choose between each and also set any kdl-go Options.
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
func (p *KDL) Merge(src, dest map[string]interface{}) error {
	if dest == nil {
		dest = make(map[string]interface{})
	}
	if src == nil {
		return nil
	}
	return nil
}

func (p *KDL) MergeArguments(src []*k.Value, dest map[string]interface{}) (map[string]interface{}, error) {
	if dest == nil {
		dest = make(map[string]interface{})
	}
	if src == nil {
		return dest, nil
	}
	if len(src) > 0 || p.Options.MergeOptions.ArgumentsAlwaysIncluded {
		_, exists := dest[p.Options.MergeOptions.ArgumentsPropertyKey]
		switch {
		case exists && (p.Options.MergeOptions.Arguments == Strict || p.Options.MergeOptions.Strict):
			return dest, fmt.Errorf("arguments already exist in destination and strict mode is enabled")
		case exists && p.Options.MergeOptions.Arguments == Skip:
			return dest, nil
		}

		switch {
		case p.Options.MergeOptions.Arguments == Overwrite || !exists || len(src) == 0 || p.Options.MergeOptions.ArgumentsAlwaysProperty || p.Options.MergeOptions.ArgumentsAlwaysArray:
			switch {
			case len(src) > 1 || p.Options.MergeOptions.ArgumentsAlwaysProperty:
				if p.Options.MergeOptions.ArgumentsAlwaysArray || len(src) > 1 {
					dest[p.Options.MergeOptions.ArgumentsPropertyKey] = src
				} else {
					dest[p.Options.MergeOptions.ArgumentsPropertyKey] = src[0]
				}
			case len(src) == 1:
				if p.Options.MergeOptions.ArgumentsAlwaysArray {
					dest[p.Options.MergeOptions.ArgumentsPropertyKey] = src
				} else {
					dest[p.Options.MergeOptions.ArgumentsPropertyKey] = src[0]
				}
			default:
				dest[p.Options.MergeOptions.ArgumentsPropertyKey] = src
			}
		default:
			return dest, fmt.Errorf("unimplemented merge strategy: %v", p.Options.MergeOptions.Arguments)
		}
	}
	return dest, nil
}

func (p *KDL) MergeProperties(src *k.Properties, dest map[string]interface{}) (map[string]interface{}, error) {
	if dest == nil {
		dest = make(map[string]interface{})
	}
	if src == nil {
		return dest, nil
	}

	switch {
	case p.Options.MergeOptions.Properties == Strict || p.Options.MergeOptions.Strict:
		for k, v := range *src {
			if _, exists := dest[k]; exists {
				return dest, fmt.Errorf("property %s already exists in destination and strict mode is enabled", k)
			}
			dest[k] = v
		}
	case p.Options.MergeOptions.Properties == Overwrite:
		for k, v := range *src {
			dest[k] = v
		}
	case p.Options.MergeOptions.Properties == Skip:
		for k, v := range *src {
			if _, exists := dest[k]; exists {
				continue
			}
			dest[k] = v
		}
	default:
		return dest, fmt.Errorf("unimplemented merge strategy: %v", p.Options.MergeOptions.Properties)
	}

	return dest, nil
}

func (p *KDL) MergeNode(src *k.Node, dest map[string]interface{}) (map[string]interface{}, error) {
	if dest == nil {
		dest = make(map[string]interface{})
	}
	if src == nil {
		return dest, nil
	}

	_, destExists := dest[src.Name.ValueString()] // value will be needed for overlay and append
	// strategy := Strategy{p.Options.ParseStrategy, p.Options.MergeOptions.Nodes, reflect.TypeOf(src), reflect.TypeOf(destValue), !destExists, p.Options.MergeOptions.Strict}

	switch {
	case !destExists && p.Options.MergeOptions.Strict:
		return dest, fmt.Errorf("node %s already exists in destination", src.Name)
	case !destExists && p.Options.MergeOptions.Nodes == Skip:
		return dest, nil
	case p.Options.ParseStrategy == NodeMap && (p.Options.MergeOptions.Nodes == Overwrite || !destExists) && len(src.Children) == 0:
		dest[src.Name.ValueString()] = src
	case p.Options.ParseStrategy == NodeMap && (p.Options.MergeOptions.Nodes == Overwrite || !destExists):
		var result map[string]interface{}
		var err error
		switch {
		case len(src.Properties) == 0 && len(src.Arguments) == 0:
			result, err = p.MergeProperties(&src.Properties, nil)
			if err != nil {
				return dest, err
			}
		}
		switch {
		case
			len(src.Properties) == 0 &&
				len(src.Children) == 0 &&
				!p.Options.MergeOptions.ArgumentsAlwaysProperty &&
				(len(src.Arguments) > 0 || p.Options.MergeOptions.ArgumentsAlwaysIncluded):
			switch {
			case p.Options.MergeOptions.ArgumentsAlwaysArray || len(src.Arguments) > 1:
				dest[src.Name.ValueString()] = src.Arguments
			case len(src.Arguments) == 1:
				dest[src.Name.ValueString()] = src.Arguments[0]
			default:
				dest[src.Name.ValueString()] = src.Arguments
			}
		default:
			result, err = p.MergeArguments(src.Arguments, result)
			if err != nil {
				return dest, err
			}
		}
		switch {
		case len(src.Children) > 0:
			result, err = p.MergeNodes(src.Children, result)
			if err != nil {
				return dest, err
			}
			dest[src.Name.ValueString()] = result
		}
	default:
		return dest, fmt.Errorf("unimplemented parse strategy: %v merge strategy: %v", p.Options.ParseStrategy, p.Options.MergeOptions.Nodes)
	}
	return dest, nil
}

func (p *KDL) MergeNodes(src []*k.Node, dest map[string]interface{}) (map[string]interface{}, error) {
	if src == nil {
		return dest, nil
	}
	if dest == nil {
		dest = make(map[string]interface{})
	}
	for _, node := range src {
		if dest, err := p.MergeNode(node, dest); err != nil {
			return dest, err
		}
	}
	return dest, nil
}

func (p *KDL) Unmarshal(b []byte) (map[string]interface{}, error) {
	var input interface{}
	var err error
	if p.Options.ParseStrategy == StringMap {
		if err := kdl.Unmarshal(b, &input); err != nil {
			return nil, err
		}
	} else {
		byteReader := bytes.NewReader(b)
		if input, err = kdl.Parse(byteReader); err != nil {
			return nil, err
		}
	}
	if input == nil {
		return nil, nil
	}

	inputType := reflect.TypeOf(input)

	switch {
	case inputType == reflect.TypeOf(k.Document{}):
		var dest map[string]interface{}

		dest, err := p.MergeNodes(input.(*k.Document).Nodes, dest)
		log.Println("dest: ", dest)

		if err != nil {
			return dest, err
		}
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
