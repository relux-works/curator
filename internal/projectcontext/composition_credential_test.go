package projectcontext

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/envauthority"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envprofile"
	"github.com/relux-works/curator/internal/managerlock"
)

type compositionCredentialFixture struct {
	fixture
	request     ComposeRequest
	preview     CompositionPreview
	composition Composition
	owner       string
}

func newCompositionCredentialFixture(t *testing.T, mode, entry string) compositionCredentialFixture {
	t.Helper()
	f := compositionModeFixture(t, "claude_code", nil, mode)
	approveCompositionDeclaration(t, f, Declaration{Version: 1, Items: []Request{}})
	req := compositionRequest(f)
	req.Selections = nil
	p, err := f.store.PreviewComposition(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	x := compositionCredentialFixture{fixture: f, request: req, preview: p, owner: principal(t, f)}
	if entry == "resolve" || entry == "authorize" {
		x.composition, err = (Authority{Store: f.store, CWD: f.root}).PublishComposition(context.Background(), req, p.SHA256, x.owner)
		if err != nil {
			t.Fatal(err)
		}
	}
	return x
}

func compositionCredentialMarker(t *testing.T, f fixture, node, orphan bool) error {
	t.Helper()
	b, err := os.ReadFile("../envmarker/testdata/seed-c/node.json")
	if err != nil {
		t.Fatal(err)
	}
	var marker map[string]any
	if err = json.Unmarshal(b, &marker); err != nil {
		t.Fatal(err)
	}
	name := "base"
	if orphan {
		name = "orphan"
	}
	profile := marker["profile"].(map[string]any)
	profile["name"], profile["root"] = name, name
	profile["lock_sha256"] = f.registration.Profile.LockSHA256
	marker["members"].([]any)[0].(map[string]any)["name"] = name
	if !node {
		marker["credential_source"] = map[string]any{"mode": "per-home"}
	}
	b, err = json.Marshal(marker)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = envmarker.Parse(b); err != nil {
		t.Fatalf("schema setup: %v", err)
	}
	write(t, filepath.Join(f.store.Home, "environments", name, "claude_code", envmarker.Name), string(b))
	err = envauthority.Validate(f.store.Home)
	if node {
		if err == nil || !strings.HasPrefix(err.Error(), "credential_injection_unavailable:") {
			t.Fatalf("node authority setup: %v", err)
		}
	} else if err != nil {
		t.Fatalf("per-home authority setup: %v", err)
	}
	return err
}

func (f compositionCredentialFixture) invoke(t *testing.T, entry string) (int, error) {
	t.Helper()
	ctx := context.Background()
	a := Authority{Store: f.store, CWD: f.root}
	var held *managerlock.HomeLock
	if entry == "resolve" || entry == "authorize" || entry == "visit" {
		m, err := managerlock.New(f.store.Home)
		if err != nil {
			t.Fatal(err)
		}
		held, err = m.AcquireExistingHomeOnly(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			if err := held.AssertHeld(); err != nil {
				t.Errorf("composition closed borrowed lock: %v", err)
			}
			if err := held.Close(); err != nil {
				t.Error(err)
			}
		}()
	}
	calls := 0
	var err error
	switch entry {
	case "preview":
		_, err = f.store.PreviewComposition(ctx, f.request)
	case "publish":
		_, err = a.PublishComposition(ctx, f.request, f.preview.SHA256, f.owner)
	case "resolve":
		_, err = a.Resolve(ctx, held.Held(), f.composition.Ref, f.composition.SHA256)
	case "authorize":
		err = a.Authorize(ctx, held.Held(), f.composition, f.owner)
	case "visit":
		err = envprofile.VisitCompositionProfile(f.store.Home, "base", f.registration.Profile.LockSHA256, func(*contextlock.Lock, contextlock.Member, string) error { calls++; return nil })
	default:
		t.Fatalf("unknown composition entry: %s", entry)
	}
	return calls, err
}

// Record file bytes, logical modes, directory membership and symlink targets.
// Access times and other volatile filesystem metadata are deliberately excluded.
func compositionCredentialSnapshot(t *testing.T, home string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.WalkDir(home, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		name, err := filepath.Rel(home, path)
		if err != nil {
			return err
		}
		value := info.Mode().String()
		if info.Mode().IsRegular() {
			b, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			value += fmt.Sprintf("/%x", sha256.Sum256(b))
		} else if info.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(path)
			if err != nil {
				return err
			}
			value += "/" + target
		}
		out[name] = value
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// Claim: each of the five real entries rejects selected and orphan node markers
// with the shared gate's exact diagnostic in both modes, without any manager
// change or visitor callback. Paired per-home entries succeed, and every
// borrowed lock remains owned by its caller. This adopts the selected-home
// review witness and extends it to the independently reproduced orphan scope.
func TestCompositionCredentialAuthority(t *testing.T) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("CIP filesystem binding supports macOS/Linux")
	}
	for _, entry := range []string{"preview", "publish", "resolve", "authorize", "visit"} {
		for _, mode := range []string{"strict", "native-project"} {
			for _, orphan := range []bool{false, true} {
				for _, node := range []bool{false, true} {
					t.Run(fmt.Sprintf("%s/%s/orphan=%v/node=%v", entry, mode, orphan, node), func(t *testing.T) {
						f := newCompositionCredentialFixture(t, mode, entry)
						gate := compositionCredentialMarker(t, f.fixture, node, orphan)
						before := compositionCredentialSnapshot(t, f.store.Home)
						calls, err := f.invoke(t, entry)
						if node {
							if err == nil || err.Error() != gate.Error() {
								t.Errorf("composition bypassed manager authority: entry=%s error=%v", entry, err)
							}
							if calls != 0 {
								t.Errorf("node authority invoked visitor before refusal: callbacks=%d", calls)
							}
							if !reflect.DeepEqual(before, compositionCredentialSnapshot(t, f.store.Home)) {
								t.Error("node authority refusal changed manager state")
							}
						} else {
							if err != nil {
								t.Fatalf("per-home positive: %v", err)
							}
							if entry == "visit" && calls == 0 {
								t.Error("per-home visitor did not execute")
							}
							if entry == "publish" {
								if _, err := os.Stat(filepath.Join(f.store.Home, "project-context", "compositions", f.preview.SHA256+".json")); err != nil {
									t.Fatalf("per-home publication missing: %v", err)
								}
							} else if !reflect.DeepEqual(before, compositionCredentialSnapshot(t, f.store.Home)) {
								t.Error("per-home read changed manager state")
							}
						}
					})
				}
			}
		}
	}
}

