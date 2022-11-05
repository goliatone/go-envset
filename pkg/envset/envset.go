package envset

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/ini.v1"
)

//DefaultSection is the name of the default ini section
const DefaultSection = ini.DEFAULT_SECTION

//RunOptions is used to configure a run command
type RunOptions struct {
	Filename            string
	Cmd                 string
	Args                []string
	Isolated            bool
	Expand              bool
	Required            []string
	Inherit             []string
	Ignored             []string
	ExportEnvName       string
	CommentSectionNames []string
}

//Run will run the given command after loading the environment
func Run(environment string, options RunOptions) error {
	env, err := getEnv(options)
	if err != nil {
		return err
	}

	// check to see if we have this section at all
	names := env.SectionStrings()
	if sort.SearchStrings(names, environment) == len(names) {
		return envFileErrorNotFound{err, "section not found"}
	}

	sec, err := env.GetSection(environment)
	if err != nil {
		return envSectionErrorNotFound{err, "section not found"}
	}

	// we don't have any values here.
	// Is that what the user wants?
	if len(sec.KeyStrings()) == 0 && options.Isolated {
		if environment == DefaultSection {
			//running in DEFAULT but loaded an env file without a section name
			for _, n := range names {
				if n == DefaultSection {
					continue
				}
				fmt.Printf("- %s\n", n)
			}
		}

		return envSectionErrorNotFound{err, fmt.Sprintf("environment %s has not key=values", environment)}
	}

	//Ensure we export the env name to the environment
	//e.g. APP_ENV=development
	if !sec.HasKey(options.ExportEnvName) {
		sec.NewKey(options.ExportEnvName, environment)
	} else {
		sec.DeleteKey(options.ExportEnvName)
		sec.NewKey(options.ExportEnvName, environment)
	}

	//Build context object from section key/values
	context := LoadIniSection(sec)

	//Replace ${VAR} and $(command) in values
	err = context.Expand(options.Expand)
	if err != nil {
		return fmt.Errorf("context expand: %w", err)
	}

	//Once we have resolved all ${VAR}/$(command) we build cmd.Env value
	vars := context.ToKVStrings()

	//Replace '${VAR}' in the executable cmd arguments
	//note that if these are not in single quotes they will
	//be resolved by the shell when we call envset and we will
	//read the the result of that replacement, even if is empty.
	InterpolateKVStrings(options.Args, context, options.Expand)

	command := exec.Command(options.Cmd, options.Args...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	//If we want to check for required variables do it now.
	missing := context.GetMissingKeys(options.Required)
	if len(missing) > 0 {
		return fmt.Errorf("missing required keys: %s", strings.Join(missing, ","))
	}

	//If we want to run in an isolated context we just use
	//our variables from the loaded file
	if options.Isolated {
		command.Env = vars
		//add value for any inherited env vars we have in options
		for _, k := range options.Inherit {
			if v := os.Getenv(k); v != "" {
				command.Env = append(command.Env, fmt.Sprintf("%s=%s", k, v))
			}
		}
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
//We don't need to do variable replacement if we print since
//the idea is to use it as a source
func Print(environment string, options RunOptions) error {
	env, err := getEnv(options)
	if err != nil {
		return err
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
	if len(sec.KeyStrings()) == 0 && options.Isolated {
		if environment == DefaultSection {
			//running in DEFAULT but loaded an env file without a section name
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
	err = context.Expand(options.Expand)
	if err != nil {
		return fmt.Errorf("context expand: %w", err)
	}

	//----- actual print action
	if options.Isolated == false {
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
		return "", fmt.Errorf("get wd: %w", err)
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

func getEnv(options RunOptions) (*ini.File, error) {
	//TODO: This might be an issue here!
	filename, err := FileFinder(options.Filename)
	if err != nil {
		return nil, fmt.Errorf("file finder %s: %w", options.Filename, err)
	}

	//EnvFile.Load(filename)
	//TODO: handle other formats, e.g JSON/YML
	env, err := ini.LoadSources(ini.LoadOptions{
		UnparseableSections:     options.CommentSectionNames,
		SkipUnrecognizableLines: true,
	}, filename)

	if err != nil {
		if ini.IsErrDelimiterNotFound(err) {
			fmt.Printf("The file \"%s\" has an error and we can't parse it.\n", options.Filename)
			fmt.Println("It looks as if you forgot a variable name.")
			delErr := err.(ini.ErrDelimiterNotFound)
			if errors.As(err, &delErr) {
				fmt.Printf("The offending line content: %s\n", delErr.Line)
			}
		}
		//error parsing data source: unknown type
		return nil, fmt.Errorf("file load: %w", err)
	}
	return env, err
}
