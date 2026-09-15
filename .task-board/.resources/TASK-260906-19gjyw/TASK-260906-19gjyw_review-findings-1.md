# TASK-260906-19gjyw — review findings, cycle 1

Verdict: **CHANGES REQUESTED** — routed to `to-dev`.
repeat-of: none (first review cycle on this leaf).

Subject: `relux-works/curator`, branch `feat/consume-overlay-rule`, head `38702164`
(two signed commits past `7c74a492`; both `%G? = G`, author `Ivan Oparin <oparin@me.com>`).
PR https://github.com/relux-works/curator/pull/62. Authority: curator-spec `87a0d0060bad`.
Change Request `CR-TASK-260906-19gjyw-1` revision 1.

Everything below was re-measured in a throwaway copy of the producer's worktree
(`.temp/review/curator`, `rsync` excluding `.git`/`.temp`/`.task-board`). Nothing was
written into the producer's worktree or the control root. The two `test-gate` lanes were
never run concurrently, and no `-race` suite ran alongside anything here.

---

## 0. The empty repository delta

`repository_delta=empty` is **structurally correct here and is not the producer's doing.**
The story worktree this Change Request snapshots is a checkout of **curator-spec**
(`base 87a0d0060bad`, candidate tree `2e6ca472`); I confirmed `HEAD^{tree}` of the story
worktree equals both the base tree and the candidate tree, so the patch has zero paths
because nothing in curator-spec changed — correctly, since this leaf's scope is the
**curator** repository. The deliverable is curator `feat/consume-overlay-rule@38702164`,
which is outside the CR's snapshot scope by construction. This run's own brief was
publish-only ("make no code changes"), so an empty tree is the truthful record of it.

**For the orchestrator:** accepting or rejecting this CR does not move the curator work
either way. The board's integration machinery would integrate an empty tree. That mismatch
is worth deciding deliberately rather than inheriting; it is not a finding against the
producer.

---

## 1. Acceptance-criteria coverage as measured by this review

**5 of 7 AC rows fully driven; 2 carry measured gaps.**

| # | AC row | Verdict |
|---|---|---|
| 1 | path overlay reachable from all three surfaces, joins the closure with its weight, through `run()` | driven — verified |
| 2 | one exported discriminator matches the landed schema, never probes the filesystem | driven — verified, 121 spellings, three engines, 0 disagreements |
| 3 | a path overlay carrying range/tag/branch/revision/directory stays `profile_source_invalid` | driven for four of five; `branch` is an accepted stated bound |
| 4 | `profile install` uses the same helper with **no source silently changing kind** | **gap — F2**: one kind-change class is unenumerated and untested |
| 5 | candidate lane green on three runners with `CI_REQUIRE_FULL_ROOT=1`, closing the 30 subcases | driven — verified from the hosted jobs' own logs and reproduced locally |
| 6 | the stale bound and ledger rows 303–305 are retired and **truthful** | **gap — F1**: one newly registered ledger row is not truthful |
| 7 | every new refusal driven through `run()` and proven by a narrowing mutant | driven — 16 of 17 mutants re-run and killed; one false kill claim (F1) |

---

## F1 — MAJOR. `TestInstallNeverDemotesANetworkIdentityToAPath` checks none of the class it names; three artifacts repeat the claim

`internal/envprofile/envprofile_f13f14_test.go`.

**Measured.** I applied M10 exactly as the report describes it — deleted the widening in
`installOperandKind`, leaving `return identity.ClassifySource(trimmed)` — and ran each of the
three named tests as its own process:

| test | exit | verdict |
|---|---|---|
| `TestInstallNeverDemotesANetworkIdentityToAPath` | **0** | **SURVIVED**, logging `checked 5 network-identity operands` |
| `TestOperandShadowedByDirectoryResolvesAsGit` | 1 | killed (`shadowed operand err = <nil>, want profile_source_invalid`) |
| `TestInstallOperandKindIsSyntactic` | 1 | killed |

