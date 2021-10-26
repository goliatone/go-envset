package template

import (
	"os"
	"path/filepath"

	"github.com/goliatone/go-envset/pkg/config"
	"github.com/goliatone/go-envset/pkg/envset"
	"github.com/urfave/cli/v2"
)

//GetCommand exports template command
func GetCommand(cnf *config.Config) *cli.Command {
	return &cli.Command{
		//TODO: This actually should load a template file and resolve it using the context.
		//Default template should generate envset.example
		Name:        "template",
		Usage:       "make a template file from an environment",
		Description: "create a new template or update file to document the variables in your environment",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "print", Usage: "only print the contents to stdout, don't write file"},
			&cli.StringFlag{Name: "filename", Usage: "template file `name`", Value: cnf.Template.File},
			&cli.StringFlag{Name: "filepath", Usage: "template file `path`", Value: cnf.Template.Path},
			&cli.StringFlag{Name: "env-file", Value: cnf.Filename, Usage: "load environment from `FILE`"},
			&cli.BoolFlag{Name: "overwrite", Usage: "overwrite file, this will delete any changes"},
		},
		Action: func(c *cli.Context) error {
			print := c.Bool("print")
			filename := c.String("env-file")
			template := c.String("filename")
			dir := c.String("filepath")
			overwrite := c.Bool("overwrite")

			dir, err := filepath.Abs(dir)
			if err != nil {
				return err
			}

			if _, err = os.Stat(dir); os.IsNotExist(err) {
				if err = os.MkdirAll(dir, os.ModePerm); err != nil {
					return err
				}
			}
			//TODO: This should take a a template file which we use to run against our thing
			template = filepath.Join(dir, template)

			return envset.DocumentTemplate(filename, template, overwrite, print)
		},
	}
}
