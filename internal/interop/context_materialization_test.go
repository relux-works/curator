package interop

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextmaterialize"
	"github.com/relux-works/curator/internal/contextpkg"
)

type environmentsVector struct {
	HeaderTypeLine string `json:"header_type_line"`
	HeaderCases    []struct {
		Name          string     `json:"name"`
		Lock          vectorLock `json:"lock"`
		LockSHA256    string     `json:"lock_sha256"`
		Precedence    struct{ Winner, Placement string }
		EmittedOrder  []string `json:"emitted_order"`
		ExpectedBytes string   `json:"expected_bytes"`
		LineCount     int      `json:"line_count"`
		SHA256        string   `json:"sha256"`
	} `json:"header_cases"`
	MaterializationCases []struct {
		Name        string     `json:"name"`
		Surface     string     `json:"surface"`
		Form        string     `json:"form"`
		Environment string     `json:"environment"`
		Lock        vectorLock `json:"lock"`
		LockSHA256  string     `json:"lock_sha256"`
		Precedence  struct{ Winner, Placement string }
		Packages    map[string]struct {
			HasContext bool `json:"has_context"`
			Modules    []struct {
				Path         string   `json:"path"`
				Class        string   `json:"class"`
				Environments []string `json:"environments"`
				Content      string   `json:"content"`
			} `json:"modules"`
		} `json:"packages"`
		EmittedOrder  []string `json:"emitted_order"`
		FileWritten   bool     `json:"file_written"`
		SurfaceSHA256 string   `json:"surface_sha256"`
		Files         []struct {
			Path     string `json:"path"`
			Bytes    int    `json:"bytes"`
			Expected string `json:"expected"`
			SHA256   string `json:"sha256"`
		} `json:"files"`
	} `json:"materialization_cases"`
}

func loadEnvironmentsVector(t *testing.T) (string, environmentsVector) {
	t.Helper()
	root := suiteRoot(t)
	vectorPath := filepath.Join(root, "vectors", "environments.json")
	payload, err := os.ReadFile(vectorPath)
	if errors.Is(err, os.ErrNotExist) {
		t.Skipf("conformance root %s publishes no vectors/environments.json (pre-environments suite; root-content)", root)
	}
	if err != nil {
		t.Fatal(err)
	}
	var vector environmentsVector
	if err := json.Unmarshal(payload, &vector); err != nil {
		t.Fatalf("decoding %s: %v", vectorPath, err)
	}
	if vector.HeaderTypeLine != contextmaterialize.HeaderTypeLine {
		t.Skipf("conformance root %s publishes header type line %q, not %q (is a pre-revision root)", root, vector.HeaderTypeLine, contextmaterialize.HeaderTypeLine)
	}
	return root, vector
}

func memberNames(order []contextlock.Member) []string {
	var names []string
	for _, member := range order {
		names = append(names, member.Name)
	}
	return names
}

// TestConformanceEnvironmentsHeader drives contextmaterialize.EmittedOrder and
// Header — the production header writer behind every root-context surface —
// against the header family of vectors/environments.json.
func TestConformanceEnvironmentsHeader(t *testing.T) {
	_, vector := loadEnvironmentsVector(t)
	if len(vector.HeaderCases) == 0 {
		t.Fatal("the vector declares no header cases")
	}
	for _, tc := range vector.HeaderCases {
		t.Run(tc.Name, func(t *testing.T) {
			lock := vectorLockToLock(t, tc.Lock)
			hash, err := lock.Hash()
			if err != nil {
				t.Fatal(err)
			}
			if hash != tc.LockSHA256 {
				t.Fatalf("lock hash %s, want %s", hash, tc.LockSHA256)
			}
			precedence := contextmaterialize.Precedence{Winner: tc.Precedence.Winner, Placement: tc.Precedence.Placement}
			order, err := contextmaterialize.EmittedOrder(lock, precedence)
			if err != nil {
				t.Fatal(err)
			}
			if got := memberNames(order); !reflect.DeepEqual(got, tc.EmittedOrder) {
				t.Fatalf("emitted order %v, want %v", got, tc.EmittedOrder)
			}
			header, err := contextmaterialize.Header(lock, hash, precedence, order)
			if err != nil {
				t.Fatal(err)
			}
			if string(header) != tc.ExpectedBytes {
				t.Fatalf("header bytes differ:\n got %q\nwant %q", header, tc.ExpectedBytes)
			}
			if got := strings.Count(string(header), "\n"); got != tc.LineCount {
				t.Fatalf("line count %d, want %d", got, tc.LineCount)
			}
			if got := contextmaterialize.FileHash(header); got != tc.SHA256 {
				t.Fatalf("header sha256 %s, want %s", got, tc.SHA256)
			}
		})
	}
}

