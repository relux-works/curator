# TASK-260916-1qfpu4 — revision 2 independent review

Candidate: curator-spec Story worktree, HEAD `684c9f1324d46b4938b2e5943f20c89e27971ec8`.
Reviewed resources: campaign producer rules, producer brief, evidence including Revision 2, rev1/rev2 spec patches, rework/review briefs; audit E5 at docs/security-audit-2026-09.md:299 and Appendix B:427.

Verdict: **accepted**. Revision 2 resolves F1, F2, and F3. No outstanding rework findings. Acceptance routes to integrating; it does not assert merged/done.

## Item-by-item review

| Item | Evidence and exact quotation | Result |
|---|---|---|
| Single normative write discipline | protocol/environments.md:1626: “Every materialization, takeover, repair, or backup write”; :1632: “operation-private name in the same directory, then renames over the target (atomic replace)”; :1633: “MUST NOT follow a symbolic link”; :1635: “manager itself did not create in this operation”; :1636: “semantics on open, `lstat`-class semantics on inspection.” | Pass; applies to all three modes, backup, marker, adapter ledger. |
| Target-link branches and takeover | environments.md:1644 “a manager-owned link”; :1649 “the backup preserves the symlink with the same link text, never dereferenced”; :1652 `environment_surface_unmanaged_conflict`; :2099 “Every takeover write is a section 8.3.1 write”. | Pass; authorization replaces the entry after backup, never writes through it. |
| Private destination refusal | environments.md:1661 “same refusal fires when a backup, marker, or ledger destination path”; :1663 “traverses or names a link the manager did not create in this operation.” | Pass; F2 modeled distinctly at tools/validate.py:6383. |
| Materialization and provisioning pointers | environments.md:657 “Every write below follows the section 8.3.1 write discipline”; :935 “Its write is a section 8.3.1 write”; :1517 “Every write in that transaction is a section 8.3.1 write.” | Pass; §5, §5.8, §8.1. |
| Shadowing, drift, repair | environments.md:1328 “The existence check uses `lstat`-class semantics”; :1681 “inspection uses `lstat`-class semantics”; :2273 “Repair writes are section 8.3.1 writes”. | Pass; §7.5, §8.4, §10.1. |
| One justified diagnostic; closed sets | environments.md:1658 and diagnostics table :1701; profiles/manager.md:2348; vector diagnostics array; tools/validate.py:6187. Exact new spelling `environment_write_would_follow_link`; existing foreign-manager/unmanaged-conflict branches retained. | Pass; uncovered parent traversal and private destination cases justify the code. No new knobs, marker fields, schemas, or lock keys; §12.1/§12.2 knob/lock additions are inapplicable. |
| Direct rollout and posture | CHANGELOG.md:10 “E5: nofollow write discipline”; :29 “Direct rollout, under the hood”; environments.md:1669 “the row is non-current”; §12 :2619/:2650 includes link-blocked paths/state. | Pass; no warn-first revisions required for E5. |
| Conformance inventory and untouched targets | environments.md:2866 “write-discipline vectors”; :2873 names both backup positions; :2876 “former target is byte-identical afterwards”. New vector has 11 cases and fixture SHA-256 expectations. | Pass; authorized/unapproved takeover, parent refusal both authorization states, planted-link and manager-owned repair, both backup positions, inside-link ledger refusal, clean/recorded positives. |
| F1 scenario pinning | tools/validate.py:6199 `WRITE_NOFOLLOW_SCENARIOS`; :6374 “inputs do not match its pinned scenario”; :6862 registered in main. tools/test_validate.py:2883 `test_substituted_scenario_rejected_through_main`. | Runtime results below. |
| F2 negative evidence | vector:177 `backup-symlinked-target-refused`, target symlink outside, parent null, nofollow diagnostic, unchanged target digest; test_validate.py:2963 rejects foreign-manager diagnostic; :2970 rejects replacement. Parent refusal :153 retained. | Pass; marker/ledger are explicitly text-only in producer evidence, backups vectorize both link positions. |
| F3 / empty curator CR | Exact base-to-candidate diff (`0f0ae61766026cf2389ec4b91c09134cdb8aac62` → `3c30a7f627678f20a8c70ce0f776d7ef7ef7260f`) and curator working status empty. | Pass. An empty curator delta is correct: this leaf delivers curator-spec changes via the separate spec worktree and attached patch. No curator implementation or LOGBOOK edit belongs in it. This does not claim the spec is merged. |
| Patch identity and scope | Attached rev2 patch and `git diff HEAD` byte-identical; stable patch ID `f9fad8dc3801642ae3df91ea135677d32da5e524`. Exactly 8 paths: CHANGELOG, protocol, manager profile, new vector, manifest, rc.9, validator and validator tests. | Pass; validator changes are spec conformance harness, not manager implementation. No S5/E7/proposals 0014–0018 content. |
| Rev1 preservation and pins | Rev1/rev2 patch comparison: CHANGELOG and manager profile unchanged; protocol only §13 backup-case inventory updated; all 10 original cases and remaining vector bytes identical after removing the new case block. All 31 pre-existing vector files match HEAD byte-for-byte. Manifest adds exactly one vector; rc.9 updates two manifest pins. | Pass. |

