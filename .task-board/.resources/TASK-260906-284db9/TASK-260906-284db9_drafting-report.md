# TASK-260906-284db9 — promote SPEC_PIN to the qualified released revision

Branch `chore/promote-spec-pin`, worktree `/Users/iv/Developer/ReluxWorks/.worktrees/curator-spec-pin`,
base `eca87fe38957ed08f4819836cc6fe12a3efd18dd`.

---

## 0. Where this work is, and why it is not in the Story workspace

The managed Story workspace for `STORY-260907-2bddfc` is a **curator-spec** worktree:

```
.temp/STORY-260907-2bddfc/worktree
repository_id  bound to https://github.com/relux-works/curator-spec
branch         task-board/story/STORY-260907-2bddfc
base           87a0d0060bad64ab883d007dcdf35df7485368bf   (curator-spec main)
```

This task changes the **curator** repository: `.github/workflows/ci.yml`,
`.github/ci/root-artifacts.tsv`, `.github/ci/platform-cases.tsv`,
`.github/ci/gate-selftest.sh`, `internal/interop/environments/`. None of those paths exists in
curator-spec, so the change cannot be made in the managed workspace at all. curator-spec shares
curator's board (`curator-spec/task-board.config.json` → `local.board_dir: ../curator/.task-board`),
which is how a curator task came to be provisioned a curator-spec workspace.

The work is therefore in the worktree the task **Scope** names by path and base. Base was verified
current, not stale:

| Check | Result |
| --- | --- |
| `git log --oneline eca87fe3..main` | empty — `main` (`a66eec88`) is an ancestor of the base |
| `git log --oneline main..eca87fe3` | 13 commits — the PR #63 / PR #64 stack, including `c8bc452f` "Make a dropped environments vector family fail the candidate lane" |

**Consequence for the handoff:** the completion guard snapshots the Story workspace, which is clean
and in the wrong repository, so it will record `repository_delta=empty`. The real delta is the
working tree of `chore/promote-spec-pin`. The orchestrator needs to re-provision the Story workspace
against curator, or take the branch directly.

---

## 1. The pin

| | old | new |
| --- | --- | --- |
| revision | `0ed5c691e9208eea52f21db2fc05e226ce3516fd` | `87a0d0060bad64ab883d007dcdf35df7485368bf` |
| named by | the rc.9 release commit on curator-spec main | signed released tag **`v1.0.0-rc.11`** |
| `conformance/v1/manifest.json` sha256 | `sha256:803918bf8672f76cf990985e51db213b826674cd5bb54fbf47731b8404b44403` | `sha256:0e195ecd26af2fcb5e5afb5c3fef0ebe60981888b057e8efc60c0b48cf584529` |
| `protocol_version` | 1.0.0-rc.9 | 1.0.0-rc.9 |
| files in the manifest | 691 | 1047 |

`release/1.0.0-rc.9.json` at curator-spec main states
`downstream_consumption.required_manifest_sha256 = sha256:0e195ecd26af2fcb5e5afb5c3f…`, which is
exactly the **new** pin's manifest. The old pin was `803918bf…`, i.e. curator was testing against a
root the protocol release record no longer names as required, with
`committed_release_pin_advanced: false` recording that the downstream pin had not caught up.

The `ci.yml` policy holds. `v1.0.0-rc.11` is an **annotated, signed** tag:

```
$ git -c gpg.ssh.allowedSignersFile=maintainers.allowed_signers tag -v v1.0.0-rc.11
Good "git" signature for oparin@me.com with ECDSA key SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM
object 87a0d0060bad64ab883d007dcdf35df7485368bf
```

The pin still carries the 40-hex revision, never the tag name: `gate-selftest.sh` reads `SPEC_PIN`
out of the workflow and runs `candidate-suite.sh verify-ref` on it, which refuses a branch, a tag,
`HEAD`, a short hash and uppercase hex.

### The rewritten comment

