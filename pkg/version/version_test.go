package version

import (
	"testing"
)

func Test_GetVersion(t *testing.T) {
	Tag = "0.0.0"
	Time = "Time"
	User = "username"

	expected := "0.0.0-Time:username"

	result := GetVersion()
	if result == "" {
		t.Errorf("GetVersion failed, expected %v got empty value", expected)
	}

	if result != expected {
		t.Errorf("GetVersion failed, expected %v got %v", expected, result)
	}
}
