# TASK-260910-hwxr26 — independent review revision 10

Verdict: ACCEPT. No actionable findings remain within this leaf's scope.

## Candidate and hosted evidence
- CR-TASK-260910-hwxr26-10; base 9b185d2503d7bfb4fe66a33ad591a50abe9f8c90; candidate tree 90f20e92641d282d2128053546ba6006bf9362c8.
- All 14 changed working files compared byte-for-byte to the candidate before and after review. No production or committed test files modified; probes and mutant used Go overlays under ignored .temp.
- Independently queried GitHub run 35300110268 and downloaded its complete log (914303 bytes). Success; head 245e763033c3e1028874e7a2eb896a224724a30f resolves locally to the exact candidate tree. https://github.com/relux-works/curator/actions/runs/35300110268
- Hosted Ubuntu/macOS/Windows tests, lint, conformance and Ubuntu/macOS race lanes passed. Rose-air and candidate-suite jobs skipped: no claim of independent ARM64 validation. Full hosted tests/goldens reused from this exact candidate, not rerun locally.
- Vendored source-audit/source-types schemas match the accepted curator-spec checkout byte-for-byte (cmp exit 0 each).

## Independent checks (zsh; all go invocations -p 1 -count=1)
1. go test ./internal/install -run '^TestDraftAuditUnreadableBindingRefuses$' -timeout=90s: exit 0. Five of five subtests, including both positive controls and three broken-link cases. Each negative runs dry-run and mutation, checks unchanged report/link and absent dangling target: 6/6 refusal calls.
2. go test -overlay .temp/review-hwxr26-r10/overlay.json ./internal/install -run '^TestReview(Rev9|ReadFailures)$' -timeout=90s -v: exit 0. Nine of nine adversarial subtests: escaped duplicate object/report keys, foreign-arm field, null report field, dangling binding, self-loop with corrupt report, directory, unreadable permissions and orphan report. Prior reviewer probe reproduced against this candidate; additional non-symlink cases assert records unchanged. Permission fixture positively checked unreadability. No skips.
3. go test ./internal/install -run '^TestDraftAudit(DuplicateKeysRefuse|MalformedShapeRefuses|ClosedPackageShapeRefuses|RenewalNeverHidesRefusal|RenewalMarkersNeverRenewable|PolicyRenewalBrokenEvidence|BrokenReportRefuses|FirstIssuanceVsBrokenRecord|BindingLifecycle|PinAdmits|RevocationBlocks|StrictLocalRequiresAttestation|AdvisoryLocalPasses|EnforcedScriptRefused)$' -timeout=100s: exit 0, 30.809s.
4. go test ./internal/install -run '^TestLegacyInstallUntouchedWhenDraftOff$' -timeout=90s: exit 0.
5. go test ./internal/audit -timeout=60s: exit 0.
6. go test ./internal/artifactpolicy ./internal/scriptpolicy ./internal/registry -run '^(TestEffectiveLabels|TestCheckLocalPackage|TestResolveWithoutIdentityIsUnknown)$' -timeout=60s: exit 0, all three packages.
7. git diff --check: exit 0.
8. Narrowing mutant: overlay replaces sourceAuditBindingVacant with os.Stat(objectPath), returning true on any error and ignoring report vacancy. go test -overlay .temp/review-hwxr26-r10/mutant.json ./internal/install -run '^TestDraftAuditUnreadableBindingRefuses$' -timeout=90s -v: exit 1. Three of three negative rows kill it; two of two controls pass. Dangling link bypasses source audit; loop cases reach writes and fail only with ELOOP. Candidate files never changed.

An initial broader four-package command (audit/artifactpolicy/scriptpolicy/registry, timeout=90s) was terminated after about 75 seconds while artifactpolicy stalled; shell exit 143. Audit had passed. This is NOT counted as a passing full-package run. Separate audit and narrow policy/registry checks above completed; wider coverage comes from the exact-tree hosted evidence.

## Scope and earlier findings
- install.Project calls the draft gate at install.go:465 before shared audit, registry resolution, cache and compiler; DraftSourcesV1 separation retained.
- sourceaudit.go:911-940 proves both entries vacant using Lstat; present/non-ENOENT errors cannot authorize first issuance. CheckSourceAudit refuses broken existing records before StoreSourceAudit. Directory, permissions, dangling/self-loop and orphan-report probes confirm this.
- Existing raw bytes reach protocoljson.Validate before lossy decoding for BOTH object and report; closed unions, required non-null shapes and trailing document checks remain effective.
- Identity/context/report integrity and completeness, live findings/pins/revocation, policy labels, decision and future timestamp checks precede renewable stale/policy outcomes. Renewal uses typed errors, not untrusted diagnostic text. Prior regression tables and valid controls rerun successfully.
- Local snapshots cannot satisfy required network attestation; pins, revocations and script/assurance checks retained. Source audit does not mint registry attestations.
- Symlink fixture skip reason 'symlinks unavailable: ...' matches the existing host-capability pattern in skip-classes.tsv:69; no added class. Local symlink rows executed, not skipped.
- Positive controls prove source-audit gate establishment/validation, not completion of sibling marker-v5 publication; that remains explicitly outside this non-final leaf.

Run reports no bound goal; checked again before verdict. Checklist already complete. No LOGBOOK.md edit per campaign constraint. Evidence and review note persisted through task-board before acceptance.
