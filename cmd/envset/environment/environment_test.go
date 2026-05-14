package environment

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/goliatone/go-envset/pkg/config"
	"github.com/goliatone/go-envset/pkg/exec"
	"github.com/urfave/cli/v2"
)

func TestRestartFlagFalseOverridesConfigForever(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".envset"), []byte("[development]\nEXPECTED=value\n"), 0644); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get wd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(previousDir); err != nil {
			t.Fatalf("restore wd: %v", err)
		}
	})

	countFile := filepath.Join(dir, "runs")
	cnf := testConfig()
	cnf.Restart = true
	cnf.RestartForever = true
	cnf.ExcludeFromRestart = []string{"development"}

	app := cli.NewApp()
	app.ExitErrHandler = func(_ *cli.Context, _ error) {}
	app.Commands = []*cli.Command{
		GetCommand("development", exec.ExecCmd{
			Cmd: "sh",
			Args: []string{
				"-c",
				fmt.Sprintf("if [ -f %q ]; then printf x >> %q; exit 0; fi; printf x >> %q; exit 1", countFile, countFile, countFile),
			},
		}, cnf),
	}

	err = app.Run([]string{"envset", "development", "--restart=false"})
	if err == nil {
		t.Fatal("expected command failure")
	}

	content, err := os.ReadFile(countFile)
	if err != nil {
		t.Fatalf("read run count: %v", err)
	}
	if got := len(content); got != 1 {
		t.Fatalf("expected command to run once, got %d runs", got)
	}
}

func TestGlobalRestartFlagFalseOverridesEnvironmentDefault(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".envset"), []byte("[development]\nEXPECTED=value\n"), 0644); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get wd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(previousDir); err != nil {
			t.Fatalf("restore wd: %v", err)
		}
	})

	countFile := filepath.Join(dir, "runs")
	cnf := testConfig()
	cnf.Restart = true
	cnf.RestartForever = false

	app := cli.NewApp()
	app.ExitErrHandler = func(_ *cli.Context, _ error) {}
	app.Flags = []cli.Flag{
		&cli.BoolFlag{Name: "restart", Value: cnf.Restart},
		&cli.BoolFlag{Name: "forever", Value: cnf.RestartForever},
		&cli.IntFlag{Name: "max-restarts", Aliases: []string{"max-restart"}, Value: cnf.MaxRestarts},
	}
	app.Commands = []*cli.Command{
		GetCommand("development", exec.ExecCmd{
			Cmd:  "sh",
			Args: []string{"-c", fmt.Sprintf("printf x >> %q; exit 1", countFile)},
		}, cnf),
	}

	err = app.Run([]string{"envset", "--restart=false", "development"})
	if err == nil {
		t.Fatal("expected command failure")
	}

	content, err := os.ReadFile(countFile)
	if err != nil {
		t.Fatalf("read run count: %v", err)
	}
	if got := len(content); got != 1 {
		t.Fatalf("expected command to run once, got %d runs", got)
	}
}

func testConfig() *config.Config {
	return &config.Config{
		Filename:            ".envset",
		Environments:        &config.Environments{Names: []string{"development"}},
		CommentSectionNames: &config.CommentSectionNames{},
		Expand:              true,
		Isolated:            true,
		ExportEnvName:       "APP_ENV",
		Restart:             true,
		MaxRestarts:         3,
	}
}
