// Package basicflag implements a koanf.Provider that reads commandline
// parameters as conf maps using the Go's flag package.
package basicflag

import (
	"errors"
	"flag"

	"github.com/knadh/koanf/maps"
)

// Opt represents optional options (yup) passed to the provider.
type Opt struct {
	KeyMap KoanfIntf
}

// KoanfIntf is an interface that represents a small subset of methods
// used by this package from Koanf{}.
type KoanfIntf interface {
	Exists(string) bool
}

// Pflag implements a pflag command line provider.
type Pflag struct {
	delim   string
	flagset *flag.FlagSet
	cb      func(key string, value string) (string, any)
	opt     *Opt
}

// Provider returns a commandline flags provider that returns
// a nested map[string]any of environment variable where the
// nesting hierarchy of keys are defined by delim. For instance, the
// delim "." will convert the key `parent.child.key: 1`
// to `{parent: {child: {key: 1}}}`.
//
// It takes an optional (but recommended) Opt{} argument containing a Koanf instance.
// It checks if the defined flags have been set by other providers (e.g., a config file).
// If not, default flag values are merged. If they exist, flag values are merged only if
// explicitly set in the command line. The function is variadic to maintain backward compatibility.
// See https://github.com/knadh/koanf/issues/255
func Provider(f *flag.FlagSet, delim string, opt ...*Opt) *Pflag {
	pf := &Pflag{
		flagset: f,
		delim:   delim,
	}

	if len(opt) > 0 {
		pf.opt = opt[0]
	}

	return pf
}

// ProviderWithValue works exactly the same as Provider except the callback
// takes a (key, value) with the variable name and value and allows you
// to modify both. This is useful for cases where you may want to return
// other types like a string slice instead of just a string.
//
// It takes an optional Opt{} (but recommended) argument with a Koanf instance (opt.KeyMap) to see if
// the flags defined have been set from other providers, for instance,
// a config file. If they are not, then the default values of the flags
// are merged. If they do exist, the flag values are not merged but only
// the values that have been explicitly set in the command line are merged.
// It is a variadic function as a hack to ensure backwards compatibility with the
// function definition.
// See https://github.com/knadh/koanf/issues/255
func ProviderWithValue(f *flag.FlagSet, delim string, cb func(key string, value string) (string, any), ko ...KoanfIntf) *Pflag {
	pf := &Pflag{
		flagset: f,
		delim:   delim,
		cb:      cb,
	}

	if len(ko) > 0 {
		pf.opt = &Opt{
			KeyMap: ko[0],
		}
	}
	return pf
}

// Read reads the flag variables and returns a nested conf map.
func (p *Pflag) Read() (map[string]any, error) {
	var changed map[string]struct{}

	// Prepare a map of flags that have been explicitly set by the user as aa KeyMap instance of Koanf
	// has been provided.
	if p.opt != nil && p.opt.KeyMap != nil {
		changed = map[string]struct{}{}

		p.flagset.Visit(func(f *flag.Flag) {
			key := f.Name
			if p.cb != nil {
				key, _ = p.cb(f.Name, "")
			}
			if key == "" {
				return
			}

			changed[key] = struct{}{}
		})
	}

	mp := make(map[string]any)
	p.flagset.VisitAll(func(f *flag.Flag) {
		var (
			key     = f.Name
			val any = f.Value.String()
		)
		if p.cb != nil {
			k, v := p.cb(f.Name, f.Value.String())
			// If the callback blanked the key, it should be omitted
			if k == "" {
				return
			}

			key = k
			val = v
		}

		// If the default value of the flag was never changed by the user,
		// it should not override the value in the conf map (if it exists in the first place).
		if changed != nil {
			if _, ok := changed[key]; !ok {
				if p.opt.KeyMap.Exists(key) {
					return
				}
			}
		}

		mp[key] = val
	})
	return maps.Unflatten(mp, p.delim), nil
}

// ReadBytes is not supported by the basicflag koanf.
func (p *Pflag) ReadBytes() ([]byte, error) {
	return nil, errors.New("basicflag provider does not support this method")
}
