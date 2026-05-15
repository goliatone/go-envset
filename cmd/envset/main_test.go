package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	osexec "os/exec"
	"path"
	"path/filepath"
	"testing"

	"github.com/rendon/testcli"
	"github.com/stretchr/testify/assert"
)

var bin string

func init() {
	cur, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("get wd: %v", err))
	}
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
	if err := os.Remove(bin); err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "remove test bin: %v\n", err)
		code = 1
	}
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

func Test_EnvironmentGlobalRunFlagsOverrideCommandDefaults(t *testing.T) {
	dir := setupPrecedenceTestDir(t)
	previousDir := cd(dir, t)
	defer cd(previousDir, t)

	t.Setenv("ENVSET_HOST_ONLY", "host-value")

	tests := []struct {
		name    string
		args    []string
		wantOK  bool
		wantOut string
	}{
		{
			name:   "env file",
			args:   []string{"--env-file=.custom-envset", "development", "--restart=false", "--", "sh", "-c", "test \"$A\" = custom"},
			wantOK: true,
		},
		{
			name:   "isolated false",
			args:   []string{"--isolated=false", "development", "--restart=false", "--", "sh", "-c", "test \"$ENVSET_HOST_ONLY\" = host-value"},
			wantOK: true,
		},
		{
			name:   "inherit",
			args:   []string{"--inherit=ENVSET_HOST_ONLY", "development", "--restart=false", "--", "sh", "-c", "test \"$ENVSET_HOST_ONLY\" = host-value"},
			wantOK: true,
		},
		{
			name:   "required",
			args:   []string{"--required=MISSING_REQUIRED", "development", "--restart=false", "--", "true"},
			wantOK: false,
		},
		{
			name:   "export env name",
			args:   []string{"--export-env-name=ENVSET_ENV", "development", "--restart=false", "--", "sh", "-c", "test \"$ENVSET_ENV\" = development"},
			wantOK: true,
		},
		{
			name:   "expand false",
			args:   []string{"--expand=false", "development", "--restart=false", "--", "sh", "-c", "test \"$B\" = '${ENVSET_HOST_ONLY}'"},
			wantOK: true,
		},
		{
			name:   "local env file wins over global",
			args:   []string{"--env-file=.custom-envset", "development", "--env-file=.local-envset", "--restart=false", "--", "sh", "-c", "test \"$A\" = local"},
			wantOK: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testcli.Run(bin, tt.args...)
			if tt.wantOK && !testcli.Success() {
				t.Fatalf("Expected to succeed, stdout: %q stderr: %q error: %q", testcli.Stdout(), testcli.Stderr(), testcli.Error())
			}
			if !tt.wantOK && testcli.Success() {
				t.Fatalf("Expected to fail, stdout: %q stderr: %q", testcli.Stdout(), testcli.Stderr())
			}
			if tt.wantOut != "" && !testcli.StdoutContains(tt.wantOut) {
				t.Fatalf("Expected stdout %q to contain %q", testcli.Stdout(), tt.wantOut)
			}
		})
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

func Test_TemplateGlobalEnvFileOverridesCommandDefault(t *testing.T) {
	dir := setupPrecedenceTestDir(t)
	previousDir := cd(dir, t)
	defer cd(previousDir, t)

	testcli.Run(bin, "--env-file=.custom-envset", "template", "--print")

	if !testcli.Success() {
		t.Fatalf("Expected to succeed, stdout: %q stderr: %q error: %q", testcli.Stdout(), testcli.Stderr(), testcli.Error())
	}
	if !testcli.StdoutContains("CUSTOM_ONLY={{CUSTOM_ONLY}}") {
		t.Fatalf("Expected stdout %q to contain custom template key", testcli.Stdout())
	}
	if testcli.StdoutContains("DEFAULT_ONLY={{DEFAULT_ONLY}}") {
		t.Fatalf("Expected stdout %q to not contain default template key", testcli.Stdout())
	}
}

func Test_MetadataGlobalEnvFileOverridesCommandDefault(t *testing.T) {
	dir := setupPrecedenceTestDir(t)
	previousDir := cd(dir, t)
	defer cd(previousDir, t)

	testcli.Run(bin, "--env-file=.custom-envset", "metadata", "--print", "--values")

	if !testcli.Success() {
		t.Fatalf("Expected to succeed, stdout: %q stderr: %q error: %q", testcli.Stdout(), testcli.Stderr(), testcli.Error())
	}
	if !testcli.StdoutContains("CUSTOM_ONLY") {
		t.Fatalf("Expected stdout %q to contain custom metadata key", testcli.Stdout())
	}
	if testcli.StdoutContains("DEFAULT_ONLY") {
		t.Fatalf("Expected stdout %q to not contain default metadata key", testcli.Stdout())
	}
}

func rm(dir string, t *testing.T) {
	err := os.RemoveAll(dir)
	if err != nil {
		t.Fatalf("error removing dir: %q ", err)
	}
}

func cd(dir string, t *testing.T) string {
	t.Helper()

	cur, err := os.Getwd()
	if err != nil {
		t.Fatalf("error cd dir: %q ", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("error cd dir %s: %q ", dir, err)
	}
	return cur
}

func md5sum(filePath string, t *testing.T) string {
	t.Helper()

	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("error removing dir: %q ", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			t.Fatalf("error closing file %s: %q", filePath, err)
		}
	}()

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

func setupPrecedenceTestDir(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, ".envsetrc"), `filename=.envset
expand=true
isolated=true
export_environment=APP_ENV
restart=true
max_restarts=3
restart_forever=false

[environments]
name=development
`)
	writeFile(t, filepath.Join(dir, ".envset"), `[development]
A=default
B=${ENVSET_HOST_ONLY}
DEFAULT_ONLY=1
`)
	writeFile(t, filepath.Join(dir, ".custom-envset"), `[development]
A=custom
CUSTOM_ONLY=1
`)
	writeFile(t, filepath.Join(dir, ".local-envset"), `[development]
A=local
LOCAL_ONLY=1
`)
	cmd := osexec.Command("git", "init")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init %s: %v\n%s", dir, err, out)
	}
	cmd = osexec.Command("git", "remote", "add", "origin", "https://example.com/envset-test.git")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git remote add %s: %v\n%s", dir, err, out)
	}

	return dir
}

func writeFile(t *testing.T, filename, content string) {
	t.Helper()
	if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", filename, err)
	}
}
