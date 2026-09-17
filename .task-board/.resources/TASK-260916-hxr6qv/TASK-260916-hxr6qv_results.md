# TASK-260916-hxr6qv results — revision-2 bounded resolution and provenance

Role: developer. Worktree: `.temp/STORY-260916-v58b5y/worktree`, branch
`task-board/story/STORY-260916-v58b5y`, base checkpoint `c89473c`
(TASK-260916-27cv45 schema-2 loader). No commits made; all changes
uncommitted for orchestrator integration.

## What changed

- `internal/buildrepo/transport.go`
  - `TransportAttempt` carries revision-2 endpoint properties: `MirrorOf`,
    `Alias`, `ResolvedHost`, `ResolvedPort`, `HasExplicitPort`.
  - `parseTransportPlan` is revision-2 aware: port-bearing URLs parse under
    the section 5 endpoint grammar (URI forms only, decimal 1-65535, no
    leading zeros; scp-like colon segment stays a path); the section 6
    resolved-host predicate is revalidated
    (`repository_mirror_undeclared` for unattested mirrors,
    `repository_policy_invalid` for every misuse); then the section 7
    strict-lane refusal fires `build_repository_identity_invalid` for any
    explicit port or alias field — never stripped, never ignored. Declared
    mirrors without ports or aliases are ordinary lane URLs. Revision-1
    attempts take the original checks byte-identically (same messages,
    same order). The stderr classification table is untouched.
  - New `CodeRepositoryMirrorUndeclared`. New closed-grammar helpers
    (`parseTransportEndpointURL`, `splitTransportEndpointPort`,
    `parseTransportPort`).
  - `AttemptRecord` gains machine-private section 7 provenance: canonical
    `Identity`, `ResolvedHost`/`ResolvedPort`, `Alias`, `MirrorOf`. No
    stderr, no secrets. Error and exhaustion text unchanged for
    revision-1 plans.
- `internal/install/drafttransport.go`: `draftTransportPlan` converts
  resolutions field-for-field including revision-2 provenance;
  `draftPlanNeedsSSH` is port-aware (`ssh://` prefix fallback);
  `acquireDraftNetwork` passes `deps.DraftTransportTrace` to the executor.
- `internal/install/external.go`: new `ExternalDeps.DraftTransportTrace`
  hook (nil by default preserves previous behavior).
- Docs: `docs/draft-transport-resolution.md` gains the revision-2
  executor, lane-refusal, provenance, and caller sections;
  `docs/draft-source-policy.md` now points at the implemented leaf.

## Tests (all through the production entry unless noted)

- `internal/buildrepo/transport_v2_test.go` (new, executor-level, no I/O):
  mirror admission, 16-row refusal table (section 6 classes, lane
  refusal, unchanged 2-attempt bound), refusal narrowing (port removed
  admits; provenance removed mistranslates, never strips), revision-1
  diagnostics byte-identical, section 5 URL grammar table, exhaustion
  closed-vocabulary with revision-2 records.
- `internal/install/drafttransport_v2_test.go` (new, through
  `acquireDraftNetwork` with the fake-git lane): mirror admitted with
  canonical provenance; mirror-first availability fallback; pin selects
  mirror and forbids fallback; fail-closed TLS stops before mirror;
  fallback-none stops after first failure; exhaustion sanitized (no
  secrets or URLs in errors, broker queries addressed to connection
  hosts); 15-row refusal table with zero fetches and zero records;
  revision-1 reader rejects revision-2 shapes; hostile insteadOf and
  GIT_SSH process config ignored; revision-2 plan conversion; port-aware
  SSH discovery.

## Verification (real exit codes, set -o pipefail shells)

- `go test -p 1 -count=1 ./internal/config/` → ok (6.053s)
- `go test -p 1 -count=1 ./internal/buildrepo/` → ok (250.330s)
- `go test -p 1 -count=1 ./internal/install/` → ok (461.812s)
- `go test -p 1 -count=1 -run DraftTransport ./cmd/curator/` → ok (521.833s)
- `go vet ./internal/install/ ./internal/buildrepo/ ./cmd/...` → clean
- `gofmt -l` on touched trees → clean
- `golangci-lint run` (full configured gate) → 0 issues
- `sh scripts/remote-gate.sh` → success: CI run 35178895680
  (branch gate/STORY-260916-v58b5y/260917-033801-84686-1, since cleaned
  up by the script). All 7 changed files verified byte-identical between
  the tested snapshot and this worktree (`git show <snap>:path | cmp`).

## Findings and decisions

- No executor fetch-path change was needed for mirrors: a declared
  mirror URL is already an ordinary lane URL, so SSH wrapper binding,
  HTTPS credential host checks, and raw-object proof apply identically.
  Ports and aliases cannot be fetched in this lane without widening the
  frozen lane grammar or the closed SSH wrapper argv, so refusal is the
  only compliant behavior here; Skillfile-lane admission (spec section 4)
  has no network caller in this tree and stays future work.
- Broker addressing uses the connection host (mirror host for mirror
  attempts); revision-1 behavior is identical (URL host equals key host).
- `references/negative-evidence.md` named in the brief does not exist in
  this tree; refusal rows follow the existing negative-row test shape
  instead (exact class plus zero fetches at the production entry).
- Prior TASK-260916-27cv45 evidence (schema-2 loader, goldens) accepted
  as attached: revision-1 suite green in the full package runs above,
  including `TestSchema1GoldenUnchanged` and the switch-off legacy
  selection tests.
