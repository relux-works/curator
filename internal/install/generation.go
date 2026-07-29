package install

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/relux-works/curator/internal/devsub"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/scopes"
)

// A declaration document is one shared file whose *bytes* select installed
// content: the project manifest, the machine-wide manifest, the project
// development substitution manifest, and the machine-home hybrid activation
// manifest. Every one of them is read in the window before the manager-home lock
// by a writer that takes none of this run's locks (Spec §6.1 step 6), so each is
// an optimistic observation revalidated under the lock in step 10.
//
// An observation is only sound if its recorded generation belongs to the bytes
// the closure was actually resolved from. Digesting the path and then reading it
// as a second, separate operation does not give that: a supported
// A -> B -> A rewrite around the read leaves both the recorded and the rechecked
// digest at A while the closure was built from the transient B, and the run
// commits context, shims, adapters, and consumer state for a declaration set
// that is neither the one it recorded nor the one on disk. `curator add|remove`,
// `curator global add|remove`, and `curator hybrid add|rm` all rewrite their
// manifest in place with os.WriteFile, so that window is not hypothetical.
//
// readDocument therefore reads a declaration document exactly *once* and returns
// the bytes together with the generation of those exact bytes. The parser
// consumes the returned payload, so "the generation we recorded" and "the bytes
// we parsed" are the same object by construction and cannot drift apart. A torn
// read is safe for the same reason: whatever bytes came back are what the parse
// saw and what the generation covers, so the under-lock recheck sees the settled
// file, reports a change, and restarts instead of committing.

// declarationGenerationDomain separates declaration generations from every other
// digest in the manager. A generation is only ever compared with another
// generation of the same document, so it needs no cross-format meaning.
var declarationGenerationDomain = []byte("curator-declaration-generation-v1\x00")

// documentAbsent is the generation of a document that is not there. It is
// explicit so the recheck can tell "still absent" from "appeared", exactly like
// transaction.DigestAbsent does for a target preimage.
const documentAbsent = "absent"

// afterDocumentOpen is nil in production. Tests assign it to write a declaration
// document in the window between the open and the read, which is the only window
// an in-place writer can use to make a parse consume bytes a separate path
// digest would have missed, and prove that the recorded generation still belongs
// to the bytes the parse actually consumed.
var afterDocumentOpen func(path string)

// document is one immutable read of a declaration document.
type document struct {
	// payload is the exact bytes read, and the only bytes any parser may see.
	payload []byte
	// exists reports whether the document was present at all. An absent document
	// has no payload and each caller applies its own absent semantics.
	exists bool
	// generation identifies payload, or documentAbsent when exists is false.
	generation string
}

// readDocument reads one declaration document exactly once and returns its bytes
// with the generation of those bytes.
func readDocument(path string) (result document, err error) {
	file, err := os.Open(path) // #nosec G304 -- a manifest path derived from the project root or the manager home
	if os.IsNotExist(err) {
		return document{generation: documentAbsent}, nil
	}
	if err != nil {
		return document{}, err
	}
	defer func() { err = errors.Join(err, file.Close()) }()
	if afterDocumentOpen != nil {
		afterDocumentOpen(path)
	}
	info, err := file.Stat()
	if err != nil {
		return document{}, err
	}
	if !info.Mode().IsRegular() {
		return document{}, fmt.Errorf("%s is not a regular file", path)
	}
	payload, err := io.ReadAll(file)
	if err != nil {
		return document{}, err
	}
	return document{payload: payload, exists: true, generation: digestDeclaration(payload)}, nil
}

// documentGeneration re-reads one declaration document and reports the
// generation of the bytes a parse would consume now. It is the recheck side of
// readDocument and must stay the only reader that rechecks a document
// observation: two readers that disagree about what a path means would restart
// every run forever.
func documentGeneration(path string) string {
	current, err := readDocument(path)
	if err != nil {
		// An unreadable document is itself a stable observation: the recheck
		// reproduces the same marker and only a change restarts.
		return "unreadable:" + err.Error()
	}
	return current.generation
}

// digestDeclaration identifies one declaration payload. The file mode is
// deliberately not part of it: the parser consumes bytes, so a chmod selects
// nothing installed and must not manufacture a restart.
func digestDeclaration(payload []byte) string {
	hash := sha256.New()
	_, _ = hash.Write(declarationGenerationDomain)
	_ = binary.Write(hash, binary.BigEndian, uint64(len(payload)))
	_, _ = hash.Write(payload)
	return "sha256:" + hex.EncodeToString(hash.Sum(nil))
}

// readManifestDocument reads and parses one Skillfile — project or machine-wide,
// which share both format and writer — and returns the generation of the exact
// bytes it parsed. An absent manifest is (nil, documentAbsent, nil), matching
// manifest.Load, so each scope keeps its own absent semantics.
func readManifestDocument(root string) (*manifest.Manifest, string, error) {
	path := manifest.PathIn(root)
	current, err := readDocument(path)
	if err != nil {
		return nil, "", err
	}
	if !current.exists {
		return nil, current.generation, nil
	}
	parsed, err := manifest.ParseBytes(current.payload, path)
	if err != nil {
		return nil, "", err
	}
	return parsed, current.generation, nil
}

// readSubstitutionsDocument reads and parses the project development
// substitution manifest and returns the generation of the exact bytes it parsed.
// An absent file yields an empty map, matching devsub.Load.
func readSubstitutionsDocument(projectRoot string) (map[string]devsub.Substitution, string, error) {
	path := devsub.PathIn(projectRoot)
	current, err := readDocument(path)
	if err != nil {
		return nil, "", err
	}
	if !current.exists {
		return map[string]devsub.Substitution{}, current.generation, nil
	}
	parsed, err := devsub.ParseBytes(current.payload, projectRoot)
	if err != nil {
		return nil, "", err
	}
	return parsed, current.generation, nil
}

// readHybridDocument reads and parses the machine-home hybrid activation
// manifest and returns the generation of the exact bytes it parsed. An absent
// file yields no declarations, matching scopes.LoadHybridDecls.
func readHybridDocument(home string) ([]scopes.HybridDecl, string, error) {
	path := scopes.HybridManifestPath(home)
	current, err := readDocument(path)
	if err != nil {
		return nil, "", err
	}
	if !current.exists {
		return nil, current.generation, nil
	}
	parsed, err := scopes.ParseHybridDecls(current.payload, path)
	if err != nil {
		return nil, "", err
	}
	return parsed, current.generation, nil
}
