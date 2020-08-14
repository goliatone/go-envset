package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/vaughan0/go-ini"
)

var environments []string

func init() {
	//TODO: we need to get this from envsetrc
	//use https://github.com/jinzhu/configor
	environments = []string{"development", "production", "staging", "testing", "local"}

	_, filename, _, _ := runtime.Caller(1)
	dirname := filepath.Dir(filename)

	//recursively walk directory structure upward, trying to
	//find our file until we reach root
	for dirname != "/" {
		dirname = filepath.Clean(dirname + "/..")
		filename = filepath.Join(dirname, ".envsetrc")

		config, err := ini.LoadFile(filename)
		if err == nil {
			envs := config["environments"]
			fmt.Printf("config: %q\n", envs)
			fmt.Printf("path: %s\n", filename)
			// environments = config.Get("environment")
			break
		}
	}
}

func main() {
	//Show help
	if len(os.Args) == 1 {
		showHelpMessage()
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
	environment := os.Args[1]
	vars := make([]string, 0)

	for name, section := range env {
		if name == environment {
			if !validEnvironment(environment, environments) {
				notValidEnvironmentMessage(environment)
				os.Exit(1)
			}

			found = true
			for key, value := range section {
				// os.Setenv(key, value)
				vars = append(vars, fmt.Sprintf("%s=%s", key, value))
				// vars[key] = value
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

	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = vars
	cmd.Run()
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

func showHelpMessage() {
	fmt.Println("envset: the environment management friend")
}
