# Review verdict — BUG-260920-3v7x6j revision 1 (CR-BUG-260920-3v7x6j-1)

Reviewer run RUN-260920-738ae2 (claude-opus-5, independent). Verdict: **ACCEPT**
(`accept_cr(BUG-260920-3v7x6j, revision=1, evidence=BUG-260920-3v7x6j_review-verdict-rev1.md)`).

## 1. Exact candidate

- Story worktree `.temp/STORY-260919-37szes/worktree`, HEAD `9be96486` (= CR base), 5 modified
  paths; temp-index `git write-tree` over the worktree = `581dd81a32452ce5b4e0ca7d424580f1a0c880d6`
  = CR candidate tree OID.
- Patch resource `BUG-260920-3v7x6j_change-request_rev1.patch` sha256
  `614c69730f5664cfc8beb8f58f72b449f9510d4319781683365049c919fdfe93` (matches the CR record).
- Disposable clones `/tmp/3v7x6j-review/cand` and `/tmp/3v7x6j-review/mut`: `git clone --no-checkout`
  of the control root → `checkout --detach 9be96486` → `git apply --index <patch>` → commit;
  `HEAD^{tree}` = `581dd81a…` in both; every driver printed `git write-tree` = `581dd81a…` at start
  and end (mutants restored with `git checkout --`). The Story worktree was never written.
- Hosted gate (validation log): `scripts/remote-gate.sh` → run 35499742447, conclusion `success`
  on every lane (lint, test/race macOS+ubuntu, test windows, gate self-tests, interop);
  `gh run view … --json headSha` = `ab9e431b04d8c236818eff629c0e1326d84a5d27`;
  `git rev-parse ab9e431b^{tree}` = `581dd81a…` = candidate. Gate == reviewed tree.

## 2. What the change does (read)

- `internal/config/sourcepolicy.go`: `Attempt.ConnectionURL() (string, bool)` renders the §5
  connection target: alias-less attempts return the listed URL verbatim (legacy/v1, explicit-port and
  declared-mirror attempts byte-identical); an aliased attempt keeps scheme, userinfo and path and
  replaces host (+port: alias port, else URL port, else transport default — the loader already folds
  this into `ResolvedPort`/`HasExplicitPort`, `checkEndpointSemantics` :800–808). scp-like + alias
  port renders `ssh://[user@]host:port/~/path`. `ok=false` (fail closed) on any carried value outside
  the grammar (`validAliasName` on the resolved host, port-flag/value agreement, port range, URL shape).
- `cmd/curator/project_resolve.go:214–229` (`fetchDraftRepoAllowingFallback`): every attempt's target
  is derived before any clone; a mistranslation fails `repository_policy_invalid` with a sanitized
  message (no URL/host); `gitops.CloneIsolated(targets[index], …)` replaces `attempt.URL`.
  Diagnostics (`draftAttemptClause`: transport + provider + class only), the §2/§6 fallback gate and
  the exhaustion shape are untouched. The refresh branch (`FetchIsolated(repoDir)`) fetches the
  checkout's origin, which after this change is the substituted address.
- Tests: 13 substitution + 18 fail-closed + 3 loader round-trip unit cases; corpus row
  `v2-alias-resolution` converted from known-gap lock to a drive (shim serves ONLY the substituted
  address `https://mirror.corp.example:8443/kit.git`, `clones == [substituted]`, lock identity
  canonical, stderr must not carry the resolved host or the alias name); the now-unused
  `semanticKnownGap` helper removed (outcome vocabulary and tally kept).

## 3. Scope decision — executor lane deliberately unchanged (verified against the contract)

The brief and the rev1 review note name the executor (`internal/buildrepo/transport.go:502, 931`) as
a second substitution site. I checked this against curator-spec `protocol/repository-transport.md`
(main, pin 802caee) and the corpus:

- §4: "For the strict external-build lane (manager section 11) only the port-free, alias-free subset
  applies: endpoints with an explicit port … or an `alias` field are refused there … and are admitted
  only for Skillfile source acquisition". §5: "An endpoint naming an `alias` MUST NOT be selected for
  the strict external-build lane". §7: such an endpoint "fails `build_repository_identity_invalid`
  before network I/O … The manager MUST NOT strip the port or ignore the alias to force admission."
- Corpus `v2-external-build-alias-refused` expects
  `build_repository_identity_invalid;alias-rewriting-has-no-resolved-host-input-in-strict-lane;zero-attempts`.
- Code: the executor is reached only through `install.acquireDraftNetwork` (external-repository
  lane, `internal/install/external.go:407`); `parseTransportPlan` refuses any port/alias attempt at
  `transport.go:489` before `sources[index]` exists, so lines 502/931 are reachable only by port-free,
  alias-free attempts, for which the resolved host equals the URL host (connection already equals the
  resolved address). Mutant M8 (refusal removed) is killed by internal/buildrepo (6 failures),
  internal/install `TestAcquireDraftNetworkRevision2Refusals` and the two `v2-external-build-*-refused`
  corpus rows — the gate is live.
- The only production lane that admits aliases is Skillfile-source acquisition through
  `curator project resolve|refresh` (`acquireDraftGitRoots → fetchDraftRepoAllowingFallback`), which is
  where the fix lands; `internal/install` draft sources read the frozen lock/bindings and never fetch
  (`draftsources.go`). Substituting inside the executor would violate §7 and break the pinned rows.

I accept the producer's reading: the AC's "connection target" requirement is satisfied at the only
lane where an alias can be selected; the "executor lane" wording in the brief was carried over from
the 1xya7x rev3 finding and is not satisfiable under the accepted contract.

## 4. Independent reruns (clone `cand`, tree 581dd81a…, `#!/bin/bash` driver, `go test -p 1 -count=1`,
exit codes from the driver log; darwin/amd64 go1.26.0, host load 5–9)

| Command | Result |
|---|---|
| `go build ./...` | exit 0 |
| `go vet ./internal/config ./cmd/curator ./internal/crossconformance` | exit 0 |
| `gofmt -l` on the three dirs | 0 lines |
| `go test -run TestAttemptConnectionURL -v ./internal/config/` | exit 0, 34 `--- PASS`, 0 FAIL |
| `go test ./internal/config/` | exit 0 (`ok 0.787s`) |
| `go test -run 'TestDraftAttempt\|TestEndpointExhaustion\|TestDraftStripCloneFraming\|TestWithDraftRemediation\|TestDraftFetchFailure\|TestDraftAvailabilityFallback\|TestDraftMachinePolicy' ./cmd/curator/` | exit 0 (`ok 8.682s`) |
| `go test -run 'TestProjectResolve\|TestProjectRefresh' ./cmd/curator/` | exit 0 (`ok 110.143s`) |
| `go test -run 'TestDraftSourcesSemanticCases/v2-' -v ./internal/crossconformance/` | 21/21 subtests `--- PASS`, 0 FAIL, 0 SKIP, 0 KNOWN-GAP lines; `v2-alias-resolution` PASS (1.42 s); parent FAIL only by the executed-count guard `executed(21) != total(94)` (by design under `-run`) |

Full 94-row matrix locally: two attempts, both killed by the host (first by my own wait-loop
timeout, second `signal: killed` of the fresh test binary after 75.8 s before its first output — the
same pathology the producer reported). The hosted gate ledger is the tally arbiter (§5).

## 5. Hosted per-lane ledger (artifacts `test-evidence-<os>` of run 35499742447, `test/go-test.json`
+ `test/observed-cases.tsv`)

