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
		Copyright: "(c) 2021 Goliatone",
		Usage:     "Load environment variables to your shell and run a command",
		HelpName:  "envset",
		UsageText: "envset [environment] -- [command]\n\nEXAMPLE:\n\t envset development -- node index.js\n\t eval $(envset development --isolated=true)\n\t envset development -- say '${MY_GREETING}'",
	}
}

func main() {
	a := cliArgs(os.Args)
	c := cmdFromArgs(os.Args)
	run(a, c)
}

func run(args []string, exec execCmd) {
	cnf, err := config.Load(".envsetrc")
	if err != nil {
		log.Println("Error loading configuration:", err)
		log.Panic("Ensure you have a valid .envsetrc")
	}

	subcommands := []*cli.Command{}

	for _, env := range cnf.Environments.Name {
		subcommands = append(subcommands, &cli.Command{
			Name:        env,
			Usage:       fmt.Sprintf("load \"%s\" environment in current shell session", env),
			UsageText:   fmt.Sprintf("envset %s [options] -- [command] [arguments...]", env),
			Description: "This will load the environment and execute the provided command",
			Category:    "ENVIRONMENTS",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "isolated",
					Usage: "if true we run shell with only variables defined",
					Value: cnf.Isolated, //call with --isolated=false to show all
				},
				&cli.BoolFlag{
					Name:  "expand",
					Usage: "if true we use expand environment variables",
					Value: cnf.Expand,
				},
				&cli.StringFlag{
					Name:  "env-file",
					Usage: "file name with environment definition",
					Value: cnf.Filename,
				},
				&cli.StringSliceFlag{
					Name:    "required",
					Aliases: []string{"R"},
					Usage:   "list of key names that are required to run",
				},
				&cli.StringFlag{
					Name:    "export-env-name",
					Aliases: []string{"N"},
					Usage:   "name of exported variable with current environment name",
					Value:   cnf.ExportEnvName,
				},
			},
			Action: func(c *cli.Context) error {
				//TODO: we want to support .env.local => [local]

				env := c.Command.Name

				ro := envset.RunOptions{
					Cmd:           exec.Cmd,
					Args:          exec.Args,
					Isolated:      c.Bool("isolated"),
					Expand:        c.Bool("expand"),
					Required:      c.StringSlice("required"),
					Filename:      c.String("env-file"),
					ExportEnvName: c.String("export-env-name"),
				}

				if c.NArg() == 0 {
					//we can do: eval `envset development`
					//we can do: envset development > /tmp/env1 | source
					//https://stackoverflow.com/questions/36074851/persist-the-value-set-for-an-env-variable-on-the-shell-after-the-go-program-exit
					//TODO: Pass required so we show missing ones?
					return envset.Print(env, ro.Filename, ro.Isolated, ro.Expand)
				}

				return envset.Run(env, ro)
			},
		})
	}

	appendCommand(&cli.Command{
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

					//TODO: we want to return a KeysDiff
					//where we have an annotation of what was wrong:
					//src missing, src extra, different hash
					s3 := envset.CompareSections(*s1, *s2)

					if s3.IsEmpty() == false {
						fmt.Println("")
						if print {
							for _, k := range s3.Keys {
								fmt.Printf("key %s: %s\n", k.Name, k.Comment)
							}
						}
						//Exit with error e.g. to fail CI
						return cli.Exit("Metadata test failed!", 1)
					}

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
	})

	app.Commands = append(app.Commands, subcommands...)

	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:  "env",
			Usage: "env name matching a section. If not set matches env vars in global scope",
			Value: envset.DefaultSection,
		},
		&cli.StringFlag{
			//This can be an absolute path. If a file name then we recursively look up
			Name:  "env-file",
			Usage: "`file` with environment definition",
			Value: cnf.Filename,
		},
		&cli.BoolFlag{
			Name:  "isolated",
			Usage: "if false the environment inherits the shell's environment",
			Value: cnf.Isolated, //call with --isolated=false to show all
		},
		&cli.BoolFlag{
			Name:  "expand",
			Usage: "if true we expand environment variables",
			Value: cnf.Expand,
		},
		&cli.StringSliceFlag{
			Name:    "required",
			Aliases: []string{"R"},
			Usage:   "list of key names that are required to run",
		},
	}

	app.Action = func(c *cli.Context) error {

		//How do we end up here?
		//If we try to execute a command for an inexistent environment, e.g:
		// envset ==> show help
		//we just called: envset
		if c.NArg() == 0 && c.NumFlags() == 0 {
			cli.ShowAppHelpAndExit(c, 0)
		}

		env := c.String("env")

		o := envset.RunOptions{
			Cmd:      exec.Cmd,
			Args:     exec.Args,
			Expand:   c.Bool("expand"),
			Isolated: c.Bool("isolated"),
			Filename: c.String("env-file"),
		}
		//Run if we have something like this:
		// envset --env-file=.env -- node index.js
		// envset --env-file=.envset --env=development -- node index.js
		if exec.Cmd != "" && o.Filename != cnf.Filename {
			return envset.Run(env, o)
		}

		// envset undefined ==> show error: environment undefined does not exist
		// envset undefined -- node index.js ==> show error: environment undefined does not exist
		if c.NArg() >= 1 && !c.Command.HasName(c.Args().First()) {
			return cli.Exit(fmt.Sprintf("%s: not a valid environment name", c.Args().First()), 1)
		}

		//we called something like:
		//envset --env-file=.env
		//envset --env-file=.envset --env=development
		return envset.Print(env, o.Filename, o.Isolated, o.Expand)
	}

	//TODO: we should process args to remove executable context
	//and return the arguments that are only for envset
	err = app.Run(args)
	if err != nil {
		log.Fatal(err)
	}
}

func appendCommand(command *cli.Command) {
	app.Commands = append(app.Commands, command)
}

type execCmd struct {
	Cmd  string
	Args []string
}

func cmdFromArgs(args []string) execCmd {
	cmd := ""
	idx := 0
	a := make([]string, 0)

	for i, v := range args {
		if v == "--" {
			idx = i + 1
			break
		}
	}

	if idx > 0 && len(args) >= idx {
		a = args[idx:]
	}

	return execCmd{
		Cmd:  cmd,
		Args: a,
	}
}

func cliArgs(args []string) []string {
	o := make([]string, 0)
	for _, v := range args {
		if v == "--" {
			break
		}
		o = append(o, v)
	}

	return o
}
