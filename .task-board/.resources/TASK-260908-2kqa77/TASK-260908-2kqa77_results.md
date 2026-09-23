# TASK-260908-2kqa77 results — goreleaser config value gate (rev4)

## Rev4: rework-2 — the awk walk is deleted, yaml.v3 enforces the values (this run)

Verdict on rev3: CHANGES_REQUESTED — R1 High (a first-position block
scalar or nested map forges a missing channel value while the gate
exits 0), R2 Medium (duplicate keys admitted against the declared
yaml.v3 contract), R3 Medium (the substring wiring pin accepts a
commented `run:` line and an `if: false` step). Per the rework-2
ruling ("stop extending the text walk") the check moved into Go:

- `tools/goreleaserconfig/gate.go` (new): `Check`/`CheckFile` parse
  with `gopkg.in/yaml.v3` and require every
  `homebrew_casks[*].skip_upload`, `scoops[*].skip_upload` and
  `release.prerelease` to be exactly the string `auto`,
  case-sensitively. Failure wording keeps the rev1 contract (field +
  entry index + observed value).
- `tools/goreleaserconfig/wiring.go` (new): `CheckWiring` parses
  `ci.yml` with yaml.v3 and requires exactly one live (no `if` key)
  lint step running `GateRun =
  "go test -count=1 ./tools/goreleaserconfig/"`.
- `tools/goreleaserconfig/gate_test.go` + `wiring_test.go` (new):
  table-driven executed rows (43 `=== RUN`, every fail row asserts
  the exact finding count plus the wording).
- `tools/goreleaserconfig/testdata/`: the reviewer's R1/R2 fixtures
  byte for byte (`block_first.yml`, `committed_block_bypass.yml`,
  `duplicate_good_last.yml`, copied from
  `TASK-260908-2kqa77_review-evidence-rev3.tar.gz`).
- `.github/workflows/ci.yml`: the lint step "Verify GoReleaser rc
  channel values" now runs the Go test (same step name, same job).
- `.github/ci/goreleaser-config-gate.sh`: DELETED (a rev1–rev3 working
  file, never on trunk).
- `.github/ci/gate-selftest.sh`: the awk-behavioural section is
  replaced with (a) a deletion pin (the `.sh` must stay absent) and
  (b) a stdlib-only python3 structural wiring pin with executed rows
  (the `gate-selftest` job has no `setup-go` step, so the value rows
  live in the Go test, which the lint lane executes).
- `CHANGELOG.md`: entry rewritten for the Go implementation.

Production call sites:

1. `lint` job, step "Verify GoReleaser rc channel values"
   (`run: go test -count=1 ./tools/goreleaserconfig/`) — every push/PR.
2. Every `go list ./...` lane (`test` on three OSes, rose-air, `race`)
   picks the package up automatically via `suite-plan.sh` — no
   per-lane list to rot.
3. `gate-selftest.sh` wiring pin (all three runners, python3 only).

No `.goreleaser.yml` semantic change, no `release.yml` change, no
ancestry-gate change, no goreleaser runs.

### R1 (High) — closed: the bypasses are impossible by construction

The walk is a real parser now: block scalar bodies are scalar text,
nested maps are nested maps, and only an entry-level `skip_upload`
key counts. Committed as executed rows (all fail, each naming
`homebrew_casks[0].skip_upload is absent`):

- R1a minimal — the verdict's inline fixture, verbatim
  (`testdata/block_first.yml`).
- R1a full config — the reviewer's `committed_block_bypass.yml`,
  verbatim (`testdata/committed_block_bypass.yml`).
- R1b — first-position `repository:` map with a nested `skip_upload`
  (the verdict's "replace `description: |` with `repository:`"
  variant, minimal form, inline).
- Kept: the nested-after-ordinary-key row (rev1 N1) still fails.

### R2 (Medium) — closed, with a real implementation finding

