package projectadapter

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	_ "embed"
	"fmt"
	"io"
)

// The archive contains the unmodified native system-skill bundle emitted by
// the qualified Codex 0.162.1 macOS arm64 executable. All 48
// asset files also match rust-v0.162.1 Git blobs; the 49th is its native marker.
// See codex-system-skills.NOTICE.md for provenance and included licenses.
//
//go:embed codex-system-skills.tar.gz
var codexSystemBundle []byte

const codexSystemBundleSHA = "dbfd0e484b3dae9a72f197153c60cdf2e6bcf57536294def3d081a663718770f"

func codexSystemSkills(archive []byte) (map[string][]byte, error) {
	if fmt.Sprintf("%x", sha256.Sum256(archive)) != codexSystemBundleSHA {
		return nil, fmt.Errorf("project_context_unsupported: pinned native bundle digest differs")
	}
	zipped, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return nil, err
	}
	defer zipped.Close()
	reader := tar.NewReader(zipped)
	files := map[string][]byte{}
	for {
		header, err := reader.Next()
		if err == io.EOF {
			return files, nil
		}
		if err != nil {
			return nil, err
		}
		// The archive is a build-time pinned source, never a caller/native-home
		// input. Keep extraction in memory; no tar path is used for direct writes.
		b, err := io.ReadAll(reader)
		if err != nil {
			return nil, err
		}
		files["skills/.system/"+header.Name] = b
	}
}