## Validation method and bounds

Reviewer independently executed all three Makefile validate gates. Per the headless bounded-call requirement, the unittest gate was partitioned into four contiguous discovery-order slices (100/100/100/65), covering all 365 tests exactly once. No producer test result substitutes for reviewer execution. The monolithic `make validate` invocation was not used because its unittest gate exceeds the per-call time bound. Python PATH: `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin:$PATH`; shell zsh with `set -o pipefail`; tests use `python3 -B`.

The semantic validator ran against the actual read-only candidate. Unit tests ran against four independent byte copies because the new entry-point regression test temporarily rewrites vector/manifest/release data. Candidate files and index were not edited. Scratch paths under `/tmp/e5-review.ujCUGY/`.

Independent replacement probe: enumerated the 11 published vector names (not the validator pin table); replaced each body with the next published, internally consistent case while retaining its original name, and called literal `validate.main()`. Only the vector load was substituted in memory, leaving candidate bytes untouched and all main checks active. The shipped regression additionally exercises on-disk substituted corpora with recomputed manifest/release pins.

This is structural spec conformance evidence, not proof of manager filesystem behavior; that remains TASK-260916-19shmj. Marker/ledger direct-link cases remain text-only as explicitly disclosed in revision 2 evidence. No new human decision or external blocker identified.

## Regeneration and identity evidence

`make regenerate-check` ran in `/tmp/e5-review.ujCUGY/regen` after `git init -q` and staging the complete candidate as the scratch index (no commits, no candidate index changes). It exited 0:

```text
go run ./tools/generate-vectors -root .
git diff --exit-code -- conformance/v1 release/1.0.0-rc.5.json release/1.0.0-rc.6.json release/1.0.0-rc.7.json release/1.0.0-rc.8.json release/1.0.0-rc.9.json
```

The check compares generated content against the candidate, avoiding the known false red against the uncommitted Story base. An additional byte comparison found **1381/1381 tracked candidate files identical** to the regenerated scratch tree, differences `[]` (exit 0). Git emitted CRLF staging warnings for two existing byte fixtures; the subsequent raw-byte comparison verifies no normalization drift.

Candidate/attached patch SHA-256 before and after validation: `696e2147c13147c3892b10ec1371faea9629a469d3804936f00715440ee941a1`. `git diff --check` exit 0.

## Independent validation transcript

All three `make validate` recipe gates passed (decomposed, not a monolithic make invocation):

