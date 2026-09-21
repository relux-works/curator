# TASK-260920-3ccq6b results — draft literal-lane env overrides + ssh-config residuals

Status: implementation complete; ALL rows observed passing on a healthy
host; mutant table executed for real (7/7 killed); READY FOR REVIEW.
(Continuation run per 3ccq6b-continue-1: prior §8 blocker closed, §3/§4
now fully observed. One test-only fix: row-(e) control, §7.)
(Revision 2 per 3ccq6b-rework-1: Windows skip-vocabulary fix, test-only;
see §10. No production change; §1/§2 rulings and bounds stand.)

## 1. Rulings (brief §R1–R4, implemented as ruled)

R1 — Allow-list, not scrub-list. `isolatedGitEnv`
(`internal/gitops/gitops.go`) builds the draft-lane git environment from
`draftAllowedEnv`: `PATH`; `HOME`/`USERPROFILE` (only so OpenSSH finds its
default `known_hosts`/identities); `TMPDIR`/`TMP`/`TEMP`; `TZ`;
`SYSTEMROOT`/`WINDIR`/`COMSPEC`/`PATHEXT`/`SYSTEMDRIVE`;
`SSH_AUTH_SOCK`; `GIT_ASKPASS`. Matching is case-insensitive
(Windows spelling). Everything else is dropped by construction
(`GIT_SSH*`, `GIT_PROXY_COMMAND`, `GIT_EXEC_PATH`, `GIT_DIR`,
`GIT_SSL_*`, `GIT_CURL_*`, `GIT_TRACE*`, `GIT_HTTP_*`, `GIT_CONFIG*`,
`GIT_TEMPLATE_DIR`, `GIT_NAMESPACE`,
`GIT_ALTERNATE_OBJECT_DIRECTORIES`, `http_proxy`/`https_proxy`/
`all_proxy`/`no_proxy` any case, `SSH_ASKPASS`, `XDG_CONFIG_HOME`,
`LOCALAPPDATA`/`APPDATA`, all unrelated names). `LOCALAPPDATA`/`APPDATA`
excluded by decision: git-for-windows needs them only for credential
helpers/caches, which this lane disables. Pins kept
(`GIT_CONFIG_GLOBAL`/`SYSTEM` → fresh empty file,
`GIT_CONFIG_NOSYSTEM=1`, `GIT_ALLOW_PROTOCOL`, `GIT_TERMINAL_PROMPT=0`)
plus `GIT_PROTOCOL_FROM_USER=0`, `LANG=C`, `LC_ALL=C` (pinned, not
honoured), and per-invocation `-c http.sslVerify=true`,
`-c http.followRedirects=false`, `-c credential.helper=` before the
subcommand: `credential.helper=` was already implied empty by the empty
user/system config for a fresh clone and `http.sslVerify=true` was
already the default, but the `-c` pins also hold against a repo-local
`.git/config` on the fetch path; `http.followRedirects=false` is pinned
so a redirect to an unlisted host never leaves the lane. Proxy config
deliberately NOT pinned via `-c`: proxies are refused by the allow-list
plus the empty user/system config (see bounds).

R2 — ssh isolation via `GIT_SSH_COMMAND` (chosen over a `GIT_SSH`
wrapper: no extra executable materialized; git-for-windows bundled
sh/ssh resolve through the honoured `PATH`; single POSIX-quoted temp
path honoured by that sh on every platform):
`ssh -F <fresh empty> -o BatchMode=yes -o StrictHostKeyChecking=yes
-o ProxyCommand=none -o ProxyJump=none -o PermitLocalCommand=no
-o ForwardAgent=no -o ClearAllForwardings=yes -o RequestTTY=no
-o CanonicalizeHostname=no -o UpdateHostKeys=no -o ConnectionAttempts=1`
— the brief's list verbatim, mirroring `ExactSSHCommand` core without
pinning identities/known_hosts (compiled-in `~/.ssh/known_hosts` +
default identity files + agent stay in force; never
`StrictHostKeyChecking=no`/`accept-new`). Operator consequences
documented in `docs/cli.md` + `docs/troubleshooting.md`
(`repository_endpoint_unavailable`): aliases/`ProxyCommand`/
`IdentityFile` not read; use agent/default key + literal host; unknown
hosts fail closed (`ssh-keyscan`/first manual `ssh`).

