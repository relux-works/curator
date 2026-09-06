# TASK-260905-30zs8t — review findings, stage (a) core, cycle 2

Subject: curator worktree `/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-a-core`,
branch `feat/agent-environments-stage-a`, head `ac9d0037`. The branch was rebased onto curator
main `a2406dfe` (which carries the byte-exact acquisition, PR #58) since the cycle-1 review, so
the reviewable delta is `git diff 4e6bc403^..ac9d0037` — 33 files, +8809, 8 signed commits.
Authority: curator-spec main `f39f4a9`.

Verdict: **CHANGES REQUESTED** — F1–F6 are all fixed and held under attack, but two new
**blocking** findings were found by attacking gates cycle 1 did not reach.

Reviewer scratch: the worktree's `.temp/review-2/` (harnesses, differential streams, produced
locks and markers). Read-only on tracked files.

---

## Rework verification — F1–F6, reproduced against this head

Every item below was re-attacked through the production entry points, not read from the report.

### F1 — FIXED (verified)

`packageRoot` (`internal/envprofile/gitsource.go:290-297`) is the single join and every consumer
of a resolved member uses it: `envprofile.go:527` (Update new-member pre-check), `:604` / `:617` /
`:625` (`auditAndStore` detect, manifest load, MCP command check) and `switch.go:305`
(materialization load). Five attacks through `envprofile.Install`, byte-identical AKIA payloads:

```
control 0: root package, secret at root                    REFUSED: profile_source_invalid: member context:acme carries a blocking context-secret-material finding
attack 1: --directory sub, secret under sub/               REFUSED: ... member context:acme ...
attack 1b: --directory sub, secret in sub/CONTEXT.md       REFUSED: ... member context:acme ...
attack 2: transitive requires.contexts{directory:sub}      REFUSED: ... member context:dep ...
attack 3: transitive requires.mcp{directory:sub} in args   REFUSED: ... member mcp:mdep ...
control 1: --directory sub, secret OUTSIDE the package root  INSTALLED (correct: outside scope)
```

Harness `.temp/review-2/auditattack/main.go`. The cycle-1 bypass is closed for both shapes the
reviewer used, plus the MCP-args shape the brief did not name.

**Third `Detect` site, checked and cleared.** `envprofile.go:784` still passes bare `entry`
(migrated global skills). That is not a drift: `context-lock-v1#/$defs/member` forbids
`directory` on a `skill` member (`allOf[0]`), and `manifest.Decl` carries no directory, so no
skill member can address a subdirectory. Worth a comment, not a change.

### F2 — FIXED (verified)

`.temp/review-2/defaultattack/main.go`, three isolated fresh manager homes:

| machine state | `List` | `Use default` | `Sync` | `Use --clear --env claude_code` | default lock |
|---|---|---|---|---|---|
| no global skills | ok | ok (4/4 adapters) | ok (4) | ok (1) | 1 member, state-pinned, stable across calls |
| one tag-pinned git global skill | ok | ok | ok | ok | 2 members; `skill hello` with `commit=8fad227e8b59` and its source |
| branch-pinned global skill | ok | ok | ok | ok | 1 member — the skill is silently omitted |

The store entry the lock pins exists (`contexts/context/default/45548456…/agent-context.json`)
and the migrated skill's entry is materialized (`contexts/skill/hello/8fad227e…/hello/SKILL.md`).
The four `cli/curator.md` rows that were dead on a fresh machine at cycle 1 all work.
No root-context file is written for `default`, as §9.4 requires.

The branch-pinned/local omission is the producer's stated bound and it is correct on the merits:
`context-lock-v1` `oneOf` admits `state_sha256` only for `kind: context`, so a skill member must
carry `source` + `commit` and a branch pin has none. It is silent, though — no warning tells the
operator a global skill did not migrate. Non-blocking; worth a warning in the stage that wires
live declarations.

### F3 — FIXED with a defensible stated bound (verified)

