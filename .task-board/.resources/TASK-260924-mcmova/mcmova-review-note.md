# Review note — TASK-260924-mcmova spec erratum, CR revision 1 (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

SECURITY-RELEVANT normative text. Review against `mcmova-brief.md` in a disposable clone of curator-spec.
1. "Hard-link substitution" is DEFINED (a link that makes the resolved identity differ from the platform-owned file the
   manager intended), not just carved out.
2. The Windows allowance is bounded EXACTLY: default Windows search list only; canonical `%SystemRoot%\System32` derived
   from the MANAGER's captured SystemRoot (never the caller's/package's environment); target physically below it; the extra
   links are the platform component store. Anything broader (other directories, caller-supplied search dirs, symlinks,
   reparse points, python/node interpreters) remains rejected — hunt for any wording that widens it.
3. core.md and profiles/manager.md agree; any conformance vector/case that encodes the rule is updated consistently;
   CHANGELOG (unreleased) entry; no other normative change (diff every changed file).
4. Validation evidence: the three `validate:` recipe lines + regenerate-check exit 0 (results table) and the runtime log green.
accept_cr or changes requested with file:line and the exact wording problem. No LOGBOOK.md.