R3 — Legacy v1 and resolved lanes byte-identical: `run` keeps
`append(os.Environ(), …)`; zero files under `internal/buildrepo` or
`internal/install`; `TestDraftTransportLegacyGolden` /
`TestProjectResolveLegacyUntouched` green (observed this run, §3).

R4 — Askpass wording fixed to "answers git's username and password
prompts (or embed the username in the endpoint URL)" in both
`docs/cli.md` and `docs/troubleshooting.md`.

## 2. Bounds

- PATH-resolved tooling: `git` and `ssh` resolve through the honoured
  `PATH` (draft-lane tooling bound; resolved lane pins tooling).
- Proxies not honoured: proxy env dropped by the allow-list, proxy
  config blocked by the empty user/system pins; a repo-local
  `http.proxy` in a curator-owned checkout is outside the
  user-configuration threat model and is not pinned.
- Default credentials only: compiled-in `~/.ssh/known_hosts`, default
  identity files, `SSH_AUTH_SOCK` agent, `GIT_ASKPASS` (HTTPS only;
  supplies credentials, cannot redirect the endpoint).
- `HOME`/`USERPROFILE` honoured ONLY for ssh defaults; git user config
  still pinned to empty, so `HOME/.gitconfig` is never read.
- OpenSSH resolves `~/.ssh/config` from the passwd entry, NOT `$HOME`
  (proven macOS + Linux, §7): the temp-HOME row proves the `-F` empty
  mechanism + argv, not HOME redirection.
- `git://` admitted by no Skillfile/policy grammar: no `project
  resolve` production entry exists; behavioral proof at gitops layer.

## 3. Rows (a)–(g) — OBSERVED (healthy host, real exit codes)

macOS arm64, go1.26.0; `set -o pipefail`; per-test `go test -run …
-count=1 -v -timeout 240s` (120s for fast legs); EXIT is the go exit.
Times are the `--- PASS` test time (package `ok` time in parens).

