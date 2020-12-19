package main

import (
	"fmt"
	"goliatone/go-envset/pkg/config"
	"goliatone/go-envset/pkg/envset"
	"goliatone/go-envset/pkg/version"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/urfave/cli/v2"
)

var cnf *config.Config

func init() {
	cnf, _ = config.Load(".envsetrc")
}

func main() {
	app := &cli.App{
		Name:     "envset",
		Version:  version.BuildVersion,
		Compiled: time.Now(),
		Authors: []*cli.Author{
			&cli.Author{
				Name:  "Goliat One",
				Email: "hi@goliat.one",
			},
		},
		Copyright: "(c) 2020 Goliatone",
		Usage:     "Load environment variables to your shell and run a command",
		HelpName:  "envset",
		UsageText: "envset [environment] -- [command]\n\nEXAMPLE:\n\t envset development -- node index.js\n\t eval $(envset development --isolated=true)\n\t envset development -- say '${MY_GREETING}'",
	}

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
	app.Commands = append(app.Commands, &cli.Command{
		Name: "metadata",
		Usage: "generate a metadata file from environment file",
		Description: "creates a metadata file with all the given environments",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "print", Usage: "only print the contents to stdout, don't write file"},
			&cli.StringFlag{Name: "filename", Usage: "metadata file name", Value: "metadata.json"},
			&cli.StringFlag{Name: "filepath", Usage: "template file path", Value: "./.envmeta"},
			&cli.StringFlag{Name: "env-file", Value: ".envset", Usage: "load environment from `FILE`"},
			&cli.BoolFlag{Name: "overwrite", Usage: "overwrite template, this will delete any changes"},
			&cli.BoolFlag{Name: "hash", Usage: "only encode the hash, skip values"},
		},
		Action: func(c *cli.Context) error {
			print := c.Bool("print")
			envfile := c.String("env-file")
			filename := c.String("filename")
			dir := c.String("filepath")
			overwrite := c.Bool("overwrite")
			hash := c.Bool("hash")

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
			filename = filepath.Join(dir, filename)

			return envset.CreateMetadataFile(envfile, filename, overwrite, print, hash)
		},
	})
	app.Commands = append(app.Commands, &cli.Command{
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
			&cli.BoolFlag{Name: "overwrite", Usage: "overwrite template, this will delete any changes"},
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

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
