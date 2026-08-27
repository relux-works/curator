# TASK-260822-f4qv7w review verdict — RUN-260823-da95fc

**Verdict: accepted.** Content, architecture fit, determinism, and the delivery
gate are all independently reproduced below.

## 1. Reviewed bytes and lineage

- Reviewed commit: `dd9c9fc079470f03f247b71efb52d4de6b204e78` ("Add script-worker
  conformance vectors"), signed, author `Ivan Oparin <oparin@me.com>`.
- Remote branch `origin/spec/script-worker-v1-normative` head equals `dd9c9fc`
  (`git ls-remote` verified this cycle).
- Merge history present on the branch: `d3db977` (spec/sw-core-prose),
  `c3e80ea` (spec/sw-manager-security), `a690d63` (spec/sw-schema) — checklist
  item 2 satisfied.
- `git merge-base --is-ancestor dd9c9fc 6001dc3` → **yes**. The reviewed bytes
  are a true ancestor of the qualified rc.9 candidate
  `6001dc33281b94a4ec7442ab15278550dd0f51d9` on `candidate/schema-8-rc.9`.
- `git diff dd9c9fc 6001dc3 -- conformance/v1/vectors/script-host-execution-policy.json`
  is a single line: `protocol_version` `1.0.0-rc.8` → `1.0.0-rc.9`. That bump is
  owned by TASK-260822-c0rxj7's rc.9 migration; every behavioural byte of this
  task survived into the candidate verbatim.
- `release/1.0.0-rc.8.json` blob at `6001dc3` is `e05e4e92…`, identical to
  `origin/main`. rc.8 stayed frozen.

## 2. Content review against the normative prose

Every identifier the vector asserts traces to normative text on the same branch:

| Vector token | Prose authority |
| --- | --- |
| `script-worker-v1-native-control-inventory-v1` inventory table | `protocol/core.md` §4 inventory table, cell-for-cell match on all 8 controls × 3 platforms |
| `script_execution_control_unavailable` | core.md:562, profiles/manager.md:699, SECURITY.md:479 |
| `script_execution_capability_evidence_invalid` | core.md:510–515 closure table |
| `script_execution_hardened_claim_forbidden` | core.md:517–519, manager.md:844 |
| `script-command-declared-only` / `script-command-unfiltered-declared-network` | core.md:584–586, manager.md:955/959 |
| `script-capability-evidence-v1` `excluded_from` incl. `install-marker` | core.md:525-531 ("MUST NOT appear in a cache key, receipt input, marker record, or claim") |

**Checklist item 1 (the review-rework regression) is closed.** The Linux
`active-process-count-limit` cell is
`{availability: host-conditional, mechanism: delegated-cgroup-v2-pids.max,
unavailable_reason: null}` in `native_control_inventory`, matching core.md
78d544d and its `RLIMIT_NPROC` prohibition paragraph. Both required Linux lane
cases exist in `preflight_cases`:

- `linux-pids-max-probe-available-evidence-applied-invocation-succeeds`
- `linux-pids-max-probe-unavailable-evidence-unavailable-invocation-succeeds`
  (`probe_result: unavailable`, `evidence_status: unavailable`,
  `invocation_succeeds: true`, `expected_error: null`)

and the Linux evidence example reports the control `host-conditional` /
`unavailable`, i.e. honestly not applied.

**All five AC behaviours have positive and negative coverage:**

| Behaviour | Cases | Negatives |
| --- | ---: | --- |
| opt-in parsing | 6 | `interpreter-without-policy`, `policy-without-interpreter`, `unknown-policy` (script-worker-v2) |
| deny-by-default derivation | 4 | absent fields derive offline network / interpreter-only exec / empty secrets+env_read / four private roots; declared network is reporting-only with warning; declared exec is manager-resolved with `inherited_path: false`; declared secrets stay identifiers |
| mandatory-control preflight rejection | 5 | reject at `install` and at `invoke`, both `worker_started: false`; fixed-unavailable and host-conditional-absent controls explicitly do NOT reject |
| evidence-record closure | 14 | 10 × `..._evidence_invalid` (status contradictions, missing/duplicate/extra entry, unknown record version, cached probe, second record, foreign build record version) + 3 × `..._hardened_claim_forbidden` (foreign policy, deferred script guarantee, deferred build guarantee) |
| legacy declared-only labeling | 4 | schema 7 and schema-8-without-policy both label `script-command-declared-only`; enforced labels empty; enforced + host-glob network labels `script-command-unfiltered-declared-network` |

Both hooks named in the task notes are closed: `mixed_build_cases` gained
`schema8-script-worker` (manifest_schema 8 → marker_version 4, drivers and
`receipt_versions` ordered exactly like the existing `schema7-mixed` row), and
`conformance/v1/expected/install-marker-v4.json` now exists with its manifest
digest `sha256:905101ee…`. The marker-v4 *prose* binding lands in the candidate
at `e66cb72` (core.md §10 "managers supporting schema 8 MUST read marker schemas
1, 2, 3, and 4"), so the fixture is not orphaned in the landed lineage.

## 3. Independent gate runs (isolated worktrees, this reviewer)

Isolated `git worktree` at `6001dc3`, fresh venv with `jsonschema` 4.26.0,
go1.25.5 darwin/arm64:

| Gate | Exit | Result |
| --- | ---: | --- |
| `tools/validate.py` | 0 | validated 53 schemas and 691 vector files |
| `python -B -m unittest discover -s tools -p 'test_*.py'` | 0 | Ran 98 tests, OK |
| `go test ./tools/...` | 0 | ok generate-vectors 0.442s |
| `gofmt -l tools` | 0 | no output |

**Determinism.** Two consecutive `go run ./tools/generate-vectors -root .`
passes at `6001dc3`: SHA-256 inventories over all 671 generated JSON files are
byte-identical, and `git diff --exit-code -- conformance/v1 release/1.0.0-rc.5..rc.9.json`
exits 0 — regeneration reproduces the committed bytes exactly.

Repeated on the reviewed commit `dd9c9fc` itself: two passes byte-identical over
633 files; `conformance/v1` clean, and the only dirty path is
`release/1.0.0-rc.8.json`. That reproduces the developer's stop-the-line
mechanically.

**Teeth check.** Mutating the Linux cell in the committed vector to
`available / RLIMIT_NPROC` makes `validate.py` fail
(`vector digest mismatch for vectors/script-host-execution-policy.json`) and
`TestScriptWorkerConformanceContract` fail with the explicit RLIMIT_NPROC
assertion; regenerating restores the file byte-exactly. The validator, the Go
contract test, and the manifest digest are all live, not decorative. The Python
mutation suite covers five drift shapes in-memory, including the exact
RLIMIT_NPROC regression this task was reworked for.

## 4. Delivery gate — AC "spec CI green on the branch"

Literal reading: `spec/script-worker-v1-normative` @ `dd9c9fc` is **red**
(Specification CI 32632173590, all three OS lanes fail at `validate.py` with
`rc.8 downstream candidate pin does not match the suite manifest`).

That branch is structurally un-greenable standalone: the new vectors change
`conformance/v1/manifest.json`, and the rc.8 release metadata pins the live
manifest, so green would require rewriting immutable published rc.8. That is the
forced fit the developer refused and the board routed to TASK-260822-c0rxj7,
which is now `done`. I verified the mechanism myself (§3, dd9c9fc regeneration
dirties only rc.8). Demanding literal branch-green would require re-introducing
the rejected forced fit, so the AC is satisfied through the candidate:

- curator-spec **Specification CI 32659168954** on exactly
  `6001dc33281b94a4ec7442ab15278550dd0f51d9`: **success** — Formatting, Links,
  Release target provenance, Specification on ubuntu-latest / macos-latest /
  windows-latest.
- curator **CI 32659157687** (candidate-conformance lane): **success**, all 14
  jobs, Candidate suite green on ubuntu / macos / windows.

Both runs execute this task's validator, Go contract test, and regenerate-check
over the exact reviewed bytes, on three platforms — strictly broader coverage
than a standalone branch run would have given.

## 5. Architecture fit

The vectors follow the established suite shape: a single generator-owned vector
file per policy (mirroring `go-host-execution-policy.json`), a
`validate_*_execution_policy` function wired into `validate_vector_semantics`, a
Go contract test in `generate-vectors/main_test.go`, a Python mutation test in
`test_validate.py`, manifest digest entries, and a `conformance/README.md` entry.
Nothing hand-maintained, nothing bypassing the generator.

## 6. Non-blocking follow-ups (do not gate this task)

1. `script_execution_policy_unsupported` — the fourth policy diagnostic
   (core.md:565, manager.md:845) has zero conformance coverage anywhere in
   `conformance/` or `tools/`. It is outside this task's five named behaviours
   (it is a manager-capability case, not a vector-shape case), but it is the one
   normative diagnostic of `script-worker-v1` with no vector. Worth a follow-up
   in the manager-capability lane.
2. `validate_script_host_execution_policy` asserts exact set membership for
   `opt_in_cases`, `mandatory_controls`, and the native-control inventory, but
   only per-name lookups for `preflight_cases` and `capability_evidence_cases`.
   An extra unasserted case could be added to either without failing a gate.
   Cosmetic hardening, not a defect in the delivered bytes.

## 7. Handoff

Reviewer archetype — no `commit_ack` supplied. The reviewed scope is already
committed and pushed at `dd9c9fc` and carried in the qualified candidate
`6001dc3`; no further commit is owed for this task.
