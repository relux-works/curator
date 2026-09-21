# BUG-260920-3ukdk4 results — draft literal-URL lane ignores user git config

## Outcome

Ready for review. `curator project resolve`/`refresh` on a literal `git:` source
can no longer be redirected by user/system git configuration: clone and fetch
on the draft literal-URL lane run with `GIT_CONFIG_GLOBAL`/`GIT_CONFIG_SYSTEM`
pinned to a fresh empty file, `GIT_CONFIG_NOSYSTEM=1`, all `GIT_CONFIG_*`
selectors scrubbed, and `GIT_TERMINAL_PROMPT=0`, independent of `HOME`.
Corpus case `v2-user-insteadof-ignored` flips to driven-pass; the narrowing
mutant (user config re-enabled) is killed at the production entry and
in-package; legacy v1 lane and resolved lane unchanged.

## Contract ruling (ambient credentials)

repository-transport.md §§3/5: "User Git/SSH configuration is usable only
through explicit operator admission. This revision supports endpoint URLs and
authentication-provider selection only; it does not import insteadOf,
ProxyCommand, core.sshCommand, helpers, include files, environment overrides,
host aliases or arbitrary remapping. Credentials remain in the existing trusted
broker." + "No interactive credential discovery occurs in a headless run."
skillfile-sources.md adds no credential language.

Ruling applied to the draft literal lane:

- Persistent user/system configuration (`~/.gitconfig`,
  `$XDG_CONFIG_HOME/git/config`, `/etc/gitconfig`, includes, helpers,
  insteadOf/pushInsteadOf, url rewriting, core.sshCommand): NEVER consulted.
  `GIT_CONFIG_*` environment selectors choose persistent config, so they are
  pinned/scrubbed, not honored.
- Per-invocation operator context (`PATH`, `SSH_AUTH_SOCK`, `GIT_ASKPASS`,
  `GIT_SSH`, author identity) is the operator's explicit invocation, not
  "user configuration", and is preserved — SSH agent/key auth keeps working.
- `GIT_TERMINAL_PROMPT=0` (no interactive prompts) and the existing
  `GIT_ALLOW_PROTOCOL` pin are kept.

`core.sshCommand`/helper/pushInsteadOf breadth: same mechanism (the config
file is never read), proven structurally by the empty pinned files plus
behavioral insteadOf rows. Per-invocation `GIT_SSH`/`GIT_SSH_COMMAND` remain
honored by design (explicit operator admission); the hostile-SSH refusal case
`v2-user-ssh-alias-ignored` stays driven at planning with zero fetch attempts.

## Changes (uncommitted, story worktree)

- `internal/gitops/gitops.go`: `run` split into `run` + `runWithEnv` (legacy
  env construction byte-identical); new `isolatedGitEnv` (fresh empty config
  file per call, fail-closed; case-insensitive scrub of `GIT_CONFIG_GLOBAL`,
  `GIT_CONFIG_SYSTEM`, `GIT_CONFIG_NOSYSTEM`, `GIT_CONFIG_COUNT`,
  `GIT_CONFIG_PARAMETERS`, `GIT_CONFIG_KEY_*`, `GIT_CONFIG_VALUE_*`;
  re-pins `GIT_ALLOW_PROTOCOL`, `GIT_TERMINAL_PROMPT=0`); new
  `CloneIsolated`/`FetchIsolated` used only by the draft lane. Legacy
  `Clone`/`Fetch`/`HasRemote` delegate to the same shared cores with `run`.
- `cmd/curator/project_resolve.go`: `fetchDraftRepoAllowingFallback` (shared
  by resolve and refresh) uses `CloneIsolated`/`FetchIsolated`.
- `internal/gitops/isolated_test.go` (new): behavioral rows (hostile
  GLOBAL/SYSTEM selectors, hostile `$HOME/.gitconfig` with selectors unset,
  `GIT_CONFIG_COUNT`/`KEY_0`/`VALUE_0` injection; each with a raw-git
  positive control proving the payload redirects when consulted) for clone
  and fetch, plus a structural `isolatedGitEnv` pinning test.
- `internal/crossconformance/draftsources_semantic_v2_test.go`:
  `driveV2UserInsteadOfGap` rewritten as driven `driveV2UserInsteadOf`
  (production entry `project resolve`; hostile config via selectors AND
  child HOME file; asserts declared URL attempted once literally and lock
  binds the declared commit, never evil); header comment updated.

## Evidence (all exit codes real, shell `sh`, `set -o pipefail`)

Pre-fix reproduction (gap signature locks):

- `go test ./internal/crossconformance/ -run
  'TestDraftSourcesSemanticCases/v2-user-insteadof-ignored'` →
  subtest PASS logging `KNOWN-GAP ... lock binds the substituted commit`
  (parent exit 1 is only the `-run` tally guard). Confirms the exploit.

Post-fix (final tree):

- `go test ./internal/gitops/ -p 1 -count=1` → ok, exit 0 (20.8s).
- New isolation tests → all PASS, exit 0 (5.1s).
- `TestDraftSourcesSemanticCases/v2-user-insteadof-ignored` → subtest PASS,
  no KNOWN-GAP marker, exit 1 parent only via tally guard (6.9s).
