# Review note — TASK-260917-3j8sjw registry service: content-hash framing version (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review rev1 (curator-skill-registry, 9 paths, +247/-11, CI green) against `3j8sjw-brief.md` and curator-spec b1a2efb (registry.md
equal-version matching; registry-log-entry-v2 / registry-bundle-v2 / log-response-v3). The spec pin in ci.yml moved from 47c3c8cb to
b1a2efb. Verify:
1. Records and snapshots carry hash_version in the spec's v2 shapes. Existing v1 records stay v1, and the serialized bytes of v1
   records are unchanged.
2. Matching refuses a version mismatch both ways. Publication refuses a record whose declared version disagrees with its hash form, with
   a stable error. Bundle import validates the version. P4 high-water and the R-series hardening are intact (their tests still pass).
3. The spec vectors that apply to the service are driven from the pinned b1a2efb. Name which ones, and which ones do not apply.
4. Run the suite yourself (pytest) and give the real exit code. Mutant: remove the version comparison in matching → a test fails.
5. The deployment note in docs/content-hash-versions.md is accurate. CHANGELOG entry present. No stray files; nothing touches unrelated
   service behaviour.
accept_cr, or changes requested with file:line. Never spell any employer name.
