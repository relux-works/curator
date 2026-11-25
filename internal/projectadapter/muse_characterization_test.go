package projectadapter

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// Keep the existing offline project-context contract observable while the
// harness API migration proceeds independently.
func TestMuseKeptOfflineProjectContextCharacterization(t *testing.T) {
	in := offlineFixture()
	home := offlineHome(t)
	got, err := MaterializeMuseOffline(context.Background(), home, in)
	if err != nil {
		t.Fatal(err)
	}
	wantParents := map[string]string{
		"XDG_CONFIG_HOME": filepath.Join(home, "config"),
		"XDG_DATA_HOME":   filepath.Join(home, "data"),
		"XDG_STATE_HOME":  filepath.Join(home, "state"),
		"XDG_CACHE_HOME":  filepath.Join(home, "cache"),
	}
	if !reflect.DeepEqual(got.Parents, wantParents) || !reflect.DeepEqual(got.Generated, in.Generated) || got.Mode != ModeStrict || got.Memory != "off" || !bytes.Equal(got.PolicyRecord, in.PolicyRecord) {
		t.Fatalf("offline project result changed: %+v", got)
	}
	for name, want := range in.Files {
		actual, err := os.ReadFile(filepath.Join(home, filepath.FromSlash(name)))
		if err != nil || !bytes.Equal(actual, want) {
			t.Fatalf("offline file %s: %q err=%v", name, actual, err)
		}
	}
	if err := VerifyMuseOffline(home, got.Generated); err != nil {
		t.Fatalf("offline verification: %v", err)
	}
}

// Required native MCP still refuses before creating the offline allocation.
func TestMuseKeptRequiredMCPRefusalCharacterization(t *testing.T) {
	in := offlineFixture()
	in.MCPItems = []MuseMCPItem{{ItemID: "required-source", Required: true}}
	home := offlineHome(t)
	got, err := MaterializeMuseOffline(context.Background(), home, in)
	var unsupported *Unsupported
	if err == nil || got.Home != "" || !errors.As(err, &unsupported) || unsupported.Surface != MCP || unsupported.Reason != "native_writer_unqualified" {
		t.Fatalf("required MCP refusal changed: result=%+v err=%v", got, err)
	}
	if _, err := os.Lstat(home); !os.IsNotExist(err) {
		t.Fatalf("required MCP refusal allocated a home: %v", err)
	}
}

// Native launch remains refused until the missing v1 channels are qualified.
func TestMuseKeptNativeLaunchRefusalCharacterization(t *testing.T) {
	got, err := MuseLaunch(MuseVersion, "macos", ModeStrict, "exec")
	var unsupported *Unsupported
	if err == nil || got.Env != nil || len(got.Args) != 0 || !errors.As(err, &unsupported) || unsupported.Surface != NativeSourceExclusion || unsupported.Reason != "complete_project_source_exclusion_unqualified" {
		t.Fatalf("native launch refusal changed: result=%+v err=%v", got, err)
	}
}
