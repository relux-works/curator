package main

import (
	"os"
	"reflect"
	"slices"
	"testing"

	"github.com/relux-works/curator/internal/config"
)

// Each configured declaration must keep its own fields in the list, including
// equal source/ref keys. Listing must preserve both declaration and file state.
func TestProfileComposeListPreservesDeclarations(t *testing.T) {
	weight10, weight20 := 10, 20
	const repo = "https://example.com/policies.git"
	tests := []struct {
		name  string
		decls []config.OverlayDeclaration
		want  string
	}{
		{
			name: "same repository and revision, different directories",
			decls: []config.OverlayDeclaration{
				{Source: repo, Revision: "abc123", Directory: "profiles/first", Weight: &weight10},
				{Source: repo, Revision: "abc123", Directory: "profiles/second", Weight: &weight20},
			},
			want: repo + "\trevision=abc123\tprofiles/first\t10\n" + repo + "\trevision=abc123\tprofiles/second\t20\n",
		},
		{
			name: "same directory, different weights",
			decls: []config.OverlayDeclaration{
				{Source: repo, Tag: "v1.0.0", Directory: "profiles/shared", Weight: &weight10},
				{Source: repo, Tag: "v1.0.0", Directory: "profiles/shared", Weight: &weight20},
			},
			want: repo + "\ttag=v1.0.0\tprofiles/shared\t10\n" + repo + "\ttag=v1.0.0\tprofiles/shared\t20\n",
		},
		{
			name: "local paths retain sorted order and default weight",
			decls: []config.OverlayDeclaration{
				{Source: "./second", Weight: &weight20},
				{Source: "./first"},
			},
			want: "./first\tpath\t-\t(default)\n./second\tpath\t-\t20\n",
		},
		{
			name: "distinct requirement keys retain existing ordering",
			decls: []config.OverlayDeclaration{
				{Source: repo, Range: "^1"},
				{Source: repo, Tag: "v1.0.0"},
				{Source: repo, Revision: "abc123"},
			},
			want: repo + "\trevision=abc123\t-\t(default)\n" + repo + "\ttag=v1.0.0\t-\t(default)\n" + repo + "\trange=^1\t-\t(default)\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := writeMachineConfig(t, `{"schema_version":2,"skills_root":"x","projects":{}}`)
			source.cfg.Env.Overlays["sample"] = tt.decls
			before := slices.Clone(tt.decls)
			fileBefore, err := os.ReadFile(source.path)
			if err != nil {
				t.Fatal(err)
			}
			code, stdout, stderr := runProfile(t, source, "profile", "compose", "sample", "list")
			if code != exitOK || stdout != tt.want || stderr != "" {
				t.Fatalf("list = %d\nstdout: %q\nwant: %q\nstderr: %s", code, stdout, tt.want, stderr)
			}
			if !reflect.DeepEqual(tt.decls, before) {
				t.Fatalf("list mutated declarations: got %+v, want %+v", tt.decls, before)
			}
			fileAfter, err := os.ReadFile(source.path)
			if err != nil {
				t.Fatal(err)
			}
			if string(fileAfter) != string(fileBefore) {
				t.Fatal("list changed config file")
			}
		})
	}
}
