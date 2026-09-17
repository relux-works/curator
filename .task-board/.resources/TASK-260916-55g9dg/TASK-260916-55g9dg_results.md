# TASK-260916-55g9dg results — direct-only `class: system` modules (E2, manager)

Story STORY-260916-2d9coh, wave 1. Spec: curator-spec `0da4020`,
`protocol/environments.md` §3/§5.5/§12 (+ `profiles/manager.md` §1 lock set).
Worktree branch `task-board/story/STORY-260916-2d9coh`, uncommitted.

Outcome: implemented and verified. The E2 admission rule is enforced at
resolution (error policy, pre-publish, lock untouched) and at
materialization (drop policy, admitted bytes only), with the two §12.1
knobs, the §12.2 lock direction, status posture, vectors, and CHANGELOG.

## 1. Per-file changes

Production:

- `internal/config/environments.go` — `Environments` gains
  `TransitiveSystemModules` (closed enum `drop`|`error`, default `drop`)
  and `SystemModuleWaivers` (`[{package, reason}]`, default empty);
  `parseSystemModuleWaivers` enforces the package identifier grammar,
  non-empty reason, and no unknown fields; `parseSystemEnvironments`
  refuses a system file carrying `drop` (locks to `error` only);
  `LockableEnvKeys` + `EnvLockKey` admit `transitive_system_modules`
  only (waivers not lockable); `render` writes both knobs.
- `internal/config/config.go` — `LockableKeys` gains
  `environments.transitive_system_modules` (waivers deliberately absent).
- `internal/contextmaterialize/admission.go` (new) — admission core:
  diagnostics `context_system_module_dropped` /
  `context_system_module_transitive`, `Admission` policy, `DirectSet`
  (root + active overlays + `required_by`-named-by-root-or-overlay),
  `ClassifySystemModules` (emitted × manifest order),
  `FirstTransitiveSystemModule` (resolution scan, applies-to-≥1-adapter),
  `TransitiveSystemModuleError` naming package + module.
- `internal/contextmaterialize/contextmaterialize.go` — `SystemPrompt`
  takes `Admission` and returns `(document, written, dropped, err)`:
  under drop the document is exactly the admitted modules' bytes;
  under error the first dropped module refuses with
  `context_system_module_transitive` and no document.
- `internal/envprofile/envprofile.go` — `Policy` gains
  `TransitiveSystemModules` (+ `SystemModuleWaiver`, `Policy.Admission()`,
  `PolicyFromConfig` mapping); `checkAdmissionPrePublish` gates install,
  reinstall, and update before any lock publish and before the
  identical-lock fast path (a violating lock fails the update even when
  nothing moved); unreadable manifest fails closed (`§8.4`).
- `internal/envprofile/managed.go` — `systemPrompt` passes the policy
  and appends one `context_system_module_dropped` warning naming package
  + path per dropped module (flows to `ResolveResult.Warnings` through
  re-verification; under error the repair fails with the refusal).
- `internal/envprofile/status.go` — `ProfileState` gains
  `transitive_system_modules` (effective value) and
  `dropped_system_modules` (package+path, emitted order, drop only;
  empty under error); read-only recompute from lock + store manifests.
- `cmd/curator/envstatus.go` — profile row prints the policy value and
  one `dropped system module <package> <path>` line per drop.
- `CHANGELOG.md` — Unreleased E2 entry (default `drop`, `error` opt-in).

Tests:

- `internal/config/environments_test.go` — lockable-subset transcription
  + payloads for both knobs, `EffectiveJSONShape` knob list, new
  `TestTransitiveSystemModulesDefaults/Values`,
  `TestSystemTransitiveDirection` (error locks, drop refused locked and
  unlocked, waivers refused carried/not-lockable).
- `internal/config/environments_conformance_test.go` —
  `prunePostRevisionKnobs`: a vector revision predating a knob cannot
  expect it, so exactly the closed `postPinKnobs` list is pruned from
  the rendered object when the case's own expectation lacks it. Full
  exact comparison at the new root; old keys still exact at the pin.
- `internal/contextmaterialize/admission_test.go` (new) — direct-set
  partition (root/overlay/requires, overlay-required, root-shared),
  drop bytes + dropped, per-module drops, error refusal naming the
  first module with no document, waiver admission under both policies,
  selector exclusion, all-dropped unwritten, policy-enum rejection,
  resolution-scan order + unregistered-selector exclusion.
- `internal/envprofile/admission_test.go` (new) — production entry
  points over a git root→mid→leaf closure: error-policy install
  refuses with the typed diagnostic and writes no lock; waiver
  admits; error-policy update fails and the old lock bytes are
  identical; drop repair warns, writes exactly the admitted bytes,
  and the fragment carries `system_prompt`; all-dropped writes no
  file and no fragment section; status reports policy + drops with
  the home row current; error status reports policy with empty
  drops; `PolicyFromConfig` mapping + zero-policy drop default.
