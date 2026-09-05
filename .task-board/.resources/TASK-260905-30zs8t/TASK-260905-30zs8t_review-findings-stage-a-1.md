# TASK-260905-30zs8t — review findings, stage (a) core, cycle 1

Subject: curator worktree `/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-a-core`,
branch `feat/agent-environments-stage-a`, head `7238412c` on `bb14375a`.
Authority: curator-spec main `f39f4a9`.
Verdict: **CHANGES REQUESTED** — 2 blocking, 2 major, 2 minor.

Everything below was reproduced by the reviewer against the committed tree.
Reviewer scratch (harnesses, differential streams): the worktree's
`.temp/review-1/`. Read-only on tracked files.

---

## F1 — BLOCKING — `context-secret-material` is completely bypassed for any package addressed with a `directory`

`repeat-of: none`

**Where.** `internal/envprofile/envprofile.go:480` and `:545`

```go
report, err := contextaudit.Detect(entry, pinOf(resolved), nil)
```

`entry` comes from `gitManager.entryPath` (`internal/envprofile/gitsource.go:292-297`),
which returns the **snapshot root** and never appends `resolved.Directory`.
`contextaudit.InScope` (`internal/contextaudit/contextaudit.go:51-60`) admits only
top-level `context/`, `agent-context.json`, `agent-mcp.json`, `CONTEXT.md`. So for a
package at `sub/`, nothing is ever scanned.

Fifteen lines below the second call the *manifest* load already does it correctly
(`envprofile.go:559-560`):

```go
root := entry
if resolved.Directory != "" {
    root = filepath.Join(entry, filepath.FromSlash(resolved.Directory))
}
```

**What is wrong.** §9.1 makes this detector always-strict, always-blocking and
unpinnable: "Because profile installation is always strict, a member carrying such
a finding fails installation." A `--directory` install, or any transitive
`requires.contexts.<n>.directory`, walks straight past it. `--directory` is on the
shipped CLI (`curator profile install -h`), so this is operator-reachable directly,
not only transitively.

**Evidence — driven through the production entry points, byte-identical payloads.**

```
control: secret at snapshot root (no --directory)          REFUSED: profile_source_invalid: member context:acme carries a blocking context-secret-material finding
attack 1: same secret under sub/ with --directory sub      *** INSTALLED *** name=acme hash=sha256:b809bf69...
attack 2: transitive requires{directory:sub} member         *** INSTALLED *** name=clean hash=sha256:895bb133...
control 2: transitive member, secret at its snapshot root  REFUSED: profile_source_invalid: member context:dep2 carries a blocking context-secret-material finding
```

The two controls prove the detector itself works; only its scope is wrong.
Then `envprofile.Use` materializes the unaudited bytes into the operator's agent home:

```
installed acme lock=sha256:... activated=true warnings=[]
use err=<nil> results=[{claude_code OK:true} {codex_cli OK:true} {opencode OK:true} {pi OK:true}]
--- materialized CLAUDE.md ---
<!--
curator-root-context-v2
...
---

## Context: acme 1.0.0

aws key AKIA1234567890ABCDEF here

CONTAINS THE UNAUDITED SECRET: true
```

Harness: `.temp/review-1/auditattack/main.go`.

**Why the suite missed it.** `TestInstallRejectsSecretMember`
(`internal/envprofile/envprofile_test.go:83`) is the only negative test for this gate
and it puts the package at the snapshot root. The bypass path around the check was
never exercised — the standard "bypass path around the check" shape.

**Fix.** Audit at the package root, not the snapshot root: pass
`filepath.Join(entry, filepath.FromSlash(resolved.Directory))` at both call sites,
for `context` and `mcp` members alike.

**Regression cover required.** A `--directory` variant and a transitive
`requires.contexts{directory}` variant of `TestInstallRejectsSecretMember`, plus a
NARROWING mutant: restore the snapshot-root path and show the new tests fail while
the existing root-case test still passes.

