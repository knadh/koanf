// Package cliflagv3 implements a koanf.Provider that reads commandline
// parameters as conf maps using urfave/cli/v3 flag.
package cliflagv3

import (
	"errors"
	"slices"
	"strings"

	"github.com/knadh/koanf/maps"
	"github.com/urfave/cli/v3"
)

// CliFlag implements a cli.Flag command line provider.
type CliFlag struct {
	cmd    *cli.Command
	delim  string
	config *Config
}

type Config struct {
	Defaults []string
}

// Provider returns a commandline flags provider that returns
// a nested map[string]any of environment variable where the
// nesting hierarchy of keys are defined by delim. For instance, the
// delim "." will convert the key `parent.child.key: 1`
// to `{parent: {child: {key: 1}}}`.
func Provider(f *cli.Command, delim string) *CliFlag {
	return &CliFlag{
		cmd:   f,
		delim: delim,
		config: &Config{
			Defaults: []string{},
		},
	}
}

// ProviderWithConfig returns a commandline flags provider with a
// Configuration struct attached.
func ProviderWithConfig(f *cli.Command, delim string, config *Config) *CliFlag {
	return &CliFlag{
		cmd:    f,
		delim:  delim,
		config: config,
	}
}

// ReadBytes is not supported by the cliflagv3 provider.
func (p *CliFlag) ReadBytes() ([]byte, error) {
	return nil, errors.New("cliflagv3 provider does not support this method")
}

// Read reads the flag variables and returns a nested conf map.
func (p *CliFlag) Read() (map[string]any, error) {
	out := make(map[string]any)

	// Get command lineage (from root to current command)
	lineage := p.cmd.Lineage()
	if len(lineage) > 0 {
		// Build command path and process flags for each level
		var cmdPath []string
		for i := len(lineage) - 1; i >= 0; i-- {
			cmd := lineage[i]
			cmdPath = append(cmdPath, cmd.Name)
			prefix := strings.Join(cmdPath, p.delim)
			p.processFlags(cmd.Flags, prefix, out)
		}
	}

	if p.delim == "" {
		return out, nil
	}

	return maps.Unflatten(out, p.delim), nil
}

func (p *CliFlag) processFlags(flags []cli.Flag, prefix string, out map[string]any) {
	for _, flag := range flags {
		name := flag.Names()[0]
		if p.cmd.IsSet(name) || slices.Contains(p.config.Defaults, name) || p.cmd.Value(name) != nil {
			value := p.getFlagValue(name)
			if value != nil {
				// Build the full path for the flag
				fullPath := name
				if prefix != "global" {
					fullPath = prefix + p.delim + name
				}

				p.setNestedValue(fullPath, value, out)
			}
		}
	}
}

// setNestedValue sets a value in the nested configuration structure
func (p *CliFlag) setNestedValue(path string, value any, out map[string]any) {
	parts := strings.Split(path, p.delim)
	current := out

	// Navigate/create the nested structure
	for i := 0; i < len(parts)-1; i++ {
		if _, exists := current[parts[i]]; !exists {
			current[parts[i]] = make(map[string]any)
		}
		current = current[parts[i]].(map[string]any)
	}

	// Set the final value
	current[parts[len(parts)-1]] = value
}

// getFlagValue extracts the typed value from the flag.
func (p *CliFlag) getFlagValue(name string) any {
	// Find the flag definition
	flag := p.findFlag(name)
	if flag == nil {
		return nil
	}
	return flag.Get()
}

// findFlag looks up a flag by name
func (p *CliFlag) findFlag(name string) cli.Flag {
	// Check global flags
	for _, f := range p.cmd.Flags {
		if slices.Contains(f.Names(), name) {
			return f
		}
	}

	return nil
}
