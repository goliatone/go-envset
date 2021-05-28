package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/goliatone/go-envset/pkg/config"

	"github.com/goliatone/go-envset/pkg/envset"

	build "github.com/goliatone/go-envset/pkg/version"

	"github.com/tcnksm/go-gitconfig"
	"github.com/urfave/cli/v2"
)

var app *cli.App
var cnf *config.Config

func init() {
	cnf, _ = config.Load(".envsetrc")

	cli.VersionFlag = &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{"V"},
		Usage:   "print the application version",
	}

	app = &cli.App{
		Name:     "envset",
		Version:  build.Tag,
		Compiled: time.Now(),
		Authors: []*cli.Author{
			{
				Name:  "Goliat One",
				Email: "hi@goliat.one",
			},
		},
		Copyright: "(c) 2020 Goliatone",
		Usage:     "Load environment variables to your shell and run a command",
		HelpName:  "envset",
		UsageText: "envset [environment] -- [command]\n\nEXAMPLE:\n\t envset development -- node index.js\n\t eval $(envset development --isolated=true)\n\t envset development -- say '${MY_GREETING}'",
	}
}

func main() {

	run(os.Args)
}

func run(args []string) {

	subcommands := []*cli.Command{}

	for _, env := range cnf.Environments.Name {
		subcommands = append(subcommands, &cli.Command{
			Name:        env,
			Usage:       fmt.Sprintf("load \"%s\" environment in current shell session", env),
			UsageText:   fmt.Sprintf("envset %s [options] -- [command] [arguments...]", env),
			Description: "This will load the environment and execute the provided command",
			// ArgsUsage:   "[arrgh]",
			Category: "ENVIRONMENTS",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "isolated",
					Usage: "if true we run shell with only variables defined",
					Value: true, //call with --isolated=false to show all
				},
				&cli.BoolFlag{
					Name:  "expand",
					Usage: "if true we use expand environment variables",
					Value: true, //call with --isolated=false to show all
				},
				&cli.StringFlag{
					Name:  "env-file",
					Usage: "file name with environment definition",
					Value: ".envset",
				},
				&cli.StringSliceFlag{
					Name:    "required",
					Aliases: []string{"R"},
					Usage:   "list of key names that are required to run",
				},
			},
			Action: func(c *cli.Context) error {
				expand := c.Bool("expand")
				isolated := c.Bool("isolated")
				//TODO: we want to support .env.local => [local]
				filename := c.String("env-file")
				env := c.Command.Name

				if c.NArg() == 0 {
					//we can do: eval `envset development`
					//we can do: envset development > /tmp/env1 | source
					//https://stackoverflow.com/questions/36074851/persist-the-value-set-for-an-env-variable-on-the-shell-after-the-go-program-exit
					//TODO: Pass required so we show missing ones?
					return envset.Print(env, filename, isolated, expand)
				}

				cmd := c.Args().First()
				arg := c.Args().Slice()[1:]

				//TODO: Get from config
				required := c.StringSlice("required")
				return envset.Run(env, filename, cmd, arg, isolated, expand, required)
			},
		})
	}

	appendCommand(&cli.Command{
		Name:        "metadata",
		Usage:       "generate a metadata file from environment file",
		Description: "creates a metadata file with all the given environments",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "print", Usage: "only print the contents to stdout, don't write file"},
			&cli.StringFlag{Name: "filename", Usage: "metadata file name", Value: "metadata.json"},
			&cli.StringFlag{Name: "filepath", Usage: "metadata file path", Value: "./.envmeta"},
			&cli.StringFlag{Name: "env-file", Value: ".envset", Usage: "load environment from `FILE`"},
			&cli.BoolFlag{Name: "overwrite", Usage: "set true to prevent overwrite metadata file", Value: true},
			&cli.BoolFlag{Name: "values", Usage: "add flag to show values in the output"},
			&cli.BoolFlag{Name: "globals", Usage: "include global section", Value: false},
			&cli.StringFlag{Name: "secret", Usage: "secret used to encode hash values", EnvVars: []string{"ENVSET_HASH_SECRET"}},
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
				Name: "compare",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name: "command2flag",
						// ...
					},
				},
				Action: func(c *cli.Context) error {
					env := envset.EnvFile{}
					env.FromJSON("./.envmeta/metadata.json")
					str, _ := env.ToJSON()
					fmt.Println(str)
					return nil
				},
			},
		},
	})

	appendCommand(&cli.Command{
		//TODO: This actually should load a template file and resolve it using the context.
		//Default template should generate envset.example
		Name:        "template",
		Usage:       "make a template file from an environment",
		Description: "create a new template or update file to document the variables in your environment",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "print", Usage: "only print the contents to stdout, don't write file"},
			&cli.StringFlag{Name: "filename", Usage: "template file name", Value: "envset.example"},
			&cli.StringFlag{Name: "filepath", Usage: "template file path", Value: "."},
			&cli.StringFlag{Name: "env-file", Value: ".envset", Usage: "load environment from `FILE`"},
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
	})

	app.Commands = append(app.Commands, subcommands...)

	app.Flags = []cli.Flag{
		&cli.StringFlag{
			//This can be an absolute path. If a file name then we recursively look up
			Name:  "env-file",
			Usage: "file with environment definition",
		},
		&cli.BoolFlag{
			Name:  "isolated",
			Usage: "if true we run shell with only variables defined",
			Value: true, //call with --isolated=false to show all
		},
		&cli.BoolFlag{
			Name:  "expand",
			Usage: "if true we expand environment variables defined",
			Value: true, //call with --isolated=false to show all
		},
		&cli.StringSliceFlag{
			Name:    "required",
			Aliases: []string{"R"},
			Usage:   "list of key names that are required to run",
		},
	}

	app.Action = func(c *cli.Context) error {
		expand := c.Bool("expand")
		filename := c.String("env-file")
		isolated := c.Bool("isolated")

		if filename == "" {
			cli.ShowAppHelpAndExit(c, 0)
		}

		env := envset.DefaultSection
		if c.NArg() == 0 {
			return envset.Print(env, filename, isolated, expand)
		}

		cmd := c.Args().First()
		arg := c.Args().Slice()[1:]

		required := c.StringSlice("required")
		return envset.Run(env, filename, cmd, arg, isolated, expand, required)
	}

	err := app.Run(args)
	if err != nil {
		log.Fatal(err)
	}
}

func appendCommand(command *cli.Command) {
	app.Commands = append(app.Commands, command)
}
