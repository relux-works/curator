# TASK-260916-1qfpu4 — revision 1 review

Verdict: **changes_requested**; route to **to-dev**. No candidate code or repository files edited. CR `CR-TASK-260916-1qfpu4-1`, revision 1, contains the curator LOGBOOK handoff entry; the normative candidate is the separate curator-spec Story worktree and attached spec patch, as expressly directed by the review brief. No merge/integration claimed.

## Candidate identity

Spec worktree: `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260916-73a5zg/worktree`.
HEAD: `684c9f1324d46b4938b2e5943f20c89e27971ec8`.
Attached patch and `git diff HEAD` are byte-equal; stable patch-id for both: `24cdcd76759a378ff0ab1c20b6a337b4b269965f`.
Patch SHA256: `054ec2e3687068a175af111e6e3daa7a3167fef767f45d5f2fe98ac0af0d29e7`.
Read producer rules, producer brief, evidence and patch, then audit E5 and Appendix B. Used project-management skill and its negative-evidence contract. No architecture diagram work applies.

## Required corrections

### F1 — Required scenarios are not pinned to discriminating inputs (high)

`tools/validate.py:6351` checks only the case-name inventory, then calls the same generic model for every case. The name only enters an error label at `tools/validate.py:6213`. An internally consistent passing case can replace each required negative or repair case without rejection. This directly violates producer rule 7 and the round-specific review requirement.

Independent probe: for each case, copy `materialize-clean-path-written`, preserve the original name, and invoke `validate_environments_write_nofollow_vectors`. All nine altered scenarios pass; the unchanged clean control also passes. Measured scenario-preservation rejection coverage: **0/9**. These are narrowing replacements, not deletions. The eight published negative tests (`tools/test_validate.py:2842` onward) change expected outputs, fixture digests or inventory; none makes the required internally consistent scenario replacement.

Correction: bind each required scenario to its operation, authorization, marker ownership, target kind/link destination, parent-link state and other discriminating inputs, with tests that substitute internally consistent alternative cases under each retained name. Retain semantic output checks too. Prove replacement rejection through the validator entry point, not merely inventory checks.

### F2 — Backup destination target-link branch contradicts the normative rule (medium)

`protocol/environments.md:1661` says the same nofollow refusal applies when a backup, marker or ledger destination “traverses or names a link the manager did not create in this operation.” But `tools/validate.py:6254` routes every symlink target through managed-surface ownership/foreign-manager logic, without distinguishing a backup write. The current backup vector (`conformance/v1/vectors/environments-write-nofollow.json:153`) has an absent target and a symlinked parent, so it exercises the generic parent traversal branch only.

Independent probe: retain the backup case, change `parent_link` to null and `target` to `{kind: symlink, link_owner: foreign, link_points: outside}`, keep the normative `environment_write_would_follow_link` refusal and untouched-target expectation. Validator rejects it: `environments-write-nofollow case backup-symlinked-destination-refused diagnostic does not follow the section 8.3.1 disposition`. Its model instead requires `environment_foreign_manager_detected`, contrary to §8.3.1.

Correction: model manager-private destination links distinctly and add a directly symlinked backup-target case with no parent symlink. Assert the nofollow diagnostic and untouched former target, and a negative test for the wrong foreign-manager diagnostic. Keep the existing parent refusal case. Ensure marker/ledger destinations share the correct private-destination rule; report explicitly what is vectorized versus text-only.

## Per-item brief review

All file:line references below are in the spec candidate unless explicitly marked curator.