**Cause.** The property loop filters on `identity.Parse(strings.TrimSpace(tc.operand))`
being non-empty. `identity.Parse` returns `("", nil)` for a bare canonical identity —
no scheme, no colon, no `@`, so `scpRE` does not match and the function takes its local
branch. Those bare spellings — `github.com/evil-org/pkg`, `github.com/relux-works/pkg`,
`example.com/org/pkg` — are **exactly and only** the operands the widening exists for. All
three are `continue`d. The 5 the test does check (`https://…`, `ssh://…`,
`git@example.com:org/pkg.git`, `c:example/pkg`, `github.com:/org/pkg`) are decided `git` or
`invalid` by `ClassifySource` alone and are unaffected by the widening in either direction.
The `if checked == 0 { t.Fatal }` guard and the `t.Logf("checked %d")` line are the
coverage-ratio mechanism, and the number they report is 5 for a class of size 0.

**The same false claim appears in three places:**

1. The test's own comment: *"A mutant that drops the widening and follows the overlay
   classification verbatim turns `github.com/evil-org/pkg` into a path install and **fails
   here**."* — false by measurement.
2. Drafting report §4, row M10: *"`envprofile TestOperandShadowedByDirectoryResolvesAsGit`
   (+2)"* — 2 of the 3 fail, not 3.
3. `.github/ci/platform-cases.tsv`, the newly registered row *"no install operand carrying
   a core 6.1 canonical network identity is classified path, which is what would hand it
   the allowlist bypass"* — the durable ledger now asserts a property no test asserts. AC
   row 6 requires the ledger rows to be **truthful**; this one is not.

**What is NOT wrong.** The widening itself is correct and I verified every step of the
producer's reasoning independently:

* `canonicalGit` (`envprofile.go:1340`) returns the operand verbatim when
  `identity.ValidCanonical(trimmed)` — so a bare `host/path` is accepted as a git operand.
* `ensureRepo` (`gitsource.go:83`) sets `cloneURL = "https://" + canonical` when
  `identity.ValidCanonical(canonical)` and no raw spelling is recorded — so it really is a
  network clone.
* The planted-directory attack reproduces. Under M10, `TestOperandShadowedByDirectoryResolvesAsGit`
  fails with **`err = <nil>`**: `Install` of `github.com/evil-org/pkg` with a planted
  `./github.com/evil-org/pkg` and `AllowedSources: ["github.com/relux-works"]` **succeeds**,
  installing the planted bytes past the machine allowlist. With the widening it is refused
  with the allowlist reason and clones nothing. F14 is genuinely defended.
* I enumerated the install kind changes myself (see F2) rather than trusting the table:
  **no operand that reached the network before reaches a local path now.**

**Repair.** One line: filter the property loop on the predicate the widening actually uses,
e.g. `canonical, err := identity.Parse(t); if err != nil || (canonical == "" && !identity.ValidCanonical(t)) { continue }`,
so the three bare canonical identities are checked and M10 kills it. Then correct the M10
row in the report and the ledger row's wording. **No product code needs to change.**

---

## F2 — MINOR (but it is the AC row's own word). The "complete list of install kind changes" omits a refused→path class

Drafting report §3 heads its table *"Complete list of install kind changes — nothing here is
silent"*, and AC row 4 requires *"no source silently changing kind"*.

I enumerated the change myself by restoring stage (c)'s `isPathOperand` verbatim from
`7c74a492` beside `installOperandKind` and driving 57 operands through both, printing the
old git destination for each. Result: every row in the report's table reproduces, **and one
class is missing.**

| operand | before | after | not in the report's table |
|---|---|---|---|
| `packages/team:context` | git → refused at `canonicalGit` (`malformed or unsupported network source`) | **path** | yes |
| `a/b:c` (same shape) | git → refused | **path** | yes |

