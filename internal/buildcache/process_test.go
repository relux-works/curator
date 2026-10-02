package buildcache

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Drive Sweep itself against protected publications in both namespaces; the
// fake table supplies only process images, never marker or journal references.
func TestSweepLiveProcessBuilds(t *testing.T) {
	for _, sourceAware := range []bool{false, true} {
		for _, row := range []struct {
			name, command, image string
			err                  error
			keep, warn           bool
		}{
			{name: "in-use", command: "runner", image: "bin/runner", keep: true},
			{name: "other-file-under-build", command: "runner", image: "other/image", keep: true},
			{name: "unused", command: "runner"},
			{name: "sibling-prefix", command: "runner", image: "../prefix/bin/runner"},
			{name: "enumeration-error", command: "runner", err: errors.New("injected process-table failure"), keep: true, warn: true},
			{name: "daemon", command: "tb-sessiond", image: "bin/tb-sessiond", keep: true},
			{name: "unused-daemon", command: "tb-sessiond"},
		} {
			namespace := "legacy"
			if sourceAware {
				namespace = "receipt-3"
			}
			t.Run(namespace+"/"+row.name, func(t *testing.T) {
				store := newSweepStore(t)
				input := testInput(row.command)
				if sourceAware {
					input = sourceAwareTestInput(row.command)
				}
				publication, _ := testPublication(t, store.Home(), input, []byte("executable"))
				published, err := store.Publish(publication, testHomeLock{})
				if err != nil {
					t.Fatal(err)
				}
				dir, _, err := store.pathsIn(published.CacheKey, sourceAware)
				if err != nil {
					t.Fatal(err)
				}
				backdate(t, dir, 30*24*time.Hour)
				calls := 0
				store.ExecutablePaths = func() ([]string, error) {
					calls++
					if row.image == "" {
						return nil, row.err
					}
					image := filepath.Join(dir, filepath.FromSlash(row.image))
					if row.name == "sibling-prefix" {
						image = dir + "-sibling" + string(filepath.Separator) + "runner"
					}
					return []string{image}, row.err
				}
				result, err := store.Sweep(SweepRequest{}, testHomeLock{})
				if err != nil {
					t.Fatal(err)
				}
				_, statErr := os.Stat(dir)
				if row.keep && statErr != nil {
					t.Fatalf("in-use or uncertain build removed: %v; %+v", statErr, result)
				}
				if !row.keep && (!os.IsNotExist(statErr) || len(result.Removed) != 1 || result.Removed[0] != string(published.CacheKey)) {
					t.Fatalf("unused build not swept: %v; %+v", statErr, result)
				}
				if row.keep && len(result.Removed) != 0 {
					t.Fatalf("retained build reported removed: %+v", result)
				}
				if calls != 1 || (len(result.Warnings) > 0) != row.warn {
					t.Fatalf("table calls=%d warnings=%v", calls, result.Warnings)
				}
				if row.warn && !warningsContaining(result, "injected process-table failure") {
					t.Fatalf("enumeration failure not explained: %+v", result)
				}
			})
		}
	}
}

func TestSweepRejectsPartialProcessTable(t *testing.T) {
	store := newSweepStore(t)
	key := publishTestEntry(t, store, "runner", "executable")
	backdate(t, entryPathOf(store, key), 30*24*time.Hour)
	for _, image := range []string{"", "runner"} {
		store.ExecutablePaths = func() ([]string, error) { return []string{image}, nil }
		result, err := store.Sweep(SweepRequest{}, testHomeLock{})
		if err != nil || len(result.Removed) != 0 || !entryExists(t, store, key) || !warningsContaining(result, "not absolute") {
			t.Fatalf("malformed image accepted: %+v, %v", result, err)
		}
	}
	store.ExecutablePaths = func() ([]string, error) {
		return []string{filepath.Join(t.TempDir(), "unrelated")}, errors.New("partial read")
	}
	result, err := store.Sweep(SweepRequest{}, testHomeLock{})
	if err != nil || len(result.Removed) != 0 || !entryExists(t, store, key) || !warningsContaining(result, "partial read") {
		t.Fatalf("partial process table accepted: %+v, %v", result, err)
	}
}

func TestSweepRetainsExecutableThroughSymlink(t *testing.T) {
	store := newSweepStore(t)
	key := publishTestEntry(t, store, "runner", "executable")
	dir := entryPathOf(store, key)
	backdate(t, dir, 30*24*time.Hour)
	link := filepath.Join(t.TempDir(), "alias")
	if err := os.Symlink(dir, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	files, err := os.ReadDir(filepath.Join(dir, "bin"))
	if err != nil || len(files) != 1 {
		t.Fatalf("bin files: %v, %v", files, err)
	}
	store.ExecutablePaths = func() ([]string, error) { return []string{filepath.Join(link, "bin", files[0].Name())}, nil }
	result, err := store.Sweep(SweepRequest{}, testHomeLock{})
	if err != nil || len(result.Removed) != 0 || !entryExists(t, store, key) {
		t.Fatalf("symlink executable swept: %+v, %v", result, err)
	}
}
