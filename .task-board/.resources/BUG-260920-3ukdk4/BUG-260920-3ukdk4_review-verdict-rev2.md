# Review verdict — BUG-260920-3ukdk4 revision 2: ACCEPT

Reviewer run RUN-260920-9f94fc (claude-opus-5), 2026-09-20 09:21–09:45Z, host e11-1
(load 5.7–8). Change Request CR-BUG-260920-3ukdk4-2 rev 2, base
`7fa08e84bc85921dca450fa71e40740cc2052a6a`, candidate tree
`b22e4e3adb2f96163694603c5319828223b6191c`. Read-only review: nothing in the
Story worktree was modified (temp-index tree of the worktree before and after
the review = candidate tree; `git status` unchanged at 10 entries). All tests,
probes and mutants ran in disposable clones under `/tmp/rev3ukdk4r2/{cand,mut,probe}`.

## 1. Verdict in one paragraph

Revision 2 closes both required findings of the rev-1 verdict and the optional
N1 without touching the accepted isolation: `internal/gitops/gitops.go`,
`internal/gitops/isolated_test.go` and `cmd/curator/project_resolve.go` are
byte-identical to revision 1 (the rev1→rev2 delta is 7 files: the new
production-entry refresh row, docs/cli.md, docs/troubleshooting.md, CHANGELOG,
the docs pin, and the closure fallback + its unit row). R1: the fetch call site
`project_resolve.go:201` is now pinned at the production entry by
`TestDraftLiteralRefreshIgnoresUserConfig` (non-corpus, ratio base stays 94);
my exact M7 sed mutation fails it (exit 1), a re-clone shape mutant fails it,
and a no-op-fetch mutant fails it, so the row proves both halves (fetch happens
AND is isolated). R2: the operator doc now states the lane correctly
(credentials from the invoking environment only; user/system git configuration
never consulted; no interactive prompt), the pin was updated consistently, the
troubleshooting remedy and CHANGELOG entry exist. N1: `pinGitAliases` falls
back to `FetchIsolated`, pinned by a unit row that the ambient mutant fails
(exit 1) — and that row passes on the hosted Windows lane with real git. The
corpus row `v2-user-insteadof-ignored` passes as a driven row and the ratio
line counts it (hosted: ubuntu 89/3/1/1, macOS 90/3/1/0; no KNOWN-GAP marker);
legacy v1 golden byte-identical; resolved lane untouched; diagnostics stay
sanitized. Gate green on the exact candidate tree. ACCEPT.

## 2. Exact-tree and gate proof

| Check | Result |
|---|---|
| Story worktree tree (`GIT_INDEX_FILE=$(mktemp) git read-tree HEAD && git add -A && git write-tree`) | `b22e4e3a…` = candidate ✓ (before and after the review) |
| Patch resource sha256 | `6378a9da…f89c` = board record ✓ |
| Disposable clone `cand`: base `7fa08e84` + `git apply --index rev2.patch` + commit → `HEAD^{tree}` | `b22e4e3a…` ✓ |
| Gate commit `2c5fc327943fa6d0933a32fe3bc9571e1e7a75b4` | parent `7fa08e84`, tree `b22e4e3a…` ✓ |
| Hosted run 35488447833 (`gh run view --json headSha,conclusion`) | headSha `2c5fc327…`, `success`; Lint, Naming, Interop, Test ×3, Race ×2, Gate self-test ×3 all `success`; candidate suite + rose-air skipped (as configured) |
| rev1 tree `25b9d748…` → rev2 tree | 7 files changed; `gitops.go`, `isolated_test.go`, `project_resolve.go` NOT in the delta (byte-identical to the reviewed rev1) |

## 3. What I reran myself (`sh -c 'set -o pipefail; …'`, real exit codes; logs attached as `BUG-260920-3ukdk4_review-rev2-logs.txt`)

Candidate clone `cand` (tree `b22e4e3a…`):

