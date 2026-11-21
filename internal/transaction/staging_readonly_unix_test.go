//go:build !windows

package transaction

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// Losing the writable descriptor after a durable partial copy must still let
// recovery discard the owned staging file without changing source/live bytes
// or source mode. Writable mode is the nearby control for the same crash.
func TestReadOnlyStagingCrashRecovery(t *testing.T) {
	for _, mode := range []os.FileMode{0o400, 0o444, 0o600} {
		t.Run(fmt.Sprintf("%04o", mode), func(t *testing.T) {
			root := t.TempDir()
			home := filepath.Join(root, "home")
			live, source := preparationCrashFixture(t, root, "file")
			if err := os.Chmod(source, mode); err != nil {
				t.Fatal(err)
			}
			preimage, err := DigestPath(live)
			if err != nil {
				t.Fatal(err)
			}
			desired, err := DigestPath(source)
			if err != nil {
				t.Fatal(err)
			}
			journal := crashDuringPreparationAt(t, home, live, source, "readonly-stage-crash", PointAfterStagingChunkSync, 2)
			target := journal.Targets[0]
			info, err := os.Stat(target.StagedPath)
			if err != nil || info.Mode().Perm() != mode || info.Size() != 2*stagingCopyChunkSize {
				t.Fatalf("intended partial copy/mode not reached: %v, %v", info, err)
			}
			if err := mustEngine(t, home).Recover(testLock{}); err != nil {
				t.Fatal(err)
			}
			if got, err := DigestPath(live); err != nil || got != preimage {
				t.Fatalf("recovery changed live preimage: %q, %v", got, err)
			}
			if got, err := DigestPath(source); err != nil || got != desired {
				t.Fatalf("recovery changed source bytes or mode: %q, %v", got, err)
			}
			if _, err := os.Lstat(target.StagedPath); !os.IsNotExist(err) {
				t.Fatalf("recovery left partial staging: %v", err)
			}
		})
	}
}
