# TASK-260905-30zs8t — review findings, stage (a) core, cycle 3

Subject: curator worktree `/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-a-core`,
branch `feat/agent-environments-stage-a`, head `314ae748` — one signed commit on the cycle-2 head
`ac9d0037`, 8 files, +647/−39. Nine commits on curator main `a2406dfe`, all `G`
`Ivan Oparin <oparin@me.com>`. Authority: curator-spec main `f39f4a9`.

Verdict: **CHANGES REQUESTED** — F8 and F9 are genuinely fixed and hold under attack, but the
same §9.1 gate F9 installed is bypassed on the one production path that does not carry a
`Policy`, and the F9 negative tests do not narrow.

Reviewer scratch: the worktree's `.temp/review-3/` (attack scripts, mutant tree, produced locks
and markers, range differential). Read-only on tracked files; the mutant tree is a `rsync` copy
under `.temp/`, never the worktree itself. `git status --short` in the worktree is clean.

**Change Request note.** `CR-TASK-260905-30zs8t-3` carries `repository_delta=empty` against the
curator-spec story workspace. That is correct by construction: every rework brief for this leaf
says the story workspace carries an empty delta by design and forbids writing into the control
root. The reviewable work is the curator branch above, and it is what this verdict judges.

---

## Rework verification — F8 and F9, reproduced against this head

Reproduced through the **built CLI binary** with real machine configuration files, not through
the library or the producer's tests.

### F8 — FIXED (verified)

`canonicalGit` (`internal/envprofile/envprofile.go:786`) now goes through `identity.Parse`, and
every boundary the brief named uses it. Three spellings of one repository, git `insteadOf`
keeping it offline, three isolated manager homes, `curator profile install` + `curator profile use`
(`.temp/review-3/attack.sh`, `attack2.sh`):

```
https://EXAMPLE.com/evil/pkg.git  -> lock source "example.com/evil/pkg"  lock sha256:03c40fd4eda…
git@example.com:evil/pkg.git      -> lock source "example.com/evil/pkg"  lock sha256:03c40fd4eda…
ssh://git@example.com/evil/pkg    -> lock source "example.com/evil/pkg"  lock sha256:03c40fd4eda…
marker profile.source (all three) -> "example.com/evil/pkg"
```

One canonical identity, one `lock_sha256`, one store entry. Validated against the **published
schemas** with ajv 2020 over `schemas/v1` at `f39f4a9`:

```
lock.json               vs context-lock-v1.schema.json:            VALID
.agent-environment.json vs agent-environment-marker-v1.schema.json: VALID
profiles/default/lock.json (local root)                           VALID
```

Malformed network sources are rejected at the boundary, at the root and transitively
(`.temp/review-3/attack8.sh` T2): a `requires.contexts[].git` of `https://example.com:8443/org/dep`
fails with `profile_source_invalid: network source must not contain an explicit port` and never
reaches a clone. The agreement check compares canonical identities
(`contextresolve.go` `canonicalForAgreement`), and `grep` finds no remaining raw-operand path:
`canonicalGit` is the only normalizer and its three callers (`Install`, `rootInput`,
`migrateGlobalSkills`) all take its error.

### F9 — FIXED for `Install`/`Update` (verified), bypassed on the migration path (F10)

The strength the brief demanded is met on the install path. With `audit.enabled` **false** in a
real machine config (`.temp/review-3/attack.sh` S1, S2):

```
config: {"audit":{"enabled":false,"revocations":["source:example.com/evil/pkg"]}}
curator profile install https://example.com/evil/pkg
  -> rc=1  profile_source_invalid: member context:pwned audit blocked: source example.com/evil/pkg is revoked

config: {"allowed_sources":["github.com/relux-works"]}
curator profile install https://example.com/evil/pkg
  -> rc=1  profile_source_invalid: source example.com/evil/pkg is outside the machine's allowed sources
  -> profile-repos/ never created: the refusal precedes the clone
```

Transitive members are gated too (`attack8.sh`): a dependency at `evil.example.com/org/dep` under
an allowlist of `good.example.com/org` is refused after the root clones but before the dependency
does, and the same dependency under `source:evil.example.com/org/dep` is refused as revoked with
`audit.enabled` false. `audit.Gate` really runs — a verdict record appears under
`<home>/audit/<content-hash>/`.

---

## New findings