`yaml.Unmarshal` into a `yaml.Node` does NOT reject duplicate keys —
first wins. The first cut of this gate reported
`homebrew_casks[0].skip_upload = "true"` (pass-shaped single finding
on the wrong value... precisely: it found first-wins `true` and the
R2 row failed on wording) instead of rejecting the document. The
"already defined" rejection the ruling relies on happens only when
decoding into a map/`any`/struct — probed on this host:
`mapping key "x" already defined at line 2` for `any` and `map`,
nil error for `Node` (throwaway probe test, removed afterwards).
Both `Check` and `CheckWiring` therefore decode into `any` first as
the strict pass and walk the `Node` second. Guarded on every run by:

- `testdata/duplicate_good_last.yml` (verbatim: `skip_upload: true`
  then `skip_upload: auto`) → 1 finding, `invalid YAML` +
  `already defined`.
- wiring table row "duplicate run keys fail instead of resolving by
  first/last wins" → same strict-decoding failure.
- python pin: duplicate `run:` keys in one step disqualify it
  (`run_keys == 1` rule), with its own executed self-test row.

### R3 (Medium) — closed: the pin reads structure, twice

Two independent pins, same policy (exactly one live lint step: parsed
`run` equals the gate command, no `if` key, duplicates disqualify):

- Go `TestCheckWiring` + `TestCommittedWiring` (real YAML parse,
  runs in the lint lane): commented run, `if: false`, `if: true`
  (strict: any condition needs human review), deleted step, changed
  path, missing `jobs` each fail; quoted value and trailing comment
  pass as the same invocation.
- self-test python3 pin (stdlib only, runs on all three runners
  without Go): committed pass, commented / `if: false` / deleted /
  changed-path fail, quoted (double and single) and trailing-comment
  pass, unreadable workflow fails closed.

Cross-check on identical fixtures (self-test fixture generators run
in /tmp, both pins executed — all agree):

| fixture | Go pin | python pin |
|---|---|---|
| committed ci.yml | 0 findings | exit 0 |
| commented run line | 1 finding | exit 1 |
| if:false step | 1 finding | exit 1 |
| deleted step | 1 finding | exit 1 |
| changed path | 1 finding | exit 1 |
| quoted run value | 0 findings | exit 0 |

### Value mutant table (all executed Go subtests, exit 0 overall)

Fail rows (each asserts count + `field = "observed", want "auto"`):

row C both-absent (count 2), row C narrowed (count 1), row E `Auto`
cask, `Auto` scoop, `sometimes`, boolean `true`, quoted `"true"`,
prerelease `ato`, prerelease `Auto`, absent prerelease, R1a minimal,
R1a full-config, R1b repository-first, R2 duplicates, bad second
entry (`scoops[1]`), nested-after-ordinary-key, removed stanza, null
stanza, empty list stanza, tab indentation (`invalid YAML`), padded
`" auto "`, missing release section, missing file (`cannot read`).

Pass rows: minimal good document, committed `.goreleaser.yml`,
good second entry, unquoted `auto`, single-quoted `auto`, trailing
comments.

### gate-selftest pin (new shape)

Section `goreleaser rc channel values: the lint lane runs the Go
gate`: the deletion pin plus the python wiring pin (pass + wording,
commented + wording, if:false, deleted, changed-path, quoted,
single-quoted, trailing-comment, duplicate-run, unreadable): 13 rows,
all `ok`; overall `gate-selftest: 198 passed, 0 failed`, exit 0.

### Narrow test commands + exit codes (shell: bash on macOS)

- `go test -count=1 ./tools/goreleaserconfig/` → exit **0**
  (`ok`, 43 `=== RUN`, 0 `--- FAIL`).
- `bash .github/ci/gate-selftest.sh` → exit **0**
  (`gate-selftest: 198 passed, 0 failed`), all 13 goreleaser rows `ok`.
- `golangci-lint run ./tools/...` → exit **0**, `0 issues`
  (repo `.golangci.yml`, same default set the lint lane runs).
- `go vet ./tools/...` → exit **0**; `gofmt -l tools/ .github` →
  empty; `go build ./...` → ok; `bash -n gate-selftest.sh` → clean.
- `bash .github/ci/no-broad-suppression.sh` (CI form) and with
  `tools` → `ok` both (the one `//nolint:gosec` names its linter).
- `git diff --quiet .goreleaser.yml .github/workflows/release.yml`
  → clean (both untouched).
