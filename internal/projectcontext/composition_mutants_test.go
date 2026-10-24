package projectcontext

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// Claim: these bounded source plants violate composition admission or identity,
// and the corresponding regression witness fails
// for its intended assertion. A setup/compile failure is not a caught mutant.
// Opt in explicitly; the ordinary suite keeps only production behavior tests.
func TestCompositionNarrowingMutants(t *testing.T) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("CIP filesystem binding supports macOS/Linux")
	}
	if os.Getenv("CURATOR_PROJECT_COMPOSITION_MUTANTS") != "1" {
		t.Skip("explicit bounded mutation lane")
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct{ name, file, test, old, new, witness string }{
		{"ignore-durable-root-lapses", "composition_held.go", "TestCompositionRetainedRootLapse", `op.loadLapses(root)`, `error(nil)`, "want project_registration_mismatch"},
		{"ignore-missing-cwd-root-loss", "registration.go", "TestCompositionMissingCWDObservesRootLoss", `if err != nil {
		return Registration{}, op.observeUnavailableCWD(cwd)
	}`, `if err != nil {
		return Registration{}, fail("project_registration_mismatch", "CWD unavailable")
	}`, "restored nested registration revived"},
		{"retain-null-waiver-map", "composition.go", "TestCompositionPolicyMetadata", `if policy.System.Waivers == nil {`, `if policy.System.Waivers == nil && false {`, "default policy publication cannot resolve"},
		{"project-only-owners", "composition.go", "TestCompositionCommandCollision", "s := skills[name]\n\t\tfor _, command", "s := skills[name]\n if s.Layer != \"project\" { continue }\n\t\tfor _, command", "want command collision"},
		{"project-first-root", "composition_types.go", "TestCompositionRootBytes", `return bytes.Join(parts, []byte{'\n'})`, `if len(parts)>1 { parts[0],parts[len(parts)-1]=parts[len(parts)-1],parts[0] }; return bytes.Join(parts, []byte{'\n'})`, "root bytes differ"},
		{"no-lf-separator", "composition_types.go", "TestCompositionRootBytes", `return bytes.Join(parts, []byte{'\n'})`, `return bytes.Join(parts, nil)`, "root bytes differ"},
		{"drop-provider-recheck", "composition.go", "TestCompositionRequiredProvider", "for _, r := range reqs {", "for _, r := range reqs[:0] {", "want required provider command missing"},
		{"allow-mcp-replacement", "composition.go", "TestCompositionMCPConflict", "ok && identity(prior) != identity(next)", "ok && identity(prior) != identity(next) && false", "want same-ID MCP replacement"},
		{"ignore-receipt-drift", "composition_authority.go", "TestCompositionAuthority", "identity(fresh) != identity(saved)", "identity(fresh) != identity(saved) && false", "want project_context_drift"},
		{"ignore-publication-principal", "composition_authority.go", "TestCompositionAuthority", "if ownerRef != op.state.Principal {", "if ownerRef != op.state.Principal && false {", "want project_item_unapproved"},
		{"ignore-authorization-principal", "composition_authority.go", "TestCompositionAuthority", "if !tagged(ownerRef, \"principal\") || ownerRef != op.state.Principal {", "if (!tagged(ownerRef, \"principal\") || ownerRef != op.state.Principal) && false {", "want project_item_unapproved"},
		{"ignore-cwd-binding", "composition_authority.go", "TestCompositionAuthority", "r, err := op.registered(id)", "if op != nil { return nil }; r, err := op.registered(id)", "want project_registration_mismatch"},
		{"ignore-held-witness", "../managerlock/managerlock.go", "TestCompositionHeldAuthority", "func (h Held) Validate(home string) error {", "func (h Held) Validate(home string) error { if home != \"\" { return nil }", "zero witness accepted"},
		{"ignore-preview-digest", "composition_authority.go", "TestCompositionAuthority", "!hexRE.MatchString(expected) || plan.Preview.SHA256 != expected", "(!hexRE.MatchString(expected) || plan.Preview.SHA256 != expected) && false", "want project_context_drift"},
		{"allow-adapter-range", "composition_types.go", "TestCompositionAdapterRelease", "invalidProviderVersion(a.ProviderVersion)", "false", "want invalid adapter release contract"},
		{"allow-memory-on", "composition.go", "TestCompositionClosedPreparation", `req.Policy.Version != 1 || req.Policy.Memory != "off"`, `req.Policy.Version != 1`, "want project_context_invalid"},
		{"native-includes-project-context", "composition.go", "TestCompositionNativeContextRefusal", `if req.Mode == "native-project" {`, `if req.Mode == "native-project" && false {`, "want native-project cannot select context items"},
		{"native-requires-exclusion", "composition_types.go", "TestCompositionAdapterModes", `if mode == "strict" {`, `if validSourceMode(mode) {`, "native mode incorrectly requires exclusion"},
		{"mode-omitted-from-digest", "composition.go", "TestCompositionNativeProfileOnly", `out.SHA256 = digest(b)`, `digestRecord := record; digestRecord.Mode = "strict"; out.SHA256 = identity(digestRecord)`, "mode omitted from digest"},
		{"ignore-mode-binding", "composition.go", "TestCompositionNativeContextRefusal", `req.Mode != r.Record.Mode`, `false`, "want project_context_invalid"},
		{"ignore-mode-advertisement", "composition_types.go", "TestCompositionAdapterModes", `if !advertised {`, `if !advertised && false {`, "want project_context_unsupported"},
		{"allow-unknown-mode", "composition_types.go", "TestCompositionAdapterModes", `if !validSourceMode(mode) {`, `if !validSourceMode(mode) && false {`, "want project_context_invalid"},
		{"allow-duplicate-modes", "composition_types.go", "TestCompositionAdapterModes", `(i > 0 && m == a.Modes[i-1])`, `(i > 0 && m == a.Modes[i-1] && false)`, "want project_context_invalid"},
		{"allow-unknown-registration-mode", "registration.go", "TestCompositionRegistrationMode", `if !validSourceMode(mode) {`, `if !validSourceMode(mode) && false {`, "want project_context_invalid"},
		{"ignore-retained-registration-mode", "store.go", "TestCompositionRegistrationMode", `!validSourceMode(v.Mode)`, `false`, "want project_context_drift"},
		{"retain-implicit-mode-alias", "composition_authority.go", "TestCompositionNativeProfileOnly", `req.Mode = again.Preview.Record.Mode`, `_ = req.Mode`, "equivalent explicit mode changed publication"},
		{"entry-only-command-identity", "composition.go", "TestCompositionCommandArtifact", `artifact := s.ArtifactSHA256`, `artifact := s.ArtifactSHA256; if c := s.spec.Commands[command]; c.Type == "script" { artifact = digest(s.files[c.UnixPath]) }`, "changed audited command artifact has unchanged command-owner artifact hash"},
		{"inspection-source-mode-off", "inspection.go", "TestCompositionInspectionStatus", `Mode:         registration.Mode`, `Mode:         "off"`, "inspection conflates source mode with admission"},
		{"inspection-without-native-notice", "inspection.go", "TestCompositionInspectionStatus", `out.Notice = "project sources not isolated"`, `out.Notice = ""`, "inspection isolation notice differs"},
		{"retain-selection-aliases", "composition_authority.go", "TestCompositionPublicationSelections", `req.Selections = canonicalCompositionSelections(req.Selections, op.state.Reviews[req.CheckoutID].Review)`, `_ = req.Selections`, "same composition identity rejected during repeat publication"},
		{"ignore-immutable-publication-metadata", "composition_authority.go", "TestCompositionPublicationMetadata", `if !bytes.Equal(old, b) {`, `if !bytes.Equal(old, b) && false {`, "want immutable composition metadata changed"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			file := filepath.Join(cwd, tc.file)
			source, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Count(string(source), tc.old) != 1 {
				t.Fatalf("plant anchor count differs: %s", tc.name)
			}
			mutant := strings.Replace(string(source), tc.old, tc.new, 1)
			dir := t.TempDir()
			replacement := filepath.Join(dir, "admission.go")
			if err := os.WriteFile(replacement, []byte(mutant), 0600); err != nil {
				t.Fatal(err)
			}
			overlay := filepath.Join(dir, "overlay.json")
			b, _ := json.Marshal(map[string]any{"Replace": map[string]string{file: replacement}})
			if err := os.WriteFile(overlay, b, 0600); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(compositionGo(), "test", "-overlay", overlay, "-count=1", "-run", "^"+tc.test+"$", ".")
			cmd.Dir = cwd
			for _, e := range os.Environ() {
				if !strings.HasPrefix(e, "GIT_DIR=") && !strings.HasPrefix(e, "GIT_WORK_TREE=") {
					cmd.Env = append(cmd.Env, e)
				}
			}
			out, err := cmd.CombinedOutput()
			if err == nil || !strings.Contains(string(out), "--- FAIL: "+tc.test) || !strings.Contains(string(out), tc.witness) || strings.Contains(string(out), "[build failed]") {
				t.Fatalf("mutant did not hit intended witness: %v\n%s", err, out)
			}
			t.Logf("%s: caught intended forbidden-operation witness", tc.name)
		})
	}
}

