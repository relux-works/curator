# TASK-260906-1xbrz6 review verdict — revision 1

Verdict: CHANGES REQUESTED. Route to to-dev; no acceptance or commit acknowledgement.

Reviewed candidate c4e98ae2d40888c304786d7a7ae5d8d83870dcbe against base 05053cd70bb64d68b3aa12e1af286203575d974c. git diff of the working tree against the candidate exited 0. No source files modified. Run goal queried: not goal-bound.

## Required rework

F1 (major): tools/validate.py:4209 and :4284 only require token presence, omitting the operative exclusion and fail-closed predicates. Replacing “they fail closed” with “they fail open”, or “operations are outside this closed set” with “operations are inside this closed set”, both PASS validate_takeover_closed_set_text. Thus contradictory normative text passes the new pin. Pin the actual exclusion/fail-closed sentence, including its subjects (activation/global operations), in all four locations; whitespace-normalized exact sentence assertions are sufficient and simpler than attempting semantic parsing. Add negative tests for these predicate changes, not merely deletion of reference tokens.

F2 (moderate): tools/validate.py:4245 and :4253 do not enforce the claimed sentence/section boundaries. Renaming heading “### 9.5 Onboarding” to “### 9.5 Onboarding-renamed” still passes because matching is a prefix split. Splitting the exclusion at “closed set. on” also passes because sentence splitting recognizes only uppercase/backtick successors. Anchor complete heading lines, delimit sections at appropriate same-or-higher heading levels, and test sentence movement and splitting independently. An exact normalized sentence pin would also remove the sentence-splitting ambiguity.

## Executable reproduction (read-only; observed exit 0)

Run from this Story worktree:
```sh
PYTHONDONTWRITEBYTECODE=1 PYTHONPATH=tools /Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin/python - <<'PY'
import validate
e = (validate.ROOT / "protocol/environments.md").read_text()
cases = {
    "fail-open": ("they fail closed exactly as\nsection 8.3 states",
                  "they fail open exactly as\nsection 8.3 states"),
    "not-outside": ("operations are outside this closed set: on",
                    "operations are inside this closed set: on"),
    "rename-heading": ("### 9.5 Onboarding\n",
                       "### 9.5 Onboarding-renamed\n"),
    "scatter-lowercase": ("operations are outside this closed set: on\n",
                         "operations are outside this closed set. on\n"),
}
for name, (old, new) in cases.items():
    assert e.count(old) == 1
    try:
        validate.validate_takeover_closed_set_text(
            environments_text=e.replace(old, new, 1))
    except validate.ValidationFailure as exc:
        print(name, "REJECTED", exc)
    else:
        print(name, "SURVIVED")
PY
```
Observed: all four SURVIVED (0/4 rejected). This drives the production text-check function directly, not the full gate; main() registration at tools/validate.py:10815 was inspected. No claim that the whole gate was rerun on these mutants.

## Independent checks and reused evidence

- zsh command: PYTHONDONTWRITEBYTECODE=1 PYTHONPATH=tools /Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin/python -m unittest tools.test_validate.TakeoverClosedSetTextTests -v
- Exit 0, 19/19 tests pass. Independently reran the supplied widening, narrowing, reordering, deletion, scattering, trigger drift and CLI-flag mutants as part of this class. Their chosen mutations are caught; the four additional mutations above are not.
- Initial system-python attempts failed at imports (assurance, then jsonschema); those are environment failures, not product test failures. Corrected with PYTHONPATH and the existing project venv; no installation performed.
- Reused TASK-260906-1xbrz6_change-request_rev1-validation.log: configured spec-gate.sh exited 0, 598 Python tests OK, 62 schemas / 1124 vector files validated, Go generator tests OK. Did not rerun the landing suite. The attached log does not separately identify regenerate-check, so no independent regenerate-check pass is claimed.
- Read producer results and source review-cycle-2 findings. CHANGELOG entry is under Unreleased. Exactly five changed paths; cli/curator.md remains unchanged, so the separate editorial leaf was not pre-empted.
- Search of conformance JSON for takeover/profile sync found only environments-write-nofollow.json, whose operation list is write classes, not carrier commands. No evidence of a carrier-enumerating family requiring a new vector. No schema/vector changed.
- Carrier enumeration and onboarding list equality are asserted; CLI import/global flag prohibition is executed and its two negative tests pass.

## Normative recovery assessment and limits

The four added sentences agree on diagnostic, exclusion and recovery. Manager 12.5 retains read-only resolve and explicit repair, without importing a new takeover carrier. First materialization alone is not a counterexample: 9.1 installs lock/store before activation, and 9.2 use materializes from the selected lock even when current was not changed. A target foreign symlink has the backup-and-entry-replacement path in 8.3.1/9.5; an ancestor symlink is a different diagnostic (environment_write_would_follow_link), which takeover intentionally cannot cure.

Two bounds need care in the producer gap analysis: sync/use only write paths represented by an installed lock, and import retry with an already installed name meets profile_import_name_taken (9.7). The recovery can complete activation, but a literal rerun of profile import is not established as successful by that reasoning. Likewise the universal global-add recovery claim depends on publication of the newly extended lock before a surface conflict; 9.4 alone does not spell out that failure ordering. These are not verified runtime failures. Do not widen the carrier set. In rework, substantiate these sequences from the transaction contract or record the unresolved retry/publication question as the brief-required decision packet, distinguishing activation retry from repeating import. This review does not claim complete operator-state coverage.

No control-root LOGBOOK edit: the campaign forbids it. Findings are persisted here and in task notes. Rework is recoverable and does not require blocked status.