```yaml
  # The one immutable committed protocol-suite pin, referenced by every job.
  # It stays on the currently qualified released revision; promoting it is
  # owned by TASK-260720-38l1sy, after TASK-260720-25d05o qualifies the
  # release. This pin is the revision the signed released tag v1.0.0-rc.11
  # names on curator-spec: it publishes protocol 1.0.0-rc.9, whose
  # conformance/v1/manifest.json is
  # sha256:0e195ecd26af2fcb5e5afb5c3fef0ebe60981888b057e8efc60c0b48cf584529
  # -- the exact digest release/1.0.0-rc.9.json requires of a downstream
  # consumer in downstream_consumption.required_manifest_sha256. It claims no
  # later revision. A candidate suite never appears here -- it enters through
  # the non-default `candidate-conformance` job below and nowhere else.
  #
  # This root serves the WHOLE module: `suite-plan.sh` defers nothing against
  # it. The environments families and the schema-2 config families that the
  # previous pin did not publish now run with the root exported on every
  # default lane, instead of taking their unset-root path there.
  SPEC_PIN: 87a0d0060bad64ab883d007dcdf35df7485368bf
```

Every clause of the old comment was about the old pin and none of it survives: the "v1.0.0-rc.9
release commit on curator-spec main" description, the `803918bf…` digest, and the "byte-identical to
the schema-8 candidate 6001dc33" provenance clause are all replaced.

### Two other `ci.yml` comments that described the old world

* **`test` job header.** Said "packages whose vectors the pin does not publish run with
  `CURATOR_CONFORMANCE_ROOT` unset and are named in the report". Against this pin there are none;
  it now says the pin publishes every declared artefact so the deferred set is empty here, while
  keeping the mechanism description for a pin that ever stopped publishing one.
* **`candidate-conformance` job.** Said `CI_REQUIRE_FULL_ROOT=1` "concentrates the load: the three
  long-pole packages the default lane splits across its served and deferred invocations … all run
  inside the single served invocation here". The default lane no longer has a deferred invocation,
  so that difference no longer exists. The measured Windows-contention rationale for the 60m budget
  is kept verbatim, because that measurement is still why the number is what it is.

---

## 2. The suite plan, before and after

`suite-plan.sh` run against each root in turn, `GOOS=darwin`:

| root | served | deferred | excluded | deferred packages |
| --- | ---: | ---: | ---: | --- |
| old pin `0ed5c691` | 69 | **4** | 0 | `internal/config`, `internal/envfragment`, `internal/envmarker`, `internal/interop/environments` |
| new pin `87a0d006` | **73** | **0** | 0 | — |

`plan-deferred.txt` for the new root is empty. The four packages that used to run with
`CURATOR_CONFORMANCE_ROOT` unset now run with it exported, on every default lane, on all three
runners.

---

## 3. Every ledger row that lost a tolerance

A `root-unset` tolerance is legitimate only while `suite-plan.sh` can defer the package
(`skip-classes.tsv` gives the class policy `deferred-only`). Against the new pin those four packages
are never deferred, so their tolerance can never fire — and a tolerance that cannot fire states a
survivable skip the gate would in fact refuse, and would silently become live again if the pin ever
moved back.

**13 rows** lost `skip_allowed_on=linux,darwin,windows` / `class=root-unset`; every one keeps
`must_run_on=linux,darwin,windows`. Both columns went to `-`.

| package | test |
| --- | --- |
| `internal/interop/environments` | `TestConformanceSnapshotAcquisition` |
| `internal/interop/environments` | `TestConformanceContextVersions` |
| `internal/interop/environments` | `TestConformanceContextResolution` |
| `internal/interop/environments` | `TestConformanceEnvironmentsHeader` |
| `internal/interop/environments` | `TestConformanceEnvironmentsMonolithic` |
| `internal/interop/environments` | `TestConformanceContextDetectors` |
| `internal/interop/environments` | `TestConformanceEveryPathTheVectorsNameIsDeclared` |
| `internal/config` | `TestManagerConfigV2SchemaCases` |
| `internal/config` | `TestSystemConfigV2SchemaCases` |
| `internal/config` | `TestManagerConfigV2Vectors` |
| `internal/envmarker` | `TestParseAuthoritativeEnvMarkerSchemaCases` |
| `internal/envfragment` | `TestFragmentAuthoritativeSchemaCases` |
| `internal/envfragment` | `TestFragmentEmissionMatchesReference` |

The three schema drivers named in the acceptance criteria are `internal/config` (3 rows),
`internal/envfragment` (2 rows) and `internal/envmarker` (1 row). The acceptance criteria's wording
— "the two `root-artifacts.tsv` registrations", "the three ledger rows" — predates the registrations
that PR #63/#64 added; the criterion applied to the current tables is the Definition-of-Done
generalisation, "every other ledger row that loses a deferral loses its tolerance too".

