# TASK-261016-i5gwaf — review revision 1 response

Ready for review. Both P2 evidence-attestation bypasses are repaired within
M2's evidence module. The vendor result is unchanged: **exec loader exclusion
PASS; interactive skill exclusion FAILURE on both native-root representations;
§7.9 interactive release gate pending.** No runtime constant, gate or store
policy changed. Model-visible behavior and refresh remain unmeasured.

## Review findings and regression witnesses

| Finding | Correction / production call sites | Named regression |
|---|---|---|
| `contradictory-loader-attestation` (P2, `repeat-of: none`) | `checkCounts`, called by real matrix, interactive regression and workspace replay, consumes quoted fields and requires exactly one event, outcome and each requested count. Missing, duplicate, malformed, overflowing and conflicting observations refuse. | Reviewer's original `TestReviewContradictoryObservation` name retained, both original cases preserved; reverse order, identical duplicates, event duplicate, quoted decoy, invalid and overflow counts added beside a valid zero-skill control. |
| `snapshot-link-attestation` (P2, `repeat-of: none`) | Reference CLI inventories actual physical home entries using no-follow classification before the published bytes-only `check_bundle`. `checkFrozenSnapshot` validates actual source fixture and is called by `TestFrozenSnapshot` and the real vendor snapshot lane before copying. | `TestFrozenSnapshot/listed-symlink`, `/unlisted-dangling-link`, `TestFrozenSnapshotSourceInputs` and `TestReferenceSnapshotPhysicalClosure` each pair malformed inputs with a valid regular fixture and assert unchanged rejected input. Source-input tests execute the original TestFrozenSnapshot entrypoint in a child test process; no pristine reconstruction can hide an invalid source. |

All **7 original candidate test names** remain, with their vendor observations
unchanged. The reviewer's previously scratch-only named test is now candidate
source under its original name. No previously committed reviewer test existed
to remove. Source fixture mode bits alone are waived in Go because Git cannot
retain 0600; the runtime private copies and independent reference CLI require
0600. The fixtures are frozen disposable inputs; no concurrent-hostile-write,
export provenance or future materialize integration guarantee is claimed.

The two regression commands were red before the respective fixes: exit **1**
for duplicate loader fields and exit **1** for each physical-link CLI attack,
with the regular positive control green. Their logs are retained alongside
this report. One earlier reference-test setup compile attempt exited **1**
(`no new variables on left side of :=`); it was corrected before the actual
red reproduction and is not counted as a control kill.

## Identity and coverage

- Curator Story HEAD/checkpoint: `0d00a109ad1da77a5506747c89f24af9442eb341`;
  worktree remains uncommitted. Local `HEAD..main` is now **5** commits, so
  these are checkpoint measurements, not claims about the newer main.
- Normative curator-spec main re-read through GitHub at
  `d2a942cf17e4eb017c45475442da1456d8dfbe6c`; unchanged from the pinned spec.
  Read environments §§7.1 P1–P4, 7.9, 10.2, CIP-0017 implementation leaf 2,
  minimal probes and Muse v4 conformance. No normative conflict found; the
  explicit M2 boundary keeps credential/model rows out of this replay.
- Muse exact binary: `1.4.8-R6256.1`, SHA-256
  `5721763b4723f961709cd28f69aa049763965b46c4d4d70f883a0a90e5ddf39c`.
  Exact locator retained privately in ignored task-local `identity.json`;
  public commands use `$MUSE148_BINARY` per public-board hygiene.
- macOS 27.0 arm64, Go 1.25.5, tmux 3.6a; same frozen fixture digests and
  offline sandbox boundary as the canonical qualification record.
- Final replay source SHA-256: `78ecf817e292c2a7032bea08cf403dc9f2e86bd015eefdb3caf82abd02f94da1`.
- Reviewer regression source SHA-256: `bfedfd84093101834626e55476bb7361b80e8b9a955593da7f4cffc29f977103`.
- Reference physical checker SHA-256: `524a8a00d2ce68be6084d3c87ac39a2f7a051cf3b064216fc5e7f3cca50dc1e6`.

**6 of 6 in-scope AC observation rows driven; 5 of 6 meet the loader contract.**
The canonical record's coverage map still maps every supplied observation row
to its production entrypoint. The new [sanitized replay](observations-review-rework.json)
binds all 44 vendor calls to these exact source identities. Both independent
parser refusals exit 2 before loader; all 42 accepted calls exit 0. Green
regression witnesses record the interactive failure; they are not successful
qualification. Every call uses disposable HOME/XDG, auto-update disabled,
file backend, echo provider and the tested network/Keychain deny. No active
login, authentication command or real credential store was inspected.

Brief gap: no catalog-keyed surface table was supplied. Its one declared
surface, vendor-build evidence authorizing form-specific exclusion, is covered
by both named regressions and the narrowing mutants below. No supplied row is
silently omitted. Out-of-contract rows and authorizing clauses remain:

| Row | Acceptance clause authorizing the bound |
|---|---|
| Login/backend/refresh and model-visible foreign removal | M2 “Separate evidence” and no-model boundary |
| Model-visible workspace precedence | M2 workspace residual: loader order/count only, unknown model behavior |
| Export/capture/publication/destination admission/runtime-B | M2 frozen snapshot-fixture boundary |
| Actual `env materialize -> Muse` | M2 separate required SNAP-CLI integration when that command ships |
| Consumer refusal/constants and registry qualification edits | M1/M3 ownership; M2 does not edit them |

## Exact commands, exits and evidence reuse

Commands were standalone processes with direct log redirection and no tee.
Set `MUSE148_BINARY`, `MUSE148_CURATOR`, `MUSE148_RUN_ROOT` and private TMPDIR
as in the probe README; every replay/plant uses a new task-local run root.
Reference tests additionally set `MUSE148_SPEC` to the pinned checkout and
`MUSE148_SPEC_PYTHON` to its already-prepared executable. No dependency was
installed during this rework. Unless noted otherwise, logs are adjacent.

| Command | Exit | Evidence / result |
|---|---:|---|
| `go test ./.research/muse148 -run '^TestReviewContradictoryObservation$' -count=1 -v` before fix | **1** | `review-regression-before-rev2.log`; intended conflicting fields admitted |
| `go test ./.research/muse148 -run '^TestReferenceSnapshotPhysicalClosure$' -count=1 -v` before CLI fix | **1** | `reference-regression-before-rev2.log`; both bad CLI inputs exit 0, regular control passes |
| Focused `go test ./.research/muse148 -run '^(TestObservationChecker\|TestReviewContradictoryObservation\|TestFrozenSnapshot\|TestFrozenSnapshotSourceInputs\|TestReferenceSnapshotPhysicalClosure)$' -count=1 -v` | 0 | `focused-rev2-02.log`; before final switch-only lint/source-input mask cleanup; final full replay below includes all of these tests |
| First `go test ./.research/muse148 -count=1 -v -timeout=5m` | **1** | Private `replay-rev2-01.log`; snapshot regression timed out waiting for `/rules` after echo. Matrix passed; no green claim for this incomplete TUI case. No harness assertion weakened; replayed in a fresh home. |
| Final `go test ./.research/muse148 -count=1 -v -timeout=5m` | 0 | `replay-rev2-02.log`; all vendor cases and every new regression/reference test executed |
| `go build ./...` | 0 | Private `build-rev2-02.log` |
| `go vet ./.research/muse148` | 0 | Private `vet-rev2-02.log` |
| Initial `golangci-lint run ./.research/muse148/...` | **1** | Private `lint-rev2-01.log`; QF1003 replaced conditional with switch |
| `golangci-lint run ./.research/muse148/...` | 0 | Private `lint-rev2-02.log`, final source |
| `$MUSE148_SPEC_PYTHON -B .research/muse148/reference-snapshot-check.py $MUSE148_SPEC .research/muse148/testdata/snapshot` | 0 | `snapshot-reference-rev2.log` |
| Same reference CLI with `MUSE148_SNAPSHOT_PLANT=extra` | **1** | `snapshot-plant-rev2.log`; expected `ValueError: entry-set` |
| Exec omission `MUSE148_OMIT_EXCLUSION=exec go test ./.research/muse148 -run '^TestMuse148LoaderMatrix$/exec/(linked\|snapshot)/skills' -count=1 -v` | **1** | `plant-exec-rev2.log`; required-exclusion retained-skill failure on both roots, no-exclusion controls pass |
| Interactive omission `MUSE148_OMIT_EXCLUSION=interactive go test ./.research/muse148 -run '^TestMuse148LoaderMatrix$/interactive/(linked\|snapshot)/skills' -count=1 -v` | 0 | `plant-interactive-rev2.log`; explicit survivor on both roots, not protection |
| `go list ./...` | 0 | Private `default-packages-rev2.log`; hidden opt-in probe is absent |
| `git diff --exit-code 9e6b55f4128ac8e92f11398ce7a23089fcd8a72d -- cmd internal go.mod go.sum .github scripts Makefile` | 0 | Private `hosted-source-identity-rev2-03.log`; default source/tests/toolchain/config unchanged |
| `gh run view 38835589308 --json headSha,conclusion,status,url,jobs` | 0 | Fresh hosted read: completed/success, exact snapshot above, 20 success and 2 configured optional skips |
| `git diff --check` | 0 | Private `diff-check-rev2-01.log` |

