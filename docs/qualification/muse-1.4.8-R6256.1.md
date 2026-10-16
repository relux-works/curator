# Muse 1.4.8-R6256.1 loader qualification record

Task: TASK-261016-i5gwaf, M2, rc.7. This is a vendor observation record,
not runtime admission authority. **Exec loader rows pass; interactive P1–P4
skill exclusion FAILS on this build. The environments §7.9 release gate
remains pending for the interactive form.** Curator qualifies no runtime
itself; M1/M3 own registry/status and consumer changes. No runtime constants,
qualification gates or credential policy change in this leaf.

## Exact identity and observation boundary

| Identity | Value |
|---|---|
| Vendor version | `Muse Code 1.4.8 (1.4.8-R6256.1)` |
| Binary SHA-256 | `5721763b4723f961709cd28f69aa049763965b46c4d4d70f883a0a90e5ddf39c` |
| Binary locator | `$MUSE148_BINARY`; exact absolute path recorded privately in ignored task-local `identity.json` (public-board path hygiene) |
| Platform | macOS 27.0 arm64 |
| Curator checkpoint | `0d00a109ad1da77a5506747c89f24af9442eb341` |
| Normative curator-spec main | `d2a942cf17e4eb017c45475442da1456d8dfbe6c` |
| Harness | Go 1.25.5, tmux 3.6a, Apple sandbox-exec |
| Native root SHA-256 | `8c9af1ca52b75aaa1e67af6365f1e39dfcc3c56a274d3571270079ef669ac9d0` |
| Snapshot manifest SHA-256 | `924378006ba6e6e7db9d0277375456a958c14cbaff03d1ebd63dfa6e4b7ceac8` |
| Snapshot launch SHA-256 | `2991bc9c97904d77aef70bb5c3ea7a7044ff462382310ae613bbd31f461fd5f1` |

The binary locator is the separately scoped vendor executable, never the
updating launcher. The harness refuses a different binary digest. Its
`MUSE148_BINARY` environment gate is opt-in; ordinary `go test ./...` excludes
this hidden evidence package. An unset binary skips vendor tests and proves
no qualification.

Normative sources read: environments §§7.1 P1–P4, 7.9, 10.2; CIP-0017
implementation leaf 2 and minimal probe items 1–4; and
`conformance/environments-muse-v4/README.md`. Current main identity was
rechecked during rework. The 1.4.7 exclusion claim is not promoted to 1.4.8.
CIP credential probes are separate evidence under M2's explicit boundary.

Every vendor invocation runs with `MUSE_NO_AUTO_UPDATE=1`, `env -i`, disposable
HOME and XDG config/data/state/cache directories under task-local `.temp/`,
`TBH_CREDENTIAL_BACKEND=file`, and `--provider echo`. The tested OS sandbox
denies all network, security-service lookups and Keychain file access, plus
real native configuration/library directories. The offline boundary test
pairs an accessible synthetic loopback endpoint with a denied connection
(exit 7 and no second server hit), and readable ordinary bytes with a denied
synthetic Keychains file (exit 1, Operation not permitted, unchanged bytes).
No real Keychain item is inspected. No login, logout or auth command runs.

Echo reaches the real context loader without remote inference. Interactive
startup shows the skill before the echo turn; a local sentinel echo flushes
loader diagnostics, followed by `/rules` and `/exit`. This is pre-remote-
inference loader evidence, not proof of termination before a local echo turn.
No quota/auth failure is used as a network boundary. Credential files remain
absent in each vendor home and fixture source/root bytes remain unchanged.
These observations do not establish credential backend/login/refresh behavior.

## Fixtures, commands and production entrypoints

[Executable probe and fixtures](../../.research/muse148/README.md) contain all
synthetic fixture bytes. `testdata/source` is installed with the real Curator
`profile install SOURCE`, then `env resolve muse --profile muse148-probe
--repair --format json`. The resulting symlink must be a real link and match
the frozen root digest above. A synthetic `{}` auth target is used only for
Curator's v3 metadata check; Muse never runs in that fixture home.

The snapshot is a regular `0600` file at `home/config/muse/AGENTS.md` plus
CCJ-1 manifest/launch metadata declaring all four XDG suffixes. Its complete
single-file closure is independently checked with the published spec
`managed_home_snapshots.check_bundle`. Git does not retain private mode bits;
the harness writes private copies and the independent check sets mode 0600.
This frozen fixture proves representation/loader behavior, not export
provenance or the future `env materialize -> Muse` integration.

Each case invokes the exact real executable through sandbox-exec:

```text
exec:        $MUSE148_BINARY exec --json --provider echo --trust-workspace [--no-foreign-personal-context] MUSE148_OFFLINE_CONTROL
interactive: $MUSE148_BINARY --provider echo --trust-workspace [-c context.foreign_personal_skills=false]
```

Interactive uses a private background tmux terminal, never the active UI.
Parser refusals independently invoke interactive with the exec-only flag and
with `-c context.foreign_personal_skills="false"`. Both must exit **2**, with
`invalid TUI options` / `the JSON type does not fit this setting`, respectively,
and no rules-loader record. Accepted calls exit **0** and expose real records.

Observation channels: `rules.context_load`, `skills.catalog_load`, interactive
startup banner and `/rules`. [Sanitized observations](../../.research/muse148/observations.json)
preserve argv, vendor exit and loader records per case; raw paths stay private.
[Rework observations](../../.research/muse148/observations-rework.json) bind all
44 calls to replay source SHA-256 `f4110c8c614f791d89def4fc04232f411f920ab0cda37fe63d3dd3ae9bc85cdf`.
The first replay's [red test log](../../.research/muse148/replay-03.log)
remains historical evidence. The rework's green regression suite pins that
failure as observed behavior; it does not turn the qualification result green.

