package closureexec

import (
	"bytes"
	"compress/zlib"
	"context"
	"crypto/sha1" // #nosec G505 -- Git object-format sha1 compatibility, not a security digest.
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/closuregraph"
	"github.com/relux-works/curator/internal/protocoljson"
)

// SourceControlMirrorRunner is the manager-owned deterministic
// source-control-mirror-v1 transform. It consumes an admitted source tree and
// an admitted raw Git commit object, writes a minimal shallow bare repository,
// and emits one exact canonical inventory file for the derivation receipt.
type SourceControlMirrorRunner struct {
	ExecutionRoot string
	OutputRoot    string
	Delegate      PortableProcessRunner
}

// NewSourceControlMirrorRunner binds absent task-private output locations.
func NewSourceControlMirrorRunner(executionRoot, outputRoot string) (*SourceControlMirrorRunner, error) {
	execution, err := filepath.Abs(executionRoot)
	if err != nil || executionRoot == "" {
		return nil, fmt.Errorf("mirror execution root is invalid")
	}
	output, err := filepath.Abs(outputRoot)
	if err != nil || outputRoot == "" || !pathWithin(execution, output) || execution == output {
		return nil, fmt.Errorf("mirror evidence root must be a child of its execution root")
	}
	return &SourceControlMirrorRunner{ExecutionRoot: execution, OutputRoot: output}, nil
}

