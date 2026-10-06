# Reviewer instruction — TASK-261006-3ptm2x: NUL opaque gate scoped to v1 identities (security, tb-R195 strong review)
Cross-provider review: the producer was muse max; the reviewer is codex sol high. This is security-relevant: it relaxes a blocking guard.
Verify against curator-spec `protocol/core.md` §8 (the interim rule and identity versioning):
1. The producer's call-site inventory is complete: `git grep opaquescan.NULPaths` plus every v1 identity compute/trust site. Every site's version decision comes from recorded `hash_version` or the frozen shape, never from the digest bytes.
2. There is NO path where a v1 identity is computed or trusted over NUL bytes. Attack it with at least: a legacy marker with no hash_version over a NUL tree; a mixed tree; a status re-check of a v1-installed skill after a NUL file appears; a v2→v1 downgrade attempt; and the verdict cache keyed on a v1 digest.
3. v2 installs, audits and status of a NUL-bearing tree work at the production entry, with the correct v2 identity.
4. contextaudit is consistent.
5. The tests are production-entry, and the named mutants are really killed. Run them in a disposable clone, and only for packages without fake executables (R194). For cmd/curator and install, use the hosted gate evidence.
6. CHANGELOG is updated; no LOGBOOK or remote-gate.sh edits.
accept_cr, or request changes with numbered findings.