Every mutating entry point takes the lock and threads one operation:
`List` :167, `Install` :326, `Update` :455, `Remove` :565, `EnsureDefault` :701, `Use`
(`switch.go:132`), `Sync` (`switch.go:220`) all call `beginOperation`, which acquires
`managerlock.AcquireHomeOnly` and runs `transaction.Engine.Recover` before anything mutates
(`internal/envprofile/lock.go:37-57`). Manager-home records publish through one
`transaction.Plan` (`lock.go:68-116`).

**Crash attack** (`.temp/review-2/crashattack/`): 28 iterations, a child process running
`Use(beta)` SIGKILLed at a random offset — 14 in 200µs–45ms and 14 in 30µs–1.4ms, so the kill
lands on both sides of materialization:

```
it=0  kill@504µs    current=alpha  doc=alpha   marker=alpha  journals(before/after)=0/0  ok
it=2  kill@1.35ms   current=alpha  doc=beta    marker=beta   journals(before/after)=0/0  ok
...
iterations with a broken invariant: 0   (both sweeps)
```

In every iteration: the recorded current survived as `alpha`, the marker was valid JSON naming
exactly the profile whose bytes were on disk, no journal residue survived the next operation, and
the next `profile use beta` converged (current and CLAUDE.md both `beta`). A SIGKILL while the
manager-home lock is held does not wedge the home — the next `List` proceeded.

The stated subset bound (per-entry agent-home payloads are not transaction targets) is reasoned
in `switch.go:15-23` and in the report, and its reasons hold: one `Plan` is all-or-nothing with
rollback while §9.2 requires every entry attempted with partial success persisted, and §8.3
requires versioned backup generations the engine's sidecars do not provide. §9.2's observable
contract — "`profile use` of either profile … completes the scope" — is met by lock-driven
convergence, demonstrated above. Accepted as a bound.

**Backups and retention, checked (neither cycle had):** nine alternating switches leave exactly
generations `4 5 6 7 8` under `.agent-environment-backup/` with the correct prior bytes in each,
retention 5 (§8.3 default). An unmanaged pre-existing `CLAUDE.md` is refused with
`environment_surface_unmanaged_conflict` and its bytes survive (`.temp/review-2/switchattack/`).

### F4 — FIXED (verified)

`grep -rn CURATOR_STAGE_B` over the tree: zero hits outside `.task-board` resources and
`.temp/`. The skip now reads "deferred to stage (b): the referenced form and MCP channel files
land in stage (b)" and `.github/ci/skip-classes.tsv:106` registers `stage-deferred` with a
matching pattern. `.github/ci/platform-cases.tsv:187` carries the truthful reason.
`ledger-consistency.sh` passes over 103 rows.

The 7 sub-skips are still exactly `referenced-{claude-code-composed,opencode,opencode-zero-modules}`
and `mcp-{claude-code,codex-cli,opencode,pi-none}` — all genuinely stage (b); no in-scope surface
is hidden.

### F5 / F6 — FIXED (verified)

Differential range sweep against `semver@7.7.4` via `node -e`, **my own** 71 ranges × 28 versions
(`.temp/review-2/{cases.json,go-out.txt,node-out.txt}`, `cmp.py`). Every disagreement is either a
§1.4 restriction or the F5 decision:

| class | go | node | authority |
|---|---|---|---|
| `latest` | OK (`*`) | INVALID | §1.4 |
| `1.2.3 - 2.3.4` | INVALID | OK | §1.4 restriction (hyphen) |
| `v1.2.3`, `>=v1.2.3`, `^v1.2.3` | INVALID | OK | §1.4 restriction (`v` in range) |
| `1.2.3+build.5`, `>=1.2.3+build` | INVALID | OK | §1.4 (no build metadata) |
| `>= 1.2.3`, `> 1.2.3` (space) | INVALID | OK | schema `$defs/range` binds the operator to its version |
| `""`, `" "` | INVALID | OK (`*`) | schema `minLength: 1` |
| `>*`, `<*`, `>x`, `<X` | INVALID | `<0.0.0-0` | **F5 author decision** |

