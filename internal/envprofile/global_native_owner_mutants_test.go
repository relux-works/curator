package envprofile

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/godriver"
)

// Each overlay removes one ownership condition for a bounded witness. A kill
// requires the named assertion and exit 1; setup, build failures, crashes and
// missing tests do not count. Children execute one package at a time.
func TestGlobalNativeOwnershipNarrowingControls(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	for _, plant := range []struct{ name, file, before, after, pkg, run, failure, message string }{
		{"kind-only-parent-compensation", "internal/envprofile/global_operation.go", "current.mode != published.mode", "false", "./cmd/curator", "^TestGlobalNativeOwnerCompensationParentMode$/^operator-chmod$", "TestGlobalNativeOwnerCompensationParentMode/operator-chmod", "operator parent mode overwritten"},
		{"ignore-created-parent-mode", "internal/envprofile/global_operation.go", "if current == nil && undo.absentParents[parent] {", "if undo.absentParents[parent] {", "./internal/envprofile", "^TestGlobalProfileCompensationCreatedEmptyParentMode$/^changed=true$", "TestGlobalProfileCompensationCreatedEmptyParentMode/changed=true", "changed empty parent admitted"},
		{"ignore-parent-owner", "internal/envprofile/global_operation.go", "current.owner != published.owner", "false", "./internal/envprofile", "^TestGlobalProfileCompensationMetadata$/^parent-owner$/^changed=true$", "TestGlobalProfileCompensationMetadata/parent-owner/changed=true", "foreign metadata admitted"},
		{"omit-image-owner", "internal/envprofile/global_operation.go", "image := &globalImage{mode: info.Mode(), owner: owner}", "_ = owner\n\timage := &globalImage{mode: info.Mode()}", "./internal/envprofile", "^TestGlobalProfileCompensationMetadata$/^owner-capture$/^changed=false$", "TestGlobalProfileCompensationMetadata/owner-capture/changed=false", "image owner not captured"},
		{"drop-special-compensation-bits", "internal/envprofile/global_operation.go", "return mode & (os.ModePerm | os.ModeSetuid | os.ModeSetgid | os.ModeSticky)", "return mode.Perm()", "./internal/envprofile", "^TestGlobalProfileCompensationLateParentMode$/^changed=false$", "TestGlobalProfileCompensationLateParentMode/changed=false", "parent special mode lost"},
		{"omit-parent-metadata-checkpoint", "internal/envprofile/global_operation.go", "for parent := range op.undo.parentPostimages {\n\t\tfor _, path := range paths {", "for parent := range map[string]*globalImage{} {\n\t\tfor _, path := range paths {", "./internal/envprofile", "^TestGlobalProfileCompensationParentModes$/^created=false$/^changed=false$", "TestGlobalProfileCompensationParentModes/created=false/changed=false", "compensation parent postimage changed"},

		{"omit-directory-only-parent-checkpoint", "internal/envprofile/global_operation.go", "for parent := range op.undo.parentPostimages {\n\t\tif parent == root", "for parent := range map[string]*globalImage{} {\n\t\tif parent == root", "./internal/envprofile", "^TestGlobalProfileCompensationDirectoryOnlyParents$/^created=true$/^changed=false$", "TestGlobalProfileCompensationDirectoryOnlyParents/created=true/changed=false", "directory-only rollback failed"},
		{"revert-pending-lock-before-nested-native", "internal/envprofile/global_operation.go", "return undo.managerHome != \"\" && (path == CurrentFile(undo.managerHome) || strings.HasPrefix(path, ProfilesDir(undo.managerHome)+string(filepath.Separator)))", "return false", "./internal/envprofile", "^TestGlobalProfileCompensationNestedNativeHome$", "TestGlobalProfileCompensationNestedNativeHome/writable=false", "nested native compensation lost pending lock"},
		{"retry-failed-profile-compensation", "internal/install/global.go", "if profileErr != nil {", "if profileErr != nil && false {", "./internal/install", "^TestGlobalProfileCompensationStopsRestart$/^failed$", "TestGlobalProfileCompensationStopsRestart/failed", "compensation failure was retried or hidden"},
		{"overwrite-compensation-postimage", "internal/envprofile/global_operation.go", "if !reflect.DeepEqual(current, undo.postimages[path]) {", "if !reflect.DeepEqual(current, undo.postimages[path]) && false {", "./internal/envprofile", "^TestGlobalProfileCompensationPostimages$", "TestGlobalProfileCompensationPostimages/link/replacement=true", "foreign postimage admitted"},
		{"discard-pending-lock-after-compensation-error", "internal/envprofile/global_operation.go", "if err := restoreGlobalImage(path, undo.entries[path]); err != nil {\n\t\t\t\treturn err", "if err := restoreGlobalImage(path, undo.entries[path]); err != nil {\n\t\t\t\tfor _, record := range paths { if recoveryRecord(record) { _ = restoreGlobalImage(record, undo.entries[record]) } }; return err", "./internal/envprofile", "^TestGlobalProfileCompensationNestedNativeHome$/^writable=false$", "TestGlobalProfileCompensationNestedNativeHome/writable=false", "nested native compensation lost pending lock"},
		{"follow-compensation-replacement-parent", "internal/envprofile/global_operation.go", "for parent := range undo.parentPostimages {", "for parent := range map[string]*globalImage{} {", "./internal/envprofile", "^TestGlobalProfileCompensationParentReplacement$", "TestGlobalProfileCompensationParentReplacement", "compensation changed foreign parent contents"},
		{"omit-native-publication-postimage", "internal/envprofile/switch.go", "if err := op.recordHomePostimages(native); err != nil {", "if err := error(nil); err != nil {", "./cmd/curator", "^TestGlobalNativeOwnerCompensationRestart$/^none$", "TestGlobalNativeOwnerCompensationRestart/none", "clean restart failed"},
		{"omit-preview-carrier-validation", "internal/manifest/manifest.go", "func WithDecl(payload []byte, filePath, name, refKind, refValue, git, source string) ([]byte, error) {\n\tif err := protocoljson.Validate(payload); err != nil {", "func WithDecl(payload []byte, filePath, name, refKind, refValue, git, source string) ([]byte, error) {\n\tif err := protocoljson.Validate(payload); false && err != nil {", "./internal/manifest", "^TestWithDeclRejectsInvalidInput$", "TestWithDeclRejectsInvalidInput", "invalid preview admitted"},
		{"lock-before-private-build", "internal/install/global.go", "\treturn runWithRestarts(\"global\", commit, func() (Result, *restartError) {", "\tif opts.WithGlobalProfile != nil { lock, err := commit.Locks.AcquireHome(ctx); if err != nil { return failedResult(\"global\", GlobalRoot(cfg.Home()), err) }; defer lock.Close(); commit.Locks = globalHeldLocks{LockBroker: commit.Locks, home: lock} }\n\treturn runWithRestarts(\"global\", commit, func() (Result, *restartError) {", "./internal/install", "^TestGlobalProfilePrivateStaging$/^success$", "TestGlobalProfilePrivateStaging/success", "private staging held the home lock"},
		{"lock-before-real-private-build", "internal/install/global.go", "\treturn runWithRestarts(\"global\", commit, func() (Result, *restartError) {", "\tif opts.WithGlobalProfile != nil { lock, err := commit.Locks.AcquireHome(ctx); if err != nil { return failedResult(\"global\", GlobalRoot(cfg.Home()), err) }; defer lock.Close(); commit.Locks = globalHeldLocks{LockBroker: commit.Locks, home: lock} }\n\treturn runWithRestarts(\"global\", commit, func() (Result, *restartError) {", "./cmd/curator", "^TestGlobalNativeOwnerPausedBuild$", "TestGlobalNativeOwnerPausedBuild", "profile sync blocked by private build"},
		{"omit-build-permit-barrier", "cmd/curator/global_native_build_unix_test.go", "if pauseOnPermit && frame.Kind == \"permit\" {", "if false && pauseOnPermit && frame.Kind == \"permit\" {", "./cmd/curator", "^TestNativeBuildPermitBarrier$", "TestNativeBuildPermitBarrier", "build permit pause count"},
		{"ignore-regular-file-snapshot-size", "cmd/curator/global_native_owner_test.go", "if entry.Mode.IsDir() {\n\t\t\tentry.Size = 0", "if entry.Mode.IsDir() || entry.Mode.IsRegular() {\n\t\t\tentry.Size = 0", "./cmd/curator", "^TestNativeOwnerSnapshotContract$", "TestNativeOwnerSnapshotContract", "regular file size lost"},
		{"omit-profile-generation-recheck", "internal/install/global.go", "if restart := observed.recheck(plan, deps.Cache); restart != nil {", "if restart := observed.recheck(plan, deps.Cache); false && restart != nil {", "./internal/install", "^TestGlobalProfileStagingRevalidatesDeclaration$", "TestGlobalProfileStagingRevalidatesDeclaration", "stale declaration reached profile publication before restart"},
		{"omit-prospective-addition", "internal/install/global.go", "if addition := opts.GlobalAddition; addition != nil {", "if addition := opts.GlobalAddition; addition != nil && addition.Name == \"never-staged\" {", "./internal/install", "^TestGlobalProfileStagingAddition$/^exact$", "TestGlobalProfileStagingAddition/exact", "prospective addition was not staged"},
		{"admit-divergent-addition", "internal/install/global.go", "err != nil || !bytes.Equal(document.payload, globalManifestPayload)", "err != nil || (!bytes.Equal(document.payload, globalManifestPayload) && len(document.payload) == 0)", "./internal/install", "^TestGlobalProfileStagingAddition$/^diverged$", "TestGlobalProfileStagingAddition/diverged", "divergent addition reached publication"},
		{"bypass-global-removal-lock", "cmd/curator/main.go", "if _, err := operation.AcquireProjects(ctx, install.GlobalRoot(cfg.Home())); err != nil {", "if _, err := func() ([]managerlock.ProjectIdentity, error) { _ = ctx; return nil, nil }(); err != nil {", "./cmd/curator", "^TestGlobalNativeOwnerConcurrentRemove$/^direct-only$", "TestGlobalNativeOwnerConcurrentRemove/direct-only", "remove completed during paused install"},
		{"release-removal-lock-before-read", "cmd/curator/main.go", "if err := manifest.RemoveDecl(install.GlobalRoot(cfg.Home()), args[1]); err != nil {", "_ = operation.Close()\n\t\tif err := manifest.RemoveDecl(install.GlobalRoot(cfg.Home()), args[1]); err != nil {", "./cmd/curator", "^TestGlobalNativeOwnerRemoveBeforeInstall$", "TestGlobalNativeOwnerRemoveBeforeInstall", "paused remove did not hold the global operation lock"},
		{"omit-direct-declarations", "internal/envprofile/global.go", "if global != nil {", "if global != nil && len(global.Skills) == 0 {", "./cmd/curator", "^TestGlobalNativeOwnerNewDeclarationAfterProfile$", "TestGlobalNativeOwnerNewDeclarationAfterProfile", "profile native owner lost"},
		{"retain-old-direct-pin", "internal/envprofile/global.go", "\tif addition != nil {\n\t\tidentity", "\tfor _, member := range oldLock.Members { if member.Kind == contextlock.KindSkill { input.Direct = append(input.Direct, contextresolve.Requirement{Kind: contextlock.KindSkill, Name: member.Name, Source: member.Source, Revision: member.Commit, Directory: member.Directory}) } }\n\tif addition != nil {\n\t\tidentity", "./cmd/curator", "^TestGlobalNativeOwnerNewDeclarationAfterProfile$", "TestGlobalNativeOwnerNewDeclarationAfterProfile", "want 0"},
		{"resurrect-removed-declaration", "internal/envprofile/global.go", "\tif addition != nil {\n\t\tidentity", "\tfor _, member := range oldLock.Members { if member.Kind == contextlock.KindSkill { input.Direct = append(input.Direct, contextresolve.Requirement{Kind: contextlock.KindSkill, Name: member.Name, Source: member.Source, Revision: member.Commit, Directory: member.Directory}) } }\n\tif addition != nil {\n\t\tidentity", "./cmd/curator", "^TestGlobalNativeOwnerRemoval$/^direct-only$", "TestGlobalNativeOwnerRemoval/direct-only", "native skill appeared before global install"},
		{"omit-profile-compensation", "internal/envprofile/global_operation.go", "if resultErr == nil {", "if true {", "./cmd/curator", "^TestGlobalNativeOwnerCanonicalRefusal$", "TestGlobalNativeOwnerCanonicalRefusal", "canonical refusal changed profile/native state"},
		{"omit-parent-mode-compensation", "internal/envprofile/global_operation.go", "for parent, mode := range undo.parentModes {", "for parent, mode := range map[string]os.FileMode{} {", "./cmd/curator", "^TestGlobalNativeOwnerCanonicalRefusal$", "TestGlobalNativeOwnerCanonicalRefusal", "canonical refusal changed profile/native state"},
		{"follow-compensation-parent", "internal/envprofile/global_operation.go", "if op.undo.managerHome != \"\" {", "if false {", "./internal/envprofile", "^TestGlobalProfileCompensationParentLinkRefuses$", "TestGlobalProfileCompensationParentLinkRefuses", "snapshot followed a manager parent link"},
		{"rollback-unmanaged-extended-lock", "internal/envprofile/global_operation.go", "if errors.As(err, &conflict) && info.Lock != nil {", "if errors.As(err, &conflict) && info.Lock == nil {", "./cmd/curator", "^TestGlobalAddConflictRetainsLockAndProfileSyncTakeoverRecovers$", "TestGlobalAddConflictRetainsLockAndProfileSyncTakeoverRecovers", "default lock has no new-skill member"},
		{"omit-pending-install-record", "internal/envprofile/global_operation.go", "retainedRecords = map[string][]byte{lockPath(home, info.Name): retainedLock, sourcePath(home, info.Name): source}", "_ = source; retainedRecords = map[string][]byte{lockPath(home, info.Name): retainedLock}", "./cmd/curator", "^TestGlobalNativeOwnerFreshAddConflict$", "TestGlobalNativeOwnerFreshAddConflict", "profile native owner lost"},
		{"second-publisher", "internal/adapters/stage.go", "if preserved[name] {", "if preserved[name] && name != \"shared-skill\" {", "./cmd/curator", "^TestGlobalNativeOwnerCleanBootstrapInstall$", "TestGlobalNativeOwnerCleanBootstrapInstall", "want 0"},
		{"double-claim", "internal/adapters/stage.go", "if expected[name] && !preserved[name] {", "if expected[name] {", "./internal/adapters", "^TestGlobalProfileOwnerPreservesEntries$/^selected$", "TestGlobalProfileOwnerPreservesEntries/selected", "profile entry claimed by legacy ledger"},
		{"retired-owner-removal", "internal/adapters/stage.go", "if !expected[name] && !preserved[name] {", "if !expected[name] {", "./internal/adapters", "^TestGlobalProfileOwnerPreservesEntries$/^retired$", "TestGlobalProfileOwnerPreservesEntries/retired", "profile entry was replaced or removed"},
		{"ownership-read-error", "internal/adapters/stage.go", "if err != nil {\n\t\t\t\treturn err\n\t\t\t}\n\t\t\tpreserved[name]", "if err != nil && err.Error() != \"ownership state unreadable\" {\n\t\t\t\treturn err\n\t\t\t}\n\t\t\tpreserved[name]", "./internal/adapters", "^TestGlobalProfileOwnerPreservesEntries$/^selected$", "TestGlobalProfileOwnerPreservesEntries/selected", "ownership error was ignored"},
		{"source-disagreement", "internal/envprofile/native_skill_owner.go", "member.Source != requested.Source", "(member.Source != requested.Source && name != \"myskill\")", "./internal/envprofile", "^TestNativeSkillOwnerProof$/^source$", "TestNativeSkillOwnerProof/source", "invalid native ownership admitted"},
		{"pin-disagreement", "internal/envprofile/native_skill_owner.go", "member.Commit != requested.Commit", "(member.Commit != requested.Commit && name != \"myskill\")", "./internal/envprofile", "^TestNativeSkillOwnerProof$/^commit$", "TestNativeSkillOwnerProof/commit", "invalid native ownership admitted"},
		{"directory-disagreement", "internal/envprofile/native_skill_owner.go", "member.Directory != requested.Directory", "(member.Directory != requested.Directory && name != \"myskill\")", "./internal/envprofile", "^TestNativeSkillOwnerProof$/^directory$", "TestNativeSkillOwnerProof/directory", "invalid native ownership admitted"},
		{"marker-lock", "internal/envprofile/native_skill_owner.go", "if marker.Profile.LockSHA256 != strings.TrimPrefix(hash, \"sha256:\") {", "if marker.Profile.LockSHA256 != strings.TrimPrefix(hash, \"sha256:\") && name != \"myskill\" {", "./internal/envprofile", "^TestNativeSkillOwnerProof$/^lock$", "TestNativeSkillOwnerProof/lock", "invalid native ownership admitted"},
		{"marker-content", "internal/envprofile/native_skill_owner.go", "digest != surface.ContentSHA256", "(digest != surface.ContentSHA256 && name != \"myskill\")", "./internal/envprofile", "^TestNativeSkillOwnerProof$/^surface$", "TestNativeSkillOwnerProof/surface", "invalid native ownership admitted"},
		{"marker-entry", "internal/envprofile/native_skill_owner.go", "if !recorded {", "if !recorded && name != \"myskill\" {", "./internal/envprofile", "^TestNativeSkillOwnerProof$/^unrecorded$", "TestNativeSkillOwnerProof/unrecorded", "invalid native ownership admitted"},
		{"live-link", "internal/envprofile/native_skill_owner.go", "if link == selected.root {", "if link == selected.root || name == \"myskill\" {", "./internal/envprofile", "^TestNativeSkillOwnerProof$/^live-link$", "TestNativeSkillOwnerProof/live-link", "invalid native ownership admitted"},
		{"copy-record", "internal/envprofile/native_skill_owner.go", "if copied {", "if copied || name == \"myskill\" {", "./internal/envprofile", "^TestNativeSkillOwnerCopyProof$/^fallback$", "TestNativeSkillOwnerCopyProof/fallback", "unrecorded copied entry admitted"},
		{"copy-content", "internal/envprofile/native_skill_owner.go", "if hash == selected.hash {", "if hash == selected.hash || name == \"myskill\" {", "./internal/envprofile", "^TestNativeSkillOwnerCopyProof$/^fallback$", "TestNativeSkillOwnerCopyProof/fallback", "drifted copied entry admitted"},
		{"ignore-ledgers", "internal/adapters/adapters.go", "func unmanagedConflict(target string, inLedger bool, source string) (bool, error) {", "func unmanagedConflict(target string, inLedger bool, source string) (bool, error) {\nif filepath.Base(target) == \"shared-skill\" { return false, nil }", "./internal/adapters", "^TestGlobalProfileOwnerForeignRefuses$", "TestGlobalProfileOwnerForeignRefuses", "foreign adapter entry admitted"},
	} {
		t.Run(plant.name, func(t *testing.T) {
			if plant.name == "discard-pending-lock-after-compensation-error" || plant.name == "revert-pending-lock-before-nested-native" {
				if os.Geteuid() == 0 {
					t.Skip("permission control requires an unprivileged process")
				}
			}
			if plant.name == "omit-build-permit-barrier" || plant.name == "lock-before-real-private-build" || plant.name == "kind-only-parent-compensation" || plant.name == "ignore-created-parent-mode" || plant.name == "drop-special-compensation-bits" || plant.name == "omit-parent-metadata-checkpoint" || plant.name == "omit-directory-only-parent-checkpoint" || plant.name == "bypass-global-removal-lock" || plant.name == "release-removal-lock-before-read" || plant.name == "discard-pending-lock-after-compensation-error" || plant.name == "omit-native-publication-postimage" || plant.name == "revert-pending-lock-before-nested-native" {
				switch runtime.GOOS {
				case "aix", "android", "darwin", "dragonfly", "freebsd", "hurd", "illumos", "ios", "linux", "netbsd", "openbsd", "solaris":
				default:
					t.Skip("concurrency witness requires the Unix Git/FIFO harness")
				}
			}
			if plant.name == "lock-before-real-private-build" && godriver.InventoryPlatform(runtime.GOOS) == "" {
				t.Skipf("rc5-native-control-inventory-v1 defines no record for host %s", runtime.GOOS)
			}
			original := filepath.Join(root, filepath.FromSlash(plant.file))
			payload, err := os.ReadFile(original)
			if err != nil {
				t.Fatal(err)
			}
			source := strings.ReplaceAll(string(payload), "\r\n", "\n")
			if count := strings.Count(source, plant.before); count != 1 {
				t.Fatalf("plant anchor count = %d, want 1", count)
			}
			dir := t.TempDir()
			replacement := filepath.Join(dir, "replacement.go")
			if err := os.WriteFile(replacement, []byte(strings.Replace(source, plant.before, plant.after, 1)), 0600); err != nil {
				t.Fatal(err)
			}
			overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{original: replacement}})
			if err != nil {
				t.Fatal(err)
			}
			overlayPath := filepath.Join(dir, "overlay.json")
			if err := os.WriteFile(overlayPath, overlay, 0600); err != nil {
				t.Fatal(err)
			}
			command := exec.Command("env", "-u", "GIT_DIR", "-u", "GIT_WORK_TREE", "go", "test", "-p=2", "-overlay", overlayPath, "-count=1", "-timeout=2m", "-v", plant.pkg, "-run", plant.run)
			command.Dir = root
			if plant.pkg == "./cmd/curator" {
				command.Env = append(os.Environ(), "CURATOR_TEST_CLI_HELPER=1")
			}
			output, err := command.CombinedOutput()
			exit, ok := err.(*exec.ExitError)
			if !ok || exit.ExitCode() != 1 || !strings.Contains(string(output), "--- FAIL: "+plant.failure+" ") || !strings.Contains(string(output), plant.message) || strings.Contains(string(output), "panic:") {
				t.Fatalf("plant survived or lacked intended assertion: %v\n%s", err, output)
			}
			t.Logf("expected-red exit 1; killed by %s: %s", plant.failure, plant.message)
		})
	}
}
