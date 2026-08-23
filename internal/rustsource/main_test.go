package rustsource

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	if handled, code := DispatchInternalWorker(os.Args[1:], os.Stdin, os.Stdout); handled {
		os.Exit(code)
	}
	os.Exit(m.Run())
}