- Cross-check (throwaway, /tmp only): Go pin and python pin run on
  the six identical workflow fixtures — agree 6/6 (table above).

### Bounds and notes

- The python pin is the weaker instrument (no PyYAML on runners): it
  pins the jobs/lint/steps structure relevant to this gate, not full
  YAML. The Go pin (real parser) is authoritative and runs in the
  lint lane; both agree on all six cross-check fixtures above.
- The awk-portability source greps are gone with the awk gate, per
  the rework-2 bound note (they were a historical-regression
  tripwire; nothing to trip over now).
- `if: true` on the gate step fails the pin by design (strict
  no-`if` policy: any condition needs human review before it may
  carry a must-run gate).
- Not run: the full landing suite (runs once at handoff per campaign
  rules), `test-gate.sh` (needs a conformance root checkout; the new
  package takes no root-dependent path), hosted lanes (arbitrate at
  handoff).
- gawk/Git-Bash awk are out of the picture: no awk remains in this
  leaf (only POSIX `index()`-style awk one-liners generating
  fixtures in self-test).

The remainder of this file is the rev3–rev1 record, still accurate
except where this section supersedes it (shell gate → Go gate).

---
# TASK-260908-2kqa77 results — goreleaser config value gate (rev3)

## Rev3: rework-1 completion — awk-portability pin rows (this run)

State found: the worktree already held the rev2 tree (hyphen-last
`[+0-9-]` fix in `is_block`; rev2 hosted run 35725881034 all green per
the rev2 validation log). Rev2 repaired the defect but skipped rework-1
item 4 ("No fixture, self-test, workflow, or CHANGELOG change") — no
awk-portability self-test row existed. This revision adds it.

Change (2 files; no CHANGELOG/workflow change — the rev1 entries still
describe the gate):

1. `.github/ci/goreleaser-config-gate.sh`: reworded the `is_block`
   comment so the forbidden literal `[+-0-9]` appears nowhere in the
   gate source (code was already the fixed `[+0-9-]`; zero behavioural
   change).
2. `.github/ci/gate-selftest.sh`: two new rows after the `ci.yml` wiring
   pin:
   - `the gate has no mid-class hyphen range gawk rejects` — fails if
     the literal `[+-0-9]` is present in the gate source (`grep -qF`,
     portable on all three runners);
   - `the block detector keeps the hyphen-last class` —
     `assert_contains` for `[+0-9-]` in the gate source.

   Rationale: behaviour alone cannot pin this on macOS because BSD awk
   tolerates the bad class (rev1 was green on macOS). The source-level
   grep pin fails on every runner, Linux included, if the range
   regresses.

Bracket-expression audit (rework-1 item 1) — every `[...]` in the gate's
awk program:

- `[ \t\r]` (trim, x2), `[ \t]` (tab check, blank line, comment line,
  doc markers x2, dash lines x3, rest-strip) — no hyphen, no range.
- `[A-Za-z0-9_.-]` (split_key) — hyphen last; ranges A-Z/a-z/0-9
  explicit and valid.
- `[+0-9-]` (is_block) — hyphen last (the rev2 fix).
- `[ ]` (indent/dash handling) — single space, no range.
- `---` and `-` outside brackets (doc-marker, dash-line regexes) —
  literal dashes outside classes.

Divergence scan: no 3-argument `match` (only 2-arg `match` + `RLENGTH`,
POSIX), no `\x` escapes, no `gensub`/`asort`, no `length(array)`, no
`delete`, no `{n,m}` intervals, no `\s`/`\d`. `/dev/stderr` (x8) —
exercised green on ubuntu gawk, macOS BSD awk, and Windows Git-Bash awk
by the rev2 hosted run.

Rework-1 item 3 (the 8 failing rows / second-bug check): all four named
rows (`the prerelease failure names the observed value`, `the
second-entry failure names scoops[1]`, `a good second entry passes`,
`unquoted auto passes (same parsed string)`) plus siblings are `ok`
locally (log lines 222-246), and rev2's hosted gate-selftest jobs were
green on all three runners — the awk parse failure was the single root
cause; no second bug. The hosted lanes re-arbitrate once at this
handoff.

