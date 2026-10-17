# TASK-261016-ont5q1 \u2014 M1 stop-line evidence

Status: blocked before implementation. No tracked files, index, commits or branch tips changed. This is not a review handoff or a capability claim.

## Identities and source reads

- Curator checkpoint, HEAD and freshly fetched origin/main: `782823d849841ae74efa160243cbac56721aa733`.
- Freshly fetched curator-spec origin/main: `15d970111277b1f68b3f7a4ddd63bc339b2c0a3f`.
- Read the attached common rules and owner decisions, CIP-0017, environments sections 5.4, 7.1, 7.4, 7.9, 8.2\u20138.4.1, 10.1\u201310.3 and 12, the v4 conformance README, login-readiness and rc7-carriers. References below are bound to that spec OID.
- Go 1.25.5; darwin/arm64; synthetic temporary homes only. No real credential payload was read. No vendor runtime or login was invoked.

## Stop-line: leaf wording contradicts the normative root-surface rule

The attached leaf acceptance 2 says: `No applicable root module means no fact.`

Normative `protocol/environments.md` section 5.4, lines 1070\u20131080, distinguishes two cases:

1. A declared context with zero applicable modules still materializes the header alone; it is a valid root-context surface.
2. A root declaring no context has no root-context surface and writes no file.

Section 10.2, lines 3574\u20133584, requires the fact iff resolve verifies the selected home's root-context surface. It explicitly derives the fact from surface/marker rules, not file size. Thus a verified header-only root in case 1 must emit the fact. Reading the leaf's sentence literally forbids it.

Production already preserves this distinction: `internal/contextmaterialize/contextmaterialize.go:Monolithic` checks `HasContext`, creates the header, then skips inapplicable modules. `internal/envprofile/managed.go:homePlan.rootContext` records the written surface. Implementing an applicable-module-count gate would discard a real verified surface; inferring a surface from any AGENTS.md would be equally invalid.

The brief explicitly orders: where leaf and normative text disagree, stop and report. No product-code workaround or tests encoding the contradictory rule were added.

Recommendation / exact owner input: correct acceptance 2 to say `No declared root-context surface means no fact; a declared, verified header-only surface still emits the fact.` Reissue/clarify that leaf before implementation resumes. Alternative: change the normative rule through its owner first; this would change existing zero-module behavior and is not recommended.

## Dependency observation: common store result is not in the checkpoint

The leaf requires consumption of CIP-0015's common store/admission result and forbids a second credential-store resolver.

- `internal/envregistry/credential_store.go:MachineConfig.CredentialStoreFor` validates the persisted machine selector, defaulting to native; it is not an effective-store result.
- `internal/envprofile/credential_policy.go:ResolveRequest.admitCredentialStore` is an activation guard returning only error. It admits native and rejects explicit Muse file/keychain; its Muse guidance says successor file-policy delivery is not enabled.
- Public `Resolve` calls that guard before provisioning; `Adapter.resolveIsolation` still rejects Muse isolated.
- No common resolved backend/source/store result to consume was found in the checkpoint. M1 must not duplicate that owner implementation.

Needed integration input: identify and supply the CIP-0015 shared-policy candidate/API and serialize the managed.go/registry/status edits, or explicitly assign the common policy extension to this owner. This is a dependency observation in addition to the root-contract stop-line, not a claim that ordinary coding is impossible.

The newer normative rc7-carriers section 4 explicitly supersedes older contradictory diagnostic wording: explicit Muse keychain uses `environment_credential_store_unsupported`, while an unsupported native selector retains `environment_credential_unsupported`. Future implementation must use that coordinated table.

## Executed baseline commands

Each command ran directly, with stdout/stderr redirected to a task-local log; neither used tee or a status-hiding pipeline. Both processes exited before evidence attachment.

| Command | Exit | What it establishes | Limit |
| --- | ---: | --- | --- |
| `go test -count=1 -timeout 90s ./internal/contextmaterialize -run 'Test(NoChapterCase|ReferencedOpenCodeZeroModules|ReferencedNoContext)$' -v` | 0 | Existing no-context and zero-module materialization tests pass | The zero-module case here is the existing referenced OpenCode test, not a new Muse resolve test |
| `go test -count=1 -timeout 90s ./internal/envprofile -run '^TestCredentialStorePolicyActivation$' -v` | 0 | Existing public Resolve explicit-store activation refusals pass, including Muse file/keychain, with native Codex positive control | Does not establish M1 file admission or v4 transport |

Logs are attached as separate task-scoped outcome resources. No prior evidence was reused. No full gate, lint, build, Muse qualification, new acceptance suite or mutation control ran: there is no candidate, and the explicit normative-conflict stop precedes implementation. These are unrun, not passing evidence.

## Coverage and handoff bounds

M1 acceptance coverage: **0 of 4 AC rows implemented and driven for this candidate**. No candidate exists.

| AC | Required production entry | Remaining |
| --- | --- | --- |
| 1 | `envprofile.Resolve` \u2192 `buildFragment` \u2192 `renderFragment` | v4 policy/five-key JSON/env/shell parity, order and HOME invariant |
| 2 | `envprofile.Resolve` / `repairUnderLock` \u2192 `verifyHome` \u2192 `buildFragment` | Root fact and missing/drifted/denied/repair controls; header-only meaning requires correction |
| 3 | `envprofile.Resolve` using CIP-0015 common effective policy | Native/unset/file mapping, keychain refusal, shared pre-override source and isolated no-link integration |
| 4 | `envprofile.Resolve` and `envprofile.StatusOf` | Preserved bytes/liveness plus version-separated status evidence and unknown qualification/refresh/discovery |

All six paired-test rows in the leaf remain unimplemented for M1. The brief supplies that paired-test table but no catalog-indexed attack-surface table; this is an explicit brief gap, not an implicit coverage waiver. There is no coverage map claiming killed mutants. No reviewer tests were removed; repository source is unchanged.

| Mutant | Narrowing | Named failing test | Result / bound |
| --- | --- | --- | --- |
| Omit only renderer value five | Keep XDG output but lose backend in env/shell | None executed | Unrun; no parity-detection evidence |
| Admit one forbidden root/store state | Keep gate while weakening one refusal class | None executed | Unrun; no gate-strength claim |

Out-of-contract work: M2 vendor exclusion qualification and M3/M4 consumer readiness are separate release dependencies under the leaf's Dependencies section; no such capability is claimed. CIP-0015 preparation/login ownership remains with that dependency. No acceptance row is silently treated as out of contract. There is no rework diff or change outside declared paths.

## Task-scoped logbook

- Initial development transition refused because estimate was missing (exit 1); estimated 5, retried development successfully (exit 0).
- Verified Story checkpoint equals fetched default branch; preserved managed Story branch and uncommitted-delivery contract.
- Stopped before implementation on the documented normative disagreement, preserving the common-policy ownership boundary rather than adding a second resolver.
- No DoD implementation, gate or mutant checkbox was checked. No developer handoff attempted because its preconditions are unmet.