```text
$ python3 -B tools/validate.py
validated 62 schemas and 1094 vector files
exit 0

$ python3 -B /tmp/e5-review.ujCUGY/slice.py 0
Ran 100 tests in 314.268s
OK
exit 0
$ python3 -B /tmp/e5-review.ujCUGY/slice.py 1
Ran 100 tests in 234.079s
OK
exit 0
$ python3 -B /tmp/e5-review.ujCUGY/slice.py 2
Ran 100 tests in 320.562s
OK
exit 0
$ python3 -B /tmp/e5-review.ujCUGY/slice.py 3
Ran 65 tests in 333.109s
OK
exit 0

$ go test ./tools/...
ok github.com/relux-works/curator-spec/tools/generate-vectors 1.512s
exit 0
```

The exact slicing runner (each slice ran in a separate complete candidate copy):

```python
import unittest,sys
suite=unittest.defaultTestLoader.discover('tools',pattern='test_*.py')
def flatten(s):
 for t in s:
  if isinstance(t,unittest.TestSuite): yield from flatten(t)
  else: yield t
alltests=list(flatten(suite)); n=int(sys.argv[1]); tests=alltests[n*100:(n+1)*100]
print(f'Discovered {len(alltests)} tests; slice {n}: {len(tests)}',flush=True)
r=unittest.TextTestRunner(verbosity=1).run(unittest.TestSuite(tests));sys.exit(not r.wasSuccessful())
```

Total: **365/365 tests passed**, including all 12 `WriteNofollowVectorTests`, its 11 on-disk replacement subtests through `validate.main()`, and the wrong foreign-manager diagnostic negative. No tests were omitted or accepted solely from producer evidence.

## Independent substitution probe

```text
takeover-symlinked-target-authorized-replaced REJECTED 1
takeover-symlinked-target-unauthorized-stopped REJECTED 1
materialize-symlinked-parent-refused REJECTED 1
takeover-symlinked-parent-authorized-still-refused REJECTED 1
repair-planted-link-replaced REJECTED 1
repair-manager-owned-link-replaced REJECTED 1
backup-symlinked-destination-refused REJECTED 1
backup-symlinked-target-refused REJECTED 1
materialize-clean-path-written REJECTED 1
takeover-inside-link-unauthorized-unmanaged-conflict REJECTED 1
materialize-recorded-file-replaced REJECTED 1
Replacements rejected 11/11
exit 0
```

Reviewer probe source (no candidate writes):

```python
import sys,pathlib,copy,contextlib,io,json
from unittest.mock import patch
sys.path.insert(0,str(pathlib.Path.cwd()/'tools'))
import validate
p=validate.SUITE/'vectors/environments-write-nofollow.json'
v=validate.load_json(p); original=validate.load_json
names=[c['name'] for c in v['cases']]
passed=0
for i,name in enumerate(names):
 changed=copy.deepcopy(v)
 donor=copy.deepcopy(v['cases'][(i+1)%len(names)])
 donor['name']=name;changed['cases'][i]=donor
 def load(path):
  return changed if path==p else original(path)
 err=io.StringIO()
 with patch.object(validate,'load_json',side_effect=load),contextlib.redirect_stderr(err),contextlib.redirect_stdout(io.StringIO()):
  status=validate.main()
 ok=status==1 and f'case {name} inputs do not match its pinned scenario' in err.getvalue()
 print(name, 'REJECTED' if ok else 'FAILED',status,flush=True)
 passed+=ok
print(f'Replacements rejected {passed}/{len(names)}',flush=True)
sys.exit(passed!=len(names))
```

## Final disposition

**Accepted**, with 11/11 independent replacement rejections, 11/11 shipped on-disk replacement subtests, all 365 tests green, validator and Go gate green, and regeneration exit 0. F1 scenario pinning, F2 private-destination routing, F3 empty curator delta are resolved. Final patch bytes still match the attached revision 2 resource.

Run goal queried before verdict: `Active Goal: none (run is not goal-bound)`. No directives recorded. No LOGBOOK was edited: campaign rules and F3 explicitly forbid it; this outcome records the review evidence. Producer-side integration remains outstanding.
