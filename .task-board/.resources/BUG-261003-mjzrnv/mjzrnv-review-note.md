# Review note — BUG-261003-mjzrnv freeze tests tag-aware (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Same-provider review (operator rule, tb-R164). Brief: freeze-fix-brief.md (attached to the producer run). Release blocker for re-tagging v1.0.0-rc.14.

Verify independently (use a venv with jsonschema; never push tags):
1. Both states pass: (a) no rc.14 tag: `python -m unittest tools.test_validate` and `make validate`; (b) a temporary LOCAL lightweight tag `v1.0.0-rc.14` on the candidate: same two commands pass. Delete the local tag afterwards.
2. The protections are not weakened: tools/validate.py (the production guard) is unchanged; a byte mutation of release/1.0.0-rc.13.json still fails validate; with rc.14 tagged, a mutation of release/1.0.0-rc.14.json fails too; with rc.14 untagged, the active record may still change.
3. The inventory derives published records from real release tags, not a fixed list; check it cannot silently pass with zero tags discovered (e.g. shallow clone / no tags fetched). If it can, that is a defect unless the test explicitly fails closed.
4. Scope: only tests + one CHANGELOG line; conformance/, schemas/, release/*.json untouched; manifest digest 6f832d81… unchanged (`make regenerate-check`). LOGBOOK.md untouched.

accept_cr if all hold; otherwise changes requested with concrete findings. Record real exit codes.
