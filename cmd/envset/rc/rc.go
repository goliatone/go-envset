package rc

import (
	"fmt"

	"github.com/goliatone/go-envset/pkg/config"
	"github.com/urfave/cli/v2"
)

//GetCommand returns a new cli.Command for the
//rc command.
func GetCommand(cnf *config.Config) *cli.Command {
	return &cli.Command{
		Name:        "rc",
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
			// print := c.Bool("print")
			// envfile := c.String("env-file")
			// filename := c.String("filename")
			// originalDir := c.String("filepath")
			// overwrite := c.Bool("overwrite")
			// values := c.Bool("values")
			// globals := c.Bool("globals")
			// secret := c.String("secret")

			fmt.Printf("%s", config.GetDefaultConfig())

			return nil
		},
	}
}
