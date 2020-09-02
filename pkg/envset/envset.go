package envset

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/ini.v1"
)

//Run will run the given command after loading the environment
func Run(environment, name, cmd string, args []string, isolated, expand bool) error {
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

	//Build context object from section key/values
	context := LoadIniSection(sec)

	//Replace ${VAR} and $(command) in values
	err = context.Expand(expand)
	if err != nil {
		return err
	}

	//Once we have resolved all ${VAR}/$(command) we build cmd.Env value
	vars := context.ToKVStrings()

	//Replace ${VAR} in the executable cmd arguments
	InterpolateKVStrings(args, context, expand)

	command := exec.Command(cmd, args...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	//If we want to run in an isolated context we just use
	//our variables from the loaded file
	if isolated {
		command.Env = vars
		//we actually add our context to the os
	} else {
		for k, v := range context {
			os.Setenv(k, v)
		}
	}

	return command.Run()
}

//Print will show the current environment
//We dont need to do variable replacement if we print since
//the idea is to use it as a source
func Print(environment, name string, isolated, expand bool) error {
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

	context := map[string]string{}

	//Build context object from section key/values
	context := LoadIniSection(sec)

	//Replace ${VAR} and $(command) in values
	err = context.Expand(expand)
	if err != nil {
		return err
	}

	//----- actual print action
	if isolated == false {
		for _, e := range os.Environ() {
			fmt.Println(e)
		}
	}

	for k, v := range context {
		//TODO: do proper scaping, here we want to check if its not already been "..."
		if strings.Contains(v, " ") {
			v = fmt.Sprintf("\"%s\"", v)
		}
		fmt.Printf("%s=%s\n", k, v)
	}

	return nil
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