Awk implementations (rework-1 item 2): executed — BSD awk only
(`/usr/bin/awk`, macOS; `awk -W version` / `awk --posix` flags are
ignored, so no strict-POSIX mode was available). NOT executed — gawk
(not installed on this host), Windows Git-Bash awk. gawk-compat rests
on the portable hyphen-last form plus the full POSIX-only audit above;
the hosted ubuntu/windows/macOS gate-selftest jobs at handoff are the
arbiter.

Re-verification (shell `/bin/bash` 3.2 on macOS):

- `bash .github/ci/goreleaser-config-gate.sh` on the committed file →
  exit **0**.
- Required mutants re-probed directly (fixtures in
  `/tmp/grl-rev3.kjJmOV`, outside the repo): row C absent → 1, row E
  `"Auto"` → 1 (names `homebrew_casks[0].skip_upload = "Auto"`),
  `sometimes` → 1, `true` → 1, prerelease `"ato"` → 1 (names
  `release.prerelease = "ato"`).
- Pin negative proof: a regressed copy (sed `[+0-9-]`→`[+-0-9]`,
  `/tmp/grl-pinneg.M6EKiW/regressed-gate.sh`) trips the absence pin and
  drops the positive pin — both new rows are real negative tests.
- `bash .github/ci/gate-selftest.sh` → exit **0**,
  `gate-selftest: 209 passed, 0 failed` (207 + 2 new rows), all 24
  goreleaser rows `ok`, 0 `FAIL` lines (full log
  `/tmp/grl-rev3-selftest.log`).
- `bash -n` on both scripts: clean. `git status`: exactly 4 paths (new
  gate; modified selftest, `ci.yml`, CHANGELOG).
  `.goreleaser.yml`/`release.yml` untouched (verified via
  `git diff --quiet`). No Go sources touched, so `gofmt`/`vet`/
  `golangci-lint` surface is unchanged. `shellcheck` not installed
  (unverified, not claimed). No LOGBOOK.md edit (campaign rules forbid
  host-state writes outside the worktree).
- `ci.yml` wiring pin kept:
  `ci.yml invokes the goreleaser config gate exactly once` → `ok`.

The remainder of this file is the rev2 + rev1 record, still accurate.

---

# TASK-260908-2kqa77 results — goreleaser config value gate (rev2)

## Rev2: gawk-locale regex fix (changes-requested repair)

Rev1's landing validation failed 8 gate-selftest rows on ubuntu-latest
and windows-latest (macos green). Root cause: the awk block-scalar
detector in `.github/ci/goreleaser-config-gate.sh` used the character
class `[+-0-9]`, where the mid-class `-` forms a `+`..`0` range. BSD awk
accepts it; gawk in a UTF-8 locale rejects it at parse time:

- ubuntu: `awk: cmd. line:59: error: Invalid range end: /^[|>][+-0-9]*$/`
- windows (Git Bash gawk): `error: invalid range endpoint: ...` (same line)

Every gate invocation died before matching anything, so pass-rows
exited 1 and the `assert_contains` wording pins found only the awk
error. The `split_key` class `/^[A-Za-z0-9_.-]+$/` parses earlier in
the program and drew no complaint, so it was the only bad class.

Fix (one class, plus a comment pinning why): hyphen-last literal form
`/^[|>][+0-9-]*$/`, which denotes the identical set on every awk.
No fixture, self-test, workflow, or CHANGELOG change — the rev1
self-test table already covers the gate; it was the gate that could
not start.

Rev2 re-verification (shell `bash` on macOS darwin/arm64, BSD awk;
no gawk binary exists on this host, so gawk-compat rests on the
portable hyphen-last form plus the reasoning above):

- `bash .github/ci/goreleaser-config-gate.sh` on the committed file →
  exit **0** (`2 channel entries and release.prerelease are "auto"`).
