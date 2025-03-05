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
	prefix    string
	delim     string
	transform func(key string, value string) (string, any)
	environ   func() []string
}

// Opt represents optional configuration passed to the provider.
type Opt struct {
	// Prefix limits the provider to only capture env vars that begin
	// with the prefix.
	Prefix string

	// TransformFunc is an optional callback that takes in the env
	// var name and value and returns a transformed version. Common
	// transformations for the name include stripping prefixes,
	// replacing _ with ., and so on. The value can be transformed
	// to other types, like a slice of strings instead of just a string.
	TransformFunc func(k, v string) (string, any)

	// EnvironFunc is the function that feeds environment variables
	// to the provider.
	EnvironFunc func() []string
}

// Provider returns an environment variables provider that returns
// a nested map[string]interface{} of environment variable where the
// nesting hierarchy of keys is defined by delim. For instance, the
// delim "." will convert the key `parent.child.key: 1`
// to `{parent: {child: {key: 1}}}`.
//
// If prefix is specified (case-sensitive), only the env vars with
// the prefix are captured. cb is an optional callback that takes
// a string and returns a string (the env variable name) in case
// transformations have to be applied, for instance, to lowercase
// everything, strip prefixes and replace _ with . etc.
// If the callback returns an empty string, the variable will be
// ignored.
//
// It takes an optional Opt argument containing a function to override
// the default source for environment variables, which can be useful
// for mocking and parallel unit tests.
func Provider(delim string, opt ...Opt) *Env {
	e := &Env{
		delim:   delim,
		environ: os.Environ,
	}

	if len(opt) == 0 {
		return e
	}

	o := opt[0]

	if o.EnvironFunc != nil {
		e.environ = o.EnvironFunc
	}
	e.transform = o.TransformFunc
	e.prefix = o.Prefix

	return e
}

// ReadBytes is not supported by the env provider.
func (e *Env) ReadBytes() ([]byte, error) {
	return nil, errors.New("env provider does not support this method")
}

// Read reads all available environment variables into a key:value map
// and returns it.
func (e *Env) Read() (map[string]any, error) {
	// Collect the environment variable keys.
	var keys []string
	for _, k := range e.environ() {
		if e.prefix != "" {
			if strings.HasPrefix(k, e.prefix) {
				keys = append(keys, k)
			}
		} else {
			keys = append(keys, k)
		}
	}

	mp := make(map[string]any)
	for _, k := range keys {
		parts := strings.SplitN(k, "=", 2)

		// If there's a transformation callback,
		// run it through every key/value.
		if e.transform != nil {
			key, value := e.transform(parts[0], parts[1])
			// If the callback blanked the key, it should be omitted
			if key == "" {
				continue
			}
			mp[key] = value
		} else {
			mp[parts[0]] = parts[1]
		}

	}

	if e.delim != "" {
		return maps.Unflatten(mp, e.delim), nil
	}

	return mp, nil
}