A colon in a later segment made the operand *unclassifiable* before and makes it a **local
path install** now. That is a refuse→accept direction, it is the one direction the table
never shows, and `installOperandCases` contains no colon-in-a-later-segment operand, so
nothing pins it. Risk is low — the spelling carries no network identity and the schema does
decide it is a path — but the table asserts completeness and the AC asserts silence.

**Repair.** Add the row to §3, and add one operand (`packages/team:context` →
`identity.SourcePath`) to `installOperandCases`.

*(Checked and confirmed benign, for the record: `GitHub.com/example/x`, `~pkg`,
`github.com/../../etc/passwd` all move git→path but fall inside the table's stated
"no colon, not a canonical identity" class, and all were local clones before.)*

---

## F3 — TRIVIAL. A doc comment is attached to the wrong function

`internal/envprofile/envprofile.go`, around line 1284. The block that begins
`// installOperandKind classifies a 'profile install <git-url|path>' operand …` runs without
a break into `// sourceKindRefusal is the refusal for a spelling that is neither kind. …`
and sits above `func sourceKindRefusal`. `installOperandKind` — declared below it — has no
doc comment at all, and godoc renders `installOperandKind`'s entire rationale, including the
widening's justification, as documentation for `sourceKindRefusal`. `golangci-lint run ./...`
is clean (0 issues), so nothing catches it.

---

## 2. What I verified and found correct

### 2.1 The discriminator against the committed schema — 121 spellings, three engines, zero disagreements

`identity.ClassifySource` never touches the filesystem: `internal/identity/sourcekind.go`
imports `regexp` and nothing else. That is a static fact, not an inference.

I built a differential harness rather than reading the transcription:

* **Corpus.** All 41 published `manager-config-v2` overlay `source` spellings, extracted
  from the cases themselves, plus every prose edge the briefs name, plus adversarial
  additions: scheme-case permutations, scheme-charset edges (`ht-tp://`, `ht.tp://`, `h://`),
  every drive spelling, SCP host/path edges, all ECMA-262 whitespace code points in both
  positions (`\v`, `U+00A0`, `U+1680`, `U+2028`, `U+3000`, `U+FEFF`, and `U+180E` which is
  *not* ECMA whitespace), astral characters (surrogate-pair vs rune width), embedded
  newlines (multiline-anchor probes), and leading/trailing space. **121 spellings.**
* **Engine A — Node/V8.** Reads the five `pattern` strings **verbatim out of the committed
  `$defs/overlay`** (no transcription by me) and asserts the three arms use identical
  patterns across `allOf[0..2]`, then composes them.
* **Engine B — ajv 8 (draft 2020-12) over the whole schema.** No hand-composition at all:
  for each spelling it validates `{source: X}` and `{source: X, range: "^1"}` against the
  full committed schema and derives `path` / `git` / `invalid` from which one validates.
  As a control, ajv agrees with `conformance/v1/schema-cases/index.json` on **71 of 71**
  `manager-config-v2` cases, 0 disagreements.
* **Engine C — the Go helper**, dumped from a test inside the package.

**Three-way result: 121 spellings, 0 disagreements** (41 path, 40 git, 40 invalid).

Every edge the review brief names is in that set and agrees: `C:\…`, `C:/…`, `C://…`,
bare `C:` (refused), `c:\users\…`, `c:example/x` (git, one-character §6.1 host),
`packages/team:context` (path), `github.com:\example\x` (refused), `file:` (refused),
`svn://` (refused), and each scheme in `HTTPS://` / `HtTpS://` spellings.

