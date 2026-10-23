package projectcontext

import (
	"bytes"
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextmaterialize"
	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/projectadapter"
)

type claudeProfileFixture struct {
	files    map[string][]byte
	receipt  []File
	called   bool
	err      error
	ref, sha string
}

func (s *claudeProfileFixture) ProfileDocuments(_ context.Context, ref, sha string) (map[string][]byte, []File, error) {
	s.called = true
	if ref != s.ref || sha != s.sha {
		return nil, nil, errors.New("profile source binding differs")
	}
	return s.files, s.receipt, s.err
}

func profileFixture(t *testing.T, p CompositionPreview, files map[string][]byte) *claudeProfileFixture {
	t.Helper()
	r, err := compositionReceipt(files, nil, p.Receipt.AdapterSHA256, p.Receipt.PolicySHA256, p.Receipt.WorktreeRef)
	if err != nil {
		t.Fatal(err)
	}
	return &claudeProfileFixture{files: files, receipt: r.Files, ref: p.Ref, sha: p.SHA256}
}

// Claim: offline native-project preparation uses only the dedicated profile
// document source, including empty/root/system variants. It retains whole
// admitted skills and MCP, recomputes its candidate receipt, and does not modify
// the strict composition. This fixture does not claim I-COMP r3 authority or
// lease publication; the native launch remains refused independently below.
func TestClaudeNativeProjectPreparation(t *testing.T) {
	_, req, p, _ := claudeFixture(t, false, true)
	original := append([]byte{}, p.Files["CLAUDE.md"]...)
	for _, docs := range []map[string][]byte{
		{"CLAUDE.md": []byte("Profile-only root.\n")},
		{"CLAUDE.md": []byte(""), ".agent-context/system-prompt.md": []byte("Profile system only.\n")},
		{},
	} {
		source := profileFixture(t, p, docs)
		out, err := PrepareClaudeMode(context.Background(), projectadapter.NativeProject, p, *req.Claude, source)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(out.Files["CLAUDE.md"], docs["CLAUDE.md"]) {
			t.Fatal("native profile-only root differs")
		}
		if !reflect.DeepEqual(out.Files[".agent-context/system-prompt.md"], docs[".agent-context/system-prompt.md"]) {
			t.Fatal("native profile-only system differs")
		}
		if !bytes.Equal(out.Files["skills/manual/references/manual.md"], p.Files["skills/manual/references/manual.md"]) || !bytes.Equal(out.Files[contextmaterialize.MCPClaudePath], p.Files[contextmaterialize.MCPClaudePath]) {
			t.Fatal("native approved skill or MCP lost")
		}
		if out.Certification.Mode != projectadapter.NativeProject || out.Certification.Notice != projectadapter.NativeProjectNotice || out.Certification.Form != "managed-home" {
			t.Fatal("native mode notice missing")
		}
		if out.Receipt.PolicySHA256 == p.Receipt.PolicySHA256 || identity(out.Receipt) == identity(p.Receipt) {
			t.Fatal("native candidate receipt reused strict authority")
		}
		if len(out.Receipt.Files) != len(out.Files) {
			t.Fatal("native candidate receipt set differs")
		}
		for _, f := range out.Receipt.Files {
			if b, ok := out.Files[f.Path]; !ok || int64(len(b)) != f.Size || digest(b) != f.SHA256 {
				t.Fatal("native candidate receipt bytes differ")
			}
		}
		for path, b := range out.Files {
			if bytes.Contains(b, []byte("Project document.")) || bytes.Contains(b, []byte("Unapproved")) {
				t.Fatalf("native project prose copied: %s", path)
			}
		}
		if len(out.Files["settings.json"]) == 0 || len(out.Files[claudeOwnershipPath]) == 0 {
			t.Fatal("native manager policy missing")
		}
		out.Files["skills/manual/references/manual.md"][0] ^= 1
		if bytes.Equal(out.Files["skills/manual/references/manual.md"], p.Files["skills/manual/references/manual.md"]) {
			t.Fatal("approved skill bytes aliased")
		}
		if len(out.Files["CLAUDE.md"]) > 0 {
			out.Files["CLAUDE.md"][0] ^= 1
			if bytes.Equal(out.Files["CLAUDE.md"], docs["CLAUDE.md"]) {
				t.Fatal("profile bytes aliased")
			}
		}
	}
	if !bytes.Equal(p.Files["CLAUDE.md"], original) {
		t.Fatal("strict composition changed")
	}
}

