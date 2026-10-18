//go:build snapshotcontrols && (darwin || linux)

package envsnapshot

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

// Each isolated narrowing plant must produce a named behavioral failure with
// exit 1. Compilation/setup errors, skips and missing results are not kills.
// These are library-admission controls, not materialize CLI qualification.
func TestDestinationNarrowingControls(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	corpusPath, err := filepath.Abs(corpus)
	if err != nil {
		t.Fatal(err)
	}
	type plant struct{ name, file, old, new, test string }
	plants := []plant{
		{"cleanup-foreign-stage-link", "publication_unix.go", "return failure\n}", "if len(objects) != 0 { _ = unix.Unlinkat(int(objects[0].parent.Fd()), \".curator-snapshot-foreign\", 0) }; return failure\n}", "TestPublicationStagingOnlyCleanup/io-failure"},
		{"cleanup-omits-one-file", "publication_unix.go", "obj := objects[i]", "obj := objects[i]; if obj.name == \"AGENTS.md\" { continue }", "TestPublicationStagingOnlyCleanup/before-rename"},
		{"post-verify-one-byte-mismatch", "publication_unix.go", "v, err := verifyDirectory(actual, Digest(p.Manifest))", "v, err := verifyDirectory(actual, Digest(p.Manifest)); if e, ok := err.(*Error); ok && e.Reason == \"file-identity\" { return &Verified{p.Manifest, p.Launch}, nil }", "TestPublicationPostVerification/bytes"},
		{"M2-linked-ancestor", "fs_unix.go", "if directory {\n\t\tflags |= unix.O_DIRECTORY\n\t}", "if directory { flags |= unix.O_DIRECTORY }; if directory && name == \"target\" { flags &^= unix.O_NOFOLLOW }", "TestDestinationBoundary/linked-ancestor"},
		{"M6-saved-manifest-bytes", "destination_unix.go", "v, err := verifyDirectory(verifyHandle, \"\")", "v, err := verifyDirectory(verifyHandle, \"\"); if e, ok := err.(*Error); ok && e.Reason == \"file-identity\" { raw, readErr := readMetadata(target, \"manifest.json\"); v = &Verified{manifest: raw}; err = readErr }", "TestDestinationPriorIdentity/bytes"},
		{"owner-root-parent", "destination_unix.go", "int64(stat.Uid) != int64(os.Geteuid()) && (caller || stat.Uid != 0)", "int64(stat.Uid) != int64(os.Geteuid()) && stat.Uid != 0", "TestDestinationOwnershipAndReadFailure"},
		{"mode-group-writable-ancestor", "destination_unix.go", "info.Mode().Perm()&0022 != 0", "(info.Mode().Perm()&0022 != 0 && info.Mode().Perm() != 0770)", "TestDestinationBoundary/shared-ancestor"},
		{"protected-equal-identity", "destination_unix.go", "if e != nil || os.SameFile(info, other) {", "if e != nil || (os.SameFile(info, other) && ancestor != lineage[0]) {", "TestDestinationProtectedIdentity"},
		{"ancestor-directory-replacement", "destination_unix.go", "err != nil || closeErr != nil || !os.SameFile(retained.info, info)", "err != nil || closeErr != nil || (!os.SameFile(retained.info, info) && retained.name != \"parent\")", "TestDestinationReplacementRaces/ancestor-directory"},
		{"target-identical-replacement", "destination_unix.go", "aerr != nil || berr != nil || !os.SameFile(a, b)", "aerr != nil || berr != nil || (!os.SameFile(a, b) && b.Mode() != os.ModeDir|0700)", "TestDestinationReplacementRaces/target-identical"},
		{"empty-target", "destination_unix.go", "return \"\", nil, invalidDestination(\"empty-target\")", "return DestinationAbsent, nil, nil", "TestDestinationReadOnlyAdmission"},
		{"private-parent", "destination_unix.go", "(private && info.Mode().Perm() != 0700)", "(private && info.Mode().Perm() != 0700 && info.Mode().Perm() != 0755)", "TestDestinationPrivateParent"},
		{"unrelated-one-entry", "destination_unix.go", "if !hasManifest {", "if !hasManifest && len(entries) != 1 {", "TestDestinationBoundary/unrelated"},
	}
	// The unrelated-target plant must admit just the one-entry case without
	// accidentally dying in SNAP-ID. All other nonempty targets still verify.
	plants[len(plants)-1].new = "if !hasManifest && len(entries) == 1 { return DestinationAbsent, nil, nil }; if !hasManifest {"
	if runtime.GOOS == "darwin" {
		plants = append(plants, plant{"publication-replaces-empty", "publication_darwin.go", "unix.RenameatxNp(parent, stage, parent, target, unix.RENAME_EXCL)", "unix.Renameat(parent, stage, parent, target)", "TestPublicationNoReplaceRace/empty"})
		plants = append(plants, plant{"acl-one-permit-kind", "destination_acl_darwin.go", "flags&0xf != 2", "flags&0xf != 2 && !(flags&0xf == 1 && binary.NativeEndian.Uint32(sec[44+i*24+20:44+i*24+24]) == 0x24)", "TestDestinationKernelACL"})
		plants = append(plants, plant{"prior-acl-one-member", "destination_unix.go", "if err == nil {\n\t\t\terr = destinationACL(f)", "if err == nil && entry.Name() != \"AGENTS.md\" {\n\t\t\terr = destinationACL(f)", "TestDestinationPriorKernelACL"})
	} else {
		plants = append(plants, plant{"publication-replaces-empty", "publication_linux.go", "unix.Renameat2(parent, stage, parent, target, unix.RENAME_NOREPLACE)", "unix.Renameat(parent, stage, parent, target)", "TestPublicationNoReplaceRace/empty"})
		plants = append(plants, plant{"acl-default-only", "destination_acl_linux.go", "!errors.Is(err, unix.ENODATA)", "!errors.Is(err, unix.ENODATA) && name != \"system.posix_acl_default\"", "TestDestinationKernelACL/system.posix_acl_default"})
		plants = append(plants, plant{"prior-acl-one-member", "destination_unix.go", "if err == nil {\n\t\t\terr = destinationACL(f)", "if err == nil && entry.Name() != \"home\" {\n\t\t\terr = destinationACL(f)", "TestDestinationPriorKernelACL"})
	}
	for _, p := range plants {
		t.Run(p.name, func(t *testing.T) {
			if p.name == "owner-root-parent" && os.Geteuid() == 0 {
				t.Fatal("owner plant unqualified under effective root; use an unprivileged test process")
			}
			scratch, err := os.MkdirTemp(filepath.Join(root, ".temp"), "destination-control-")
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := os.RemoveAll(scratch); err != nil {
					t.Error(err)
				}
			}()
			if err := filepath.WalkDir(".", func(path string, entry fs.DirEntry, walkErr error) error {
				if walkErr != nil {
					return walkErr
				}
				if entry.IsDir() {
					if path != "." && strings.HasPrefix(entry.Name(), ".") {
						return filepath.SkipDir
					}
					return os.MkdirAll(filepath.Join(scratch, path), 0700)
				}
				if path != ".gitattributes" && !strings.HasSuffix(path, ".go") && !strings.HasPrefix(path, "schemas/") {
					return nil
				}
				raw, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				if path == "codec_test.go" {
					raw = bytes.ReplaceAll(raw, []byte(strconv.Quote(corpus)), []byte(strconv.Quote(corpusPath)))
				}
				if path == p.file {
					if bytes.Count(raw, []byte(p.old)) != 1 {
						t.Fatalf("control target count %d", bytes.Count(raw, []byte(p.old)))
					}
					raw = bytes.Replace(raw, []byte(p.old), []byte(p.new), 1)
				}
				return os.WriteFile(filepath.Join(scratch, path), raw, 0600)
			}); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command("go", "test", "-count=1", "-json", "-run", "^"+p.test+"$", scratch)
			cmd.Dir = root
			output, runErr := cmd.CombinedOutput()
			exitErr, expectedExit := runErr.(*exec.ExitError)
			killed := false
			for _, line := range bytes.Split(output, []byte{10}) {
				var event struct{ Action, Test string }
				if json.Unmarshal(line, &event) == nil && event.Action == "fail" && (event.Test == p.test || strings.HasPrefix(event.Test, p.test+"/")) {
					killed = true
				}
			}
			if !expectedExit || exitErr.ExitCode() != 1 || !killed {
				sanitized := strings.ReplaceAll(strings.ReplaceAll(string(output), scratch, "<control>"), root, "<repository>")
				t.Fatalf("survivor or invalid plant: %v\n%s", runErr, sanitized)
			}
			t.Logf("%s: killed by %s; child exit 1 (expected failing command)", p.name, p.test)
		})
	}
}
