package exec

import "testing"

func TestCmdFromArgsTrailingSeparator(t *testing.T) {
	cmd := CmdFromArgs([]string{"envset", "--"})

	if cmd.Cmd != "" {
		t.Fatalf("cmd = %q, want empty", cmd.Cmd)
	}
	if len(cmd.Args) != 0 {
		t.Fatalf("args = %v, want empty", cmd.Args)
	}
}

func TestCmdFromArgsWithCommand(t *testing.T) {
	cmd := CmdFromArgs([]string{"envset", "--", "sh", "-c", "true"})

	if cmd.Cmd != "sh" {
		t.Fatalf("cmd = %q, want sh", cmd.Cmd)
	}
	if len(cmd.Args) != 2 || cmd.Args[0] != "-c" || cmd.Args[1] != "true" {
		t.Fatalf("args = %v, want [-c true]", cmd.Args)
	}
}

func TestCliArgsStripsCommandContext(t *testing.T) {
	args := CliArgs([]string{"envset", "--env", "development", "--", "sh", "-c", "true"})

	want := []string{"envset", "--env", "development"}
	if len(args) != len(want) {
		t.Fatalf("args = %v, want %v", args, want)
	}
	for i := range want {
		if args[i] != want[i] {
			t.Fatalf("args = %v, want %v", args, want)
		}
	}
}