---

## F2 — BLOCKING — the builtin `local` `default` profile is unusable, and carries no migrated global skills

`repeat-of: none`

**Where.** `internal/envprofile/envprofile.go:631-654`, `EnsureDefault`.

It hashes a **freshly created empty temp directory**, writes a lock pinning that hash,
and never creates the corresponding store entry — no `contextstore.EnsureState` call.
The lock's only member is the profile itself at `0.0.0` with the empty-string SHA-256:

```json
{"members":[{"kind":"context","name":"default","overlay":false,"required_by":[],
"state_sha256":"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
"version":"0.0.0","weight":0}],"root":"default","schema_version":1}
```

**What is wrong.** Two distinct failures.

1. Brief item 8 and §9.4 require "the builtin `local` `default` profile **whose lock
   carries the migrated global skills**". No skill member is ever added; nothing is
   migrated.
2. The lock names a store entry that does not exist, so every path that materializes
   the default profile fails on a fresh manager home — the state every operator starts
   from.

**Evidence — fresh isolated manager home, production entry points:**

```
List err: <nil>
  profile default  kind=local  lock=sha256:552290707b5a... members=1
     member kind=context name=default version=0.0.0 state=e3b0c44298fc...

Use(default) err: profile_source_invalid: context_manifest_invalid: agent-context.json is absent at <home>/contexts/context/default/e3b0c442...
Sync err:         profile_source_invalid: context_manifest_invalid: agent-context.json is absent ... results: 0
Use --clear --env claude_code err: profile_source_invalid: context_manifest_invalid: agent-context.json is absent ... results: 0
```

Three `cli/curator.md` rows are dead on a fresh machine: `curator profile sync`,
`curator profile use default`, and `curator profile use --clear --env <env-id>` —
the last being exactly §9.3's "re-materialize the scope from the machine default".

Harness: `.temp/review-1/defaultattack/main.go`. Reproduced against the real binary
too: `curator profile use --clear --env claude_code` failed identically.

**Why the suite missed it.** `TestEnsureDefaultCreatesLocalProfile`
(`internal/envprofile/envprofile_test.go:548`) asserts only that `source.Kind == local`
and that the lock validates. It never drives `Use`, `Sync`, or any materialization.
The drafting report calls item 8 "driven (TestEnsureDefaultCreatesLocalProfile)" —
driven to creation, not to use. Positive-path-only evidence for a claimed AC row.

**Fix.** Create the default profile's store entry alongside its lock, and populate the
lock with the migrated global skill members. Then drive `Use`/`Sync` against the
default profile in a test on a fresh home.

---

## F3 — MAJOR — the profile write paths take no manager-home mutation lock and write no journal

`repeat-of: none`

**Where.** `internal/envprofile/switch.go:1-20` (package doc), and the absence of any
`internal/transaction` / `internal/managerlock` import anywhere in the new packages:

```
$ grep -rn "internal/transaction\|internal/managerlock" internal/envprofile internal/contextstore internal/envmarker cmd/curator/profile.go
(no matches)
```

`cmd/curator/main.go:175` dispatches `case "profile"` without taking the lock,
unlike the established skill-install path (`internal/install/commit.go`, which uses
both `managerlock` and `transaction`).

**What is wrong.** §9.2 step 1: `profile use` re-materializes "atomically per entry,
**under the manager-home mutation lock, journaled like any other manager-home
transaction (manager §2.5)**", and "`profile use` of either profile … **completes the
scope from the journal**". Brief item 7 restates this as "`profile use` as one
`transaction` Plan of per-entry targets". What ships is a hand-rolled loop that
reproduces the *observable* M11 shape but has no mutation lock (two concurrent
`curator profile` invocations race over the same manager home and the same agent
homes) and no journal (the specified recovery path does not exist).

