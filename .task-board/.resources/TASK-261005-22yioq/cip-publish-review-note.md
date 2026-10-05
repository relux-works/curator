# Reviewer instruction — TASK-261005-22yioq: publish CIP drafts 0002–0006 into curator-spec cips/
Cross-provider review: the producer was muse max; the reviewer is codex sol high (sol medium is not admitted on the spec board).
Check:
1. Five files `cips/CIP-000N-<slug>.md` (0002–0006) follow `cips/TEMPLATE.md` and the rules in `cips/README.md` / CIP-0001. Status is **Draft**: the operator has made no acceptance decision. All required sections are present.
2. Each CIP faithfully condenses its source on curator main (`git -C /Users/administrator/Developer/ReluxWorks/curator/curator show origin/main:.research/261004_CIP-000N-*.md`): no decision was invented, no "decisions needed" item was dropped, and open questions and alternatives are kept. Evidence is linked to the curator `.research/` files by repo-relative path and commit, not copied wholesale.
3. The README index lists 0002–0006 if the README keeps an index.
4. No normative protocol or schema edits; no CHANGELOG or LOGBOOK edits; frozen and released files untouched.
5. No personal paths, host names or employer names (the repo is public). Count them with grep and report the count.
6. Links resolve. Applicable docs checks were run and their exit codes recorded.
Verdict via the board: `accept_cr` with the merged checklist, or changes requested with numbered findings. Do not edit files. Do not run go tests.