### Behaviour text that went with them

Ten of those rows ended in a clause describing the deferral, e.g. "*; a root that publishes no such
family defers the package and records a root-unset skip*". That clause is now false of the row it
sits on — the row tolerates nothing — so it was removed from all ten. No new claim was put in its
place: the columns say the row tolerates no skip, and the section comment says why.

### A renamed test

`TestTheLedgerTolerationForThisPackageIsTheDeferredRootClass` asserted the opposite of the new
state: that every `TestConformance*` row for `internal/interop/environments` declares
`class=root-unset` on all three platforms. Kept as-is it would fail; deleted, nothing would guard the
columns. It is rewritten and renamed to
**`TestNoLedgerRowForThisPackageToleratesASkip`**, asserting `skip_allowed_on == "-"` **and**
`class == "-"` for every row of the package, plus the unchanged requirement that every
`TestConformance*` case in the package has a row. Its own ledger row was renamed and its behaviour
text rewritten with it.

Checking **both** columns is deliberate and is what the token-preserving mutant M2 below attacks:
`skip_allowed_on` alone decides whether `platform-case-gate.sh` enters its toleration branch, and a
row reading `linux,darwin,windows` with `class=-` tolerates a skip for *any* reason while leaving no
`root-unset` text for a reader or a grep to notice.

### Comments and doc that described a world that ended

* `platform-cases.tsv`, environments section — rewrote the "the only skip these rows tolerate is the
  `root-unset` one" paragraph. Also corrected a precision error it carried: `suite-plan.sh` defers on
  **any one** missing declared artefact, not only on "some of them".
* `platform-cases.tsv`, schema-8 section — the blanket "every row … tolerates exactly one skip class"
  is no longer true of the `internal/config` rows; the section now says which rows tolerate nothing
  and which keep their column, and why.
* `root-artifacts.tsv` header — records that the committed pin is now a fully-serving root, so the
  deferral shapes it documents are the candidate-lane and partial-root story rather than a
  description of the default lane. The rows themselves stay: they are what makes a pin or candidate
  that stopped publishing an artefact fail by name.
* `internal/interop/environments/suite_test.go` — `suiteRoot`'s doc said its skip "cannot fire
  unnoticed in a lane that requires a fully serving root". It now records that no hosted lane reaches
  it at all, and that the guard survives only for a local `go test ./...` with no root exported.

### Rows deliberately NOT changed — a stated bound

Seven `root-unset` rows remain, and against this pin they are unreachable too:

| package | rows |
| --- | ---: |
| `internal/skillspec` | 1 |
| `internal/marker` | 1 |
| `internal/moduleroots` | 1 |
| `internal/scriptpolicy` | 4 |

These packages' artefacts were published by the **previous** pin as well — `suite-plan.sh` against
`0ed5c691` deferred four packages and none of these — so they lose no deferral in this move and are
outside the producer brief's "remove the tolerated-skip columns from exactly the rows that lose
them". They are not harmless: after this change every `root-unset` row in the ledger is unreachable,
so a general "no served package may tolerate `root-unset`" gate cannot be stated without an
exception list, and the check added to `gate-selftest.sh` enumerates the four promoted packages
instead of deriving the set. **Recommended follow-up:** a separate task removing the remaining seven,
after which the gate can be generalised. The bound is stated in the `gate-selftest.sh` comment and in
the schema-8 section comment rather than left to be inferred.

The two `internal/godriver` `root-content` rows are a different class — that case guards its own read
— and are untouched.
---

## 4. What now holds this in place

Two committed gates, both on required CI jobs.

**`internal/interop/environments.TestNoLedgerRowForThisPackageToleratesASkip`** (rewritten, renamed).
Reads `.github/ci/platform-cases.tsv` as a committed file and requires, for every row naming the
package: `must_run_on == linux,darwin,windows`, `skip_allowed_on == "-"`, `class == "-"`. Then parses
the package's own syntax tree and requires every `TestConformance*` function to have a row, failing
with `no TestConformance case found` rather than passing vacuously if the scan breaks. It reports
`12 of 12 ledger row(s) for internal/interop/environments checked; 7 TestConformance case(s) matched
to a row`. Production call site: the ledger this test reads is the one `platform-case-gate.sh` and
`ledger-consistency.sh` enforce, and the test itself is a required row in that ledger, run by
`test-gate.sh` in the `Test` and `Race` jobs on all three runners.

