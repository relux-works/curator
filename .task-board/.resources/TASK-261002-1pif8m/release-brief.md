# THE ONLY CURRENT INSTRUCTION — TASK-261002-1pif8m rc.14 / rc.3 readiness checklist (researcher, read-only)

The operator approved releasing curator v0.15.0-rc.3. Planned order: spec v1.0.0-rc.14 → curator v0.15.0-rc.3 → a curator-agent-launcher tag. Build an evidence-backed checklist of what must go into these releases and what is still open.

Read:
- curator-spec main: RELEASE.md (the release procedure), the CHANGELOG Unreleased section, release/*.json records, and .github/workflows/implementations.yml pins (Go = curator e4a6a8d5, csk = 4a88aa0e).
- curator main: CHANGELOG Unreleased, the release workflow, and docs/release*.
- The board (`task-board q`): open leaves in EPIC-260910-2hw1xb (security remediation, manager and spec) and EPIC-260910-16qce1. Also any leaves that name rc.14 or rc.3, e.g.:
  - TASK-260927-31gaka / 1e5qqm / 25hk87 (revision A/B seeds);
  - the content-hash v2 write cut-over switch "OFF until SPEC_PIN rc.14";
  - the B3 PRs curator-spec#119 / curator#101 (still CHANGES; owned by macmini-infra).

Answer:
1. **Spec rc.14:** which merged changes since rc.13 it ships; which open items must or should land first; the exact release steps per RELEASE.md (version bump files, release record, schema freeze, tags).
2. **Curator rc.3:**
   - which SPEC_PIN it needs;
   - which switches flip at the rc.14 pin (content-hash v2 writer, others);
   - which open leaves block it;
   - the release workflow steps and assets.
3. **Launcher tag:**
   - the version (v0.1.1 vs v0.2.0, given the muse and interactive features);
   - its README install line;
   - compatibility with curator rc.3.
4. **One ordered plan:** leaf-sized steps, each with dependencies, a rough effort, and what can run in parallel. Mark the items that need the operator's "да" (publishing tags and releases).

Rules:
- Read-only: no tags, no releases, no commits, no posts.
- Never put secrets in the results. Never spell any employer name.

Write the results. Then run `task-board handoff TASK-261002-1pif8m --role researcher` and END TURN.