// Run implements PortableProcessRunner without launching an adapter-controlled
// process. The closed transform implementation itself is the manager action.
func (runner *SourceControlMirrorRunner) Run(ctx context.Context, request ExecutionRequest) (PortableRunResult, error) {
	permit := request.Permit
	if runner != nil && !strings.HasPrefix(permit.InvocationKey, "source-control-mirror-v1:") {
		if runner.Delegate == nil {
			return PortableRunResult{}, failure("portable_runner_missing", "mirror runner delegate is absent")
		}
		return runner.Delegate.Run(ctx, request)
	}
	if runner == nil || permit.InvocationSubtype != DerivationMirror || !strings.HasPrefix(permit.InvocationKey, "source-control-mirror-v1:sha256:") || permit.Network != "none" {
		return PortableRunResult{}, failure("closure_derivation_unauthorized", "source-control mirror transform permit is invalid")
	}
	if len(permit.LocalOutputs) != 1 || permit.LocalOutputs[0].SchemaID != "source-control-mirror-v1" || permit.LocalOutputs[0].Path != permit.Environment["CURATOR_MIRROR_ROOT"] {
		return PortableRunResult{}, failure("closure_derivation_unauthorized", "source-control mirror local output declaration is invalid")
	}
	if err := ensureEmptyDirectory(runner.OutputRoot); err != nil {
		return PortableRunResult{}, err
	}
	mirrorRoot, err := resolveManagerPath(runner.ExecutionRoot, permit.Environment["CURATOR_MIRROR_ROOT"])
	if err != nil || !pathWithin(runner.ExecutionRoot, mirrorRoot) || pathWithin(runner.OutputRoot, mirrorRoot) || pathWithin(mirrorRoot, runner.OutputRoot) {
		return PortableRunResult{}, failure("closure_derivation_unauthorized", "source-control mirror destination is invalid")
	}
	if _, statErr := os.Lstat(mirrorRoot); !os.IsNotExist(statErr) {
		return PortableRunResult{}, failure("closure_write_undeclared", "source-control mirror destination is not absent")
	}
	var sourceRoot, commitPath string
	for _, input := range request.Inputs {
		protected, inputErr := input.ProtectedPath()
		if inputErr != nil {
			return PortableRunResult{}, inputErr
		}
		if input.IsTree() {
			if sourceRoot != "" {
				return PortableRunResult{}, failure("closure_derivation_unauthorized", "mirror transform has duplicate source trees")
			}
			sourceRoot = protected
		} else {
			if commitPath != "" {
				return PortableRunResult{}, failure("closure_derivation_unauthorized", "mirror transform has duplicate commit evidence")
			}
			commitPath = protected
		}
	}
	if sourceRoot == "" || commitPath == "" {
		return PortableRunResult{}, failure("closure_derivation_unauthorized", "mirror transform requires admitted source and commit evidence")
	}
	commitPayload, err := os.ReadFile(commitPath) // #nosec G304 -- opaque admitted commit-evidence handle.
	if err != nil {
		return PortableRunResult{}, err
	}
	revision, tree := strings.ToLower(permit.Environment["CURATOR_GIT_REVISION"]), strings.ToLower(permit.Environment["CURATOR_GIT_TREE"])
	if !gitObjectID(revision) || !gitObjectID(tree) {
		return PortableRunResult{}, failure("closure_derivation_unauthorized", "mirror transform Git identities are invalid")
	}
	if err = os.MkdirAll(filepath.Join(mirrorRoot, "objects"), 0o700); err != nil {
		return PortableRunResult{}, err
	}
	observedTree, err := writeGitTree(mirrorRoot, sourceRoot, ".")
	if err != nil {
		return PortableRunResult{}, err
	}
	if observedTree != tree {
		return PortableRunResult{}, failure("closure_derivation_drift", "admitted source does not reproduce the acquisition Git tree")
	}
	if !bytes.HasPrefix(commitPayload, []byte("tree "+tree+"\n")) {
		return PortableRunResult{}, failure("closure_derivation_drift", "acquisition commit object names another tree")
	}
	observedRevision, err := writeGitObject(mirrorRoot, "commit", commitPayload)
	if err != nil || observedRevision != revision {
		return PortableRunResult{}, failure("closure_derivation_drift", "acquisition commit object does not preserve the pinned revision")
	}
	if err = os.MkdirAll(filepath.Join(mirrorRoot, "refs", "heads"), 0o700); err != nil {
		return PortableRunResult{}, err
	}
	files := map[string]string{"HEAD": "ref: refs/heads/curator\n", "config": "[core]\n\trepositoryformatversion = 0\n\tbare = true\n", "description": "Curator source-control mirror\n", "refs/heads/curator": revision + "\n", "shallow": revision + "\n"}
	for relative, payload := range files {
		pathValue := filepath.Join(mirrorRoot, filepath.FromSlash(relative))
		if err = os.MkdirAll(filepath.Dir(pathValue), 0o700); err != nil {
			return PortableRunResult{}, err
		}
		if err = os.WriteFile(pathValue, []byte(payload), 0o600); err != nil {
			return PortableRunResult{}, err
		}
	}
	mirrorDigest, nodes, err := mirrorTreeIdentity(mirrorRoot)
	if err != nil {
		return PortableRunResult{}, err
	}
	evidence := map[string]any{"acquisition_receipt_id": strings.TrimPrefix(permit.InvocationKey, "source-control-mirror-v1:"), "git_tree": tree, "kind": permit.Environment["CURATOR_SOURCE_CONTROL_KIND"], "mirror_digest": string(mirrorDigest), "nodes": nodes, "revision": revision, "schema_id": "source-control-mirror-v1"}
	payload, err := protocoljson.MarshalCanonical(evidence)
	if err != nil || len(permit.ExpectedEvidence) != 1 || permit.ExpectedEvidence[0].SchemaID != "source-control-mirror-v1" {
		return PortableRunResult{}, failure("closure_derivation_unauthorized", "mirror transform evidence declaration is invalid")
	}
	evidencePath, err := safeExecutionPath(runner.OutputRoot, permit.ExpectedEvidence[0].Path)
	if err != nil {
		return PortableRunResult{}, err
	}
	if err = os.MkdirAll(filepath.Dir(evidencePath), 0o700); err != nil {
		return PortableRunResult{}, err
	}
	if err = os.WriteFile(evidencePath, payload, 0o600); err != nil {
		return PortableRunResult{}, err
	}
	return PortableRunResult{ExitCode: 0, OutputRoot: runner.OutputRoot, EvidenceRoot: runner.OutputRoot}, nil
}

