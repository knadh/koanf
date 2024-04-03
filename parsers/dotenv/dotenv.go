// Package dotenv implements a koanf.Parser that parses DOTENV bytes as conf maps.
package dotenv

import (
	"fmt"
	"strings"

	"github.com/joho/godotenv"
	"github.com/knadh/koanf/maps"
)

// DotEnv implements a DOTENV parser.
type DotEnv struct {
	delim  string
	prefix string

	cb        func(key string, value string) (string, interface{})
	reverseCB map[string]string
}

// Parser returns a DOTENV Parser.
func Parser() *DotEnv {
	return &DotEnv{}
}

// ParserEnv allows to make the DOTENV Parser behave like the env.Provider.
func ParserEnv(prefix, delim string, cb func(s string) string) *DotEnv {
	return &DotEnv{
		delim:  delim,
		prefix: prefix,
		cb: func(key, value string) (string, interface{}) {
			return cb(key), value
		},
		reverseCB: make(map[string]string),
	}
}

// ParserEnvWithValue allows to make the DOTENV Parser behave like the env.ProviderWithValue.
func ParserEnvWithValue(prefix, delim string, cb func(key string, value string) (string, interface{})) *DotEnv {
	return &DotEnv{
		delim:     delim,
		prefix:    prefix,
		cb:        cb,
		reverseCB: make(map[string]string),
	}
}

// Unmarshal parses the given DOTENV bytes.
func (p *DotEnv) Unmarshal(b []byte) (map[string]interface{}, error) {
	// Unmarshal DOTENV from []byte
	r, err := godotenv.Unmarshal(string(b))
	if err != nil {
		return nil, err
	}

	// Convert a map[string]string to a map[string]interface{}
	mp := make(map[string]interface{})
	for sourceKey, v := range r {
		if !strings.HasPrefix(sourceKey, p.prefix) {
			continue
		}

		if p.cb != nil {
			targetKey, value := p.cb(sourceKey, v)
			p.reverseCB[targetKey] = sourceKey
			mp[targetKey] = value
		} else {
			mp[sourceKey] = v
		}

	}

	if p.delim != "" {
		mp = maps.Unflatten(mp, p.delim)
	}
	return mp, nil
}

// Marshal marshals the given config map to DOTENV bytes.
func (p *DotEnv) Marshal(o map[string]interface{}) ([]byte, error) {
	if p.delim != "" {
		o, _ = maps.Flatten(o, nil, p.delim)
	}

	// Convert a map[string]interface{} to a map[string]string
	mp := make(map[string]string)
	for targetKey, v := range o {
		if sourceKey, found := p.reverseCB[targetKey]; found {
			targetKey = sourceKey
		}

		mp[targetKey] = fmt.Sprint(v)
	}

	// Unmarshal to string
	out, err := godotenv.Marshal(mp)
	if err != nil {
		return nil, err
	}

	// Convert to []byte and return
	return []byte(out), nil
}
