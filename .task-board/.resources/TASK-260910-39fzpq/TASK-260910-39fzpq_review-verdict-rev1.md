# TASK-260910-39fzpq — review verdict revision 1

Verdict: changes_requested. Route to to-dev; do not accept CR revision 1.

## Identity and empty curator delta

The empty curator repository delta is expected: this task changes the separate curator-spec Story worktree. The reviewed spec candidate at base 684c9f1324d46b4938b2e5943f20c89e27971ec8 has 8 changed files, 959 insertions and 18 deletions. The attached spec patch equals git diff HEAD byte-for-byte. Both stable patch IDs: 985b8821bfc7e9c85ee22e6b2b9144269507708c. SHA256: fb1f22f72e571a054d0bf4b733e05e12a62a1d18d6cf27f9285079c35b654b18.
No candidate source or index was edited by the reviewer.

## Corrections required

### F1 — high: stale home marker becomes integrity baseline for a different lock

protocol/environments.md:2294–2302 requires hashing entries named by the current lock, re-derived through section 5, and comparing with the home's recorded hashes; any mismatch is environment_store_untrusted. But lines 2254–2267 support a home “stale after profile update” and repair from the new lock. Section 5.1:789 embeds the lock hash in the generated surface; section 5.6:953–961 hashes the materialized file set. Section 9.2:1862–1869 explicitly marks managed homes stale after publishing the new lock and retains the old lock to identify them. Thus an intact new lock/store and old home can mismatch without corruption. Re-fetching the same valid new Git snapshot cannot make its generated surface match the old marker.

Separate home currency from store integrity. Specify an applicable trusted integrity baseline for the new entry when the marker belongs to another lock/configuration, then permit ordinary stale-home repair. Add cases: intact updated store + old marker is stale and repair succeeds; swapped updated store + old marker is untrusted and never adopted.

### F2 — high: absent marker bypasses byte verification before provisioning

protocol/environments.md:2304–2308 makes an unprovisioned home run boundary checks only; otherwise it is stale and repair re-verifies the contract before writing. A same-user byte swap preserves ownership, permissions, containment, type and link safety. With no marker, no hash baseline is specified, so provisioning can copy the swapped prompt and mint a marker for it. This contradicts lines 2273–2275 (“only after ... its surface hashes”; “a non-trusted entry is never re-applied”) and the brief's repair-not-persistence requirement.

Define independent integrity verification when no applicable marker exists, such as pinned-snapshot verification before provisioning. Missing hashes must not count as passed. Add intact/swapped unprovisioned-home repair cases and distinguish absent from unreadable/malformed markers. Keep the correction within S5.

### F3 — medium: fail-before-write and rebuild lack a defined separation

protocol/environments.md:677–683 says an unprovable boundary causes “every mutating profile operation” to fail “before its first write”, immediately followed by “A real operation rebuilds an untrusted entry ... into newly established protected state”. It does not distinguish an unsafe enclosing boundary from an individual entry that can safely be rebuilt. The successful repair vector models only a hash mismatch with an intact boundary.

Specify refusal versus safe recovery conditions, ordering and the verified publication boundary. Add a rebuildable entry-boundary failure and an enclosing-boundary failure that cannot safely be re-established. Reconcile dry-run and GC wording with this distinction. This is ordinary spec rework, not a human-only blocker.

### F4 — medium: named cases do not enforce branch coverage

tools/validate.py:5921–5952 checks exact names, one failed boolean and internally consistent outputs but does not bind named scenarios to their intended failure. An independent in-memory narrowing changed five cases (ownership, permissions, containment, regular types, link safety) to ownership-only, updating failing_check to ownership. The actual vector validation function accepted the narrowed corpus, exit 0:

    SURVIVED: five named negative boundary branches collapsed to ownership-only; gate accepts

Those five required boundary classes fell from 5/5 to 1/5 coverage with a green gate. Bind cases to required inputs/objects/surfaces or enforce an independent coverage matrix; add this narrowing regression. Current resolve vectors cover 3/5 protected object classes: entry, lock, marker; roots have no cases. Add environments-root/store-root cases. Boolean scenarios prove outcome consistency, not filesystem verification or manager runtime enforcement.

