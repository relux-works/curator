// Package crossmanager provides an implementation-neutral harness for running
// an external Curator Protocol corpus against independent manager processes.
package crossmanager

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

const (
	// CorpusBoundaryV1 identifies the consumer-side manifest contract. It is
	// independent of a particular candidate checkout and can be rebound to an
	// accepted corpus with the same protocol version.
	CorpusBoundaryV1 = "rc5-external-repository-interop/v1"
	// ProtocolRC5 is the exact protocol revision required by this corpus.
	ProtocolRC5 = "1.0.0-rc.5"
	// CorpusRC5 is the exact accepted corpus version.
	CorpusRC5 = "rc5-external-repository-interop-v1"
)

// Boundary pins the protocol and consumer contract expected at a corpus root.
type Boundary struct {
	Version         string `json:"version"`
	ProtocolVersion string `json:"protocol_version"`
}

// RC5Boundary is the explicit boundary used by the provisional rc.5 consumer.
var RC5Boundary = Boundary{Version: CorpusBoundaryV1, ProtocolVersion: ProtocolRC5}

type manifest struct {
	SchemaVersion   int             `json:"schema_version"`
	CorpusVersion   string          `json:"corpus_version"`
	ProtocolVersion string          `json:"protocol_version"`
	GeneratedAt     string          `json:"generated_at"`
	Generator       string          `json:"generator"`
	Files           []manifestEntry `json:"files"`
}

type manifestEntry struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
}

// Corpus is a read-only view of files authenticated by manifest.json.
type Corpus struct {
	root           string
	boundary       Boundary
	manifestSHA256 string
	files          map[string]string
}

// OpenCorpus validates a corpus manifest against the explicit boundary.
func OpenCorpus(root string, boundary Boundary) (*Corpus, error) {
	if boundary.Version != CorpusBoundaryV1 {
		return nil, fmt.Errorf("unsupported corpus boundary %q", boundary.Version)
	}
	if boundary.ProtocolVersion == "" {
		return nil, fmt.Errorf("corpus protocol version is required")
	}
	absolute, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve corpus root: %w", err)
	}
	// #nosec G304 -- absolute is the caller-selected corpus root and the leaf is fixed.
	payload, err := os.ReadFile(filepath.Join(absolute, "manifest.json"))
	if err != nil {
		return nil, fmt.Errorf("read corpus manifest: %w", err)
	}
	var parsed manifest
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&parsed); err != nil {
		return nil, fmt.Errorf("decode corpus manifest: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return nil, fmt.Errorf("decode corpus manifest: trailing JSON value")
	}
	if parsed.ProtocolVersion != boundary.ProtocolVersion {
		return nil, fmt.Errorf("corpus protocol version %q does not match boundary %q", parsed.ProtocolVersion, boundary.ProtocolVersion)
	}
	if parsed.SchemaVersion != 1 || parsed.CorpusVersion != CorpusRC5 {
		return nil, fmt.Errorf("unsupported corpus identity schema=%d corpus=%q", parsed.SchemaVersion, parsed.CorpusVersion)
	}
	files := make(map[string]string, len(parsed.Files))
	for _, entry := range parsed.Files {
		if err := validateCorpusPath(entry.Path); err != nil {
			return nil, fmt.Errorf("manifest file %q: %w", entry.Path, err)
		}
		if _, exists := files[entry.Path]; exists {
			return nil, fmt.Errorf("manifest contains duplicate file %q", entry.Path)
		}
		if _, err := parseDigest(entry.SHA256); err != nil {
			return nil, fmt.Errorf("manifest file %q: %w", entry.Path, err)
		}
		if entry.Size < 0 {
			return nil, fmt.Errorf("manifest file %q has negative size", entry.Path)
		}
		files[entry.Path] = entry.SHA256
	}
	digest := sha256.Sum256(payload)
	return &Corpus{
		root:           absolute,
		boundary:       boundary,
		manifestSHA256: "sha256:" + hex.EncodeToString(digest[:]),
		files:          files,
	}, nil
}

// Evidence returns the replaceable corpus identity used by reports.
func (c *Corpus) Evidence() CorpusEvidence {
	return CorpusEvidence{
		Boundary:        c.boundary.Version,
		ProtocolVersion: c.boundary.ProtocolVersion,
		ManifestSHA256:  c.manifestSHA256,
	}
}

// Entries returns sorted manifest paths under prefix without reading bytes.
func (c *Corpus) Entries(prefix string) []string {
	entries := make([]string, 0)
	for name := range c.files {
		if strings.HasPrefix(name, prefix) {
			entries = append(entries, name)
		}
	}
	sort.Strings(entries)
	return entries
}

// Read verifies and returns one manifest-listed file.
func (c *Corpus) Read(name string) ([]byte, string, error) {
	if err := validateCorpusPath(name); err != nil {
		return nil, "", err
	}
	want, ok := c.files[name]
	if !ok {
		return nil, "", fmt.Errorf("corpus file %q is not listed in manifest.json", name)
	}
	payload, err := os.ReadFile(filepath.Join(c.root, filepath.FromSlash(name)))
	if err != nil {
		return nil, "", fmt.Errorf("read corpus file %q: %w", name, err)
	}
	digest := sha256.Sum256(payload)
	got := "sha256:" + hex.EncodeToString(digest[:])
	if got != want {
		return nil, "", fmt.Errorf("corpus file %q digest %s does not match manifest %s", name, got, want)
	}
	return payload, got, nil
}

func validateCorpusPath(name string) error {
	if name == "" || strings.Contains(name, "\\") || path.IsAbs(name) || path.Clean(name) != name || name == "." || strings.HasPrefix(name, "../") {
		return fmt.Errorf("path must be a clean relative slash path")
	}
	return nil
}

func parseDigest(value string) ([]byte, error) {
	encoded, ok := strings.CutPrefix(value, "sha256:")
	if !ok {
		return nil, fmt.Errorf("digest must use sha256 prefix")
	}
	digest, err := hex.DecodeString(encoded)
	if err != nil || len(digest) != sha256.Size {
		return nil, fmt.Errorf("digest must contain 64 lowercase hexadecimal characters")
	}
	if encoded != strings.ToLower(encoded) {
		return nil, fmt.Errorf("digest must contain 64 lowercase hexadecimal characters")
	}
	return digest, nil
}
