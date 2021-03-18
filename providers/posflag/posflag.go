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
	delim     string
	flagset   *pflag.FlagSet
	ko        *koanf.Koanf
	cb        func(key string, value string) (string, interface{})
	renameMap map[string]string
}

type Option func(*Posflag)

// Provider returns a commandline flags provider that returns
// a nested map[string]interface{} of environment variable where the
// nesting hierarchy of keys are defined by delim. For instance, the
// delim "." will convert the key `parent.child.key: 1`
// to `{parent: {child: {key: 1}}}`.
func Provider(f *pflag.FlagSet, delim string, options ...Option) *Posflag {
	p := &Posflag{
		flagset: f,
		delim:   delim,
	}
	if options != nil {
		for _, opt := range options {
			opt(p)
		}
	}
	return p
}

// ProviderWithValue works exactly the same as Provider except the callback
// takes a (key, value) with the variable name and value and allows you
// to modify both. This is useful for cases where you may want to return
// other types like a string slice instead of just a string.
// Deprecated: this function is deprecated, use WithValue option to achieve the same result
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
		flagName := p.getFlagName(f)
		// If no value was explicitly set in the command line,
		// check if the default value should be used.
		if !f.Changed {
			if p.ko != nil {
				if p.ko.Exists(flagName) {
					return
				}
			} else {
				return
			}
		}

		var v interface{}
		switch f.Value.Type() {
		case "int":
			i, _ := p.flagset.GetInt(flagName)
			v = int64(i)
		case "int8":
			i, _ := p.flagset.GetInt8(flagName)
			v = int64(i)
		case "int16":
			i, _ := p.flagset.GetInt16(flagName)
			v = int64(i)
		case "int32":
			i, _ := p.flagset.GetInt32(flagName)
			v = int64(i)
		case "int64":
			i, _ := p.flagset.GetInt64(flagName)
			v = int64(i)
		case "float32":
			v, _ = p.flagset.GetFloat32(flagName)
		case "float":
			v, _ = p.flagset.GetFloat64(flagName)
		case "bool":
			v, _ = p.flagset.GetBool(flagName)
		case "stringSlice":
			v, _ = p.flagset.GetStringSlice(flagName)
		case "intSlice":
			v, _ = p.flagset.GetIntSlice(flagName)
		default:
			if p.cb != nil {
				key, value := p.cb(flagName, f.Value.String())
				if key == "" {
					return
				}
				mp[key] = value
				return
			} else {
				v = f.Value.String()
			}
		}

		mp[flagName] = v
	})
	return maps.Unflatten(mp, p.delim), nil
}

func (p *Posflag) getFlagName(flag *pflag.Flag) string {
	name := flag.Name
	if p.renameMap != nil {
		if keyName, found := p.renameMap[name]; found {
			name = keyName
		}
	}
	return name
}

// ReadBytes is not supported by the env koanf.
func (p *Posflag) ReadBytes() ([]byte, error) {
	return nil, errors.New("pflag provider does not support this method")
}

// Watch is not supported.
func (p *Posflag) Watch(cb func(event interface{}, err error)) error {
	return errors.New("posflag provider does not support this method")
}

// WithKoanf option adds the Koanf instance to see if the
// the flags defined have been set from other providers, for instance,
// a config file. If they are not, then the default values of the flags
// are merged. If they do exist, the flag values are not merged but only
// the values that have been explicitly set in the command line are merged.
func WithKoanf(ko *koanf.Koanf) Option {
	return func(p *Posflag) {
		p.ko = ko
	}
}

// WithValue options adds the callback
// takes a (key, value) with the variable name and value and allows you
// to modify both. This is useful for cases where you may want to return
// other types like a string slice instead of just a string.
func WithValue(cb func(key string, value string) (string, interface{})) Option {
	return func(p *Posflag) {
		p.cb = cb
	}
}

// WithRenameKeys options adds the possibility to map flags in case when flag name
// differs from setting name.
func WithRenameKeys(flagMap map[string]*pflag.Flag) Option {
	return func(p *Posflag) {
		renameMap := make(map[string]string)
		for k, v := range flagMap {
			// add a display only for an existing flags
			found := p.flagset.Lookup(v.Name)
			if found != nil {
				renameMap[v.Name] = k
			}
		}
		p.renameMap = renameMap
	}
}
