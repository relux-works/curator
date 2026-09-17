# TASK-260910-39fzpq — review verdict revision 2

Verdict: accepted. Accept CR revision 2 and route to integrating; this review does not mark the task done or claim it is merged.

## Candidate identity and empty curator delta

The curator CR delta from 0f0ae61766026cf2389ec4b91c09134cdb8aac62 to 3c30a7f627678f20a8c70ce0f776d7ef7ef7260f is empty, independently confirmed. That is the correct repository outcome for this spec-only leaf: its actual deliverable is the separate curator-spec Story candidate and attached spec patch, not a curator implementation change. Integration of that spec remains necessary for the acceptance criterion “Environments revision merged.”

Reviewed curator-spec worktree: /Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260910-148pj1/worktree, HEAD 684c9f1324d46b4938b2e5943f20c89e27971ec8. Attached TASK-260910-39fzpq_spec-patch_rev2.patch equals git diff HEAD byte-for-byte. SHA256 88840932af637e7901f9b235fec8f540244959ec397d4ec9eff2049c5eda6783; stable patch ID fc41a77779920f02cef4c8c834ab9389f1a6390b. Eight changed paths, 1574 insertions, 28 deletions. No candidate source or index was changed by this reviewer.

## Per-item review and prior findings

Paths and lines below refer to the curator-spec candidate. E = protocol/environments.md; M = profiles/manager.md.

| Requirement | Exact text / evidence and location | Result |
|---|---|---|
| §4 protected state and core §9.3 mirror | E:650 “manager-created, manager-protected, and resolved independently of package input”; E:662–676 “owned by the operator”, “private mutation permissions”, “resolves below”, “directory or a regular file”, “verified by `lstat`” | Pass; core:1632 unchanged |
| Verification scope and timing | E:653–660 “On every `env resolve`”, “again under the manager-home mutation lock”; roots, lock, home markers and every named entry explicitly included | Pass |
| Lock and marker file protection | E:201 “The lock file itself is verified”; E:1661 “The marker file itself is verified” | Pass |
| F1: independent pin baseline and home currency | E:677 “Store integrity is verified against the pin, not the marker”; E:2350 “Home currency is separate”; E:2355 “An intact updated store with an old marker is stale and repair succeeds” | Closed |
| F2: missing marker cannot bypass integrity | E:683 “This check needs no home marker”; E:685 “Missing hashes never count as passed”; E:2348 “a swapped entry is untrusted even with no marker” | Closed |
| Absent vs unreadable/malformed | E:1648 “An absent marker and an unreadable or malformed marker are distinct facts”; E:1719 `environment_marker_unreadable`, “never \"absent\"” | Pass; unsupported version remains `environment_marker_invalid` |
| Hash rule, cost and justification | E:2339–2365 “recomputes every named store entry's tree hash from its bytes”; git tree identity / path and local state_sha256; “O(store entry bytes named by the lock)” | Pass; full pin verification supersedes rev1 per-surface rule |
| F3: failure classes and ordering | E:687–705 “Two failure classes, in order”; enclosing failure “before its first write”, “nothing is rebuilt”; entry recovery “operation-private staging, atomic publication”; E:2335 “enclosing boundary → entries → pin hashes → home currency” | Closed |
| Fail-closed and repair | E:2346 “no fragment”; E:2308 “A store entry is re-applied only after that entry passed”; E:2319–2324 git reacquires pinned bytes; path/local fail repair without a second snapshot | Pass; no adoption of tampered bytes |
| Dry-run and GC split | E:2325–2328 `would-rebuild-untrusted-store` for entry failures; enclosing failure `environment_store_untrusted`, “no rebuild planned”; E:2785 “refuses collection before its first write”; E:2791 “collection never rebuilds an entry” | Pass |
| Drift and diagnostics | E:1722 “Store failure is not drift”; §8.5 E:1735–1737; §10.4 E:2542–2546 include exact diagnostic/outcome spellings | Pass |
| Posture | E:2729 “store-trust row per installed profile”; E:2731 “naming the failing check” and enclosing boundary; E:2743 store-untrusted is non-current | Pass |
| Rollout and configuration | E:708 “The contract is not configurable”; E:709 “Rollout is direct”; §12.1/§12.2 closed knob and lockable sets unchanged | Pass; no new knob/key/schema needed |
| §13 conformance | E:2969 names the new vector; E:2970–2985 lists intact, failures, roots, old-marker, unprovisioned, unreadable, recovery and negative cases | Pass |
| F4: scenario binding | tools/validate.py:5873–5940 pins object, failing check, home, surface and repair kind; :5987 checks those discriminators; :6953 registers gate in main; tools/test_validate.py:2144 five-to-one regression | Closed by independent main-entry probes below |
| Vector branches | conformance/v1/vectors/environments-store-boundary.json: 20 resolve, 4 dry-run, 10 repair, 4 status cases; all 38 opened and checked | Pass within model bound below |
| Manifest and pins | conformance/v1/manifest.json:4292 registers vector; release/1.0.0-rc.9.json:15,27 both pin manifest b42b5d7ba548f738e788aa76a89f2fcad6cf0386ae2a5ee10064e814f37b7a71 | Pass |
| Existing vectors byte identity | Every HEAD-tracked conformance file except v1 manifest compared to HEAD bytes: 1213/1213 identical, including historical conformance sets | Pass |
| CHANGELOG | CHANGELOG.md:10 “S5: protected-boundary contract”; details pin baseline, failure classes, direct rollout and diagnostics | Pass |
| Mirror and scope | M:2339–2341,2566–2580,2628–2630,2723–2730 consistent with E; only two spec docs, CHANGELOG, vector, manifest, rc.9, validator and its tests changed | Pass; no manager implementation, core, schemas, CLI, E6 admission, E7 launcher ownership or proposals 0014–0018 changes |
| Outcomes / AC | Producer evidence and revision-2 patch attached; reviewer verdict attached before acceptance | Review accepted for integration; merged AC remains producer/integration responsibility |

