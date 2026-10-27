package envprofile

import (
	"os"
	"testing"

	"github.com/relux-works/curator/internal/envregistry"
)

// credentialPreimage reads only synthetic fixture bytes to detect mutation;
// production inspection has no content-reader dependency.
func credentialPreimage(t *testing.T, paths ...string) func() {
	t.Helper()
	type image struct {
		info   os.FileInfo
		data   string
		absent bool
	}
	capture := func(path string) image {
		info, err := os.Lstat(path)
		if os.IsNotExist(err) {
			return image{absent: true}
		}
		if err != nil {
			t.Fatal(err)
		}
		out := image{info: info}
		if info.Mode()&os.ModeSymlink != 0 {
			out.data, err = os.Readlink(path)
		} else if info.Mode().IsRegular() {
			var b []byte
			b, err = os.ReadFile(path)
			out.data = string(b)
		}
		if err != nil {
			t.Fatal(err)
		}
		return out
	}
	before := map[string]image{}
	for _, path := range paths {
		before[path] = capture(path)
	}
	return func() {
		t.Helper()
		for _, path := range paths {
			a, b := before[path], capture(path)
			if a.absent != b.absent {
				t.Fatal("diagnostic changed fixture path existence")
			}
			if a.absent {
				continue
			}
			if !os.SameFile(a.info, b.info) || a.info.Mode() != b.info.Mode() || a.info.Size() != b.info.Size() || !a.info.ModTime().Equal(b.info.ModTime()) || a.data != b.data {
				t.Fatal("diagnostic changed fixture preimage")
			}
		}
	}
}

// Other providers and non-credential paths retain their existing verifier.
func TestClaudeCredentialDiagnosticsScope(t *testing.T) {
	for _, tc := range []struct{ id, path string }{{envregistry.CodexCLI, "auth.json"}, {envregistry.ClaudeCode, "other.json"}} {
		adapter, _ := envregistry.ByID(tc.id)
		req := ResolveRequest{passthroughLstat: func(string) (os.FileInfo, error) {
			t.Fatal("out-of-scope path inspected")
			return nil, os.ErrPermission
		}}
		verdict := &verification{}
		if verdict.checkClaudeCredentialLink(&req, &homePlan{adapter: adapter}, tc.path, "unused") {
			t.Fatal("out-of-scope path handled")
		}
		if len(verdict.reasons) != 0 || len(verdict.warnings) != 0 {
			t.Fatal("out-of-scope finding")
		}
	}
}
