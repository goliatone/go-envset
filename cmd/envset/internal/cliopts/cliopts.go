// Package cliopts resolves CLI options across app and command contexts.
package cliopts

import (
	"math"
	"slices"

	"github.com/goliatone/go-envset/pkg/config"
	"github.com/goliatone/go-envset/pkg/envset"
	"github.com/goliatone/go-envset/pkg/exec"
	"github.com/urfave/cli/v2"
)

const (
	// Run flag names shared by the app and environment commands.
	EnvFileFlag        = "env-file"
	IsolatedFlag       = "isolated"
	ExpandFlag         = "expand"
	RequiredFlag       = "required"
	RequiredAlias      = "R"
	ExportEnvNameFlag  = "export-env-name"
	ExportEnvNameAlias = "N"
	InheritFlag        = "inherit"
	InheritAlias       = "I"
	RestartFlag        = "restart"
	ForeverFlag        = "forever"
	MaxRestartsFlag    = "max-restarts"
	MaxRestartAlias    = "max-restart"
)

// RunOptions resolves command flags across local and parent cli contexts.
//
// Environment commands intentionally duplicate the app-level run flags so users
// can place flags before or after the environment name. urfave/cli resolves
// duplicate flag names from child to parent, so direct c.Bool/c.String calls on
// environment commands hide explicit global values behind local defaults.
func RunOptions(c *cli.Context, cnf *config.Config, env string, ecmd exec.ExecCmd) envset.RunOptions {
	required := StringSlice(c, RequiredFlag, RequiredAlias)
	required = cnf.MergeRequired(env, required)

	restart, maxRestarts := RestartOptions(c)

	return envset.RunOptions{
		Cmd:                 ecmd.Cmd,
		Args:                ecmd.Args,
		Isolated:            Bool(c, IsolatedFlag),
		Expand:              Bool(c, ExpandFlag),
		Filename:            String(c, EnvFileFlag),
		CommentSectionNames: cnf.CommentSectionNames.Keys,
		Required:            required,
		Inherit:             StringSlice(c, InheritFlag, InheritAlias),
		ExportEnvName:       String(c, ExportEnvNameFlag, ExportEnvNameAlias),
		Restart:             restart,
		MaxRestarts:         maxRestarts,
	}
}

// RestartOptions resolves restart behavior from duplicated restart flags.
func RestartOptions(c *cli.Context) (bool, int) {
	restart := Bool(c, RestartFlag)
	maxRestarts := Int(c, MaxRestartsFlag, MaxRestartAlias)
	forever := Bool(c, ForeverFlag)

	restartExplicit, _ := explicitBool(c, RestartFlag)
	foreverExplicit, _ := explicitBool(c, ForeverFlag)

	if !forever {
		return restart, maxRestarts
	}

	if restartExplicit && !restart {
		return false, maxRestarts
	}

	if foreverExplicit {
		restart = true
	}

	if !restart {
		return false, maxRestarts
	}

	return restart, math.MaxInt
}

// Bool resolves a bool flag, preferring explicit flags from child to parent.
func Bool(c *cli.Context, name string, aliases ...string) bool {
	if ok, value := explicitBool(c, name, aliases...); ok {
		return value
	}
	return c.Bool(name)
}

// String resolves a string flag, preferring explicit flags from child to parent.
func String(c *cli.Context, name string, aliases ...string) string {
	names := append([]string{name}, aliases...)
	for _, ctx := range c.Lineage() {
		if !hasLocalFlag(ctx, names...) {
			continue
		}
		return ctx.String(name)
	}
	return c.String(name)
}

// StringSlice resolves a string-slice flag, preferring explicit flags from child to parent.
func StringSlice(c *cli.Context, name string, aliases ...string) []string {
	names := append([]string{name}, aliases...)
	for _, ctx := range c.Lineage() {
		if !hasLocalFlag(ctx, names...) {
			continue
		}
		return ctx.StringSlice(name)
	}
	return c.StringSlice(name)
}

// Int resolves an int flag, preferring explicit flags from child to parent.
func Int(c *cli.Context, name string, aliases ...string) int {
	names := append([]string{name}, aliases...)
	for _, ctx := range c.Lineage() {
		if !hasLocalFlag(ctx, names...) {
			continue
		}
		return ctx.Int(name)
	}
	return c.Int(name)
}

func explicitBool(c *cli.Context, name string, aliases ...string) (bool, bool) {
	names := append([]string{name}, aliases...)
	for _, ctx := range c.Lineage() {
		if !hasLocalFlag(ctx, names...) {
			continue
		}
		return true, ctx.Bool(name)
	}
	return false, false
}

func hasLocalFlag(c *cli.Context, names ...string) bool {
	if c == nil {
		return false
	}

	for _, local := range c.LocalFlagNames() {
		if slices.Contains(names, local) {
			return true
		}
	}

	return false
}
