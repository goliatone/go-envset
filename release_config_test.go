package main

import (
	"os"
	"strings"
	"testing"
)

func TestGoReleaserInjectsVersionMetadata(t *testing.T) {
	data, err := os.ReadFile(".goreleaser.yml")
	if err != nil {
		t.Fatalf("read .goreleaser.yml: %v", err)
	}

	config := string(data)
	for _, expected := range []string{
		"-X github.com/goliatone/go-envset/pkg/version.Tag={{ .Version }}",
		"-X github.com/goliatone/go-envset/pkg/version.Time={{ .Date }}",
		"-X github.com/goliatone/go-envset/pkg/version.User=goreleaser",
		"-X github.com/goliatone/go-envset/pkg/version.Commit={{ .FullCommit }}",
	} {
		if !strings.Contains(config, expected) {
			t.Fatalf("expected GoReleaser config to contain %q", expected)
		}
	}
}

func TestGoReleaserPublishesHomebrewCask(t *testing.T) {
	data, err := os.ReadFile(".goreleaser.yml")
	if err != nil {
		t.Fatalf("read .goreleaser.yml: %v", err)
	}

	config := string(data)
	if strings.Contains(config, "\nbrews:") {
		t.Fatal("expected GoReleaser config to use homebrew_casks instead of deprecated brews")
	}
	for _, expected := range []string{
		"\nhomebrew_casks:",
		"name: envset",
		"binaries:",
		"- envset",
		"directory: Casks",
	} {
		if !strings.Contains(config, expected) {
			t.Fatalf("expected GoReleaser cask config to contain %q", expected)
		}
	}
}
