// Command regenerate independently rebuilds the interoperability fixtures.
// It deliberately imports no Curator packages so the golden expectations do
// not share implementation code with the tests that consume them.
package main

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const fixedCommit = "0123456789abcdef0123456789abcdef01234567"

var includeRoots = map[string]bool{
	"SKILL.md": true, "agents": true, "references": true, ".skill_triggers": true,
	"assets": true, "templates": true, "examples": true, "data": true,
}

var excludedPatterns = []string{
	".git", ".github", ".gitlab-ci.yml", ".venv", "__pycache__", "*.pyc",
	"node_modules", "tests", "test", "__tests__", "README*", "CHANGELOG*",
	"LICENSE*", "Makefile", "setup.py", "pyproject.toml", "requirements*.txt",
	".DS_Store", ".gitignore",
}

func main() {
	root := flag.String("root", ".", "repository root")
	flag.Parse()
	golden := filepath.Join(*root, "testdata", "golden")
	fixture := filepath.Join(golden, "skill-fixture")
	expected := filepath.Join(golden, "expected")
	must(os.MkdirAll(filepath.Join(expected, "registry"), 0o755))

	snapshotFiles := regularFiles(fixture)
	snapshotHash := contentHash(fixture, snapshotFiles)
	writeText(filepath.Join(expected, "snapshot_sha256.txt"), snapshotHash+"\n")

	contextFiles := contextFiles(fixture)
	writeJSON(filepath.Join(expected, "context_files.json"), contextFiles)
	contextHash := contentHash(fixture, contextFiles)
	writeText(filepath.Join(expected, "context_sha256.txt"), contextHash+"\n")

	markerObject := map[string]any{
		"schema_version": 1, "name": "golden-skill", "source": "golden-skill",
		"ref_kind": "revision", "ref": fixedCommit, "commit": fixedCommit,
		"content_sha256": contextHash, "locale": nil, "agents": []string{"codex_cli"},
		"commands": []string{"golden-tool"}, "dependencies": []string{},
		"skill_schema_version": 5, "runtime_roots": []string{"scripts"},
		"installed_at": "2000-01-01T00:00:00Z", "files": contextFiles,
		"activation": map[string]any{"context": true, "commands": []string{"golden-tool"}},
		"requirers":  []string{"<project>"},
	}
	writeJSON(filepath.Join(expected, "marker.json"), markerObject)

	seed := make([]byte, ed25519.SeedSize)
	for index := range seed {
		seed[index] = byte(index)
	}
	private := ed25519.NewKeyFromSeed(seed)
	public := private.Public().(ed25519.PublicKey)
	pinned := "ed25519:" + base64.StdEncoding.EncodeToString(public)
	writeText(filepath.Join(expected, "registry", "pinned_key.txt"), pinned+"\n")

	auditedBody := map[string]any{
		"name": "golden-skill", "source_identity": "git.example.com/skills/golden-skill",
		"commit": fixedCommit, "content_sha256": snapshotHash, "status": "audited",
		"audit": map[string]any{"auditor": "golden", "note": "заметка"},
	}
	audited := sign(auditedBody, private, public)
	writeJSON(filepath.Join(expected, "registry", "record_audited.json"), audited)

	revokedBody := map[string]any{
		"name": "golden-skill", "source_identity": "git.example.com/skills/golden-skill",
		"commit": fixedCommit, "content_sha256": snapshotHash, "status": "revoked",
		"audit": map[string]any{"reason": "test revocation"},
	}
	writeJSON(filepath.Join(expected, "registry", "record_revoked.json"), sign(revokedBody, private, public))
	forged := cloneObject(audited)
	forged["status"] = "revoked"
	writeJSON(filepath.Join(expected, "registry", "record_forged.json"), forged)

	snapshot := sign(map[string]any{
		"schema_version": 1,
		"merkle_root":    strings.Repeat("a", 64),
		"log_size":       2,
		"head":           strings.Repeat("b", 64),
		"version":        2,
		"created_at":     "2026-07-12T00:00:00Z",
	}, private, public)
	writeJSON(filepath.Join(expected, "registry", "snapshot.json"), snapshot)
}