**The ECMA-262 whitespace trap: re-applied, and no other dialect gap found.** M9 — replacing
`ecmaSpace` with Go's `[\t\n\f\r ]` — is killed by `identity.TestClassifySourceEcmaWhitespace`
(exit 1). The producer's spelled-out class is exactly current ECMA-262 `\s`
(WhiteSpace ∪ LineTerminator: `\t \v \f \r \n`, space, `U+00A0`, the Zs separators
`U+1680`, `U+2000–200A`, `U+202F`, `U+205F`, `U+3000`, `U+2028`, `U+2029`, `U+FEFF`) —
including correctly **excluding** `U+180E`, which older Unicode had as Zs. I looked for the
other dialect gaps the brief asks about and found none that can bite here: no backreferences
or lookaround, so RE2 and a backtracking engine answer the same existence question;
greediness cannot change a boolean match; every pattern is `^`-anchored and neither engine
is in multiline mode (my embedded-newline spellings confirm this by execution); the only
`.` characters are literal inside classes; and the UTF-16-vs-rune difference for astral
characters cannot change any of these five patterns' verdicts (verified with `U+1F600` in
three positions).

### 2.2 The dropped Windows-drive carve-out — ordering verified by construction

`driveRE` (`^[A-Za-z]:[\\/]`) is evaluated in the `else if` **above** the `scpRE` branch, so
the drive spellings never reach the retired `len(host) == 1` rule. Confirmed by execution,
not by reading — `identity.Parse` returns the local verdict `("", nil)` for `C:/x`, `C:\x`,
`c:/x`, `c:\x`, `D:\a\b`, `Z:\proj\ctx`, `c:\users\op\context`, `c:/Users/op/context`, and
the network identity `c/example/x` for `c:example/x`, `C:example/x`, `git@c:example/x`.
M16 — restoring the carve-out — is killed by
`identity.TestParseOneCharacterHostIsANetworkIdentity` (exit 1). Nothing else in the tree
depended on the carve-out: `grep` shows one discriminator with four production call sites
and no surviving `isPathOperand`.

### 2.3 The three surfaces, each addressing mode, through `run()`

* **Reader** (`internal/config/environments.go:446`) — the blanket `forms != 1` is a kind
  switch; field grammar is validated first, as the schema does in `properties`, then the
  kind rule, as the schema does in `allOf`. `render()` no longer emits `"revision": ""` for
  a path overlay.
* **`resolveOverlay`** (`internal/envprofile/overlays.go:53`) — same helper; the neither-kind
  arm refuses **before** `canonicalGit`, so nothing is cloned or snapshotted.
* **`profile compose add`** (`cmd/curator/compose.go:57`) — same helper; the previously dead
  `default: form = "path"` branch of `compose list` is now live and asserted.
* End to end: `cmd/curator TestPathOverlayFromMachineConfigJoinsTheClosure` declares a path
  overlay with `weight: 250` in machine configuration and drives `profile install --use`
  through `run()`, asserting an overlay-flagged marker member at weight 250 with a state-hash
  pin and no commit. `TestOverlayFromMachineConfigIsRefusedByKind` drives nine refusals
  through `run()` via the production `configSource` seam and requires the stderr to name the
  `environments.overlays` row.
* The refusal assertions require the **kind refusal reason** (`"neither a git source nor a
  path"` / the F12 no-network-identity wording), not merely `profile_source_invalid`, and
  assert `profile-repos` stayed empty. That is the right discipline: without it a
  fall-through to the git arm would produce the same diagnostic from a downstream clone
  failure and prove nothing.

All green locally against the manifest-verified candidate root:
`internal/identity ok`, `internal/config ok`, 30 of 30 targeted `envprofile` subtests PASS,
`cmd/curator` compose/overlay rows 27 subtests PASS.

### 2.4 Mutants — 16 of 17 re-run by me, plus both text-gate mutants

Every mutant below was written by me from the report's description, applied to a fresh copy,
run as its own process with its `=== RUN` count checked (so a no-op filter reports NOT-RUN
rather than "killed"), then reverted.