// Claim: missing dedicated profile input, malformed/drifting profile receipts,
// foreign paths and inconsistent base files fail with no candidate. Successful
// dedicated input is covered next to these cases, so rejecting all is detected.
func TestClaudeNativeProjectInputs(t *testing.T) {
	_, req, p, _ := claudeFixture(t, false, false)
	for _, name := range []string{"missing", "unavailable", "drift", "extra", "foreign", "mode", "duplicate", "source-error", "base-drift", "policy-drift", "cancelled", "mode-empty", "mode-unknown", "binding-ref", "binding-sha", "binding-materialization", "binding-environment"} {
		t.Run(name, func(t *testing.T) {
			source := profileFixture(t, p, map[string][]byte{"CLAUDE.md": []byte("Only profile.\n")})
			var input ClaudeProfileDocumentSource = source
			base := p
			ctx := context.Background()
			mode := projectadapter.NativeProject
			code := "project_context_invalid"
			switch name {
			case "missing":
				input = nil
				code = "project_context_unsupported"
			case "unavailable":
				input = ClaudeProfileDocumentsUnavailable{}
				code = "project_context_unsupported"
			case "drift":
				source.files["CLAUDE.md"][0] ^= 1
			case "extra":
				source.files["extra.md"] = []byte("Extra")
			case "foreign":
				source = profileFixture(t, p, map[string][]byte{"settings.json": []byte("{}")})
				input = source
			case "mode":
				source.receipt[0].Mode = "0700"
			case "duplicate":
				source.receipt = append(source.receipt, source.receipt[0])
			case "source-error":
				source.err = errors.New("fixture source unavailable")
				code = "project_context_unreadable"
			case "base-drift":
				base.Files = cloneClaudeFiles(p.Files)
				base.Files["CLAUDE.md"][0] ^= 1
			case "policy-drift":
				changed := *req.Claude
				changed.AllowTools = []string{"Read"}
				out, err := PrepareClaudeMode(ctx, projectadapter.NativeProject, base, changed, source)
				required(t, err, code)
				if out.Files != nil {
					t.Fatal("invalid native input produced candidate")
				}
				return
			case "mode-empty":
				mode = ""
			case "mode-unknown":
				mode = "unknown"
			case "binding-ref":
				base.Ref = "composition:other"
				source.ref = base.Ref
			case "binding-sha":
				base.SHA256 = digest([]byte("other"))
				base.Ref = "composition:" + hashID(base.SHA256)
				source.ref = base.Ref
				source.sha = base.SHA256
			case "binding-materialization":
				base.Record.MaterializationSHA256 = digest([]byte("other"))
				base.SHA256 = identity(base.Record)
				base.Ref = "composition:" + hashID(base.SHA256)
				source.ref = base.Ref
				source.sha = base.SHA256
			case "binding-environment":
				base.Record.Environment = "muse"
				base.SHA256 = identity(base.Record)
				base.Ref = "composition:" + hashID(base.SHA256)
				source.ref = base.Ref
				source.sha = base.SHA256
			case "cancelled":
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}
			out, err := PrepareClaudeMode(ctx, mode, base, *req.Claude, input)
			if name == "cancelled" {
				if !errors.Is(err, context.Canceled) {
					t.Fatal(err)
				}
			} else {
				required(t, err, code)
			}
			if out.Files != nil {
				t.Fatal("invalid native input produced candidate")
			}
		})
	}
}

