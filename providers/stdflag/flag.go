// Package stdflag implements a koanf.Provider that reads commandline parameters as conf maps.
package stdflag

import (
	"errors"
	"flag"
	"fmt"

	"github.com/knadh/koanf/maps"
)

// KoanfIntf is an interface that represents a small subset of methods
// used by this package from Koanf{}. When using this package, a live
// instance of Koanf{} should be passed.
type KoanfIntf interface {
	Exists(string) bool
}

// CallBack is a callback function that allows the caller to modify the key and value
type CallBack func(key string, value flag.Value) (string, any)

// Flag implements a pflag command line provider.
type Flag struct {
	delim   string
	flagSet *flag.FlagSet
	ko      KoanfIntf
	cb      CallBack
}

// Provider returns a commandline flags provider that returns
// a nested map[string]any of environment variable where the
// nesting hierarchy of keys are defined by delim. For instance, the
// delim "." will convert the key `parent.child.key: 1`
// to `{parent: {child: {key: 1}}}`.
//
// It takes an optional (but recommended) Koanf instance to see if the
// flags defined have been set from other providers, for instance,
// a config file. If they are not, then the default values of the flags
// are merged. If they do exist, the flag values are not merged but only
// the values that have been explicitly set in the command line are merged.
func Provider(f *flag.FlagSet, delim string, ko KoanfIntf) *Flag {
	return &Flag{
		flagSet: f,
		delim:   delim,
		ko:      ko,
	}
}

// ProviderWithKey works exactly the same as Provider except the callback
// takes the variable name allows their modification.
// This is useful for cases where complex types like slices separated by
// custom separators.
func ProviderWithKey(f *flag.FlagSet, delim string, ko KoanfIntf, cb CallBack) *Flag {
	return &Flag{
		flagSet: f,
		delim:   delim,
		ko:      ko,
		cb:      cb,
	}
}

// Read reads the flag variables and returns a nested conf map.
func (p *Flag) Read() (map[string]any, error) {
	mp := make(map[string]any)

	p.flagSet.VisitAll(func(f *flag.Flag) {
		var (
			key   string
			value any
		)

		if p.cb != nil {
			key, value = p.cb(f.Name, f.Value)
		} else {
			// All Value types provided by flag package satisfy the Getter interface
			// if user defined types are used, they must satisfy the Getter interface
			getter, ok := f.Value.(flag.Getter)
			if !ok {
				panic(fmt.Sprintf("flag %s does not implement flag.Getter", f.Name))
			}
			key, value = f.Name, getter.Get()
		}

		// if the key is set, and the flag value is the default value, skip it
		if p.ko.Exists(key) && f.Value.String() == f.DefValue {
			return
		}

		mp[key] = value
	})

	return maps.Unflatten(mp, p.delim), nil
}

// ReadBytes is not supported by the flag provider.
func (p *Flag) ReadBytes() ([]byte, error) {
	return nil, errors.New("flag provider does not support this method")
}
