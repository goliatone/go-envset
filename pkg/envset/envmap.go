package envset

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

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
	resolver := newEnvResolver(e, osExpand)
	for _, k := range sortedEnvKeys(e) {
		res, err := resolver.resolveKey(k)
		if err != nil {
			return err
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

type envResolver struct {
	source    EnvMap
	resolved  EnvMap
	resolving map[string]bool
	osExpand  bool
}

func newEnvResolver(source EnvMap, osExpand bool) *envResolver {
	return &envResolver{
		source:    source,
		resolved:  make(EnvMap, len(source)),
		resolving: make(map[string]bool, len(source)),
		osExpand:  osExpand,
	}
}

func (r *envResolver) resolveKey(key string) (string, error) {
	if val, ok := r.resolved[key]; ok {
		return val, nil
	}

	if r.resolving[key] {
		return "", fmt.Errorf("cyclic variable reference involving %s", key)
	}

	raw, ok := r.source[key]
	if !ok {
		return "", fmt.Errorf("unknown variable %s", key)
	}

	r.resolving[key] = true
	defer delete(r.resolving, key)

	res, err := interpolateVarsWithResolver(raw, func(ref string) (string, bool, error) {
		if _, ok := r.source[ref]; !ok {
			return "", false, nil
		}
		val, err := r.resolveKey(ref)
		return val, true, err
	})
	if err != nil {
		return "", fmt.Errorf("interpolate vars for %s: %w", key, err)
	}

	if hasCommandSubstitution(res) {
		cmdVars, err := r.commandEnv(key)
		if err != nil {
			return "", err
		}
		res, err = interpolateCmds(res, cmdVars)
		if err != nil {
			return "", ErrorRunningCommand{err, "error running command"}
		}
	}

	if r.osExpand {
		res = os.ExpandEnv(res)
	}

	r.resolved[key] = res
	return res, nil
}

func (r *envResolver) commandEnv(current string) (EnvMap, error) {
	env := make(EnvMap, len(r.source))
	for _, key := range sortedEnvKeys(r.source) {
		if key == current {
			continue
		}
		if val, ok := r.resolved[key]; ok {
			env[key] = val
			continue
		}
		if hasCommandSubstitution(r.source[key]) {
			continue
		}
		val, err := r.resolveKey(key)
		if err != nil {
			return nil, err
		}
		env[key] = val
	}
	return env, nil
}

func sortedEnvKeys(env EnvMap) []string {
	keys := make([]string, 0, len(env))
	for key := range env {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func hasCommandSubstitution(str string) bool {
	return strings.Contains(str, "$(")
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
	return interpolateVarsWithResolver(str, func(key string) (string, bool, error) {
		val, ok := vars[key]
		return val, ok, nil
	})
}

func interpolateVarsWithResolver(str string, resolve func(string) (string, bool, error)) (string, error) {
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
		val, ok, err := resolve(key)
		if err != nil {
			return "", err
		}
		if ok {
			out.WriteString(val)
		} else {
			out.WriteString(str[i : i+2+end+1])
		}
		i = i + 2 + end + 1
	}
	return out.String(), nil
}

/////

func (e *EnvSlice) Add(k, v string) {
	val := fmt.Sprintf("%s=%s", k, v)
	*e = append(*e, val)
}
