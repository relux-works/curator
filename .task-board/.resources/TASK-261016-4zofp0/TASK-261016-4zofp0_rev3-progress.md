# SNAP-LOCK revision 3 progress and logbook

The converged candidate preserves all revision-2 deltas and adds only TestMaterializeCaptureCredentialStoreAdmission in internal/envprofile/materialize_capture_test.go. README retains the capture paragraph and current login-reference/tooling paragraphs once each. No behavior changes.

The supplied revision-2 verdict is ACCEPT with zero findings, not a rejection. No repeat-of class exists. The new regression is driven by the convergence brief.

Normative review: curator-spec main 15d970111277b1f68b3f7a4ddd63bc339b2c0a3f; snapshots sections 1/5, CIP-0014 consistency 1-5, environments 9.2/10.1, and login-readiness section 2. Snapshot read-only verification preserves existing credential refusals. No conflicting capture exemption exists. One direct admission call remains in resolveVerified; both Resolve and capture share it. The separate existing checkIsolatedCredentialStore recheck is retained; this is not a claim that admission executes only once across the whole successful call graph. Capture intentionally resolves before and after its callback.

The new test pairs native capture with explicit Codex file/keychain refusal, checks exact Resolve diagnostic equality, nil captured output, no callback invocation, and unchanged fixture/credential state. The file-narrowing plant is rejected at the capture assertion itself; keychain admission remains guarded.

Current evidence: focused race suite exit 0; go build ./... exit 0; unchanged reviewer tests through Go overlay exit 0; all 10 standalone narrowing commands exit 1 at named failing tests, harness exit 0. Existing nine controls were rerun on the converged identity; no old mutation result is silently reused.

Signed full-gate snapshot 4cf8cc94ef36be4e95888d57b48f3cf2314b9b47, tree e9ce37ba4c43514de32154fb689df7808a7cb1e8, signature G. Hosted run https://github.com/relux-works/curator/actions/runs/38919350971 remains pending. Full-gate checklist item is unchecked until the actual process exits green. Story HEAD is unchanged and candidate remains uncommitted. This progress record is not a review handoff.
