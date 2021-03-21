package envset

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"text/template"

	"gopkg.in/ini.v1"
)

type EnvSlice []string

type EnvMap map[string]string

//NewEnvMap returns a new EnvMap
func NewEnvMap() EnvMap {
	return make(EnvMap)
}

//LocalEnv will return the local env
func LocalEnv() EnvMap {
	env := make(EnvMap)

	for _, kv := range os.Environ() {
		c := strings.SplitN(kv, "=", 2)
		key := c[0]
		value := c[1]

		env[key] = value
	}

	return env
}

//LoadJSON will load a json environment definition
func LoadJSON(b []byte) (EnvMap, error) {
	env := make(EnvMap)
	err := json.Unmarshal(b, &env)
	return env, err
}

//LoadIniSection returns a new EnvMap from a ini section
func LoadIniSection(sec *ini.Section) EnvMap {
	env := make(EnvMap)
	for _, k := range sec.KeyStrings() {
		env[k] = sec.Key(k).String()
	}
	return env
}

//Expand ${VAR} and $(command) in values
func (e EnvMap) Expand(osExpand bool) error {
	for k, v := range e {
		res := interpolateVars(v, e)
		res, err := interpolateCmds(res, e)
		if err != nil {
			return ErrorRunningCommand{err, "error running command"}
		}

		//try using built in shell variables
		if osExpand == true {
			res = os.ExpandEnv(res)
		}

		e[k] = res
	}
	return nil
}

//GetMissingKeys will compare the keys present in `keys` with the keys present in 
//the EnvMap instance and return a list of missing keys.
func (e EnvMap) GetMissingKeys(keys []string) []string {

	missing := make([]string, len(keys))
	for i, k := range keys {
		if v := e[k]; v == "" {
			missing[i] = k
		}
	}
	return missing
}

//ToExpandedKVStrings returns an expanded list of key=value strings
func (e EnvMap) ToExpandedKVStrings(osExpand bool) []string {
	vars := e.ToKVStrings()
	InterpolateKVStrings(vars, e, osExpand)
	return vars
}

//ToKVStrings will return an slice of `key=values`
func (e EnvMap) ToKVStrings() []string {

	// vars := make([]string, 0)
	// for k, v := range context {
	// 	vars = append(vars, fmt.Sprintf("%s=%s", k, v))
	// }

	env := make([]string, len(e))
	index := 0
	for key, value := range e {
		env[index] = fmt.Sprintf("%s=%s", key, value)
		index++
	}
	return env
}

//InterpolateKVStrings replace ${VAR} in the executable cmd arguments
func InterpolateKVStrings(args []string, context EnvMap, expand bool) []string {

	for i, arg := range args {
		//we use custom interpolation string for variables we load
		args[i] = interpolateVars(arg, context)

		//try using built in OS variables
		if expand == true {
			args[i] = os.ExpandEnv(args[i])
		}
	}
	return args
}

func interpolateCmds(str string, vars map[string]string) (string, error) {
	//check if str has something that looks like a command $(.*+)
	re, err := regexp.Compile(`\$\(.*\)`)
	if err != nil {
		return "", err
	}

	matches := re.FindAllString(str, -1)

	if len(matches) == 0 {
		return str, nil
	}

	//execute command:
	for _, match := range matches {
		//Get the actual $(command)
		command := strings.Replace(match, ")", "", -1)
		command = strings.Replace(command, "$(", "", -1)

		//Some commands might have arguments: $(hostname -f)
		args := strings.Split(command, " ")
		res, err := exec.Command(args[0], args[1:]...).Output()
		if err != nil {
			return "", err
		}

		//replace $() with value
		out := string(res)
		out = strings.TrimSuffix(out, "\n")
		re := regexp.MustCompile(regexp.QuoteMeta(match))
		str = re.ReplaceAllString(str, out)
	}

	return str, nil
}

func interpolateVars(str string, vars map[string]string) string {
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

/////

func (e EnvSlice) Add(k, v string) {
	val := fmt.Sprintf("%s=%s", k, v)
	e = append(e, val)
}