The `switch.go` package doc asserts "the M11 transactional shape" without qualifying
that the lock and journal are absent, and the drafting report does not list this as a
bound — it is an undeclared gap, not a stated one.

**Verified correct, for the record.** The observable shape itself is right. Live
mid-switch failure (one adapter's native home replaced by a regular file):

```
switch to beta with codex_cli broken:
  claude_code  ok=true
  codex_cli    ok=false mkdir <home>/codex: not a directory
  opencode     ok=true
  pi           ok=true
  err: profile_use_partial: the scope is partially switched; the recorded current is unchanged
  recorded current after partial switch: alpha (must still be alpha)
  claude CLAUDE.md now names: beta
```

That matches §9.2 exactly, including leaving the succeeded entries switched. Harness:
`.temp/review-1/switchattack/main.go`.

**Fix.** Either take the manager-home mutation lock and journal the per-entry plan
through `internal/transaction`, or state the omission explicitly as a bound in the
package doc, the report and the board, and carry it as a named follow-up.

---

## F4 — MAJOR — the stage-(b) conformance skip is classified `opt-in` on the strength of an environment variable that does not exist

`repeat-of: none`

**Where.** `internal/interop/context_materialization_test.go:145`

```go
t.Skipf("surface %s form %s deferred to stage (b); set CURATOR_STAGE_B=1 when it lands", tc.Surface, tc.Form)
```

and the ledger row `.github/ci/platform-cases.tsv:187`:

```
internal/interop	TestConformanceEnvironmentsMonolithic/*	...	opt-in	referenced-form and mcp cases belong to stage (b) and record an opt-in skip until CURATOR_STAGE_B=1 lands
```

**What is wrong.** `CURATOR_STAGE_B` is read nowhere in the repository — it occurs
only in those two strings. The skip is unconditional. `.github/ci/skip-classes.tsv:48`
defines the class it claims:

```
opt-in	set CURATOR_[A-Z_]+=1	allow	enabled by an explicit developer environment variable
```

The reason text matches that regex, so the gate whose stated purpose is that "a skip
whose reason matches NOTHING here is fatal: that is how a newly-introduced skip is
caught the first time it runs, instead of quietly shrinking the suite" is satisfied by
a sentence rather than by a knob.

**Evidence — the capability claim does not reproduce:**

```
=== with CURATOR_STAGE_B=1 === 7
=== without ===               7
```

Identical skip counts.

**Scope check, in the producer's favour.** The 7 sub-skips are exactly
`referenced-claude-code-composed`, `referenced-opencode`, `referenced-opencode-zero-modules`,
`mcp-claude-code`, `mcp-codex-cli`, `mcp-opencode`, `mcp-pi-none` — all genuinely stage (b)
per the brief. **No in-scope surface is hidden.** The defect is the gate bypass, not a
coverage hole.

**Fix.** Either honour `CURATOR_STAGE_B=1` (run the cases and let them fail loudly
until stage (b) lands), or add an honest class to `skip-classes.tsv` for
"deferred to a later delivery stage" and classify the row under it.

---

## F5 — MINOR — `>*`, `<*`, `>x`, `<X` mean "match everything" where node-semver means "match nothing"

`repeat-of: none`

**Where.** `internal/pkgversion/pkgversion.go:334-338`

```go
if isWild(match[2]) {
    // A bare "*", "x", or "X" — with or without an operator — is the
    // any-comparator; node-semver reads every operator on "*" as "*".
    return []Comparator{{Any: true}}, nil
}
```

**What is wrong.** The comment is false for `>` and `<`. node-semver's
`replaceXRange` turns `>*` and `<*` into `<0.0.0-0`, which matches nothing.
§1.4 binds the semantics: "Its semantics are those of node-semver (the npm
implementation … recorded against 7.7.4), restricted as stated" — and this is not one
of the two stated restrictions (hyphen ranges, `v` inside a range). These spellings are
also **admitted by the published schema**: `agent-context-v1.schema.json#/$defs/range`
allows `(?:>=|<=|>|<|=|\^|~)?(?:\*|x|X|...)`. So either they must carry node's meaning
or the schema must stop admitting them; "matches everything" is available under neither
reading.