- Neighbor v2 rows (`v2-port-endpoint`, `v2-declared-mirror`,
  `v2-mirror-first`, `v2-reader-accepts-v1-policy`,
  `v2-user-ssh-alias-ignored`) → all subtests PASS.
- `go test ./cmd/curator/ -run 'TestProjectResolve'` → ok, exit 0 (117s).
- `go test ./cmd/curator/ -run 'TestDraftTransport'` → ok, exit 0 (98s).
- `TestProjectResolveLegacyUntouched` → PASS, exit 0 (legacy v1 lane).
- `TestUserConfigIgnoredByResolvedLane` (internal/install) → PASS, exit 0
  (resolved lane unchanged).
- Full matrix `go test ./internal/crossconformance/ -timeout 1800s -run
  'TestDraftSourcesSemanticCases'` → ok, exit 0 (477s):
  `semantic cases: 90 driven, 3 known-gap, 1 bound, 0 skipped, 94 total`;
  remaining known-gaps are the untouched siblings
  (`v2-alias-resolution`, `attestation-evidence-wrong-name`,
  `attestation-evidence-wrong-context`); `v2-user-insteadof-ignored`
  `--- PASS` with no gap marker (counted driven). (First attempt hit the
  10m default timeout under host load with zero row failures; rerun with
  1800s passed.)
- `go vet` on all touched packages → exit 0; `gofmt -l` clean;
  `golangci-lint run` on gitops/cmd/curator/crossconformance → 0 issues.

Narrowing mutant (CloneIsolated/FetchIsolated reverted to legacy `run`,
production call graph intact):

- gitops isolation tests → FAIL binding evil on all 4 behavioral rows.
- `v2-user-insteadof-ignored` production-entry row → FAIL:
  `lock binds the evil commit ...: user insteadOf was applied`.
- Mutant reverted; final tree re-verified (build + tests above).

## Notes / residuals

- Host episode mid-run: all binary execs SIGKILLed (~10 min), then recovered;
  the two full-package runs that hit it were rerun green (reported above).
- Windows: the flipped row uses the POSIX transport shim like all sibling
  transport rows (declared skip on Windows); the implementation itself is
  portable (temp empty file, no `/dev/null`). Hosted gate owns Windows proof.
- `GIT_SSH`/`GIT_SSH_COMMAND` per-invocation env intentionally preserved
  (ruling above); a hostile `GIT_SSH_COMMAND` in the invoking environment is
  operator context, not user git config, and out of this leaf's scope.

---

# Revision 2 (rework R1 + R2, N1)

Ready for review. Revision 1's isolation stands (unchanged:
`internal/gitops/gitops.go`, `internal/gitops/isolated_test.go`,
`cmd/curator/project_resolve.go` call sites, the driven corpus row);
revision 2 adds the missing fetch-path production-entry pin, the
operator-doc correction, and the latent alias-fallback close.

## Changes vs revision 1 (uncommitted, story worktree)

