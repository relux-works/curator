# Review verdict — TASK-260918-3moznc revision 1

Verdict: **changes_requested**, route **to-dev**. Reviewed the curator-spec story worktree at HEAD `5146c7b9ed4b0c07b840ab58f9908f197d667478`, without editing candidate code. This is ordinary specification rework, not a stop-the-line boundary.

## Findings requiring correction

1. **Unreadable lock recovery contradicts the new fail-closed contract.** `protocol/environments.md:742` explicitly adds a lock that “cannot be read or parsed” to entry-class failures; lines 744–747 then require that “a real operation rebuilds it from the revalidated snapshot”. The new general rule at lines 2003–2008 instead says the operation needing the unreadable file fails closed and forbids absence-shaped re-materialization. §10.1 at line 2650 inherits entry-class rebuild behavior; the lock row at 2017 supplies no exception. The producer evidence acknowledges lock-rebuild mechanics are vague, but this revision actively broadens the rebuild branch to read/parse failures. Specify an unambiguous read-failure disposition and precedence across §4, §8.4.1, §10.1 and manager mirrors; preserve the settled fail-closed decision. Add a pinned unreadable-lock repair/mutation case that refuses rebuilding from unreadable evidence. Current lock vectors assert no fragment but do not constrain rebuild/mutation.
2. **Required §9.7 diagnostic admission is missing.** `protocol/environments.md:2563` (§9.7) omits the newly introduced `environment_passthrough_unreadable`. It is correctly present at lines 1636 and 2905 and in manager mirrors, but the producer deliberately reinterpreted an explicit deliverable as applying only to inventory-class codes. The task's AC/DoD and round-specific instructions explicitly require new codes in §9.7. Add the required row (a clear cross-reference to the owning §7.7 row is sufficient to avoid duplicated policy). No new knob or lock key was requested; §12.1/§12.2 changes are therefore not needed.
3. **The closed table still has an unnamed unreadable outcome.** `protocol/environments.md:2010–2012` requires every read site to name its row's diagnostic, but the backup-record row at 2019 and owning paragraph at 1903–1912 contain only behavior (“the backup inventory is unknown”), with no diagnostic. The producer evidence explicitly confirms that no existing dedicated backup code was found. This is a useful gap discovery, not fulfillment of the settled per-class diagnostic requirement. Reconcile the backup row/read site with an applicable existing diagnostic, or explicitly surface the unresolved spec gap and a bounded proposed resolution; do not silently waive this acceptance criterion or introduce an unauthorized new code. Include the backup status/reporting mirror if its contract changes.

## Deliverable review

| Requirement | Evidence / quote | Verdict |
| --- | --- | --- |
| One general rule, all failure classes | environments:1992–2008: “MUST NOT be reported, persisted, or acted upon as absence”; includes permission, I/O, type, encoding, parse, schema, non-directory components and prohibited links | Rule text present; lock interaction requires correction (F1) |
| Closed class table with owners | environments:2014–2023 has all eight classes and owning sections | Partial: backup diagnostic missing (F3) |
| Required read sites reference rule | §1.3:210; §7.4:1396/1524; §8.2:1850; §8.3:1903; §9.5:2422; §9.6:2486/2515; §10.1:2693 | All seven required section groups cite §8.4.1; diagnostic completeness subject to F3 |
| Existing diagnostic reuse | lock uses existing `environment_store_untrusted`; seed uses existing `environment_seed_unreadable`; passthrough detached does not describe a failed lstat/readlink | Reuse rationale sound for identity, but lock recovery is not reconciled (F1) |
| New diagnostic everywhere required | §7.7 and §10.4 plus manager §12.2/§12.5 name `environment_passthrough_unreadable`; §9.7 omits it | Fail (F2) |
| Status posture | environments:1397–1401 and 1908–1909: non-current/currency unknown; mirrors manager:2547/2834 | Present for passthrough and backup behavior |
| §13 conformance surface | environments:3474 onward describes markers, seven lock failures, seeds and passthrough | Present |
| Vectors and rule-7 gate | new family, tools/validate.py:8741 and registration:10069; tests:4213 | Structural gate present; independent mutation results below |
| Existing vectors unchanged | independently compared all 36 pre-existing tracked vector files to HEAD | 36/36 byte-identical |
| Manifest and release pins | only new family entry and rc.9 manifest pins; generator rerun unchanged | Pass |
| CHANGELOG/direct rollout | CHANGELOG:10 onward names STORY-260916-1ll22r; “Rollout is direct (correctness of an existing MUST)” | Pass; no warn-first rollout required for this task |
| Scope/architecture | eight changed paths: protocol, manager mirror, CHANGELOG, new vector, manifest, rc.9, validator and tests | No manager/launcher implementation or frozen schema change |
| Patch identity | attached patch and `git diff HEAD` SHA256 both `17b4dbb56059590ae448f582e68ae5a05568037d99ad0ed5808ea0cc9ca065db` | Byte-identical, stronger than patch-id comparison; new file already intent-to-added |
| Empty curator delta | exact CR base-to-candidate diff has zero paths; curator worktree status empty | Expected: task changes curator-spec, carried by its attached patch. Empty curator delta is not itself a rejection |

## Additional read-site inventory review

