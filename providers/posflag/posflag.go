// Package posflag implements a koanf.Provider that reads commandline
// parameters as conf maps using spf13/pflag, a POSIX compliant
// alternative to Go's stdlib flag package.
package posflag

import (
	"errors"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/maps"
	"github.com/spf13/pflag"
)

// Posflag implements a pflag command line provider.
type Posflag struct {
	delim   string
	flagset *pflag.FlagSet
	ko      *koanf.Koanf
	cb      func(key string, value string) (string, interface{})
}

// Provider returns a commandline flags provider that returns
// a nested map[string]interface{} of environment variable where the
// nesting hierarchy of keys are defined by delim. For instance, the
// delim "." will convert the key `parent.child.key: 1`
// to `{parent: {child: {key: 1}}}`.
//
// It takes an optional (but recommended) Koanf instance to see if the
// the flags defined have been set from other providers, for instance,
// a config file. If they are not, then the default values of the flags
// are merged. If they do exist, the flag values are not merged but only
// the values that have been explicitly set in the command line are merged.
func Provider(f *pflag.FlagSet, delim string, ko *koanf.Koanf) *Posflag {
	return &Posflag{
		flagset: f,
		delim:   delim,
		ko:      ko,
	}
}

// ProviderWithValue works exactly the same as Provider except the callback
// takes a (key, value) with the variable name and value and allows their modification.
// This is useful for cases where complex types like slices separated by
// custom separators.
func ProviderWithValue(f *pflag.FlagSet, delim string, ko *koanf.Koanf, cb func(key string, value string) (string, interface{})) *Posflag {
	return &Posflag{
		flagset: f,
		delim:   delim,
		ko:      ko,
		cb:      cb,
	}
}

// Read reads the flag variables and returns a nested conf map.
func (p *Posflag) Read() (map[string]interface{}, error) {
	mp := make(map[string]interface{})
	p.flagset.VisitAll(func(f *pflag.Flag) {
		var (
			key = f.Name
			val interface{}
		)

		switch f.Value.Type() {
		case "int":
			i, _ := p.flagset.GetInt(key)
			val = int64(i)
		case "int8":
			i, _ := p.flagset.GetInt8(key)
			val = int64(i)
		case "int16":
			i, _ := p.flagset.GetInt16(key)
			val = int64(i)
		case "int32":
			i, _ := p.flagset.GetInt32(key)
			val = int64(i)
		case "int64":
			i, _ := p.flagset.GetInt64(key)
			val = int64(i)
		case "float32":
			val, _ = p.flagset.GetFloat32(key)
		case "float":
			val, _ = p.flagset.GetFloat64(key)
		case "bool":
			val, _ = p.flagset.GetBool(key)
		case "stringSlice":
			val, _ = p.flagset.GetStringSlice(key)
		case "intSlice":
			val, _ = p.flagset.GetIntSlice(key)
		case "stringToString":
			val, _ = p.flagset.GetStringToString(key)
		case "stringToInt":
			val, _ = p.flagset.GetStringToInt(key)
		case "stringToInt64":
			val, _ = p.flagset.GetStringToInt64(key)
		default:
			val = f.Value.String()
		}

		// If there is a callback set, pass the key and value to
		// it and use the resultant transformed values instead.
		if p.cb != nil {
			k, v := p.cb(key, f.Value.String())
			if k == "" {
				return
			}

			key = k
			val = v
		}

		// If the default value of the flag was never changed by the user,
		// it should not override the value in the conf map (if it exists in the first place).
		if !f.Changed {
			if p.ko != nil {
				if p.ko.Exists(key) {
					return
				}
			} else {
				return
			}
		}

		// No callback. Use the key and value as-is.
		mp[key] = val
	})

	return maps.Unflatten(mp, p.delim), nil
}

// ReadBytes is not supported by the env koanf.
func (p *Posflag) ReadBytes() ([]byte, error) {
	return nil, errors.New("pflag provider does not support this method")
}

// Watch is not supported.
func (p *Posflag) Watch(cb func(event interface{}, err error)) error {
	return errors.New("posflag provider does not support this method")
}