## Per-deliverable review

Locations are in the candidate curator-spec worktree.

| Requirement | Exact text / evidence | Result |
|---|---|---|
| Section 4 / core 9.3 mirror | environments:650–675 “manager-created, manager-protected”; “owned by the operator”; “private mutation permissions”; “resolves below”; “directory or a regular file”; “verified by lstat” | Present; F3 |
| Timing and scope | environments:653–660 “On every env resolve”, “again under the manager-home mutation lock”; roots, lock, marker, entries named | Present |
| Lock and marker themselves | environments:201–207 “lock file itself is verified”; 1635–1642 “marker file itself is verified” | Present |
| Hash rule and cost | environments:2294–2313 “recompute ... surface hashes”; “O(applied system-prompt and root-context module bytes)” | Per-surface choice justified; F1/F2 |
| Drift and diagnostic | environments:1690–1705 “Store failure is not drift”; environment_store_untrusted table row | Present |
| Failure / repair / dry-run | environments:677–685, 2273–2308, 2495–2499; “no fragment; non-current”; would-rebuild-untrusted-store | Present; F1–F3 |
| Posture / GC | environments:2680–2686 “store-trust row per installed profile”, “naming the failing check”; 2693 non-current; 2734 retains untrusted entries | Present |
| Direct rollout / no knob | environments:687–688 “not configurable”, “Rollout is direct”; 12.1/12.2 unchanged | Correct; no new knob or lock key required |
| Section 13 and vectors | environments:2914–2923 names vector and branches; 13 resolve, 3 dry-run, 4 repair, 3 status cases | Basic brief cases present; lifecycle and coverage gaps above |
| Manifest / rc.9 pins / old vectors | manifest:4292 registration; both rc.9 manifest pins updated; all 1093 pre-existing non-manifest conformance files byte-identical to HEAD | Pass |
| CHANGELOG | CHANGELOG:10 “S5: protected-boundary contract”; direct rollout and exact spellings | Pass |
| Closed names / schema | environment_store_untrusted and would-rebuild-untrusted-store consistent in text/tables/vector/validator/changelog; no schema, marker-field or knob change | Pass |
| Scope and architecture | environments, manager consistency mirror, CHANGELOG, vector, manifest, rc.9, validator, validator tests only | Conformance tooling appropriate; no manager implementation/core/E6/E7/proposals edits |
| Outcome / AC | Producer patch and evidence attached; spec not merged | Rework required; AC not complete |

## Independent validation

Shell zsh with set -o pipefail, from the curator-spec Story worktree:

    PATH="/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin:$PATH" make validate

Independent result, exit 0:

    python3 tools/validate.py
    validated 62 schemas and 1094 vector files
    python3 -B -m unittest discover -s tools -p 'test_*.py'
    Ran 370 tests in 383.479s
    OK
    go test ./tools/...
    ok github.com/relux-works/curator-spec/tools/generate-vectors 2.181s

All three Makefile gates completed. No test command was left running. git diff HEAD --check also exited 0. Green validation does not resolve the normative conflicts or the surviving narrowing described above.

Regeneration was independently rerun on a disposable byte-copy at /tmp/TASK-260910-39fzpq-review-regen, preserving the read-only candidate. The temporary Git index held candidate bytes, with no commit. Thus regenerate-check compares generated bytes to the candidate rather than its intended uncommitted changes against HEAD.

    $ set -o pipefail; make regenerate-check
    go run ./tools/generate-vectors -root .
    git diff --exit-code -- conformance/v1 release/1.0.0-rc.5.json release/1.0.0-rc.6.json release/1.0.0-rc.7.json release/1.0.0-rc.8.json release/1.0.0-rc.9.json
    exit 0
    Regeneration candidate-byte comparison: 1220 files; 0 differences []

Producer results were read but not substituted for independent checks. These checks do not establish manager runtime behavior; that implementation is out of scope. Findings are persisted in this task outcome and board notes; campaign rules explicitly prohibit LOGBOOK.md edits. Run goal query: “Active Goal: none (run is not goal-bound)”.
