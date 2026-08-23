# Approved rc.8 recovery plan

The user accepted the recommended recovery: retain the failed signed v1.0.0-rc.7 tag and publish a reviewed v1.0.0-rc.8.

Requirements:
- Never move, delete, or rewrite rc.7.
- Fix the release workflow at the tagged commit so validation cannot create Python bytecode or other untracked generated files before the clean-checkout release gate. Prefer deterministic process-level prevention plus a regression test.
- Update every authoritative normative version, conformance pin, and release metadata surface to rc.8 while preserving historical rc.7 metadata.
- Run validation, regeneration, release-gate, review-report, clean-tree, signature, and release-asset checks.
- Use a reviewed PR and independently reviewed producer handoff before merge.
- Publish an annotated signed v1.0.0-rc.8 tag and GitHub prerelease only after acceptance.
- Keep verified implementation and platform conformance claim lists empty.
- Record commit, tag object, suite digest, release URL, asset digests, CI links, and explicit absence of verified claims as outcome evidence.