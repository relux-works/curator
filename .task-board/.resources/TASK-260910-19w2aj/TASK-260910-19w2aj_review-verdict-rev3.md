# TASK-260910-19w2aj revision 3 — CHANGES_REQUESTED

Exact candidate tree: f969280b54514ade5e4aed90428dc8b9d6736398.
Base: 62ea2d2ced3fac7c3c5f7a2ff4e887e0f92f4540.
All 13 changed files independently compared byte-for-byte with candidate blobs: 13/13 equal. No repository source or test modifications. Revision 2 → 3 is exactly the requested Windows skip-reason correction; that correction is valid. This review nevertheless reproduces two production failures in the complete candidate.

## P1: Git runtime bytes still escape frozen authentication

Locations: internal/closure/resolve.go:302-307, 365-366, 668-673.
LoadDraftFrozenNodes authenticates only ContentHashFor, which excludes declared runtime/build roots. openGitFrozen rejects links/special files but does not authenticate regular-file bytes against the locked Git commit. The previous revision-1 finding 3 explicitly required these excluded bytes to remain authenticated. Its repair is incomplete.

Independent CLI reproduction (attached Python fixture and complete log): build the exact candidate CLI; create a local Git fixture with skills/review, exported script command tool, runtime_roots=[scripts], tag v1; use a fake Git transport to map https://fixture.test/kit.git to the local bare repository. Bootstrap into an isolated temporary config/home, add project, then create the lock only through project resolve. Resolve exits 0; baseline install --dry-run exits 0. Change only cache/fixture.test/kit/<locked-commit>/snapshot/skills/review/scripts/tool.sh from echo original to echo TAMPERED, leaving the lock and Git commit unchanged. install --dry-run still exits 0 and plans commands=[tool]. No implicit resolve or live-source change was involved.

Measured adversarial bound: runtime regular-file mutation refusal 0/1; ordinary context tamper cases in the shipped suite pass. Build-file tampering is the same excluded-hash path by inspection, not a separately executed case. No source-code mutant run claimed. Actual tampered-byte publication is NOT claimed: the separate marker failure below prevents it.

Required repair: authenticate complete locked Git snapshot inventory/content, including runtime/build regular files and executable mode, without network/live adoption. Add production CLI negative controls for excluded runtime/build bytes and missing inventory members, plus narrowing-mutant evidence. Do not treat link checks or projected context hashes as full snapshot authentication.

## P1: real install of a successfully resolved Git selection fails marker publication

Locations: internal/closure/resolve.go:309-318; internal/install/install.go:1022-1027; internal/marker/marker.go:275-278 (existing validation).
Frozen nodes populate Resolved.Commit but omit Resolved.Kind/Ref and declaration Source. The unchanged marker builder consumes those empty fields and existing marker validation refuses the result. On the same UNMODIFIED fixture above, project resolve exits 0 and install --dry-run exits 0, but install app exits 1: "error: review: install marker is invalid for schema 2". This is before any cache tampering. Thus successful resolution cannot be consumed by the actual installation path for this valid selected Git script package. The schema number in that diagnostic is the install-marker schema, not the Skillfile schema.

Required repair: carry the accepted source identity through the real materialization/marker path using the approved draft contract; preserve frozen v1 validation. Add non-dry-run CLI install and pinned-consumption coverage. Coordinate with any separately owned draft-marker work if needed; do not fabricate legacy source/ref values merely to pass validation. This is ordinary implementation rework, not a human-only decision or external blocker.

## Independent checks and external evidence

Commands run in zsh, observed exits:
- go test -p 1 ./cmd/curator -run '^TestProjectResolve(LocalCreatesLockThroughCLI|LegacyUntouched|DraftOffRefuses|GitSelectionThroughCLI)$' -count=1 -timeout=90s → exit 0, 2.624s, all four selected tests.
- go test -p 1 ./internal/closure ./internal/install -run 'Test(ResolveDraft|OpenDraftFrozen|Refresh|DraftInstall|LegacyInstallUntouched|DraftSourcesSwitch)' -count=1 -timeout=90s → exit 0; closure 4.513s, install 3.437s.
- go build -o /tmp/TASK-260910-19w2aj-review-curator ./cmd/curator → exit 0.
- git diff --check base candidate → exit 0.
- python3 /tmp/TASK-260910-19w2aj-repro.py → exit 0 (harness completed; it records the unexpected product outcomes rather than claiming acceptance). Setup, resolve, and dry-run commands exit 0; baseline and tampered real install exit 1. Complete log attached separately.

Hosted gate evidence independently read, not rerun: https://github.com/relux-works/curator/actions/runs/35187218160 is success. GitHub commit API maps gate commit 8057410578eedbb449789e02db9276cdf34b2207 to exact candidate tree f969280b54514ade5e4aed90428dc8b9d6736398. Ubuntu/macOS/Windows tests, Ubuntu/macOS race, lint, naming, interop and gate self-tests succeed. Candidate suite and rose-air skipped. No local full suite/lint replay; those are hosted evidence. Passing gates do not cover the two reproduced failures.

Read producer rev3 results, prior reviewer rev1 verdict, task/rework briefs and accepted skillfile-sources contract (Git integrity and locked-consumption requirements). Prior verdict has no rev2 acceptance artifact on this task. No acceptance inferred from the rework brief's statement that four findings were fixed. Existing runtime/build refresh tests are helper-level; shipped CLI install coverage uses dry-run. No launch success asserted.

Applied project-management reviewer lifecycle and negative-evidence skill. Run goal queried: not goal-bound; directives: none. This artifact records important findings as the review logbook entry because campaign rules prohibit LOGBOOK.md edits. Reproduction fixtures and binary are isolated temporary files; no real runtime home, credential export, remote fixture access, installed script execution, commits, or production-source changes.

Verdict: changes_requested. Route TASK-260910-19w2aj to to-dev for repair and another reviewer cycle. Do not accept revision 3.