**Full-gate evidence is reused, not freshly rerun:** `sh scripts/remote-gate.sh`
previous producer exit **0**, run
[38835589308](https://github.com/relux-works/curator/actions/runs/38835589308),
with byte-identical default source/tests/config and the same recorded hosted
execution environment. Its exact identity and green terminal result were
rechecked above. Per the standing reuse instruction, no full-suite replay was
justified by changes confined to the hidden opt-in evidence module and docs.
The default `go test ./...` suite and native probe are distinct: CI did not
execute this hidden module; the new full native replay did. Previous local
unfiltered default-suite failures/timeouts remain historical red/unverified
in the earlier handoff, not newly passing evidence. Optional hosted jobs
remain skipped, not executed. Reviewer run 38844928424 was cancelled and is
not reused as green evidence.

## Narrowing mutants

For each row, copy the final evidence module into a fresh ignored task-local
directory, apply the matching committed patch under `mutants/`, then run:
`go test ./.temp/TASK-261016-i5gwaf/mutants-rev2-final/<mutant> -run '<mask>' -count=1 -v`.
Reference mutants inherit the two reference locators; only the offline control
uses binary/Curator locators and a fresh run root, and it never launches Muse.
All eight patches applied cleanly to the final source. Each command exited
**1** at the named test, not at setup. Regular positive controls passed.

| Mutant | Gate narrowed to admit | Mask / named failing test | Bound / survivor |
|---|---|---|---|
| `checker-duplicate-count` | Exactly zero-then-one user-skills duplicate | `^TestReviewContradictoryObservation$`; `/count-zero-then-one` | Killed |
| `checker-duplicate-outcome` | Exactly failed-then-ready outcome duplicate | Same mask; `/outcome-failed-then-ready` | Killed |
| `snapshot-go-listed-link` | Follow only the listed root symlink | `^Test(FrozenSnapshot\|FrozenSnapshotSourceInputs)$`; both `/listed-symlink` tests | Killed; regular controls pass, source entrypoint returns 0 only under mutant |
| `snapshot-reference-listed-link` | Follow only the listed root symlink | `^TestReferenceSnapshotPhysicalClosure$`; `/listed-symlink` | Killed; real CLI improperly exits 0 under mutant |
| `snapshot-reference-dangling-link` | Omit only the synthetic unlisted dangling link | Same mask; `/unlisted-dangling-link` | Killed; real CLI improperly exits 0 under mutant |
| `checker-one-skill` | Exactly one skill accepted as zero | `^TestObservationChecker$` | Killed: `checker admitted one foreign skill` |
| `snapshot-extra-entry` | Ignore only `home/extra` | `^TestFrozenSnapshot$`; `/extra` | Killed: `extra snapshot admitted` |
| `offline-loopback` | Allow only outbound loopback TCP, retain other denies | `^TestMuse148OfflineBoundary$` | Killed: `network deny exit=0: MUSE148_NETWORK_CONTROL` |
| Exec omission | Retain synthetic foreign skill despite required exclusion | Matrix linked/snapshot `/skills/exclude-true` | Killed on both, exit 1 |
| Interactive omission | Retain skill with boolean omitted | No named failing test — **survivor**, exit 0 | Exact build retains skill with or without boolean on both roots; cannot establish interactive protection |

Mutant logs use `<mutant>-rev2-final.log` except `offline-mutant-rev2-final.log`;
omission logs are named above. All tests execute behavior; no source-token
scanner is added. The extra-entry reference input plant also exits 1 at the
published checker. No source gate was removed merely to claim detection.

## Scope and DoD

Rework diff is bounded to `.research/muse148/` (checker, regressions, patches,
sanitized evidence) and the existing `docs/qualification/` record. Outside
that module, `LOGBOOK.md` records these review findings as required by the
assignment; the prior README/UNRESOLVED_QUESTIONS additions remain unchanged.
No runtime source, test/configuration, qualification constant or credential
policy changes. No producer commit, branch switch or main integration.

- [x] Every supplied observation row remains mapped (6 of 6 driven), including both parser refusals and their accepted controls.
- [x] Root absent/present rules/skills/exclusions, both forms and linked/snapshot roots replayed through real loader.
- [x] Interactive-without-root fallback, workspace order/count residual and unknown model precedence retained.
- [x] Exact build/digest/platform, frozen bytes, channel, commands and vendor exits retained.
- [x] Real Curator link and independently validated physical regular snapshot closure exercised.
- [x] Both P2 findings have named regression tests, paired positive controls, unchanged rejected inputs and narrowing mutant kills.
- [x] Checker, omission, physical closure and offline-boundary controls executed; interactive survivor explicitly bounded.
- [x] All prior named tests and the reviewer test retained and green on final source.
- [x] Focused/full native suite, build, vet and lint green; all earlier red/setup attempts reported truthfully.
- [x] Full hosted property/default suites reused only after exact identity/green-result verification; no new default-suite execution claimed.
- [x] No catalog surface table supplied: brief gap declared; every supplied row mapped; out-of-contract rows and clauses listed.
- [x] Interactive vendor failure recorded on both roots; pending §7.9 gate retained, no successful exclusion or runtime admission claim.
- [x] Rework remains inside the named evidence module except justified documentation; candidate uncommitted for CR snapshot.
- [x] Outcome packet and qualification update attached before developer handoff; findings recorded in worktree logbook and board note.

The last item records the handoff precondition; the board resource attachments
and developer handoff acknowledgement are the authoritative delivery records.
