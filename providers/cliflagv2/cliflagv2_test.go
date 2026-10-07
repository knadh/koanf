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

func TestCliFlagParentFlagsInSubcommand(t *testing.T) {
	run := func(t *testing.T, args []string) map[string]any {
		var all map[string]any
		cliApp := cli.App{
			Name: "app",
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "name"},
			},
			Commands: []*cli.Command{
				{
					Name: "mid",
					Flags: []cli.Flag{
						&cli.StringFlag{Name: "region"},
						&cli.IntFlag{Name: "port"},
					},
					Subcommands: []*cli.Command{
						{
							Name: "leaf",
							Flags: []cli.Flag{
								&cli.StringFlag{Name: "name"},
							},
							Action: func(ctx *cli.Context) error {
								k := koanf.New(".")
								require.NoError(t, k.Load(Provider(ctx, "."), nil))
								all = k.All()
								return nil
							},
						},
					},
				},
			},
		}
		require.NoError(t, cliApp.Run(args))
		return all
	}

	t.Run("a middle command's flags are kept", func(t *testing.T) {
		all := run(t, []string{"app", "mid", "--region", "eu", "--port", "80", "leaf"})
		require.Equal(t, map[string]any{"app.mid.region": "eu", "app.mid.port": 80}, all)
	})

	t.Run("a parent and child sharing a name keep their own values", func(t *testing.T) {
		all := run(t, []string{"app", "--name", "root", "mid", "leaf", "--name", "child"})
		require.Equal(t, map[string]any{"app.name": "root", "app.mid.leaf.name": "child"}, all)
	})

	t.Run("a child's flag does not set the parent's key", func(t *testing.T) {
		all := run(t, []string{"app", "mid", "leaf", "--name", "child"})
		require.Equal(t, map[string]any{"app.mid.leaf.name": "child"}, all)
	})
}
