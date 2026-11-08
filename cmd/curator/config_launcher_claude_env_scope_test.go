package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestConfigLauncherClaudeEnvLegacyDiscovery(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("synthetic POSIX provider")
	}
	root := t.TempDir()
	path := filepath.Join(root, "selected.json")
	if err := os.WriteFile(filepath.Join(root, "curator-zzprobe"), []byte("#!/bin/sh\nprintf 'PROVIDER_EXECUTED\\n'\n"), 0700); err != nil {
		t.Fatal(err)
	}
	for _, schema := range []int{1, 2, 3, 5, 7, 8} {
		for _, defect := range []string{"malformed", "semantic"} {
			t.Run(fmt.Sprintf("schema-%d-%s", schema, defect), func(t *testing.T) {
				text := fmt.Sprintf(`{"schema_version":%d,"skills_root":"x","projects":{},}`, schema)
				want := "malformed JSON in config " + path
				if defect == "semantic" {
					text = fmt.Sprintf(`{"schema_version":%d,"skills_root":"x","projects":{},"unknown":"DO_NOT_ECHO"}`, schema)
					want = "config: unsupported top-level configuration field"
				}
				if err := os.WriteFile(path, []byte(text), 0600); err != nil {
					t.Fatal(err)
				}
				var out, errs bytes.Buffer
				command := cli{config: fileConfigSource(path), stdout: &out, stderr: &errs, providerPathOverride: &root, userHome: func() (string, error) { return root, nil }}
				if code := command.run([]string{"zzprobe"}); code != exitOK || out.String() != "PROVIDER_EXECUTED\n" || !strings.HasPrefix(errs.String(), "curator: "+want+"\n") || strings.Contains(errs.String(), "DO_NOT_ECHO") {
					t.Fatalf("legacy discovery changed: exit/output %q, diagnostic %q", out.String(), errs.String())
				}
				after, err := os.ReadFile(path)
				if err != nil || string(after) != text {
					t.Fatal("discovery changed source")
				}
			})
		}
	}
}

func TestConfigLauncherClaudeEnvMissingSystemHint(t *testing.T) {
	path := filepath.Join(t.TempDir(), "selected.json")
	missing := filepath.Join(t.TempDir(), "missing-system.json")
	t.Setenv("CURATOR_SYSTEM_CONFIG", missing)
	for _, extensions := range []string{"", `,"extensions":{"claude_env":{"BASH_MAX_OUTPUT_LENGTH":"8000"}}`} {
		text := `{"schema_version":9,"skills_root":"x","projects":{}` + extensions + `}`
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
		var out, errs bytes.Buffer
		command := cli{config: fileConfigSource(path), stdout: &out, stderr: &errs}
		if code := command.run([]string{"config", "launcher"}); code != exitFail || out.Len() != 0 ||
			errs.String() != "curator: global config not found: "+missing+"\ncurator: "+configNotFoundHint+"\n" || strings.Contains(errs.String(), "claude_env") {
			t.Fatalf("missing-system hint/path lost: %q", errs.String())
		}
	}
}