- Direct mutant probes, each exit observed: row C absent-both → 1,
  row C narrowed (one key) → 1, row E `"Auto"` → 1 (names
  `homebrew_casks[0].skip_upload = "Auto"`), `sometimes` → 1,
  boolean `true` → 1, `prerelease: "ato"` → 1 (names
  `release.prerelease = "ato"`), `prerelease: |` block scalar → 1
  (`= "<multiline value>"`, proves the fixed class still detects
  `|`), unquoted `auto` → 0.
- `bash .github/ci/gate-selftest.sh` → exit **0**,
  `gate-selftest: 207 passed, 0 failed`; all 22 new goreleaser rows
  `ok` including the `ci.yml` wiring pin.
- `bash -n` on both scripts: clean. `git status`: exactly 4 paths
  (new `goreleaser-config-gate.sh`; modified `gate-selftest.sh`,
  `ci.yml`, `CHANGELOG.md`). No `.goreleaser.yml`, `release.yml`,
  ancestry-gate, or Go change. `shellcheck` not installed on this
  host (unverified, not claimed); no Go sources touched so
  `gofmt`/`vet`/`golangci-lint` surface is unchanged.

The remainder of this file is the rev1 record, still accurate.

---

# TASK-260908-2kqa77 results — goreleaser config value gate

## Design

New gate `.github/ci/goreleaser-config-gate.sh` (POSIX `sh`+`awk`, no new
runner binaries). It performs an indentation-aware structural walk of the
config (default `.goreleaser.yml`, overridable path argv for fixtures) and
compares each parsed scalar against the exact string `auto`,
case-sensitively:

- every `homebrew_casks[i].skip_upload` and `scoops[i].skip_upload`
- `release.prerelease`

Failure messages name field, entry index, and observed value, e.g.
`homebrew_casks[0].skip_upload = "Auto", want "auto"`, or
`scoops[0].skip_upload is absent, want "auto"`.

Field naming: the AC says `brews[*].skip_upload`, but the config declares
no `brews` stanza; the publishing surfaces are `homebrew_casks`, `scoops`,
`release` (cycle-3 verdict section 3). The gate checks the two real
stanzas, every entry of each; this mapping is documented in the gate
header.

Value semantics (parse-then-compare): quoting is YAML syntax, not value,
so unquoted `auto` passes; `Auto`, `sometimes`, boolean `true`, `ato`,
empty, and absent all fail. Fail-closed: unreadable file, missing
stanza, stanza with no entries, and tab indentation all exit 1. A
`skip_upload` nested deeper than the entry level (e.g. under
`repository:`) does not satisfy the entry. Duplicate keys take the last,
matching `gopkg.in/yaml.v3`. Block-scalar bodies (`footer: |`) are
skipped structurally. The awk is POSIX-only (verified: no 3-arg
`match`, no `\x` escapes) so it runs identically under BSD awk on
macos-latest and Git Bash on windows-latest.

Why shell+awk instead of a Go test under `tools/`: every other gate is a
`.sh` pinned the same way, the script runs with zero build on all three
self-test runners (the `gate-selftest` job has no `setup-go` step), and
an explicit `ci.yml` step makes the "runs on every push" wiring
auditable. A `tools/` Go test would additionally depend on
`suite-plan.sh` package planning.

Production call sites:
1. `lint` job in `.github/workflows/ci.yml` (new step "Verify
   GoReleaser rc channel values", exact line
   `run: bash .github/ci/goreleaser-config-gate.sh`) — runs on every
   push/PR; OS-independent static check, same home as the other static
   gates (`toolchain-identity`, `no-broad-suppression`).
2. `gate-selftest.sh` new section (behavioral mutant table + wiring pin
   asserting the exact `ci.yml` invocation count is 1) — runs on all
   three runners on every push.

No `.goreleaser.yml` semantic change, no `release.yml` change, no
ancestry-gate change, no goreleaser runs.

## Mutant table (executed, exit codes observed)

Direct gate probes (`bash .github/ci/goreleaser-config-gate.sh <fixture>`,
fixtures derived from the committed file; throwaway probe script kept at
`/tmp/grl-probe.sh`, outside the repo):

