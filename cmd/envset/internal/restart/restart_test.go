package restart

import (
	"math"
	"testing"

	"github.com/urfave/cli/v2"
)

func TestOptionsPrecedence(t *testing.T) {
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
				&cli.BoolFlag{Name: restartFlag, Value: tt.restartDefault},
				&cli.BoolFlag{Name: foreverFlag, Value: tt.foreverDefault},
				&cli.IntFlag{Name: maxRestartsFlag, Value: 3},
			}
			app.Action = func(c *cli.Context) error {
				gotRestart, gotMax = Options(c)
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

func TestOptionsGlobalFlagsOverrideShadowedCommandDefaults(t *testing.T) {
	tests := []struct {
		name                string
		args                []string
		localRestartDefault bool
		wantRestart         bool
		wantMax             int
	}{
		{
			name:                "global restart false wins over local default true",
			args:                []string{"--restart=false", "development"},
			localRestartDefault: true,
			wantRestart:         false,
			wantMax:             3,
		},
		{
			name:                "local restart wins over global restart",
			args:                []string{"--restart=false", "development", "--restart=true"},
			localRestartDefault: true,
			wantRestart:         true,
			wantMax:             3,
		},
		{
			name:                "global max restarts wins over local default",
			args:                []string{"--max-restarts=1", "development"},
			localRestartDefault: true,
			wantRestart:         true,
			wantMax:             1,
		},
		{
			name:                "global max restart alias wins over local default",
			args:                []string{"--max-restart=2", "development"},
			localRestartDefault: true,
			wantRestart:         true,
			wantMax:             2,
		},
		{
			name:                "global forever wins over local restart default false",
			args:                []string{"--forever", "development"},
			localRestartDefault: false,
			wantRestart:         true,
			wantMax:             math.MaxInt,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotRestart bool
			var gotMax int

			app := cli.NewApp()
			app.Flags = []cli.Flag{
				&cli.BoolFlag{Name: restartFlag, Value: true},
				&cli.BoolFlag{Name: foreverFlag},
				&cli.IntFlag{Name: maxRestartsFlag, Aliases: []string{maxRestartAlias}, Value: 3},
			}
			app.Commands = []*cli.Command{
				{
					Name: "development",
					Flags: []cli.Flag{
						&cli.BoolFlag{Name: restartFlag, Value: tt.localRestartDefault},
						&cli.BoolFlag{Name: foreverFlag},
						&cli.IntFlag{Name: maxRestartsFlag, Aliases: []string{maxRestartAlias}, Value: 3},
					},
					Action: func(c *cli.Context) error {
						gotRestart, gotMax = Options(c)
						return nil
					},
				},
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
