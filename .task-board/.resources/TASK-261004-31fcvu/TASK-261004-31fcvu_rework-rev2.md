# TASK-261004-31fcvu — rc3-release-notes: developer rework, revision 2

Ready for review. Only CHANGELOG.md differs from the recorded base; all work remains uncommitted. LOGBOOK.md is untouched. This outcome supersedes revision 1's duplicate/disposition claims.

## Review findings addressed

- P1: all 27 previously deleted entries are restored byte-for-byte in original order under `### Shipped in 0.15.0-rc.2 but not recorded in its notes` at the end of rc.3. The required explanatory sentence attributes these changes to rc.2. Independent section-scoped lookup finds **0/27 released twins**, while **27/27** occur in the tag's Unreleased section. Presence somewhere in a tag is not evidence of a released-note duplicate. No previous entry is now classified as dropped.
- P2: the concrete internal runner identifier is replaced with “select the explicit self-hosted runner label.” Scanned the entire rc.3 section, including historical entries, for the known identifier, machine-name patterns and personal user paths; no match. Earlier release bytes remain unchanged as required.

The fresh empty Unreleased and normalized rc.3 heading/date remain. All bytes from the rc.2 heading through EOF match the base. The rc.14 pin is `43bf0a2506d5c354a73bbc3ea4623d4653db10c7`, manifest `6f832d813efc768ea154a7d5076b512ab4be6aa9409d92e11469d21ea9bc69f5`; production writes remain `curator-content-v1`. Known issues explicitly retain fp8vx7 as risk, rc.4 v2 writer deferral, N1–N4 unfixed (#106), and B3 exclusion.

## Direct validation and bounds

Commands were standalone processes; no tee or pipeline masks a gate result.

| Command | Actual exit | Result |
| --- | --- | --- |
| `python3 /tmp/TASK-261004-31fcvu_validate.py` | 0 | 54/54 mapped non-administrative commits; 258/258 board-only commits; 32/32 entry dispositions; 27/27 exact historical texts/order; prior release bytes; pin/write policy; known issues; public wording; CHANGELOG-only scope. |
| `python3 /tmp/TASK-261004-31fcvu_validate.py --regression` | 0 | Three named tests; 27/27 single-entry omissions rejected; Unreleased cannot serve as released evidence; original and historical-subsection runner injections rejected. |
| `python3 /tmp/TASK-261004-31fcvu_validate.py --regression --narrow-release-search` | 1 | Expected-red narrowing mutant: the gate refuses only entries absent from the whole rc.2 tag, allowing tag Unreleased to masquerade as released evidence. Both preservation regressions fail (all 27 omissions admitted). |
| `python3 /tmp/TASK-261004-31fcvu_validate.py --regression --narrow-public-scan` | 1 | Expected-red narrowing mutant: only the fresh rc.3 prose is scanned, excluding the historical subsection. `test_internal_runner_identifier_rejected_in_rc3` fails on the omitted surface. |
| `python3 /tmp/TASK-261004-31fcvu_validate.py --candidate /tmp/TASK-261004-31fcvu_rev1.md` | 1 | Expected-red previous candidate: rejects exactly entries 5–20 and 22–32, all 27 review omissions. |
| `git diff --check` | 0 | Tracked diff whitespace validation. |
| `env -u GOROOT GOMAXPROCS=2 go build -p 2 -o /tmp/TASK-261004-31fcvu_curator ./cmd/curator` | 0 | Build rerun in this rework; output outside repository. |

The first document check exited 1 on entry 3 because its consolidation-token comparison did not normalize wrapped prose. Corrected that comparison only; historical preservation remains byte-exact. The corrected validator exited 0 twice. An initial run against the previous candidate exited 1 at subsection structure; after moving the preservation check ahead of structure, the recorded replay exits 1 specifically for all 27 omissions. The P1 narrowing mutant also exited 1 before its reporting was condensed; the attached log records the current concise run. Normal regression runs exited 0 before and after these validator fixes. Diff checks exited 0 on both executions. No failed gate is reported as green.

No product-code behavior changed. Tests and validator are board outcome artifacts, preserving the task's CHANGELOG-only repository scope. Go unit tests, full conformance/platform/race suites and release qualification were **not run** because this rework changes documentation only. The roughly 21,000 hosted passes are historical incident evidence from the readiness report, not a new locally executed test. Revision 1 reviewer content spot checks (11 commits) are accepted prior evidence, not rerun here; the entire current subject map and 258 administrative path boundaries were rerun independently. Source base and history range are unchanged.

## Regression contract

`ReleaseNotesRegression.test_prior_unreleased_requires_released_twin_or_verbatim_rc3` derives the 27 historical texts from base Unreleased, removes each individually from the actual candidate file, and requires the same preservation gate used by `validate` to reject it. `test_prior_unreleased_in_ambient_unreleased_is_not_released` proves neither tag nor current Unreleased can provide released-section evidence. `test_internal_runner_identifier_rejected_in_rc3` uses the label derived from unchanged base history, preventing disclosure in the test artifact.

Five entries (1–4 and 21) retain the original brief's explicit rc.3 consolidations. Their exact source texts are in the ledger; validation checks the designated group, a unique bullet and required content for each. Entry 2 deliberately supersedes stale “keep rc.13 pinned” wording with the released pin and the owned v2-write gap. This is an explicit transformation, not a duplicate/deletion exception. Every other entry must be verbatim in rc.3 or have a section-scoped released twin; the restored block also has a strict byte/order assertion. Coverage is 27/27 exact historical preservations + 5/5 explicit consolidations = 32/32, with no silently omitted entries.

The validator proves document preservation, structure and subject-map identity; semantic completeness of every commit is bounded by the reviewed subject-to-entry map and prior content inspection, not automatically proved by substring checks. Runner scanning detects the base's known label and declared private-path/machine-name patterns; the whole rc.3 section was also read. The public subject list neutralizes the internal identifier; per-subject SHA-256 values preserve exact original subject identity in the JSON without copying the identifier.


See updated results and map resources for all 32 dispositions and 312 commit subjects.
