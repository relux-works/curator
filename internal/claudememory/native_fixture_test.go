package claudememory

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
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
const actual=U2e(entries).map(x=>x.path);
if(actual.includes(input.native))throw Error('native parent memory survived exclusion');
if(!actual.includes(input.project)||!actual.includes(input.managed))throw Error('project or profile instructions were excluded');
if(o5t(input.native,'Managed'))throw Error('managed policy was excluded');
settings={};
if(!U2e(entries).some(x=>x.path===input.native))throw Error('missing-exclusion control did not detect native parent memory');
console.log('native excluded; project/profile retained; missing-exclusion control exposes native');
`)...)
			file := filepath.Join(t.TempDir(), "native-memory.cjs")
			if err := os.WriteFile(file, script, 0600); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			command := exec.CommandContext(ctx, node, file)
			command.Stdin = bytes.NewReader(input)
			if output, err := command.CombinedOutput(); err != nil {
				t.Fatalf("native memory filter control: %v\n%s", err, output)
			} else {
				t.Log(string(output))
			}
		})
	}
}