// Claim: node admission precedes even a failing protected-state/lock read.
// The paired per-home carrier instead reaches the deliberately malformed
// protected record. Thus a later visitor gate cannot hide a missing entry gate.
func TestCompositionCredentialAuthorityBeforeState(t *testing.T) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("CIP filesystem binding supports macOS/Linux")
	}
	for _, entry := range []string{"preview", "publish", "resolve", "authorize", "visit"} {
		for _, node := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/node=%v", entry, node), func(t *testing.T) {
				f := newCompositionCredentialFixture(t, "strict", entry)
				gate := compositionCredentialMarker(t, f.fixture, node, false)
				path := filepath.Join(f.store.Home, "project-context", "state.json")
				if entry == "visit" {
					path = filepath.Join(f.store.Home, "profiles", "base", "lock.json")
				}
				write(t, path, "{not-json")
				before := compositionCredentialSnapshot(t, f.store.Home)
				calls, err := f.invoke(t, entry)
				if node {
					if err == nil || err.Error() != gate.Error() {
						t.Errorf("composition authority admission missing before state: entry=%s error=%v", entry, err)
					}
				} else {
					prefix := "project_context_invalid: malformed protocol JSON object:"
					if entry == "visit" {
						prefix = `environment_store_untrusted: profile "base" lock cannot be trusted:`
					}
					if err == nil || !strings.HasPrefix(err.Error(), prefix) || !strings.Contains(err.Error(), "malformed protocol JSON object:") {
						t.Errorf("per-home malformed-state control: %v", err)
					}
				}
				if calls != 0 || !reflect.DeepEqual(before, compositionCredentialSnapshot(t, f.store.Home)) {
					t.Error("failed composition admission changed state or invoked visitor")
				}
			})
		}
	}
}

