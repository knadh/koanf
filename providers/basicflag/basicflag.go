// Package basicflag implements a koanf.Provider that reads commandline
// parameters as conf maps using the Go's flag package.
package basicflag

import (
	"errors"
	"flag"

	"github.com/knadh/koanf/maps"
)

// Pflag implements a pflag command line provider.
type Pflag struct {
	delim   string
	flagset *flag.FlagSet
	cb      func(key string, value string) (string, interface{})
}

// Provider returns a commandline flags provider that returns
// a nested map[string]interface{} of environment variable where the
// nesting hierarchy of keys are defined by delim. For instance, the
// delim "." will convert the key `parent.child.key: 1`
// to `{parent: {child: {key: 1}}}`.
func Provider(f *flag.FlagSet, delim string) *Pflag {
	return &Pflag{
		flagset: f,
		delim:   delim,
	}
}

// ProviderWithValue works exactly the same as Provider except the callback
// takes a (key, value) with the variable name and value and allows you
// to modify both. This is useful for cases where you may want to return
// other types like a string slice instead of just a string.
func ProviderWithValue(f *flag.FlagSet, delim string, cb func(key string, value string) (string, interface{})) *Pflag {
	return &Pflag{
		flagset: f,
		delim:   delim,
		cb:      cb,
	}
}

// Read reads the flag variables and returns a nested conf map.
func (p *Pflag) Read() (map[string]interface{}, error) {
	mp := make(map[string]interface{})
	p.flagset.VisitAll(func(f *flag.Flag) {
		if p.cb != nil {
			key, value := p.cb(f.Name, f.Value.String())
			// If the callback blanked the key, it should be omitted
			if key == "" {
				return
			}
			mp[key] = value
		} else {
			mp[f.Name] = f.Value.String()
		}
	})
	return maps.Unflatten(mp, p.delim), nil
}

// ReadBytes is not supported by the basicflag koanf.
func (p *Pflag) ReadBytes() ([]byte, error) {
	return nil, errors.New("basicflag provider does not support this method")
}

// Watch is not supported.
func (p *Pflag) Watch(cb func(event interface{}, err error)) error {
	return errors.New("basicflag provider does not support this method")
}
