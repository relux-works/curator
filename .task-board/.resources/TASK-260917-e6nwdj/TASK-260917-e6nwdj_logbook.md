# TASK-260917-e6nwdj review logbook

- Review-harness finding: `git archive e8b53a0` expands the export-subst fixture `conformance/v1/fixtures/byte-exact/subst.txt`. The initial archive-copy `make regenerate-check` and `make validate` both exited 2 and are not evidence against the delivery candidate. Rebuilt the disposable copy from raw `git cat-file blob` bytes. Regeneration then exited 0, with 1382/1382 tracked raw blobs still identical to e8b53a0. Future byte-exact reviews must use raw blobs or a verified byte copy, not git archive.
- Non-blocking editorial findings: protocol/environments.md:3048 joins the two conformance blocks with `afterwards. ;`; CHANGELOG.md:37 ends the E5 bullet at `checked`, with the shared explanatory tail following the S5 bullet at line 90. Neither loses a normative rule. No source edits made.
- Production-entry narrowing attack: reduced five store-boundary branch shapes to ownership-only, consistently adjusting `failing_check`, regenerated the manifest successfully, then ran `python3 tools/validate.py`. Exit 1 specifically refused the symlinked-entry-root branch missing link_safety. This demonstrates semantic refusal beyond manifest digest validation; it does not demonstrate manager filesystem enforcement.

Persisted as a task-scoped logbook outcome because no logbook CLI or connector is exposed in this session.
