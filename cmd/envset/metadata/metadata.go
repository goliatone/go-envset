package metadata

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/goliatone/go-envset/pkg/config"
	"github.com/goliatone/go-envset/pkg/envset"
	"github.com/tcnksm/go-gitconfig"
	"github.com/urfave/cli/v2"
)

//GetCommand returns a new cli.Command for the
//metadata command.
func GetCommand(cnf *config.Config) *cli.Command {
	return &cli.Command{
		Name:        "metadata",
		Usage:       "generate a metadata file from environment file",
		Description: "creates a metadata file with all the given environments",
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
			print := c.Bool("print")
			envfile := c.String("env-file")
			filename := c.String("filename")
			originalDir := c.String("filepath")
			overwrite := c.Bool("overwrite")
			values := c.Bool("values")
			globals := c.Bool("globals")
			secret := c.String("secret")

			//TODO: Handle case repo does not have a remote!
			projectURL, err := gitconfig.OriginURL()
			if err != nil {
				return err
			}

			dir, err := filepath.Abs(originalDir)
			if err != nil {
				return err
			}

			if _, err = os.Stat(dir); os.IsNotExist(err) {
				if err = os.MkdirAll(dir, os.ModePerm); err != nil {
					return err
				}
			}
			//TODO: This should take a a template file which we use to run against our thing
			filename = filepath.Join(dir, filename)

			algorithm := "sha256"
			if secret != "" {
				algorithm = "hmac"
			}

			o := envset.MetadataOptions{
				Name:          envfile,
				Filepath:      filename,
				Algorithm:     algorithm,
				Project:       projectURL,
				Globals:       globals,
				GlobalSection: "globals", //TODO: make flag
				Overwrite:     overwrite,
				Print:         print,
				Values:        values,
				Secret:        secret,
			}

			return envset.CreateMetadataFile(o)
		},
		Subcommands: []*cli.Command{
			{
				Name:        "compare",
				Usage:       "compare two metadata files",
				UsageText:   "envset metadata compare --section=[section] [source] [target]",
				Description: "compares the provided section of two metadata files",
				Category:    "METADATA",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "section",
						Usage:    "env file section",
						Required: true,
					},
					&cli.BoolFlag{
						Name:  "print",
						Usage: "print the comparison results to stdout",
						Value: cnf.Meta.Print,
					},
				},
				Action: func(c *cli.Context) error {
					print := c.Bool("print")
					name := c.String("section")
					source := c.Args().Get(0)
					target := c.Args().Get(1)

					src := envset.EnvFile{}
					src.FromJSON(source)

					s1, err := src.GetSection(name)
					if err != nil {
						return cli.Exit(fmt.Sprintf("Section \"%s\" not found in source metadata file:\n%s", name, source), 1)
					}

					tgt := envset.EnvFile{}
					tgt.FromJSON(target)
					s2, err := tgt.GetSection(name)

					if err != nil {
						return cli.Exit(fmt.Sprintf("Section \"%s\" not found in target metadata file.", name), 1)
					}

					s3 := envset.CompareSections(*s1, *s2)
					s3.Name = name

					if s3.IsEmpty() == false {
						if print {
							j, err := s3.ToJSON()
							if err != nil {
								return cli.Exit(err, 1)
							}
							return cli.Exit(j, 1)
						}
						//Exit with error e.g. to fail CI
						return cli.Exit("Metadata test failed!", 1)
					}

					return nil
				},
			},
		},
	}
}
