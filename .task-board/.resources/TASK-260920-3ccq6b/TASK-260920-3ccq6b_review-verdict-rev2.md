# TASK-260920-3ccq6b — review verdict, revision 2 (CR-TASK-260920-3ccq6b-2)

Reviewer: Relux Bot (claude-opus-5, max), 2026-09-21. Read-only review; every rerun and
mutant ran in disposable clones (`/tmp/rv-3ccq6b`, `/tmp/rv-3ccq6b-mut`), never in the
Story worktree. Shell: zsh tool calls driving `#!/bin/bash` scripts with `set -o pipefail`;
every exit code below is the real `go test` exit.

## 1. Verdict

**ACCEPT** — `accept_cr(TASK-260920-3ccq6b, revision=2, evidence=TASK-260920-3ccq6b_review-verdict-rev2.md)`.

The AC is met at the production entry: the draft literal-URL lane's git environment is an
explicit allow-list (R1), the ssh command is curator-owned with `-F <fresh empty file>` and the
ruled option list (R2), legacy v1 and resolved lanes are byte-identical (R3), the askpass
wording is fixed (R4); rows (a)–(g) run through `cli.run` `project resolve`/`refresh` with real
git, the brief's four mutants plus the producer's three narrowing mutants and five of my own are
KILLED; docs/CHANGELOG/pins in place; hosted gate green on all lanes with the Windows skip
reasons inside the ledger vocabulary. Residuals in §7 are non-blocking bounds, not defects.

## 2. Tree identity and gate

- Story worktree temp-index `write-tree` = `06956f334c373984e8966f510da143cbd168b46f` = CR
  candidate tree OID (untracked `draftsources_literal_isolation_test.go` included). 8 changed
  paths as listed in the CR; none under `internal/buildrepo` or `internal/install`; `.github/`
  untouched (ledger vocabulary not widened).
- Hosted gate run 35534261038: `conclusion=success`, gate commit `607b077a` has parent
  `680a2f47` (base OID) and tree `06956f33…` (= candidate). Jobs: Lint, Naming gate, Test
  (ubuntu/macos/windows-latest), Race (ubuntu/macos), Gate self-test (×3), Interop conformance —
  all success; rose-air and candidate suite skipped by workflow design.
- Both disposable clones were checked out at `607b077a` (`HEAD^{tree}` = candidate) and ended
  with a clean status at that tree after all reruns/mutant restores (hash-verified per mutant,
  baseline `internal/gitops/gitops.go` sha256 prefix `9426d901c30afe11` = the producer's recorded
  baseline).

## 3. Reruns (clone 1, healthy host, 21:20–21:22Z)

| Command | Result |
|---|---|
| `go build ./...` / `go vet` (gitops, crossconformance, cmd/curator) / `gofmt -l` | rc 0 / rc 0 / clean |
| `go test ./internal/gitops/ -count=1 -v` | ok 12.0 s, 34/34 PASS incl. `TestIsolatedGitEnvAllowList`, `…WindowsCaseHonoursAllowed`, `…SSHCommand`, `TestIsolatedGitConfigArgs`, `TestShellQuoteEscapesSingleQuote`, `TestCloneIsolatedIgnoresGitProxyCommand`, `TestCloneIsolatedIgnoresUserConfig` (3 subtests), `TestFetchIsolatedIgnoresUserConfig` |
| `go test ./internal/crossconformance/ -run 'TestDraftLiteral\|TestDraftSSH' -count=1 -v` | ok 14.7 s, 9/9 PASS: rows (a) `IgnoresGitSSHCommand` 3.47 s, (a½) `IgnoresGitSSH`, (b) `DropsGitProxyCommand`, (c) `IgnoresGitExecPath`, (d) `IgnoresUserSSHConfig` + `TestDraftSSHEmptyConfigIgnoresHostileFile`, (e) `IgnoresProxyEnvironment` 2.57 s, (g) `KeepsAgentAndAskpass`, refresh `TestDraftLiteralRefreshIgnoresUserConfig` |
| `TestDraftSourcesSemanticCoverage$` | PASS (94 total) |
| `TestDraftSourcesSemanticCases$/(v2-\|fallback-\|pinned-auth\|endpoint-identity-mismatch)` (every row that uses the changed shim) | 34/34 subtests PASS; parent FAIL is the by-design `executed(34) != total(94)` guard under `-run` |
| `cmd/curator -run 'TestDraftDocsPinExamples$\|TestProjectResolveLegacyUntouched$\|TestDraftTransportLegacyGolden$'` | ok 21.1 s, 3/3 PASS (golden 20.6 s) |

