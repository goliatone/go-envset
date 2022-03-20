package rc

import (
	"errors"
	"fmt"

	"github.com/goliatone/go-envset/pkg/config"
	"github.com/urfave/cli/v2"
)

//GetCommand returns a new cli.Command for the
//rc command.
func GetCommand(cnf *config.Config) *cli.Command {
	return &cli.Command{
		Name:        "config",
		Aliases:     []string{"rc"},
		Usage:       "generate an envsetrc file",
		Description: "creates an envsetrc file with either the current options or the default options",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "print", Usage: "only print the contents to stdout, don't write file"},
			&cli.StringFlag{Name: "filename", Usage: "metadata file `name`", Value: cnf.Meta.File},
			&cli.StringFlag{Name: "filepath", Usage: "metadata file `path`", Value: cnf.Meta.Dir},
			&cli.StringFlag{Name: "env-file", Value: cnf.Filename, Usage: "load environment from `FILE`"},
			&cli.BoolFlag{Name: "overwrite", Usage: "set true to prevent overwrite metadata file", Value: true},
			&cli.BoolFlag{Name: "values", Usage: "add flag to show values in the output"},
			&cli.BoolFlag{Name: "globals", Usage: "include global section", Value: false},
			&cli.StringFlag{Name: "secret", Usage: "`password` used to encode hash values", EnvVars: []string{"ENVSET_HASH_SECRET"}},
		},
		Action: func(c *cli.Context) error {
			fmt.Printf("%s", config.GetDefaultConfig())
			return nil
		},
		Subcommands: []*cli.Command{
			{
				Name:        "get",
				Usage:       "get option value for key",
				UsageText:   "get option value for key",
				Description: "retrieves configuration value of given key",
				Action: func(c *cli.Context) error {
					if c.Args().Len() == 0 {
						return errors.New("envset config get requires exactly one argument, e.g.\nenvset config get <path>")
					}
					key := c.Args().First()
					val := cnf.Get(key)
					fmt.Println(val)
					return nil
				},
			},
			{
				Name:        "list",
				Usage:       "list available key paths",
				UsageText:   "list available key paths",
				Description: "prints a list of all key paths for configuration options",
				Action: func(c *cli.Context) error {
					keys := cnf.ListKeys()
					for _, k := range keys {
						fmt.Println(k)
					}
					return nil
				},
			},
		},
	}
}