- `gofmt -l` on the 7 changed Go files → empty, exit 0. `go build ./...` exit 0. `go vet ./internal/gitops/ ./internal/closure/ ./cmd/curator/ ./internal/crossconformance/` exit 0.
- `go test -p 1 ./internal/gitops/ -count=1` → ok, exit 0, 13 s.
- Precompiled `crossconformance.test` (module root resolves to `cand` via `runtime.Caller`, so the CLI is built from the candidate): `-test.run '^TestDraftLiteralRefreshIgnoresUserConfig$|^TestDraftSourcesSemanticCoverage$'` → both PASS, **exit 0**, 5 s.
- `-test.run 'TestDraftSourcesSemanticCases/v2-user-(insteadof|ssh-alias)-ignored$'` → both subtests PASS; `semantic cases: 2 driven, 0 known-gap, 0 bound, 0 skipped, 94 total`; no KNOWN-GAP marker; parent exit 1 only via the `executed != total` tally guard (expected under `-run`; the guard at `draftsources_semantic_test.go:54-58` is intact). Remaining `semanticKnownGap` callers = the 3 untouched siblings (`v2-alias-resolution`, attestation wrong-name/-context).
- `go test -p 1 ./internal/closure/ -count=1` → ok, exit 0, 22 s.
- `go test -p 1 ./cmd/curator/ -run '^(TestDraftDocsPinExamples|TestDraftTransportLegacyGolden|TestProjectResolveLegacyUntouched|TestProjectResolveDraftOffRefuses|TestWithDraftRemediationTable|TestDraftRemediationThroughCLI)$'` → 6/6 PASS, exit 0, 29 s (legacy/resolved golden byte-identical; docs pin sees the new markers).
- Probe clone `probe` (candidate + my rev1 `zz_review_probe_test.go`): `go test -p 1 ./internal/crossconformance/ -run '^TestReviewProbe'` → P1–P7 PASS (include/includeIf rewrite, hooksPath execution, XDG channel, system-only, core.sshCommand on ssh://, refresh/fetch, env-injected config), P8 logs the declared bound, exit 0, 20 s.

Accepted from attached hosted evidence on the exact tree, not rerun locally (would not fit bounded calls): full `cmd/curator` and `internal/install` packages, the full 94-row semantic matrix, the Lint job. From `test-evidence-*` `go-test.json` (run 35488447833): ratio lines ubuntu `89 driven, 3 known-gap, 1 bound, 1 skipped` (the skip is `case-alias`, unrelated), macOS `90/3/1/0`, Windows `42/2/1/49` (POSIX-shim rows skip, declared); `v2-user-insteadof-ignored` pass ubuntu+macOS, skip Windows; `TestDraftLiteralRefreshIgnoresUserConfig` pass ubuntu+macOS, skip Windows; `TestResolveDraftAliasFetchFallbackIgnoresUserConfig` **pass ×3 incl. Windows (real git)**; `TestUserConfigIgnoredByResolvedLane` pass ×3; `TestDraftDocsPinExamples` pass ×3; gitops isolation tests pass ×3.

## 4. Mutants (`mut` clone; `git checkout -- .` between mutants; restored to `b22e4e3a…`)

| Mutant | Row | Result |
|---|---|---|
| **M7** (rev1 finding) `sed 's/gitops.FetchIsolated(repoDir)/gitops.Fetch(repoDir)/' cmd/curator/project_resolve.go` | `TestDraftLiteralRefreshIgnoresUserConfig` | **KILLED**: `refresh = 1`, `repository_endpoint_unavailable … fetch of the existing checkout failed (unknown: unclassified failure)` — the ambient fetch went to evil and hit the tag clobber; stderr carries class vocabulary only (no URL/tool leak). exit 1 |
| M7b shape: fetch branch disabled (`err == nil && false`) → refresh re-clones | same row | **KILLED**: `refresh = 1` on the clone path, exit 1 — the row provably runs through the fetch branch |
| No-op `FetchIsolated` (`return nil` before fetching) | same row | **KILLED**: `refresh lock = <stale>, want advanced declared`, exit 1 — the row's positive half (a real fetch of the declared repo) is not vacuous |
| **N1** `fetch = gitops.FetchIsolated` → `gitops.Fetch` in `internal/closure/resolve.go:654` | `TestResolveDraftAliasFetchFallbackIgnoresUserConfig` | **KILLED**: `git fetch --all --tags --prune failed: From file:///…/evil … ! [rejected] v1 -> v1 (would clobber existing tag)`, exit 1 |
| rev1 M1–M4, M8 (gitops narrowing + clone call site) | unchanged files, same tests | not rerun — the mutated files are byte-identical to rev1 where they were killed |

## 5. Findings against the review note

1. Isolation mechanism — unchanged from rev1 (`isolatedGitEnv`: fresh empty file pinned as GLOBAL and SYSTEM, `GIT_CONFIG_NOSYSTEM=1`, `GIT_CONFIG_COUNT/PARAMETERS/KEY_*/VALUE_*` scrubbed case-insensitively, `GIT_ALLOW_PROTOCOL` re-pinned, `GIT_TERMINAL_PROMPT=0`, fail-closed on temp-file error, HOME-independent by construction). Re-probed P1–P7 at the production entry on rev2: all bind the declared commit; `core.sshCommand` on `ssh://` fails closed with a sanitized availability class.
2. Ambient-credential ruling — consistent with repository-transport §2 ("No interactive credential discovery occurs in a headless run") and §3/§5 ("does not import insteadOf, ProxyCommand, core.sshCommand, helpers, include files, environment overrides…"): persistent user/system configuration never consulted; per-invocation environment (`SSH_AUTH_SOCK`, `GIT_ASKPASS`, `GIT_SSH`/`GIT_SSH_COMMAND`) preserved and disclosed in docs/cli.md; the resolved lane's broker/askpass fixtures are untouched (no file under `internal/install` or `internal/buildrepo` in the diff). The environment-override residue is the filed sibling TASK-260920-3ccq6b (backlog) and is stated as a bound in results.md in my wording.
3. Corpus row — `driveV2UserInsteadOf` registered as a driven row, no gap marker, ratio line counts it; `TestDraftLiteralRefreshIgnoresUserConfig` is NOT registered (no `registerDraftSemantic`), the base stays 94 and `TestDraftSourcesSemanticCoverage` passes; `executed == total` guard intact.
4. Narrowing mutants — M7 killed by the new row (§4); the rev1 kills stand on byte-identical files.
5. Legacy v1 lane — `TestDraftTransportLegacyGolden` + `TestProjectResolveLegacyUntouched` PASS; `run` still builds `append(os.Environ(), "GIT_ALLOW_PROTOCOL=…")`; remaining ambient call sites (`closure.go:367/386/398/412/421`, `main.go:686` skills-root fetch, `envprofile/gitsource.go`) are legacy/transitive/env-profile lanes, out of this leaf by the brief. `pinGitAliases` is reachable only from `ResolveDraft` (draft lane), so N1 cannot alter legacy behaviour; production leaves `FetchRepo` nil (`project_resolve.go:140`), so the isolated fallback is the real production function when an alias root is not pre-marked. Diagnostics sanitized (M7/M7b stderr: class vocabulary only).
6. Docs — `docs/cli.md:503-509` carries the requested sentence verbatim plus the N3 Windows note; `cmd/curator/draft_diagnostics_test.go:1141-1142` pins `"credentials from the invoking environment"` and `"is never consulted"`; `docs/troubleshooting.md:471-477` remedy sits under `### repository_endpoint_unavailable`; CHANGELOG entry under `## Unreleased` → `### Fixed`. No residual "ambient Git credentials" phrase in docs/, cmd/, internal/, README.

## 6. Non-blocking notes (record only; not gating this revision)

- N6 (doc precision): the troubleshooting remedy says "a non-interactive `GIT_ASKPASS` program that prints the password"; git also routes the *username* prompt through askpass when the endpoint URL carries no username, so a program that prints only the password would answer the username prompt with it. Precise wording: "answers git's username and password prompts (or embed the username in the endpoint URL)". Fold into the next docs pass or the sibling 3ccq6b.
- N2 → TASK-260920-3ccq6b (backlog), confirmed on the board. N3 recorded in docs/cli.md and results.md. N4/N5 (env-injected config in-package + P7; `EqualFold` vs `==` POSIX-equivalent) stand as bounds.
- Windows proof for the new rows: N1 closure row passes on windows-latest with real git; the refresh row is a declared POSIX-shim skip there (consistent with every sibling transport row).

## 7. Routing

`accept_cr(BUG-260920-3ukdk4, revision=2, evidence=BUG-260920-3ukdk4_review-verdict-rev2.md)`.
No human decision is needed. Attachments: this verdict; `BUG-260920-3ukdk4_review-rev2-logs.txt` (raw logs of every command in §3–§4 with exit codes).
