package muse

import (
	"reflect"
	"testing"

	"github.com/relux-works/curator/internal/harnessapi"
)

func TestProbeAdapterCharacterization(t *testing.T) {
	adapter := New()
	if adapter.ID() != ID || adapter.APIVersion() != harnessapi.AdapterAPIVersionV1 {
		t.Fatalf("adapter identity: id=%q version=%d", adapter.ID(), adapter.APIVersion())
	}
	if want := []harnessapi.Capability{harnessapi.ReleaseProbing}; !reflect.DeepEqual(adapter.Capabilities().Values(), want) {
		t.Fatalf("capabilities=%v, want %v", adapter.Capabilities().Values(), want)
	}
	probe, err := adapter.ReleaseProbe(harnessapi.Context{})
	if err != nil {
		t.Fatal(err)
	}
	if probe.ExecutableAlias() != "muse" || !reflect.DeepEqual(probe.Arguments(), []string{"--version"}) || probe.Timeout() != harnessapi.ProbeTimeoutV1 || len(probe.Environment()) != 0 {
		t.Fatalf("probe tuple changed: alias=%q args=%v timeout=%s env=%v", probe.ExecutableAlias(), probe.Arguments(), probe.Timeout(), probe.Environment())
	}
}
