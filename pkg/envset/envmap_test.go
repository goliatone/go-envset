package envset

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/ini.v1"
)

func Test_LocalEnv(t *testing.T) {
	defer unsetEnv(t, "TEST_KEY_")()

	setEnv(t, "TEST_KEY_1", "value1")
	setEnv(t, "TEST_KEY_2", "value2")
	setEnv(t, "TEST_KEY_3", "value3")

	result := LocalEnv()

	if result["TEST_KEY_1"] != "value1" {
		t.Errorf("LocalEnv failed, expected %v got %v", "value1", result["TEST_KEY_1"])
	}

	if result["TEST_KEY_2"] != "value2" {
		t.Errorf("LocalEnv failed, expected %v got %v", "value2", result["TEST_KEY_2"])
	}

	if result["TEST_KEY_3"] != "value3" {
		t.Errorf("LocalEnv failed, expected %v got %v", "value3", result["TEST_KEY_3"])
	}
}

func Test_LoadJSON(t *testing.T) {
	fixture := []byte("{\"TEST_KEY_1\": \"value1\", \"TEST_KEY_2\": \"value2\",\"TEST_KEY_3\": \"value3\"}")

	result, err := LoadJSON(fixture)

	if err != nil {
		t.Errorf("LoadJSON failed, unexpected error %v", err)
	}

	if result["TEST_KEY_1"] != "value1" {
		t.Errorf("LoadJSON failed, expected %v got %v", "value1", result["TEST_KEY_1"])
	}

	if result["TEST_KEY_2"] != "value2" {
		t.Errorf("LoadJSON failed, expected %v got %v", "value1", result["TEST_KEY_2"])
	}

	if result["TEST_KEY_3"] != "value3" {
		t.Errorf("LoadJSON failed, expected %v got %v", "value1", result["TEST_KEY_3"])
	}
}

func Test_LoadIniSection(t *testing.T) {

	cfg, err := ini.Load(
		[]byte(`
[test]
TEST_KEY_1=value1
TEST_KEY_2=value2
TEST_KEY_3=value3
		`),
	)
	if err != nil {
		t.Fatalf("ini load: %v", err)
	}

	sec, err := cfg.GetSection("test")
	if err != nil {
		t.Fatalf("get section: %v", err)
	}

	result := LoadIniSection(sec)

	if result["TEST_KEY_1"] != "value1" {
		t.Errorf("LoadJSON failed, expected %v got %v", "value1", result["TEST_KEY_1"])
	}

	if result["TEST_KEY_2"] != "value2" {
		t.Errorf("LoadJSON failed, expected %v got %v", "value1", result["TEST_KEY_2"])
	}

	if result["TEST_KEY_3"] != "value3" {
		t.Errorf("LoadJSON failed, expected %v got %v", "value1", result["TEST_KEY_3"])
	}
}

func Test_Expand(t *testing.T) {}

func Test_Expand_NestedVariablesDeterministic(t *testing.T) {
	for range 100 {
		env := EnvMap{
			"A": "${B}",
			"B": "${C}",
			"C": "ok",
		}

		if err := env.Expand(false); err != nil {
			t.Fatalf("expand: %v", err)
		}

		if env["A"] != "ok" {
			t.Fatalf("A = %q, want ok", env["A"])
		}
	}
}

func Test_Expand_CommandSubstitutionSeesResolvedEnv(t *testing.T) {
	env := EnvMap{
		"COMMAND": "$(printf \"$VALUE\")",
		"VALUE":   "${BASE}-suffix",
		"BASE":    "prefix",
	}

	if err := env.Expand(false); err != nil {
		t.Fatalf("expand: %v", err)
	}

	if env["COMMAND"] != "prefix-suffix" {
		t.Fatalf("COMMAND = %q, want prefix-suffix", env["COMMAND"])
	}
}

func Test_Expand_CyclicVariableReference(t *testing.T) {
	env := EnvMap{
		"A": "${B}",
		"B": "${A}",
	}

	if err := env.Expand(false); err == nil {
		t.Fatal("expected cycle error")
	}
}

func Test_GetMissingKeys(t *testing.T) {
	fixture := []byte("{\"TEST_KEY_1\": \"value1\", \"TEST_KEY_2\": \"value2\",\"TEST_KEY_3\": \"value3\"}")

	result, err := LoadJSON(fixture)
	if err != nil {
		t.Errorf("LoadJSON failed, unexpected error %v", err)
	}

	missing := result.GetMissingKeys([]string{"TEST_KEY_1", "TEST_KEY_2", "TEST_KEY_3"})
	if len(missing) != 0 {
		t.Errorf("Missing keys should be 0: %v", missing)
	}

	missing = result.GetMissingKeys([]string{"MISSING_KEY"})
	if !reflect.DeepEqual(missing, []string{"MISSING_KEY"}) {
		t.Errorf("Missing keys should be [MISSING_KEY]: %v", missing)
	}

	missing = result.GetMissingKeys([]string{"TEST_KEY_1", "MISSING_KEY", "TEST_KEY_2"})
	if !reflect.DeepEqual(missing, []string{"MISSING_KEY"}) {
		t.Errorf("Missing keys should not include blank entries: %v", missing)
	}
}

