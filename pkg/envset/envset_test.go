package envset

import (
	"path/filepath"
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
