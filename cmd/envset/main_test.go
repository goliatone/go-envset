package main

import (
	"testing"

	"github.com/rendon/testcli"
)

func Test_CommandHelp(t *testing.T) {
	testcli.Run("envset", "-h")
	if !testcli.Success() {
		t.Fatalf("Expected to succeed, but failed: %q with message: %q", testcli.Error(), testcli.Stderr())
	}
}

func Test_Version(t *testing.T) {
	testcli.Run("envset", "-V")

	if !testcli.Success() {
		t.Fatalf("Expected to succeed, but failed: %q with message: %q", testcli.Error(), testcli.Stderr())
	}

	if !testcli.StdoutContains("version") {
		t.Fatalf("Expected %q to contain %q", testcli.Stdout(), "version?")
	}
}
