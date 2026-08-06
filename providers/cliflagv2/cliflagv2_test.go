package cliflagv2

import (
	"fmt"
	"os"
	"testing"

	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestCliFlag(t *testing.T) {
	cliApp := cli.App{
		Name: "testing",
		Action: func(ctx *cli.Context) error {
			p := Provider(ctx, ".")
			x, err := p.Read()
			require.NoError(t, err)
			require.NotEmpty(t, x)

			fmt.Printf("x: %v\n", x)

			k := koanf.New(".")
			err = k.Load(p, nil)

			fmt.Printf("k.All(): %v\n", k.All())

			return nil
		},
		Flags: []cli.Flag{
			cli.HelpFlag,
			cli.VersionFlag,
			&cli.StringFlag{
				Name:    "test",
				Usage:   "test flag",
				Value:   "test",
				Aliases: []string{"t"},
				EnvVars: []string{"TEST_FLAG"},
			},
		},
		Commands: []*cli.Command{
			{
				Name:        "x",
				Description: "yeah yeah testing",
				Action: func(ctx *cli.Context) error {
					p := Provider(ctx, ".")
					x, err := p.Read()
					require.NoError(t, err)
					require.NotEmpty(t, x)
					fmt.Printf("x: %s\n", x)

					k := koanf.New(".")
					err = k.Load(p, nil)

					fmt.Printf("k.All(): %v\n", k.All())

					require.Equal(t, k.String("testing.x.lol"), "dsf")
					return nil
				},
				Flags: []cli.Flag{
					cli.HelpFlag,
					cli.VersionFlag,
					&cli.StringFlag{
						Name:     "lol",
						Usage:    "test flag",
						Value:    "test",
						Required: true,
						EnvVars:  []string{"TEST_FLAG"},
					},
				},
			},
		},
	}

	x := append([]string{"testing", "--test", "gf", "x", "--lol", "dsf"}, os.Args...)
	err := cliApp.Run(append(x, os.Environ()...))
	require.NoError(t, err)
}

// A flag whose name is a delimiter-prefix of another flag's name, for eg:,
// `--db` and `--db.host`, used to panic on an unchecked type assertion while
// building the nested map. The deeper key must win instead.
func TestCliFlagV2PrefixCollision(t *testing.T) {
	cliApp := cli.App{
		Name: "app",
		Action: func(ctx *cli.Context) error {
			p := ProviderWithConfig(ctx, ".", &Config{Defaults: []string{"db", "db.host"}})
			mp, err := p.Read()
			require.NoError(t, err)
			require.Equal(t, map[string]any{
				"app": map[string]any{
					"db": map[string]any{"host": "localhost"},
				},
			}, mp)
			return nil
		},
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "db", Value: "prod"},
			&cli.StringFlag{Name: "db.host", Value: "localhost"},
		},
	}
	require.NoError(t, cliApp.Run([]string{"app"}))
}

// The same collision across command levels: a subcommand whose name matches a
// flag on its parent command.
func TestCliFlagV2SubCommandNameCollision(t *testing.T) {
	cliApp := cli.App{
		Name:  "app",
		Flags: []cli.Flag{&cli.StringFlag{Name: "db", Value: "prod"}},
		Commands: []*cli.Command{
			{
				Name:  "db",
				Flags: []cli.Flag{&cli.StringFlag{Name: "host", Value: "localhost"}},
				Action: func(ctx *cli.Context) error {
					p := ProviderWithConfig(ctx, ".", &Config{Defaults: []string{"db", "host"}})
					mp, err := p.Read()
					require.NoError(t, err)
					require.Equal(t, map[string]any{
						"app": map[string]any{
							"db": map[string]any{"host": "localhost"},
						},
					}, mp)
					return nil
				},
			},
		},
	}
	require.NoError(t, cliApp.Run([]string{"app", "db"}))
}

// With an empty delimiter there is no nesting to do, so flag keys must be
// returned as-is. They used to be split into one level per character.
func TestCliFlagV2NoDelim(t *testing.T) {
	cliApp := cli.App{
		Name: "app",
		Action: func(ctx *cli.Context) error {
			p := ProviderWithConfig(ctx, "", &Config{Defaults: []string{"host"}})
			mp, err := p.Read()
			require.NoError(t, err)
			require.Equal(t, map[string]any{"apphost": "localhost"}, mp)
			return nil
		},
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "host", Value: "localhost"},
		},
	}
	require.NoError(t, cliApp.Run([]string{"app"}))
}
