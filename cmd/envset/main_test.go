package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"os"
	osexec "os/exec"
	"path"
	"path/filepath"
	"testing"

	"github.com/rendon/testcli"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

var bin string

func init() {
	cur, _ := os.Getwd()
	bin = path.Join(cur, "envset")
}

func TestMain(m *testing.M) {
	cur, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "get wd: %v\n", err)
		os.Exit(1)
	}

	bin = filepath.Join(os.TempDir(), "envset-test-bin")
	cmd := osexec.Command("go", "build", "-o", bin, ".")
	cmd.Dir = cur
	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "build envset: %v\n%s\n", err, out)
		os.Exit(1)
	}

	code := m.Run()
	_ = os.Remove(bin)
	os.Exit(code)
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
	assertNotGroupOrWorldWritable(t, ".meta")
	assertRegularFileMode(t, path.Join(".meta", "data.json"))

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

func Test_MetadataCompareInvalidJSONReportsLoadError(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "source.json")
	target := filepath.Join(dir, "target.json")

	if err := os.WriteFile(source, []byte("{not-json"), 0644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	valid := `{"algorithm":"sha256","sections":[{"name":"development","values":[]}]}`
	if err := os.WriteFile(target, []byte(valid), 0644); err != nil {
		t.Fatalf("write target: %v", err)
	}

	testcli.Run(bin,
		"metadata",
		"compare",
		"--section=development",
		source,
		target,
	)

	if testcli.Success() {
		t.Fatal("Expected metadata compare to fail")
	}
	if !testcli.StdoutContains("Unable to load source metadata file") && !testcli.StderrContains("Unable to load source metadata file") {
		t.Fatalf("Expected load error, stdout: %q stderr: %q", testcli.Stdout(), testcli.Stderr())
	}
}

func Test_MetadataOverwriteTightensFilePermissions(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".envset")
	metaDir := filepath.Join(dir, "meta")
	metaFile := filepath.Join(metaDir, "data.json")

	if err := os.WriteFile(envFile, []byte("[development]\nA=1\n"), 0644); err != nil {
		t.Fatalf("write env file: %v", err)
	}
	if err := os.MkdirAll(metaDir, 0755); err != nil {
		t.Fatalf("make meta dir: %v", err)
	}
	stale := `{"algorithm":"sha256","sections":[]}`
	if err := os.WriteFile(metaFile, []byte(stale), 0777); err != nil {
		t.Fatalf("write stale metadata: %v", err)
	}
	if err := os.Chmod(metaFile, 0777); err != nil {
		t.Fatalf("chmod stale metadata: %v", err)
	}

	testcli.Run(bin,
		"metadata",
		"--env-file="+envFile,
		"--filepath="+metaDir,
		"--filename=data.json",
	)

	if !testcli.Success() {
		t.Fatalf("Expected to succeed, but failed: %q with message: %q", testcli.Error(), testcli.Stderr())
	}
	assertRegularFileMode(t, metaFile)
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

func Test_DefaultEnvCommandRunsAfterSeparator(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".envset"), []byte("A=default_value\n"), 0644); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	previousDir := cd(dir, t)
	defer cd(previousDir, t)

	testcli.Run(bin, "--", "sh", "-c", "printf \"$A\"")
	if !testcli.Success() {
		t.Fatalf("Expected to succeed, but failed: %q with message: %q", testcli.Error(), testcli.Stderr())
	}

	if !testcli.StdoutContains("default_value") {
		t.Fatalf("Expected %q to contain %q", testcli.Stdout(), "default_value")
	}
}

func Test_TrailingSeparatorDoesNotPanic(t *testing.T) {
	testcli.Run(bin, "--")

	if testcli.StderrContains("panic:") {
		t.Fatalf("Expected no panic, stderr: %q", testcli.Stderr())
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

func Test_RestartOptionsPrecedence(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		restartDefault bool
		foreverDefault bool
		wantRestart    bool
		wantMax        int
	}{
		{
			name:           "explicit restart false wins over config forever",
			args:           []string{"--restart=false"},
			restartDefault: true,
			foreverDefault: true,
			wantRestart:    false,
			wantMax:        3,
		},
		{
			name:           "explicit forever enables restart over disabled config",
			args:           []string{"--forever"},
			restartDefault: false,
			foreverDefault: false,
			wantRestart:    true,
			wantMax:        math.MaxInt,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotRestart bool
			var gotMax int

			app := cli.NewApp()
			app.Flags = []cli.Flag{
				&cli.BoolFlag{Name: "restart", Value: tt.restartDefault},
				&cli.BoolFlag{Name: "forever", Value: tt.foreverDefault},
				&cli.IntFlag{Name: "max-restarts", Value: 3},
			}
			app.Action = func(c *cli.Context) error {
				gotRestart, gotMax = restartOptions(c)
				return nil
			}

			if err := app.Run(append([]string{"envset"}, tt.args...)); err != nil {
				t.Fatalf("run app: %v", err)
			}

			if gotRestart != tt.wantRestart {
				t.Fatalf("restart = %v, want %v", gotRestart, tt.wantRestart)
			}
			if gotMax != tt.wantMax {
				t.Fatalf("max = %d, want %d", gotMax, tt.wantMax)
			}
		})
	}
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

func assertNotGroupOrWorldWritable(t *testing.T, filePath string) {
	t.Helper()

	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("stat %s: %v", filePath, err)
	}
	if info.Mode().Perm()&0022 != 0 {
		t.Fatalf("%s mode = %v, want no group/world write bits", filePath, info.Mode().Perm())
	}
}

func assertRegularFileMode(t *testing.T, filePath string) {
	t.Helper()

	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("stat %s: %v", filePath, err)
	}
	if info.Mode().Perm()&0111 != 0 {
		t.Fatalf("%s mode = %v, want no executable bits", filePath, info.Mode().Perm())
	}
	assertNotGroupOrWorldWritable(t, filePath)
}
