package projectadapter

import (
	"context"
	"errors"
	"github.com/relux-works/curator/internal/stateread"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

// Claim (R3): unchanged empty, root-only and system-only profiles verify and
// reuse. Adding an absent root/override instruction is rejected without repair
// or receipt recapture; existing receipted roots retain byte-drift protection.
func TestCodexInstructionInventory(t *testing.T) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("qualified filesystem boundary")
	}
	for _, profile := range []string{"empty", "root", "system"} {
		for _, addition := range []string{"AGENTS.md", "AGENTS.override.md"} {
			t.Run(profile+"/"+addition, func(t *testing.T) {
				r := request(t)
				r.ProfileDocuments = map[string][]byte{}
				r.ProfileGenerated = map[string]Surface{}
				switch profile {
				case "root":
					r.ProfileDocuments["AGENTS.md"] = []byte("PROFILE_ROOT\n")
				case "system":
					r.ProfileDocuments[".agent-context/system-prompt.md"] = []byte("PROFILE_SYSTEM\n")
				}
				for name, b := range r.ProfileDocuments {
					r.ProfileGenerated[name] = surface(b, 0600)
				}
				p := render(t, r)
				home := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(r.Home)))))
				h := held(t, home)
				if err := p.Materialize(context.Background(), h, home); err != nil {
					t.Fatal(err)
				}
				if err := p.Verify(context.Background()); err != nil {
					t.Fatal("unchanged profile refused:", err)
				}
				if err := p.Materialize(context.Background(), h, home); err != nil {
					t.Fatal("unchanged reuse refused:", err)
				}
				target := filepath.Join(r.Home, addition)
				b := []byte("UNRECEIPTED_INSTRUCTION\n")
				if err := os.WriteFile(target, b, 0600); err != nil {
					t.Fatal(err)
				}
				baseline := p.Receipt()
				for _, check := range []func() error{func() error { return p.Verify(context.Background()) }, func() error { return p.Materialize(context.Background(), h, home) }} {
					if err := check(); err == nil || !strings.Contains(err.Error(), "project_context_drift") {
						t.Fatalf("instruction inventory drift accepted: %v", err)
					}
					got, err := os.ReadFile(target)
					if err != nil || string(got) != string(b) {
						t.Fatal("instruction drift repaired")
					}
					if !reflect.DeepEqual(baseline, p.Receipt()) {
						t.Fatal("instruction receipt recaptured")
					}
				}
			})
		}
	}
}

// Claim (R2): the independently pinned native bundle is published completely
// before startup, under the same immutable receipt and private modes as profile
// files. An admitted skill cannot replace that reserved native namespace.
func TestCodexSystemBundle(t *testing.T) {
	p := render(t, request(t))
	count := 0
	for name, b := range p.Files() {
		if strings.HasPrefix(name, "skills/.system/") {
			count++
			if p.Receipt()[name] != surface(b, 0600) {
				t.Fatal("native bundle not receipted")
			}
		}
	}
	if count != 49 {
		t.Fatalf("pinned native bundle missing: got %d files", count)
	}
	for _, name := range []string{"skills/.system", "skills/.system/skill-creator/SKILL.md", "skills/.SYSTEM/extra.md"} {
		r := request(t)
		r.Files[name] = []byte("PROJECT_REPLACEMENT")
		r.Generated[name] = surface(r.Files[name], 0600)
		if _, err := RenderCodex(r); err == nil || !strings.Contains(err.Error(), "reserved native skill namespace") {
			t.Fatalf("reserved native bundle replacement accepted: %v", err)
		}
	}
}

// Claim (R2): native bundle byte, marker, mode, missing, extra and linked-path
// drift refuse without repairs or receipt changes, next to unchanged reuse.
func TestCodexSystemBundleDrift(t *testing.T) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("qualified filesystem boundary")
	}
	for _, kind := range []string{"bytes", "marker", "mode", "missing", "extra", "file-link", "directory-link"} {
		t.Run(kind, func(t *testing.T) {
			r := request(t)
			p := render(t, r)
			home := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(r.Home)))))
			h := held(t, home)
			if err := p.Materialize(context.Background(), h, home); err != nil {
				t.Fatal(err)
			}
			if err := p.Materialize(context.Background(), h, home); err != nil {
				t.Fatal("unchanged bundle reuse refused:", err)
			}
			target := filepath.Join(r.Home, "skills/.system/skill-creator/SKILL.md")
			must := func(err error) {
				t.Helper()
				if err != nil {
					t.Fatal(err)
				}
			}
			switch kind {
			case "bytes":
				must(os.WriteFile(target, []byte("MALICIOUS_NATIVE_SKILL"), 0600))
			case "marker":
				must(os.WriteFile(filepath.Join(r.Home, "skills/.system/.codex-system-skills.marker"), []byte("wrong\n"), 0600))
			case "mode":
				must(os.Chmod(target, 0644))
			case "missing":
				must(os.Remove(target))
			case "extra":
				must(os.WriteFile(filepath.Join(r.Home, "skills/.system/extra.md"), []byte("EXTRA_NATIVE_SKILL"), 0600))
			case "file-link":
				must(os.Remove(target))
				must(os.Symlink("../../skill-installer/SKILL.md", target))
			case "directory-link":
				dir := filepath.Dir(target)
				moved := filepath.Join(t.TempDir(), "moved")
				must(os.Rename(dir, moved))
				must(os.Symlink(moved, dir))
			}
			before := bundleObservation(t, r.Home)
			receipt := p.Receipt()
			for _, check := range []func() error{func() error { return p.Verify(context.Background()) }, func() error { return p.Materialize(context.Background(), h, home) }} {
				if err := check(); err == nil || !strings.Contains(err.Error(), "project_context_drift") {
					t.Fatalf("native bundle drift accepted: %v", err)
				}
				if !reflect.DeepEqual(before, bundleObservation(t, r.Home)) {
					t.Fatal("native bundle drift repaired")
				}
				if !reflect.DeepEqual(receipt, p.Receipt()) {
					t.Fatal("native bundle receipt recaptured")
				}
			}
		})
	}
}