Hosted full matrix (test-evidence artifacts of run 35534261038): ubuntu `semantic cases: 92
driven, 0 known-gap, 1 bound, 1 skipped, 94 total`, `TestDraftSourcesSemanticCases` PASS 86 s;
windows `42 driven … 51 skipped, 94 total` PASS. Ratio unchanged from the previous leaf: no
corpus row added (correct — the spec has no literal-ssh fetch case; `v2-user-ssh-alias-ignored`
is the logical-declaration refusal), `executed == total` preserved.

## 4. Findings against the review note

1. **Allow-list (R1)** — `draftAllowedEnv` (`internal/gitops/gitops.go`) is exactly the brief's
   list: `PATH`, `HOME`, `USERPROFILE`, `TMPDIR`, `TMP`, `TEMP`, `TZ`, `SYSTEMROOT`, `WINDIR`,
   `COMSPEC`, `PATHEXT`, `SYSTEMDRIVE`, `SSH_AUTH_SOCK`, `GIT_ASKPASS`; `LOCALAPPDATA`/`APPDATA`
   excluded with the stated reason (credential helpers/caches only, disabled on this lane);
   `LANG=C`/`LC_ALL=C` pinned; `GIT_CONFIG_GLOBAL`/`SYSTEM` → fresh empty file,
   `GIT_CONFIG_NOSYSTEM=1`, `GIT_ALLOW_PROTOCOL`, `GIT_TERMINAL_PROMPT=0`,
   `GIT_PROTOCOL_FROM_USER=0`, `GIT_SSH_COMMAND=<curator-owned>` appended after the filter.
   `isDraftAllowedEnv` is `strings.EqualFold` (Windows spelling); `TestIsolatedGitEnvAllowList`
   sets 26 hostile `GIT_*` names, three mixed-case spellings, 8 proxy spellings, `SSH_ASKPASS`,
   `XDG_CONFIG_HOME`, `LOCALAPPDATA`/`APPDATA` and asserts each absent and each listed name
   present once; `…WindowsCaseHonoursAllowed` proves the honour direction; both PASS on
   windows-latest in the gate evidence (real case-insensitive environment). `HOME` honoured with
   git user config still pinned: `TestCloneIsolatedIgnoresUserConfig/home file without selectors`
   plants `HOME/.gitconfig` insteadOf → declared commit bound (PASS here and on Windows).
   Proxies any case: unit table + row (e). `GIT_ASKPASS`/`SSH_AUTH_SOCK`: unit + row (g).
2. **Production-entry rows** — all through `runCurator` → compiled CLI `project resolve`/
   `refresh` (`cli.run` → `fetchDraftRepoAllowingFallback` → `gitops.CloneIsolated`/
   `FetchIsolated`, `cmd/curator/project_resolve.go:201,229`) with real git behind a logging
   PATH shim; each negative row has a live positive control (evil ssh serves the evil commit;
   poisoned `git-remote-https` invoked; proxy listener accepts the control connection) and asserts
   the `repository_endpoint_unavailable` class, sanitized stderr (no URL, tool path, evil host),
   exactly one clone of the declared literal endpoint, no lock published.
   Row (d) judgement: the committed row proves the lane's ssh argv (`-F`, `BatchMode=yes`,
   `StrictHostKeyChecking=yes`, `ProxyCommand=none`, `ProxyJump=none`, no `no`/`accept-new`)
   at the production entry with a fake `ssh`, and `TestDraftSSHEmptyConfigIgnoresHostileFile`
   proves the `-F` mechanism with real `ssh -G`. Because OpenSSH resolves `~/.ssh/config` from
   the passwd entry, the temp-HOME plant cannot exercise the real file, so I closed the gap with
   my own production-entry probe (§5, probe A): behind the lane, REAL OpenSSH 9.9p2 read exactly
   one configuration file — the fresh `curator-sshconfig-*` lane file — and neither this host's
   existing `/Users/administrator/.ssh/config` (read by default `ssh`, control shown), nor its
   `Include`d `~/.colima/ssh_config`, nor `/etc/ssh/ssh_config`; it connected to `fixture.test`
   literally and executed no proxy command. "ssh config isolation implemented" is therefore
   established at the production entry, not only at the unit layer.