func Test_GetMissingKeys_EmptyValue(t *testing.T) {
	env := EnvMap{
		"EMPTY_KEY": "",
	}

	missing := env.GetMissingKeys([]string{"EMPTY_KEY"})
	if !reflect.DeepEqual(missing, []string{"EMPTY_KEY"}) {
		t.Errorf("empty value should count as missing: %v", missing)
	}
}

func Test_Expand_CommandSubstitution(t *testing.T) {
	env := EnvMap{
		"JOINED":   "$(printf a)$(printf b)",
		"QUOTED":   "$(printf \"a b\")",
		"PIPE":     "$(printf abc | tr a-z A-Z)",
		"FROM_ENV": "$(printf \"$BASE-suffix\")",
		"BASE":     "prefix",
	}

	if err := env.Expand(false); err != nil {
		t.Fatalf("expand: %v", err)
	}

	tests := map[string]string{
		"JOINED":   "ab",
		"QUOTED":   "a b",
		"PIPE":     "ABC",
		"FROM_ENV": "prefix-suffix",
	}
	for key, want := range tests {
		if env[key] != want {
			t.Fatalf("%s = %q, want %q", key, env[key], want)
		}
	}
}

func Test_Expand_CommandSubstitutionFailure(t *testing.T) {
	env := EnvMap{
		"FAILS": "$(exit 7)",
	}

	if err := env.Expand(false); err == nil {
		t.Fatal("expected command substitution failure")
	}
}

func Test_EnvSliceAdd(t *testing.T) {
	env := EnvSlice{}

	env.Add("KEY", "value")

	if !reflect.DeepEqual(env, EnvSlice{"KEY=value"}) {
		t.Fatalf("env = %v, want [KEY=value]", env)
	}
}

func Test_InterpolateKVStringsPreservesBareShellVars(t *testing.T) {
	args := []string{"sh", "-c", "printf \"$A\""}

	got, err := interpolateKVStrings(args, EnvMap{"A": "envset_value"}, true)
	if err != nil {
		t.Fatalf("interpolate args: %v", err)
	}

	if got[2] != "printf \"$A\"" {
		t.Fatalf("arg = %q, want bare shell variable preserved", got[2])
	}
}

func Test_ToKVStrings(t *testing.T) {
	fixture := []byte("{\"TEST_KEY_1\": \"value1\", \"TEST_KEY_2\": \"value2\",\"TEST_KEY_3\": \"value3\"}")

	env, err := LoadJSON(fixture)
	if err != nil {
		t.Errorf("ToKVStrings failed, unexpected error %v", err)
	}

	result := env.ToKVStrings()

	expected := map[string]bool{
		"TEST_KEY_1=value1": true,
		"TEST_KEY_2=value2": true,
		"TEST_KEY_3=value3": true,
	}

	for i := range result {
		if _, ok := expected[result[i]]; ok != true {
			t.Error("ToKVStrings failed, unexpected error")
		}
	}
}

func setEnv(t *testing.T, key, value string) {
	t.Helper()

	if err := os.Setenv(key, value); err != nil {
		t.Fatalf("set env %s: %v", key, err)
	}
}

func unsetEnv(t *testing.T, prefix string) (restore func()) {
	t.Helper()

	before := map[string]string{}

	for _, e := range os.Environ() {
		if !strings.HasPrefix(e, prefix) {
			continue
		}
		parts := strings.SplitN(e, "=", 2)
		before[parts[0]] = parts[1]
		if err := os.Unsetenv(parts[0]); err != nil {
			t.Fatalf("unset env %s: %v", parts[0], err)
		}
	}

	return func() {
		after := map[string]string{}

		for _, e := range os.Environ() {
			if !strings.HasPrefix(e, prefix) {
				continue
			}
			parts := strings.SplitN(e, "=", 2)
			after[parts[0]] = parts[1]

			v, ok := before[parts[0]]
			if !ok {
				if err := os.Unsetenv(parts[0]); err != nil {
					t.Fatalf("unset env %s: %v", parts[0], err)
				}
				continue
			}
			if parts[1] != v {
				// If the env var has changed, set it back.
				setEnv(t, parts[0], v)
			}
		}
		for k, v := range before {
			if _, ok := after[k]; !ok {
				// Add missing k.
				setEnv(t, k, v)
			}
		}
	}
}
