package version

import (
	"github.com/goliatone/go-envset/pkg/config"
	"github.com/goliatone/go-envset/pkg/version"
	"github.com/urfave/cli/v2"
)

//GetCommand returns a new cli.Command for the
//version command
func GetCommand(cnf *config.Config) *cli.Command {
	return &cli.Command{
		Name:  "version",
		Usage: "return version information for envset program",
		Action: func(ctx *cli.Context) error {
			return version.Print(ctx.App.Writer)
		},
	}
}