func compositionGo() string {
	if binary := os.Getenv("CURATOR_GO"); binary != "" {
		return binary
	}
	return "go"
}

// Claim: removing each entry's shared authority calls makes that entry's node
// witness fail before malformed protected state is read. Each plant must compile
// and exit at the named behavioral assertion; a later visitor refusal is not
// evidence that an earlier entry gate ran. Production sources stay untouched.
func TestCompositionCredentialNarrowingControls(t *testing.T) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("CIP filesystem binding supports macOS/Linux")
	}
	if os.Getenv("CURATOR_PROJECT_COMPOSITION_MUTANTS") != "1" {
		t.Skip("explicit bounded mutation lane")
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		entry, file, function, home string
		calls                       int
	}{
		{"preview", "composition.go", "func (s Store) PreviewComposition(", "s.Home", 1},
		{"publish", "composition_authority.go", "func (a Authority) PublishComposition(", "a.Store.Home", 2},
		{"resolve", "composition_authority.go", "func (a Authority) Resolve(", "a.Store.Home", 1},
		{"authorize", "composition_authority.go", "func (a Authority) Authorize(", "a.Store.Home", 1},
		{"visit", "../envprofile/composition.go", "func VisitCompositionProfile(", "home", 2},
	}
	for _, tc := range cases {
		t.Run(tc.entry, func(t *testing.T) {
			file := filepath.Clean(filepath.Join(cwd, tc.file))
			b, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			source := string(b)
			if strings.Count(source, tc.function) != 1 {
				t.Fatal("authority function anchor not unique")
			}
			start := strings.Index(source, tc.function)
			end := len(source)
			if n := strings.Index(source[start:], "\nfunc "); n >= 0 {
				end = start + n
			}
			body := source[start:end]
			call := "envauthority.Validate(" + tc.home + ")"
			if strings.Count(body, call) != tc.calls {
				t.Fatalf("authority call anchor count differs: %s", tc.entry)
			}
			// Retain an import reference while removing every gate call of this entry.
			body = strings.ReplaceAll(body, call, "func() error { _ = envauthority.Validate; return nil }()")
			dir := t.TempDir()
			replacement := filepath.Join(dir, "authority.go")
			if err := os.WriteFile(replacement, []byte(source[:start]+body+source[end:]), 0600); err != nil {
				t.Fatal(err)
			}
			overlay := filepath.Join(dir, "overlay.json")
			mapping, err := json.Marshal(map[string]any{"Replace": map[string]string{file: replacement}})
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(overlay, mapping, 0600); err != nil {
				t.Fatal(err)
			}
			witness := "TestCompositionCredentialAuthorityBeforeState/" + tc.entry + "/node=true"
			cmd := exec.Command(compositionGo(), "test", "-p=2", "-overlay", overlay, "-count=1", "-timeout=2m", "-v", "-run", "^TestCompositionCredentialAuthorityBeforeState$/^"+tc.entry+"$/^node=true$", ".")
			cmd.Dir = cwd
			for _, e := range os.Environ() {
				if !strings.HasPrefix(e, "GIT_DIR=") && !strings.HasPrefix(e, "GIT_WORK_TREE=") {
					cmd.Env = append(cmd.Env, e)
				}
			}
			out, err := cmd.CombinedOutput()
			var exit *exec.ExitError
			if !errors.As(err, &exit) || exit.ExitCode() != 1 || !strings.Contains(string(out), "--- FAIL: "+witness) || !strings.Contains(string(out), "composition authority admission missing before state") || strings.Contains(string(out), "[build failed]") {
				t.Fatalf("authority plant did not hit intended witness: %v\n%s", err, out)
			}
			t.Logf("%s: caught intended authority-before-state witness (exit 1)", tc.entry)
		})
	}
}