## F10 — BLOCKING — the builtin default profile's global-skill migration reads its own weaker copy of the machine configuration, so a system-locked allowlist and system-locked revocations do not apply

`repeat-of: TASK-260905-30zs8t_review-findings-stage-a-2.md#F9` (same class: a §9.1 gate that is
present and correct but is not in force on a production path)

**Where.** `internal/envprofile/envprofile.go:934`

```go
func loadPolicyForHome(home string) Policy {
	payload, err := os.ReadFile(filepath.Join(home, "config.json")) // #nosec G304 -- manager home path
	if err != nil {
		return Policy{}
	}
	…
	if err := decoder.Decode(&raw); err != nil {
		return Policy{}
	}
```

Used at `:883`, the only policy source for `migrateGlobalSkills`, which is reached from
`ensureDefault` and therefore from **every** profile command on a machine that has no `default`
profile yet — `List` (`:172`), `Install` (`:349`), `Update` (`:491`), `Remove` (`:605`),
`EnsureDefault` (`:819`), `Use` (`switch.go:132`), `Sync` (`switch.go:220`).

**What is wrong.** The CLI derives its `Policy` from `config.Load` (`cmd/curator/profile.go:63`,
`:200`), which overlays the **system configuration** and enforces its locked keys
(`internal/config/config.go:224` `Load` → `:248` `applySystem`). `allowed_sources` and `audit` are
both in `LockableKeys` (`config.go:48-53`) — they are precisely the keys an organization locks.
`loadPolicyForHome` re-implements loading by reading one file directly, so it sees neither the
system overlay nor the `CURATOR_CONFIG` path override (`config.go:190` `UserPath`). §9.1 requires
every member new to the lock to pass the source allowlist and the strict §7 audit with
revocation; these members are new to the lock.

**Evidence — the same identity, the same shell, two commands, opposite outcomes
(`.temp/review-3/attack3.sh`):**

```
CURATOR_SYSTEM_CONFIG = {"locked":["audit","allowed_sources"],
                         "allowed_sources":["github.com/relux-works"],
                         "audit":{"enabled":true,"mode":"strict",
                                  "revocations":["source:example.com/skills/hello"]}}
<home>/config.json    = {"schema_version":1,"skills_root":"skills","projects":{}}
<home>/global/Skillfile.json declares skill "hello" at https://example.com/skills/hello tag v1.0.0

control:  curator profile install https://example.com/skills/hello
          rc=1  profile_source_invalid: source example.com/skills/hello is outside the machine's allowed sources

attack:   curator profile list
          rc=0
          default lock members: [("context","default",""), ("skill","hello","example.com/skills/hello")]
          store entry materialized: contexts/skill/hello/6854cd446f2b…
```

The revoked, non-allowlisted source is cloned, hashed, stored and pinned into the lock that
`profile use default` then links into every agent home.

**Second reproduction — the config-path override (`.temp/review-3/attack4.sh` S5).** No system
config at all; the operator's own config carries both gates but lives at
`CURATOR_CONFIG=<home>/curator.json`:

```
control:  curator profile install https://example.com/skills/hello   -> rc=1 (outside allowed sources)
attack:   curator profile list                                        -> rc=0, skill migrated
```

**Control that isolates the cause (S6).** With the identical policy written to `<home>/config.json`
the gate fires: `curator profile list` → `rc=1 profile_source_invalid: migrate global skill "hello":
… outside the machine's allowed sources`. The gate is correct; only its *policy source* is wrong.

**The doc contradicts the code.** `loadPolicyForHome`'s comment says "a malformed file fails the
migration rather than silently bypassing the gates". The function cannot fail — it has no error
return — and returns an empty policy on a decode error, which is exactly the silent bypass the
comment disclaims.

**Fix.** Do not re-parse configuration in `envprofile`. Either thread the already-loaded `Policy`
into `EnsureDefault`/`ensureDefault`/`migrateGlobalSkills` from the CLI (which already has the
overlaid `cfg`), or have `loadPolicyForHome` call `config.Load(config.UserPath(), …)` and return
an error the migration propagates. A read or parse failure must fail the migration, not empty the
policy — a failed read is not an absent policy.

**Regression cover required.** A test driving `EnsureDefault` (or `List`) on an isolated home with
`CURATOR_SYSTEM_CONFIG` set to a config that locks `allowed_sources` and `audit.revocations`
excluding/revoking a declared global skill: the migration must be refused. A second with
`CURATOR_CONFIG` pointing at a non-`config.json` file name. Plus a NARROWING mutant: make the
loader read the user file but drop the system overlay — the test must fail.

