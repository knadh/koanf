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

// Read reads the flag variables and returns a nested conf map.
func (p *Pflag) Read() (map[string]interface{}, error) {
	mp := make(map[string]interface{})
	p.flagset.VisitAll(func(f *flag.Flag) {
		mp[f.Name] = f.Value.String()
	})
	return maps.Unflatten(mp, p.delim), nil
}

// ReadBytes is not supported by the env koanf.
func (p *Pflag) ReadBytes() ([]byte, error) {
	return nil, errors.New("pflag provider does not support this method")
}