| Requirement | Evidence / exact excerpt | Result |
|---|---|---|
| One rule; atomic entry replacement; RFC 2119 nofollow | environments:1626–1636: “Every materialization, takeover, repair, or backup write”; “operation-private name in the same directory, then renames over the target (atomic replace)”; “MUST NOT follow a symbolic link”; “`O_NOFOLLOW`-class semantics on open, `lstat`-class semantics on inspection.” | Text satisfies core rule. |
| Every mode, backups, marker and ledger | environments:1627–1629: “(`copied`, `linked`, `managed-home`), a section 8.3 backup, the section 8.2 marker, or the adapter ledger”. | Covered in canonical rule. |
| Target ownership and takeover | environments:1644–1654: “a marker-recorded surface entry”; “backup preserves the symlink with the same link text, never dereferenced”; “`environment_surface_unmanaged_conflict` unless a takeover authorization covers that path”. | Closed disposition; repair of planted link remains possible. |
| Non-operation-created parent link | environments:1656–1663: “no takeover flag authorizes traversal”; private destination “traverses or names a link”. | Text covered; F2 validator mismatch. |
| Materialization references §5/§7.5/§8.1 | environments:657: “Every write below follows the section 8.3.1 write discipline”; :935 “Its write is a section 8.3.1 write”; :1328 “existence check uses `lstat`-class semantics”; :1517 “Every write in that transaction is a section 8.3.1 write.” | Present. |
| Takeover §9.5 | environments:2080: “backed up as a symlink with the same link text, never dereferenced”; :2099 “Every takeover write is a section 8.3.1 write”. | Present; backup precedes replacement. |
| Drift §8.4 and repair §10.1 | environments:1681: “identified with `readlink`, never opened through the link”; :2273 “Repair writes are section 8.3.1 writes”. | Present. |
| One new diagnostic, closed vocabulary | environments:1701 and manager:2348 have identical `environment_write_would_follow_link`; environment prose, CHANGELOG, vector diagnostic inventory and validator use same spelling. Existing foreign-manager and unmanaged-conflict codes reused. | Justification is the previously unnamed parent/private destination traversal stop. No new schema field, CLI knob, marker field or lock key. |
| §12.1 knob / §12.2 lock rows | Producer evidence: “nothing to configure or lock”; unconditional discipline, no config change. | Not applicable; existing closed sets unchanged. |
| Posture reporting | environments:1667–1670: “`env status` reports every managed-surface, backup, and marker path ... the row is non-current”; :2650 includes “link-blocked”. | Reporting row present. Ledger reporting is not expressly enumerated; keep it consistent with the private-destination scope when revising F2. |
| Rollout and CHANGELOG | CHANGELOG:10 “E5: nofollow write discipline”; :28 “Direct rollout, under the hood”. | Direct rollout honored; warn-first revisions not applicable. |
| §13 conformance surface | environments:2865–2877 names `vectors/environments-write-nofollow.json`, both takeover authorizations, both parent cases, repairs and backup refusal, and “former target is byte-identical afterwards”. | Inventory present; enforcement fails F1/F2. |
| Positive/negative vector content | New file: takeover authorized/unauthorized; parent unauthorized/authorized; repair planted/manager-owned; backup parent; clean-path; inside-link conflict; recorded-file. Foreign cases carry fixture SHA256-after equal to recomputed before bytes. | Published cases express requested minimum branches. Structural expectations only; no manager execution claimed. F1 means branch preservation is not gated. |
| Manifest and rc.9 pins | manifest:4292 registers new family; release rc.9 changes only `candidate_protocol_pin.manifest_sha256` and `downstream_consumption.required_manifest_sha256`. | Regeneration checked in scratch copy; existing vectors 31/31 byte-identical to HEAD. |
| Scope / architecture | Eight changed spec files: environments, manager profile, CHANGELOG, vector, manifest, rc.9, validator and validator tests. No implementation, schemas, proposals 0014–0018, S5 or E7 changes. | Spec-harness edits are necessary under producer rule 7. Curator CR adds only the producer LOGBOOK entry (actual delta present, correcting stale “EMPTY” wording in review brief). |
| Evidence and acceptance criteria | Producer patch and evidence attached and read; spec revision not yet accepted or merged. | Corrections required before another reviewer cycle and producer integration. |

## Validation and adversarial evidence

Validation commands and observations are recorded below. Source candidate remains unchanged; mutation probes and regeneration run only in `/tmp/e5-review-candidate` or in-memory objects. Scratch Git index holds candidate content so `make regenerate-check` compares regenerated bytes against this uncommitted candidate rather than its old HEAD. No commits made.


The full `make validate` recipe was executed as its three component gates, with unittest discovery divided into contiguous batches of at most 100 tests to respect the headless bounded-call constraint. No aggregate `make validate` exit code is claimed. All gates are independently rerun; producer suite results are not substituted for these observations. Shell: zsh; Python commands use `PATH="/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin:$PATH"`. The batch harness uses the exact `unittest.defaultTestLoader.discover('tools', pattern='test_*.py')` inventory, flattens in discovery order, then runs slices 0:100, 100:200, 200:300, 300:400 via TextTestRunner; all 362 tests are covered once.

