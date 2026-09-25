// System-module admission vectors (environments §3, §5.5, §5.7): the five
// system-module-* materialization cases of vectors/environments.json run
// here, selected by name, byte for byte through the production SystemPrompt
// assembler and the ClassifySystemModules split.
//
// The general consumer in internal/interop/environments iterates whatever
// cases the root publishes, so a root predating the admission family runs
// it green with no admission case at all. This driver is the explicit
// subset accounting: it runs all five, and fails loud on a partial
// publication — or an empty one, since the committed pin serves the
// family and a root-content skip would hide a regression. It lives in
// the production package rather
// than internal/interop/environments because that package's committed
// contract forbids any skip but the deferred root
// (TestNoCaseHereSkipsForAnythingButTheDeferredRoot,
// TestNoLedgerRowForThisPackageToleratesASkip): a root-content driver there
// would weaken the gate it exists to enforce.
package contextmaterialize

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/conformancecoverage"
	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextpkg"
)

// systemModuleAdmissionCases are the five §3 admission cases by name:
// direct, transitive-drop (byte-identical output), transitive-error,
// waived, and overlay-direct.
var systemModuleAdmissionCases = []string{
	"system-module-direct",
	"system-module-transitive-drop",
	"system-module-transitive-error",
	"system-module-transitive-waived",
	"system-module-overlay-direct",
}

// admissionVectorCase is one materialization case of
// vectors/environments.json as this driver consumes it.
type admissionVectorCase struct {
	Name        string          `json:"name"`
	Surface     string          `json:"surface"`
	Environment string          `json:"environment"`
	Lock        json.RawMessage `json:"lock"`
	LockSHA256  string          `json:"lock_sha256"`
	Precedence  struct {
		Winner    string `json:"winner"`
		Placement string `json:"placement"`
	} `json:"precedence"`
	Packages map[string]struct {
		HasContext bool `json:"has_context"`
		Modules    []struct {
			Path         string   `json:"path"`
			Class        string   `json:"class"`
			Environments []string `json:"environments"`
			Content      string   `json:"content"`
		} `json:"modules"`
	} `json:"packages"`
	EmittedOrder []string `json:"emitted_order"`
	FileWritten  bool     `json:"file_written"`
	Files        []struct {
		Path     string `json:"path"`
		Bytes    int    `json:"bytes"`
		Expected string `json:"expected"`
		SHA256   string `json:"sha256"`
	} `json:"files"`
	SurfaceSHA256 string `json:"surface_sha256"`
	MachinePolicy *struct {
		TransitiveSystemModules string `json:"transitive_system_modules"`
		SystemModuleWaivers     []struct {
			Package string `json:"package"`
			Reason  string `json:"reason"`
		} `json:"system_module_waivers"`
	} `json:"machine_policy"`
	Admitted []struct {
		Package string `json:"package"`
		Path    string `json:"path"`
	} `json:"admitted"`
	Dropped []struct {
		Package string `json:"package"`
		Path    string `json:"path"`
	} `json:"dropped"`
	Warnings []struct {
		Diagnostic string `json:"diagnostic"`
		Package    string `json:"package"`
		Path       string `json:"path"`
	} `json:"warnings"`
	Error        string `json:"error"`
	ErrorPackage string `json:"error_package"`
	ErrorModule  string `json:"error_module"`
}

// TestSystemModuleAdmissionVectors drives the production system-prompt
// assembler against the five admission cases, byte for byte.
func TestSystemModuleAdmissionVectors(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	payload, err := os.ReadFile(filepath.Join(root, "vectors", "environments.json")) // #nosec G304 -- explicit conformance input
	if err != nil {
		t.Fatal(err)
	}
	var vector struct {
		MaterializationCases []admissionVectorCase `json:"materialization_cases"`
	}
	if err := json.Unmarshal(payload, &vector); err != nil {
		t.Fatal(err)
	}
	byName := map[string]admissionVectorCase{}
	selectedNames := map[string]bool{}
	for _, tc := range vector.MaterializationCases {
		byName[tc.Name] = tc
	}
	var missing []string
	for _, name := range systemModuleAdmissionCases {
		selectedNames[name] = true
		if _, ok := byName[name]; !ok {
			missing = append(missing, name)
		}
	}
	if len(missing) != 0 {
		t.Fatalf("conformance root %s publishes only %d of %d system-module admission cases, missing %s",
			root, len(systemModuleAdmissionCases)-len(missing), len(systemModuleAdmissionCases), strings.Join(missing, ", "))
	}
	conformancecoverage.RunOutcomes(t, "environments/materialization-cases", vector.MaterializationCases,
		func(tc admissionVectorCase) string { return tc.Name }, func(t *testing.T, tc admissionVectorCase) conformancecoverage.Observation {
			if !selectedNames[tc.Name] {
				return conformancecoverage.Observation{BoundReason: "non-system-module materialization surfaces are driven by the full environments vector consumer"}
			}
			runAdmissionVectorCase(t, root, tc)
			return conformancecoverage.Observation{}
		})
}