- R1 — `internal/crossconformance/draftsources_semantic_v2_test.go`:
  new NON-corpus test `TestDraftLiteralRefreshIgnoresUserConfig`
  beside `driveV2UserInsteadOf` (reviewer probe P6 shape, adapted to
  local helpers; no `registerDraftSemantic` call, so the semantic
  ratio stays 94 and `TestDraftSourcesSemanticCoverage` is
  unaffected). Clean `project resolve` with `branch: main`, advance
  the declared bare repo, plant insteadOf via selectors AND the
  child's `~/.gitconfig`, run `project refresh`: asserts exit 0,
  the shim log shows exactly one clone (fetch, not re-clone), and
  the lock binds the advanced DECLARED commit, never evil's main.
  Includes the raw-git positive control (clone declared, hostile
  fetch lands on evil's main) proving the payload redirects fetches
  when consulted.
- R2 — `docs/cli.md`: the lane sentence restated per the verdict
  ("with credentials from the invoking environment only (SSH agent
  via `SSH_AUTH_SOCK`, `GIT_ASKPASS`, `GIT_SSH`/`GIT_SSH_COMMAND`);
  user and system Git configuration — credential helpers,
  `insteadOf`/`pushInsteadOf`, URL rewriting and includes,
  `core.sshCommand` — is never consulted, and no interactive prompt
  occurs") plus the N3 Windows system-config note.
  `docs/troubleshooting.md` (`repository_endpoint_unavailable`):
  remedy for private HTTPS Skillfile sources (helper never
  consulted, no prompt; provide a non-interactive `GIT_ASKPASS`
  program or an SSH endpoint with an agent via `SSH_AUTH_SOCK`).
  `cmd/curator/draft_diagnostics_test.go:1141`: the
  `"ambient Git credentials"` pin replaced with
  `"credentials from the invoking environment"` and
  `"is never consulted"`. `CHANGELOG.md` Unreleased → Fixed entry
  (security-relevant behaviour change).
- N1 — `internal/closure/resolve.go` (`pinGitAliases`): the fallback
  `fetch = gitops.Fetch` is now `gitops.FetchIsolated`, so the
  class is closed even where production does not pre-mark an alias
  root. Legacy/transitive fallbacks in `internal/closure/closure.go`
  untouched by design. Unit row
  `TestResolveDraftAliasFetchFallbackIgnoresUserConfig` in
  `internal/closure/resolve_test.go`: hostile user config
  redirecting the alias remote, `Fetch: true`, `FetchRepo: nil`
  (exercises the production fallback); asserts the lock binds the
  declared tag with a raw-fetch positive control.

## Evidence (all exit codes real, shell `sh`, `set -o pipefail`)

- `go test -p 1 ./internal/gitops/ -count=1` → ok, exit 0 (18.0s).
- `TestDraftLiteralRefreshIgnoresUserConfig` (new R1 row) → PASS,
  exit 0 (7.6s first run; 3–8s on reruns).
- `TestResolveDraftAliasFetchFallbackIgnoresUserConfig` (new N1
  row) → PASS, exit 0 (1.6–6.4s).
- Corpus row: `-run
  'DraftSourcesSemanticCases/v2-user-insteadof-ignored'` →
  `--- PASS ... (3.63s)`, `semantic cases: 1 driven, 0 known-gap,
  0 bound, 0 skipped, 94 total`, no KNOWN-GAP marker; parent
  exit 1 ONLY via the `-run` tally guard (`executed(1) !=
  total(94)`), as designed.
- `TestDraftSourcesSemanticCoverage` → PASS, exit 0 (matrix/corpus
  agreement; no unregistered driver; ratio base stays 94).
- `cmd/curator -run
  'TestDraftDocsPinExamples|TestWithDraftRemediationTable|TestDraftRemediationThroughCLI'`
  → ok, exit 0 (6.1s).
- `go test -p 1 ./internal/closure/ -count=1` → ok, exit 0 (46.5s).
- Legacy: `cmd/curator -run
  'TestDraftTransportLegacyGolden|TestProjectResolveLegacyUntouched'`
  → both PASS, exit 0 (21.1s). Resolved lane untouched (no file
  under `internal/install` or `internal/buildrepo` in the diff).
- M7 (`sed -i '' 's/gitops.FetchIsolated(repoDir)/gitops.Fetch(repoDir)/'
  cmd/curator/project_resolve.go`, the reviewer's exact mutation) →
  new R1 row FAILS, exit 1 (`refresh` fails closed
  `repository_endpoint_unavailable` after the hostile fetch;
  `tail` shows the sanitized class, no URL/tool leak). Mutant
  reverted; file verified byte-identical to the fix
  (`FetchIsolated` at :201, diff shows only intended lines).
- N1 mutant (fallback `FetchIsolated` → `Fetch`) → new closure row
  FAILS, exit 1 (`[rejected] v1 -> v1 (would clobber existing
  tag)` — the ambient fetch went to evil). Reverted; green rerun
  exit 0.
- `go build ./...` → exit 0; `go vet` on
  gitops/closure/cmd/curator/crossconformance → exit 0; `gofmt -l`
  on all touched files → empty.

## Bounds (reviewer wording, recorded not widened)

- N2: per-invocation environment overrides (`GIT_SSH_COMMAND`,
  `GIT_SSH`, `GIT_PROXY_COMMAND`, `GIT_EXEC_PATH`, `GIT_ASKPASS`)
  are still honoured on the draft literal lane; user `~/.ssh/config`
  (Host aliases, `ProxyCommand`) is still read by the real ssh for
  literal `ssh://` declarations on this lane (contract §5 "remain
  NOT imported") — reasoned, not demonstrated (P8 bound). Filed as
  a sibling by the orchestrator; out of this leaf.
- N3: Windows `GIT_CONFIG_NOSYSTEM=1` also drops Git for Windows'
  distro system config (`http.sslBackend`, `http.sslCAInfo`) —
  HTTPS verification on Windows unverified (gate has no network);
  noted in `docs/cli.md`.
- N4/N5 bounds from rev1 stand (env-injected config covered
  in-package + reviewer P7; `EqualFold` vs `==` equivalent on
  POSIX).

## Host anomaly (evidence-backed, worked around)

Mid-run every test-binary exec via `/tmp` stalled ~80s (wall, ~0
CPU, zero output) then SIGKILLed — including an unchanged package's
binary and a trivial `fmt.Println` binary, ruling out my code and
any single mutant. `go build`, `go vet`, `gofmt`, and pre-existing
signed binaries kept working. The identical trivial binary copied
to the workspace ran instantly (exit 0), scoping the fault to exec
from `/tmp`. All evidence above gathered after ~08:03 local ran
with `TMPDIR=$PWD/tmp-exec-test` (test binaries + fixtures
outside `/tmp`); the directory was removed afterwards and `git
status` shows only the nine intended files. Pre-episode results
(M7 first kill, initial R1/N1 greens, first gitops full-package
ok) were re-proven after the workaround where they gate
acceptance.
