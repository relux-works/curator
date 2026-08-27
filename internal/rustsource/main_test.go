package rustsource

import (
	"os"
	"testing"
)

func requireNativeCargoDescriptor(t *testing.T) {
	t.Helper()
	target, approved := NativeCargoDescriptorAvailable()
	if target != "" && !approved {
		t.Skipf("no operator-approved Cargo descriptor for native target %s", target)
	}
}

func TestMain(m *testing.M) {
	if handled, code := DispatchInternalWorker(os.Args[1:], os.Stdin, os.Stdout); handled {
		os.Exit(code)
	}
	os.Exit(m.Run())
}
