# TASK-261014-4xkh3q: Muse root-context review handoff

Spec: curator-spec eb1f08c6c66cedc52d9e09553fb791367f334704, environments §§7.1, 7.2, 7.9, 10.2, 12, 13.

Candidate base: 5234cf38833554542e68afade15951c6e56ed89e. The base equaled refreshed origin/main before work; zero missing upstream commits. The Story branch remains uncommitted. The remote gate uses an isolated snapshot without changing the Story HEAD or index. An alternate-index tree comparison returned exit 0: the current candidate tree exactly equals hosted snapshot 599cf6b128a73f79c1d675fddbb3b6f03ce83986.

Changes:

- Muse's registry declares home-relative `config/muse/AGENTS.md`, monolithic-only/default monolithic, verified release 1.4.7, and 32768-byte advisory. Existing generic root-context materialization and verification handle the nested path; no caller special case or fragment/schema extension was added.
- The CLI now threads configured forms through `machineFromConfig` into `Resolve`. A hosted diagnostic exposed the previously missing binding; without it, CLI Muse `referenced` silently used the default. The change is generic, not a Muse caller special case.
- Monolithic output has normal marker/hash accounting and drift/refusal/repair. No context declaration writes no root surface; zero applicable modules still writes the exact generation header.
- The Muse standing note reports root delivery, conditional exec exclusion, serve exclusion gap, observed workspace precedence, and unverified model-visible foreign removal, skills, system-prompt, refresh, and internal state/cache layout.
- Existing Muse registry/layout tests retain their names. The old published registry vector predates the amendment: its null root and 1.4.1 release remain historical facts rather than current root admission. Its XDG/layout/auth assertions remain active. The new CLI vectors live at `cmd/curator/testdata/muse-root-context.json`.
- README documents the behavior, tools, and external launch requirement.

## AC coverage and production entry points

7 of 8 AC rows driven through curator production entries; the external exec-argv row is a stated ownership bound under brief point 4.

| AC row | Named test | Production call site |
| --- | --- | --- |
| Monolithic default, path, exact bytes, marker/hash | TestMuseRootContextVectors/monolithic; TestMuseHasNoAdmittedRootContextForm (metadata supplement) | run -> cmdEnvResolve -> envprofile.Resolve -> homePlan.rootContext -> contextmaterialize.Monolithic |
| Zero applicable modules produce header only | TestMuseRootContextVectors/zero-applicable | cmdEnvResolve -> Resolve -> homePlan.rootContext -> Monolithic |
| No context declaration produces no root file/marker surface | TestMuseRootContextVectors/no-context | cmdEnvResolve -> Resolve -> homePlan.rootContext -> Monolithic |
| Referenced refusal, no fallback output or marker | TestMuseRootContextVectors/referenced-refused | cmdEnvResolve(machineFromConfig) -> envprofile.Resolve -> MachineConfig.Form -> Adapter.ResolveForm |
| Changed root refused and reported as drift; repair restores bytes | TestMuseRootContextDrift/changed | cmdEnvResolve -> Resolve -> verifyHome/checkSurfaces; cmdEnvStatus -> StatusOf |
| Removed root refused and reported missing; repair restores bytes | TestMuseRootContextDrift/removed | cmdEnvResolve -> Resolve -> verifyHome/checkSurfaces; cmdEnvStatus -> StatusOf |
| Standing status note and residuals | TestMuseRootContextStatus | cmdEnvStatus -> StatusOf |
| Exec flag iff materialized | External ownership bound; no synthetic argv conformance claim | No Muse exec-argv composer in curator |

The brief supplied no attack-surface table. That is a brief gap; no coverage-map-template rows were provided. The table above covers the concrete acceptance behaviors instead. This is an initial implementation, not rework. Scope is registry/status plus their existing tests, the one-line CLI form binding required to reach the refusal, CLI vectors/behavioral mutants, and required README maintenance. No previously committed reviewer test was removed or renamed. No new gate inspects source text; mutation anchors locate production edits, and the mutant harness executes CLI behavioral tests.

## Point-4 launch handoff and out-of-contract rows

Search of production Go sources found registry `ExecFlags` release metadata and the environment fragment producer/validator, but no Muse exec launch-argv builder or fragment executor. `cmdEnvResolve` returns the fragment; the inert launch in the existing Muse test is test code. No flag or new descriptor was added here.

External owners: curator-agent-launcher/task-board, whichever composes managed Muse `exec` argv. Bind the flag to actual root-context materialization, including the header-only zero-applicable case. Omit it for a profile with no context declaration. The rule is independent of permission mode and channel opt-ins, and must be coordinated in the same release as this target admission.