// Claim: Preview and Publish refuse node authority before accessing an absent
// lock directory. Per-home controls reach that lock error and succeed once the
// same provisioned directory is restored; neither refusal creates lock state.
func TestCompositionCredentialAuthorityBeforeLock(t *testing.T) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("CIP filesystem binding supports macOS/Linux")
	}
	for _, entry := range []string{"preview", "publish"} {
		for _, node := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/node=%v", entry, node), func(t *testing.T) {
				f := newCompositionCredentialFixture(t, "strict", entry)
				gate := compositionCredentialMarker(t, f.fixture, node, false)
				lockDir := filepath.Join(f.store.Home, "state", "locks", "v1")
				saved := filepath.Join(t.TempDir(), "locks")
				if err := os.Rename(lockDir, saved); err != nil {
					t.Fatal(err)
				}
				before := compositionCredentialSnapshot(t, f.store.Home)
				_, err := f.invoke(t, entry)
				if node {
					if err == nil || err.Error() != gate.Error() {
						t.Fatalf("authority did not precede lock access: %v", err)
					}
				} else if !errors.Is(err, os.ErrNotExist) {
					t.Fatalf("per-home did not reach missing lock: %v", err)
				}
				if !reflect.DeepEqual(before, compositionCredentialSnapshot(t, f.store.Home)) {
					t.Fatal("pre-lock refusal changed state")
				}
				if err := os.Rename(saved, lockDir); err != nil {
					t.Fatal(err)
				}
				if !node {
					if _, err := f.invoke(t, entry); err != nil {
						t.Fatalf("restored per-home control: %v", err)
					}
				}
			})
		}
	}
}

// Claim: the shared operation constructor rechecks real authority after lock
// acquisition and before protected-state reads. A marker inserted after the
// successful pre-lock check yields the exact node refusal and no subsequent
// writes; per-home succeeds. The injection is independent of the held recheck.
// This is a deterministic constructor seam, not a claim of hostile-UID safety.
func TestCompositionCredentialAuthorityUnderLock(t *testing.T) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("CIP filesystem binding supports macOS/Linux")
	}
	for _, node := range []bool{false, true} {
		t.Run(fmt.Sprintf("node=%v", node), func(t *testing.T) {
			f := newCompositionCredentialFixture(t, "strict", "preview")
			injected := false
			var gate error
			var before map[string]string
			op, err := f.store.beginValidated(false, func() error {
				verdict := envauthority.Validate(f.store.Home)
				if !injected {
					if verdict != nil {
						t.Fatalf("pre-lock authority setup: %v", verdict)
					}
					injected = true
					gate = compositionCredentialMarker(t, f.fixture, node, false)
					before = compositionCredentialSnapshot(t, f.store.Home)
				}
				return verdict
			})
			if op != nil {
				op.close()
			}
			if !injected {
				t.Fatal("pre-lock marker injection not applied")
			}
			if node {
				if err == nil || err.Error() != gate.Error() || op != nil {
					t.Fatalf("under-lock authority admission missing: %v", err)
				}
			} else if err != nil || op == nil {
				t.Fatalf("per-home under-lock control: %v", err)
			}
			if !reflect.DeepEqual(before, compositionCredentialSnapshot(t, f.store.Home)) {
				t.Fatal("held admission changed protected state")
			}
		})
	}
}

