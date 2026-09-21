# Review verdict — BUG-260920-3ukdk4 revision 1: CHANGES_REQUESTED

Reviewer run RUN-260920-01ac41 (claude-opus-5), 2026-09-20 03:05–03:35Z, host e11-1
(load 7–15). Change Request CR-BUG-260920-3ukdk4-1 rev 1, base
`7fa08e84bc85921dca450fa71e40740cc2052a6a`, candidate tree
`25b9d748d150d2fecb34ad0c3fa37126d8b10500`. Read-only review: nothing in the
Story worktree was modified (one accidental `git add --intent-to-add` on the
untracked test file was undone with `git reset -q -- <path>`; working tree and
index are back to the handoff state).

## 1. Verdict in one paragraph

The fix is correct and complete on its behavioural surface: the draft
literal-URL lane (`cmd/curator/project_resolve.go` →
`gitops.CloneIsolated`/`FetchIsolated`) no longer consults user or system git
configuration through any channel I could plant (GLOBAL/SYSTEM selectors,
`~/.gitconfig`, `$XDG_CONFIG_HOME/git/config`, system-only, `include`/`includeIf`
rewriting, `core.hooksPath` execution, `core.sshCommand` on an `ssh://`
declaration, `GIT_CONFIG_COUNT/KEY/VALUE` injection), the corpus row
`v2-user-insteadof-ignored` passes as a DRIVEN row and the ratio line counts it
(90/3/1/0/94 locally and on the hosted macOS lane), legacy v1 and the resolved
lane are untouched, and the hosted gate ran on the exact candidate tree. Two
things stop acceptance: **(R1)** the fetch half of the fix at the production
call site (`project_resolve.go:201`) is pinned by no committed test — the
call-site mutant `FetchIsolated → Fetch` survives the whole candidate suite
(gitops, crossconformance row, cmd/curator resolve/refresh masks), and that is
the steady-state path every `resolve`/`refresh` after the first clone takes;
**(R2)** `docs/cli.md:503-505` still tells operators that `project resolve`
/`refresh` "clone and fetch with the operator's ambient Git credentials",
which after this change is wrong for the most common ambient mechanism
(credential helpers configured in user/system git config are now never
consulted). Both are small, bounded reworks; nothing else needs to change.

## 2. Exact-tree and gate proof

| Check | Result |
|---|---|
| Story worktree tree (temp index: `GIT_INDEX_FILE=$(mktemp) git read-tree HEAD && git add -A && git write-tree`) | `25b9d748…` = candidate ✓ |
| Patch resource sha256 | `7d66b56e…b791` = board record ✓ |
| Disposable clone: base `7fa08e84` + `git apply --index rev1.patch` + commit → `HEAD^{tree}` | `25b9d748…` ✓ (all throwaway tests and mutants ran in `/tmp/rev3ukdk4/{cand,probe,mut}`, never in the worktree) |
| Gate commit `c2784dc240e8ae5ffe5ba917405fdbab97979e10` | parent `7fa08e84`, tree `25b9d748…` ✓ |
| Hosted run 35482663765 (`gh run view --json headSha,conclusion`) | headSha `c2784dc…`, `success`; jobs: Interop, Test ×3, Race ×2, Gate self-test ×3, Lint, Naming all `success`; rose-air + candidate suite skipped |
| Hosted `go-test.json` (artifacts `test-evidence-*`) | ubuntu: `semantic cases: 89 driven, 3 known-gap, 1 bound, 1 skipped, 94 total`; macos: `90 driven, 3 known-gap, 1 bound, 0 skipped`; windows: `42 driven, 2 known-gap, 1 bound, 49 skipped` (row `v2-user-insteadof-ignored` **skip** on Windows — POSIX shim, declared). Row `pass` on ubuntu+macos; the 3 remaining known-gaps are the untouched siblings (`v2-alias-resolution`, `attestation-evidence-wrong-name/-context`). `TestUserConfigIgnoredByResolvedLane` pass ×3; the three new gitops isolation tests pass ×3 **including Windows** (real git). |

## 3. What I reran myself (bash drivers, `set -o pipefail`, real exit codes, logs under `/tmp/rev3ukdk4/logs`)

Candidate clone `cand` (tree `25b9d748…` before and after every driver):

