# TASK-260924-11burj — playbook collection acceptance, end to end (THE ONLY CURRENT INSTRUCTION)

Read `campaign-producer-rules.md` and the binding operator memo `skillfile-operator-memo-20260924.md` (acceptance criteria). Curator main
now has Skillfile schema 2 ON by default (1aa9wb), the fresh-machine lock replay (m28s6b) and the manifest dependency `directory`
selector (1kpw4w). The real relux-works/curator-playbook repository has no skills/ folder or tags yet, so build FIXTURES with the
§8 layout of curator-playbook spec/process-configuration.md:
- repository A (local bare git, tag v1.0.0): skills/orchestrator, skills/developer, skills/reviewer (each a valid skill; one of them
  declares a manifest dependency with `directory` on a subfolder skill of repository B);
- repository B (local bare git, tag v1.0.0): roles/qa/SKILL.md etc. (a skill NOT at the repository root).
Drive the REAL CLI (built from this tree, production entry — not internal calls) in temp homes:
1. Skillfile schema 2 with ONE collection entry {from: <repo A>, directory: "skills", include: ["*"]} → `curator install` installs
   every skill in skills/ plus the transitive subfolder dependency from B; Skillfile.lock.json written; audit records present; `status`
   up to date. Then `curator update` after adding skills/writer at tag v1.1.0 → the new skill appears, lock/audit updated.
2. Fresh machine: new empty home, same project with the committed Skillfile.lock.json → install replays from the lock (exact locked
   commits, identity + content_sha256 match), lock byte-identical, no re-resolution; tamper case → source_snapshot_changed; unreachable
   source → source_snapshot_unavailable.
3. Negative rows: include pattern matching nothing; directory escaping the repo; dependency directory with no SKILL.md.
Commit the scenario as an automated test (cmd/curator or an e2e package) using only local fixtures (no network), bounded runtime.
CHANGELOG POLICY: do not edit CHANGELOG.md — put release-note text for BOTH capabilities (collection selector; manifest dependency
directory) in results under "## CHANGELOG entry (for release prep)". Attach results (scenario table with commands, exit codes, lock
excerpts), check DoD, `task-board handoff TASK-260924-11burj --role developer`. A write-boundary `policy warn` block is a warning.