No blocking findings remain from F1–F4. No human-only decision is required.

## Independent adversarial probes

Called the real tools/validate.py main() in-process, intercepting only load_json for the S5 vector with a copied input. Every other gate remained enabled; candidate files and manifest were unchanged. This tests the semantic entry-point call, not merely the helper and not a manifest-digest rejection.

- Published corpus, including intact updated store plus old marker: main exit 0.
- Five boundary branches collapsed to ownership-only, adjusting failing_check consistently: main exit 1, “symlinked-entry-root-untrusted: this branch must fail link_safety”.
- Intact updated store plus old marker mislabeled untrusted: main exit 1, expected `environment_home_stale`.
- Swapped updated store plus old marker mislabeled stale: main exit 1, expected `environment_store_untrusted`.
- Swapped unprovisioned entry mislabeled stale: main exit 1, expected `environment_store_untrusted`.

Measured 5/5 expected main-entry outcomes (one positive, four refusals). Published cases cover 5/5 boundary check classes and 5/5 protected object classes, and pin-hash/home-currency branches. This is not a claim of all 25 object/check cross-products. The vectors model check results as booleans and home state as labels; they verify outcome consistency and branch identity, not actual OS ownership/DACL/lstat operations, tree hashing, atomic publication or manager runtime execution. Those implementation proofs belong to TASK-260910-32gki6. The full-tree normative rule includes a system-prompt swap without a marker; the no-marker vector models a root-context swap through the same pin-hash failure branch.

## Independent regeneration

Ran make regenerate-check on a disposable byte-copy at /tmp/TASK-260910-39fzpq-review2-regen, with its own temporary Git index staged to candidate bytes, no commit. This avoids modifying the read-only candidate and prevents the intended uncommitted spec delta from being mistaken for a regeneration difference.

```text
go run ./tools/generate-vectors -root .
git diff --exit-code -- conformance/v1 release/1.0.0-rc.5.json release/1.0.0-rc.6.json release/1.0.0-rc.7.json release/1.0.0-rc.8.json release/1.0.0-rc.9.json
exit 0
Candidate vs regenerated conformance/ and release/: 1220 files; 0 byte differences.
```

git diff HEAD --check: exit 0. No producer gate transcript was substituted for the independent checks. Run goal query: “Active Goal: none (run is not goal-bound)”. No directives recorded. Campaign rules prohibit LOGBOOK.md edits; closure is recorded in this task-scoped verdict and board notes instead.

## Independent make validate

Shell: zsh with `set -o pipefail`, workdir the candidate, PATH prepended with `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin`.

The requested `make validate` was started and independently ran the schema/vector gate successfully:

```text
python3 tools/validate.py
validated 62 schemas and 1094 vector files
```

That gate exited 0 (make advanced to unittest). The full unittest command was deliberately interrupted near the headless time limit; make consequently exited 2 with KeyboardInterrupt, not a test assertion failure. **No claim is made that the single make validate invocation exited 0.** This follows the assignment's higher-priority operational requirement to split long verification into bounded calls.

The captured unittest stream contains exactly 205 successful `.` records and no failure/error record before interruption in `RepositoryDescriptorIdentityTests.test_absence_guard_fires_when_the_retired_name_returns`. The actual unittest discovery inventory was independently loaded (378 tests), not estimated from source counts. Only the fully completed first 162 tests, through `EnvironmentVectorTests.test_written_surface_cannot_claim_absence`, are credited from that run. The overlapping and remaining 216 tests were rerun with verbose unittest output as the following bounded command:

```sh
PATH="/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin:$PATH" PYTHONPATH=tools python3 -B -m unittest -v \
  tools.test_validate.ManagerConfigVectorTests \
  tools.test_validate.ManagerLifecycleValidationTests \
  tools.test_validate.RegistryPageBoundaryVectorTests \
  tools.test_validate.RepositoryDescriptorIdentityTests \
  tools.test_validate.SharedFixtureMarkerTests \
  tools.test_validate.ShellHookTrustVectorTests \
  tools.test_validate.SnapshotAcquisitionVectorTests \
  tools.test_validate.SourceSignersVectorTests \
  tools.test_validate.StoreBoundaryVectorTests \
  tools.test_validate.SystemConfigV2SchemaTests \
  tools.test_validate.UmbrellaProviderVectorTests \
  tools.test_validate.WireSemanticValidationTests \
  tools.test_validate.WorkflowRegenerationScopeTests \
  tools.test_verify_release_commit tools.test_verify_release_merge_policy
```

The independent final Makefile gate was also run directly:

```text
$ go test ./tools/...
ok github.com/relux-works/curator-spec/tools/generate-vectors 6.055s
exit 0
```

Bounded unittest rerun result:

```text
Ran 216 tests in 459.354s
OK
exit 0
```

Total independently observed successful test coverage: 162 retained completed outcomes + 216 completed rerun outcomes = **378/378 discovered tests**. No producer result is needed to fill a coverage gap. The interrupted overall command is disclosed above; its incomplete tail is never credited. This bounded completion satisfies the headless-run instruction without misreporting a full-command exit code. No test process remains running.

The attached TASK-260910-39fzpq_review-validation-rev2.log preserves the original interruption, complete verbose bounded rerun, regeneration output and exact discovery inventory. Final candidate patch SHA256 was rechecked unchanged before attaching this verdict.
