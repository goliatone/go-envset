package envset 

import (
	"os"
	"strings"
	"testing"

	"gopkg.in/ini.v1"
)

func Test_LocalEnv(t *testing.T) {
	defer unsetEnv("TEST_KEY_")()

	os.Setenv("TEST_KEY_1", "value1")
	os.Setenv("TEST_KEY_2", "value2")
	os.Setenv("TEST_KEY_3", "value3")

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

	cfg, _ := ini.Load(
		[]byte(`
[test]
TEST_KEY_1=value1
TEST_KEY_2=value2
TEST_KEY_3=value3
		`),
	)

	sec, _ := cfg.GetSection("test")

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

func Test_Expand(t *testing.T){}

func Test_GetMissingKeys(t *testing.T) {
	fixture := []byte("{\"TEST_KEY_1\": \"value1\", \"TEST_KEY_2\": \"value2\",\"TEST_KEY_3\": \"value3\"}")
	
	result, err := LoadJSON(fixture)
	if err != nil {
		t.Errorf("LoadJSON failed, unexpected error %v", err)
	}

	missing := result.GetMissingKeys([]string{"TEST_KEY_1", "TEST_KEY_2", "TEST_KEY_3"})
	if missing[0] != "" {
		t.Errorf("Missing keys should be 0: %v", missing)
	}

	missing = result.GetMissingKeys([]string{"MISSING_KEY"})
	if missing[0] != "MISSING_KEY" {
		t.Error("Missing keys should be 1")
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


func unsetEnv(prefix string)(restore func()) {
	before := map[string]string{}
	
	for _, e := range os.Environ() {
		if !strings.HasPrefix(e, prefix) {
			continue
		}
		parts := strings.SplitN(e, "=", 2)
		before[parts[0]] = parts[1]
		os.Unsetenv(parts[0])
	}

	return func() {
		after := map[string]string{}

		for _, e := range os.Environ() {
			if !strings.HasPrefix(e, prefix) {
				continue
			}
			parts := strings.SplitN(e, "=", 2)
			after[parts[0]] = parts[1]

			v, ok :=  before[parts[0]]
			if !ok {
				os.Unsetenv(parts[0])
				continue
			}
			if parts[1] != v {
				//If the env var has changed, set it back
				os.Setenv(parts[0], v)
			}
		}
		for k, v := range before {
			if _, ok := after[k]; !ok {
				//add missing k
				os.Setenv(k, v)
			}
		}
	}
}