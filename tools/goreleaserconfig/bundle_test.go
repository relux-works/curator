package goreleaserconfig

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// The committed build must use the launcher's pinned version and toolchain,
// and must never cross-build the Unix-only launcher for Windows.
func TestLauncherPinMatchesBuild(t *testing.T) {
	root := filepath.Join("..", "..")
	raw, err := os.ReadFile(filepath.Join(root, "release", "curator-run.json"))
	if err != nil {
		t.Fatal(err)
	}
	var pin struct {
		Version   string `json:"version"`
		GoVersion string `json:"go_version"`
	}
	if err := json.Unmarshal(raw, &pin); err != nil {
		t.Fatal(err)
	}
	raw, err = os.ReadFile(filepath.Join(root, ".goreleaser.yml"))
	if err != nil {
		t.Fatal(err)
	}
	var config struct {
		Before struct {
			Hooks []string `yaml:"hooks"`
		} `yaml:"before"`
		Builds []struct {
			ID                        string `yaml:"id"`
			Dir                       string `yaml:"dir"`
			Env, Goos, Flags, Ldflags []string
		} `yaml:"builds"`
		AURs []struct {
			Name string `yaml:"name"`
			Skip string `yaml:"skip_upload"`
		} `yaml:"aurs"`
		Signs []struct {
			Cmd       string   `yaml:"cmd"`
			Args      []string `yaml:"args"`
			Artifacts string   `yaml:"artifacts"`
		} `yaml:"signs"`
		Notarize struct {
			MacOS []struct {
				Enabled string   `yaml:"enabled"`
				IDs     []string `yaml:"ids"`
			} `yaml:"macos"`
		} `yaml:"notarize"`
	}
	if err := yaml.Unmarshal(raw, &config); err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(config.Before.Hooks, "python3 release/bundle.py prepare") {
		t.Fatal("signed launcher preparation is absent")
	}
	found := false
	for _, b := range config.Builds {
		if b.ID != "curator-run" {
			continue
		}
		found = true
		if b.Dir != ".tmp/curator-run" || !slices.Equal(b.Goos, []string{"darwin", "linux"}) {
			t.Fatalf("launcher source/platform mismatch: %+v", b)
		}
		for _, flag := range []string{"-trimpath", "-buildvcs=false"} {
			if !slices.Contains(b.Flags, flag) {
				t.Errorf("canonical launcher flag missing: %s", flag)
			}
		}
		for _, value := range []string{"CGO_ENABLED=0", "GOFLAGS=-mod=readonly", "GOTOOLCHAIN=go" + pin.GoVersion} {
			if !slices.Contains(b.Env, value) {
				t.Errorf("pinned launcher environment missing: %s", value)
			}
		}
		if strings.Join(b.Ldflags, " ") != "-s -w -buildid= -X main.buildVersion="+pin.Version {
			t.Fatal("launcher version/flags differ from pin")
		}
	}
	if !found {
		t.Fatal("launcher build is absent")
	}
	if len(config.AURs) != 1 || config.AURs[0].Name != "curator-bin" || config.AURs[0].Skip != "auto" {
		t.Fatal("AUR must preserve the stable channel for prereleases")
	}
	// These are the production wiring, not just tests of an unused helper:
	// checksum signing must pass the payload gate and notarization must select both binaries.
	if len(config.Signs) != 1 || config.Signs[0].Cmd != "bash" || config.Signs[0].Artifacts != "checksum" ||
		!slices.Equal(config.Signs[0].Args, []string{"release/sign-checksums.sh", "${artifact}", "${certificate}", "${signature}"}) {
		t.Fatal("checksum signing bypasses the payload gate")
	}
	mac := config.Notarize.MacOS
	if len(mac) != 1 || mac[0].Enabled != "{{ not .IsSnapshot }}" || !slices.Equal(mac[0].IDs, []string{"curator", "curator-run"}) {
		t.Fatal("real releases must sign and notarize both Darwin binaries")
	}
}
