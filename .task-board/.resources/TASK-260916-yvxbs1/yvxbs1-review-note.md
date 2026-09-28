# Review note — TASK-260916-yvxbs1 E6 path-kind admission and boundary (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review rev5 (base 86552087 = trunk, tree 3e55fd15, 32 paths, gate green on every lane) against `yvxbs1-sec-brief.md`, `yvxbs1-decision-1.md`,
`yvxbs1-gatefix-3.md` and curator-spec rc.13 (cite clauses). Verify through the production entry:
1. A path-kind package declaring an MCP server is refused with the spec diagnostic (reuses the trunk MCP surfacing).
2. class: system modules in path packages follow rc.13 §3: a trusted DIRECT path overlay's system module is admitted
   (path-overlay-system-module-admitted); a transitive one refused via the existing admission (E2, trunk).
3. environments.md §4 "`path` source directories": the five boundary checks (ownership, private mutation permissions / owner-only DACL,
   containment, regular file types, lstat link safety) run for every path source the lock names "on every `env resolve`, and again under the
   manager-home mutation lock for every mutating profile operation — install, update, use, sync, repair, garbage collection". List the call
   sites and check each operation is covered; a failing directory is entry-class environment_store_untrusted (no fragment, non-current,
   NO rebuild/re-copy, dry-run reports environment_store_untrusted, never would-rebuild-untrusted-store). NOTE: an earlier orchestrator note
   wrongly said "only at read points" — the spec wins; flag any operation left unchecked.
4. The Windows DACL check is not weakened (inherited SYSTEM/Administrators ACEs on an operator directory fail); Windows fixtures create
   protected trees; the ownership rows are driven through the owner-lookup seam (not bounded).
5. `profile update default` still refuses with profile_update_blocked before any source.json read.
6. rc.13 vectors driven; gap rows before/after; mutants (MCP refusal removed; boundary check skipped on env resolve; DACL check relaxed; link
   check following links) killed with real exit codes. Rebase fidelity: nothing of trunk (E4/E3/ryh3kw) reverted — compare the conflicted
   files (envprofile.go, managed.go, status.go, envregistry.go, root-artifacts.tsv) with trunk. No CHANGELOG/LOGBOOK, no stray files.
Bounded runs. accept_cr or changes requested with file:line. No LOGBOOK.md.
