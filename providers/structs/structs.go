// Package structs implements a koanf.Provider that takes a struct and tag
// and returns a nested config map (using fatih/structs) to provide it to koanf.
package structs

import (
	"errors"

	"github.com/fatih/structs"
)

// Structs implements a structs provider.
type Structs struct {
	s   interface{}
	tag string
}

// Provider returns a provider that takes a  takes a struct and a struct tag
// and uses structs to parse and provide it to koanf.
func Provider(s interface{}, tag string) *Structs {
	return &Structs{s: s, tag: tag}
}

// ReadBytes is not supported by the structs provider.
func (s *Structs) ReadBytes() ([]byte, error) {
	return nil, errors.New("structs provider does not support this method")
}

// Read reads the struct and returns a nested config map.
func (s *Structs) Read() (map[string]interface{}, error) {
	ns := structs.New(s.s)
	ns.TagName = s.tag

	return ns.Map(), nil
}

// Watch is not supported by the structs provider.
func (s *Structs) Watch(cb func(event interface{}, err error)) error {
	return errors.New("structs provider does not support this method")
}