| # | Fixture | Exit | Observed message (abridged) |
|---|---|---|---|
| P0 | committed `.goreleaser.yml` | 0 | `2 channel entries and release.prerelease are "auto"` |
| C | row C: both `skip_upload` keys removed | 1 | `homebrew_casks[0]... is absent`, `scoops[0]... is absent` |
| C1 | row C narrowed: first key removed only | 1 | `homebrew_casks[0].skip_upload is absent` |
| E | row E: `"Auto"` in cask entry | 1 | `homebrew_casks[0].skip_upload = "Auto"` |
| E2 | `"Auto"` in scoop entry | 1 | `scoops[0].skip_upload = "Auto"` |
| M1 | `sometimes` (unquoted) | 1 | `= "sometimes"` |
| M2 | `true` (YAML boolean) | 1 | `= "true"` |
| M3 | `"true"` (quoted string) | 1 | `= "true"` |
| M4 | `prerelease: "ato"` | 1 | `release.prerelease = "ato"` |
| M5 | `prerelease: "Auto"` | 1 | `release.prerelease = "Auto"` |
| M6 | `prerelease` key removed | 1 | `release.prerelease is absent` |
| Q1 | unquoted `auto` (all three) | 0 | pass — same parsed string |
| Q2 | single-quoted `'auto'` | 0 | pass |
| Q3 | trailing `# comment` on value lines | 0 | pass |
| S1 | appended bad 2nd scoop entry | 1 | `scoops[1].skip_upload = "sometimes"` |
| S2 | appended good 2nd scoop entry | 0 | `3 channel entries ... are "auto"` |
| N1 | key moved under `repository:` (nested decoy) | 1 | `homebrew_casks[0]... is absent` |
| F1 | `homebrew_casks` stanza deleted | 1 | `section "homebrew_casks" is absent` |
| F2 | missing file | 1 | `cannot read ...` |
| T1 | `" auto "` (padded) | 1 | `= " auto "` (strict) |

Cross-check: `gopkg.in/yaml.v3` parses the committed file's three values
as the Go string `"auto"` (temp `go run` program, removed afterwards),
agreeing with the gate's PASS.

## gate-selftest pin

`gate-selftest.sh` gains section
`goreleaser-config-gate.sh: the rc channel values parse as exactly auto`
— 16 behavioral `assert`s (committed-file pass; rows C, C-narrowed,
E×2, sometimes, boolean-true, ato, prerelease-Auto, prerelease-absent,
second-entry bad/good, unquoted-auto pass, nested decoy,
stanza-removed, missing-file), 5 `assert_contains` pins on the failure
wording (field + entry + observed value), and 1 wiring pin (`ci.yml`
invokes the exact gate line exactly once). All fixtures are generated
from the committed `.goreleaser.yml` at self-test time, so they track
the file.

## Narrow test command + exit code

Shell: `bash` on macOS (darwin/arm64).

- `bash .github/ci/gate-selftest.sh` → exit **0**,
  `gate-selftest: 207 passed, 0 failed`, all 22 new goreleaser rows
  `ok` (log lines 222–244, full log at `/tmp/grl-selftest.log` on the
  producer host).
- Direct gate probes per the mutant table above: every row's exit code
  observed as listed (all fail-rows exit 1, all pass-rows exit 0).

## Other validation

- `bash -n` on the new gate and on `gate-selftest.sh`: clean.
- `ci.yml` / `release.yml` re-parsed as YAML via `gopkg.in/yaml.v3`
  (temp program): both valid.
- `git status`: exactly 4 paths changed/added
  (`goreleaser-config-gate.sh` new; `gate-selftest.sh`, `ci.yml`,
  `CHANGELOG.md` modified). No Go sources touched, so `gofmt`/`vet`/
  `golangci-lint` surface is unchanged; `shellcheck` is not installed on
  this host.
- CHANGELOG: Unreleased/Added entry referencing TASK-260908-2kqa77.

## Boundary notes

- The gate covers the channels the config declares. Adding a brand-new
  publisher stanza (e.g. `brews:`, `nfpms`-style asset surface beyond
  `release`) is a visible diff owned by release review, not by this
  gate — same division as the verdict's "three fields cover everything
  this config declares".
- Non-blocking by placement: `lint` job only, matrix lanes untouched;
  not rc-blocking (no `release.yml` change).