type compositionCredentialContext struct {
	context.Context
	check func()
}

func (c compositionCredentialContext) Err() error { c.check(); return c.Context.Err() }

// Claim: a node marker appearing at the last context check before first
// publication prevents the immutable record write. The empty namespace may
// already exist at this seam; no later changes are allowed. Per-home publishes.
func TestCompositionCredentialPublicationBoundary(t *testing.T) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("CIP filesystem binding supports macOS/Linux")
	}
	for _, node := range []bool{false, true} {
		t.Run(fmt.Sprintf("node=%v", node), func(t *testing.T) {
			f := newCompositionCredentialFixture(t, "strict", "publish")
			dir := filepath.Join(f.store.Home, "project-context", "compositions")
			injected := false
			var gate error
			var before map[string]string
			ctx := compositionCredentialContext{context.Background(), func() {
				if injected {
					return
				}
				if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
					return
				} else if err != nil {
					t.Fatal(err)
				}
				injected = true
				gate = compositionCredentialMarker(t, f.fixture, node, false)
				before = compositionCredentialSnapshot(t, f.store.Home)
			}}
			_, err := (Authority{Store: f.store, CWD: f.root}).PublishComposition(ctx, f.request, f.preview.SHA256, f.owner)
			if !injected {
				t.Fatal("publication boundary plant not applied")
			}
			record := filepath.Join(dir, f.preview.SHA256+".json")
			if node {
				if err == nil || err.Error() != gate.Error() {
					t.Fatalf("publication authority recheck missing: %v", err)
				}
				if _, err := os.Stat(record); !errors.Is(err, os.ErrNotExist) {
					t.Fatalf("node published record: %v", err)
				}
				if !reflect.DeepEqual(before, compositionCredentialSnapshot(t, f.store.Home)) {
					t.Fatal("publication refusal wrote after marker change")
				}
			} else {
				if err != nil {
					t.Fatalf("per-home publication control: %v", err)
				}
				if _, err := os.Stat(record); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

// Claim: if a visitor installs node authority during its first callback, no
// second protected member is visited and its caller still owns the home lock.
// The per-home neighbor visits the full two-member closure with no later writes.
func TestCompositionCredentialVisitorReentry(t *testing.T) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("CIP filesystem binding supports macOS/Linux")
	}
	for _, node := range []bool{false, true} {
		t.Run(fmt.Sprintf("node=%v", node), func(t *testing.T) {
			f := compositionFixture(t, "claude_code", map[string]string{"tool": compositionSkillSource(t, "tool", "")})
			m, err := managerlock.New(f.store.Home)
			if err != nil {
				t.Fatal(err)
			}
			held, err := m.AcquireExistingHomeOnly(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := held.AssertHeld(); err != nil {
					t.Error(err)
				}
				if err := held.Close(); err != nil {
					t.Error(err)
				}
			}()
			calls := 0
			var gate error
			var before map[string]string
			err = envprofile.VisitCompositionProfile(f.store.Home, "base", f.registration.Profile.LockSHA256, func(*contextlock.Lock, contextlock.Member, string) error {
				calls++
				if calls == 1 {
					gate = compositionCredentialMarker(t, f, node, false)
					before = compositionCredentialSnapshot(t, f.store.Home)
				}
				return nil
			})
			if node {
				if err == nil || err.Error() != gate.Error() || calls != 1 {
					t.Fatalf("visitor authority recheck missing: callbacks=%d error=%v", calls, err)
				}
			} else if err != nil || calls != 2 {
				t.Fatalf("per-home visitor control: callbacks=%d error=%v", calls, err)
			}
			if !reflect.DeepEqual(before, compositionCredentialSnapshot(t, f.store.Home)) {
				t.Fatal("visitor reentry changed protected state")
			}
		})
	}
}