- `gofmt -l` on the 4 changed files → empty, exit 0. `go build ./...` exit 0. `go vet ./internal/gitops/ ./cmd/curator/ ./internal/crossconformance/` exit 0.
- `go test ./internal/gitops/ -count=1 -p 1 -v -run 'TestCloneIsolatedIgnoresUserConfig|TestFetchIsolatedIgnoresUserConfig|TestIsolatedGitEnvPinsEmptyConfig'` → 3/3 PASS (subtests `global_and_system_selectors`, `home_file_without_selectors`, `environment_injected_config` PASS), exit 0, 3.7 s.
- `go test ./internal/gitops/ -count=1 -p 1` → ok, exit 0, 13.7 s.
- Precompiled `crossconformance.test`, `-test.run 'TestDraftSourcesSemanticCases/v2-user-(insteadof|ssh-alias)-ignored$'` → both subtests PASS (1.43 s / 3.87 s), `semantic cases: 2 driven, 0 known-gap…`, parent exit 1 only via the `executed != total` tally guard (expected under a `-run` filter).
- Full matrix `crossconformance.test -test.run '^TestDraftSourcesSemanticCases$' -test.timeout=1800s` → **exit 0, 371.8 s**, `semantic cases: 90 driven, 3 known-gap, 1 bound, 0 skipped, 94 total`; `v2-user-insteadof-ignored` `--- PASS` with no KNOWN-GAP marker; known-gaps = the 3 untouched siblings, bound = `capture-mutation`. The executed==total assertion is intact (`draftsources_semantic_test.go:54-58`).
- `go test ./cmd/curator/ -count=1 -run '^TestDraftTransportLegacyGolden$'` → PASS, exit 0, 22.1 s (legacy/resolved lane golden byte-identical). `TestProjectResolveLegacyUntouched`, `TestProjectResolveDraftOffRefuses` → PASS (precompiled binary). Note: the precompiled-binary run of the golden test FAILED only because the golden normalizes `core.askPass=$T/go-build$N/b001/curator.test` and my binary lived at `/tmp/rev3ukdk4/bin/curator.test` — a harness artefact, cleared by the `go test` rerun above.
- Resolved lane unchanged structurally: the diff touches no file under `internal/install` or `internal/buildrepo`, and neither package imports `internal/gitops` (importers: cmd/curator, closure, contextstore, envprofile, snapshot). Hosted gate: `internal/install` pass ×3, `TestUserConfigIgnoredByResolvedLane` pass ×3.
- Legacy env construction byte-identical by inspection: `run` = `runWithEnv(dir, append(os.Environ(), "GIT_ALLOW_PROTOCOL="+AllowedProtocols), …)`; `Clone/Fetch/HasRemote` delegate to the same cores with `run`; only `run(` → `runFn(` substitutions in the cores.

Accepted from attached evidence, not rerun: the full `cmd/curator` and `internal/install` packages (hosted gate exit 0 on the exact tree; local full runs would not fit the bounded-call budget).

## 4. Reviewer probes at the production entry (`probe` clone, `zz_review_probe_test.go`, attached)

Each probe plants a hostile configuration through ONE channel, runs a raw-git
positive control that must bind the evil fixture (payload live), then drives
`curator project resolve|refresh` through the compiled CLI and the transport
shim and asserts the lock binds the declared commit and the canonical identity.
Candidate: **7/7 PASS, exit 0** (28 s incl. CLI build).

| # | Channel | Control | Candidate |
|---|---|---|---|
| P1 | global file with only `[include] path=` + `[includeIf "gitdir:…/**"] path=` → payload file with `insteadOf` | evil | declared, shim log = declared URL once |
| P2 | `core.hooksPath` + `post-checkout` marker hook, plus insteadOf | evil + hook marker written | declared, **no marker** (config-driven execution never runs) |
| P3 | payload only at child `$XDG_CONFIG_HOME/git/config`, selectors unset | evil | declared |
| P4 | `GIT_CONFIG_SYSTEM=hostile`, `GIT_CONFIG_NOSYSTEM=0`, no global payload | evil | declared |
| P5 | `ssh://git@fixture.test/kit.git` declaration; `core.sshCommand=<stand-in that serves the evil bare via upload-pack>` in selectors AND child `~/.gitconfig` | raw clone succeeds with evil HEAD | resolve **fails closed** `repository_endpoint_unavailable` (availability), stand-in never invoked, no lock written, stderr carries class only (no URL/tool path) |
| P6 | **refresh path**: clean initial resolve (`branch: main`), advance declared `main`, then plant insteadOf (selectors + child `~/.gitconfig`) and run `project refresh` | raw fetch under hostile env moves `origin/main` to evil | lock = advanced declared commit, shim log still exactly 1 clone (fetch, not re-clone) |
| P7 | `GIT_CONFIG_COUNT=1`, `GIT_CONFIG_KEY_0=url.<evil>.insteadOf`, `GIT_CONFIG_VALUE_0=<declared>` in the child env | evil | declared |