Zero satisfaction disagreements outside these classes; the caret-on-`0.x`/`0.0.x`, tilde,
comparator, x-range, partial-coercion, `-0`-bound, `||`, prerelease-admission and total-order
cases all match node exactly. Comparator-set spellings differ cosmetically
(`^0.0.0` → go `[">=0.0.0","<0.0.1-0"]` vs node `["<0.0.1-0"]`) with no behavioural consequence.

F5 note for the spec owner, not a finding: `agent-context-v1.schema.json#/$defs/range` still
admits `>*`, so a manifest that validates against the published schema is refused by curator with
`context_manifest_invalid`. The producer flagged this; it wants an erratum.

F6: `.temp/review-2/resolveattack/main.go`, graphs F/H/I/J/K/L/M/N:

```
F: root weights {"sk": 900} over a skill member  -> skill sk weight=900
H: root weights {"mm": 700} over an mcp member   -> mcp mm  weight=700
I: agreeing edges on a skill member              -> skill sk weight=33
J: disagreeing edges on a skill member           -> context_weight_conflict (both requirers named)
K: non-root weights map                          -> context_weights_not_root
L: root weights names a member outside the closure -> context_weight_unknown
M: disagreeing edges WITH rule 3 present         -> weight 99 + WARN context_weight_conflict  (§6 rule 2 carve-out)
N: root edge + root map on one package           -> context_weights_duplicate
```

All eight match §6 rules 1–4 verbatim, including the rule-2 warning carve-out and
`context_weights_duplicate`, which neither cycle had exercised.

### F7 — recorded bound, unchanged

Both audit call sites still pass `nil` waivers and no machine-config surface parses
`secret_material_waivers`. Correctly carried forward to the stage that lands manager-config
schema 2.

---

## New findings

## F8 — BLOCKING — the canonical source identity is never computed, so every `git` profile writes a lock and a marker that fail the published schemas

`repeat-of: none`

**Where.** `internal/envprofile/envprofile.go:678-681`

```go
// canonicalGit normalizes a git operand onto its canonical source identity.
func canonicalGit(operand string) string {
	return strings.TrimSuffix(strings.TrimSpace(operand), "/")
}
```

The comment is false: the function trims whitespace and one trailing slash and nothing else. It
is the only normalization on the path — `gitManager.Identity` (`gitsource.go:77-83`) returns it
verbatim, `Install` stores it as `Source.Git` (`envprofile.go:382`), `migrateGlobalSkills` uses it
for skill members (`:771`), and it lands unchanged in the lock member's `source` and in the
marker's `profile.source`.

**What is wrong.** §1: a `git` source is "a network git source under the **core §6.1 canonical
identity**". §1.3: each member records its "**canonical source identity** (`source`)". The repo
already computes exactly that — `internal/identity.Parse` returns `host/path`, transport removed,
host lowercased, trailing `.git` stripped, and errors on a malformed network source. `envprofile`
never imports it.

**Evidence — four spellings of one repository, driven through `envprofile.Install` with git
`insteadOf` making each URL resolve to the same local repo offline
(`.temp/review-2/identityattack/main.go`):**

```
operand                                     canonical identity (core §6.1)  lock_sha256 produced
https://github.com/companyA/acme            github.com/companyA/acme        sha256:132bd617e9e7f670...
https://github.com/companyA/acme.git        github.com/companyA/acme        sha256:18fc2134ed262a74...
git@github.com:companyA/acme.git            github.com/companyA/acme        sha256:ab4b2e32828b334a...
ssh://git@github.com/companyA/acme          github.com/companyA/acme        sha256:65fab5e0716d0ac3...
   lock source recorded:   "https://github.com/companyA/acme" / "…acme.git" / "git@github.com:…" / "ssh://…"
   marker profile.source:  same raw strings
```