**Evidence — differential sweep against `semver@7.7.4` via `node -e`,
88 ranges × 27 versions and 41 ranges × 32 versions:**

- satisfaction disagreements outside this class: **0** — every caret (incl. `^0.x`,
  `^0.0.x`), tilde, comparator, x-range, partial coercion, `-0` upper bound, `||`,
  `latest`, prerelease-admission and total-order case matches node exactly; hyphen
  ranges, `v`-inside-range and build metadata are correctly rejected.
- disagreements inside this class: **88**, all of shape `go=true node=false`.

Production consequence through `contextresolve.Resolve` (not just the parser):
a manifest requirement `{"range": ">*"}` selects `alpha@2.0.0` where node-semver
yields no candidate and §1.4 step 2 demands `context_range_conflict`.

Harnesses: `.temp/review-1/rangediff/`, `node-run.js`, `cmp.py`,
`.temp/review-1/resolveattack/` graph G.

**Fix.** Reject `>`/`<` on a bare wildcard as `profile_source_invalid`, or map them to
`<0.0.0-0` as node does. Correct the comment either way. Add the four spellings to the
negative range tests.

---

## F6 — MINOR — a root `weights` entry naming a `skill` or `mcp` member is silently ignored

`repeat-of: none`

**Where.** `internal/contextresolve/contextresolve.go:748` — the whole effective-weight
computation sits inside `if sel.kind == contextlock.KindContext {`, so a non-context
member always locks at `weight 0`.

**What is wrong.** §6: "**Every closure member** has one effective weight, computed by
exactly these rules in order", rule 3 being the root's `weights` map — whose
`propertyNames` in `agent-context-v1.schema.json` is a bare identifier, so it may name a
skill or MCP member. §1.3 requires the lock to record "its effective `weight`
(section 6)" for each member. The entry is not rejected either: `context_weight_unknown`
tests membership in `r.selected`, which includes skills and MCP packages.

**Evidence** (`.temp/review-1/resolveattack/` graph F) — root manifest
`"weights": {"sk": 900}` over a skill member `sk`:

```
context  root  v1.0.0  weight=0  required_by=[]
skill    sk    v1.0.0  weight=0  required_by=[root]
```

Silently 0, no diagnostic. Since the lock hash is the profile's effective pin, a
conforming manager that applies rule 3 to `sk` writes different lock bytes for the
same inputs — a cross-manager divergence on the identity surface. Not covered by
`vectors/environments.json`.

**Fix.** Either apply rules 1–3 to every member kind, or reject a `weights` entry
naming a non-context member. This may want a spec erratum rather than a code change —
flagging it for the spec owner rather than prescribing.

---

## F7 — accepted bound, recorded — scoped waivers have no production path

Both production call sites pass `nil` waivers (`envprofile.go:480`, `:545`), and no
machine-config surface parses `secret_material_waivers`. The unmatched-waiver branch
at `envprofile.go:553` is therefore unreachable in production.

Brief item 5 says "scoped waivers `{pin, file, span, reason}` **from machine config**",
which is unmet. The same brief puts manager-config schema 2 out of scope for stage (a),
so the config surface genuinely cannot land here, and the producer declared it. **Accepted
as a bound**, with the note that the "from machine config" clause of item 5 must be
picked up by the stage carrying schema 2, and that `Waiver.Reason` is currently not
required by the matcher (a reason-less waiver still clears a finding).

---

## Verified correct — attacked, held

These were attacked independently, not read.