func writeGitTree(gitRoot, sourceRoot, relative string) (string, error) {
	root := sourceRoot
	if relative != "." {
		root = filepath.Join(sourceRoot, filepath.FromSlash(relative))
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		return "", err
	}
	type treeEntry struct{ name, sortName, mode, object string }
	values := make([]treeEntry, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		logical := name
		if relative != "." {
			logical = relative + "/" + name
		}
		info, infoErr := entry.Info()
		if infoErr != nil || info.Mode()&fs.ModeSymlink != 0 {
			return "", failure("closure_derivation_drift", "admitted source contains an invalid Git tree member")
		}
		if entry.IsDir() {
			object, treeErr := writeGitTree(gitRoot, sourceRoot, logical)
			if treeErr != nil {
				return "", treeErr
			}
			values = append(values, treeEntry{name: name, sortName: name + "/", mode: "40000", object: object})
			continue
		}
		if !info.Mode().IsRegular() {
			return "", failure("closure_derivation_drift", "admitted source contains a special Git tree member")
		}
		payload, readErr := os.ReadFile(filepath.Join(sourceRoot, filepath.FromSlash(logical))) // #nosec G304 -- Walk is rooted in the admitted source tree.
		if readErr != nil {
			return "", readErr
		}
		object, writeErr := writeGitObject(gitRoot, "blob", payload)
		if writeErr != nil {
			return "", writeErr
		}
		mode := "100644"
		if info.Mode().Perm()&0o100 != 0 {
			mode = "100755"
		}
		values = append(values, treeEntry{name: name, sortName: name, mode: mode, object: object})
	}
	sort.Slice(values, func(i, j int) bool { return values[i].sortName < values[j].sortName })
	var tree bytes.Buffer
	for _, entry := range values {
		tree.WriteString(entry.mode + " " + entry.name)
		tree.WriteByte(0)
		decoded, _ := hex.DecodeString(entry.object)
		tree.Write(decoded)
	}
	return writeGitObject(gitRoot, "tree", tree.Bytes())
}

func writeGitObject(gitRoot, kind string, payload []byte) (string, error) {
	header := []byte(fmt.Sprintf("%s %d\x00", kind, len(payload)))
	object := append(header, payload...)
	digest := sha1.Sum(object) // #nosec G401 -- exact Git sha1 object identity.
	id := hex.EncodeToString(digest[:])
	pathValue := filepath.Join(gitRoot, "objects", id[:2], id[2:])
	if _, err := os.Lstat(pathValue); err == nil {
		return id, nil
	}
	if err := os.MkdirAll(filepath.Dir(pathValue), 0o700); err != nil {
		return "", err
	}
	file, err := os.OpenFile(pathValue, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600) // #nosec G304 -- path is the manager root plus a fixed-shape Git object ID derived above.
	if err != nil {
		return "", err
	}
	writer := zlib.NewWriter(file)
	_, writeErr := writer.Write(object)
	closeWriterErr, closeFileErr := writer.Close(), file.Close()
	if writeErr != nil {
		return "", writeErr
	}
	if closeWriterErr != nil {
		return "", closeWriterErr
	}
	return id, closeFileErr
}

func mirrorTreeIdentity(root string) (closuregraph.ID, []any, error) {
	nodes := []any{}
	err := filepath.WalkDir(root, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil || current == root || entry.IsDir() {
			return walkErr
		}
		info, err := entry.Info()
		if err != nil || !info.Mode().IsRegular() || info.Mode()&fs.ModeSymlink != 0 {
			return failure("closure_derivation_drift", "mirror output contains an invalid node")
		}
		payload, err := os.ReadFile(current) // #nosec G304 -- WalkDir supplies a contained mirror output member.
		if err != nil {
			return err
		}
		relative, _ := filepath.Rel(root, current)
		nodes = append(nodes, map[string]any{"executable": info.Mode().Perm()&0o100 != 0, "path": filepath.ToSlash(relative), "sha256": string(digestBytes(payload)), "size": int64(len(payload))})
		return nil
	})
	if err != nil {
		return "", nil, err
	}
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].(map[string]any)["path"].(string) < nodes[j].(map[string]any)["path"].(string)
	})
	id, err := closuregraph.DomainID("source-control-mirror-tree-v1", map[string]any{"nodes": nodes})
	return id, nodes, err
}

func gitObjectID(value string) bool {
	if len(value) != 40 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil && value == strings.ToLower(value)
}
