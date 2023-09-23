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
	// cb is a callback function that allows the caller to modify the key and value as a string
	cb func(key string, value string) (string, interface{})
	// flagCB is a callback function that allows the caller to modify the key and flag.Value
	flagCb CallBack
}

// Provider returns a commandline flags provider that returns
// a nested map[string]interface{} of environment variable where the
// nesting hierarchy of keys are defined by delim. For instance, the
// delim "." will convert the key `parent.child.key: 1`
// to `{parent: {child: {key: 1}}}`.
// It takes two WithEnableMerge and WithCallBack options.
func Provider(f *flag.FlagSet, delim string, opts ...Option) *Pflag {
	p := Pflag{
		flagset: f,
		delim:   delim,
	}

	for _, opt := range opts {
		opt(&p)
	}

	return &p
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
		if p.ko != nil {
			p.readWithMerge(mp, f)
			return
		}

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

func (p *Pflag) readWithMerge(mp map[string]interface{}, f *flag.Flag) {
	var (
		key   string
		value any
	)

	if p.flagCb != nil {
		key, value = p.flagCb(f.Name, f.Value)
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
}

// ReadBytes is not supported by the basicflag koanf.
func (p *Pflag) ReadBytes() ([]byte, error) {
	return nil, errors.New("basicflag provider does not support this method")
}