| # | narrowed to admit | named test | ran | exit | verdict |
|---|---|---|---|---|---|
| M1 | reader: a revision on a path source | `config …/invalid-overlay-path-requirement-form.json` | 2 | 1 | killed |
| M2 | reader: a form-free git source | `config …/invalid-overlay-no-requirement-form.json` | 2 | 1 | killed |
| M3a | discriminator: `svn` added to the admitted schemes | `config …/invalid-overlay-unknown-scheme-with-form.json` | 2 | 1 | killed |
| M3b | discriminator: the refusal arm skips exactly `svn://` | `config …/invalid-overlay-unknown-scheme.json` | 2 | 1 | killed |
| M4 | reader: `directory: packages/team` on a path source | `config …/invalid-overlay-path-directory.json` | 2 | 1 | killed |
| M5 | discriminator: an underscore SCP host | `config …/invalid-overlay-scp-host-grammar-with-form.json` | 2 | 1 | killed |
| M6 | discriminator: a one-character URL scheme | `config …/valid-overlay-path-windows-double-slash.json` | 2 | 1 | killed |
| M7 | discriminator: drive rule, forward slash only | `config …/valid-overlay-path-windows-lowercase-drive.json` | 2 | 1 | killed |
| M8 | discriminator: drops plain `http` | `config …/valid-overlay-git-http-uppercase.json` | 2 | 1 | killed |
| M9 | discriminator: whitespace class becomes Go `\s` | `identity TestClassifySourceEcmaWhitespace` | 1 | 1 | killed |
| M10 | install: drops the network-identity widening | `envprofile TestOperandShadowedByDirectoryResolvesAsGit` | 1 | 1 | killed — **but see F1** |
| M11 | install: neither-kind refusal admits exactly `D:` | `envprofile TestInstallRefusesANonSourceOperand` | 7 | 1 | killed |
| M12 | `resolveOverlay`: neither-kind arm admits exactly `C:` locally | `envprofile TestOverlayInvalidSourceKindIsSourceInvalid` | 8 | 1 | killed |
| M13 | `resolveOverlay`: path-form gate admits a `directory` | `envprofile TestPathOverlayFormIsSourceInvalid` | 5 | 1 | killed |
| M14 | `compose add`: path-form gate admits exactly `--revision` | `curator TestProfileComposeAddRefusesAFormOnAPathSource` | 15 | 1 | killed at `path_with_revision`, all 13 siblings PASS |
| M15 | `compose add`: git-form gate admits a form-free git source | same | 15 | 1 | killed |
| M16 | canonicalization: restores the single-letter-host carve-out | `identity TestParseOneCharacterHostIsANetworkIdentity` | 1 | 1 | killed |
| M17 | rendering: unconditional `revision` key on a path overlay | `config TestManagerConfigV2Vectors/schema2-overlay-path-source` | 2 | 1 | killed |

Every one is a genuine narrowing — the gate stays present and admits exactly one member.
**No delete-only mutant in the table**, so this is not a mutant table wearing a costume.
M14 is the cleanest specimen: the narrowed member fails while its 13 siblings pass.

**Both token-preserving text-gate mutants reproduce.** `ledger-consistency.sh` is the
source-text-inspecting gate this change touches, and both attacks keep the searched-for
token and change behaviour, with the behavioural suite executed for each:

| | preserves | changes | `ledger-consistency.sh` | `go test ./internal/identity` |
|---|---|---|---|---|
| T1 | the ledger row's `TestClassifySourceMatrix` token, verbatim | the Go case is renamed | **exit 1** — "required on darwin/windows but not compiled into that build" | **exit 0 (green)** |
| T2 | the token and the case | `must_run_on` narrowed to `darwin` | **exit 1** — "compiled into the linux/windows build but undeclared there" | **exit 0 (green)** |

A harness that ran only the static checker, or only the suite, would miss one of these. The
producer's ran both. Correct.

### 2.5 The candidate lane — measured, not argued