// Claim: the pre-lock gate alone cannot satisfy held admission, first
// publication or callback reentry. Each omitted recheck must fail its dedicated
// boundary witness rather than compiling badly or missing the injection seam.
func TestCompositionCredentialBoundaryNarrowingControls(t *testing.T) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("CIP filesystem binding supports macOS/Linux")
	}
	if os.Getenv("CURATOR_PROJECT_COMPOSITION_MUTANTS") != "1" {
		t.Skip("explicit bounded mutation lane")
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct{ name, file, test, old, replacement, witness string }{
		{"under-lock", "store.go", "TestCompositionCredentialAuthorityUnderLock", `}()
	if validate != nil {`, `}()
	if validate != nil && false {`, "expected pre-lock and held checks"},
		{"publication", "composition_authority.go", "TestCompositionCredentialPublicationBoundary", `if err = envauthority.Validate(a.Store.Home); err != nil {`, `if err = error(nil); err != nil {`, "publication authority recheck missing"},
		{"visitor-reentry", "../envprofile/composition.go", "TestCompositionCredentialVisitorReentry", `if err := envauthority.Validate(home); err != nil {
			return err
		}`, `if err := error(nil); err != nil {
			return err
		}`, "visitor authority recheck missing"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			file := filepath.Clean(filepath.Join(cwd, tc.file))
			b, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Count(string(b), tc.old) != 1 {
				t.Fatal("boundary plant anchor not unique")
			}
			dir := t.TempDir()
			replacement := filepath.Join(dir, "boundary.go")
			if err := os.WriteFile(replacement, []byte(strings.Replace(string(b), tc.old, tc.replacement, 1)), 0600); err != nil {
				t.Fatal(err)
			}
			overlay := filepath.Join(dir, "overlay.json")
			mapping, err := json.Marshal(map[string]any{"Replace": map[string]string{file: replacement}})
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(overlay, mapping, 0600); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(compositionGo(), "test", "-p=2", "-overlay", overlay, "-count=1", "-timeout=2m", "-v", "-run", "^"+tc.test+"$/^node=true$", ".")
			cmd.Dir = cwd
			for _, e := range os.Environ() {
				if !strings.HasPrefix(e, "GIT_DIR=") && !strings.HasPrefix(e, "GIT_WORK_TREE=") {
					cmd.Env = append(cmd.Env, e)
				}
			}
			out, err := cmd.CombinedOutput()
			var exit *exec.ExitError
			if !errors.As(err, &exit) || exit.ExitCode() != 1 || !strings.Contains(string(out), "--- FAIL: "+tc.test+"/node=true") || !strings.Contains(string(out), tc.witness) || strings.Contains(string(out), "[build failed]") {
				t.Fatalf("boundary plant did not hit intended witness: %v\n%s", err, out)
			}
			t.Logf("%s: caught intended boundary witness (exit 1)", tc.name)
		})
	}
}
