package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"

	// "path/filepath"
	// "runtime"

	"github.com/vaughan0/go-ini"
)

var environments []string

func init() {
	//TODO: we need to get this from envsetrc
	//use https://github.com/jinzhu/configor
	environments = []string{"development", "production", "staging", "testing", "local"}
	/*
		//Get the path to the current script
		_, filename, _, _ := runtime.Caller(0)
		//get the path from the script
		dirname := filepath.Dir(filename)

		//recursively walk directory structure upward, trying to
		//find our file until we reach root
		for dirname != "/" {
			filename = filepath.Join(dirname, ".envsetrc")
			fmt.Printf("path: %s\n", filename)
			config, err := ini.LoadFile(filename)
			if err == nil {
				// envs := config["environments"]
				// fmt.Printf("config: %q\n", config)
				// fmt.Printf("envs: %q\n", envs)
				break
			}

			dirname = filepath.Clean(dirname + "/..")
		}*/
}

func main() {
	//Show help
	if len(os.Args) == 1 {
		showHelpMessage(os.Args[0])
		os.Exit(0)
	}
	//This should recursively try to find an envset file
	//from here upwards.
	env, err := ini.LoadFile(".envset")
	if err != nil {
		fmt.Println("Error parsing envset")
		os.Exit(1)
	}

	found := false
	//os.Args[0] is the main command `envset`
	//os.Args[1] is the environment
	environment := os.Args[1]

	switch environment {
	case "sync":
		syncDemoFile()
		os.Exit(0)
		break
	case "help":
		showHelpMessage(os.Args[0])
		os.Exit(0)
		break
	}

	vars := make([]string, 0)
	var tplVars map[string]string
	for name, section := range env {
		if name == environment {
			if !validEnvironment(environment, environments) {
				notValidEnvironmentMessage(environment)
				os.Exit(1)
			}

			found = true
			tplVars = map[string]string{}
			for key, value := range section {
				// os.Setenv(key, value)
				vars = append(vars, fmt.Sprintf("%s=%s", key, value))
				tplVars[key] = value
			}
		}
	}

	if found == false {
		fmt.Printf("Error, environment %q not found.\n", environment)
		os.Exit(1)
	}

	if len(os.Args) < 3 {
		showCurrentEnvironment(vars)
		os.Exit(0)
	}

	command := os.Args[3]
	args := os.Args[4:]

	//TODO: Run args string on replacement function so that we can replace ${MY_VAR}s
	for i, arg := range args {
		fmt.Printf("arg: %s\n", arg)
		args[i] = interpolate(arg, tplVars)
	}

	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = vars
	cmd.Run()
}

func interpolate(str string, vars map[string]string) string {
	s := strings.Replace(str, "${", "${.", -1)
	tmpl, err := template.New(str).Option("missingkey=error").Delims("${", "}").Parse(s)
	if err != nil {
		fmt.Printf("Error parsing command arguments: %+v", err)
		os.Exit(1)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, vars)
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
	fmt.Printf("replaced to: %s len %d\n", out, buf.Len())

	if out == "<no value>" {
		return str
	}
	return out
}

func validEnvironment(env string, list []string) bool {
	for _, v := range list {
		if v == env {
			return true
		}
	}
	return false
}

func notValidEnvironmentMessage(env string) {
	fmt.Println("Environment not recognized")
}

func showCurrentEnvironment(vars []string) {
	for _, value := range vars {
		fmt.Printf("%s\n", value)
	}
}

func showHelpMessage(progname string) {
	fmt.Printf("%s: the environment management friend\n", progname)
	usage := `
envset --help           => Usage()
envset                  => ListEnvs()
envset [environment]    => ShowEnv()
envset [environment] -- [cmd] <args>   => ExecuteCommand()
envset init             => InitializeProject()
envset sync             => SyncProject()
envset patch create     => PatchCreate()
envset patch submit     => PatchSubmit()
envset patch apply      => PatchApply()

Configuration options:
* name for template file default to envset.tpl
* filepath for envset.tpl


* How to watch a file and trigger a function on change
* How to create a git hook in go? or bash?


git hooks:
* before commit: 
    * Check if local envset file has changed, if so check to see if template new values were filled with data
* on checkout:
    * Check to see if we have an environment for this branch?
    * Check to see if remote source has changed

Look at [git-lfs](https://github.com/git-lfs/git-lfs) to see how they extend git with a new lfs subcommand so that we can have git env track
	`
	fmt.Printf("%s\n", usage)
}

func syncDemoFile() {

}