Four different `lock_sha256` for one repository at one commit. §1.3 says "the lock hash is the
same on every machine that locks the same bytes", and the lock hash is the profile's **effective
pin** — the §5.1 generation header, the §8.2 marker, `env status`. §1 says "SSH and HTTPS URLs of
one repository yield one identity."

**The produced artifacts fail their published schemas.** ajv against `schemas/v1` at `f39f4a9`
(`.temp/review-1/validate.js`, real bytes captured from a real install):

```
produced-lock.json vs context-lock-v1.schema.json:            INVALID
  /members/0/source must match pattern "^(?!.*\.git$)[a-z0-9][a-z0-9.-]*/..."
marker-git-32.json vs agent-environment-marker-v1.schema.json: INVALID
  /profile/source must match pattern "^(?!.*\.git$)[a-z0-9][a-z0-9.-]*/..."
```

Every lock and every marker this implementation writes for a `git`-sourced profile is invalid.

**Third consequence — source agreement compares raw strings.**
`internal/contextresolve/contextresolve.go:418-425`:

```go
if declaredSource != "" && declaredSource != c.requirement.Source {
    return &Error{Diagnostic: DiagSourceMismatch, ...}
```

§1.4: "Every requirement on one package name MUST agree on the canonical source identity." Two
manifests naming the same repository as `https://github.com/x/y` and `git@github.com:x/y` are one
identity by the spec and a `context_source_mismatch` here.

**Why the suite missed it.** Every `envprofile`, `envmarker` and `cmd/curator` profile test uses a
`file://` operand, whose canonical identity is legitimately empty, so the raw string is never
distinguishable from the canonical one; and `conformance/v1/vectors/environments.json` already
supplies canonical identities (`github.com/companyA/root-context-core`), so the vector suite feeds
the implementation the answer. No test anywhere validates a **produced** lock or marker against
`schemas/v1` — `grep -rln "schema.json" internal/envprofile internal/envmarker internal/contextlock cmd/curator` is empty. Green on a fixture the real world never presents.

**Fix.** Canonicalize through `identity.Parse` at the operand boundary (`Install`, the requirement
reader, `migrateGlobalSkills`), keep the raw URL only where the spec asks for it (`path`
`source_path` is a different field), reject a malformed network source rather than passing it
through, and compare canonical identities in the agreement check.

**Regression cover required.** A test that installs one repository under an `https`, an `scp`-style
and an `ssh://` spelling (git `insteadOf` makes this offline, as above) and asserts one identical
`lock_sha256` and one store entry; a test that ajv- or code-validates a produced git-root lock and
marker against `context-lock-v1` / `agent-environment-marker-v1`; and a NARROWING mutant that
canonicalizes the host but keeps the trailing `.git` — the identity test must fail.

---

## F9 — BLOCKING — `profile install` applies neither the machine source allowlist nor audit revocations

`repeat-of: none`

**Where.** `internal/envprofile` imports neither `internal/config` nor `internal/identity` nor
`internal/audit`. The only gate on an installed member is `contextaudit.Detect`
(`envprofile.go:527`, `:604`), i.e. the two classes §9.1 *adds*. The pipeline it adds them to is
not run.

**What is wrong.** §9.1: "Every closure member passes the same gates: **canonical source identity
and the core §6.1 allowlist for `git` sources**, the MCP package allowlist for `mcp` members …
and the audit below", and "Profile installation always runs the **manager §7 source audit in
strict mode** over every member … The audit pipeline is unchanged — raw-tree hashing, the static
canary whose failure always blocks, deterministic detectors, **revocation**".

The repository already implements both for the skill pipeline:
`internal/closure.gateSource` (`closure.go:463-478`) — "applies the machine allowlist **before any
network clone**" — and `internal/audit.Gate` (`audit.go:101`), whose `revocationReason`
(`audit.go:195`) blocks a revoked content hash or source unconditionally. `allowed_sources` and
`audit.revocations` are **schema-1** machine-config keys (`config.go:51`, `:107`), already parsed
today, so nothing about this is blocked by the manager-config schema-2 bound.

