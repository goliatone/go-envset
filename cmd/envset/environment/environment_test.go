package environment

import (
	"fmt"
	"math"
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

func TestRestartOptionsPrecedence(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		restartDefault bool
		foreverDefault bool
		wantRestart    bool
		wantMax        int
	}{
		{
			name:           "explicit restart false wins over config forever",
			args:           []string{"--restart=false"},
			restartDefault: true,
			foreverDefault: true,
			wantRestart:    false,
			wantMax:        3,
		},
		{
			name:           "explicit forever enables restart over disabled config",
			args:           []string{"--forever"},
			restartDefault: false,
			foreverDefault: false,
			wantRestart:    true,
			wantMax:        math.MaxInt,
		},
		{
			name:           "config forever only applies when restart is enabled",
			args:           nil,
			restartDefault: false,
			foreverDefault: true,
			wantRestart:    false,
			wantMax:        3,
		},
		{
			name:           "config restart forever sets max restarts",
			args:           nil,
			restartDefault: true,
			foreverDefault: true,
			wantRestart:    true,
			wantMax:        math.MaxInt,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotRestart bool
			var gotMax int

			app := cli.NewApp()
			app.Flags = []cli.Flag{
				&cli.BoolFlag{Name: "restart", Value: tt.restartDefault},
				&cli.BoolFlag{Name: "forever", Value: tt.foreverDefault},
				&cli.IntFlag{Name: "max-restarts", Value: 3},
			}
			app.Action = func(c *cli.Context) error {
				gotRestart, gotMax = restartOptions(c)
				return nil
			}

			if err := app.Run(append([]string{"envset"}, tt.args...)); err != nil {
				t.Fatalf("run app: %v", err)
			}

			if gotRestart != tt.wantRestart {
				t.Fatalf("restart = %v, want %v", gotRestart, tt.wantRestart)
			}
			if gotMax != tt.wantMax {
				t.Fatalf("max = %d, want %d", gotMax, tt.wantMax)
			}
		})
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