3. **`-c` pins and shim change** — `isolatedGitConfigArgs` = `-c http.sslVerify=true -c
   http.followRedirects=false -c credential.helper=` prepended in `CloneIsolated` and
   `FetchIsolated` only (`run` untouched). Probe B (§5) observed them at the production entry
   on clone, fetch and `remote` (1/1/1 pinned, 0 unpinned network invocations). The
   `installDraftTransportShim` change (`clone` detected anywhere in argv instead of `$1`) is
   required by the `-c` prefix and cannot widen: an argument equal to the bare word `clone` never
   occurs in the lane's fetch/remote/rev-parse invocations, and the 34 transport/fallback/pinned
   rows plus the refresh row re-passed on it.
4. **Mutants** — see §6: the brief's four (M1–M4), the producer's narrowing three (M5–M7) and my
   own five (EqualFold→exact, `accept-new`, clone call-site→legacy, fetch call-site→legacy, `-c`
   pins dropped at the call sites) — outcomes observed, not predicted.
5. **R3** — `git diff --stat base..candidate` names no file under `internal/buildrepo` or
   `internal/install`; `run` still `runWithEnv(dir, append(os.Environ(), "GIT_ALLOW_PROTOCOL="+…))`;
   `TestDraftTransportLegacyGolden` (20.6 s) and `TestProjectResolveLegacyUntouched` PASS.
   `internal/closure/resolve.go:654` still routes the alias-fallback fetch to `FetchIsolated`.
6. **Docs** — `docs/cli.md:503–527` carries the allow-list statement, the proxy bound ("does not
   honour proxy environment"), the ssh command and operator consequences (aliases/ProxyCommand/
   IdentityFile not read; agent or default-named key + literal host; `ssh-keyscan`), the PATH
   tooling bound, and the N6 askpass wording; `docs/troubleshooting.md:470–487` under
   `### repository_endpoint_unavailable` (the class the operator sees) with the same consequences
   and "proxy environment is not honoured on this lane"; `TestDraftDocsPinExamples` pins 6 new
   cli markers and 5 troubleshooting markers (PASS); CHANGELOG under `## Unreleased → ### Fixed`.
   No stale "GIT_SSH honoured" wording remains in docs/README. Windows skip reasons: the eight new
   skips classify `platform-control / allowed-platform-control / test transport wrapper is
   POSIX-only` in the windows `skips-observed.tsv`; the listener fallback uses the ledger's
   `this host cannot create` (`host-capability`); `.github/ci/*.tsv` untouched; Gate self-test
   (windows-latest) success.

## 5. Reviewer probes (throwaway `zz_review_probe_test.go` in clone 2, not part of the candidate)

- **Probe A — real ssh at the production entry (row d)**: PATH `ssh` wrapper = `exec <real ssh>
  -v "$@" 2>>log`; hostile `Host fixture.test → HostName evil.test + ProxyCommand` planted under
  the child HOME; `curator project resolve` on `ssh://git@fixture.test/kit.git`. Result: exit ≠ 0
  with `repository_endpoint_unavailable`, one clone of the declared endpoint, and the verbose log
  shows `Reading configuration data /var/folders/…/T/curator-sshconfig-2379978387` as the ONLY
  configuration read, then `Connecting to fixture.test port 22.` → DNS failure; no
  `/Users/administrator/.ssh/config`, no `/etc/ssh/ssh_config`, no `evil.test`, no proxy
  execution. Control on the same host: default `ssh -v -G fixture.test` reads
  `/Users/administrator/.ssh/config`, `/Users/administrator/.colima/ssh_config`,
  `/etc/ssh/ssh_config`; `ssh -v -F <empty> -G` reads only the given file. PASS 4.9 s.
- **Probe B — `-c` pins reach git**: argv-logging git shim (rewrite to the bare fixture);
  resolve + refresh succeed; argv lines: `<-c><http.sslVerify=true><-c><http.followRedirects=false>
  <-c><credential.helper=><clone>…` ×1, `…<fetch>…` ×1, `…<remote>` ×1, unpinned clone/fetch 0.
  PASS 1.6 s.

