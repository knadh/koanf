// Package basicflag implements a koanf.Provider that reads commandline
// parameters as conf maps using the Go's flag package.
package basicflag

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

// Pflag implements a pflag command line provider.
type Pflag struct {
	ko      KoanfIntf
	delim   string
	flagset *flag.FlagSet
	cb      CallBack
}

// Provider returns a commandline flags provider that returns
// a nested map[string]interface{} of environment variable where the
// nesting hierarchy of keys are defined by delim. For instance, the
// delim "." will convert the key `parent.child.key: 1`
// to `{parent: {child: {key: 1}}}`.
func Provider(f *flag.FlagSet, delim string, ko KoanfIntf) *Pflag {
	return &Pflag{
		flagset: f,
		delim:   delim,
		ko:      ko,
	}
}

// ProviderWithValue works exactly the same as Provider except the callback
// takes a (key, value) with the variable name and value and allows you
// to modify both. This is useful for cases where you may want to return
// other types like a string slice instead of just a string.
func ProviderWithValue(f *flag.FlagSet, delim string, ko KoanfIntf, cb CallBack) *Pflag {
	return &Pflag{
		ko:      ko,
		flagset: f,
		delim:   delim,
		cb:      cb,
	}
}

// Read reads the flag variables and returns a nested conf map.
func (p *Pflag) Read() (map[string]interface{}, error) {
	mp := make(map[string]interface{})

	p.flagset.VisitAll(func(f *flag.Flag) {
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

		if key == "" {
			return
		}

		// if the key is set, and the flag value is the default value, skip it
		if p.ko.Exists(key) && f.Value.String() == f.DefValue {
			return
		}

		mp[key] = value
	})

	return maps.Unflatten(mp, p.delim), nil
}

// ReadBytes is not supported by the basicflag koanf.
func (p *Pflag) ReadBytes() ([]byte, error) {
	return nil, errors.New("basicflag provider does not support this method")
}