**Evidence — one run, machine config that forbids the source twice, install succeeds
(`.temp/review-2/allowlistattack/main.go`; git `insteadOf` resolves the attacker URL to a local
repo so no network is touched):**

```
machine config allowed_sources=[github.com/relux-works] audit.enabled=true audit.revocations=[source:https://evil.example.com/*]
identity.Canonical("https://evil.example.com/pkg/root") = "evil.example.com/pkg/root" ; identity.Allowed = false
profile install *** SUCCEEDED *** name=pwned lock=sha256:bc7970de79f875b8... activated=true
   lock member kind=context name=pwned source="https://evil.example.com/pkg/root" commit=b0fe84200bff
```

The payload is a context module — prompt bytes that `profile use` then links into every agent
home. An operator who has locked their machine to one org's sources, and who has revoked a source
after an incident, gets neither gate on the profile surface. `curator global install` of the same
source refuses.

**Standard shape.** "The check is present but uncalled from production", twice over:
`identity.Allowed` and `audit.Gate` both exist, are correct, and are never reached from any
profile path. `contextresolve.MCPAllowlist` (`contextresolve.go:134`, used at `:456`) is the third
instance — `grep -rn MCPAllowlist` finds the field, its one use, and **no producer anywhere**, so
`mcp_package_not_allowed` cannot fire in production or in any test.

