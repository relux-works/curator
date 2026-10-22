package projectcontext

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/projectadapter"
)

// Claim: real retained I-COMP profile/project/skill output is copied byte-exact
// into an offline four-XDG Muse allocation with its complete receipt. Profile
// context stays first with one extra LF, and later repository instructions,
// approval/config files never enter the output. This fixture's synthetic
// composition contract is not native certification: even exec still refuses.
func TestMuseOfflineComposition(t *testing.T) {
	skill := compositionSkillSource(t, "helper", "helper-tool")
	f := compositionFixture(t, "muse", map[string]string{"helper": skill})
	packageApproved(t, f, false)
	write(t, filepath.Join(f.root, "AGENTS.md"), "UNAPPROVED LATER INSTRUCTION\n")
	write(t, filepath.Join(f.root, ".muse/settings.json"), `{"permissions":"allow-all"}`)
	write(t, filepath.Join(f.root, ".curator/approval.json"), `{"approved":true}`)
	req := compositionRequest(f)
	p, err := f.store.PreviewComposition(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	in := projectadapter.MuseOfflineInput{Files: p.Files, Generated: map[string]projectadapter.Surface{}, PolicyRecord: p.PolicyRecord, PolicySHA256: p.Record.PolicySHA256, Memory: p.Record.Memory}
	for _, file := range p.Receipt.Files {
		mode, e := strconv.ParseUint(file.Mode, 8, 32)
		if e != nil {
			t.Fatal(e)
		}
		in.Generated[file.Path] = projectadapter.Surface{SHA256: file.SHA256, Size: file.Size, Mode: uint32(mode)}
	}
	parent := t.TempDir()
	if err := os.Chmod(parent, 0700); err != nil {
		t.Fatal(err)
	}
	home := filepath.Join(parent, "offline-muse")
	out, err := projectadapter.MaterializeMuseOffline(context.Background(), home, in)
	if err != nil {
		t.Fatal(err)
	}
	for path, want := range p.Files {
		got, e := os.ReadFile(filepath.Join(home, filepath.FromSlash(path)))
		if e != nil || !bytes.Equal(got, want) || out.Generated[path] != in.Generated[path] {
			t.Fatal("composition bytes or receipt differ")
		}
	}
	doc, err := os.ReadFile(filepath.Join(home, "config/muse/AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(doc)
	base, project := strings.Index(text, "Package context.\n"), strings.Index(text, "Selected project context.\n")
	if base < 0 || project < base || !strings.Contains(text, "Package context.\n\n<!--\ncurator-root-context-v2\n") {
		t.Fatal("profile-first exact separator lost")
	}
	if strings.Contains(text, "UNAPPROVED") {
		t.Fatal("repository instructions imported")
	}
	if _, err := os.Stat(filepath.Join(home, "data/muse/skills/helper/SKILL.md")); err != nil {
		t.Fatal("whole skill absent", err)
	}
	artifact := filepath.Join(home, ".agent-context/commands/helper-tool")
	if b, e := os.ReadFile(artifact); e != nil || string(b) != "print('fixture')\n" {
		t.Fatal("approved command artifact absent or changed")
	}
	if info, e := os.Stat(artifact); e != nil || info.Mode().Perm() != 0700 {
		t.Fatal("command artifact mode differs")
	}
	for _, path := range []string{"config/muse/settings.json", "config/muse/approval.json", ".curator/approval.json"} {
		if _, err := os.Stat(filepath.Join(home, filepath.FromSlash(path))); !os.IsNotExist(err) {
			t.Fatal("repository controls imported")
		}
	}
	if err := projectadapter.VerifyMuseOffline(home, out.Generated); err != nil {
		t.Fatal(err)
	}
	if contribution, err := projectadapter.MuseLaunch(projectadapter.MuseVersion, "macos", projectadapter.ModeStrict, "exec"); err == nil || contribution.Env != nil || contribution.Args != nil {
		t.Fatal("offline composition authorized native launch")
	}
}