| Lane | ratio line | `v2-alias-resolution` | semantic subtests |
|---|---|---|---|
| macos-latest | `semantic cases: 93 driven, 0 known-gap, 1 bound, 0 skipped, 94 total` | pass | 94 pass |
| ubuntu-latest | `92 driven, 0 known-gap, 1 bound, 1 skipped, 94 total` | pass | 93 pass, 1 skip (`case-alias`, declared FS reason) |
| windows-latest | `42 driven, 0 known-gap, 1 bound, 51 skipped, 94 total` | skip `test transport wrapper is POSIX-only` (declared, as its siblings) | 43 pass, 51 skip |

Parent `TestDraftSourcesSemanticCases` passed on all three lanes → executed == total (94) with the
rev4 guard; known-gap count is 0 on every lane (was 4 at 1xya7x rev4; the three sibling fixes and this
one account for 89 → 93 driven). The known-gap marker is gone from the tree (`grep semanticKnownGap` →
no callers). Windows has no POSIX shim, so the substituted connection is proven on macOS/ubuntu only;
the `ConnectionURL` unit tests ran on every lane: `observed-cases.tsv` shows 34 `TestAttemptConnectionURL*`
rows per lane (ubuntu, macOS, windows), 0 fail, 0 skip.

## 6. Reviewer probes at the production entry (clone `mut` + `zz_review_probe_test.go`, compiled CLI +
default-deny transport shim recording clone targets)

| Probe | Result |
|---|---|
| P1 review-note item 6: alias policy + hostile git config planted via `GIT_CONFIG_GLOBAL`, `GIT_CONFIG_SYSTEM`, `GIT_CONFIG_NOSYSTEM=0` and child-HOME `.gitconfig` (`url.file://evil.insteadOf=file://bare`, `url.https://never.invalid/x.git.insteadOf=<substituted URL>`, `core.sshCommand`) | PASS 9.55 s: resolve exit 0, `clones=[https://mirror.corp.example:8443/kit.git]`, lock commit = declared fixture commit (not the evil one), identity `fixture.test/kit` — the substituted connection is not re-rewritten by user configuration |
| P2 substituted connection fails (shim answers `Could not resolve host: mirror.corp.example`) | PASS 1.75 s: exit ≠ 0, one clone of the substituted address; stderr = `warning: kit: endpoint 1 (https, provider "team-https"): availability: …` + `curator: repository_endpoint_unavailable: kit: identity fixture.test/kit: endpoint 1 (…); verify the network path…`; no `mirror.corp.example`, `corp-mirror`, `8443`, listed URL or raw git text |
| P3 alias endpoint (availability failure) then declared mirror under `availability-auth` | PASS 2.07 s: exit 0, `clones=[substituted alias URL, https://mirror.fixture.test/kit.git]`, sanitized warning, lock canonical |
| P4 refresh after an alias resolve (`branch: main`, bare advanced) | PASS 7.48 s: lock moved `4e024a5c → fe78abf6`, exactly one clone (the substituted address) across resolve+refresh — refresh fetches the substituted checkout's origin, no re-clone from the listed URL |

## 7. Mutants (clone `mut`, restored between, end tree 581dd81a…)

| Mutant | Tests | Outcome |
|---|---|---|
| M1 call-site narrowing: `CloneIsolated(attempt.URL, …)` (connect to the declared host) | unit exit 0; row `v2-alias-resolution` **FAIL** `resolve = 1, want success` (4.26 s) | KILLED at the production entry |
| M2 https branch drops the port | unit FAIL (3 cases); row FAIL at the in-row `ConnectionURL` pin | KILLED |
| M2b = M2 with the in-row pin deleted | row **FAIL** `resolve = 1, want success` (4.36 s): the shim refuses `https://mirror.corp.example/kit.git` | KILLED at the production entry (port is part of the exact-match target) |
| M3 substitution disabled (`if a.Alias == ""` → `if true`) | unit FAIL; row FAIL | KILLED |
| M4 resolved port off by one | unit FAIL; row FAIL | KILLED |
| M5 call-site fail-open (`!ok` → `target = attempt.URL`) | 21/21 v2 rows PASS; cmd/curator masks PASS (73.7 s) | SURVIVED — equivalent through production: `fetchDraftRepoAllowingFallback` only receives loader-planned attempts, which always form a target (`parseAliasTable` host grammar, `checkEndpointSemantics` port folding); the `ok=false` floor exists for hand-built attempts and is covered by 18 unit refusal cases. Bound, not a defect |
| M8 executor §7 refusal removed (`transport.go:489` → `if false`) | internal/buildrepo FAIL (6: `TestPortAliasRefusalNarrows`, `TestValidateTransportPlanRevision2Refusals/*`), install `TestAcquireDraftNetworkRevision2Refusals` FAIL, rows `v2-external-build-alias-refused` + `-port-refused` FAIL | KILLED — the strict-lane gate behind §3 is live |