- `internal/interop/environments/context_materialization_test.go` —
  vector struct gains `machine_policy`/`admitted`/`dropped`/`warnings`/
  `error*`; system-prompt cases drive `SystemPrompt` with the case
  policy and assert bytes + admitted + dropped + warnings, or the
  refusal naming diagnostic + package + module.

## 2. AC mapping (acceptance criteria → code)

- Knobs parsed, validated, written; schema cases from the root:
  `internal/config/environments.go:40,43,99,769,912,1027`,
  `internal/config/config.go:65`. Schema-case tests consume the root
  index dynamically (`environments_conformance_test.go`), so the new
  cases execute at `0da4020` and are absent (not failed) at the pin.
- Direct = root / active overlay / requires-named; waived admitted:
  `internal/contextmaterialize/admission.go:72` (`DirectSet`),
  `:62` (`Waived`), `:135` (`ClassifySystemModules`).
- Drop skips with `context_system_module_dropped` naming package +
  module; bytes = admitted only:
  `internal/contextmaterialize/contextmaterialize.go:257`,
  `internal/envprofile/managed.go:1787`.
- Error fails resolution with `context_system_module_transitive`,
  lock unchanged: `internal/envprofile/envprofile.go:1209`
  (`checkAdmissionPrePublish`, called pre-publish at install,
  reinstall, and update — including before the identical-lock fast
  path — plus `FirstTransitiveSystemModule` at
  `internal/contextmaterialize/admission.go:160`).
- `context-system-module-present` stays always-warn:
  `internal/envprofile/envprofile.go` audit warnings untouched
  (proven by `TestDropRepairWarnsAndFragmentFollowsAdmitted`).
- Fragment follows the admitted set:
  `internal/envfragment/envfragment.go` `buildFragment` keys
  `system_prompt` off the marker surface, which the admitted-only
  materialization writes; `works.relux.curator.system-modules` is
  the launcher-owned ax key reflecting that presence (no manager
  field of that name exists; manager contract §10.2 presence is
  pinned by the fragment tests).
- Status reports policy + every dropped module; drop never
  non-current: `internal/envprofile/status.go:110,114,390,404`,
  `cmd/curator/envstatus.go:72`; currency is findings-only
  (`status.go` `homeState`), proven by
  `TestStatusReportsPolicyAndDropped`.
- Vectors + unit tests + CHANGELOG + transcripts: §4, §5.

## 3. Profile shipped

`drop` default (non-breaking), `error` opt-in — exactly as §12.1
states. No warn-first split applies (brief: direct, non-breaking).

## 4. Validation transcripts (real exit codes, shell `bash`)

Conformance root for the new-revision runs:
`CURATOR_CONFORMANCE_ROOT=…/curator-spec/conformance/v1` (`0da4020`).
Pin root: `/tmp/pinroot-87a0d00/conformance/v1`, extracted read-only
via `git archive`/`git show` from `87a0d00` (the committed SPEC_PIN).

- `go build ./...` → exit 0.
- `go vet` on `internal/config internal/contextresolve
  internal/contextmaterialize internal/contextaudit
  internal/envfragment internal/envprofile cmd/curator
  internal/interop/environments` → exit 0.
- `gofmt -l internal/ cmd/` → empty (clean). (`gofmt -l .` lists
  only pre-existing `.task-board/.resources` probe files, untouched.)
- `golangci-lint run` (v2.12.2, repo `.golangci.yml`) on all eight
  narrow packages → `0 issues.`

New-root (`0da4020`) suites, each run directly (no `tee`; exit code
via `PIPESTATUS`):

- `go test ./internal/interop/environments/` → exit 0 (`ok`, 2.0s):
  all materialization cases byte-exact, including the five
  admission cases `system-module-direct`,
  `system-module-transitive-drop`, `system-module-transitive-error`
  (refusal naming diagnostic + package + module, `file_written`
  false), `system-module-transitive-waived`,
  `system-module-overlay-direct`.
- `go test ./internal/contextmaterialize/ -count=1` → exit 0.
- `go test ./internal/contextresolve/ ./internal/contextaudit/
  ./internal/envfragment/` → exit 0 each.
- `go test ./internal/envprofile/` in three `-run` shards + one
  leftover (partition verified to cover all 154 tests; the full
  package exceeds the single-call time bound, so sharding is the
  documented split, not a narrowing): shard A (install/update/
  reinstall/drop/status/policy/…) → exit 0, 504.6s; shard B
  (import/resolve/repair/managed/…) → exit 0, 180.8s; shard C
  (path/overlay/weight/git/…) → exit 0, 198.5s;
  `TestNonDirectoryPathIsSourceInvalid` → PASS.