func bundleObservation(t *testing.T, home string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.WalkDir(home, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		st, err := entry.Info()
		if err != nil {
			return err
		}
		payload := ""
		if st.Mode()&os.ModeSymlink != 0 {
			payload, err = os.Readlink(path)
		} else if st.Mode().IsRegular() {
			var b []byte
			b, err = os.ReadFile(path)
			payload = string(b)
		}
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(home, path)
		if err != nil {
			return err
		}
		out[rel] = st.Mode().String() + payload
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// Claim: the embedded bundle digest rejects changed input before decoding;
// valid pinned bytes contain the expected native marker and full inventory.
func TestCodexBundleProvenance(t *testing.T) {
	files, err := codexSystemSkills(codexSystemBundle)
	if err != nil || len(files) != 49 || string(files["skills/.system/.codex-system-skills.marker"]) != "4af37f669ff79b4f\n" {
		t.Fatalf("pinned bundle differs: %v", err)
	}
	changed := append(append([]byte{}, codexSystemBundle...), 0)
	if _, err = codexSystemSkills(changed); err == nil || !strings.Contains(err.Error(), "pinned native bundle digest differs") {
		t.Fatalf("unpinned native bundle accepted: %v", err)
	}
}

// Claim (R4): the actual root-relative reader accepts genuine missing leaves
// while a broken-link or regular-file ancestor remains typed unreadable. The
// mandatory repository scanner separately proves production seam routing.
func TestCodexManagerMetadata(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink fixture requires Unix")
	}
	base := t.TempDir()
	if err := os.WriteFile(filepath.Join(base, "file"), []byte("fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("missing-target", filepath.Join(base, "broken")); err != nil {
		t.Fatal(err)
	}
	root, err := os.OpenRoot(base)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	for _, name := range []string{"absent", "missing-parent/absent"} {
		metadata, err := managerMetadata(root, base, name)
		if err != nil || metadata.Kind != stateread.KindAbsent {
			t.Fatalf("genuine absence refused: %v", err)
		}
	}
	for _, name := range []string{"broken/child", "file/child"} {
		metadata, err := managerMetadata(root, base, name)
		var readError *stateread.Error
		if !errors.As(err, &readError) || readError.Kind != stateread.KindUnreadable || metadata.Kind != stateread.KindUnreadable {
			t.Fatalf("unreadable ancestor collapsed to absence: %s: %v", name, err)
		}
	}
}

// Claim (R1): valid inputs still produce an inspectable immutable plan, while
// native-project permission certification, launch and prepared lease publication
// all refuse at this pin. Strict keeps its separate exclusion refusal.
func TestCodexNativeProjectRefusal(t *testing.T) {
	p := render(t, request(t))
	if len(p.Files()) == 0 || len(p.Launch().Args) == 0 {
		t.Fatal("valid preview lost")
	}
	var refusal *Unsupported
	if err := p.RequireLaunch(); !errors.As(err, &refusal) || refusal.Surface != Permissions || refusal.Code() != "project_context_unsupported" {
		t.Fatalf("incompatible native-project launch accepted: %v", err)
	}
	report := p.Report()
	for _, surface := range report.Certified {
		if surface == Permissions {
			t.Fatal("unsafe native-project permissions certified")
		}
	}
	found := false
	for _, r := range report.Refused {
		if r.Surface == Permissions && r.Reason == refusal.Reason {
			found = true
		}
	}
	if !found {
		t.Fatal("permission refusal missing from report")
	}
	if err := p.RequireStrict(); !errors.As(err, &refusal) || refusal.Surface != NativeSourceExclusion {
		t.Fatalf("strict refusal changed: %v", err)
	}
}
