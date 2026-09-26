# Review note — TASK-260924-19n6g2 curator-spec v1.0.0-rc.13 release prep (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review against `19n6g2-brief.md`. Disposable clone; bounded runs (host memory is tight).
1. PRECEDENT: diff the shape of this candidate with the rc.12 (and rc.11) release-prep commits (`git log --grep rc.12` / the commits that
   added release/1.0.0-rc.12.json): same kinds of files and edits; every deviation named and justified (e.g. ci.yml / release.yml /
   Makefile / tools changes).
2. `release/1.0.0-rc.9.json` restored to tag v1.0.0-rc.9 bytes: PR #88 (accepted erratum TASK-260924-mcmova) had changed that file — decide
   whether restoring the historical release record is correct (released metadata must stay byte-identical) and that nothing of #88's
   intended effect is lost (where did #88's rc.9.json change belong instead?).
3. The new independent `conformance/skillfile-sources-v1/manifest.json` and the rc.13 metadata tying that suite to core tag v1.0.0-rc.10
   (manifest SHA-256 803918bf…): correct, reproducible by the generator, and consumable by a partial client pinned to rc.10.
4. Every vector's version bump is generator output (regenerate-check clean), no semantic vector change; CHANGELOG rc.13 section keeps every
   Unreleased entry and highlights the accepted suite, no v9 `directory`, the #88 erratum and per-implementation CI claims.
5. `tools/release_gate.py --version 1.0.0-rc.13` passes on the candidate; the tag command in results is correct; release.yml changes do not
   weaken signing/provenance.
accept_cr or changes requested with file:line. No LOGBOOK.md.
