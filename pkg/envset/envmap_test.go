package envset 

import (
	"os"
	"strings"
	"testing"
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
		t.Errorf("LocalEnv failed, expected %v got %v", "value1", result["TEST_KEY_2"])
	}

	if result["TEST_KEY_3"] != "value3" {
		t.Errorf("LocalEnv failed, expected %v got %v", "value1", result["TEST_KEY_3"])
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