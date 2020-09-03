package envset

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/ini.v1"
)

const DefaultSection = "DEFAULT"

//Run will run the given command after loading the environment
func Run(environment, name, cmd string, args []string, isolated, expand bool, required []string) error {
	//TODO: This might be an issue here!
	filename, err := FileFinder(name)
	if err != nil {
		return err
	}

	//EnvFile.Load(filename)
	env, err := ini.Load(filename)

	if err != nil {
		return envFileErrorNotFound{err, "file not found"}
	}

	//check to see if we have this section at all
	names := env.SectionStrings()
	if sort.SearchStrings(names, environment) == len(names) {
		return envFileErrorNotFound{err, "section not found"}
	}

	sec, err := env.GetSection(environment)
	if err != nil {
		return envSectionErrorNotFound{err, "section not found"}
	}

	//we don't have any values here. Is that what the user
	//wants?
	if len(sec.KeyStrings()) == 0 && isolated {
		if environment == DefaultSection {
			//running in DEFAULT but loaded an env file without a section name
			fmt.Println("we have a the follow sections but not what you want")
			for _, n := range names {
				if n == DefaultSection {
					continue
				}
				fmt.Printf("- %s\n", n)
			}
		}

		return envSectionErrorNotFound{err, fmt.Sprintf("environment %s has not key=values", environment)}
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

	//Replace '${VAR}' in the executable cmd arguments
	//note that if these are not in single quited they will
	//be resolved by the shell when we call envset and we will
	//read the the result of that replacement, even if is empty.
	InterpolateKVStrings(args, context, expand)

	command := exec.Command(cmd, args...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	//If we want to check for required variables do it now.
	missing := context.GetMissingKeys(required)
	if len(missing) > 0 {
		return fmt.Errorf("missing required keys: %s", strings.Join(missing, ","))
	}

	//If we want to run in an isolated context we just use
	//our variables from the loaded file
	if isolated {
		command.Env = vars
		//we actually add our context to the os
	} else {

		local := LocalEnv()
		for k, v := range context {
			//TODO: what do we get if we have unset variables
			if _, ok := local[k]; !ok {
				os.Setenv(k, v)
			}
		}
	}

	return command.Run()
}

//Print will show the current environment
//We dont need to do variable replacement if we print since
//the idea is to use it as a source
func Print(environment, name string, isolated, expand bool) error {
	filename, err := FileFinder(name)
	if err != nil {
		return err
	}

	//TODO: handle other formats, e.g JSON/YML
	env, err := ini.Load(filename)

	if err != nil {
		return envFileErrorNotFound{err, "file not found"}
	}

	//check to see if we have this section at all
	names := env.SectionStrings()
	if sort.SearchStrings(names, environment) == len(names) {
		return envFileErrorNotFound{err, "section not found"}
	}

	sec, err := env.GetSection(environment)
	if err != nil {
		return envSectionErrorNotFound{err, "section not found"}
	}

	//we don't have any values here. Is that what the user
	//wants?
	if len(sec.KeyStrings()) == 0 && isolated {
		if environment == DefaultSection {
			//running in DEFAULT but loaded an env file without a section name
			fmt.Println("we have a the follow sections but not what you want")
			for _, n := range names {
				if n == DefaultSection {
					continue
				}
				fmt.Printf("- %s\n", n)
			}
		}

		return envSectionErrorNotFound{err, fmt.Sprintf("environment %s has not key=values", environment)}
	}

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

	//vars := context.GetEnvSlice()
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
func FileFinder(filename string) (string, error) {
	if filepath.IsAbs(filename) {
		return filename, nil
	}

	//we want to start crawling at the current directory path
	dirname, err := os.Getwd()
	if err != nil {
		return "", err
	}

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