## 8. Review-note items

1. Substituted endpoint CONNECTS to the resolved host:port on the CLI resolve lane: yes — the git
   command line carries `https://mirror.corp.example:8443/kit.git` (shim log), M1/M2b killed. Executor
   lane: not applicable by contract (§3). Declared identity, lock (`Package.Repository`
   `fixture.test/kit`, no port), bindings/markers unchanged.
2. Provenance: the strict-lane sink `draft-transport-provenance.jsonl` is untouched (records the
   resolved host as before); the CLI lane never had a sink (unchanged); no endpoint provenance in
   the lock or user-facing errors (P2/P3 scans).
3. Accepted rev2 grammar and refusals unchanged: all 11 refusal rows, exhaustion, executor cap and
   `v2-external-build-*` rows pass locally and on all gate lanes; loader code untouched.
4. Corpus row driven: known-gap marker removed, ratio counts it (0 known-gap), executed == total.
5. Narrowing mutant killed at the production entry (M1; plus M2b for the port).
6. git-config isolation holds for the substituted connection (P1).
7. Legacy v1 lane byte-identical: alias-less attempts return `a.URL` verbatim (`ConnectionURL`
   first branch; unit rows "legacy https/scp/port/mirror passes through"); the frozen v1 lane never
   enters this code; `TestProjectResolve*`/`TestProjectRefresh*` and `v2-reader-accepts-v1-policy`
   green.

## 9. Residuals (non-blocking, for the Story close / spec owner)

- R1 scp-like URL + alias port renders `ssh://[user@]host:port/~/path` (git's home-relative URI
  spelling; `git-upload-pack` expands `~/` via `enter_repo`; the admitted scp path grammar
  `[A-Za-z0-9._/-]+` has no `~`, so no `~user` ambiguity). §5 leaves this corner undefined; servers
  with custom SSH front-ends may not accept `~/`. The fail-closed alternative would be refusing
  scp+alias-port. Spec-owner note, not a defect.
- R2 `docs/draft-source-policy.md` ("This layer plans only") and `docs/draft-transport-resolution.md`
  do not mention `Attempt.ConnectionURL` or that CLI resolve connects to the substituted address.
  Docs are outside this leaf's AC.
- R3 The substituted address persists only as the checkout's `remote.origin.url` under the manager
  home (machine-private); a later alias re-point in policy reaches an existing checkout only through
  the fetch-failure → re-clone path (pre-existing checkout-reuse design, same for listed-URL changes).
- R4 M5 bound (above).
- R5 Windows: the substituted connection is unproven there by the POSIX-only shim design (declared
  skip, same as every sibling row); unit coverage of the target rendering ran on all lanes.

## 10. Producer evidence cross-check

Producer results (`BUG-260920-3v7x6j_results.md`): commands and exit codes reproduced by my reruns
(config, cmd/curator masks, v2 rows), mutant kill reproduced (M1), scope decision verified against the
spec text and the corpus. Producer's full-matrix and lint claims are accepted from the hosted gate
(Lint job success; ledger in §5), not from local observation.

Checklist items 10–13 checked by this run before `accept_cr`; no `commit_ack`; the leaf routes to
`integrating` for the producer-side checkpoint.