- `go test ./cmd/curator/` in four `-run` shards (partition
  verified to cover all 163 tests): env/profile/import →
  exit 0, 267.3s; build/status/classify → exit 0, 75.7s; gc/
  global/cli/misc → exit 0, 194.6s; compiled/toolchain/creds
  leftover → exit 0, 342.6s.
- `go test ./internal/config/ -count=1` → exit 1 (expected-red:
  41 subtests fail, and every one carries E4's
  `provider_directories` knob — verified programmatically: zero
  failing subtests without it. All E2-invalid schema cases pass,
  all new unit tests pass; see §6). Failing set: every
  `valid*` schema case and every valid manager vector
  (`valid vector rejected: … unsupported field
  "provider_directories"` or the expected member carrying it).

Pin-root (`87a0d00`, the committed SPEC_PIN lane) suites:

- `go test ./internal/interop/environments/` → exit 0 (old cases
  byte-exact under default drop; admission cases absent, nothing
  skipped-or-failed).
- `go test ./internal/config/` → exit 0 (old schema cases +
  pruned vectors + new unit tests). The pin lane stays green.

Negative evidence (narrowing mutants, each restored after):

- `DirectSet` without the overlay-requirer rule → unit
  `TestDirectSet` FAILs and vector
  `system-module-overlay-direct` FAILs (admitted list short).
- `Admission.Waived` forced false → unit
  `TestSystemPromptWaiverAdmits` FAILs and vector
  `system-module-transitive-waived` FAILs.
- The error-policy install/update tests assert the typed
  `TransitiveSystemModuleError` (`errors.As`) plus lock
  absence/byte-identity; the drop tests assert exact file bytes.

Manual CLI proof (built binary, temp config): `env config show
transitive_system_modules` → `"drop"` default; `set … '"error"'`
→ persists; `set … '"quarantine"'` → exit 1
(`environments.transitive_system_modules: must be drop or
error`); waiver list round-trips.

## 5. Deliberately out of scope

- E4 `provider_directories` (knob, trust, dispatch): a sibling
  task owns it. Parsing it here without E4 semantics would be a
  hollow stub, so the new-root `valid*` config cases stay red
  until E4 lands (§4/§6 evidence).
- E1 signer rules; pi `SYSTEM.md` channel changes; SPEC_PIN bump;
  ax integration; proposals 0014–0018; tags/releases.
- `internal/contextresolve.Resolve` itself is untouched: module
  manifests are not resolver inputs, so the §3 error refusal is
  enforced pre-publish in `envprofile` over the identical closure
  + `required_by` edges (equivalent directness: one name, one
  kind), and at materialization for stale locks. Observable
  behavior is exactly "resolution fails, lock unchanged".
- No `platform-cases.tsv` / `root-artifacts.tsv` change: no new
  vector file or schema family was consumed — only new cases
  inside already-declared families, which the dynamic tests pick
  up automatically (no new skip class needed).

## 6. Spec gaps and findings (reported, not patched)

- F1 (resolution-time "applicable" has no environment): §3 says
  "resolution of the same module MUST fail" where "the same
  module" is a non-admitted *applicable* system module, but
  applicability is environment-relative. Implemented: a module
  refusing resolution iff it applies to ≥1 registered adapter
  environment; a module selecting only unregistered environments
  selects nothing (§3) and never refuses. Unit-pinned
  (`TestFirstTransitiveSystemModuleOrder`).
- F2 (materialization under error for a stale violating lock,
  e.g. after a drop→error flip): §3 only defines resolution
  failure. Implemented fail-closed: `SystemPrompt` returns
  `context_system_module_transitive` (the error vector pins
  `file_written: false` at the materialization layer), which
  also preserves "MUST contain only admitted system modules".
  The fix is machine-config-only (add a waiver), no
  re-resolution needed.
- F3 (E2 schema/vector cases bundle E4's knob): every E2 schema
  case and every valid manager vector also carries
  `provider_directories`, so the invalid E2 cases currently
  reject on the E4 field before reaching the E2 checks (the
  dedicated unit tests pin the E2 checks directly), and the
  valid E2 cases cannot go green until E4 lands. Pre-existing
  overlap, no spec change proposed.
- F4 (update identical-lock fast path vs refusal): §9.2 says an
  identical-lock update "changes nothing and says so"; under
  error with a violating lock this would hide the violation
  forever. Implemented: the refusal runs before the fast path,
  so the update fails and the old lock stands. Documented in
  `checkAdmissionPrePublish`.