Reviewed against the producer's 27-site inventory. Verdicts describe references/classification, not acceptance of the whole revision.

| Site | Verdict |
| --- | --- |
| §1.1 path operands | Already conformant; citation updated; lock diagnostic row added |
| §4 boundary | Already fail-closed for boundary; new lock-read extension introduces F1 |
| §6 overlay source | Existing §1.1 diagnostic reference; indirectly covered |
| §7.1 XDG reconciliation | Silent → now covered by seed row; keeps link on uncertain target read |
| §7.3 launcher path check | Existing fail-closed path check; citation updated |
| §7.5 shadowing probe | Silent → now covered; unverifiable presence reported |
| §7.6 target probe | Silent → now covered; unverifiable probe reported; consent remains |
| §7.9 release detection | Existing unknown-not-matching rule; informative scope |
| §9.4 sync | Silent → now covered: unreadable lock not omitted |
| §9.4 migration | Added fail-before-write behavior, but no explicit §8.4.1 citation or named diagnostic; producer classifies core install records outside table |
| §9.2 use/update/remove | Indirect §1.3 coverage only; no added per-operation reference |
| §10.4 | Added surface/lock/passthrough rows |
| §11 provider roots | Already distinct outcomes; citation updated |
| §12 profile list | Silent → now covered; unreadable profile remains listed |
| §12 currency | Existing reference updated |
| §12 GC | Reference added; unreadable locks named; uncertain references retained |
| §2/§3/§5 snapshot/store reads | Existing §4 validation dependency; no direct new rule reference |
| manager §12.2/§12.5/§12.7 | References and diagnostic mirrors added |
| manager §1 configuration | Existing fail-closed contract; producer explicitly excludes from closed table |

The claim “every read site references the rule” is broader than the direct references delivered: several sites above rely on transitive coverage or are declared out-of-table. Tighten that evidence claim, and add an explicit reference/diagnostic where a newly covered state read (notably migration) is intended to fall under this task's universal rule.

## Independent validation

Shell: zsh, `set -o pipefail`; PATH prepended with `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin`.

- `make validate`: transcript attached separately; final result appended below.
- Independently replayed same-class, name-preserving replacements of each unreadable positive with a self-consistent absent case: **21/21 refused** by `validate_environments_read_failure_vectors`. Coverage is structural scenario pinning, not real filesystem/manager behavior. The manager implementation is explicitly out of scope.
- Read the 31-name substitution-through-main test and all 14 new test methods. Note the producer says “20” unreadable positives but its breakdown and actual corpus total **21**, plus 5 absence positives, 4 negatives, and 1 present-directory detached case = 31.
- `make regenerate` on a disposable copy: exit **0**.
- Literal `make regenerate-check` on that copy: exit **2**, because its `git diff --exit-code` compares the intentional uncommitted manifest/new-vector/rc.9 changes against the original index. This is not generated drift. SHA256 inventories of all conformance/v1 and release files before and after both generator runs were identical (fixpoint proof). Candidate was not regenerated in place.
- No tests can resolve the contradictory normative recovery behavior or missing diagnostic table rows.

No logbook CLI/tool is available and campaign rules prohibit editing LOGBOOK.md. Findings are preserved in this task-scoped verdict and board notes instead; no prohibited logbook edit was made.

### Final validation results and bounds

`PATH="/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin:$PATH" make validate` began in the candidate tree. Its first gate passed:

```text
python3 tools/validate.py
validated 62 schemas and 1118 vector files
```

The full unittest discovery did not finish. I deliberately interrupted its Python process during fixture copying and switched to bounded checks; make exited **2** with KeyboardInterrupt. This is an incomplete reviewer rerun, not a discovered test failure. No claim of independently green 552-test coverage is made. The producer reports 552 tests passing; that remains producer evidence only.

Reviewer-rerun focused command loaded `ReadFailureVectorTests` with unittest and excluded only `test_substituted_scenario_rejected_through_main` (the 31-replay method). Exit **0**:

```text
.............
Ran 13 tests in 0.011s
OK
```

Additionally ran the actual `validate.main()` through the test harness on the disposable copy, with a same-class absent-lock body under the unreadable-lock name and recomputed manifest/rc.9 pins. Expected validator refusal, harness exit **0**:

```text
REAL MAIN same-class lock absence substitution: exit 1
validation failed: environments-read-failure case lock-read-io-error-untrusted inputs do not match its pinned scenario
```

This proves the gate is called from the real entry point and cannot be bypassed merely by repinning hashes. Independently tested 21/21 unreadable positives directly; entry-point substitution coverage independently rerun is **1/31** names, not 31/31.

`go test ./tools/...` exit **0**:

```text
ok github.com/relux-works/curator-spec/tools/generate-vectors 3.258s
```

Final candidate diff SHA256 remained `17b4dbb56059590ae448f582e68ae5a05568037d99ad0ed5808ea0cc9ca065db`; curator delta remains empty. All started commands completed or were explicitly interrupted before verdict recording. `task-board spawn goal "$TASK_BOARD_RUN_ID"` reports this run is not goal-bound.

Next producer: resolve F1–F3, tighten the universal-coverage evidence claim, update patch/evidence and run complete validation in bounded subsets where necessary, then hand off for another reviewer cycle.