**Fix.** Read the machine config in the profile path and gate every `git` member's canonical
identity before the clone (reusing `closure.gateSource`'s shape), populate `MCPAllowlist` from the
same config, and run the §7 audit over every member new to the lock. Note that §9.1's "an advisory
profile install does not exist" is stronger than `audit.Gate`, which no-ops when
`cfg.Audit.Enabled` is false — the profile path must apply revocation and the canary regardless.
If any part of that genuinely cannot land in stage (a), it must be a **stated** bound in the
package doc and the report, which it is not today.

**Regression cover required.** Negative tests driving `envprofile.Install`: a source outside
`allowed_sources` refused before the clone; a revoked source and a revoked content hash refused;
an `mcp` member outside the MCP allowlist refused with `mcp_package_not_allowed`. Each with a
NARROWING mutant — e.g. an allowlist prefix match that drops the segment boundary, so
`github.com/relux-works-evil` is admitted and the test must fail.

---

## Verified correct — attacked, held

| Surface | Attack | Result |
|---|---|---|
| Ranges/versions | own 71 ranges × 28 versions differentially vs `semver@7.7.4` | every deviation is a §1.4/schema restriction or the F5 decision; 0 unexplained |
| Resolution | own graphs A–N: downward re-selection, never-increases, empty intersection, all four weight rules, rule-2 carve-out, duplicate/unknown/not-root | replay exactly as §1.4 and §6 state |
| Materialization | 13 vector cases byte-for-byte against `expected/environments/*` incl. both precedence primitives, no-chapter, zero-modules, system-prompt; file bytes, file hash and surface hash all asserted | PASS |
| Lock | CCJ-1 order and `lock_sha256` stable and correct for `path`/`local` roots; `(kind,name)` order over a mixed-kind lock | correct — see F8 for the `source` field |
| Marker | ajv against `agent-environment-marker-v1` for `path` and `local` roots; copied-surface reason present | VALID, `reason: claude-code-root-context` — see F8 for `git` roots |
| Detector | 5 install attacks incl. MCP `args`; unpinnable (no pin path in `Detect`); `context-system-module-present` always-warn surfaced as an install warning | correct |
| Switching | 28 SIGKILL iterations; one-adapter failure; unmanaged-file conflict; 9-switch backup retention | `profile_use_partial`, current unchanged, marker consistent, generations `4..8`, `environment_surface_unmanaged_conflict` |
| CLI | all 8 `cli/curator.md` profile rows and their refusals through the built binary | present: `profile_install_ref_conflict` ×3 shapes, `--use` takes no name, `environment_unknown`, `environment_target_unknown`, `profile_in_use`, `profile_update_blocked`, compose refused with the schema-2 bound, split-brain shown in `list` |
| Conformance | 5 top-level PASS, exactly 7 stage-(b) sub-skips, 0 FAIL, candidate root | matches the report |
| Mutant table | producer's 7 mutants inspected for shape | all NARROWING (gate stays, admits exactly one class member); not a delete-only table |
| Scope/signatures | `git log --format='%G? %GS'` over `a2406dfe..HEAD` | 8 commits, all `G`, `Ivan Oparin <oparin@me.com>`; diff additive; no unrelated behaviour change |

## Gates rerun by the reviewer at `ac9d0037`

| Gate | Result |
|---|---|
| `go build ./...` | exit 0 |
| `go vet ./...` | exit 0 |
| `gofmt -l cmd internal` | clean |
| `golangci-lint run` on envprofile, pkgversion, contextresolve | 0 issues |
| `go test -count=1 -race` on the 9 new/touched packages | all `ok` (envprofile 19.2s) |
| interop vectors, `CURATOR_CONFORMANCE_ROOT=<spec>/conformance/v1` | 5 PASS, 7 stage-deferred sub-skips, 0 FAIL |
| `bash .github/ci/gate-selftest.sh` | 81 passed, 0 failed |
| `bash .github/ci/no-broad-suppression.sh` | ok |
| `bash .github/ci/ledger-consistency.sh <evidence>` | ok, 103 rows across linux/darwin/windows |

`go test ./cmd/curator/` (~5 min) was **not** rerun; the producer's two first-hand runs
(pre-rebase 307s, post-rebase 314s, exit 0) are cited, not re-verified. The suite is green while
F8 and F9 both hold — it proves the code compiles and the fixture paths work, not that the gates
gate.

## AC coverage as measured

Rework items: **6 of 6** (F1–F6) fixed and independently verified; F7 stays a recorded bound.

Producer-brief items, measured against `f39f4a9`:

- items 1, 2, 4, 6, 8 — delivered and verified.
- item 3 (resolution and lock) — resolution, weights, order, CCJ-1 and `lock_sha256` correct; the
  lock's `source` field is not the canonical identity, so every `git` lock is schema-invalid (F8).
- item 5 (always-strict audit) — the two §9.1 classes are correct and correctly scoped; the
  pipeline they attach to (§6.1 allowlist, §7 revocation/canary, MCP allowlist) is absent (F9);
  waivers remain the F7 bound.
- item 7 (linked switching and CLI) — delivered with the stated F3 subset bound; the marker is
  schema-invalid for `git` roots (F8).

**6 of 8 brief items fully delivered.** All 8 `cli/curator.md` profile rows have a named driving
test (`cmd/curator/profile_test.go` ×11, `internal/envprofile/envprofile_test.go` ×24);
`remove --purge` is driven at the library level only (`TestRemovePurgeCleansHomes`), not through
`cmd/curator` — a small gap, not a finding.

## Reviewer housekeeping

Running the built binary's `curator init` from the worktree root created `Skillfile.json` and
appended a managed block to `.gitignore`. Both were reverted immediately
(`rm Skillfile.json`, `git checkout -- .gitignore`); `git status --short` is clean and the diff of
the reverted block is recorded in this run's transcript. Every other probe ran with `HOME`,
`CLAUDE_CONFIG_DIR`, `CODEX_HOME`, `XDG_CONFIG_HOME`, `PI_CODING_AGENT_DIR` and `CURATOR_CONFIG`
pointed at temporary directories; `~/.curator/profiles` does not exist, so no operator state was
created this cycle. Scratch lives in the worktree's `.temp/review-2/`.
