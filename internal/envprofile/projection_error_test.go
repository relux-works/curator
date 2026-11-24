package envprofile

import (
	"errors"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envregistry"
)

var errPiSkillProjection = errors.New("injected Pi user-scope skill projection failure")

func failPiSkillProjection(t *testing.T) {
	t.Helper()
	prior := projectedSkillsAdapterByID
	projectedSkillsAdapterByID = func(id string) (envregistry.Adapter, error) {
		if id == envregistry.Pi {
			return envregistry.Adapter{}, errPiSkillProjection
		}
		return prior(id)
	}
	t.Cleanup(func() { projectedSkillsAdapterByID = prior })
}

func TestNativeSkillsStateReturnsProjectionFailure(t *testing.T) {
	projected, err := envregistry.ByID(envregistry.Pi)
	if err != nil || projected.SkillsDir != "skills" {
		t.Fatalf("Pi positive projection control = %+v, %v; want skills", projected, err)
	}
	fx := writeManagedFixture(t, "pi-projection-state")
	failPiSkillProjection(t)
	row, err := nativeSkillsState(fx.home, fx.profile, envregistry.Pi, fx.request(envregistry.Pi).NativeHomeOf)
	if !errors.Is(err, errPiSkillProjection) {
		t.Fatalf("native skills state error = %v, want projection failure", err)
	}
	if row != nil {
		t.Fatalf("native skills state = %+v after projection failure, want no result", row)
	}
}

func TestUseWithPolicyReturnsProjectionFailureInsteadOfFallbackTarget(t *testing.T) {
	pinHomes(t)
	setActivationRuntimes(t, allActivationRuntimes()...)
	positive := writeManagedFixture(t, "pi-use-positive")
	results, err := UseWithPolicy(positive.home, positive.profile, envregistry.Pi, "", false, Policy{})
	if err != nil {
		t.Fatalf("valid Pi native skill projection could not be used: %v", err)
	}
	foundPi := false
	for _, result := range results {
		if result.Adapter == envregistry.Pi {
			foundPi = result.OK
		}
	}
	if !foundPi {
		t.Fatalf("valid Pi use did not materialize a Pi entry: %+v", results)
	}

	failing := writeManagedFixture(t, "pi-use-error")
	profileBefore := snapshotPlanState(t, failing.home)
	nativeBefore := snapshotActivationTree(t, failing.native[envregistry.Pi])
	currentBefore, err := Current(failing.home)
	if err != nil {
		t.Fatal(err)
	}
	failPiSkillProjection(t)
	results, err = UseWithPolicy(failing.home, failing.profile, envregistry.Pi, "", false, Policy{})
	if err == nil || !strings.Contains(err.Error(), DiagUsePartial) {
		t.Fatalf("Pi profile-use error = %v, results=%+v; want the normal partial-use error", err, results)
	}
	if len(results) != 1 || results[0].Adapter != envregistry.Pi || results[0].OK ||
		!strings.Contains(results[0].Detail, errPiSkillProjection.Error()) {
		t.Fatalf("Pi projection failure result = %+v; want preserved projection detail without fallback success", results)
	}
	if got := snapshotPlanState(t, failing.home); got != profileBefore {
		t.Fatal("profile-use projection failure changed the profile or store publication state")
	}
	if got := snapshotActivationTree(t, failing.native[envregistry.Pi]); !reflect.DeepEqual(got, nativeBefore) {
		t.Fatal("profile-use projection failure wrote through a fallback native target")
	}
	currentAfter, err := Current(failing.home)
	if err != nil {
		t.Fatal(err)
	}
	if currentAfter != currentBefore {
		t.Fatalf("profile-use projection failure changed current profile from %q to %q", currentBefore, currentAfter)
	}
}

func TestCaptureNativeReturnsProjectionFailure(t *testing.T) {
	adapter, ok := adapterByID(envregistry.Pi)
	if !ok {
		t.Fatal("Pi adapter is missing")
	}
	native := t.TempDir()
	positive := &operation{undo: &globalUndo{entries: map[string]*globalImage{}}}
	if err := positive.captureNative(native, adapter, []skillOf{{name: "skill-pi"}}); err != nil {
		t.Fatalf("valid Pi rollback capture failed: %v", err)
	}
	if _, ok := positive.undo.entries[filepath.Join(native, "skills", "skill-pi")]; !ok {
		t.Fatalf("Pi positive capture did not include declared skill path: %v", positive.undo.entries)
	}

	failPiSkillProjection(t)
	failing := &operation{undo: &globalUndo{entries: map[string]*globalImage{}}}
	err := failing.captureNative(native, adapter, []skillOf{{name: "skill-pi"}})
	if !errors.Is(err, errPiSkillProjection) {
		t.Fatalf("Pi rollback capture error = %v, want projection failure", err)
	}
	if len(failing.undo.entries) != 0 {
		t.Fatalf("rollback capture recorded fallback paths after projection failure: %v", failing.undo.entries)
	}
}

func TestStatusOfReturnsProjectionFailureInsteadOfUnknownAdapterRow(t *testing.T) {
	// With no managed profiles, StatusOf reaches adapterStates directly and
	// cannot fail earlier while constructing native-skill scope rows.
	req := StatusRequest{Home: t.TempDir(), Machine: envregistry.DefaultMachineConfig()}
	positive, err := StatusOf(req)
	if err != nil {
		t.Fatalf("valid status projection failed: %v", err)
	}
	var piState *AdapterState
	for i := range positive.Adapters {
		if positive.Adapters[i].ID == envregistry.Pi {
			piState = &positive.Adapters[i]
			break
		}
	}
	if piState == nil || len(piState.Supported) == 0 {
		t.Fatalf("valid Pi status projection = %+v, want supported forms", piState)
	}

	failPiSkillProjection(t)
	status, err := StatusOf(req)
	if !errors.Is(err, errPiSkillProjection) {
		t.Fatalf("status error = %v, status=%+v; want projection failure", err, status)
	}
	if status != nil {
		t.Fatalf("status = %+v after projection failure, want no partial status", status)
	}
}