// runAdmissionVectorCase executes one admission case through SystemPrompt
// and ClassifySystemModules: lock hash, emitted order, the admitted and
// dropped sets, the drop warnings (or the error-policy refusal naming
// package and module), and the expected file bytes and surface hash.
func runAdmissionVectorCase(t *testing.T, root string, tc admissionVectorCase) {
	t.Helper()
	if tc.Surface != "system-prompt" {
		t.Fatalf("admission case %q has surface %q, want system-prompt", tc.Name, tc.Surface)
	}
	if tc.MachinePolicy == nil {
		t.Fatalf("admission case %q carries no machine_policy", tc.Name)
	}
	lock, err := contextlock.Parse(tc.Lock)
	if err != nil {
		t.Fatalf("vector lock does not parse: %v", err)
	}
	hash, err := lock.Hash()
	if err != nil {
		t.Fatal(err)
	}
	if hash != tc.LockSHA256 {
		t.Fatalf("lock hash %s, want %s", hash, tc.LockSHA256)
	}
	precedence := Precedence{Winner: tc.Precedence.Winner, Placement: tc.Precedence.Placement}
	packages := map[string]Package{}
	for name, pkg := range tc.Packages {
		content := Package{HasContext: pkg.HasContext}
		for _, module := range pkg.Modules {
			content.Modules = append(content.Modules, Module{
				Module: contextpkg.Module{Path: module.Path, Class: module.Class, Environments: module.Environments},
				Bytes:  []byte(module.Content),
			})
		}
		packages[name] = content
	}
	order, err := EmittedOrder(lock, precedence)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, member := range order {
		names = append(names, member.Name)
	}
	if !reflect.DeepEqual(names, tc.EmittedOrder) {
		t.Fatalf("emitted order %v, want %v", names, tc.EmittedOrder)
	}
	admission := Admission{Transitive: tc.MachinePolicy.TransitiveSystemModules}
	for _, waiver := range tc.MachinePolicy.SystemModuleWaivers {
		if admission.Waivers == nil {
			admission.Waivers = map[string]bool{}
		}
		admission.Waivers[waiver.Package] = true
	}
	document, written, dropped, err := SystemPrompt(lock, precedence, tc.Environment, packages, admission)
	if tc.Error != "" {
		if err == nil {
			t.Fatalf("expected %s, got no error", tc.Error)
		}
		if !strings.Contains(err.Error(), tc.Error) ||
			!strings.Contains(err.Error(), tc.ErrorPackage) ||
			!strings.Contains(err.Error(), tc.ErrorModule) {
			t.Fatalf("error %q names neither %s nor %s/%s", err, tc.Error, tc.ErrorPackage, tc.ErrorModule)
		}
		assertAdmissionDropped(t, dropped, tc)
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	if written != tc.FileWritten {
		t.Fatalf("file_written %v, want %v", written, tc.FileWritten)
	}
	admitted, classified := ClassifySystemModules(lock, order, packages, tc.Environment, admission)
	var gotAdmitted [][2]string
	for _, module := range admitted {
		gotAdmitted = append(gotAdmitted, [2]string{module.Package, module.Module.Path})
	}
	var wantAdmitted [][2]string
	for _, module := range tc.Admitted {
		wantAdmitted = append(wantAdmitted, [2]string{module.Package, module.Path})
	}
	if len(gotAdmitted) == 0 {
		gotAdmitted = [][2]string{}
	}
	if len(wantAdmitted) == 0 {
		wantAdmitted = [][2]string{}
	}
	if !reflect.DeepEqual(gotAdmitted, wantAdmitted) {
		t.Fatalf("admitted %v, want %v", gotAdmitted, wantAdmitted)
	}
	assertAdmissionDropped(t, classified, tc)
	assertAdmissionDropped(t, dropped, tc)
	var gotWarnings [][3]string
	for _, module := range dropped {
		gotWarnings = append(gotWarnings, [3]string{DiagSystemModuleDropped, module.Package, module.Path})
	}
	var wantWarnings [][3]string
	for _, warning := range tc.Warnings {
		wantWarnings = append(wantWarnings, [3]string{warning.Diagnostic, warning.Package, warning.Path})
	}
	if len(gotWarnings) == 0 {
		gotWarnings = [][3]string{}
	}
	if len(wantWarnings) == 0 {
		wantWarnings = [][3]string{}
	}
	if !reflect.DeepEqual(gotWarnings, wantWarnings) {
		t.Fatalf("warnings %v, want %v", gotWarnings, wantWarnings)
	}
	if !written {
		if len(tc.Files) != 0 {
			t.Fatalf("the vector lists files for an unwritten surface")
		}
		return
	}
	if len(tc.Files) != 1 {
		t.Fatalf("expected exactly one file for a system-prompt surface, vector lists %d", len(tc.Files))
	}
	want, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(tc.Files[0].Expected))) // #nosec G304 -- explicit conformance input
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(document, want) {
		t.Fatalf("%s bytes differ from %s:\n got %q\nwant %q", tc.Files[0].Path, tc.Files[0].Expected, document, want)
	}
	if len(document) != tc.Files[0].Bytes || FileHash(document) != tc.Files[0].SHA256 {
		t.Fatalf("%s: %d bytes %s, want %d bytes %s",
			tc.Files[0].Path, len(document), FileHash(document), tc.Files[0].Bytes, tc.Files[0].SHA256)
	}
	if got := SurfaceHash(map[string][]byte{tc.Files[0].Path: document}); got != tc.SurfaceSHA256 {
		t.Fatalf("surface hash %s, want %s", got, tc.SurfaceSHA256)
	}
}

// assertAdmissionDropped pins the dropped set in emitted order.
func assertAdmissionDropped(t *testing.T, dropped []DroppedModule, tc admissionVectorCase) {
	t.Helper()
	var got [][2]string
	for _, module := range dropped {
		got = append(got, [2]string{module.Package, module.Path})
	}
	var want [][2]string
	for _, module := range tc.Dropped {
		want = append(want, [2]string{module.Package, module.Path})
	}
	if len(got) == 0 {
		got = [][2]string{}
	}
	if len(want) == 0 {
		want = [][2]string{}
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%s: dropped %v, want %v", tc.Name, got, want)
	}
}
