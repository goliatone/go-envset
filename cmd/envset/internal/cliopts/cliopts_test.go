package cliopts

import (
	"math"
	"reflect"
	"testing"

	"github.com/goliatone/go-envset/pkg/config"
	"github.com/goliatone/go-envset/pkg/envset"
	"github.com/goliatone/go-envset/pkg/exec"
	"github.com/urfave/cli/v2"
)

func TestRunOptionsGlobalFlagsOverrideCommandDefaults(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		validate func(t *testing.T, got resolvedOptions)
	}{
		{
			name: "env file",
			args: []string{"--env-file=global.envset", "development"},
			validate: func(t *testing.T, got resolvedOptions) {
				assertEqual(t, got.run.Filename, "global.envset")
			},
		},
		{
			name: "isolated",
			args: []string{"--isolated=false", "development"},
			validate: func(t *testing.T, got resolvedOptions) {
				assertEqual(t, got.run.Isolated, false)
			},
		},
		{
			name: "expand",
			args: []string{"--expand=false", "development"},
			validate: func(t *testing.T, got resolvedOptions) {
				assertEqual(t, got.run.Expand, false)
			},
		},
		{
			name: "required",
			args: []string{"--required=GLOBAL_REQUIRED", "development"},
			validate: func(t *testing.T, got resolvedOptions) {
				assertDeepEqual(t, got.run.Required, []string{"GLOBAL_REQUIRED"})
			},
		},
		{
			name: "required alias",
			args: []string{"-R=GLOBAL_REQUIRED", "development"},
			validate: func(t *testing.T, got resolvedOptions) {
				assertDeepEqual(t, got.run.Required, []string{"GLOBAL_REQUIRED"})
			},
		},
		{
			name: "export env name",
			args: []string{"--export-env-name=ENVSET_ENV", "development"},
			validate: func(t *testing.T, got resolvedOptions) {
				assertEqual(t, got.run.ExportEnvName, "ENVSET_ENV")
			},
		},
		{
			name: "export env name alias",
			args: []string{"-N=ENVSET_ENV", "development"},
			validate: func(t *testing.T, got resolvedOptions) {
				assertEqual(t, got.run.ExportEnvName, "ENVSET_ENV")
			},
		},
		{
			name: "inherit",
			args: []string{"--inherit=GLOBAL_INHERIT", "development"},
			validate: func(t *testing.T, got resolvedOptions) {
				assertDeepEqual(t, got.run.Inherit, []string{"GLOBAL_INHERIT"})
			},
		},
		{
			name: "inherit alias",
			args: []string{"-I=GLOBAL_INHERIT", "development"},
			validate: func(t *testing.T, got resolvedOptions) {
				assertDeepEqual(t, got.run.Inherit, []string{"GLOBAL_INHERIT"})
			},
		},
		{
			name: "restart false",
			args: []string{"--restart=false", "development"},
			validate: func(t *testing.T, got resolvedOptions) {
				assertEqual(t, got.run.Restart, false)
			},
		},
		{
			name: "forever",
			args: []string{"--forever", "development"},
			validate: func(t *testing.T, got resolvedOptions) {
				assertEqual(t, got.run.Restart, true)
				assertEqual(t, got.run.MaxRestarts, math.MaxInt)
			},
		},
		{
			name: "max restarts",
			args: []string{"--max-restarts=1", "development"},
			validate: func(t *testing.T, got resolvedOptions) {
				assertEqual(t, got.run.MaxRestarts, 1)
			},
		},
		{
			name: "max restart alias",
			args: []string{"--max-restart=2", "development"},
			validate: func(t *testing.T, got resolvedOptions) {
				assertEqual(t, got.run.MaxRestarts, 2)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runResolverApp(t, tt.args)
			tt.validate(t, got)
		})
	}
}

