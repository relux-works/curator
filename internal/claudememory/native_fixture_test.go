package claudememory

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/protocoljson"
)

// This optional exact-release oracle executes the distributed matcher and
// parent-memory filter, not the Claude application, an API or authentication.
// Fixtures have no symlinks: the injected path resolver is identity-only. It
// proves exclusion and retained project/profile entries at this filter boundary,
// not complete session discovery, import handling or runtime qualification.
func TestPinnedClaudeMemoryExclusion(t *testing.T) {
	path := os.Getenv("CURATOR_CLAUDE_MEMORY_SPECIMEN")
	if path == "" {
		t.Skip("set CURATOR_CLAUDE_MEMORY_SPECIMEN to the pinned public Claude 2.1.314 distribution")
	}
	blob, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(blob)
	if hex.EncodeToString(sum[:]) != "21d6e7b9a6f07b4cfcb7df13760b3b72aa25e084e13a5354413382987ba832ed" {
		t.Fatal("Claude memory specimen differs from the pinned release")
	}
	section := func(start, end string) []byte {
		t.Helper()
		a := bytes.Index(blob, []byte(start))
		if a < 0 {
			t.Fatal("missing native extraction start")
		}
		z := bytes.Index(blob[a:], []byte(end))
		if z < 0 {
			t.Fatal("missing native extraction end")
		}
		return blob[a : a+z]
	}
	matcher := section("var zf=x(function(hge,RI)", "var dl=x(function(eN)")
	filter := section("function t5t(){let e=pt().claudeMdExcludes", "function r5t(")
	node, err := exec.LookPath("node")
	if err != nil {
		t.Fatal("native oracle requires Node.js")
	}
	for _, name := range []string{"operator", "operator[1]*?{x}(y)!+@"} {
		t.Run(name, func(t *testing.T) {
			home := filepath.Join(t.TempDir(), name)
			native := filepath.Join(home, ".claude", "CLAUDE.md")
			project := filepath.Join(home, "project", "CLAUDE.md")
			managed := filepath.Join(home, "managed", "CLAUDE.md")
			for _, file := range []string{native, project, managed} {
				if err := os.MkdirAll(filepath.Dir(file), 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(file, []byte("fixture instructions\n"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			settings, err := Merge(nil, native)
			if err != nil {
				t.Fatal(err)
			}
			input, err := json.Marshal(map[string]any{"settings": json.RawMessage(settings), "native": filepath.ToSlash(native), "project": filepath.ToSlash(project), "managed": filepath.ToSlash(managed)})
			if err != nil {
				t.Fatal(err)
			}
			script := []byte(`const fs=require('node:fs'), path=require('node:path');
const x=fn=>{let mod;return()=>{if(!mod){mod={exports:{}};fn(mod.exports,mod)}return mod.exports}};
`)
			script = append(script, matcher...)
			script = append(script, []byte(`
const input=JSON.parse(fs.readFileSync(0,'utf8'));
let settings=input.settings;
const pt=()=>settings, le=()=>null, wx=path.posix.dirname;
const Po=(_context,p)=>({resolvedPath:p,isSymlink:false,isCanonical:true});
const AL=()=>false, Bq=()=>false;
const V6t={default:O6t()};
`)...)
			script = append(script, filter...)
			script = append(script, []byte(`
const entries=[{path:input.native,type:'Project'},{path:input.project,type:'Project'},{path:input.managed,type:'User'}];
const configured=U2e(entries);
const managedExcluded=o5t(input.native,'Managed');
settings={};
const missing=U2e(entries);
process.stdout.write(JSON.stringify({configured,missing,managedExcluded}));
`)...)
			file := filepath.Join(t.TempDir(), "native-memory.cjs")
			if err := os.WriteFile(file, script, 0600); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			command := exec.CommandContext(ctx, node, file)
			command.Stdin = bytes.NewReader(input)
			output, err := command.CombinedOutput()
			if command.Process != nil {
				t.Logf("owned native filter PID=%d", command.Process.Pid)
			}
			if err != nil {
				t.Fatalf("native memory filter invocation: %v\n%s", err, output)
			}
			if err := checkNativeMemoryResult(output, filepath.ToSlash(native), filepath.ToSlash(project), filepath.ToSlash(managed)); err != nil {
				t.Fatal(err)
			}
			t.Log("native excluded; project/profile retained; Managed policy retained; missing-exclusion control exposes native")
		})
	}
}

// Authored verdict logic stays in Go; JavaScript returns the upstream subject's
// results without deciding their validity. Pointer presence distinguishes a
// missing/null Managed result from the required explicit false value.
type nativeMemoryEntry struct {
	Path string `json:"path"`
	Type string `json:"type"`
}

type nativeMemoryResult struct {
	Configured      []nativeMemoryEntry `json:"configured"`
	Missing         []nativeMemoryEntry `json:"missing"`
	ManagedExcluded *bool               `json:"managedExcluded"`
}

func checkNativeMemoryResult(payload []byte, native, project, profile string) error {
	if err := protocoljson.Validate(payload); err != nil {
		return fmt.Errorf("invalid native memory result: %w", err)
	}
	var result *nativeMemoryResult
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&result); err != nil {
		return fmt.Errorf("invalid native memory result: %w", err)
	}
	if result == nil {
		return fmt.Errorf("invalid native memory result: expected an object")
	}
	expected := []nativeMemoryEntry{{project, "Project"}, {profile, "User"}}
	if !slices.Equal(result.Configured, expected) {
		return fmt.Errorf("configured memory entries differ: %v", result.Configured)
	}
	missing := append([]nativeMemoryEntry{{native, "Project"}}, expected...)
	if !slices.Equal(result.Missing, missing) {
		return fmt.Errorf("missing-exclusion control entries differ: %v", result.Missing)
	}
	if result.ManagedExcluded == nil {
		return fmt.Errorf("managed-policy exclusion result missing")
	}
	if *result.ManagedExcluded {
		return fmt.Errorf("managed policy was excluded")
	}
	return nil
}

// The Go oracle accepts the retained-entry positive control, detects each
// forbidden native result and refuses missing/malformed result data itself.
func TestNativeMemoryResultOracle(t *testing.T) {
	const valid = `{"configured":[{"path":"project","type":"Project"},{"path":"profile","type":"User"}],"missing":[{"path":"native","type":"Project"},{"path":"project","type":"Project"},{"path":"profile","type":"User"}],"managedExcluded":false}`
	check := func(payload string) error {
		return checkNativeMemoryResult([]byte(payload), "native", "project", "profile")
	}
	if err := check(valid); err != nil {
		t.Fatalf("valid native result rejected: %v", err)
	}
	for _, tc := range []struct{ name, payload, witness string }{
		{"empty", "", "invalid native memory result"},
		{"truncated", "{", "invalid native memory result"},
		{"null", "null", "invalid native memory result"},
		{"missing-results", "{}", "configured memory entries differ"},
		{"null-configured", strings.Replace(valid, `"configured":[{"path":"project","type":"Project"},{"path":"profile","type":"User"}]`, `"configured":null`, 1), "configured memory entries differ"},
		{"native-survives", strings.Replace(valid, `"configured":[`, `"configured":[{"path":"native","type":"Project"},`, 1), "configured memory entries differ"},
		{"project-lost", strings.Replace(valid, `{"path":"project","type":"Project"},`, "", 1), "configured memory entries differ"},
		{"profile-lost", strings.Replace(valid, `,{"path":"profile","type":"User"}`, "", 1), "configured memory entries differ"},
		{"wrong-type", strings.Replace(valid, `"type":"User"`, `"type":"Project"`, 1), "configured memory entries differ"},
		{"missing-control", strings.Replace(valid, `"missing":[{"path":"native","type":"Project"},`, `"missing":[`, 1), "missing-exclusion control entries differ"},
		{"null-control", strings.Replace(valid, `"missing":[{"path":"native","type":"Project"},{"path":"project","type":"Project"},{"path":"profile","type":"User"}]`, `"missing":null`, 1), "missing-exclusion control entries differ"},
		{"missing-managed", strings.Replace(valid, `,"managedExcluded":false`, "", 1), "managed-policy exclusion result missing"},
		{"null-managed", strings.Replace(valid, `"managedExcluded":false`, `"managedExcluded":null`, 1), "managed-policy exclusion result missing"},
		{"wrong-managed-type", strings.Replace(valid, `"managedExcluded":false`, `"managedExcluded":"false"`, 1), "invalid native memory result"},
		{"managed-excluded", strings.Replace(valid, `"managedExcluded":false`, `"managedExcluded":true`, 1), "managed policy was excluded"},
		{"duplicate-result", strings.Replace(valid, `"managedExcluded":false`, `"managedExcluded":false,"managedExcluded":true`, 1), "invalid native memory result"},
		{"trailing-result", valid + "{}", "invalid native memory result"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := check(tc.payload); err == nil || !strings.Contains(err.Error(), tc.witness) {
				t.Fatalf("native result did not fail intended witness %q: %v", tc.witness, err)
			}
		})
	}
}
