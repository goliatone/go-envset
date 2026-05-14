package restart

import (
	"math"

	"github.com/urfave/cli/v2"
)

const (
	restartFlag     = "restart"
	foreverFlag     = "forever"
	maxRestartsFlag = "max-restarts"
	maxRestartAlias = "max-restart"
)

// Options resolves restart flags across command contexts.
//
// Environment commands define local flags with the same names as the app-level
// flags. urfave/cli resolves duplicate names from child to parent, so a global
// explicit value like `envset --restart=false development` would otherwise be
// hidden by the development command default.
func Options(c *cli.Context) (bool, int) {
	restart := boolValue(c, restartFlag)
	max := intValue(c, maxRestartsFlag, maxRestartAlias)
	forever := boolValue(c, foreverFlag)

	restartExplicit, _ := explicitBool(c, restartFlag)
	foreverExplicit, _ := explicitBool(c, foreverFlag)

	if !forever {
		return restart, max
	}

	if restartExplicit && !restart {
		return false, max
	}

	if foreverExplicit {
		restart = true
	}

	if !restart {
		return false, max
	}

	return restart, math.MaxInt
}

func boolValue(c *cli.Context, name string) bool {
	if ok, value := explicitBool(c, name); ok {
		return value
	}
	return c.Bool(name)
}

func explicitBool(c *cli.Context, name string) (bool, bool) {
	for _, ctx := range c.Lineage() {
		if !hasLocalFlag(ctx, name) {
			continue
		}
		return true, ctx.Bool(name)
	}
	return false, false
}

func intValue(c *cli.Context, name string, aliases ...string) int {
	names := append([]string{name}, aliases...)
	for _, ctx := range c.Lineage() {
		if !hasLocalFlag(ctx, names...) {
			continue
		}
		return ctx.Int(name)
	}
	return c.Int(name)
}

func hasLocalFlag(c *cli.Context, names ...string) bool {
	if c == nil {
		return false
	}

	for _, local := range c.LocalFlagNames() {
		for _, name := range names {
			if local == name {
				return true
			}
		}
	}

	return false
}