| Surface | Attack | Result |
|---|---|---|
| Versions/ranges | 129 ranges × 27–32 versions differentially vs `semver@7.7.4` | exact match outside F5; all prerelease rules, `-0` bounds and rejections correct |
| Resolution | own graphs: forced downward re-selection, never-increases after a requirer leaves, empty intersection, weight precedence, weight conflict | all replay exactly as §1.4 steps 1–4 and §6 rules 1–4 state |
| Lock | CCJ-1 bytes hand-recomputed from `registry.md` §1; `(kind,name)` bytewise order over a 5-member mixed-kind lock; ajv validation against `context-lock-v1` | hand hash `sha256:597e3f01…` == vector; order correct; VALID |
| Materialization | header + chapter + join hand-written from §5/§5.1 in Python and diffed against `expected/environments/monolithic-claude-code/CLAUDE.md` | byte-identical, 693 bytes; file hash and surface hash match |
| Detector | pattern classes and placeholder rule vs `vectors/context-detectors.json` | classes, `placeholder_rule` and `span_rule` match the vector verbatim; unpinnable (no pin path exists in `Detect`); `context-system-module-present` is always-warn |
| Switching | live one-adapter failure | every entry attempted, per-adapter results, `profile_use_partial`, recorded current unchanged, succeeded homes switched — §9.2 exactly |
| Marker | ajv against `agent-environment-marker-v1` | VALID; carries `members`, `precedence`, `surfaces` with `reason: claude-code-root-context` |
| CLI | every `cli/curator.md` profile row and its refusals | present; `profile_install_ref_conflict` on two requirement flags, `--use` takes no name, `environment_target_unknown` on `--target`, `compose` refuses with the schema-2 bound |
| Conformance | rerun with `CURATOR_CONFORMANCE_ROOT=<spec>/conformance/v1` | 5 top-level PASS, exactly 7 sub-skips (3 `referenced-*`, 4 `mcp-*`), 0 FAIL — the report's tally is accurate |
| Scope/signatures | `git log --format='%G? %GS'` | 4 commits, all `G`, `Ivan Oparin <oparin@me.com>`; diff additive, no unrelated behaviour change |

## Gates rerun by the reviewer at `7238412c`

| Gate | Result |
|---|---|
| `go build ./...` | exit 0 |
| `go vet ./...` | exit 0 |
| `gofmt -l cmd internal` | clean |
| `go test -count=1 -race` on the 9 new packages | all `ok` |
| interop context/environments vectors, candidate root | PASS, 7 stage-(b) sub-skips |

`go test ./cmd/curator/` (~4–8 min) was **not** rerun; the producer's exit 0 at 278s is
cited, not re-verified. The full suite is green while F1 and F2 both hold — the suite
proves the code compiles and the happy paths work, not that the gates gate.

## AC coverage as measured, not as prose

The report claims "8 of 8 brief items delivered". Measured against the brief:

- items 1, 2, 3, 4, 6 — delivered and verified (item 2 with F5 outstanding).
- item 5 — delivered except the machine-config waiver surface (F7, accepted bound)
  **and** with the detector's scope broken for directory-addressed packages (F1).
- item 7 — observable shape delivered; mutation lock and journal absent and undeclared (F3).
- item 8 — **not delivered**: no migration, and the default profile cannot materialize (F2).

**5 of 8 brief items fully delivered.** Of the 12 `cli/curator.md` profile rows, 3
(`sync`, `use <default>`, `use --clear`) fail on a fresh machine because of F2.

## Reviewer housekeeping

The reviewer's real-binary probe of `profile use --clear` ran against the operator's
manager home and caused `EnsureDefault` to create `/Users/iv/.curator/profiles/default/`
(`lock.json`, `source.json`), which did not exist before. Both files were copied to
`.temp/review-1/operator-home-profiles-created-by-review/` and the directory was removed,
restoring the pre-probe state. No agent home, no `contexts/` entry and no pre-existing
file was touched. This is separate from — and much smaller than — the incident the
drafting report already records under "Anomalies".
