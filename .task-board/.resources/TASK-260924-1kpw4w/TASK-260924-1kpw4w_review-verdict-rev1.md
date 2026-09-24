# TASK-260924-1kpw4w review verdict — CR rev1: ACCEPTED

Candidate tree e524eb59 (the worktree's write-tree matches it), base 948ae7c9. Reviewed read-only; I restored every mutant afterwards and the tree has no stray files.

## Checks against the review note
1. Grammar: `identifiers.ValidDirectory` is one validator shared by the Skillfile selector (manifest/sources.go parseSelector) and schema-9 `dependencies.skills[].directory` (skillspec/parse.go). The key is rejected below schema 9. An absent directory normalizes to ".".
2. Identity, closure, lock and audit: Node.Directory is threaded through resolveNode, acquireGitPinned/acquireLocal, memberForNode (network-git package directory = member directory) and LoadDraftFrozenNodes. `unify` refuses the same name when the folders differ. A diamond on one folder unifies. Different folders of one repository share one checkout, keyed by canonical identity. SelectSkillPackage requires SKILL.md with a matching name. It refuses symlinked components with `source_selection_invalid ... escapes source`, and a git snapshot link that intersects the selection gets the same diagnostic. A configured-git subdirectory is refused.
3. Absent directory: TestLegacyDependencyWithoutDirectoryLockBytes pins sha256:685572494b16…d66c, and the test passes.
4. Vectors: I compared the blob OIDs from the accepted 2am4qa rev2 patch with `git hash-object`. The 23 files (manifest-dependency-directories.json plus the agent/csk-skill-v9 schema-cases) and install-marker-v5.schema.json (5c6a573) are byte-identical. The README and index.json are not copied; they have no counterpart in this repo layout.
5. Real CLI: TestDraftSourcesCLIDirectoryManifestDependencyInstallRefreshAndAudit passes (install, refresh, reinstall; lock member/package directory, marker Package.Directory and source-audit directories).
6. There is no CHANGELOG edit.

## Reruns (mine)
go build ./... and go vet on the touched packages are clean. go test for closure, skillspec, manifest, identifiers, marker, sourcelock and audit: all ok. The crossconformance -run CLIDirectory test: ok.

## Mutants I re-applied
- M-a: disabled the directory-inequality check in closure.unify. The SameNameDifferentFolderConflicts test fails, so the mutant is killed.
- M-b: narrowed ValidDirectory glob set "*?[]" to "*". internal/manifest tests fail, so it is killed there. The skillspec v9 vectors alone do not kill it (only the `*` glob vector exists). This is a stated bound; the shared validator covers it.

## Residuals (non-blocking)
- R1: In the legacy (non-draft) Skillfile lane, a schema-9 skill with a subfolder dependency resolves through closure.Build. Neither the legacy marker (marker.Marker has no directory field outside v5) nor audit.Subject records the directory. The draft lane records it everywhere.
- R2: The dependency grammar's glob class is only pinned by manifest selector tests. Adding `?`/`[` vectors in a spec follow-up would pin it on the dependency path too.
