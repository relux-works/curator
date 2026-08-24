package crossconformance_test

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/relux-works/curator/internal/closuregraph"
)

// sourcePackage is one reusable pure-source dependency payload. Every adapter
// path in this suite consumes the same three packages so that a difference in
// outcome is a difference in the adapter, not in the fixture.
type sourcePackage struct {
	name, version string
	metadata      map[string]any
	files         map[string][]byte
}

// crossPackages is the shared dependency set: one package with a transitive
// edge, one leaf, and one target-conditional optional package that must stay
// in the selection-neutral capture and be pruned only by a binding.
func crossPackages() []sourcePackage {
	return []sourcePackage{
		{name: "a", version: "1.0.0", metadata: map[string]any{"dependencies": map[string]string{"b": "^1.0.0"}}, files: map[string][]byte{"index.js": []byte("module.exports = require('b')\n")}},
		{name: "b", version: "1.0.0", metadata: map[string]any{}, files: map[string][]byte{"index.js": []byte("module.exports = 'b'\n")}},
		{name: "opt", version: "1.0.0", metadata: map[string]any{"os": []string{"linux"}}, files: map[string][]byte{"index.js": []byte("module.exports = 'optional'\n")}},
	}
}

// buildTGZ produces the npm-shaped package tarball every Node manager accepts
// as raw registry bytes.
func buildTGZ(t *testing.T, pkg sourcePackage) []byte {
	t.Helper()
	metadata := map[string]any{"name": pkg.name, "version": pkg.version}
	for key, value := range pkg.metadata {
		metadata[key] = value
	}
	files := map[string][]byte{"package/package.json": mustJSON(t, metadata)}
	for name, payload := range pkg.files {
		files["package/"+name] = payload
	}
	return tarball(t, files)
}

// tarball gzips a deterministic tar of the given members.
func tarball(t *testing.T, files map[string][]byte) []byte {
	t.Helper()
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	var output bytes.Buffer
	compressor := gzip.NewWriter(&output)
	archive := tar.NewWriter(compressor)
	for _, name := range names {
		payload := files[name]
		if err := archive.WriteHeader(&tar.Header{Name: name, Mode: 0o644, Size: int64(len(payload))}); err != nil {
			t.Fatal(err)
		}
		if _, err := archive.Write(payload); err != nil {
			t.Fatal(err)
		}
	}
	if err := archive.Close(); err != nil {
		t.Fatal(err)
	}
	if err := compressor.Close(); err != nil {
		t.Fatal(err)
	}
	return output.Bytes()
}

// cacheZIP produces the normalized modern Yarn cache archive layout for one
// package: every member lives under node_modules/<name>/ and is stored.
func cacheZIP(t *testing.T, pkg sourcePackage) []byte {
	t.Helper()
	metadata := map[string]any{"name": pkg.name, "version": pkg.version}
	for key, value := range pkg.metadata {
		metadata[key] = value
	}
	members := map[string][]byte{"package.json": mustJSON(t, metadata)}
	for name, payload := range pkg.files {
		members[name] = payload
	}
	names := make([]string, 0, len(members))
	for name := range members {
		names = append(names, name)
	}
	sort.Strings(names)
	var output bytes.Buffer
	writer := zip.NewWriter(&output)
	for _, name := range names {
		header := &zip.FileHeader{Name: "node_modules/" + pkg.name + "/" + name, Method: zip.Store}
		header.SetMode(0o644)
		entry, err := writer.CreateHeader(header)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = entry.Write(members[name]); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return output.Bytes()
}

// withExtraFiles returns the package with additional payload members. The
// cross suite uses it to inject one shared compiled fixture into every
// ecosystem's dependency bytes without changing anything else.
func withExtraFiles(pkg sourcePackage, extra map[string][]byte) sourcePackage {
	if len(extra) == 0 {
		return pkg
	}
	files := map[string][]byte{}
	for name, payload := range pkg.files {
		files[name] = payload
	}
	for name, payload := range extra {
		files[name] = payload
	}
	return sourcePackage{name: pkg.name, version: pkg.version, metadata: pkg.metadata, files: files}
}

func mustJSON(t *testing.T, value any) []byte {
	t.Helper()
	payload, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return payload
}

func writeFile(t *testing.T, path string, payload []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, payload, 0o600); err != nil {
		t.Fatal(err)
	}
}

func writeTemp(t *testing.T, name string, payload []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	writeFile(t, path, payload)
	return path
}

func sriSHA512(payload []byte) string {
	sum := sha512.Sum512(payload)
	return "sha512-" + base64.StdEncoding.EncodeToString(sum[:])
}

func digestID(payload []byte) closuregraph.ID {
	sum := sha256.Sum256(payload)
	return closuregraph.ID("sha256:" + hex.EncodeToString(sum[:]))
}

// releaseTree restores write permission so a capture store rooted in a test
// temporary directory can be removed by the test framework.
func releaseTree(t *testing.T, root string) {
	t.Helper()
	t.Cleanup(func() {
		_ = filepath.WalkDir(root, func(current string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return nil
			}
			if entry.IsDir() {
				_ = os.Chmod(current, 0o700)
			} else {
				_ = os.Chmod(current, 0o600)
			}
			return nil
		})
	})
}
