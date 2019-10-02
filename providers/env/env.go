// Package env implements a koanf.Provider that reads environment
// variables as conf maps.
package env

import (
	"errors"
	"os"
	"strings"

	"github.com/knadh/koanf/maps"
)

// Env implements an environment variables provider.
type Env struct {
	prefix string
	delim  string
	cb     func(s string) string
}

// Provider returns an environment variables provider that returns
// a nested map[string]interface{} of environment variable where the
// nesting hierarchy of keys are defined by delim. For instance, the
// delim "." will convert the key `parent.child.key: 1`
// to `{parent: {child: {key: 1}}}`.
//
// If prefix is specified (case sensitive), only the env vars with
// the prefix are captured. cb is an optional callback that takes
// a string and returns a string (the env variable name) in case
// transformatios have to be applied, for instance, to lowercase
// everything, strip prefixes and replace _ with . etc.
func Provider(prefix, delim string, cb func(s string) string) *Env {
	return &Env{
		prefix: prefix,
		delim:  delim,
		cb:     cb,
	}
}

// ReadBytes is not supported by the env provider.
func (e *Env) ReadBytes() ([]byte, error) {
	return nil, errors.New("env provider does not support this method")
}

// Read reads all available environment variables into a key:value map
// and returns it.
func (e *Env) Read() (map[string]interface{}, error) {
	// Collect the environment variable keys.
	var keys []string
	for _, k := range os.Environ() {
		if e.prefix != "" {
			if strings.HasPrefix(k, e.prefix) {
				keys = append(keys, k)
			}
		} else {
			keys = append(keys, k)
		}
	}

	mp := make(map[string]interface{})
	for _, k := range keys {
		parts := strings.SplitN(k, "=", 2)

		// If there's a string transformation callback,
		// run it through every string.
		if e.cb != nil {
			parts[0] = e.cb(parts[0])
		}
		mp[parts[0]] = parts[1]
	}

	return maps.Unflatten(mp, e.delim), nil
}

// Watch is not supported.
func (e *Env) Watch(cb func(event interface{}, err error)) error {
	return errors.New("env provider does not support this method")
}