I confirmed it from the jobs' own logs rather than the report's sentence. Run
[34071813375](https://github.com/relux-works/curator/actions/runs/34071813375),
`workflow_dispatch`, head `38702164aee7a4658b15fc15966fec21dae61f8b`, conclusion success.

| runner | job | `CANDIDATE_REF` / `candidate_revision` | `CI_REQUIRE_FULL_ROOT` | suite-plan | `test-gate: stage served` | wall |
|---|---|---|---|---|---|---|
| ubuntu-latest | 101590439643 | `87a0d0060bad…` / `87a0d0060bad…` | `1` | served=71 deferred=0 excluded=1 | exit=0 | 3m38s |
| macos-latest | 101590439723 | `87a0d0060bad…` / `87a0d0060bad…` | `1` | served=72 deferred=0 excluded=0 | exit=0 | 8m7s |
| windows-latest | 101590439718 | `87a0d0060bad…` / `87a0d0060bad…` | `1` | served=72 deferred=0 excluded=0 | exit=0 | 32m17s |

`internal/config :: TestManagerConfigV2SchemaCases` and `TestManagerConfigV2Vectors` are
`ok` on all three; `TestPathOverlayDeclarationParses` and `TestGitOverlayDeclarationParses`
are `ok` on macos and windows. The ubuntu `excluded=1` is `internal/godriver`, excluded by
the root's own `vectors/conformance-claim-v3-qualification.json` — a designed platform
exclusion, and the producer recorded it rather than smoothing it over. PR #62 at the same
head reports 11 SUCCESS checks with `Candidate suite` SKIPPED, as expected for a
`pull_request` event.

**The dispatch measures what the row asks**, and the producer's earlier refusal to infer
linux and windows from a darwin run was the right call.

**The 30 subcases reproduce.** I materialized the candidate root as a plain checkout
(never `git archive` — see §2.6) and verified it file-by-file against
`conformance/v1/manifest.json`: **1047 of 1047 match, 0 mismatches, 0 missing**. Against that
root, `go test ./internal/config` at base `7c74a492` fails with exactly **30 failing
subcases** (15 in `TestManagerConfigV2SchemaCases`, 15 in `TestManagerConfigV2Vectors`) and
at head `38702164` is `ok`. The task's premise and its closure are both confirmed by
execution.

### 2.6 Both reported anomalies are real

**(a) `git archive` corrupts a materialized conformance root.**
`git check-attr -a conformance/v1/fixtures/byte-exact/subst.txt` reports `export-subst: set`
from the fixture's nested `.gitattributes`. `git archive HEAD <that path>` yields **65 bytes**
against **40 bytes** on disk, differing from char 9 — the exact numbers the report gives.
Any root materialized with `git archive` is silently wrong. Confirmed.

**(b) `go test -run` splits on unbracketed slashes and exits 0.**
`go test -run '^(TestClassifySourceMatrix/plain)$'` prints
`testing: warning: no tests to run` / `ok … [no tests to run]` and **exits 0** — Go splits the
pattern on `/` into `^(TestClassifySourceMatrix` and `plain)$`, and does not error on either
unbalanced fragment. A mutant harness keying on exit code alone reports that as a survivor.
Confirmed; the producer's fix (count `=== RUN` and report NOT-RUN) is the right one, and I
used the same guard on every mutant above.

**Open question for the epic, reported as unknown.** The review brief asks whether any
*earlier* mutant result in this epic could be a false survivor for the same reason. I could
not establish it either way: the earlier sweep artifacts on the board
(`TASK-260906-2x4s7i_mutant-sweep-1..8`, and the stage (c) sweeps) record mutant names and
verdicts but **do not record the literal `-run` patterns used**, so the filter shape is not
recoverable from the evidence. I report this as **unknown, not clear** — the detection
criterion for whoever re-checks is: any sweep whose harness anchored a whole `Parent/child`
filter as one regexp and keyed "survived" on exit code without checking that any test ran.

### 2.7 The retired bound and the ledger

Stage (c)'s bound is genuinely gone: `grep -rniE "19gjyw|consumption gap|until the spec fix
lands|no operator surface can declare"` over `cmd internal .github docs` returns nothing
related. The `default: form = "path"` branch in `cmdComposeList` is live and asserted by
`TestProfileComposeAddPathRow` (`\tpath\t` in the list output).

Rows 303, 304, 305 and 376 now describe what their tests assert — I checked each against the
test body, including 304's new "and survives update" clause (`UpdateWithPolicy` is driven and
the member is asserted to survive) and 303's "all four Windows drive spellings" (the test
drives 8 path spellings including all four). `ledger-consistency.sh <evidence-dir>` reports
**225 rows checked across linux darwin windows, ok**. The one exception is the new row named
in **F1**.

