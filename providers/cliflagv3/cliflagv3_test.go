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
					p := Provider(cmd, ".")
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
						Sources:  cli.EnvVars("TEST_FLAG"),
					},
				},
			},
		},
	}

	x := []string{"testing", "--test", "gf", "x", "--lol", "dsf"}
	err := cliApp.Run(context.Background(), append(x, os.Environ()...))
	require.NoError(t, err)
}