func regularFiles(root string) []string {
	var files []string
	must(filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.Type().IsRegular() {
			rel, relErr := filepath.Rel(root, path)
			if relErr != nil {
				return relErr
			}
			files = append(files, filepath.ToSlash(rel))
		}
		return nil
	}))
	sort.Strings(files)
	return files
}

func contextFiles(root string) []string {
	var manifestData struct {
		RuntimeRoots []string       `json:"runtime_roots"`
		Commands     map[string]any `json:"commands"`
	}
	payload, err := os.ReadFile(filepath.Join(root, "csk-skill.json"))
	must(err)
	must(json.Unmarshal(payload, &manifestData))

	var files []string
	for _, rel := range regularFiles(root) {
		parts := strings.Split(rel, "/")
		if !includeRoots[parts[0]] && !(parts[0] == "scripts" && len(manifestData.Commands) == 0) {
			continue
		}
		if excluded(parts) || underRoot(rel, manifestData.RuntimeRoots) {
			continue
		}
		files = append(files, rel)
	}
	return files
}

func excluded(parts []string) bool {
	for _, part := range parts {
		for _, pattern := range excludedPatterns {
			if matched, _ := filepath.Match(pattern, part); matched {
				return true
			}
		}
	}
	return false
}

func underRoot(path string, roots []string) bool {
	for _, root := range roots {
		if path == root || strings.HasPrefix(path, strings.TrimRight(root, "/")+"/") {
			return true
		}
	}
	return false
}

func contentHash(root string, files []string) string {
	digest := sha256.New()
	for index, rel := range files {
		if index > 0 {
			_, _ = digest.Write([]byte{0})
		}
		_, _ = digest.Write([]byte(rel))
		_, _ = digest.Write([]byte{0})
		payload, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		must(err)
		_, _ = digest.Write(payload)
	}
	return "sha256:" + hex.EncodeToString(digest.Sum(nil))
}

func sign(body map[string]any, private ed25519.PrivateKey, public ed25519.PublicKey) map[string]any {
	record := cloneObject(body)
	signature := ed25519.Sign(private, canonicalBytes(record))
	keyHash := sha256.Sum256(public)
	record["sig"] = map[string]any{
		"key_id": hex.EncodeToString(keyHash[:])[:16], "algorithm": "ed25519",
		"signature": base64.StdEncoding.EncodeToString(signature),
	}
	return record
}

func canonicalBytes(value any) []byte {
	switch typed := value.(type) {
	case map[string]any:
		keys := make([]string, 0, len(typed))
		for key := range typed {
			if key != "sig" {
				keys = append(keys, key)
			}
		}
		sort.Strings(keys)
		parts := make([]string, 0, len(keys))
		for _, key := range keys {
			keyBytes, _ := json.Marshal(key)
			parts = append(parts, string(keyBytes)+":"+string(canonicalBytes(typed[key])))
		}
		return []byte("{" + strings.Join(parts, ",") + "}")
	case []any:
		parts := make([]string, len(typed))
		for index, item := range typed {
			parts[index] = string(canonicalBytes(item))
		}
		return []byte("[" + strings.Join(parts, ",") + "]")
	default:
		payload, _ := json.Marshal(typed)
		return payload
	}
}

func cloneObject(value map[string]any) map[string]any {
	payload, err := json.Marshal(value)
	must(err)
	var clone map[string]any
	must(json.Unmarshal(payload, &clone))
	return clone
}

func writeJSON(path string, value any) {
	payload, err := json.MarshalIndent(value, "", "  ")
	must(err)
	writeText(path, string(payload)+"\n")
}

func writeText(path, text string) {
	must(os.WriteFile(path, []byte(text), 0o644))
}

func must(err error) {
	if err != nil {
		panic(fmt.Sprintf("regenerate golden fixtures: %v", err))
	}
}