**`.github/ci/gate-selftest.sh`, section "the packages this pin promoted tolerate no skip"** (new).
Run by the `Gate self-test` job on all three runners. Two halves:

* *Static.* For each of `internal/config`, `internal/envfragment`, `internal/envmarker`,
  `internal/interop/environments`: the shipped ledger must carry at least one row (else the check is
  reported as vacuous and fails), and no row may have `skip_allowed_on != "-"` or `class != "-"`. The
  failure message names each offending row with both columns.
* *Behavioural.* For each of linux/darwin/windows it builds the same otherwise-satisfying `go test
  -json` stream the neighbouring "shipped ledger is satisfiable" loop builds — every case the ledger
  requires on that runner, observed passing — and changes exactly one event:
  `TestConformanceContextVersions` **skips** with `CURATOR_CONFORMANCE_ROOT is not set`, with
  `CI_DEFERRED_PKGS=internal/interop/environments`. That is the precise combination the old columns
  made survivable. `platform-case-gate.sh` — the production gate, run as its own process — must exit
  1, and `skips-observed.tsv` must record the verdict `FATAL-not-tolerated` for that case, proving
  the refusal came from the ledger columns and not from the class policy.

  The exit code alone would not have proved anything, and the first version of this check was wrong
  for exactly that reason: a stream carrying only the skip exits 1 regardless, because every other
  required case is then missing. It was rewritten to run against a stream that would otherwise pass.

### Coverage, as a ratio

Acceptance-criteria rows driven through a production entry point by a named committed test or a
named gate run: **8 of 8**.

| # | AC row | Driven by | Production call site |
| --- | --- | --- | --- |
| 1 | SPEC_PIN names the rc.11 released revision, comment describes that pin | `gate-selftest.sh` → `candidate-suite.sh verify-ref "$PIN"` (pin read out of `ci.yml`) | `Gate self-test` job; the pin is what `actions/checkout` resolves in four jobs |
| 2 | The three schema drivers lose their tolerance, required on every lane | `gate-selftest.sh` static half (`internal/config`, `internal/envfragment`, `internal/envmarker`) | `Gate self-test` job |
| 3 | Every other row that loses a deferral loses its tolerance | same check + `TestNoLedgerRowForThisPackageToleratesASkip` for `internal/interop/environments` | `Gate self-test`, `Test`, `Race` |
| 4 | `suite-plan` reports deferred=0 for the four packages | `suite-plan.sh` against the rc.11 root, §2 above | run inside `test-gate.sh` on every lane |
| 5 | `test-gate` green against a materialised rc.11 root, sequential | `test-gate.sh`, §5 below | the `Test` and `Race` jobs |
| 6 | The candidate lane still fails closed on a missing environments family | `fail-closed.sh` → `suite-plan.sh` under `CI_REQUIRE_FULL_ROOT=1`, §6 below; plus the pre-existing synthetic loop in `gate-selftest.sh` | `candidate-conformance` job |
| 7 | `gate-selftest.sh` and `ledger-consistency.sh` green | run directly, §5 | `Gate self-test` and `Test` jobs |
| 8 | Roots materialised as plain checkouts verified against `manifest.json` | `verify-root.py`, §7 below | method, not shipped code |

**Stated bound:** rows 1 and 5 are proved on this host only. The pin's behaviour on the hosted
ubuntu/windows runners, and the `Race` lane's behaviour there, are the orchestrator's hosted
measurement. Nothing here claims a hosted result.
---

## 5. Gates

Every command run as its own process, unpiped, with its real exit code. All against the final tree,
on darwin/arm64, Go from `go.mod`. The three `test-gate` lanes were run **sequentially**, never
concurrently and never alongside one another.

