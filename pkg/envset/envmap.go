package envset

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"gopkg.in/ini.v1"
)

// EnvSlice type to hold envset entries
type EnvSlice []string

// EnvMap type to hold envset map
type EnvMap map[string]string

// NewEnvMap returns a new EnvMap
func NewEnvMap() EnvMap {
	return make(EnvMap)
}

// LocalEnv will return the local env
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

// LoadJSON will load a json environment definition
func LoadJSON(b []byte) (EnvMap, error) {
	env := make(EnvMap)
	if err := json.Unmarshal(b, &env); err != nil {
		return env, fmt.Errorf("load json: %w", err)
	}
	return env, nil
}

// LoadIniSection returns a new EnvMap from a ini section
func LoadIniSection(sec *ini.Section) EnvMap {
	env := make(EnvMap)
	for _, k := range sec.KeyStrings() {
		env[k] = sec.Key(k).String()
	}
	return env
}

// Expand ${VAR} and $(command) in values
func (e EnvMap) Expand(osExpand bool) error {
	for k, v := range e {
		res, err := interpolateVars(v, e)
		if err != nil {
			return fmt.Errorf("interpolate vars: %w", err)
		}

		res, err = interpolateCmds(res, e)
		if err != nil {
			return ErrorRunningCommand{err, "error running command"}
		}

		//try using built in shell variables
		if osExpand {
			res = os.ExpandEnv(res)
		}

		e[k] = res
	}
	return nil
}

// GetMissingKeys will compare the keys present in `keys` with the keys present in
// the EnvMap instance and return a list of missing keys.
func (e EnvMap) GetMissingKeys(keys []string) []string {

	missing := make([]string, 0)
	for _, k := range keys {
		if v := e[k]; v == "" {
			missing = append(missing, k)
		}
	}
	return missing
}

// ToExpandedKVStrings returns an expanded list of key=value strings
func (e EnvMap) ToExpandedKVStrings(osExpand bool) []string {
	vars := e.ToKVStrings()
	InterpolateKVStrings(vars, e, osExpand)
	return vars
}

// ToKVStrings will return an slice of `key=values`
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

// InterpolateKVStrings replace ${VAR} in the executable cmd arguments
func InterpolateKVStrings(args []string, context EnvMap, expand bool) []string {
	args, _ = interpolateKVStrings(args, context, expand)
	return args
}

func interpolateKVStrings(args []string, context EnvMap, expand bool) ([]string, error) {
	for i, arg := range args {
		//we use custom interpolation string for variables we load
		interpolated, err := interpolateVars(arg, context)
		if err != nil {
			return args, err
		}
		args[i] = interpolated

		//try using built in OS variables
		if expand {
			args[i] = expandBracedEnv(args[i])
		}
	}
	return args, nil
}

func expandBracedEnv(str string) string {
	var out strings.Builder
	for i := 0; i < len(str); {
		if i+2 >= len(str) || str[i] != '$' || str[i+1] != '{' {
			out.WriteByte(str[i])
			i++
			continue
		}

		end := strings.IndexByte(str[i+2:], '}')
		if end == -1 {
			out.WriteString(str[i:])
			break
		}

		key := str[i+2 : i+2+end]
		out.WriteString(os.Getenv(key))
		i = i + 2 + end + 1
	}
	return out.String()
}

func interpolateCmds(str string, vars map[string]string) (string, error) {
	var out strings.Builder
	for i := 0; i < len(str); {
		if i+1 >= len(str) || str[i] != '$' || str[i+1] != '(' {
			out.WriteByte(str[i])
			i++
			continue
		}

		command, end, err := scanCommandSubstitution(str, i)
		if err != nil {
			return "", err
		}

		if command == "" {
			i = end
			continue
		}

		res, err := runShellCommand(command, vars)
		if err != nil {
			return "", err
		}

		out.WriteString(strings.TrimSuffix(res, "\n"))
		i = end
	}

	return out.String(), nil
}

func scanCommandSubstitution(str string, start int) (string, int, error) {
	depth := 1
	bodyStart := start + 2
	quote := byte(0)
	escaped := false

	for i := bodyStart; i < len(str); i++ {
		ch := str[i]

		if escaped {
			escaped = false
			continue
		}

		if ch == '\\' {
			escaped = true
			continue
		}

		if quote != 0 {
			if ch == quote {
				quote = 0
			}
			continue
		}

		if ch == '\'' || ch == '"' {
			quote = ch
			continue
		}

		if ch == '$' && i+1 < len(str) && str[i+1] == '(' {
			depth++
			i++
			continue
		}

		if ch == ')' {
			depth--
			if depth == 0 {
				return str[bodyStart:i], i + 1, nil
			}
		}
	}

	return "", 0, fmt.Errorf("unterminated command substitution")
}

func runShellCommand(command string, vars map[string]string) (string, error) {
	cmd := exec.Command("/bin/sh", "-c", command)
	cmd.Env = os.Environ()
	for k, v := range vars {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	res, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("exec command: %w", err)
	}

	return string(res), nil
}

func interpolateVars(str string, vars map[string]string) (string, error) {
	s := strings.Replace(str, "${", "${.", -1)
	t, err := template.New(str).Option("missingkey=error").Delims("${", "}").Parse(s)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	var buf bytes.Buffer
	err = t.Execute(&buf, vars)
	if err != nil {
		return str, nil
	}

	if buf.Len() == 0 {
		return str, nil
	}

	out := buf.String()

	if out == "<no value>" {
		return str, nil
	}

	return out, nil
}

/////

func (e *EnvSlice) Add(k, v string) {
	val := fmt.Sprintf("%s=%s", k, v)
	*e = append(*e, val)
}