Exact normative sentence, environments §10.2:

> A managed Muse launch on the `exec` form whose environment materializes the root-context target in section 7.1 MUST pass `--no-foreign-personal-context`.

The same paragraph also states:

> An implementation MUST NOT emit the flag without materializing that target.

Out-of-contract rows:

- Exec argv and its permission/channel-independent positive and negative vectors: brief point 4 explicitly assigns a precise handoff when argv is composed only outside curator. The external consumer must run these vectors; this repository's fragment shape remains unchanged.
- Native Muse loading, skills discovery, serve exclusion, workspace conflict resolution, and model-visible foreign removal: brief points 1 and 3 use the admitted spec evidence and require residual reporting, not live-tool probes or new channel admission. Tests use inert binaries and temporary homes, never live homes.

## Verification

Hosted run [38489868868](https://github.com/relux-works/curator/actions/runs/38489868868) is success on exact snapshot 599cf6b128a73f79c1d675fddbb3b6f03ce83986. Remote-gate returned real exit 0. Twenty jobs passed: three full Test lanes (Ubuntu, macOS, Windows), two Race lanes (Ubuntu, macOS), nine Go-driver lanes, lint, naming, interop and three gate self-tests. The self-hosted trunk-only and optional candidate lanes were skipped by workflow conditions; neither is claimed as run. Full property suites and original reviewer tests are green. Uploaded, complete go-test JSON from all five full lanes was read directly: each has all required named behavior/registry tests, the unchanged dry-run regression, and seven named expected-red mutant kills (child exit 1). No current survivor. The task-scoped test-evidence JSON records the named tests, jobs and kills.

Commands were invoked directly with stdout/stderr redirected to task-local logs, not piped through tee. Local tests were not run under R223; hosted CI is the test arbiter. No already-attached green test evidence was substituted.

| Command / attempt | Real exit | Outcome |
| --- | ---: | --- |
| go build ./... (01) | 0 | Compiled implementation |
| go vet ./... (01) | 1 | Test referenced a nonexistent marker field; corrected to the lock's content-hash version |
| go vet ./... (02) | 0 | Corrected test compiles and validates |
| go build ./... (02) | 0 | Intermediate candidate compiles |
| go vet ./... (03) | 0 | Intermediate candidate validates |
| git diff --check (02) | 0 | No whitespace errors |
| gofmt -l changed Go files; assert empty output | 0 | Formatted |
| sh scripts/remote-gate.sh (01) | 137 | Intentionally stopped obsolete snapshot after self-review found old Muse assertions; hosted run canceled, not green evidence |
| sh scripts/remote-gate.sh (02) | 137 | Intentionally stopped after detecting a wrong expected header weight; canceled uploaded evidence then exposed missing CLI forms binding. Partial diagnostic logs are not passing full-suite evidence. |
| sh scripts/remote-gate.sh (03) | 137 | Stopped obsolete snapshot before applying the required CLI form binding; hosted run canceled, not green evidence |
| go build ./... (03) | 0 | Current implementation compiles |
| go vet ./... (04, 05) | 0 | Corrected fixture and current candidate validate |
| git diff --check (03) | 0 | Current candidate has no whitespace errors |
| sh scripts/remote-gate.sh (04) | 1 | Run 38484575100: macOS Test/Race and Windows Test jobs failed on mutant harness portability; Linux Test/Race passed. macOS also timed out existing TestAuthoritativeDryRunCasesMutateNothingPersistent on the extended package lock hold. These are diagnostic, not current-candidate green evidence. |
| go build ./... (04, 05) | 0 | Helper protocol and portable overlays compile |
| go vet ./... (06, 07) | 0 | Current candidate validates |
| go build ./... (06) | 0 | Standalone current-candidate compile |
| git diff --check (06, 07) | 0 | Current candidate has no whitespace errors |
| gofmt -l changed Go files (02); separate assert-empty command | 0 / 0 | Current candidate formatted |
| sh scripts/remote-gate.sh (05) | 137 | Stopped obsolete snapshot before Windows anchor correction; hosted run canceled, not green evidence |
| sh scripts/remote-gate.sh (06) | 0 | Hosted snapshot 599cf6b128a73f79c1d675fddbb3b6f03ce83986; run 38489868868 success; all configured full-suite lanes green |

## Narrowing mutant evidence

Verified from complete uploaded output on Test/Ubuntu, Test/macOS, Test/Windows, Race/Ubuntu and Race/macOS: every row below returned child go-test exit 1 and failed its named behavioral test. The harness rejects compilation and unrelated failures. Seven of seven mutants killed on each lane (35 named kills); no current survivor. Each child runs `go test -overlay <temporary overlay> -count=1 -timeout=5m -v -run ^<parent>$/^<case>$ .` in the CLI package, driving the production entry point. The existing CLI helper process protocol preserves git isolation while sharing the parent-held toolchain lock.

| Mutant | What it narrows the gate to | Named test required to fail | Survival bound |
| --- | --- | --- | --- |
| configured-form-ignored | Keep the CLI form binding but ignore exactly Muse referenced, preserving other configured forms | TestMuseRootContextVectors/referenced-refused | Killed on all five lanes, child exit 1; no current survivor |
| referenced-fallback | Accept exactly Muse referenced by silently selecting monolithic; all other unsupported forms still refuse | TestMuseRootContextVectors/referenced-refused | Killed on all five lanes, child exit 1; no current survivor |
| no-context-admitted | Admit only Muse profiles without a context declaration | TestMuseRootContextVectors/no-context | Killed on all five lanes, child exit 1; no current survivor |
| wrong-target | Keep the root materializer active but put only Muse output at home/AGENTS.md | TestMuseRootContextVectors/monolithic | Killed on all five lanes, child exit 1; no current survivor |
| wrong-bytes | Keep the header and target, append extra bytes only for Muse | TestMuseRootContextVectors/zero-applicable | Killed on all five lanes, child exit 1; no current survivor |
| changed-root-admitted | Keep root verification but admit only Muse's root with the forged bytes `changed` plus LF | TestMuseRootContextDrift/changed | Killed on all five lanes, child exit 1; no current survivor |
| removed-root-admitted | Keep verification but admit only absent Muse config/muse/AGENTS.md | TestMuseRootContextDrift/removed | Killed on all five lanes, child exit 1; no current survivor |

The Go overlay harness TestMuseRootContextNarrowingMutants performs these mutations without editing the candidate. This evidence is behavioral, including negative output, refusal, and drift cases; it does not attest to external launcher argv.

## Diagnostic bounds and portability repairs

Attempt 04 killed all seven mutants on Linux. On macOS all seven had no named behavioral failure because child TestMain waited on the toolchain lock already held by its parent: those seven are survivors/unrun, bounded to Linux proof only. On Windows five were killed and the two drift mutants had no named failing test because CRLF source anchors did not match: both are survivors/unrun, leaving Windows drift mutation evidence unestablished on that snapshot. The harness correctly rejected these unrelated failures. Reused the existing cliHelperProcess child protocol, preserving git isolation, and normalized overlay source line endings; no product behavior or verification gate was weakened. Attempt 06 reruns the full suite for the changed test identity.

Historical attempt-04 mutation evidence (not substituted for the corrected candidate):

| Mutant | Narrowed behavior | Linux named failure (exit 1) | macOS survivor bound | Windows named failure or survivor bound |
| --- | --- | --- | --- | --- |
| configured-form-ignored | Ignore exactly configured Muse referenced | TestMuseRootContextVectors/referenced-refused | No named failure; nested lock timeout, macOS mutation bound unestablished | TestMuseRootContextVectors/referenced-refused (exit 1) |
| referenced-fallback | Admit exactly Muse referenced | TestMuseRootContextVectors/referenced-refused | No named failure; nested lock timeout, macOS mutation bound unestablished | TestMuseRootContextVectors/referenced-refused (exit 1) |
| no-context-admitted | Admit Muse with no context | TestMuseRootContextVectors/no-context | No named failure; nested lock timeout, macOS mutation bound unestablished | TestMuseRootContextVectors/no-context (exit 1) |
| wrong-target | Use only Muse home-level target | TestMuseRootContextVectors/monolithic | No named failure; nested lock timeout, macOS mutation bound unestablished | TestMuseRootContextVectors/monolithic (exit 1) |
| wrong-bytes | Append bytes only to Muse header | TestMuseRootContextVectors/zero-applicable | No named failure; nested lock timeout, macOS mutation bound unestablished | TestMuseRootContextVectors/zero-applicable (exit 1) |
| changed-root-admitted | Admit only Muse root containing changed plus LF | TestMuseRootContextDrift/changed | No named failure; nested lock timeout, macOS mutation bound unestablished | No named failure; CRLF anchor miss, Windows mutation bound unestablished |
| removed-root-admitted | Admit only absent Muse root | TestMuseRootContextDrift/removed | No named failure; nested lock timeout, macOS mutation bound unestablished | No named failure; CRLF anchor miss, Windows mutation bound unestablished |