## 6. Mutant table (clone 2; applied with perl, hash-verified, restored with `git checkout --`,
final tree = candidate)

| # | Mutant (file) | Killer row(s) | Observed |
|---|---|---|---|
| M1 | remove the allow-list filter (`if false && !isDraftAllowedEnv(name)`) | (a) `TestDraftLiteralIgnoresGitSSHCommand`, (b) `TestDraftLiteralDropsGitProxyCommand`, (c) `TestDraftLiteralIgnoresGitExecPath`, (e) `TestDraftLiteralIgnoresProxyEnvironment`; unit `TestIsolatedGitEnvAllowList` | **KILLED** ×5, rc 1: "git saw the ambient GIT_SSH_COMMAND" (env log carries the evil `GIT_SSH=` passthrough), "…GIT_PROXY_COMMAND", "…GIT_EXEC_PATH", "…proxy environment"; unit `GIT_SSH = ["evil-GIT_SSH"], want it absent` |
| M2 | drop `-F <empty>` from `draftSSHCommand` | (d) `TestDraftLiteralIgnoresUserSSHConfig`; unit `TestIsolatedGitEnvSSHCommand` | **KILLED** ×2: `ssh argv misses <-F>`; `GIT_SSH_COMMAND misses "ssh -F "` |
| M3 | drop `ProxyCommand=none` | (d); unit `TestIsolatedGitEnvSSHCommand` | **KILLED** ×2 (re-executed after the stall): `ssh argv misses <ProxyCommand=none>`; unit misses the option |
| M4 | pass the ambient `GIT_SSH_COMMAND` through after the pin (ambient wins) | (a); unit `TestIsolatedGitEnvAllowList`, `TestIsolatedGitEnvSSHCommand` | **KILLED** ×3: `resolve succeeded with a hostile GIT_SSH_COMMAND, want the endpoint refusal` (the evil ssh served the evil repo); unit sees two `GIT_SSH_COMMAND` entries / the ambient payload |
| M5 | narrow: add `https_proxy` to the allow-list | (e); unit | **KILLED** ×2: `git saw the ambient proxy environment` (listener took the lane's connection); unit `https_proxy = [...], want it absent` |
| M6 | narrow: drop `GIT_ASKPASS` | (g) `TestDraftLiteralKeepsAgentAndAskpass`; unit ×2 | **KILLED** ×3: `git missed GIT_ASKPASS`; `GIT_ASKPASS was not preserved` |
| M7 | narrow: drop `SSH_AUTH_SOCK` | (g); unit ×2 | **KILLED** ×3: `git missed SSH_AUTH_SOCK`; `SSH_AUTH_SOCK was not preserved` |
| M9 (mine) | `strings.EqualFold` → `==` in `isDraftAllowedEnv` (Windows case claim) | unit `TestIsolatedGitEnvWindowsCaseHonoursAllowed` | **KILLED**: mixed-case `sSh_AuTh_SoCk` value dropped (`TestIsolatedGitEnvAllowList` PASS — the honour-direction row is the one that narrows) |
| M11 (mine) | `StrictHostKeyChecking=yes` → `accept-new` | (d); unit `TestIsolatedGitEnvSSHCommand` | **KILLED** ×2: `ssh argv misses <StrictHostKeyChecking=yes>`; unit misses `StrictHostKeyChecking=yes` |
| M13 (mine, call site) | `cmd/curator/project_resolve.go:229` `CloneIsolated` → legacy `Clone` | (a), (c), (e) | **KILLED** ×3: resolve succeeded with the hostile `GIT_SSH_COMMAND`; git saw `GIT_EXEC_PATH`; git saw the proxy environment |
| M14 (mine, call site) | `project_resolve.go:201` `FetchIsolated` → legacy `Fetch` | `TestDraftLiteralRefreshIgnoresUserConfig` (3ukdk4 row, on the changed shim) | **KILLED**: `refresh = 1`, `repository_endpoint_unavailable … fetch of the existing checkout failed (unknown: unclassified failure)` — sanitized class vocabulary only |
| M15 (mine) | drop `withIsolatedConfigArgs` at both call sites (`-c` pins gone) | every committed gitops + crossconformance row; my probe B | **SURVIVES the committed suite** (gitops 9/9 PASS, rows 9/9 PASS); **KILLED only by probe B**: `pinned clone=0 fetch=0 … unpinned clone/fetch=2` → residual R-A |

11 of 11 committed-row mutants killed; the brief's four (M1–M4) and the producer's M5–M7
reproduced with the same assertions the producer recorded. Restore verified per mutant
(`gitops.go` back to `9426d901c30afe11`, `project_resolve.go` back to `01d6214729acae97`);
clone 2 ends at tree `06956f33…` with only my untracked probe file.

## 7. Bounds and non-blocking residuals (record only)

- R-A `-c` pins have no committed production-entry row: `TestIsolatedGitConfigArgs` pins the
  helper, but a call-site drop (my M15) is not caught by any committed row — probe B is the row
  shape that kills it (argv-logging shim asserting the three `-c` pairs before `clone`/`fetch`).
  The pins are demonstrably live in the candidate (probe B); this is a regression-protection gap
  for a belt-and-braces pin, not a behaviour defect. Suggested for the next test pass of this
  Story.
- R-B Windows `ssh://` draft clones are unverified: every ssh row is a declared POSIX-shim skip on
  windows-latest (as for all sibling transport rows); what the gate does prove on Windows is the
  allow-list environment with real git-for-windows (`TestCloneIsolatedIgnoresUserConfig`,
  `TestFetchIsolatedIgnoresUserConfig`, row (e) `TestDraftLiteralIgnoresProxyEnvironment` PASS)
  and `ssh -F <Windows temp path>` via direct exec (`TestDraftSSHEmptyConfigIgnoresHostileFile`
  PASS). The single-quoted `GIT_SSH_COMMAND` through git-for-windows' bundled `sh` is reasoned,
  not observed — report as unverified, not passing.
- R-C Post-transport local object reads on the draft lane (`gitops.Resolve`/`Extract` →
  `run`, ambient environment) are outside this leaf (brief scope: the isolated clone/fetch
  functions and their call sites) and identical to the legacy lane; `GIT_DIR`-class ambient
  overrides would affect them as they always have. Not a transport-contract item; noted for
  completeness.
- R-D The draft ssh options omit the resolved lane's `ConnectTimeout`/`-T`/explicit auth-method
  toggles by the brief's verbatim list; under `-F <empty>` + `BatchMode=yes` the omitted ones
  are OpenSSH defaults except `ConnectTimeout` (OS TCP timeout applies to an unreachable host).
  By ruling; no change requested.
- R-E `GIT_PROTOCOL_FROM_USER=0` is inert next to `GIT_ALLOW_PROTOCOL` (listed protocols are
  `always`); harmless and requested by the brief.
- R-F Row (b) has no `git://` production entry by construction: `cutRevision2Scheme`/
  `buildrepo.ParseSource` admit only `https://`, `ssh://` and scp-like forms, so the behavioural
  `GIT_PROXY_COMMAND` proof at the gitops layer (`TestCloneIsolatedIgnoresGitProxyCommand`, evil
  proxy never invoked, fail-closed) plus the production-entry env-drop row is the right shape.
- Rev1 gate race-lane failure `attestation-evidence-wrong-key` is BUG-260920-2d9gfv, unrelated.

## 8. Host incident during this review

Two exec-stall windows (fresh `go test` binaries `signal: killed` at ~88 s or immediately)
hit the mutant driver: ≈21:26–21:36Z (M3–M6 first attempts killed; M7 onward ran) and
≈21:41–21:59Z (a fresh Go probe binary stranded 85 s then `Killed: 9`; load spiked to 211 when
the backlog released). Killed runs were never taken as verdicts: M3–M6 were re-executed by a
health-gated retry driver (fresh Go binary `println` probe; note that a *copied platform
binary* such as `cp /bin/echo` is SIGKILLed on this host even when healthy, so it is not a
valid probe) and §6 records the observed outcome of the successful attempt only (M5's row (e)
took 314 s at load ≈200 but produced a real assertion failure). Reruns in §3 and probes in §5
completed before the first window.

## 9. Routing

`accept_cr(TASK-260920-3ccq6b, revision=2, evidence=TASK-260920-3ccq6b_review-verdict-rev2.md)`;
no human decision needed. Attachments: this verdict; `TASK-260920-3ccq6b_review-rev2-logs.txt`
(raw driver logs, per-test outputs and the probe test source).
