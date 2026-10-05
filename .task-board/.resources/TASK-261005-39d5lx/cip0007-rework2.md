# THE ONLY CURRENT INSTRUCTION — TASK-261005-39d5lx rework 2 (researcher; curator-spec). One finding only.
The reviewer (RUN-261005-5e340e) confirmed that all three High findings and the other Medium findings are addressed. ONE Medium finding remains (repeat of rev1/F1); read `TASK-261005-39d5lx_review-verdict-rev2.md` finding 1.
- `.research/261005_manager-provisioned-cli-tools.md:155` puts the two prohibited vendor tokens inside the published scan expression, and lines 100–102 claim zero matches.
- Fix: remove the literal token expression from the public note. Keep only GENERIC scan labels (e.g. "two vendor names from the issue text") and the real counts.
- Write the scan pattern ONLY into a 0600 temp file outside the repo, and read it from there. Never put it on argv or into any committed file.
- Rescan the FINAL saved bytes of both documents after writing. Record the pre-correction and final counts truthfully.
- Add the named check `final-public-doc-name-scan` to the evidence. Its negative control places a prohibited token only in a temp copy of the evidence doc and detects it; a CIP-only scan must miss it. Keep the control ephemeral and never commit the token.
No other edits. Rerun `python -B tools/validate.py` (exit code). Then `task-board handoff TASK-261005-39d5lx --role researcher` and END YOUR TURN.
