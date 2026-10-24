package projectlease_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/scopes"
	"github.com/relux-works/curator/internal/transaction"
)

// Subprocess observation helper for internal admission tests: it acquires the
// real same-home lock and runs Collect with the engine's actual referenced build keys.
// Output reports observations; the parent checks the required refusal/effects.
func TestLeaseCollectorProcessHelper(t *testing.T) {
	home := os.Getenv("CURATOR_TEST_PENDING_GC_HOME")
	if home == "" {
		return
	}
	p := struct {
		KeyCount int
		Acquired bool
		Error    string
		Removed  []string
	}{}
	defer func() {
		raw, err := json.Marshal(p)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Printf("collect-result:%s\n", raw)
	}()
	lm, err := managerlock.New(home)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if os.Getenv("CURATOR_TEST_PENDING_GC_MODE") == "try" {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 25*time.Millisecond)
		defer cancel()
	}
	lock, err := lm.AcquireHomeOnly(ctx, false)
	if err != nil {
		p.Error = err.Error()
		return
	}
	defer func() { _ = lock.Close() }()
	p.Acquired = true
	engine, err := transaction.New(home)
	if err != nil {
		t.Fatal(err)
	}
	keys, err := engine.ReferencedBuildKeys(lock)
	if err != nil {
		t.Fatal(err)
	}
	p.KeyCount = len(keys)
	result, err := scopes.Collect(scopes.MaintenanceRequest{Home: home, Lock: lock, JournalKeys: keys})
	p.Removed = result.RemovedRuntime
	if err != nil {
		p.Error = err.Error()
	}
}