- `python3 tools/validate.py`: exit 0, `validated 62 schemas and 1094 vector files`.
- `go test ./tools/...`: exit 0, `ok github.com/relux-works/curator-spec/tools/generate-vectors (cached)`; Go reused its valid package cache, explicitly not a fresh uncached test execution.
- `make regenerate-check` in the candidate scratch copy: exit 0:
  ```text
  go run ./tools/generate-vectors -root .
  git diff --exit-code -- conformance/v1 release/1.0.0-rc.5.json release/1.0.0-rc.6.json release/1.0.0-rc.7.json release/1.0.0-rc.8.json release/1.0.0-rc.9.json
  ```
  This baseline regeneration proof ran before any scratch mutation. The actual candidate worktree was not regenerated or staged by the reviewer.

F1 real-entry reproduction: only in scratch, replace all case bodies with the clean-path body while preserving their names, then `make regenerate` to update manifest/release pins and `python3 -B tools/validate.py`. Both exit 0:
```text
go run ./tools/generate-vectors -root .
validated 62 schemas and 1094 vector files
```
The real gate calls `validate_environments_write_nofollow_vectors` from `main()` at `tools/validate.py:6725`. Thus F1 is an actual entry-point acceptance, not a disconnected helper observation. No filesystem safety behavior is proven by this spec-only harness; implementation verification remains TASK-260916-19shmj.

### Reproduction probe (run with candidate tools on sys.path)
```python
import copy
import validate
v = validate.load_json(validate.SUITE / "vectors/environments-write-nofollow.json")
clean = next(c for c in v["cases"] if c["name"] == "materialize-clean-path-written")
for i, case in enumerate(v["cases"]):
    changed = copy.deepcopy(v)
    changed["cases"][i] = dict(copy.deepcopy(clean), name=case["name"])
    validate.validate_environments_write_nofollow_vectors(changed) # all pass
changed = copy.deepcopy(v)
backup = next(c for c in changed["cases"] if c["operation"] == "backup")
backup["parent_link"] = None
backup["target"] = {"kind": "symlink", "link_owner": "foreign", "link_points": "outside"}
validate.validate_environments_write_nofollow_vectors(changed) # incorrectly rejects normative diagnostic
```

### Logbook entry — 2026-09-17

FINDING (TASK-260916-1qfpu4, E5 review rev1): a closed case-name inventory plus generic outcome derivation does not preserve scenario coverage. Nine of nine non-control scenarios can be replaced with a clean success and the production validation entry remains green after repinning. The backup target-link diagnostic also disagrees with §8.3.1. Route ordinary rework to to-dev; no external blocker or human decision is needed. This entry is persisted within the task-scoped review outcome; repository LOGBOOK files are left untouched under the explicit read-only review and campaign constraints.

## Completed unittest transcript

`python3 -B /tmp/e5-tests.py 0` — exit 0:
```text
Batch 0: 100 of 362 discovered tests
....................................................................................................
----------------------------------------------------------------------
Ran 100 tests in 154.469s

OK
```

`python3 -B /tmp/e5-tests.py 1` — exit 0:
```text
Batch 1: 100 of 362 discovered tests
....................................................................................................
----------------------------------------------------------------------
Ran 100 tests in 107.534s

OK
```

`python3 -B /tmp/e5-tests.py 2` — exit 0:
```text
Batch 2: 100 of 362 discovered tests
....................................................................................................
----------------------------------------------------------------------
Ran 100 tests in 166.360s

OK
```

`python3 -B /tmp/e5-tests.py 3` — exit 0:
```text
Batch 3: 62 of 362 discovered tests
..............................................................
----------------------------------------------------------------------
Ran 62 tests in 13.247s

OK
```

All 362/362 discovered Python tests passed, including all nine new WriteNofollowVectorTests. These greens do not invalidate F1/F2; the tests omit scenario substitutions and the direct backup-target branch. Final `git diff HEAD --check` exited 0, and final spec diff SHA256 remained `054ec2e3687068a175af111e6e3daa7a3167fef767f45d5f2fe98ac0af0d29e7`.

Run goal check: `task-board spawn goal "$TASK_BOARD_RUN_ID"` returned `Active Goal: none (run is not goal-bound)`. No directives were recorded. Verdict branch is changes_requested, with task status to-dev; no accept_cr or commit_ack is supplied.
