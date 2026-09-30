# TASK-260916-2x2f7h review verdict — CR rev1: ACCEPTED

Base 4ad8042b, tree bd777646. Changed files: CHANGELOG.md and protocol/environments.md only. No LOGBOOK.md and no stray files.

1. S5 residual, environments.md §10.1 (lines ~2864-2877). The old sentence "`env resolve --repair` is not persistence for a tampered store" is replaced with an accurate statement. When repair finds a stale home, it re-materializes the store bytes that passed their checks, so an attacker who can write the store and set a matching pin gets the payload restored on each such launch. The accepted bound is the operator-owned protected-state and pin contract (§4), and the text states that S5 promises no freshness or revocation. It cites E7 and §4, and CHANGELOG calls it "S5 §10.1". Scoping the replay to "stale home" is more accurate than the brief's "every launch": §10.1 repairs only when lock-free verification finds the home stale. The removed sentence was informative, not a MUST, and every MUST remains ("MUST NOT adopt candidate bytes", the §8.3.1 writes).
2. The §7.8 asymmetry row is in environments.md §7.8 (line 1770), after the per-adapter residual table, so the brief's location is correct. Runtime enforcement: Claude's strict flag excludes other configuration; Codex's -p merges over the base and does not suppress base mcp_servers. Manager responsibility cites the existing launcher MUST (pre-exec stat) and the §7.4 revision A/B seed. The row adds no new MUST and is consistent with the rows at 1759/1761 and with §7.4 at 1547.
3. No MUST changed, so no vector was needed. regenerate-check exit 0 with no diff.
4. The CHANGELOG entry is under Unreleased.

Validators, rerun by the reviewer:
- tools/validate.py: exit 0 (64 schemas, 1170 vectors). Run in a venv with jsonschema; the system python lacks that module, so `make validate` exits 2 on this host for an environment reason.
- go test ./tools/...: exit 0.
- make regenerate-check: exit 0.
- Python unittest suite (~20 min): not rerun. I accept the producer's 641 OK for docs-only changes.