## Coverage map

**6 of 6 in-scope AC observation rows driven through production entrypoints;
5 of 6 satisfy the loader contract.** The supplied seventh observation row
expressly requires separate evidence. No catalog-keyed attack-surface table
was supplied: this is a brief gap. The declared surface is vendor-build
evidence authorizing form-specific exclusion; every supplied observation row
is mapped below. Green witness tests and vendor qualification are separate.

| AC observation | Named tests / production call site | Exec | Interactive | Narrowing control or bound |
|---|---|---|---|---|
| Parser acceptance/refusal | `TestMuse148ParserRefusals`, `TestMuse148LoaderMatrix`; real exec/TUI parser | Exact flag accepted | Boolean accepted; exec flag and string false rejected independently, exit 2, before loader | Independent malformed-type and wrong-form calls; no help-only inference |
| Root fallback and persistent skills | `TestMuse148LoaderMatrix`; real rules/skills loader | Observed | Observed | Each isolated foreign source with root absent/present and exclusion absent/present |
| Required exclusion with native root | Matrix plus `TestMuse148InteractiveExclusionRegression`; real skills loader/banner | PASS: 1 → 0 skill | **FAIL: 1 → 1 skill**, linked and regular-file roots | Exec omission killed; interactive omission cannot be distinguished from broken vendor exclusion |
| Boolean without native root | Matrix `interactive/absent`; real TUI loader | Exec's explicit exclusion removes fallback | Fallback remains; skill remains | Paired absent/present native root, separate Claude/Codex rules |
| Curator link and frozen snapshot identity | `TestFrozenSnapshot`, matrix; real Curator install/resolve and Muse loader | Both load one native user source | Both load one native user source | Altered bytes, extra entry, link rejected; independent published snapshot checker |
| Workspace conflict | `TestMuse148WorkspaceResidual`; real loader and TUI `/rules` | Native and workspace sources counted | User then project listed; project count 1 → 0 on subsequent reload | Enumeration only; model-visible precedence remains unknown |
| Credentials/model removal | Not driven | Not measured | Not measured | Out of contract: M2 “Separate evidence” |

The matrix runs 36 calls (2 forms × 3 root states × 3 separate foreign sources
× 2 exclusion states), parser 2, workspace 4, explicit regression witness 2:
**44 vendor entry cases** in the rework replay. Without exclusion a foreign
skill visibly loads, so zero counts cannot pass vacuously. All checker reads
require explicit successful records; missing, failed, truncated or
contradictory observations cannot attest absence.

Out-of-contract rows: credentials/login/refresh and model-visible removal
(M2 “Separate evidence”); model-visible workspace precedence (M2 workspace
unknown); capture locks, export provenance, publication, destination admission,
runtime-B and materialize integration (M2 snapshot-fixture clause); runtime
consumer refusals/constants (M1/M3 ownership). No consumer qualification is
inferred from vendor tests or borrowed credential measurements.

## Validation and release effect

Reproduction commands and per-command exit codes are in the
[task handoff](../../.research/muse148/TASK-261016-i5gwaf_outcome.md).
The original interactive omission is a non-discriminating survivor: it cannot
prove exclusion where the real build already retains the skill. The checker
still rejects that skill when asked to attest zero. No exclusion or consumer
gate is weakened. Keep the §7.9 interactive release gate pending on this build;
D10 credential-mode admission remains independent. A corrected vendor build
requires a new exact-build record and replay, coordinated with M1/M3.

## Review revision 1: attestation corrections and replay

The review found two P2 evidence bypasses, rather than a different vendor
result: conflicting duplicate loader fields were accepted, and the reference
fixture wrapper followed listed symlinks while dropping unlisted dangling
links. Both are repaired in the evidence module. `checkCounts` now requires
one unambiguous event/outcome/count field; the physical snapshot inventory is
classified without following links before calling the published bytes-only
checker. The actual frozen source is checked before creating private copies.
Git source mode bits alone are waived in the Go check; private copies and the
independent reference check still require 0600. These frozen-fixture checks
make no concurrent-hostile-write or materialize-provenance claim.

The original reviewer test name `TestReviewContradictoryObservation` is kept,
with both original cases, their reversals and neighboring malformed/duplicate
cases. `TestFrozenSnapshotSourceInputs` drives the original frozen-fixture
test entrypoint against physical inputs; `TestReferenceSnapshotPhysicalClosure`
drives the real reference CLI. Listed-symlink and unlisted-dangling-link
refusals have regular-file controls and unchanged-input assertions. Five new
narrowing mutants plus the three retained controls all fail at their named
tests; interactive omission remains a declared survivor.

[Review rework evidence](../../.research/muse148/TASK-261016-i5gwaf_review-rework.md)
records exact commands, exits, regression-before-fix failures, mutant table,
scope, AC map and hosted-gate reuse identity. The final full opt-in suite
exited **0** and replayed all **44** real vendor calls; the first attempt's
missing TUI `/rules` observation remains a reported failure, not qualification.
[Current sanitized observations](../../.research/muse148/observations-review-rework.json)
bind the replay to source SHA-256
`78ecf817e292c2a7032bea08cf403dc9f2e86bd015eefdb3caf82abd02f94da1`.
The checkpoint remains `0d00a109ad1da77a5506747c89f24af9442eb341` and is now
five commits behind local main; these measurements qualify only the stated
checkpoint/build/fixtures, not later source. Default runtime/tests/config
remain byte-identical to the earlier green hosted gate; the hidden probe was
separately rerun. No runtime constant or admission gate changed. **Exec loader
rows pass; interactive skill exclusion still FAILS on both roots; §7.9
interactive release gate stays pending.** Model-visible behavior and refresh
remain unmeasured.
