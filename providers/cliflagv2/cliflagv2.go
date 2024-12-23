// Package cliflagv2 implements a koanf.Provider that reads commandline
// parameters as conf maps using ufafe/cli/v2 flag.
package cliflagv2

import (
	"strings"

	"github.com/knadh/koanf/maps"
	"github.com/urfave/cli/v2"
)

// CliFlag implements a cli.Flag command line provider.
type CliFlag struct {
	ctx   *cli.Context
	delim string
}

// Provider returns a commandline flags provider that returns
// a nested map[string]interface{} of environment variable where the
// nesting hierarchy of keys are defined by delim. For instance, the
// delim "." will convert the key `parent.child.key: 1`
// to `{parent: {child: {key: 1}}}`.
func Provider(f *cli.Context, delim string) *CliFlag {
	return &CliFlag{
		ctx:   f,
		delim: delim,
	}
}

// Read reads the flag variables and returns a nested conf map.
func (p *CliFlag) Read() (map[string]interface{}, error) {
	out := make(map[string]interface{})

	// Get command lineage (from root to current command)
	lineage := p.ctx.Lineage()
	if len(lineage) > 0 {
		// Build command path and process flags for each level
		var cmdPath []string
		for i := len(lineage) - 1; i >= 0; i-- {
			cmd := lineage[i]
			if cmd.Command == nil {
				continue
			}
			cmdPath = append(cmdPath, cmd.Command.Name)
			prefix := strings.Join(cmdPath, p.delim)
			p.processFlags(cmd.Command.Flags, prefix, out)
		}
	}

	if p.delim == "" {
		return out, nil
	}

	return maps.Unflatten(out, p.delim), nil
}

func (p *CliFlag) processFlags(flags []cli.Flag, prefix string, out map[string]interface{}) {
	for _, flag := range flags {
		name := flag.Names()[0]
		if p.ctx.IsSet(name) {
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
func (p *CliFlag) setNestedValue(path string, value interface{}, out map[string]interface{}) {
	parts := strings.Split(path, p.delim)
	current := out

	// Navigate/create the nested structure
	for i := 0; i < len(parts)-1; i++ {
		if _, exists := current[parts[i]]; !exists {
			current[parts[i]] = make(map[string]interface{})
		}
		current = current[parts[i]].(map[string]interface{})
	}

	// Set the final value
	current[parts[len(parts)-1]] = value
}

// getFlagValue extracts the typed value from the flag
func (p *CliFlag) getFlagValue(name string) interface{} {
	switch {
	case p.ctx.IsSet(name):
		switch {
		case p.ctx.String(name) != "":
			return p.ctx.String(name)
		case p.ctx.StringSlice(name) != nil:
			return p.ctx.StringSlice(name)
		case p.ctx.Int(name) != 0:
			return p.ctx.Int(name)
		case p.ctx.Int64(name) != 0:
			return p.ctx.Int64(name)
		case p.ctx.IntSlice(name) != nil:
			return p.ctx.IntSlice(name)
		case p.ctx.Float64(name) != 0:
			return p.ctx.Float64(name)
		case p.ctx.Bool(name):
			return p.ctx.Bool(name)
		case p.ctx.Duration(name).String() != "0s":
			return p.ctx.Duration(name)
		default:
			return p.ctx.Generic(name)
		}
	}
	return nil
}
