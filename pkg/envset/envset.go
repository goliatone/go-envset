package envset

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"syscall"

	"gopkg.in/ini.v1"
)

// DefaultSection is the name of the default ini section
const DefaultSection = ini.DEFAULT_SECTION

// RunOptions is used to configure a run command
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
	Restart             bool
	MaxRestarts         int
}

type runOutput struct {
	completed bool
	reload    bool
	err       error
}

func (r runOutput) tryRestart() bool {
	return r.reload || (r.completed && r.err != nil)
}

var runs = 1
var command *exec.Cmd

// Run will run the given command after loading the environment
func Run(environment string, options RunOptions) error {
	ch := make(chan runOutput)
	defer close(ch)

	rl := make(chan os.Signal, 1)
	signal.Notify(rl, syscall.SIGUSR2)
	defer close(rl)

	go doRun(environment, options, ch)

	var res runOutput

	select {
	case res = <-ch:
		// Run is done...
		// Current implementation means that if we had
		// an restart due to error then we are also
		// reloading the env.
	case <-rl:
		// command.Process.Signal(syscall.SIGUSR2)
		command.Process.Kill()
		res = runOutput{
			reload: true,
		}
	}

	if res.tryRestart() {
		if options.Restart {
			if runs < options.MaxRestarts {
				runs = runs + 1
				return Run(environment, options)
			}
		}
		return res.err
	}

	return res.err
}

func doRun(environment string, options RunOptions, ch chan runOutput) {
	env, err := getEnvFile(options)
	if err != nil {
		ch <- runOutput{
			err: err,
		}
		return
	}

	context, err := getContext(environment, env, options)
	if err != nil {
		ch <- runOutput{
			err: err,
		}
		return
	}

	//Replace ${VAR} and $(command) in values
	err = context.Expand(options.Expand)
	if err != nil {
		ch <- runOutput{
			err: fmt.Errorf("context expand: %w", err),
		}
		return
	}

	//Once we have resolved all ${VAR}/$(command) we build cmd.Env value
	vars := context.ToKVStrings()

	//Replace '${VAR}' in the executable cmd arguments
	//note that if these are not in single quotes they will
	//be resolved by the shell when we call envset and we will
	//read the the result of that replacement, even if is empty.
	InterpolateKVStrings(options.Args, context, options.Expand)

	//If we want to check for required variables do it now.
	missing := context.GetMissingKeys(options.Required)
	if len(missing) > 0 {
		ch <- runOutput{
			err: fmt.Errorf("missing required keys: %s", strings.Join(missing, ",")),
		}
		return
	}

	command = exec.Command(options.Cmd, options.Args...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	//If we want to run in an isolated context we just use
	//our variables from the loaded file
	if options.Isolated {
		command.Env = vars
		//add value for any inherited env vars we have in options
		for _, k := range options.Inherit {
			if v, ok := os.LookupEnv(k); !ok {
				command.Env = append(command.Env, fmt.Sprintf("%s=%s", k, v))
			}
		}
	} else {
		local := LocalEnv()
		for k, v := range context {
			if _, ok := local[k]; !ok {
				os.Setenv(k, v)
			}
		}
	}

	err = command.Run()

	ch <- runOutput{
		err:       err,
		completed: true,
	}
}

// Print will show the current environment
// We don't need to do variable replacement if we print since
// the idea is to use it as a source
func Print(environment string, options RunOptions) error {
	env, err := getEnvFile(options)
	if err != nil {
		return err
	}

	context, err := getContext(environment, env, options)
	if err != nil {
		return err
	}

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

// FileFinder will find the file and return its full path
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

func getContext(environment string, env *ini.File, options RunOptions) (EnvMap, error) {
	sec, err := getSec(environment, env, options)
	if err != nil {
		return EnvMap{}, err
	}
	//Build context object from section key/values
	context := LoadIniSection(sec)
	return context, nil
}

func getEnvFile(options RunOptions) (*ini.File, error) {
	filename, err := FileFinder(options.Filename) //TODO: This might be an issue here!
	if err != nil {
		return nil, fmt.Errorf("file finder %s: %w", options.Filename, err)
	}

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

func getSec(environment string, env *ini.File, options RunOptions) (*ini.Section, error) {
	// check to see if we have this section at all
	names := env.SectionStrings()
	if sort.SearchStrings(names, environment) == len(names) {
		return nil, fmt.Errorf("section not defined in envsetrc")
	}

	sec, err := env.GetSection(environment)
	if err != nil {
		return nil, envSectionErrorNotFound{
			err,
			fmt.Sprintf("run: section [%s] not found in env file", environment),
		}
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

		return nil, envSectionErrorNotFound{err, fmt.Sprintf("environment %s has not key=values", environment)}
	}

	//Ensure we export the env name to the environment
	//e.g. APP_ENV=development
	if !sec.HasKey(options.ExportEnvName) {
		sec.NewKey(options.ExportEnvName, environment)
	}

	return sec, nil
}