func TestRunOptionsLocalFlagsOverrideGlobalFlags(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		validate func(t *testing.T, got resolvedOptions)
	}{
		{
			name: "env file",
			args: []string{"--env-file=global.envset", "development", "--env-file=local.envset"},
			validate: func(t *testing.T, got resolvedOptions) {
				assertEqual(t, got.run.Filename, "local.envset")
			},
		},
		{
			name: "isolated",
			args: []string{"--isolated=false", "development", "--isolated=true"},
			validate: func(t *testing.T, got resolvedOptions) {
				assertEqual(t, got.run.Isolated, true)
			},
		},
		{
			name: "expand",
			args: []string{"--expand=false", "development", "--expand=true"},
			validate: func(t *testing.T, got resolvedOptions) {
				assertEqual(t, got.run.Expand, true)
			},
		},
		{
			name: "required",
			args: []string{"--required=GLOBAL_REQUIRED", "development", "--required=LOCAL_REQUIRED"},
			validate: func(t *testing.T, got resolvedOptions) {
				assertDeepEqual(t, got.run.Required, []string{"LOCAL_REQUIRED"})
			},
		},
		{
			name: "export env name",
			args: []string{"--export-env-name=GLOBAL_ENV", "development", "--export-env-name=LOCAL_ENV"},
			validate: func(t *testing.T, got resolvedOptions) {
				assertEqual(t, got.run.ExportEnvName, "LOCAL_ENV")
			},
		},
		{
			name: "inherit",
			args: []string{"--inherit=GLOBAL_INHERIT", "development", "--inherit=LOCAL_INHERIT"},
			validate: func(t *testing.T, got resolvedOptions) {
				assertDeepEqual(t, got.run.Inherit, []string{"LOCAL_INHERIT"})
			},
		},
		{
			name: "restart",
			args: []string{"--restart=false", "development", "--restart=true"},
			validate: func(t *testing.T, got resolvedOptions) {
				assertEqual(t, got.run.Restart, true)
			},
		},
		{
			name: "forever",
			args: []string{"--forever", "development", "--restart=false"},
			validate: func(t *testing.T, got resolvedOptions) {
				assertEqual(t, got.run.Restart, false)
				assertEqual(t, got.run.MaxRestarts, 3)
			},
		},
		{
			name: "max restarts",
			args: []string{"--max-restarts=1", "development", "--max-restarts=2"},
			validate: func(t *testing.T, got resolvedOptions) {
				assertEqual(t, got.run.MaxRestarts, 2)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runResolverApp(t, tt.args)
			tt.validate(t, got)
		})
	}
}

func TestRunOptionsDefaults(t *testing.T) {
	got := runResolverApp(t, []string{"development"})

	assertEqual(t, got.run.Filename, ".envset")
	assertEqual(t, got.run.Isolated, true)
	assertEqual(t, got.run.Expand, true)
	assertDeepEqual(t, got.run.Required, []string(nil))
	assertDeepEqual(t, got.run.Inherit, []string(nil))
	assertEqual(t, got.run.ExportEnvName, "APP_ENV")
	assertEqual(t, got.run.Restart, true)
	assertEqual(t, got.run.MaxRestarts, 3)
}

type resolvedOptions struct {
	run envset.RunOptions
}

func runResolverApp(t *testing.T, args []string) resolvedOptions {
	t.Helper()

	cnf := &config.Config{
		Filename:            ".envset",
		CommentSectionNames: &config.CommentSectionNames{},
		Required:            map[string][]string{},
		Expand:              true,
		Isolated:            true,
		ExportEnvName:       "APP_ENV",
		Restart:             true,
		RestartForever:      false,
		MaxRestarts:         3,
	}

	var got resolvedOptions

	app := cli.NewApp()
	app.Flags = testRunFlags(".envset", true, true, "APP_ENV", true, false, 3)
	app.Commands = []*cli.Command{
		{
			Name:  "development",
			Flags: testRunFlags(".envset", true, true, "APP_ENV", true, false, 3),
			Action: func(c *cli.Context) error {
				got.run = RunOptions(c, cnf, "development", exec.ExecCmd{Cmd: "sh"})
				return nil
			},
		},
	}

	if err := app.Run(append([]string{"envset"}, args...)); err != nil {
		t.Fatalf("run app: %v", err)
	}

	return got
}

func testRunFlags(filename string, isolated, expand bool, exportEnvName string, restart, forever bool, maxRestarts int) []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{Name: EnvFileFlag, Value: filename},
		&cli.BoolFlag{Name: IsolatedFlag, Value: isolated},
		&cli.BoolFlag{Name: ExpandFlag, Value: expand},
		&cli.StringSliceFlag{Name: RequiredFlag, Aliases: []string{RequiredAlias}},
		&cli.StringFlag{Name: ExportEnvNameFlag, Aliases: []string{ExportEnvNameAlias}, Value: exportEnvName},
		&cli.StringSliceFlag{Name: InheritFlag, Aliases: []string{InheritAlias}},
		&cli.BoolFlag{Name: RestartFlag, Value: restart},
		&cli.BoolFlag{Name: ForeverFlag, Value: forever},
		&cli.IntFlag{Name: MaxRestartsFlag, Aliases: []string{MaxRestartAlias}, Value: maxRestarts},
	}
}

func assertEqual[T comparable](t *testing.T, got, want T) {
	t.Helper()
	if got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func assertDeepEqual(t *testing.T, got, want interface{}) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}