// TestConformanceEnvironmentsMonolithic drives contextmaterialize.Monolithic
// and SystemPrompt — the production assemblers behind `profile use` — against
// every root-context case in the monolithic form and every system-prompt case
// of vectors/environments.json, byte for byte against the expected files, and
// checks the §5.6 surface hash. Referenced-form and MCP cases belong to a
// later stage and are recorded as skipped subtests, never as passes.
func TestConformanceEnvironmentsMonolithic(t *testing.T) {
	root, vector := loadEnvironmentsVector(t)
	if len(vector.MaterializationCases) == 0 {
		t.Fatal("the vector declares no materialization cases")
	}
	for _, tc := range vector.MaterializationCases {
		t.Run(tc.Name, func(t *testing.T) {
			if tc.Surface == "mcp" || tc.Form == contextmaterialize.FormReferenced {
				t.Skipf("surface %s form %s deferred to stage (b); set CURATOR_STAGE_B=1 when it lands", tc.Surface, tc.Form)
			}
			lock := vectorLockToLock(t, tc.Lock)
			hash, err := lock.Hash()
			if err != nil {
				t.Fatal(err)
			}
			if hash != tc.LockSHA256 {
				t.Fatalf("lock hash %s, want %s", hash, tc.LockSHA256)
			}
			precedence := contextmaterialize.Precedence{Winner: tc.Precedence.Winner, Placement: tc.Precedence.Placement}
			packages := map[string]contextmaterialize.Package{}
			for name, pkg := range tc.Packages {
				content := contextmaterialize.Package{HasContext: pkg.HasContext}
				for _, module := range pkg.Modules {
					content.Modules = append(content.Modules, contextmaterialize.Module{
						Module: contextpkg.Module{Path: module.Path, Class: module.Class, Environments: module.Environments},
						Bytes:  []byte(module.Content),
					})
				}
				packages[name] = content
			}
			order, err := contextmaterialize.EmittedOrder(lock, precedence)
			if err != nil {
				t.Fatal(err)
			}
			if got := memberNames(order); !reflect.DeepEqual(got, tc.EmittedOrder) {
				t.Fatalf("emitted order %v, want %v", got, tc.EmittedOrder)
			}
			var document []byte
			var written bool
			switch tc.Surface {
			case "root-context":
				document, written, err = contextmaterialize.Monolithic(lock, hash, precedence, tc.Environment, packages)
			case "system-prompt":
				document, written, err = contextmaterialize.SystemPrompt(lock, precedence, tc.Environment, packages)
			default:
				t.Fatalf("unknown surface %q", tc.Surface)
			}
			if err != nil {
				t.Fatal(err)
			}
			if written != tc.FileWritten {
				t.Fatalf("file_written %v, want %v", written, tc.FileWritten)
			}
			if !written {
				if len(tc.Files) != 0 {
					t.Fatalf("the vector lists files for an unwritten surface")
				}
				return
			}
			if len(tc.Files) != 1 {
				t.Fatalf("expected exactly one file for a %s %s surface, vector lists %d", tc.Form, tc.Surface, len(tc.Files))
			}
			file := tc.Files[0]
			want, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(file.Expected)))
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(document, want) {
				t.Fatalf("%s bytes differ from %s:\n got %q\nwant %q", file.Path, file.Expected, document, want)
			}
			if len(document) != file.Bytes || contextmaterialize.FileHash(document) != file.SHA256 {
				t.Fatalf("%s: %d bytes %s, want %d bytes %s", file.Path, len(document), contextmaterialize.FileHash(document), file.Bytes, file.SHA256)
			}
			if got := contextmaterialize.SurfaceHash(map[string][]byte{file.Path: document}); got != tc.SurfaceSHA256 {
				t.Fatalf("surface hash %s, want %s", got, tc.SurfaceSHA256)
			}
		})
	}
}
