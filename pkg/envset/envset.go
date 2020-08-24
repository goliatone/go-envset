package envset

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"

	"gopkg.in/ini.v1"
)

//Run will run the given command after loading the environment
func Run(environment, name, cmd string, args []string, isolated bool) error {
	filename, err := FileFinder(name, 2)
	if err != nil {
		return err
	}

	//EnvFile.Load(filename)
	env, err := ini.Load(filename)

	if err != nil {
		return envFileErrorNotFound{err, "file not found"}
	}

	sec, err := env.GetSection(environment)
	if err != nil {
		return envSectionErrorNotFound{err, "section not found"}
	}

	vars := make([]string, 0)
	context := map[string]string{}

	for _, k := range sec.KeyStrings() {
		context[k] = sec.Key(k).String()
		vars = append(vars, fmt.Sprintf("%s=%s\n", k, sec.Key(k).String()))
	}

	//Replace ${VAR} in arguments
	for i, arg := range args {
		//we use custom interpolation string for variables we load
		args[i] = interpolate(arg, context)
		//try using built in OS variables
		if isolated == false {
			args[i] = os.ExpandEnv(args[i])
		}
	}

	command := exec.Command(cmd, args...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	if isolated {
		command.Env = vars
	} else {
		for k, v := range context {
			os.Setenv(k, v)
		}
	}

	return command.Run()
}

//Print will show the current environment
func Print(environment, name string, isolated bool) error {
	filename, err := FileFinder(name, 2)
	if err != nil {
		return err
	}

	//EnvFile.Load(filename)
	env, err := ini.Load(filename)

	if err != nil {
		return envFileErrorNotFound{err, "file not found"}
	}

	sec, err := env.GetSection(environment)
	if err != nil {
		return envSectionErrorNotFound{err, "section not found"}
	}

	if isolated == false {
		for _, e := range os.Environ() {
			fmt.Println(e)
		}
	}

	for _, k := range sec.KeyStrings() {
		value := sec.Key(k).String()
		//if value has spaces then wrap in ""
		if strings.Contains(value, " ") {
			value = fmt.Sprintf("\"%s\"", value)
		}
		fmt.Printf("%s=%s\n", k, value)
	}

	return nil
}

func interpolate(str string, vars map[string]string) string {
	s := strings.Replace(str, "${", "${.", -1)
	t, err := template.New(str).Option("missingkey=error").Delims("${", "}").Parse(s)
	if err != nil {
		fmt.Printf("Error parsing command arguments: %+v", err)
		os.Exit(1)
	}

	var buf bytes.Buffer
	err = t.Execute(&buf, vars)
	if err != nil {
		//if strict we should exit
		//else return str
		if false {
			fmt.Printf("Error parsing command arguments: %+v", err)
			os.Exit(1)
		} else {
			return str
		}
	}

	if buf.Len() == 0 {
		return str
	}

	out := buf.String()

	if out == "<no value>" {
		return str
	}

	return out
}

//FileFinder will find the file and return its full path
func FileFinder(filename string, skip int) (string, error) {
	_, caller, _, _ := runtime.Caller(skip)
	dirname := filepath.Dir(caller)
	var file string
	for dirname != "/" {
		file = filepath.Join(dirname, filename)
		_, err := os.Stat(file)
		if err == nil {
			return file, nil
		}
		dirname = filepath.Clean(dirname + "/..")
	}
	return "", envFileErrorNotFound{nil, "file not found"}
}
