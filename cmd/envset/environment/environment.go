package environment

import (
	"fmt"

	"github.com/goliatone/go-envset/pkg/config"
	"github.com/goliatone/go-envset/pkg/envset"
	"github.com/goliatone/go-envset/pkg/exec"
	"github.com/urfave/cli/v2"
)

//GetCommand export command
func GetCommand(env string, ecmd exec.ExecCmd, cnf *config.Config) *cli.Command {
	return &cli.Command{
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
			&cli.StringSliceFlag{
				Name:    "inherit",
				Aliases: []string{"I"},
				Usage:   "list of env vars to inherit from shell",
			},
		},
		Action: func(c *cli.Context) error {
			//TODO: we want to support .env.local => [local]

			env := c.Command.Name

			ro := envset.RunOptions{
				Cmd:           ecmd.Cmd,
				Args:          ecmd.Args,
				Isolated:      c.Bool("isolated"),
				Expand:        c.Bool("expand"),
				Required:      c.StringSlice("required"),
				Inherit:       c.StringSlice("inherit"),
				Filename:      c.String("env-file"),
				ExportEnvName: c.String("export-env-name"),
			}

			if ecmd.Cmd == "" {
				return envset.Print(env, ro.Filename, ro.Isolated, ro.Expand)
			}

			return envset.Run(env, ro)
		},
	}
}
