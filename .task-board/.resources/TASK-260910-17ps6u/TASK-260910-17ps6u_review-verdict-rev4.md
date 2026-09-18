# Revision 4 independent review

Verdict: ACCEPT. Revision-3 F1 is resolved; no actionable findings remain within this leaf's scope.

Candidate d7ce6af19764479a9d85258a46e04d3cae375c47; base 7ce27b20c43baf489d4d11e9c55dc05eb2dc108e. All 13 changed files compared byte-for-byte with candidate before and after verification. Revision 3 -> 4 changes only internal/marker/marker.go and marker_v5_test.go. No repository production/test files modified; adversarial tests and mutant use Go overlays in ignored .temp.

## Finding closure and production coverage

marker.go:408 validates the retained raw package object before accepting decoded identity. Explicit closed shapes match the three source-types-v1 package arms and lockedCommit member sets: required typed non-null members, foreign fields forbidden even null/empty, nested commit closed. protocoljson.Validate at Read rejects duplicate keys and trailing data before decoding. It is a transport validator, not a schema validator, so the additional raw-shape checks are appropriate. Read still uses its existing nil-on-invalid API (it does not expose a typed error); CLI marker-band classification remains invalid-marker for readable malformed schema 5.

Independent TestReviewRawUnion: 24/24 malformed cases refused by Read and never Current. Includes local source/repository/directory/commit null and empty, network source/snapshot null and empty, configured repository/snapshot null and empty, nested commit extra fields null/empty, duplicate keys in every arm, trailing document. Existing valid controls and required-field/type/duplicate nested-commit regression table independently rerun via the whole marker package.

Independent TestReviewInstallMalformedMarker: 24/24 cases consumed through install.Project (draftLocalInstall -> Project -> stageNode -> marker.Current). Each malformed recorded marker is non-current and replaced with a valid frozen local marker, never reported up-to-date. The Git-arm rows here are hostile recorded markers on a local installation, not a claim of end-to-end Git acquisition coverage. Valid initial install is the control. Existing malformed installed-marker repair semantics are preserved; no new refusal-before-repair semantics invented.

Narrowing mutant retains decoded semantic validation but drops only validV5PackageShape. Killed independently (1/1 mutant, zero survivors): producer regression table fails 14 cases; independent reader table fails 15 cases with readable/current=true; production install probe local-source-null reports up-to-date. This demonstrates the raw gate is called from installation, not merely helper-tested. No further live-link/lock mutants rerun this revision; unchanged prior review evidence remains bounded as recorded in review-verdict-rev3.md.

## Independently executed checks

Shell zsh; Go tests use -p 1 -count=1. Long commands were retained and polled until completion, not left running at turn end.

- go test -p 1 ./internal/marker -count=1 -timeout=120s: exit 0, 0.625s.
- go test -p 1 ./internal/install -run 'TestDraftLocal|TestLegacyInstallUntouchedWhenDraftOff' -count=1 -timeout=240s: exit 0, 19.722s. Frozen scripts/dependency materialization, refresh, tampered/missing snapshots, script policy, system/skill dependency refusal, and draft-off legacy checks pass.
- go test -overlay .temp/review-17ps6u-r4/probe.json -p 1 ./internal/marker -run TestReviewRawUnion -count=1 -timeout=120s: exit 0, 0.412s.
- go test -overlay .temp/review-17ps6u-r4/probe.json -p 1 ./internal/install -run TestReviewInstallMalformedMarker -count=1 -timeout=240s: exit 0, 80.336s.
- go test -overlay .temp/review-17ps6u-r4/mutant.json -p 1 ./internal/marker -run 'TestReviewRawUnion|TestMarkerV5PackageClosedShape' -count=1 -timeout=120s: exit 1, mutant killed (0.700s).
- go test -overlay .temp/review-17ps6u-r4/mutant.json -p 1 ./internal/install -run '^TestReviewInstallMalformedMarker$/^local-source-null$' -count=1 -timeout=120s: exit 1, mutant killed (9.757s).
- go test -p 1 ./internal/runtimestore ./internal/scopes -run 'SourceV1|Draft' -count=1 -timeout=90s: exit 0, 0.548/0.411s.
- go test -p 1 ./cmd/curator -run '^TestMarkerRefusalSeparatesUnsupportedFromInvalid$' -count=1 -timeout=90s: exit 0, 0.564s.
- go vet ./internal/marker ./internal/install: exit 0 (standalone verification).
- gofmt -l internal/marker/marker.go internal/marker/marker_v5_test.go: no output; git diff --check: exit 0.

Probe generator and logs attached in TASK-260910-17ps6u_review-evidence-rev4.tar.gz. Run make-probes.py from the candidate root to reconstruct relocatable overlay paths.

## Hosted identity and inherited bounds

Independently queried and downloaded full 921194-byte log of https://github.com/relux-works/curator/actions/runs/35315795438. Conclusion success. Gate head 69fffd3a470aaf483dda215caadf7e61d3b9f8b0 resolves to tree d7ce6af19764479a9d85258a46e04d3cae375c47 exactly. Hosted Ubuntu/macOS/Windows tests, Ubuntu/macOS race, lint, interop, naming and gate self-tests green. Rose-air and candidate-suite skipped; no passing claim for those lanes. Full-suite/golden evidence reused from this exact hosted candidate, not rerun locally.

The scope ruling from rev3 stands: skillfile-sources.md section 4 lines 190-248 mandates marker package/lock and source-v1 GC retention. Local runtime materialization/keys/refresh/GC, marker-band test, and Windows fixture fixes are unchanged. The Windows materialization assertions run before the declared POSIX skip; the subsequent reinstall assertion is also skipped, as already bounded in rev3. Dependency fixture executes provider and consumer shims separately; no claim for lookup inside the consumer script. Git runtime/marker migration and receipt-v3 remain sibling/integration work, not completed by this local-snapshot leaf. No frozen-v1 schemas or goldens changed. Accepted source-audit sibling rev10 results and verdict read; its typed gate remains untouched.

Goal queried before verdict: run not goal-bound; no directives. Live checklist complete. Campaign forbids LOGBOOK.md edits; review recorded as board outcome and notes. Acceptance routes to integrating, not done; no commit/commit_ack supplied.
