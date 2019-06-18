// Package posflag implements a koanf.Provider that reads commandline
// parameters as conf maps using spf13/pflag, a POSIX compliant
// alternative to Go's stdlib flag package.
package posflag

import (
	"errors"

	"github.com/knadh/koanf/maps"
	"github.com/spf13/pflag"
)

// Posflag implements a pflag command line provider.
type Posflag struct {
	delim   string
	flagset *pflag.FlagSet
}

// Provider returns a commandline flags provider that returns
// a nested map[string]interface{} of environment variable where the
// nesting hierarchy of keys are defined by delim. For instance, the
// delim "." will convert the key `parent.child.key: 1`
// to `{parent: {child: {key: 1}}}`.
func Provider(f *pflag.FlagSet, delim string) *Posflag {
	return &Posflag{
		flagset: f,
		delim:   delim,
	}
}

// Read reads the flag variables and returns a nested conf map.
func (p *Posflag) Read() (map[string]interface{}, error) {
	mp := make(map[string]interface{})
	p.flagset.VisitAll(func(f *pflag.Flag) {
		// If no value was explicity specified by the user,
		// ignore the flag.
		if !f.Changed {
			return
		}

		var v interface{}
		switch f.Value.Type() {
		case "int":
			i, _ := p.flagset.GetInt(f.Name)
			v = int64(i)
		case "int8":
			i, _ := p.flagset.GetInt8(f.Name)
			v = int64(i)
		case "int16":
			i, _ := p.flagset.GetInt16(f.Name)
			v = int64(i)
		case "int32":
			i, _ := p.flagset.GetInt32(f.Name)
			v = int64(i)
		case "int64":
			i, _ := p.flagset.GetInt64(f.Name)
			v = int64(i)
		case "float32":
			v, _ = p.flagset.GetFloat32(f.Name)
		case "float":
			v, _ = p.flagset.GetFloat64(f.Name)
		case "bool":
			v, _ = p.flagset.GetBool(f.Name)
		case "stringSlice":
			v, _ = p.flagset.GetStringSlice(f.Name)
		default:
			v = f.Value.String()
		}

		mp[f.Name] = v
	})
	return maps.Unflatten(mp, p.delim), nil
}

// ReadBytes is not supported by the env koanf.
func (p *Posflag) ReadBytes() ([]byte, error) {
	return nil, errors.New("pflag provider does not support this method")
}