| Gate | Exit | Result |
| --- | ---: | --- |
| `go build ./...` | 0 | — |
| `go vet ./...` | 0 | — |
| `gofmt -l cmd internal` | 0 | 0 files |
| `golangci-lint run ./...` | 0 | 0 issues |
| `bash .github/ci/gate-selftest.sh` | 0 | 140 passed, 0 failed; the 10 assertions this change adds are all in it (130 before) |
| `bash .github/ci/no-broad-suppression.sh` | 0 | ok |
| `bash .github/ci/ledger-consistency.sh` | 0 | 235 rows checked across linux darwin windows |
| `test-gate.sh` (default lane) | 0 | served=73 deferred=0 excluded=0; go test 0, platform-case gate 0 |
| `test-gate.sh` with `GO_TEST_FLAGS=-race` | 0 | served=73 deferred=0; go test 0, gate 0; 0 `DATA RACE` reports |
| `test-gate.sh` with `CI_REQUIRE_FULL_ROOT=1` (candidate-lane setting) | 0 | served=73 deferred=0; go test 0, gate 0 |
| `mutants.sh` | 0 | 4 mutants, 4 killed, 0 survivors |
| `fail-closed.sh` | 0 | 8 passed, 0 failed |

All three `test-gate` lanes ran **one** `go test` invocation — the `served` stage. There was no
`deferred` stage at all, on any of them; that stage is what this task removed from the default lane.
Each lane recorded 20 skips, every one classified, and **zero** of class `root-unset`.

### One red, and what it was

The first `-race` re-run exited **1** and is kept at `.temp/TASK-260906-284db9/gates/race-v2.log`.
It was not a test failure — the stream carries no `fail` event for any case. It is a Go build-cache
fault:

```
internal/snapshot/rename_noreplace_darwin.go:5:8: could not import golang.org/x/sys/unix
  (open /Users/iv/Library/Caches/go-build/8e/8e0c877aa2f6d648e3f5f0e777159b849e1244b95cb2c71a8f668db332f564fd-d:
   no such file or directory)
```

— a cache index entry pointing at a file that had been removed underneath it, which made 12 packages
`build-fail` and so made every case in them "never ran" to the ledger. `go build ./...` and
`go vet` on the three named packages were re-run immediately afterwards and both exited 0, and the
lane was re-run clean (`race-v3`, exit 0). Recorded rather than dropped; it is not a property of this
change.

---

## 6. The candidate lane

**Fail-closed, against the real rc.11 root.** `fail-closed.sh` copies the verified root, removes
exactly **one** declared environments artefact — the narrowest possible weakening — and runs
`suite-plan.sh` under `CI_REQUIRE_FULL_ROOT=1`, the candidate job's own setting. Every one of the
seven declared artefacts is load-bearing on its own:

| root | `CI_REQUIRE_FULL_ROOT=1` exit | names the package | names the artefact |
| --- | ---: | :---: | :---: |
| intact rc.11 | 0 (deferred=0) | — | — |
| minus `vectors/environments.json` | 1 | yes | yes |
| minus `vectors/context-versions.json` | 1 | yes | yes |
| minus `vectors/context-detectors.json` | 1 | yes | yes |
| minus `vectors/snapshot-acquisition.json` | 1 | yes | yes |
| minus `expected/environments` | 1 | yes | yes |
| minus `expected/byte-exact-snapshot_sha256.txt` | 1 | yes | yes |
| minus `fixtures/byte-exact` | 1 | yes | yes |

The guard PR #63 landed is intact and is now proved against the released root rather than only
against `gate-selftest.sh`'s synthetic one.

**Green, against the same revision.** `test-gate.sh` with `CI_REQUIRE_FULL_ROOT=1` against the rc.11
root exits 0 — see §5.

### A consequence worth flagging before anyone dispatches the lane

`candidate-suite.sh verify-ref` refuses a candidate revision equal to the committed pin:

```
$ SPEC_PIN=87a0d006… bash .github/ci/candidate-suite.sh verify-ref 87a0d006…
candidate-suite: candidate revision equals the committed released pin 87a0d006…;
                 a candidate run must not impersonate the qualified pin
exit=1
```

So after this promotion the `candidate-conformance` job **can no longer be dispatched against
`87a0d006` via `candidate_ref`** — by design, and correctly. Re-qualifying that exact suite now has
to go through the `candidate_root` input against a materialised root. That is a behaviour change of
the workflow's inputs caused by this pin move, and it is not a defect.

---

## 7. Mutants

Every mutant NARROWS a gate: the gate stays present and is weakened to admit exactly one member of
the class it must reject.