| Row | Test | Entry | Verdict |
|---|---|---|---|
| (a) GIT_SSH_COMMAND evil serves evil | `TestDraftLiteralIgnoresGitSSHCommand` | cli resolve, real git+ssh | PASS 4.08s (pkg 4.608s), EXIT 0; evil-serves control live |
| (a½) GIT_SSH | `TestDraftLiteralIgnoresGitSSH` | cli resolve, real git+ssh | PASS 3.19s (pkg 3.871s), EXIT 0 |
| (b) GIT_PROXY_COMMAND | `TestDraftLiteralDropsGitProxyCommand` (cli) + `TestCloneIsolatedIgnoresGitProxyCommand` (gitops git:// behavioral) | cli + gitops, real git | PASS 3.69s (pkg 4.217s) + gitops leg PASS, EXIT 0 |
| (c) GIT_EXEC_PATH poisoned helper | `TestDraftLiteralIgnoresGitExecPath` | cli resolve, real git | PASS 7.56s (pkg 8.918s), EXIT 0; control live (first attempt SIGKILLed by stall #1, reran green) |
| (d) ~/.ssh/config alias+ProxyCommand | `TestDraftLiteralIgnoresUserSSHConfig` + `TestDraftSSHEmptyConfigIgnoresHostileFile` (real `ssh -G`) | cli resolve + real ssh | PASS 3.27s (pkg 4.000s) + 0.05s (pkg 0.756s), EXIT 0 |
| (e) proxy env → zero connections | `TestDraftLiteralIgnoresProxyEnvironment` (all-OS listener) | cli resolve, real git | PASS 5.01s (pkg 5.579s), EXIT 0; control count 1 (live), lane 0 (control fix §7) |
| (f) allow-list unit table incl. Windows case | `TestIsolatedGitEnvAllowList` + `…WindowsCaseHonoursAllowed` + pins/quote tests | unit | PASS, EXIT 0 (inside `go test ./internal/gitops/` ok 14.892s) |
| (g) SSH_AUTH_SOCK + GIT_ASKPASS pass | `TestDraftLiteralKeepsAgentAndAskpass` (resolve+refresh) + unit | cli + unit | PASS 3.96s (pkg 4.585s), EXIT 0 |
| refresh/fetch half | `TestDraftLiteralRefreshIgnoresUserConfig` (existing) | cli refresh | PASS 4.32s (pkg 5.074s), EXIT 0 (first attempt SIGKILLed by stall #2, reran green) |
| existing v2-user-insteadof | `TestDraftSourcesSemanticCases/v2-user-insteadof-ignored` | cli | subtest PASS 3.21s (parent tally FAILs by design under `-run` filter: executed(1) != total(94)) |
| affected-surface sweep | 34 subtests `/(v2-\|fallback-\|pinned-auth\|endpoint-identity-mismatch)/` | cli | ALL 34 PASS (29.67s; parent tally FAIL by design, same reason) |
| legacy R3 | `TestDraftTransportLegacyGolden` (19.92s) + `TestProjectResolveLegacyUntouched` (0.05s) | cli | PASS (pkg 20.633s), EXIT 0 |
| matrix | `TestDraftSourcesSemanticCoverage` | corpus check | PASS 0.00s (pkg 0.589s), EXIT 0; tally line shows 94 total, want 94 |

Static gates: `go build ./...` EXIT 0; `go vet` on `internal/gitops` +
`internal/crossconformance` + `cmd/curator` EXIT 0; `gofmt -l` clean;
`go test -c` for all three test packages EXIT 0.

Local rows only — no `registerDraftSemantic` (spec has no literal
ssh://-fetch corpus case; `v2-user-ssh-alias-ignored` covers the logical
planning refusal). The full 94-case matrix is not rerun here (landing
suite owns it); the no-new-driver + coverage-PASS + 94-total evidence
above preserves `executed == total`, and the 34-case sweep covers every
row touching the changed code.

## 4. Mutant table (OBSERVED — each applied, killer run, reverted)

Baseline `shasum -a 256 internal/gitops/gitops.go` =
`9426d901c30afe11bdac1b85c94a54dfc112e1baacdd0be6858a4e1d4f25c90f`
recorded before M1; hash re-verified after every revert and at the end
(identical); final `git diff --stat` equals the pre-mutant baseline
(7 modified + 1 new, §9). Every killer run below is a production-entry
row through `cli.run` with real git. KILLED = test FAILs (exit 1) with
the named assertion.

| Mutant | Killed by | Observed |
|---|---|---|
| M1 remove allow-list filter (`if false && …`) | (a) `TestDraftLiteralIgnoresGitSSHCommand` | KILLED, exit 1, 9.5s: env log carries the ambient evil (`GIT_SSH=` passthrough; pinned-command assertion prints it). (First M1 run SIGKILLed by stall #3 at 78s — no verdict; mutant reverted immediately hash-verified, re-executed after recovery.) |
| M2 drop `-F` | (d) `TestDraftLiteralIgnoresUserSSHConfig` | KILLED, exit 1, 6.67s: `ssh argv misses <-F>` |
| M3 drop `ProxyCommand=none` | (d) `TestDraftLiteralIgnoresUserSSHConfig` | KILLED, exit 1, 4.87s: `ssh argv misses <ProxyCommand=none>` |
| M4 pass ambient GIT_SSH_COMMAND (appended after pin wins) | (a) `TestDraftLiteralIgnoresGitSSHCommand` | KILLED, exit 1, 3.85s: `resolve succeeded with a hostile GIT_SSH_COMMAND` (evil honored, wrong success) |
| M5 narrow: allow `https_proxy` | (e) `TestDraftLiteralIgnoresProxyEnvironment` | KILLED, exit 1, 3.36s: `git saw the ambient proxy environment` |
| M6 narrow: drop GIT_ASKPASS | (g) `TestDraftLiteralKeepsAgentAndAskpass` | KILLED, exit 1, 4.93s: `git missed GIT_ASKPASS` |
| M7 narrow: drop SSH_AUTH_SOCK | (g) `TestDraftLiteralKeepsAgentAndAskpass` | KILLED, exit 1, 3.82s: `git missed SSH_AUTH_SOCK` |

Survivors: none. Post-mutant tree is byte-identical to the tested
state (hash + diff-stat proof above), so no re-run is needed beyond
the landing suite at handoff.

## 5. Ratio line

Semantic matrix unchanged: 94 total, 0 new drivers, no corpus edit;
`executed == total` preserved (all new rows are local `TestXxx`, like
`TestDraftLiteralRefreshIgnoresUserConfig`).

## 6. Windows proof status

POSIX rows skip on Windows exactly like sibling transport rows
(`draftTransportPATH` / explicit skip); row (e) is shell-free and runs
on all OS. Unit table proves Windows case-insensitivity on every host
(`EqualFold` + mixed-case rows, observed pass on macOS). No Windows
host available here: hosted gate must prove Windows.

## 7. Findings / anomalies

- OpenSSH ignores `$HOME` for `~/.ssh/config` (uses passwd entry):
  proven on macOS (`ssh -v` reads `/Users/administrator/.ssh/config`
  with `HOME=temp`) and Linux (alpine container: passwd-home `Port
  4444` wins over `$HOME` `Port 2222`). Row (d) therefore asserts the
  `-F` empty mechanism (real `ssh -G` hostile-vs-empty +
  `ProxyCommand=none` override) plus production-entry argv, with the
  temp-HOME hostile file planted per the brief.
- `-c` transport pins required a test-shim update (not production):
  `installDraftTransportShim` gated on `$1 == "clone"`; with `git -c
  … clone` it never logged/rewrote. Fixed to detect the `clone`
  subcommand anywhere (same for the new shim). Existing v2/refresh
  rows still pass (34-case sweep §3 all green).
- Row (e) control portability (this run, test-only fix in the NEW
  file): the positive control asserted curl stderr mentions "proxy";
  macOS curl says `Recv failure: Connection reset by peer` with no
  proxy word (observed). The verdict is now the listener connection
  count (1 = git honoured the proxy — only git learned the ephemeral
  port), wording kept as a `t.Logf`. No production change; lane
  assertions untouched. `go vet` + `gofmt` clean after the fix.
- Stall characterization (this run): `cp /bin/echo /tmp/echo-copy`
  then exec → `Killed: 9`, while `/bin/echo` runs fine — the kernel
  kills exec of newly-created executable files regardless of content;
  building/copying cannot work around it, only waiting out the window.
- `GIT_CONFIG_GLOBAL=/dev/null` in test helpers is unaffected (pins
  override).
- `docs/cli.md` keeps all pre-existing pin markers
  (`credentials from the invoking environment`, `is never consulted`,
  …) and adds 6; troubleshooting keeps `verify the network path` and
  adds 5 (`TestDraftDocsPinExamples` PASS, EXIT 0).

## 8. Host stall windows (this run; prior §8 blocker CLOSED)

Three transient macOS exec-stall windows hit this run; all passed
with bounded retries (probe `/tmp/health-3ccq6b.sh`, fresh Mach-O
`/tmp/machotest/probe`; `date -u` timestamps):

- Window #1 ~17:33:30–17:40:00Z (~6.5 min): fresh-Mach-O/shebang exec
  hung, then SIGKILLed (`signal: killed` after ~85s); existing
  binaries (`git`, `go build/vet`, `gofmt`, `grep`, `/bin/sh
  <script>`) kept working. Killed the first row-(c) attempt; row (c)
  reran PASS 7.56s after recovery (probe `healthy`, exit 0,
  17:40:00Z).
- Window #2 ~17:44–17:59:51Z (~16 min): same signature (hang →
  SIGKILL-fast). Killed the first refresh attempt; refresh reran PASS
  4.32s after recovery (probe `healthy`, exit 0, 17:59:51Z).
- Window #3 ~18:03–18:20:34Z (~17 min): SIGKILL-fast throughout.
  Killed the first M1 run at 78s (no verdict taken); M1 was reverted
  immediately (hash-verified) and re-executed after recovery
  (KILLED, §4). Probe `healthy`, exit 0, 18:20:34Z.

Healthy-host evidence: every §3/§4 verdict above with real exit codes
and timings; no verdict in this document was taken from a killed or
cached run (killed attempts were rerun; `go vet`/build/compile-only
checks are labelled as such).

## 9. Files

- `internal/gitops/gitops.go` — allow-list, ssh command, `-c` pins,
  isolated env (R1+R2); `run` unchanged (R3).
- `internal/gitops/isolated_test.go` — updated pins test + (f)/(g)
  unit + git:// row (b) behavioral + quote/args pins.
- `internal/crossconformance/draftsources_literal_isolation_test.go`
  — NEW: rows (a)–(e),(g) + real-ssh `-G` leg (local, non-corpus);
  this run: row-(e) control count-first fix (test-only).
- `internal/crossconformance/draftsources_semantic_transport_test.go`
  — shim `clone`-anywhere fix for `-c` prefix.
- `docs/cli.md`, `docs/troubleshooting.md`, `CHANGELOG.md`,
  `cmd/curator/draft_diagnostics_test.go` — R4 + allow-list/ssh/proxy
  wording + 11 new pin markers.

## 10. Revision 2 (3ccq6b-rework-1; test-only, no production change)

Revision 1 gate FAILED (run 35529945079 on gate commit d0b9e056):

1. MINE — Windows platform-case gate (tier 2): `FAIL skip with an
   unrecognised reason on windows: internal/gitops ::
   TestCloneIsolatedIgnoresGitProxyCommand`, reason `evil proxy fixture
   is POSIX-only`. Fixed below.
2. NOT MINE — Race (macos-latest):
   `TestDraftSourcesSemanticCases/attestation-evidence-wrong-key`
   (`fresh install errors = [every trusted audit registry served a
   tampered snapshot]`, want `is not audited by any trusted registry`)
   — the known race-lane nondeterminism BUG-260920-2d9gfv (next leaf of
   this Story). Not chased; the republish reruns the suite. Recorded
   here as an unrelated flake with run id 35529945079.

Fix (3 lines, 2 files, zero production bytes; no ledger or vocabulary
widening — `.github/ci/skip-classes.tsv` and
`.github/ci/platform-cases.tsv` untouched):

- `internal/gitops/isolated_test.go:552` (row-b git:// leg):
  `evil proxy fixture is POSIX-only` → `test transport wrapper is
  POSIX-only` — the exact reason `TestDraftLiteralRefreshIgnoresUserConfig`
  and the 7 sibling skips in the new file print. Ledger match:
  `.github/ci/skip-classes.tsv` line 60, class `platform-control`,
  policy `allow` (`the fixture uses a POSIX shell wrapper; native
  admission remains covered on Windows`).
- `internal/crossconformance/draftsources_literal_isolation_test.go:435`
  (row-d `ssh -G` leg): `t.Skip("ssh is not available")` →
  `t.Fatal("ssh is not available")`. A latent unclassified skip: no
  vocabulary row matches it. Sibling precedent `v2FetchTool`
  (`draftsources_semantic_v2_test.go:612`) fatals on a missing required
  tool; ssh ships on every CI runner (ubuntu/macos/windows-latest —
  rev1's Windows run passed this test), so a missing ssh is a broken
  test environment, not a platform carve-out.
- Same file `:465` (row-e listener fallback): `cannot listen for the
  proxy row: %v` → `this host cannot create a loopback listener the
  proxy row needs: %v`. Ledger match: `skip-classes.tsv` line 71 regex
  `this host cannot create`, class `host-capability`, policy `allow`
  (precedent: `cmd/curator/toolchain_remedy_test.go:41`).

Classification proof (this run): replicated the gate's `classify()`
verbatim in awk against the committed `skip-classes.tsv`: new reason 1
→ `MATCH class=platform-control policy=allow`; new listener reason →
`MATCH class=host-capability policy=allow`; both old reasons →
`UNCLASSIFIED` (confirms the diagnosis and that neither old reason
could ever pass tier 2). All other skips in the 8 touched paths are
pre-existing or already use the sibling `test transport wrapper is
POSIX-only` reason (audited: `draft_diagnostics_test.go`,
`draftsources_semantic_transport_test.go` diffs contain no skip).

Rev2 observed runs (macOS arm64, healthy host, no stalls this run;
`set -o pipefail`):

- `go vet ./internal/gitops/ ./internal/crossconformance/
  ./cmd/curator/` EXIT 0; `gofmt -l` on both edited files clean.
- `go test ./internal/gitops/ -count=1` → ok 11.295s, EXIT 0.
- `go test ./internal/crossconformance/ -run
  'TestDraftLiteral|TestDraftSSH' -count=1 -v` → 9/9 PASS, ok 13.806s,
  EXIT 0: GitSSHCommand 3.54s, GitSSH 1.05s, DropsGitProxyCommand
  1.02s, GitExecPath 0.83s, UserSSHConfig 0.81s, SSHEmptyConfig 0.05s,
  ProxyEnvironment 2.64s, KeepsAgentAndAskpass 1.29s, Refresh 2.06s.
- `go test ./cmd/curator/ -run TestDraftDocsPinExamples -count=1` →
  PASS, EXIT 0.
- `TestDraftSourcesSemanticCoverage` → ok 0.473s, EXIT 0
  (94-total tally intact); `TestProjectResolveLegacyUntouched` →
  PASS 0.04s; `TestDraftTransportLegacyGolden` → PASS 19.20s, EXIT 0.

Mutant table (§4) stands without re-execution: production
`internal/gitops/gitops.go` is byte-identical to the rev1 tested state
(the delta is 3 test-only lines), and every §4 killer row re-passed
above on those identical bytes. Tree: same 8 paths as §9 (7 modified +
1 new).