P8 (residual scan, not this leaf's AC): a `ProxyCommand` marker in the child's
`~/.ssh/config` was NOT hit — inconclusive, because OpenSSH resolves the home
directory through passwd, not `$HOME`, and writing the operator's real
`~/.ssh/config` is out of bounds. The run did show the isolated `ssh://` path
failing closed with a sanitized `availability` class.

## 5. Mutants (`mut` clone; `git checkout --` between mutants; tree restored to `25b9d748…`)

| Mutant | Candidate's own tests | Reviewer probes |
|---|---|---|
| M1 `CloneIsolated`/`FetchIsolated` closures → legacy `run` (producer's mutant) | **killed**: 4 gitops behavioural rows FAIL; corpus row FAIL (`lock binds the evil commit …: user insteadOf was applied`) | all 7 FAIL |
| M2 scrub selectors, drop the `GIT_CONFIG_GLOBAL/SYSTEM=<empty>` pins | **killed**: `home_file_without_selectors` FAIL, `TestIsolatedGitEnvPinsEmptyConfig` FAIL, corpus row FAIL | P3, P5, P6 FAIL |
| M3 pin GLOBAL only; SYSTEM/NOSYSTEM ambient (not scrubbed, not pinned) | **killed**: `global_and_system_selectors` FAIL, fetch row FAIL, structural FAIL, corpus row FAIL | P1, P2, P4, P5, P6 FAIL |
| M4 stop scrubbing `GIT_CONFIG_COUNT/PARAMETERS/KEY_*/VALUE_*` | **killed in-package only**: `environment_injected_config` FAIL, structural FAIL; corpus row PASS | P7 FAIL |
| M8 call site `project_resolve.go:217` `CloneIsolated → Clone` | **killed**: corpus row FAIL | P1–P5, P7 FAIL |
| **M7 call site `project_resolve.go:201` `FetchIsolated → Fetch`** | **SURVIVES**: gitops rows unaffected (they test the functions, not the call site); corpus row PASS (fresh clone only); cmd/curator masks `TestProjectResolve|TestDraftSources|TestDraftStatus|TestDraftDiagnostics|TestDraftTransport` 34 PASS / 1 FAIL (the askpass-path golden artefact of §3, identical under the candidate) | **only P6 kills it** (`refresh = 1`: the fetch went to evil and failed on the clobbered tag) |
| M6 `strings.EqualFold` → `==` in `scrubsGitConfigEnv` | equivalent on POSIX (Windows-only distinction) — bound, not run |

## 6. Findings

### R1 (required) — fetch call site unpinned at the production entry

`cmd/curator/project_resolve.go:201` (`gitops.FetchIsolated(repoDir)`) has no
committed negative test: every `project resolve`/`refresh` after the first
clone takes this path, and mutant M7 survives the candidate suite (§5). The
results.md claim "the narrowing mutant (user config re-enabled) is killed at the
production entry and in-package" holds for the clone half only; the producer's
mutant altered the gitops functions, which the in-package tests see, not the
call sites. Rework: add one production-entry row for the refresh/fetch path —
P6 in the attached probe file is a ready shape (clean resolve with
`branch: main` → advance the declared bare → plant insteadOf via selectors and
child `~/.gitconfig` → `project refresh` → lock binds the advanced declared
commit, shim log still shows exactly one clone). Either as a non-corpus test
beside `driveV2UserInsteadOf` in `internal/crossconformance` (do not register
it as a corpus row; the ratio stays 94) or as a `capture`-driven test in
`cmd/curator` with `t.Setenv("GIT_CONFIG_GLOBAL", hostile)` (the fixture
pattern `envconfig_test.go:532-536` already uses). Reproduction: apply
`sed -i '' 's/gitops.FetchIsolated(repoDir)/gitops.Fetch(repoDir)/' cmd/curator/project_resolve.go`
and run the new test — it must fail; the candidate's current suite does not.

### R2 (required) — operator doc now misdescribes the lane

`docs/cli.md:503-505`: "`project resolve` and `project refresh` clone and fetch
with the operator's ambient Git credentials". After this change the
`credential.helper` configured in user or system git config (this host's
global config carries `credential.helper` plus per-host `gh` helpers; stock
installs ship osxkeychain / `manager` at system level) is never consulted, so a
private HTTPS Skillfile source that worked yesterday through the helper now
fails `repository_endpoint_unavailable`/authentication with no prompt. The producer's ruling (results.md) is consistent with the contract
(repository-transport §3: "does not import … helpers, include files, environment
overrides"; §2: "No interactive credential discovery occurs in a headless run"),
but it must reach the operator: restate the sentence as "with credentials from
the invoking environment only (SSH agent via `SSH_AUTH_SOCK`, `GIT_ASKPASS`,
`GIT_SSH`/`GIT_SSH_COMMAND`); user and system Git configuration — credential
helpers, `insteadOf`/`pushInsteadOf`, URL rewriting and includes,
`core.sshCommand` — is never consulted, and no interactive prompt occurs". Keep
the pinned phrase or update the pin at `cmd/curator/draft_diagnostics_test.go:1141`
(`"ambient Git credentials"`). Recommended alongside: a troubleshooting remedy
for private HTTPS sources on this lane, and a CHANGELOG `Unreleased` entry — the
repo records notable implementation changes there and this is a
security-relevant behaviour change.

### Non-blocking notes (record, do not gate rev 2 on them)

- N1 `internal/closure/resolve.go:650` (`pinGitAliases`): the fallback
  `fetch = gitops.Fetch` is the ambient variant. Production pre-marks every
  alias root (`project_resolve.go:121-124`, same `repoKey`), so it never fires
  today, but it is a latent ambient fetch on draft alias trees; making the
  fallback `gitops.FetchIsolated` (or dropping it for alias roots) closes the
  class rather than the instance.
- N2 Residuals outside this leaf's AC, for the orchestrator to scope as
  siblings: per-invocation environment overrides (`GIT_SSH_COMMAND`, `GIT_SSH`,
  `GIT_PROXY_COMMAND`, `GIT_EXEC_PATH`, `GIT_ASKPASS`) are still honoured on the
  draft literal lane (the producer's ruling discloses this; the contract lists
  "environment overrides" among non-imports); user `~/.ssh/config`
  (Host aliases, `ProxyCommand`) is still read by the real ssh for literal
  `ssh://` declarations on this lane (contract §5 "remain NOT imported") —
  reasoned, not demonstrated (P8 bound). The corpus row
  `v2-user-ssh-alias-ignored` covers only the logical declaration that fails at
  planning.
- N3 Windows: `GIT_CONFIG_NOSYSTEM=1` also drops Git for Windows' distro system
  config (`http.sslBackend`, `http.sslCAInfo`). Whether HTTPS verification still
  works there depends on compiled curl defaults — unverified (gate has no
  network). The isolation itself is proven on Windows in-package (hosted gitops
  tests, real git, file://); the corpus row is a declared POSIX skip there.
- N4 Production-entry coverage of env-injected config (`GIT_CONFIG_COUNT/KEY/VALUE`)
  is in-package only (M4); P7 covered it in review. Optional to add.
- N5 M6 (`EqualFold` vs `==`) is an equivalence bound on POSIX; the structural
  test lowercases names via `ToUpper`, so Windows case-insensitivity is an
  unpinned but harmless property.

## 7. Contract and architecture fit

- `isolatedGitEnv` (`internal/gitops/gitops.go:89-131`) mirrors
  `buildrepo.cleanGitEnvironment` for the persistent-config channels
  (GLOBAL/SYSTEM pinned to a fresh empty file per call, `GIT_CONFIG_NOSYSTEM=1`,
  `GIT_TERMINAL_PROMPT=0`, `GIT_ALLOW_PROTOCOL` re-pinned) while keeping the
  operator's invocation environment — the brief's "GIT_CONFIG_GLOBAL/SYSTEM
  pointed at empty files … keep GIT_TERMINAL_PROMPT=0" option. Fail-closed when
  the temp file cannot be created. HOME-independent by construction
  (`GIT_CONFIG_GLOBAL` set ⇒ neither `~/.gitconfig` nor `$XDG_CONFIG_HOME/git/config`
  is read) and proven by P3 and the candidate's `home_file_without_selectors` row.
- Draft-lane-only call sites; legacy `Clone`/`Fetch` untouched; the
  `envprofile` git-source lane (`internal/envprofile/gitsource.go`) still uses
  the ambient variant by design (its `insteadOf` fixtures in
  `cmd/curator/envconfig_test.go`/`profile_test.go` depend on it) — out of scope,
  noted.
- Diagnostics stay sanitized (P5/P8 stderr: class vocabulary and canonical
  identity only).
- Corpus fidelity: the vendored `v2-user-insteadof-ignored` expectation
  (`attempt-declared-url-once-literally;insteadOf-never-applied`) matches
  curator-spec main; the row asserts both halves (shim log = declared URL once;
  lock ≠ evil, = declared).

## 8. Routing

`set_status(BUG-260920-3ukdk4, status=to-dev)`; rework R1 + R2 (N1 optional),
same producer role/archetype, new revision, new gate, new review. No human
decision is needed.

Attachments: this verdict; `BUG-260920-3ukdk4_review-rev1-probes.go.txt`
(the eight probes as run, package `crossconformance`, not to be committed
as-is — P6 is the intended shape for R1).
