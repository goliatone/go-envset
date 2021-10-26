package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/goliatone/go-envset/cmd/envset/environment"
	"github.com/goliatone/go-envset/cmd/envset/metadata"
	"github.com/goliatone/go-envset/cmd/envset/template"
	"github.com/goliatone/go-envset/pkg/config"
	"github.com/goliatone/go-envset/pkg/exec"

	"github.com/goliatone/go-envset/pkg/envset"

	build "github.com/goliatone/go-envset/pkg/version"

	"github.com/urfave/cli/v2"
)

var app *cli.App
var cnf *config.Config

func init() {
	cli.VersionFlag = &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{"v"},
		Usage:   "prints the application version",
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
	args := exec.CliArgs(os.Args)
	cmd := exec.CmdFromArgs(os.Args)
	run(args, cmd)
}

func run(args []string, ecmd exec.ExecCmd) {
	cnf, err := config.Load(".envsetrc")
	if err != nil {
		log.Println("Error loading configuration:", err)
		log.Panic("Ensure you have a valid .envsetrc")
	}

	subcommands := []*cli.Command{}

	for _, env := range cnf.Environments.Name {
		subcommands = append(subcommands, environment.GetCommand(env, ecmd, cnf))
	}

	appendCommand(metadata.GetCommand(cnf))

	appendCommand(template.GetCommand(cnf))

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
			Cmd:      ecmd.Cmd,
			Args:     ecmd.Args,
			Expand:   c.Bool("expand"),
			Isolated: c.Bool("isolated"),
			Filename: c.String("env-file"),
		}
		//Run if we have something like this:
		// envset --env-file=.env -- node index.js
		// envset --env-file=.envset --env=development -- node index.js
		if ecmd.Cmd != "" && o.Filename != cnf.Filename {
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