// Claim: explicit native-project status names both missing mandatory native
// qualifications and carries the operator notice. No mode, strict, unknown,
// resume or fork can obtain runnable argv/environment through the mode gate.
// The offline preparation positive above remains available despite refusal.
func TestClaudeNativeProjectQualification(t *testing.T) {
	r := ClaudeModeCertificationReport(ClaudeProjectVersion, "macos", "arm64", projectadapter.NativeProject)
	if len(r.Certified) != 0 || len(r.Refused) != 2 || r.Refused[0].Surface != projectadapter.Permissions || r.Refused[1].Surface != projectadapter.MemoryOff {
		t.Fatal("native mandatory qualifications misreported")
	}
	if r.Notice != projectadapter.NativeProjectNotice || r.Mode != projectadapter.NativeProject || r.Form != "managed-home" {
		t.Fatal("native status notice missing")
	}
	for _, mode := range []string{"", "unknown", projectadapter.Strict, projectadapter.NativeProject} {
		for _, entry := range []ClaudeEntry{ClaudeFresh, ClaudeResume, ClaudeFork} {
			out, err := (Authority{}).ClaudeLaunchMode(context.Background(), managerlock.Held{}, "", "", "", entry, mode)
			var diagnostic *ClaudeDiagnostic
			if !errors.As(err, &diagnostic) || out.Args != nil || out.Env != nil {
				t.Fatalf("native launch qualification bypass: %v", err)
			}
			want := "project_context_unsupported"
			if mode == "" || mode == "unknown" {
				want = "project_context_invalid"
			}
			if diagnostic.Code != want {
				t.Fatalf("native mode diagnostic differs: %v", err)
			}
			if mode == projectadapter.NativeProject && (diagnostic.Notice != projectadapter.NativeProjectNotice || len(diagnostic.Refused) != 2 || !strings.Contains(err.Error(), projectadapter.NativeProjectNotice)) {
				t.Fatal("native launch refusal lost notice or missing qualifications")
			}
		}
	}
	for _, tuple := range [][3]string{{"2.1.306", "macos", "arm64"}, {ClaudeProjectVersion, "linux", "arm64"}, {ClaudeProjectVersion, "macos", "amd64"}} {
		other := ClaudeModeCertificationReport(tuple[0], tuple[1], tuple[2], projectadapter.NativeProject)
		if len(other.Certified) != 0 || len(other.Refused) < 2 {
			t.Fatal("native unknown tuple borrowed qualification")
		}
	}
	source := &claudeProfileFixture{}
	for _, mode := range []string{"", "unknown", projectadapter.Strict} {
		out, err := PrepareClaudeMode(context.Background(), mode, CompositionPreview{}, ClaudeControls{}, source)
		if err == nil || out.Files != nil || source.called {
			t.Fatal("mode opt-in bypass")
		}
	}
}

// Claim: exact file sets, including empty files and executable skill bytes,
// validate; missing/extra bytes, size/hash drift, invalid paths/modes and aliases
// fail with the inventory diagnostic. Grammar-only validation is insufficient.
func TestClaudeNativeFileSet(t *testing.T) {
	for _, files := range []map[string][]byte{{}, {"CLAUDE.md": {}}, {"skills/manual/script": []byte("echo fixture")}} {
		r, err := compositionReceipt(files, map[string]bool{"skills/manual/script": true}, "", "", "")
		if err != nil {
			t.Fatal(err)
		}
		if err = validateClaudeFileSet(files, r.Files); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"extra", "missing", "size", "hash", "mode", "path", "alias", "bound"} {
		t.Run(name, func(t *testing.T) {
			files := map[string][]byte{"CLAUDE.md": []byte("profile")}
			receipt := []File{{Path: "CLAUDE.md", Size: 7, SHA256: digest(files["CLAUDE.md"]), Mode: "0600"}}
			switch name {
			case "extra":
				files["extra.md"] = []byte("extra")
			case "missing":
				delete(files, "CLAUDE.md")
			case "size":
				receipt[0].Size++
			case "hash":
				files["CLAUDE.md"][0] ^= 1
			case "mode":
				receipt[0].Mode = "0644"
			case "path":
				files["../CLAUDE.md"] = files["CLAUDE.md"]
				delete(files, "CLAUDE.md")
				receipt[0].Path = "../CLAUDE.md"
			case "alias":
				files["claude.md"] = files["CLAUDE.md"]
				changedReceipt := receipt[0]
				changedReceipt.Path = "claude.md"
				receipt = append(receipt, changedReceipt)
			case "bound":
				receipt[0].Size = MaxFileBytes + 1
			}
			required(t, validateClaudeFileSet(files, receipt), "project_context_invalid")
		})
	}
}
