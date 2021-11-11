package metadata

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/goliatone/go-envset/pkg/config"
	"github.com/goliatone/go-envset/pkg/envset"
	"github.com/gosuri/uitable"
	colors "github.com/logrusorgru/aurora/v3"
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
						Aliases:  []string{"s"},
						Usage:    "env file section",
						Required: true,
					},
					&cli.BoolFlag{
						Name:  "print",
						Usage: "print the comparison results to stdout",
						Value: cnf.Meta.Print,
					},
					&cli.BoolFlag{
						Name:  "json",
						Usage: "print the comparison results to stdout in JSON format",
						Value: cnf.Meta.AsJSON,
					},
				},
				Action: func(c *cli.Context) error {
					print := c.Bool("print")
					json := c.Bool("json")
					name := c.String("section")

					var source string
					var target string

					if c.Args().Len() == 1 {
						source, _ = envset.FileFinder(filepath.Join(cnf.Meta.Dir, cnf.Meta.File))
						target = c.Args().Get(0)
					} else {
						source = c.Args().Get(0)
						target = c.Args().Get(1)
					}

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
						if print && !json {
							prettyPrint(s3, source, target)
							return cli.Exit("", 1)
						} else if print && json {
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

func prettyPrint(diff envset.EnvSection, source, target string) {

	sort.Slice(diff.Keys, func(p, q int) bool {
		return diff.Keys[p].Comment > diff.Keys[q].Comment
	})

	fmt.Println("")

	mit := uitable.New()
	mit.MaxColWidth = 50

	mrt := uitable.New()
	mrt.MaxColWidth = 50

	dvt := uitable.New()
	dvt.MaxColWidth = 50

	mi := 0
	mr := 0
	dv := 0

	for _, k := range diff.Keys {
		if strings.Contains(k.Comment, "missing") {
			if mi == 0 {

				mit.AddRow(
					"  "+colors.Bold("STATUS").Underline().String(),
					colors.Bold("ENV KEY").Underline(),
					colors.Bold("HASH").Underline(),
				)
				// mit.AddRow()
			}
			mi++
			mit.AddRow("👻 Missing", k.Name, strmax(k.Hash, 12, "..."))
		} else if strings.Contains(k.Comment, "extra") {
			if mr == 0 {

				mrt.AddRow(
					"  "+colors.Bold("STATUS").Underline().String(),
					colors.Bold("ENV KEY").Underline(),
					colors.Bold("HASH").Underline(),
				)
				// mrt.AddRow()
			}
			mr++
			mrt.AddRow("🌱 Added", k.Name, strmax(k.Hash, 12, "..."))
		} else if strings.Contains(k.Comment, "different") {
			if dv == 0 {
				dvt.AddRow(
					"  "+colors.Bold("STATUS").Underline().String(),
					colors.Bold("ENV KEY").Underline(),
					colors.Bold("HASH").Underline(),
				)
				// dvt.AddRow()
			}
			dv++
			dvt.AddRow("❓ Different", k.Name, strmax(k.Hash, 12, "..."))
		}
	}

	fmt.Printf("• %s: %s\n", colors.Bold("source"), source)
	fmt.Println(mit)

	fmt.Printf("\n\n• %s: %s\n", colors.Bold("target"), target)
	fmt.Println(mrt)

	fmt.Printf("\n\n• %s\n", colors.Bold("different values"))
	fmt.Println(dvt)

	fmt.Println("")

	fmt.Printf(
		"\n👻 Missing in %s (%d) | 🌱 Missing in %s (%d) | ❓ Different values (%d)\n\n",
		colors.Bold("source"),
		colors.Red(mr).Bold(),
		colors.Bold("target"),
		colors.Red(mi).Bold(),
		colors.Yellow(dv).Bold(),
	)
}

func strmax(str string, l int, suffix string) string {
	if len(str) <= l {
		return str
	}
	return str[:l] + suffix
}
