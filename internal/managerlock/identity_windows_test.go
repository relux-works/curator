//go:build windows

package managerlock

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unsafe"

	"golang.org/x/sys/windows"
)

func TestWindowsMissingHomeCaseAliasesShareIdentityAndContention(t *testing.T) {
	base := t.TempDir()
	caseSensitive, err := directoryCaseSensitive(base)
	if err != nil {
		t.Fatal(err)
	}
	if caseSensitive {
		t.Skip("temporary directory uses case-sensitive Windows path semantics")
	}

	upperHome := filepath.Join(base, "Missing", "CuratorHome")
	lowerHome := filepath.Join(base, "mISSING", "cURATORhOME")
	first, err := New(upperHome)
	if err != nil {
		t.Fatal(err)
	}
	second, err := New(lowerHome)
	if err != nil {
		t.Fatal(err)
	}
	if first.Home() != second.Home() {
		t.Fatalf("case-alias identities differ before creation: %q != %q", first.Home(), second.Home())
	}
	if first.homeLockPath() != second.homeLockPath() {
		t.Fatalf("case-alias home lock paths differ: %q != %q", first.homeLockPath(), second.homeLockPath())
	}
	if first.process != second.process {
		t.Fatal("case-alias managers split process-local ordering state")
	}

	project := t.TempDir()
	operation := first.NewOperation(false)
	cleanupOperation(t, operation)
	if _, err := operation.AcquireProjects(context.Background(), project); err != nil {
		t.Fatal(err)
	}
	if err := operation.AcquireBuildKey(context.Background(), "case-alias-order-state"); err != nil {
		t.Fatal(err)
	}
	if _, err := second.AcquireHomeOnly(context.Background(), false); !errors.Is(err, ErrLockOrder) {
		t.Fatalf("case-alias home nested with build key: %v, want ErrLockOrder", err)
	}
	if err := operation.ReleaseBuildKey(); err != nil {
		t.Fatal(err)
	}

	homeLock, err := first.AcquireHomeOnly(context.Background(), false)
	if err != nil {
		t.Fatal(err)
	}
	if got := runHelper(t, "try-home", lowerHome, "", "blocked"); got != "blocked" {
		t.Fatalf("case-alias subprocess = %q, want blocked", got)
	}
	if err := homeLock.Close(); err != nil {
		t.Fatal(err)
	}
	if got := runHelper(t, "try-home", lowerHome, "", "acquired"); got != "acquired" {
		t.Fatalf("released case-alias subprocess = %q, want acquired", got)
	}

	after, err := New(lowerHome)
	if err != nil {
		t.Fatal(err)
	}
	if after.Home() != first.Home() {
		t.Fatalf("case-alias identity changed after creation: before %q, after %q", first.Home(), after.Home())
	}
}

func TestWindowsMixedCaseSensitivityKeepsDistinctHomes(t *testing.T) {
	parent := t.TempDir()
	requireDirectoryCaseSensitivity(t, parent, true)

	upperHome := filepath.Join(parent, "Foo")
	lowerHome := filepath.Join(parent, "foo")
	for _, home := range []string{upperHome, lowerHome} {
		if err := os.Mkdir(home, 0o700); err != nil {
			t.Fatalf("create case-distinct home %q: %v", home, err)
		}
		requireDirectoryCaseSensitivity(t, home, false)
	}

	upper, err := New(upperHome)
	if err != nil {
		t.Fatal(err)
	}
	lower, err := New(lowerHome)
	if err != nil {
		t.Fatal(err)
	}
	if upper.Home() == lower.Home() {
		t.Fatalf("case-distinct physical homes share identity %q", upper.Home())
	}
	if upper.lockRoot == lower.lockRoot {
		t.Fatalf("case-distinct physical homes share lock root %q", upper.lockRoot)
	}
	if upper.process == lower.process {
		t.Fatal("case-distinct physical homes share process-local ordering state")
	}

	upperLock, err := upper.AcquireHomeOnly(context.Background(), false)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := upperLock.Close(); err != nil {
			t.Errorf("release upper home lock: %v", err)
		}
	}()
	lowerLock, err := lower.AcquireHomeOnly(context.Background(), false)
	if err != nil {
		t.Fatalf("distinct lower home did not lock independently: %v", err)
	}
	if err := lowerLock.Close(); err != nil {
		t.Fatal(err)
	}

	uppercaseThird := filepath.Join(parent, strings.ToUpper(filepath.Base(upperHome)))
	if _, err := os.Stat(uppercaseThird); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("lock state was redirected to uppercase sibling %q: %v", uppercaseThird, err)
	}
}

