package main

import (
	"os"
	"path"
	"testing"

	"github.com/rendon/testcli"
	"github.com/stretchr/testify/assert"
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

func Test_Print(t *testing.T) {
	testcli.Run("envset", "--env-file=testdata/.envset")

	if !testcli.Success() {
		t.Fatalf("Expected to succeed, but failed: %q with message: %q", testcli.Error(), testcli.Stderr())
	}

	//TODO: how do we get the full stdout? it breaks on first \n
	if !testcli.StdoutContains("GLOBAL") {
		t.Fatalf("Expected %q to contain %q", testcli.Stdout(), "GLOBAL?")
	}
}

func Test_Metadata(t *testing.T) {
	dir := cd("testdata", t)
	rm(".envmeta", t)

	testcli.Run("envset",
		"metadata",
	)

	if !testcli.Success() {
		t.Fatalf("Expected to succeed, but failed: %q with message: %q", testcli.Error(), testcli.Stderr())
	}

	assert.DirExists(t, ".envmeta")
	assert.FileExists(t, path.Join(".envmeta", "metadata.json"))

	rm(".envmeta", t)
	cd(dir, t)
}

func Test_MetadataOptions(t *testing.T) {
	rm("testdata/meta", t)

	testcli.Run("envset",
		"metadata",
		"--env-file=testdata/.envset",
		"--filepath=testdata/meta",
		"--filename=data.json",
	)

	if !testcli.Success() {
		t.Fatalf("Expected to succeed, but failed: %q with message: %q", testcli.Error(), testcli.Stderr())
	}

	assert.DirExists(t, path.Join("testdata", "meta"))
	assert.FileExists(t, path.Join("testdata", "meta", "data.json"))

	rm("testdata/meta", t)
}

func Test_DotEnvFile(t *testing.T) {
	testcli.Run("envset", "--env-file=testdata/.env")

	if !testcli.Success() {
		t.Fatalf("Expected to succeed, but failed: %q with message: %q", testcli.Error(), testcli.Stderr())
	}

	if !testcli.StdoutContains("EXPECTED") {
		t.Fatalf("Expected %q to contain %q", testcli.Stdout(), "EXPECTED?")
	}
}

func Test_Template(t *testing.T) {
	dir := cd("testdata", t)
	rm("envset.example", t)

	testcli.Run("envset", "template")

	if !testcli.Success() {
		t.Fatalf("Expected to succeed, but failed: %q with message: %q", testcli.Error(), testcli.Stderr())
	}

	assert.FileExists(t, "envset.example")

	rm("envset.example", t)
	cd(dir, t)
}

func Test_TemplateOptions(t *testing.T) {
	dir := cd("testdata", t)
	rm("output/env.tpl", t)

	testcli.Run("envset",
		"template",
		"--filename=env.tpl",
		"--filepath=output",
		"--env-file=.env",
	)

	if !testcli.Success() {
		t.Fatalf("Expected to succeed, but failed: %q with message: %q", testcli.Error(), testcli.Stderr())
	}

	assert.FileExists(t, path.Join("output", "env.tpl"))

	rm("output/env.tpl", t)
	cd(dir, t)
}

func rm(dir string, t *testing.T) {
	err := os.RemoveAll(dir)
	if err != nil {
		t.Fatalf("error removing dir: %q ", err)
	}
}

func cd(dir string, t *testing.T) string {
	cur, err := os.Getwd()
	if err != nil {
		t.Fatalf("error cd dir: %q ", err)
	}
	os.Chdir(dir)
	return cur
}