---

## F11 — MAJOR — the F9 negative tests do not narrow: three mutants that weaken the gates instead of deleting them all survive the suite

`repeat-of: TASK-260905-30zs8t_review-findings-stage-a-2.md#F9` (the finding's "Regression cover
required" named a narrowing mutant by name; it was not delivered)

**Where.** `internal/envprofile/envprofile_f8f9_test.go:123` `TestSourceAllowlistRefusesBeforeClone`,
`:144` `TestRevokedSourceIsRefused`, `:192` `TestStrictAuditCanaryPasses`; the mutant table in
`TASK-260905-30zs8t_rework-report-2.md`.

**What is wrong.** The three F9 rows of the producer's table are *delete-only* — "`gateSource`
always allows", "`strictAuditMember` result ignored", "`MCPAllowlist` not populated". Cycle 2
required a narrowing mutant and named it: "an allowlist prefix match that drops the segment
boundary, so `github.com/relux-works-evil` is admitted and the test must fail." The delivered
allowlist test uses a **different host** (`evil.example.com` against an allowlist of
`github.com/relux-works`), so it cannot distinguish segment-aware matching from any coarser rule.

**Evidence — my own mutants on an `rsync` copy of the tree (`.temp/review-3/mutant/`), each run
against `./internal/envprofile ./internal/contextresolve ./internal/audit ./internal/identity`:**

| # | Mutant (gate stays present, admits one class member) | Result |
|---|---|---|
| M-A | `gateSource` matches on the **host only**, ignoring the path — an allowlist of `github.com/relux-works` admits `github.com/evil-org` | **SURVIVES** — all four packages `ok` |
| M-B | revocation applied only when `resolved.Kind == KindContext` — a revoked **skill** or **mcp** member installs | **SURVIVES** — all four packages `ok` |
| M-C | the static canary no longer blocks in `strictAuditMember` (the `if !audit.CanaryPasses()` guard removed) | **SURVIVES** — all four packages `ok` |

For contrast, two of the producer's own rows reproduce as claimed:
`gateSource` always allows → `TestSourceAllowlistRefusesBeforeClone` FAILS;
`strictAuditMember` error ignored in `auditAndStore` → `TestRevokedSourceIsRefused` FAILS
(`revocation err = <nil>`).

**These are evidence gaps, not behavioural defects.** I confirmed the shipped code is correct for
all three classes through the CLI (`.temp/review-3/attack8.sh`): a transitive member outside the
allowlist is refused, and a revoked transitive member is refused with `audit.enabled` false. M-A is
caught nowhere at the profile call site — `internal/identity.TestMatchesPrefix` covers
`h/skills-evil` at the helper, but a caller that stops calling `identity.Allowed` is invisible to
it. M-C means the canary's *blocking role on the profile path* — the one §9.1 spells "whose
failure always blocks" — has no assertion at all; `TestStrictAuditCanaryPasses` asserts the helper
returns `true`, which is true whether or not anything consults it.

**Fix.** Three narrowing tests: an allowlist entry `example.com/org` against an operand
`example.com/org-evil/pkg` (same host, adjacent path segment); a revoked `skill` member and a
revoked `mcp` member; and a canary test that drives `strictAuditMember` (or `Install`) with the
canary forced to fail and asserts the install is refused. Then re-run each as a narrowing mutant
and put the killed test in the table.

**Related, same shape.** `TestMigratedSkillSourceIsCanonical` (`:200`) is named for the canonical
identity but asserts `member.Source == "file://"+skill` — the raw shim URL. The old
`canonicalGit` (trim space, trim one slash) passes it unchanged, so the migration boundary that
F8 was asked to fix has no test that would fail if it regressed. I verified the boundary produces
`example.com/skills/hello` myself (`attack3.sh`); the suite does not.

---

## F12 — MAJOR — a `file://` git operand is reachable from the CLI and writes a lock and a marker that fail the published schemas

`repeat-of: TASK-260905-30zs8t_review-findings-stage-a-2.md#F8` (same defect, remaining spelling)

**Where.** `internal/envprofile/envprofile.go:786` `canonicalGit` returns the trimmed raw URL when
`identity.Parse` yields an empty identity; `gitsource.go:104` `gateSource` returns `nil` for any
`file://` prefix. The rework report states this as a bound: "File:// git remotes are a test-only
shim (no network identity, schema-invalid locks/markers, allowlist bypass)."

**What is wrong.** It is not test-only. It is the plain CLI (`.temp/review-3/attack5.sh` A):

```
curator profile install file:///…/filepkg --use   -> rc=0
lock   source: file:///…/filepkg
marker source: file:///…/filepkg
lock.json               vs context-lock-v1.schema.json:             INVALID  (/members/0/source pattern)
.agent-environment.json vs agent-environment-marker-v1.schema.json: INVALID  (/profile/source pattern,
                                                                              /profile oneOf: none)
```

§1 admits three source kinds and a `git` source is "a network git source under the core §6.1
canonical identity". A `file://` remote has no canonical identity, and `context-lock-v1` has no
member shape for it: a `commit`-pinned member needs `source` matching the identity pattern, and
`state_sha256` is admitted only for `kind: context`. There is no valid lock for this operand, which
is the spec saying curator should not accept it. A bound may defer a surface; it cannot legalize a
production command that writes artifacts failing the published schemas — that is the F8 defect,
narrowed to one spelling rather than closed.

**Fix.** Reject a `file://` operand (and a `file://` requirement source) at the same boundary that
rejects a malformed network source, with `profile_source_invalid`. The new F8 tests already prove
the replacement: `insteadOf` gives hermetic, offline, network-identity fixtures. If converting the
existing `file://`-based tests is too large for this stage, then say so and gate the operand at the
**CLI** so no operator-reachable path produces an invalid artifact — but do not leave it reachable
and call it a shim.

---

## Stated bounds — judged

| Bound (rework report 2) | Judgement |
|---|---|
| MCP package allowlist has no schema-1 machine-config surface; CLI passes empty (permits all) | **Defensible.** `mcp_package_allowlist` is a manager-config **v2** key (`environments.mcp_package_allowlist`, environments §12.1 line 2227; `schemas/v1/manager-config-v2.schema.json:169`), and manager-config schema 2 is explicitly out of scope for stage (a). §2 says "an empty allowlist permits every network identity", so the stage-(a) behaviour equals the spec default. The enforcement path exists and fires (`TestMCPAllowlistRefusesOutsidePackage`, and the diagnostic now has a producer where cycle 2 found none). |
| `file://` git remotes are a test-only shim | **Not defensible as written** — see F12. Reachable from the CLI, writes schema-invalid artifacts. |
| `fetchRaw` first-spelling-wins for transport selection | **Defensible.** Every spelling clones identical bytes; transport choice is operator-visible and spec-neutral. Worth noting a transitive requirement declared `git@host:org/dep` is canonicalized before any raw is recorded and therefore clones over `https://`, which can surprise an SSH-only operator. Non-blocking. |
| F7 scoped waivers, F2 migration warning | Unchanged, correctly carried to the stage that lands the operator surfaces. |

---

## Verified correct — attacked, held

| Surface | Attack | Result |
|---|---|---|
| F8 canonical identity | 3 spellings via `insteadOf` through the CLI; ajv against `schemas/v1` | one identity, one `lock_sha256`, lock + marker VALID |
| F8 malformed source | explicit port at the root and as a transitive requirement | refused at the boundary, no clone |
| F9 allowlist | root and transitive member outside `allowed_sources` | refused before the clone; `profile-repos` empty for the root case |
| F9 revocation with `audit.enabled=false` | real config file, canonical-identity revocation, root and transitive | refused: "audit blocked: source … is revoked" |
| F1 detector scope (re-run) | secret at package root / under `--directory sub/context` / under `sub/CONTEXT.md` / transitive `requires.contexts{directory}` | all four refused `context-secret-material`; secret outside the package root correctly installs |
| F2 fresh home (re-run) | isolated home: `profile list`, `use default`, `sync`, `use --clear --env claude_code` | all rc=0 |
| F3 lock and journal (re-run) | 8 SIGKILL iterations at 0.1–30 ms during `profile use beta` | 0 broken invariants; recorded current stayed `alpha`, marker and materialized bytes always agreed, next `use` converged, backups retained `13..17` (5 generations) |
| F4 skip honesty (re-run) | `grep CURATOR_STAGE_B` over the tree; every env var named in a skip | 0 hits; all four named variables are read (`CURATOR_CONFORMANCE_ROOT` 43 sites); `stage-deferred` registered at `skip-classes.tsv:106` and `platform-cases.tsv:187` |
| F5/F6 ranges (re-run) | own 24 ranges × 20 versions differentially vs `semver@7.7.4` | 14 exact agreements; all 10 deviations are §1.4 restrictions (hyphen, `v`-in-range, build metadata, spaced operator, empty) or the F5 decision (`>*`,`<*`,`>x`,`<X`) or `latest`; 0 unexplained |
| Materialization | 13 monolithic vector cases byte-for-byte, both precedence primitives, no-chapter, zero-modules, system-prompt | PASS (`contextmaterialize`, `contextlock`, `envmarker`, `pkgversion` are untouched by this rework — `git diff ac9d0037..314ae748` is empty for all four) |
| Marker shape | ajv + inspection of a `git`-root marker | VALID; `surfaces.root-context.copies[0].reason = "claude-code-root-context"`, `mode: linked`, precedence recorded |
| CLI rows | refusal spot-check through the built binary | `profile_install_ref_conflict`, `--use` takes no name, `profile_not_found` ×3, `environment_unknown` |
| Producer mutant table | 2 of 7 rows re-run | both kill their named test as claimed |
| Scope / signatures | `git log --format='%G? %GS'` over `a2406dfe..HEAD` | 9 commits, all `G`, `Ivan Oparin <oparin@me.com>`; no unrelated behaviour change |

## Gates rerun by the reviewer at `314ae748`

| Gate | Result |
|---|---|
| `go build ./...` | exit 0 |
| `go vet ./...` | exit 0 |
| `gofmt -l cmd internal` | clean |
| `golangci-lint run` (envprofile, contextresolve, audit, identity, cmd/curator) | 0 issues |
| `go test -count=1 -race` on 11 new/touched packages | all `ok` (envprofile 21.1s) |
| interop vectors, `CURATOR_CONFORMANCE_ROOT=<spec>/conformance/v1` | 5 top-level PASS (detectors, header, monolithic, resolution, versions), 0 FAIL |
| conformance sub-skips | exactly 7: `referenced-{claude-code-composed,opencode,opencode-zero-modules}` + `mcp-{claude-code,codex-cli,opencode,pi-none}` — all genuinely stage (b) |
| `platform-case-gate.sh` on a scoped 3-package stream | all 7 skips classified `stage-deferred` / `tolerated-by-ledger`; the only failures are "required case never ran" for packages this scoped stream did not run |
| `bash .github/ci/gate-selftest.sh` | 81 passed, 0 failed (exit 0) |
| `bash .github/ci/no-broad-suppression.sh` | ok (exit 0) |
| `bash .github/ci/ledger-consistency.sh` | ok, 103 rows across linux/darwin/windows |

`go test ./cmd/curator/` (~5 min) was **not** rerun this cycle; the producer's first-hand run
(`ok … 273.524s`, exit 0) is cited, not re-verified. The suite being green while F10, F11 and F12
all hold is the point of F11: it proves the code compiles and the fixture paths work.

## AC coverage as measured

Rework-2 author decisions: **11 of 12 sub-items** delivered and independently verified. The one
that is not: F9's "read the machine configuration in the profile path" holds for `Install` and
`Update` and does not hold for `ensureDefault`/`migrateGlobalSkills` (F10).

Producer-brief items against `f39f4a9`: items 1, 2, 4, 6, 8 delivered and verified; item 3
(resolution and lock) now correct including the canonical `source` field, except for the `file://`
spelling (F12); item 5 (always-strict audit) correct on the install and update paths, bypassed on
the migration path (F10), with the negative evidence gap of F11; item 7 (linked switching and CLI)
delivered with the accepted F3 subset bound. **7 of 8 brief items fully delivered.**

## Reviewer housekeeping

Every probe ran with `HOME`, `CURATOR_CONFIG`, `CURATOR_SYSTEM_CONFIG`, `CLAUDE_CONFIG_DIR`,
`CODEX_HOME`, `XDG_CONFIG_HOME`, `GIT_CONFIG_GLOBAL` and `PI_CODING_AGENT_DIR` pointed at
directories under `.temp/review-3/`. `~/.curator` was never touched. The mutant tree is a copy
under `.temp/review-3/mutant/` and every mutation was reverted from a saved original;
`git status --short` in the worktree is clean and `git diff` against `314ae748` is empty.
Nothing was written into the control root.