func TestWindowsCaseSensitivePrefixMultiComponentFirstUse(t *testing.T) {
	parent := t.TempDir()
	requireDirectoryCaseSensitivity(t, parent, true)

	home := filepath.Join(parent, "Missing", "Nested", "CuratorHome")
	preCreationVariant := filepath.Join(parent, "Missing", "nESTED", "cURATORhOME")
	first, err := New(home)
	if err != nil {
		t.Fatal(err)
	}
	second, err := New(preCreationVariant)
	if err != nil {
		t.Fatal(err)
	}

	homeLock, err := first.AcquireHomeOnly(context.Background(), false)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := homeLock.Close(); err != nil {
			t.Errorf("release first-use home lock: %v", err)
		}
	}()
	created := []string{
		filepath.Join(parent, "Missing"),
		filepath.Join(parent, "Missing", "Nested"),
		home,
	}
	caseSemantics := make([]bool, len(created))
	for index, path := range created {
		caseSensitive, err := directoryCaseSensitive(path)
		if err != nil {
			t.Fatalf("inspect manager-created directory %q: %v", path, err)
		}
		caseSemantics[index] = caseSensitive
	}

	// Change spelling only where the containing directory actually performs
	// case-insensitive lookup. The first component is under the sensitive test
	// parent and therefore keeps its exact spelling.
	components := []string{"Missing", "Nested", "CuratorHome"}
	parentSensitive := true
	for index := range components {
		if !parentSensitive {
			components[index] = swapCase(components[index])
		}
		parentSensitive = caseSemantics[index]
	}
	alias := filepath.Join(append([]string{parent}, components...)...)
	after, err := New(alias)
	if err != nil {
		t.Fatal(err)
	}
	if after.Home() != first.Home() {
		t.Fatalf("multi-component identity changed after first use: before %q, after %q", first.Home(), after.Home())
	}
	if got := runHelper(t, "try-home", alias, "", "blocked"); got != "blocked" {
		t.Fatalf("multi-component alias subprocess = %q, want blocked", got)
	}
	if err := second.prepare(); err != nil {
		t.Fatal(err)
	}
	if caseSemantics[0] {
		if second.Home() == first.Home() || second.process == first.process {
			t.Fatal("pre-creation case variant merged despite a sensitive containing directory")
		}
	} else if second.Home() != first.Home() || second.process != first.process {
		t.Fatal("pre-creation case alias did not join stabilized physical identity")
	}

	distinctHome := filepath.Join(parent, "missing", "Nested", "CuratorHome")
	distinct, err := New(distinctHome)
	if err != nil {
		t.Fatal(err)
	}
	if distinct.Home() == first.Home() {
		t.Fatalf("case-distinct first component shares identity %q", first.Home())
	}
	distinctLock, err := distinct.AcquireHomeOnly(context.Background(), false)
	if err != nil {
		t.Fatalf("case-distinct multi-component home did not lock independently: %v", err)
	}
	if err := distinctLock.Close(); err != nil {
		t.Fatal(err)
	}
}

func swapCase(value string) string {
	var result strings.Builder
	for _, character := range value {
		switch {
		case character >= 'a' && character <= 'z':
			result.WriteRune(character - ('a' - 'A'))
		case character >= 'A' && character <= 'Z':
			result.WriteRune(character + ('a' - 'A'))
		default:
			result.WriteRune(character)
		}
	}
	return result.String()
}

func requireDirectoryCaseSensitivity(t *testing.T, path string, enabled bool) {
	t.Helper()
	current, err := directoryCaseSensitive(path)
	if err != nil {
		t.Skipf("per-directory case sensitivity unavailable: %v", err)
	}
	if current == enabled {
		return
	}

	pathUTF16, err := windows.UTF16PtrFromString(path)
	if err != nil {
		t.Fatal(err)
	}
	handle, err := windows.CreateFile(
		pathUTF16,
		windows.FILE_READ_ATTRIBUTES|windows.FILE_WRITE_ATTRIBUTES,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_FLAG_BACKUP_SEMANTICS,
		0,
	)
	if err != nil {
		t.Skipf("open directory to change case sensitivity: %v", err)
	}
	defer windows.CloseHandle(handle)

	var flags uint32
	if enabled {
		flags = windows.FILE_CS_FLAG_CASE_SENSITIVE_DIR
	}
	if err := windows.SetFileInformationByHandle(
		handle,
		windows.FileCaseSensitiveInfo,
		(*byte)(unsafe.Pointer(&flags)),
		uint32(unsafe.Sizeof(flags)),
	); err != nil {
		t.Skipf("set directory case sensitivity to %t: %v", enabled, err)
	}
	current, err = directoryCaseSensitive(path)
	if err != nil {
		t.Fatal(err)
	}
	if current != enabled {
		t.Fatalf("directory case sensitivity = %t, want %t", current, enabled)
	}
}