## Revision 6 (refresh)

Revision 5's content was ACCEPTED; integration refused because trunk advanced to
`fad881368b632f43a18b01681a8fc7def110cbae` and changed `.github/workflows/ci.yml`,
which this candidate also changes. I refreshed the same candidate onto that trunk.
No Go gate or test source changed in this revision.

### Candidate convergence

Before refresh, the accepted revision had 10 changed paths, matching its rev5
Change Request patch. I combined the incoming `ci.yml` pin update with the gate
step and unioned trunk's `CHANGELOG.md` entries with this task's entry, then ran:

- `task-board worktree refresh-candidate TASK-260908-2kqa77` → exit **0**;
  outcome `refresh_advanced`, trunk/reviewed-trunk/branch OID all
  `fad881368b632f43a18b01681a8fc7def110cbae`.
- The first post-refresh checkout view still contained old-base contents on
  unrelated paths. I restored those paths to the refreshed HEAD, preserving the
  accepted task patch and the two merged files. Final `git status --short` is
  exactly the same 10 task paths as rev5: three modified files
  (`.github/ci/gate-selftest.sh`, `.github/workflows/ci.yml`, `CHANGELOG.md`)
  plus the seven files under `tools/goreleaserconfig/`.

Per-file comparison with rev5:

- `.github/ci/gate-selftest.sh`: unchanged.
- `.github/workflows/ci.yml`: retains the single GoReleaser gate step; includes
  trunk's rc.12 pin `dced9b8317e0e8af79edf2d0539b32bd22b6c85b`.
- `CHANGELOG.md`: retains the GoReleaser entry and includes the incoming trunk
  entries via the configured union merge.
- `tools/goreleaserconfig/{gate.go,gate_test.go,wiring.go,wiring_test.go}` and
  all three YAML fixtures: unchanged from rev5.

### Regression row and narrowing mutant

The later Review Round Brief describes revision 5 as rejected, but the attached
`TASK-260908-2kqa77_review-verdict-rev5.md` is an ACCEPTED identity review; the
last content rejection was revision 3 and the rev4 verdict accepted its fixes.
To exercise the brief's named-test/mutant requirement against the actual
candidate, I ran the existing regression row
`TestCheck/row_C_narrowed:_one_entry_losing_its_key_is_already_a_publish_path`.
I temporarily changed the absent-`skip_upload` branch in `checkListStanza` to
ignore missing values. With that narrowing mutant, the named row failed as
required: `got 0 finding(s), want 1`, exit **1** (expected red). I restored
`gate.go`; its SHA-256 before and after is
`e0521681446083fb0611e46f3f3cab63069906a2f6b5fcc232ba457d55034fc7`.

### Refreshed-tree validation

Shell: bash on macOS.

- `set -o pipefail; go test -count=1 -v ./tools/goreleaserconfig/` → exit **0**.
  Executed row C (both absent), the narrowed row C, row E (`Auto`),
  `sometimes`, boolean `true`, prerelease `ato`, committed config, block
  scalar, nested-map, duplicate-key, bad-second-entry, and workflow wiring rows.
- `set -o pipefail; sh .github/ci/gate-selftest.sh` → exit **0**,
  `gate-selftest: 198 passed, 0 failed`; includes GoReleaser wiring rows.
- Narrowing mutant command
  `set -o pipefail; go test -count=1 -v ./tools/goreleaserconfig/ -run 'TestCheck/row_C_narrowed'`
  → exit **1**, expected red as described above.
- `set -o pipefail; golangci-lint run ./tools/...` → exit **0**, 0 issues.
- `set -o pipefail; go vet ./tools/goreleaserconfig/` → exit **0**.
- `set -o pipefail; go build ./tools/goreleaserconfig/` → exit **0**.
- Post-restore `set -o pipefail; go test -count=1 ./tools/goreleaserconfig/`
  → exit **0**.
- `git diff --quiet HEAD -- .goreleaser.yml .github/workflows/release.yml`
  → exit **0**; both remain unchanged.

The configured hosted landing gate is left to the developer handoff runtime and
will run once there; it was not run manually.
