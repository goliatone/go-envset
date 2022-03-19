package main

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
	"path"
	"testing"

	"github.com/rendon/testcli"
	"github.com/stretchr/testify/assert"
)

var bin string

func init() {
	cur, _ := os.Getwd()
	bin = path.Join(cur, "envset")
}

func Test_CommandHelp(t *testing.T) {
	testcli.Run(bin, "-h")
	if !testcli.Success() {
		t.Fatalf("Expected to succeed, but failed: %q with message: %q", testcli.Error(), testcli.Stderr())
	}
}

func Test_Version(t *testing.T) {
	testcli.Run(bin, "-v")

	if !testcli.Success() {
		t.Fatalf("Expected to succeed, but failed: %q with message: %q", testcli.Error(), testcli.Stderr())
	}

	if !testcli.StdoutContains("version") {
		t.Fatalf("Expected %q to contain %q", testcli.Stdout(), "version?")
	}
}

func Test_Print(t *testing.T) {
	testcli.Run(bin, "--env-file=testdata/.envset")

	if !testcli.Success() {
		t.Fatalf("Expected to succeed, but failed: %q with message: %q", testcli.Error(), testcli.Stderr())
	}

	//TODO: how do we get the full stdout? it breaks on first \n
	if !testcli.StdoutContains("GLOBAL") {
		t.Fatalf("Expected %q to contain %q", testcli.Stdout(), "GLOBAL?")
	}
}

func Test_ExecCmd(t *testing.T) {
	dir := cd("testdata", t)

	testcli.Run(bin,
		"development",
		"--",
		"sh",
		"test.sh",
	)

	if !testcli.Success() {
		t.Fatalf("Expected to succeed, but failed: %q with message: %q", testcli.Error(), testcli.Stderr())
	}

	if !testcli.StdoutContains("out: envset_result") {
		t.Fatalf("Expected %q to contain %q", testcli.Stdout(), "envset_result?")
	}

	cd(dir, t)
}

func Test_APP_ENV(t *testing.T) {
	dir := cd("testdata", t)

	testcli.Run(bin,
		"development",
		"--",
		"sh",
		"app_env.sh",
	)

	if !testcli.Success() {
		t.Fatalf("Expected to succeed, but failed: %q with message: %q", testcli.Error(), testcli.Stderr())
	}

	if !testcli.StdoutContains("development") {
		t.Fatalf("Expected %q to contain %q", testcli.Stdout(), "APP_ENV?")
	}

	cd(dir, t)
}

func Test_Metadata(t *testing.T) {
	dir := cd("testdata", t)
	rm(".meta", t)

	testcli.Run(bin,
		"metadata",
	)

	if !testcli.Success() {
		t.Fatalf("Expected to succeed, but failed: %q with message: %q", testcli.Error(), testcli.Stderr())
	}

	assert.DirExists(t, ".meta")
	assert.FileExists(t, path.Join(".meta", "data.json"))

	rm(".meta", t)
	cd(dir, t)
}

func Test_Metadata_Print(t *testing.T) {
	dir := cd("testdata", t)
	rm(".meta", t)

	testcli.Run(bin,
		"metadata",
		"--print",
	)

	if !testcli.Success() {
		t.Fatalf("Expected to succeed, but failed: %q with message: %q", testcli.Error(), testcli.Stderr())
	}

	if !testcli.StdoutContains("development") {
		t.Fatalf("Expected %q to contain %q", testcli.Stdout(), "APP_ENV?")
	}

	if !testcli.StdoutContains("\"algorithm\": \"sha256\"") {
		t.Fatalf("Expected %q to contain %q", testcli.Stdout(), "sha256 default algorithm?")
	}

	assert.NoDirExists(t, ".meta")
	assert.NoFileExists(t, path.Join(".meta", "data.json"))

	rm(".meta", t)
	cd(dir, t)
}

func Test_Metadata_Secret_Print(t *testing.T) {
	dir := cd("testdata", t)
	rm(".meta", t)

	testcli.Run(bin,
		"metadata",
		"--print",
		"--secret=secret",
	)

	if !testcli.Success() {
		t.Fatalf("Expected to succeed, but failed: %q with message: %q", testcli.Error(), testcli.Stderr())
	}

	if !testcli.StdoutContains("\"algorithm\": \"hmac\"") {
		t.Fatalf("Expected %q to contain %q", testcli.Stdout(), "hmac algorithm?")
	}
	assert.NoDirExists(t, ".meta")
	assert.NoFileExists(t, path.Join(".meta", "data.json"))

	rm(".meta", t)
	cd(dir, t)
}

func Test_Metadata_MD5_Print(t *testing.T) {
	dir := cd("testdata", t)
	rm(".meta", t)

	testcli.Run(bin,
		"metadata",
		"--print",
		"--hash-algo",
		"md5",
	)

	if !testcli.Success() {
		t.Fatalf("Expected to succeed, but failed: %q with message: %q", testcli.Error(), testcli.Stderr())
	}

	if !testcli.StdoutContains("\"algorithm\": \"md5\"") {
		t.Fatalf("Expected %q to contain %q", testcli.Stdout(), "md5 algorithm?")
	}
	assert.NoDirExists(t, ".meta")
	assert.NoFileExists(t, path.Join(".meta", "data.json"))

	rm(".meta", t)
	cd(dir, t)
}

func Test_Metadata_Idempotency(t *testing.T) {
	dir := cd("testdata", t)
	rm(".meta", t)

	testcli.Run(bin,
		"metadata",
	)

	if !testcli.Success() {
		t.Fatalf("Expected to succeed, but failed: %q with message: %q", testcli.Error(), testcli.Stderr())
	}

	filePath := path.Join(".meta", "data.json")
	assert.DirExists(t, ".meta")
	assert.FileExists(t, filePath)

	hash1 := md5sum(filePath, t)

	testcli.Run(bin,
		"metadata",
	)

	if !testcli.Success() {
		t.Fatalf("Expected to succeed, but failed: %q with message: %q", testcli.Error(), testcli.Stderr())
	}

	hash2 := md5sum(filePath, t)

	if hash1 != hash2 {
		t.Fatal("Expected meta file to not change with no data changes")
	}

	rm(".meta", t)
	cd(dir, t)
}

func Test_MetadataOptions(t *testing.T) {
	rm("testdata/meta", t)

	testcli.Run(bin,
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
	testcli.Run(bin, "--env-file=testdata/.env")

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

	testcli.Run(bin, "template")

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

	testcli.Run(bin,
		"template",
		"--filename=env.tpl",
		"--filepath=output",
		"--env-file=.env",
	)

	if !testcli.Success() {
		t.Fatalf("Expected to succeed, but failed: %q with message: %q", testcli.Error(), testcli.Stderr())
	}

	assert.FileExists(t, path.Join("output", "env.tpl"))

	rm("output", t)
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

func md5sum(filePath string, t *testing.T) string {
	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("error removing dir: %q ", err)
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		t.Fatalf("error removing dir: %q ", err)
	}
	return hex.EncodeToString(hash.Sum(nil))
}