| Mutant | What it narrows the gate to | Named test that fails | Exit |
| --- | --- | --- | ---: |
| **M1** | one environments row regains `skip_allowed_on=linux,darwin,windows` / `class=root-unset` — admits exactly the deferred-root skip of `TestConformanceContextVersions` | `TestNoLedgerRowForThisPackageToleratesASkip`; `gate-selftest.sh` → `internal/interop/environments tolerates no skip…`, and all three `a root-unset skip of TestConformanceContextVersions fails an otherwise-passing <goos> stream, even deferred` + the three `the <goos> verdict is the ledger refusing it, not a class policy` | 1 / 1 |
| **M2** (token-preserving) | same row gets `skip_allowed_on=linux,darwin,windows` with `class=-`. **`root-unset` occurrences in the mutated row: 0** — a grep for the token sees a clean ledger, while the row now tolerates a skip for *any* reason, which is strictly wider than what M1 admits | identical set to M1 | 1 / 1 |
| **M3** | one `internal/config` schema-driver row regains its tolerance | `gate-selftest.sh` → `internal/config tolerates no skip in the shipped ledger (7 rows)` | 1 |
| **M4** | `SPEC_PIN` reverted to `0ed5c691` | `suite-plan.sh` against the root that pin names: `served=69 deferred=4`, deferring exactly `internal/config`, `internal/envfragment`, `internal/envmarker`, `internal/interop/environments`; under `CI_REQUIRE_FULL_ROOT=1` it exits 1 | 0 then 1 |

**Survivors: none.** Baseline before each mutant: the Go test exits 0 and `gate-selftest.sh` reports
140 passed, 0 failed.

M3 is deliberately *not* caught by `TestNoLedgerRowForThisPackageToleratesASkip` (exit 0): that test
is scoped to `internal/interop/environments` and says so. It is caught by name in `gate-selftest.sh`,
so the mutant is killed, not survived — but the scope boundary is recorded here rather than left to
be discovered.

M2 is the mutant the DoD asks for against a gate that inspects text: it preserves the searched-for
token (by removing it) while changing behaviour, and the harness runs the **behavioural** suite —
`platform-case-gate.sh` against a real `go test -json` stream — not only the static checker. Under
M2 the gate stops reporting `FATAL-not-tolerated` for the skip, which is exactly the behaviour
change the static check alone could not see.

### A false gate caught before it shipped

The first version of the behavioural check asserted only that `platform-case-gate.sh` exits 1 against
a stream containing the single skip. That assertion passes under M1 and M2 as well — a one-event
stream fails anyway, because every other required case is then missing. It proved nothing. It was
rewritten to run against an otherwise-passing stream and to assert the recorded verdict by name.

### A harness bug that damaged the working tree, and how it was handled

The first `mutants.sh` restored between mutants with `git checkout -- <paths>`. That restores the
**index**, not the snapshot, so it discarded the uncommitted change under test: `ci.yml` and
`platform-cases.tsv` reverted to `eca87fe3`, the mutations then failed to apply (the harness's own
assertions fired and said so), and the gates went red against the base tree. Both files were
re-applied and the harness rewritten to snapshot and restore by file copy, never git. The mutant
table above is from the rebuilt harness. The `test-gate` lanes that had run before that point were
re-run against the final tree; §5 reports only the re-runs.

---

## 8. Method — how the roots were materialised

Both roots are **plain checkouts**, never `git archive`:

```
git worktree add --detach <dest> <revision>
```

and each was verified **file by file** against its own `manifest.json` by
`.temp/TASK-260906-284db9/verify-root.py`, in both directions — every declared path present with a
matching sha256, and no undeclared file on disk:

| root | declared files | missing | hash mismatched | undeclared on disk | verdict |
| --- | ---: | ---: | ---: | ---: | --- |
| `0ed5c691` | 691 | 0 | 0 | 0 | ok |
| `87a0d006` | 1047 | 0 | 0 | 0 | ok |

The `export-subst` trap was checked explicitly. `conformance/v1/fixtures/byte-exact/subst.txt` in the
materialised rc.11 root:

```
40 bytes
commit: $Format:%H$
abbrev: $Format:%h$
sha256:ec9a6c8c260b3977dad2f1cba09414599879ac5c3e7353c3205a40d5fbe122bc   (== manifest)
```

40 bytes and unsubstituted — a `git archive` root would carry 65 with the placeholders expanded.

Which artefacts moved, per declared row — all 12 absent from the old root, all 12 present in the new:

| package | artefact | old | new |
| --- | --- | :---: | :---: |
| `internal/envfragment` | `schema-cases/launch-env-fragment-v1` | — | ✓ |
| `internal/envmarker` | `schema-cases/agent-environment-marker-v1` | — | ✓ |
| `internal/config` | `schema-cases/manager-config-v2` | — | ✓ |
| `internal/config` | `schema-cases/system-config-v2` | — | ✓ |
| `internal/config` | `vectors/manager-config-v2.json` | — | ✓ |
| `internal/interop/environments` | `vectors/environments.json` | — | ✓ |
| `internal/interop/environments` | `vectors/context-versions.json` | — | ✓ |
| `internal/interop/environments` | `vectors/context-detectors.json` | — | ✓ |
| `internal/interop/environments` | `vectors/snapshot-acquisition.json` | — | ✓ |
| `internal/interop/environments` | `expected/environments` | — | ✓ |
| `internal/interop/environments` | `expected/byte-exact-snapshot_sha256.txt` | — | ✓ |
| `internal/interop/environments` | `fixtures/byte-exact` | — | ✓ |

`-run` anchoring: no `Parent/child` filter was used anywhere. The mutant harness anchors a single
top level, `-run '^TestNoLedgerRowForThisPackageToleratesASkip$'`, and the environments package was
additionally run with `-v` and its `=== RUN` lines counted: **12 top-level cases run, 12 passed, 0
skipped**, including all 7 `TestConformance*` cases.

---

## 9. Files changed

| File | Change |
| --- | --- |
| `.github/workflows/ci.yml` | `SPEC_PIN` → `87a0d006…`; the pin comment fully rewritten; the `test` job header and the `candidate-conformance` load comment corrected where they described a deferral that no longer happens |
| `.github/ci/platform-cases.tsv` | 13 rows lost both tolerance columns and their deferral clause; environments and schema-8 section comments rewritten; one ledger row renamed with its behaviour text |
| `.github/ci/root-artifacts.tsv` | header records that the committed pin is now a fully-serving root, and that the deferral shapes it documents are the candidate/partial-root story |
| `.github/ci/gate-selftest.sh` | new section: static + behavioural check that the four promoted packages tolerate no skip (+7 assertions) |
| `internal/interop/environments/contract_test.go` | ledger test rewritten and renamed; both columns checked; vacuity guard on the row count; coverage logged as a ratio |
| `internal/interop/environments/suite_test.go` | `suiteRoot` doc corrected — no hosted lane reaches its skip any more |

No `LOGBOOK.md` entry was written, per the producer brief. The findings above are recorded here and
in the board notes instead.

---

## 10. Open items for the orchestrator

1. **The Story workspace is in the wrong repository** (§0). The delta is on `chore/promote-spec-pin`
   in `/Users/iv/Developer/ReluxWorks/.worktrees/curator-spec-pin`, uncommitted. Nothing was pushed
   and no PR was opened, per the brief.
2. **Hosted measurement is yours.** Everything here is darwin/arm64. The ubuntu and windows runners,
   and the hosted `Race` lane, are unproved by this report and nothing here claims otherwise.
3. **Seven `root-unset` ledger rows remain unreachable** (§3). Recommended as a follow-up task, after
   which the `gate-selftest.sh` check can drop its enumeration and derive the package set.
4. **Re-qualifying `87a0d006` through the candidate lane now needs `candidate_root`**, not
   `candidate_ref` (§6).

---

## 11. Delivery

Two signed commits on `chore/promote-spec-pin`, human identity, nothing pushed and no PR opened.

| Commit | Signature | Author | Subject |
| --- | :---: | --- | --- |
| `a1290fca` | G | Ivan Oparin \<oparin@me.com\> | Pin the released revision the protocol record actually requires |
| `25fc913a` | G | Ivan Oparin \<oparin@me.com\> | Retire the tolerances the deferral they existed for has ended |

The split is not cosmetic: the first commit is green on its own, because moving the pin makes the
four packages served and their cases pass while their now-unreachable tolerance is simply not yet
asserted against. The second removes it and rewrites the test that asserted the old state. Each
commit was checked out into a detached worktree of its own and verified there:

| Commit | `go build ./...` | `go vet ./...` | `ledger-consistency.sh` | `gate-selftest.sh` |
| --- | ---: | ---: | ---: | ---: |
| `a1290fca` | 0 | 0 | 0 | 0 |
| `25fc913a` | 0 | 0 | 0 | 0 |

The branch bisects.
