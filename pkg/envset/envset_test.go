package envset

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Test_Run(t *testing.T) {

}

func Test_Print(t *testing.T) {

}

func Test_FileFinder(t *testing.T) {

	file, err := FileFinder(".goreleaser.yml")
	if err != nil {
		t.Errorf("FileFinder failed, unexpected error %v", err)
	}

	if filepath.IsAbs(file) == false {
		t.Errorf("FilePath should return absolute path: %s", file)
	}
}

func Test_FileFinder_Error(t *testing.T) {
	_, err := FileFinder("random_file.404")
	if err == nil {
		t.Error("FileFinder failed, expected error")
	}
}

func Test_Run_RequiredKeysPresent(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".envset")
	if err := os.WriteFile(envFile, []byte("[development]\nEXPECTED=envset_result\n"), 0644); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	err := Run("development", RunOptions{
		Filename:      envFile,
		Cmd:           "sh",
		Args:          []string{"-c", "test \"$EXPECTED\" = envset_result"},
		Isolated:      true,
		ExportEnvName: "APP_ENV",
		Required:      []string{"EXPECTED"},
	})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
}

func Test_Run_RequiredKeysMissingOnlyReportsMissing(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".envset")
	if err := os.WriteFile(envFile, []byte("[development]\nEXPECTED=envset_result\n"), 0644); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	err := Run("development", RunOptions{
		Filename:      envFile,
		Cmd:           "sh",
		Args:          []string{"-c", "exit 0"},
		Isolated:      true,
		ExportEnvName: "APP_ENV",
		Required:      []string{"EXPECTED", "MISSING"},
	})
	if err == nil {
		t.Fatal("expected required key error")
	}
	if got := err.Error(); !strings.Contains(got, "missing required keys: MISSING") {
		t.Fatalf("error = %q, want missing MISSING only", got)
	}
	if strings.Contains(err.Error(), ": ,") {
		t.Fatalf("error includes blank missing key: %q", err.Error())
	}
}

func Test_Run_UnsortedSectionNames(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".envset")
	contents := []byte("[z]\nA=z\n[m]\nA=m\n[a]\nA=a\n")
	if err := os.WriteFile(envFile, contents, 0644); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	err := Run("z", RunOptions{
		Filename:      envFile,
		Cmd:           "sh",
		Args:          []string{"-c", "test \"$A\" = z"},
		Isolated:      true,
		ExportEnvName: "APP_ENV",
	})
	if err != nil {
		t.Fatalf("run unsorted section: %v", err)
	}
}

func Test_EnvFileLoadPersistsState(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".envset")
	if err := os.WriteFile(envFile, []byte("[development]\nA=1\n"), 0644); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	env := EnvFile{}
	if err := env.Load(envFile); err != nil {
		t.Fatalf("load: %v", err)
	}
	if env.Path != envFile {
		t.Fatalf("path = %q, want %q", env.Path, envFile)
	}
	if env.File == nil {
		t.Fatal("file was not persisted")
	}
}

func Test_CompareMetadataFiles(t *testing.T) {
	tests := []struct {
		name        string
		source      *EnvFile
		target      *EnvFile
		wantChanged bool
		wantErr     bool
	}{
		{
			name: "same sections and hashes unchanged",
			source: metadataFixture(HashSHA256, map[string]string{
				"development": "abc",
			}),
			target: metadataFixture(HashSHA256, map[string]string{
				"development": "abc",
			}),
			wantChanged: false,
		},
		{
			name: "same count different section names changed",
			source: metadataFixture(HashSHA256, map[string]string{
				"production": "abc",
			}),
			target: metadataFixture(HashSHA256, map[string]string{
				"staging": "abc",
			}),
			wantChanged: true,
		},
		{
			name: "added section changed",
			source: metadataFixture(HashSHA256, map[string]string{
				"development": "abc",
				"production":  "def",
			}),
			target: metadataFixture(HashSHA256, map[string]string{
				"development": "abc",
			}),
			wantChanged: true,
		},
		{
			name: "different hash changed",
			source: metadataFixture(HashSHA256, map[string]string{
				"development": "abc",
			}),
			target: metadataFixture(HashSHA256, map[string]string{
				"development": "def",
			}),
			wantChanged: true,
		},
		{
			name: "algorithm mismatch errors",
			source: metadataFixture(HashMD5, map[string]string{
				"development": "abc",
			}),
			target: metadataFixture(HashSHA256, map[string]string{
				"development": "abc",
			}),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changed, err := CompareMetadataFiles(tt.source, tt.target)
			if tt.wantErr {
				var wrongAlgorithm *ErrorWrongAlgorithm
				if !errors.As(err, &wrongAlgorithm) {
					t.Fatalf("err = %v, want ErrorWrongAlgorithm", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("compare: %v", err)
			}
			if changed != tt.wantChanged {
				t.Fatalf("changed = %v, want %v", changed, tt.wantChanged)
			}
		})
	}
}

func metadataFixture(algorithm string, sections map[string]string) *EnvFile {
	envFile := &EnvFile{
		Algorithm: algorithm,
		Sections:  make([]*EnvSection, 0, len(sections)),
	}
	for name, hash := range sections {
		envFile.Sections = append(envFile.Sections, &EnvSection{
			Name: name,
			Keys: []*EnvKey{
				{Name: "KEY", Hash: hash},
			},
		})
	}
	return envFile
}
