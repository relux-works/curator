# Review note — TASK-260926-4hd81z (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Blocking curator-spec v1.0.0-rc.13 (PR #97). Review against `rc13pv-brief.md`: one file (internal/scriptpolicy/conformance_test.go);
the protocol-version check accepts exactly {1.0.0-rc.9, 1.0.0-rc.13} with an accurate comment; schema_version / execution_policy /
interpreter-set bindings unchanged; `go test ./internal/scriptpolicy -count=1` passes against curator's pinned suite AND against
conformance/v1 of curator-spec branch release/v1.0.0-rc.13 (head f6bd748c) via CURATOR_CONFORMANCE_ROOT — rerun both yourself; the
any-version mutant (and an rc.14 label) is killed. Hosted gate green. accept_cr or changes requested. No LOGBOOK.md.
