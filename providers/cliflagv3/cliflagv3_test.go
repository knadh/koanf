package cliflagv3

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

// mockKoanf is a mock implementation of KoanfIntf for testing.
type mockKoanf struct {
	keys map[string]bool
}

func (m *mockKoanf) Exists(key string) bool {
	return m.keys[key]
}

func TestCliFlag(t *testing.T) {
	cliApp := cli.Command{
		Name: "testing",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			p := Provider(cmd, ".")
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
				Sources: cli.EnvVars("TEST_FLAG"),
			},
		},
		Commands: []*cli.Command{
			{
				Name:        "x",
				Description: "yeah yeah testing",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					p := ProviderWithConfig(cmd, ".", &Config{Defaults: []string{"other"}})
					x, err := p.Read()
					require.NoError(t, err)
					require.NotEmpty(t, x)
					fmt.Printf("x: %s\n", x)

					k := koanf.New(".")
					err = k.Load(p, nil)

					fmt.Printf("k.All(): %v\n", k.All())

					require.Equal(t, k.String("testing.x.lol"), "dsf")
					// "default" was not explicitly set and not in Defaults,
					// so it should not be emitted.
					require.Equal(t, k.String("testing.x.default"), "")
					require.Equal(t, k.String("testing.x.other"), "")
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
						Sources:  cli.EnvVars("TEST_FLAG"),
					},
					&cli.StringFlag{
						Name:  "default",
						Usage: "default test flag",
						Value: "test",
					},
					&cli.StringFlag{
						Name:  "other",
						Usage: "other test flag",
					},
				},
			},
		},
	}

	// The Action of the testing only runs if no subcommand is specified
	x := []string{"testing", "--test", "gf"}
	err := cliApp.Run(context.Background(), append(x, os.Environ()...))
	require.NoError(t, err)

	// This runs the Action of the x
	x = []string{"testing", "x", "--lol", "dsf"}
	err = cliApp.Run(context.Background(), append(x, os.Environ()...))
	require.NoError(t, err)
}

func TestCliFlagDefaults(t *testing.T) {
	t.Run("without KeyMap unset flags are skipped", func(t *testing.T) {
		cliApp := cli.Command{
			Name: "app",
			Action: func(ctx context.Context, cmd *cli.Command) error {
				p := Provider(cmd, ".")
				mp, err := p.Read()
				require.NoError(t, err)

				// "name" was explicitly set, so it must appear.
				require.Contains(t, mp, "app")
				app := mp["app"].(map[string]any)
				require.Equal(t, "explicit", app["name"])

				// "color" was NOT set and is not in Defaults, so it must be absent.
				require.NotContains(t, app, "color")
				return nil
			},
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "name", Value: "default-name"},
				&cli.StringFlag{Name: "color", Value: "red"},
			},
		}
		err := cliApp.Run(context.Background(), []string{"app", "--name", "explicit"})
		require.NoError(t, err)
	})

	t.Run("with KeyMap existing keys are not overwritten", func(t *testing.T) {
		cliApp := cli.Command{
			Name: "app",
			Action: func(ctx context.Context, cmd *cli.Command) error {
				km := &mockKoanf{keys: map[string]bool{
					"app.color": true,
				}}
				p := ProviderWithConfig(cmd, ".", &Config{KeyMap: km})
				mp, err := p.Read()
				require.NoError(t, err)

				app := mp["app"].(map[string]any)
				require.Equal(t, "explicit", app["name"])
				require.NotContains(t, app, "color")
				require.Equal(t, "large", app["size"])
				return nil
			},
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "name", Value: "default-name"},
				&cli.StringFlag{Name: "color", Value: "red"},
				&cli.StringFlag{Name: "size", Value: "large"},
			},
		}
		err := cliApp.Run(context.Background(), []string{"app", "--name", "explicit"})
		require.NoError(t, err)
	})
}