### 2.8 Gates re-run here (darwin, sequential, each its own process)

| gate | command | exit |
|---|---|---|
| build | `go build ./...` | 0 |
| vet | `go vet ./...` | 0 |
| gofmt | `gofmt -l cmd internal` (no paths listed) | 0 |
| lint | `golangci-lint run ./...` — 0 issues | 0 |
| gate selftest | `gate-selftest.sh` — 94 passed, 0 failed | 0 |
| suppression | `no-broad-suppression.sh` — ok | 0 |
| ledger | `ledger-consistency.sh` — 225 rows | 0 |
| config vs candidate root | `go test ./internal/config` | 0 |
| identity | `go test ./internal/identity` | 0 |
| envprofile | `go test ./internal/envprofile` | 0 |

### 2.9 Accepted stated bounds

* **`branch` is refused as an unknown field, not as `profile_source_invalid`.** AC row 3
  names `branch`, but `$defs/overlay` has no `branch` property and
  `additionalProperties: false`, and `OverlayDeclaration` has no field to carry one, so the
  declaration cannot reach resolution and the resolution diagnostic is unreachable by
  construction. The refusal exists at the reader and is pinned
  (`TestSchema2KnobRejections/overlay path branch`). Sound bound; accepted.
* **No one-character-host overlay is driven through a live clone.** The fixture reaches git
  through `url.<base>.insteadOf`, and a `c:`-prefixed remote is the spelling git itself may
  read as drive-relative on Windows, which would make the row a property of the fixture.
  Pinned where portable instead. Sound bound; accepted.
* **The transcription is pinned by the corpus, not by the schema text**, because the
  conformance root publishes `conformance/v1` only. Correctly stated. My three-engine
  differential above closes that gap **for this review**, not for the repository's own CI —
  it remains a real bound there.
* **Trimming asymmetry** between `installOperandKind` (trims) and the reader (does not).
  Correctly stated and matches the schema, which does not trim. I probed it: a leading space
  makes `" git@github.com:example/x"` a **path** overlay source in both the schema and the
  Go helper (the space is inside the excluded class, so no arm fires), which is the schema's
  own decision and carries no allowlist consequence.
* Windows cannot host a colon in a path component, so
  `TestClassifySourceNeverProbesTheFilesystem` plants nothing there. The test logs the
  planted count and asserts the verdicts either way, and the bound is in the safe direction.
  Correct.

---

## 3. What to do

Three repairs, none of them in product code:

1. **F1** — make the property loop check the class it names (filter on `ValidCanonical` as
   well as a non-empty `Parse`), so that M10 kills it; then correct the M10 row in the report
   and the `platform-cases.tsv` row's description.
2. **F2** — add the colon-in-a-later-segment row to §3's table and one operand to
   `installOperandCases`.
3. **F3** — split the doc comment so `installOperandKind` carries its own.

Re-run only what the delta touches: `go test ./internal/envprofile`, `ledger-consistency.sh`,
and M10 against the corrected test. The hosted candidate lane does not need re-dispatching
for a test-filter and comment change, but say so explicitly rather than leaving it implied.
