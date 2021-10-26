package exec

type ExecCmd struct {
	Cmd  string
	Args []string
}

func CmdFromArgs(args []string) ExecCmd {
	cmd := ""
	idx := 0
	a := make([]string, 0)

	for i, v := range args {
		if v == "--" {
			idx = i + 1
			break
		}
	}

	if idx > 0 && len(args) >= idx {
		cmd = args[idx]
		a = args[idx+1:]
	}

	return ExecCmd{
		Cmd:  cmd,
		Args: a,
	}
}

func CliArgs(args []string) []string {
	o := make([]string, 0)
	for _, v := range args {
		if v == "--" {
			break
		}
		o = append(o, v)
	}

	return o
}
