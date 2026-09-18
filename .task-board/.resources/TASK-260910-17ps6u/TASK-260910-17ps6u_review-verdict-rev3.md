# Revision 3 independent review

Verdict: CHANGES_REQUESTED (to-dev).

Candidate tree 7369ebf1456a968376c7c0a8bf3eb804d12878c9; base 7ce27b20c43baf489d4d11e9c55dc05eb2dc108e. All 13 changed files compared byte-for-byte with candidate before and after review. No repository source/test edits; adversarial tests and mutant run through Go overlays in ignored .temp.

## F1 — malformed package union accepted as current (P2)
internal/marker/marker.go:428-438 validates decoded values instead of raw package field presence. The shared Package struct recognizes fields from every union arm, so DisallowUnknownFields does not enforce each arm. JSON null becomes an empty string/nil. A local-snapshot package containing source:null, repository:null, directory:null, or commit:null is accepted by marker.Read and marker.Current returns true. The source-types-v1.schema.json package local-snapshot arm explicitly has additionalProperties:false and admits only kind/snapshot; install-marker-v5 references that package contract. Empty foreign string fields have the same value-collapse problem by inspection.

Reproduction: write a valid v5 local marker using existing writeV5/v5LocalMarker helpers, decode its JSON, add one of those fields with null under package, write it back, call Read and Current against the original expected marker. Independent overlay TestReviewV5ForeignNullFields reproduced 4/4 malformed packages accepted and 4/4 current=true (expected rejection). Command: go test -overlay .temp/review-17ps6u-r3/probe.json -p 1 ./internal/marker -run '^TestReviewV5ForeignNullFields$' -count=1 -timeout=60s; exit 1, 0.335s. Probe source and log attached separately.

Required rework: validate the raw package object against the closed applicable arm before lossy decoding; enforce field presence/type/null constraints, including nested commit, and retain valid controls. Add reader/currentness regression tests for forbidden fields with null and empty values across the arms. Do not solve by silently dropping foreign fields. Re-review required.

## Independent successful checks
Shell zsh; every go test uses -p 1 -count=1.
- ./internal/install -run TestDraftLocal -timeout=180s: exit 0, 20.650s. Includes actual frozen runtime materialization, refresh, tampered/missing snapshot, enforced script policy, missing system command and skill dependency refusal rows.
- ./internal/marker ./internal/runtimestore ./internal/scopes -run 'V5|SourceV1|Draft|SchemaBand' -timeout=90s: exit 0 (0.366/0.396/0.388s).
- ./internal/install -run '^TestLegacyInstallUntouchedWhenDraftOff$' -timeout=90s: exit 0, 0.886s.
- Narrowing mutant replaces only install's source-v1 key preimage with member.Package.Snapshot (retains namespace). TestDraftLocalRuntimeMaterializesFromFrozenSnapshot kills it: exit 1, runtime script missing at required package-derived path (6.537s). Independently measured mutants 1/1 killed; no claim for unrun live-link/lock mutants. Producer reports additional mutants; not independently reproduced here.

## Hosted evidence and scope
Independently queried/downloaded complete log for https://github.com/relux-works/curator/actions/runs/35312241911 (success). Head 06d45efad5b613241a7be83f1face3ff7cebf9f5 resolves to exact candidate tree 7369ebf1456a968376c7c0a8bf3eb804d12878c9. Hosted Ubuntu/macOS/Windows test lanes and race/conformance/lint evidence reused, not full-suite rerun locally. Rose-air skipped; no rose-air claim. Windows materialization assertions precede POSIX skip; refresh test runs fully. Note skip also bypasses subsequent reinstall-current assertion on Windows, a coverage bound.

Marker/GC scope extension is justified by skillfile-sources.md lines 190-248: marker v5 package/lock and source-v1 GC retention are mandated carriers of protected runtime state. Git-member migration/build receipt v3 remain explicitly deferred sibling/integration work, not accepted as complete here. Existing source-audit sibling rev10 results/verdict read; its accepted gate untouched. Marker-band test update faithfully distinguishes readable-invalid v5 from NewestSchemaVersion+1; no weakening found. No frozen v1 schema/golden files changed. Legacy local check is a bounded test; broad golden coverage comes from exact-tree hosted gate.

Dependency positive fixture declares a dependency and executes both shims separately; it does not execute dependency lookup from inside consumer script. Thus consumer-internal lookup coverage is not claimed.

Board run is not goal-bound (queried). No external blocker; ordinary implementation rework. Campaign forbids LOGBOOK.md edits and no logbook CLI is available per existing evidence; findings persisted as this outcome plus board notes.
