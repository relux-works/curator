# Bounded layout diagnostic

Temporary diagnostic source (removed from product test package after execution):

```go
package envprofile

import (
	"path/filepath"
	"testing"

	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/identifiers"
)

// This diagnostic proves that the current native Muse resolver and managed SkillsDir cannot describe independent native XDG roots with one home-relative marker path.
func TestNATSkillMuseNativeLayoutConstraint(t *testing.T) {
	base := t.TempDir()
	pinOperatorHome(t, filepath.Join(base, "operator"))
	config := filepath.Join(base, "config-root")
	data := filepath.Join(base, "independent-data-root")
	t.Setenv("XDG_CONFIG_HOME", config)
	t.Setenv("XDG_DATA_HOME", data)
	adapter, err := envregistry.ByID(envregistry.Muse)
	if err != nil {
		t.Fatal(err)
	}
	native, err := (&ResolveRequest{}).nativeHome(envregistry.Muse)
	if err != nil || native != filepath.Join(config, "muse") {
		t.Fatalf("native resolver no longer has the diagnosed config-root contract: %v", err)
	}
	literal := filepath.Join(native, filepath.FromSlash(adapter.SkillsDir), "sample")
	dataSurface := filepath.Join(data, "muse", "skills", "sample")
	if literal == dataSurface {
		t.Fatal("literal SkillsDir unexpectedly reaches the independent data root")
	}
	rel, err := filepath.Rel(native, dataSurface)
	if err != nil {
		t.Fatal(err)
	}
	if identifiers.PortablePath(filepath.ToSlash(rel)) {
		t.Fatal("marker path validator unexpectedly admits the route outside its home")
	}
	if _, ok := adapterByID(envregistry.Muse); ok {
		t.Fatal("Muse unexpectedly participates in the in-place switch registry; reassess the constraint")
	}
	t.Log("native=config-root/muse; literal=config-root/muse/data/muse/skills/sample; independent data surface cannot be expressed as a portable relative path below native")
}

```

Standalone command exit: 0. Raw output:

```text
=== RUN   TestMuseDirectoriesAndIsolation
--- PASS: TestMuseDirectoriesAndIsolation (0.64s)
=== RUN   TestNATSkillMuseNativeLayoutConstraint
    natskill_layout_probe_test.go:42: native=config-root/muse; literal=config-root/muse/data/muse/skills/sample; independent data surface cannot be expressed as a portable relative path below native
--- PASS: TestNATSkillMuseNativeLayoutConstraint (0.00s)
PASS
ok  	github.com/relux-works/curator/internal/envprofile	0.992s

```
