# TASK-260916-yvxbs1 — review verdict, CR rev5 (tree 3e55fd15, base 86552087): ACCEPTED

Reviewer: claude-opus-5-5 (low). The worktree was confirmed byte-identical to the candidate tree (`git diff 3e55fd15` is empty). Mutants were run on `git archive 3e55fd15` copies in $TMPDIR. Host: darwin/zsh, `set -o pipefail`.

## Spec (curator-spec rc.13 @23435129)
- environments.md:463/548 — `mcp_declaration_path_source_refused` for a path-kind root, overlay or onboarding import.
- environments.md:763-800 — the "`path` source directories" rule: five checks on every env resolve and under the mutation lock for install/update/use/sync/repair/gc. A failure is entry-class `environment_store_untrusted`: no rebuild, and dry-run reports environment_store_untrusted.
- vectors/environments-path-kind-admission.json — includes path-overlay-system-module-admitted.

## Findings by review item
1. **MCP refusal.** `refusePathMCPDeclaration` (pathsource.go) is called from install root (envprofile.go ~L91/127), the overlay (overlays.go:387), the install preflight overlays, and import (via Install). Absence is read through stateread (contextpkg.LoadMCPIfPresent). Diagnostic matches the spec.
2. **System modules.** TestPathOverlayTrustedSystemModuleRemainsAdmitted, plus the vector consumer. The transitive case is left to the existing trunk admission and is not changed.
3. **Boundary call sites:**
   - resolve: managed.go Resolve (before verifyHome) and assembleHome
   - repair: repairUnderLock
   - install: validatePathPackageDirectory plus preflightPathOverlayDeclarations before ensureDefault
   - update: updateLocked, non-default
   - use: useLocked preflight
   - sync: SyncWithPolicy preflight and materializeScope
   - gc: cmd/curator collectUnderLock → PreflightCurrentPathSources
   - status: homeState finding (non-current)

   All six mutating operations and resolve are covered. No re-copy happens on failure.
4. **Windows.** checkMutationPermissions refuses any non-owner ALLOW ACE carrying mutation rights and any unsupported ACE type; this is not weakened. Fixtures use ProtectTree. Ownership is driven through the OwnerLookup seam (ValidateWithOwner; production uses the real lookup).
5. **`profile update default`.** ensureDefault then the refusal. This is equivalent to trunk (trunk also ensured default before the absent refusal); TestUpdateDefaultIsBlocked passes.
6. **Ledger and rebase fidelity.** No conformance-gaps.tsv rows were owned by wgt8vz/yvxbs1 (before 0 / after 0). Case counts 5/14/3 were added. platform-cases renamed TestSymlinkInPathIsSourceInvalid→TestSymlinkInPathIsUntrusted with the same link assertion and the new diagnostic. The changes to envprofile.go, managed.go, status.go and root-artifacts are additive over trunk; the only removals are the ensureDefault reorder and one comment. envregistry.go is untouched. No CHANGELOG, LOGBOOK or stray files.

## Reviewer reruns (real exit codes)
- `go build ./...` =0; `go vet` pathboundary/envprofile/contextpkg =0
- `go test ./internal/pathboundary ./internal/contextpkg` =0
- `go test ./internal/envprofile -run 'Path|Boundary|Overlay|Mcp|MCP|System|Guarded|UpdateDefault'` =0 (185.8s)

## Mutants (mine, -run PathKind|PathRoot|PathOverlay|PathSource)
| Mutant | Result | Killed by |
|---|---|---|
| M1 MCP present forced false | killed rc=1 | TestPathRootAndImportedMCPDeclarationsRefused, TestPathOverlayMCPDeclarationRefused |
| M2 boundary check removed from Resolve | killed rc=1 | TestPathOverlayFailureBlocksResolveAndMakesStatusNonCurrent |
| M3b permissions check disabled (&0o000) | killed rc=1 | pathboundary TestValidateRejectsGroupOrWorldWritableComponents + envprofile |
| M4 isLink never true (unix) | killed rc=1 | TestValidateRejectsInternalSymlink, TestSymlinkInPathIsUntrusted, TestPathOverlayBoundaryRejectsEscapingAndInternalLinks |
| M5 owner comparison disabled | killed rc=1 | TestValidateRejectsInjectedForeignOwner, TestPathSourceMutationsRejectUntrustedOverlayBeforeDefaultWrites/{update,use,sync} |
| **M3 narrowing 0o022→0o002 (group-write admitted)** | **SURVIVED** rc=0 | — |

## Residuals (non-blocking)
- **R1.** Group-writable-only components have no killing row; every fixture uses 0o777. The production code is correct (it checks 0o022). A 0o770/0o775 row in TestValidateRejectsGroupOrWorldWritableComponents would kill M3.
- **R2.** The Windows DACL narrowing (e.g. masking only GENERIC_ALL) was not mutated locally, because Windows is not available on this host. It is bounded by the hosted Windows lane only.
- **R3.** Overlays are validated from the current policy's EffectiveOverlays rather than the lock's recorded overlay members. This is equivalent while policy and lock agree; a policy edit that drops an overlay would leave that lock-named directory unchecked until re-resolve.
- **R4.** The update path does not preflight the default profile's path overlays before ensureDefault; install does.
