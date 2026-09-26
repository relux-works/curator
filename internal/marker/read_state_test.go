package marker

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/stateread"
)

func TestReadStateDistinguishesAbsentAndUnreadable(t *testing.T) {
	dir := t.TempDir()
	if got, kind, err := ReadState(dir); err != nil || got != nil || kind != stateread.KindAbsent {
		t.Fatalf("absent ReadState = (%v, %q, %v), want nil, absent, nil", got, kind, err)
	}

	path := filepath.Join(dir, Name)
	if err := os.WriteFile(path, []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS == "windows" {
		t.Skip("platform-control: POSIX mode-bit unreadability")
	}
	if err := os.Chmod(path, 0); err != nil {
		t.Skipf("host-capability: this environment can read a mode-000 file: chmod failed: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, 0o644) })
	if _, err := os.ReadFile(path); err == nil {
		t.Skip("host-capability: this environment can read a mode-000 file; permission-only marker row unavailable")
	}
	got, kind, err := ReadState(dir)
	if got != nil || kind != stateread.KindUnreadable || err == nil {
		t.Fatalf("unreadable ReadState = (%v, %q, %v), want nil, unreadable, error", got, kind, err)
	}
	var readErr *stateread.Error
	if !strings.Contains(err.Error(), stateread.DiagUnreadable) || !strings.Contains(err.Error(), path) || !errors.As(err, &readErr) || readErr.Kind != stateread.KindUnreadable {
		t.Fatalf("unreadable marker error = %v, want typed %s for %s", err, stateread.DiagUnreadable, path)
	}
}

func TestCurrentDistinguishesAbsentAndUnreadable(t *testing.T) {
	dir, expected := install(t)
	if current, err := Current(t.TempDir(), expected); err != nil || current {
		t.Fatalf("Current with absent marker = (%v, %v), want false, nil", current, err)
	}
	if runtime.GOOS == "windows" {
		t.Skip("platform-control: POSIX mode-bit unreadability")
	}
	path := filepath.Join(dir, Name)
	if err := os.Chmod(path, 0); err != nil {
		t.Skipf("host-capability: this environment can read a mode-000 file: chmod failed: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, 0o644) })
	if _, err := os.ReadFile(path); err == nil {
		t.Skip("host-capability: this environment can read a mode-000 file; permission-only marker row unavailable")
	}
	if current, err := Current(dir, expected); err == nil || current || !strings.Contains(err.Error(), stateread.DiagUnreadable) {
		t.Fatalf("Current with unreadable marker = (%v, %v), want false and typed unreadable error", current, err)
	}
}

func TestReadStateDistinguishesInvalidMarkerFromUnreadable(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, Name)
	if err := os.WriteFile(path, []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, kind, err := ReadState(dir)
	var invalid *InvalidError
	if got != nil || kind != stateread.KindPresent || !errors.As(err, &invalid) {
		t.Fatalf("invalid ReadState = (%v, %q, %v), want nil, present, typed invalid-marker error", got, kind, err)
	}
	if invalid.Path != path || !strings.Contains(err.Error(), DiagInvalid) || strings.Contains(err.Error(), stateread.DiagUnreadable) {
		t.Fatalf("invalid marker diagnostic = %v, want %s for %s and no %s", err, DiagInvalid, path, stateread.DiagUnreadable)
	}
}

func TestCurrentTreatsInvalidMarkerAsNonCurrent(t *testing.T) {
	dir, expected := install(t)
	if err := os.WriteFile(filepath.Join(dir, Name), []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if current, err := Current(dir, expected); err != nil || current {
		t.Fatalf("Current with invalid marker = (%v, %v), want false, nil for re-derivation", current, err)
	}
}

func TestProductionMarkerCallersUseReadState(t *testing.T) {
	root := filepath.Clean(filepath.Join("..", ".."))
	for _, subtree := range []string{"internal", "cmd"} {
		err := filepath.WalkDir(filepath.Join(root, subtree), func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
			if err != nil {
				return err
			}
			markerImports := map[string]bool{}
			for _, imported := range file.Imports {
				importPath := strings.Trim(imported.Path.Value, "\"")
				if !strings.HasSuffix(importPath, "/internal/marker") {
					continue
				}
				alias := "marker"
				if imported.Name != nil {
					alias = imported.Name.Name
				}
				markerImports[alias] = true
			}
			ast.Inspect(file, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}
				selector, ok := call.Fun.(*ast.SelectorExpr)
				if !ok || selector.Sel.Name != "Read" {
					return true
				}
				base, ok := selector.X.(*ast.Ident)
				if ok && markerImports[base.Name] {
					t.Errorf("%s calls marker.Read, which collapses absence and read failure; use marker.ReadState", path)
				}
				return true
			})
			return nil
		})
		if err != nil {
			t.Fatalf("scan %s for collapsing marker reads: %v", subtree, err)
		}
	}
}
