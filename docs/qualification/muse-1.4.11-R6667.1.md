# Muse 1.4.11-R6667.1 loader qualification record

Task: TASK-261021-mw4laf, rc.7. This is a vendor observation record,
not runtime admission authority. **Exec loader rows qualify; interactive
P1–P4 skill exclusion FAILS. The environments §7.9 exclusion replay passes
for exec and remains pending for interactive, unchanged from 1.4.10.**
Consumer owners retain admission authority. No runtime constants, registry
pins, qualification gates or credential policy change in this leaf.

## Exact identity and observation boundary

| Identity | Value |
|---|---|
| Vendor version | `Muse Code 1.4.11 (1.4.11-R6667.1)` |
| Binary SHA-256 | `6c04ce1719844685153ef6f04b98113cb61b942d97f0a8b0bfb3932c343984d9` |
| Binary locator | `$MUSE1411_BINARY`; pin-store source and private copy locators recorded only in ignored task-local `identity.json` |
| Platform | macOS 27.0 arm64 |
| Curator checkpoint | `cdcba3bd7a8de6028d3d1dbf621ce4cb19b130a7`, zero commits behind local main when measured |
| Normative curator-spec main | `e41f96387287f5e76604627f9e63662f7d6d86d4` |
| Harness | Go 1.25.5, tmux 3.6a, Apple sandbox-exec |
| Native root SHA-256 | `8c9af1ca52b75aaa1e67af6365f1e39dfcc3c56a274d3571270079ef669ac9d0` |
| Snapshot manifest SHA-256 | `924378006ba6e6e7db9d0277375456a958c14cbaff03d1ebd63dfa6e4b7ceac8` |
| Snapshot launch SHA-256 | `2991bc9c97904d77aef70bb5c3ea7a7044ff462382310ae613bbd31f461fd5f1` |

The executable was copied read-only from the task-board Muse pin store;
both source and task-local copy matched the pinned digest. The manager's
store and updating launcher were untouched. The harness refuses different
bytes. Without `MUSE1411_BINARY`, every vendor test skips before creating
a fixture. Ordinary `go test ./...` excludes this hidden evidence package;
the explicit ungated package run checks the skips and non-vendor controls.

Normative sources reread from a snapshot of fetched current spec main:
environments §§7.1 P1–P4, 7.9, 10.2; CIP-0017, especially implementation
leaf 2 and minimal probes 1–5; and
`conformance/environments-muse-v4/README.md`. Earlier 1.4.7 context and
1.4.8 credential evidence is not promoted to this build.

Every vendor invocation uses `MUSE_NO_AUTO_UPDATE=1`, `env -i`, disposable
HOME and all four XDG parents under task-local `.temp/`,
`TBH_CREDENTIAL_BACKEND=file`, and `--provider echo`. The OS sandbox denies
all network, security-service lookups, Keychain file access and the real
native configuration/library directories. The paired offline controls
observe curl exits 0 / 7 with exactly one server hit, and ordinary versus
synthetic Keychains file read exits 0 / 1, unchanged bytes and
`Operation not permitted`. No real Keychain item, login, logout or auth
command is used. No remote inference runs or shared inference quota is used.

Interactive runs in a private background tmux terminal. Startup, a local
echo sentinel, `/rules`, then `/exit` flush the loader diagnostics without
using the active UI. This measures pre-remote-inference loading, not
model-visible foreign-context removal. Credential files remain absent in
vendor homes; fixture sources and native root bytes remain unchanged.

## Fixtures, commands and production entrypoints

[Harness and fixtures](../../.research/muse1411/README.md) mirror 1.4.10
and the 1.4.8 template, including the original reviewer regressions.
Unchanged `testdata` bytes retain the synthetic `muse148-probe` name and
frozen digests. Historical snapshot metadata describes the reused fixture,
not a new export. Current Curator runs real `profile install SOURCE`, then
`env resolve muse --profile muse148-probe --repair --format json`; the
result must be a real symlink matching the frozen root. Synthetic `{}` auth
bytes serve only the resolver's metadata check. Muse never runs in that home.

The retained reference wrapper calls the published
`managed_home_snapshots.check_bundle` after classifying complete physical
closure without following links. Listed links and unlisted dangling links
are refused before collecting bytes. Private snapshot copies are regular
0600 files with all four declared XDG suffixes; only source mode bits are
waived because Git cannot retain 0600. This measures representation and
loading, not export provenance or hostile concurrent filesystem mutation.

```text
exec:        $MUSE1411_BINARY exec --json --provider echo --trust-workspace [--no-foreign-personal-context] MUSE1411_OFFLINE_CONTROL
interactive: $MUSE1411_BINARY --provider echo --trust-workspace [-c context.foreign_personal_skills=false]
```

`TestMuse1411LoaderMatrix` drives 36 cases, `TestMuse1411ParserRefusals` 2,
`TestMuse1411ExecInteractiveKeyObservation` 1,
`TestMuse1411InteractiveExclusionRegression` 2, and
`TestMuse1411WorkspaceResidual` 4: **45 loader/parser cases**, plus one
isolated version inspection. There are 43 vendor exits 0 and two parser
exits 2. The green gated suite pins the interactive failure witness;
it does not qualify interactive exclusion.

| Probe | Exec observation and vendor exit | Interactive observation and vendor exit | Compared with 1.4.10 |
|---|---|---|---|
| P1: exec exclusion spelling | Accepted; root-absent foreign rules 1 → 0 and skills 1 → 0; native-root skills 1 → 0; all exit 0 | Exec-only spelling covered by P2 | Same |
| P2: form/parser boundary | `-c context.foreign_personal_skills=false` accepted and reaches loader, exit 0; parser-only case, exclusion semantics unmeasured | Exec-only flag refused, exit 2, `invalid TUI options`; string `"false"` independently refused, exit 2, `the JSON type does not fit this setting`; neither reaches rules loader. Boolean accepted, exit 0 | Same; anticipated exec-key refusal is still not observed |
| P3: foreign rules fallback | Separate Claude/Codex rules: absent root gives `user_sources=1 foreign_sources=1`; linked/snapshot root gives `user_sources=1 foreign_sources=0`. Exec exclusion removes absent-root fallback; all exit 0 | Same root contrast; no-root boolean retains fallback at 1. Native root suppresses fallback with or without boolean; all exit 0 | Same |
| P4: persistent skills and required exclusion | Linked/snapshot native roots retain one skill without exclusion and zero with the flag, native root preserved; all exit 0 | Both root representations retain one skill with or without boolean, including retained-skill banner. No-root boolean also retains one; all exit 0. **Required exclusion fails** | Same; 1.4.8 and 1.4.10 interactive failure persists |

Workspace residual: both forms load native and workspace rules, exit 0 for
both root representations. Interactive `/rules` enumerates both; project
count changes from 1 at startup to 0 on subsequent reload. This establishes
enumeration only; model-visible precedence/conflict resolution is unknown.
No loader-relevant flag/config behavior changed in the exercised interface,
including interactive `context.foreign_personal_skills`. Additive MSP
session/leave, inbox notification and recoveryMode behavior is outside this
loader replay and was not qualified.

[Sanitized observations](../../.research/muse1411/observations.json) bind
each case's argv, vendor exit, loader records and bounded visible markers to
source digests. The [row comparison](../../.research/muse1411/comparison-1.4.10.json)
finds **45 of 45 cases unchanged** for argv (sentinel normalized), exits,
ordered loader event/outcome/count fields and visible marker sets. Timings
and sentinel-dependent rendered byte lengths are excluded from that comparison.

## Coverage map

**6 of 6 task AC rows driven**, with qualification/refusal separate from
successful evidence collection. No catalog-keyed surface table was supplied:
this is a brief gap. There are no undeclared in-scope rows.

| AC row | Named tests / production call site | Result and narrowing evidence |
|---|---|---|
| 1: P1–P4 both forms | `TestMuse1411LoaderMatrix`, `TestMuse1411ParserRefusals`, `TestMuse1411ExecInteractiveKeyObservation`; real vendor exec/TUI parsers and rules/skills loaders | All observations have vendor exits; omission plant fails `/exec/linked/skills/exclude-true` |
| 2: §7.9 and 1.4.10 comparison | `TestMuse1411InteractiveExclusionRegression`; real TUI loader/banner; comparison of both sanitized case sets | Exec qualifies, interactive refused on links and snapshots; 45 cases unchanged |
| 3: pin and missing opt-in | `TestVendorIntakeBoundary`, `TestMuse1411Identity`; vendor-test entry → `setup` → sandboxed version | Wrong digest refused and unset gate skips before fixture; single foreign-digest allowance killed at `/wrong-digest` |
| 4: offline boundary | `TestMuse1411OfflineBoundary`; sandbox-exec → curl/cat; matrix → sandbox-exec → env -i → vendor echo | Paired controls pass; localhost-only and one-synthetic-Keychains-file allowances killed; executed argv contains no auth/remote command |
| 5: planted foreign skill | `TestMuse1411LoaderMatrix/exec/linked/skills`, `TestObservationChecker`, `TestReviewContradictoryObservation`; real loader → `checkCounts` | Exec omission and checker one-skill allowance killed; contradictory observations refused |
| 6: build/default tests/hygiene | Recorded standalone Go/lint commands; `TestFrozenSnapshot`, `TestFrozenSnapshotSourceInputs`, `TestReferenceSnapshotPhysicalClosure`; real Curator install/resolve and published snapshot checker | Local build/vet and opt-in suite green; snapshot mutants killed. Exact green default-suite evidence reused on unchanged inputs; scope/public hygiene checked |

Out-of-contract rows: credentials/login/refresh, broker keys and
model-visible removal (task's offline P1–P4 scope; CIP-0017 minimal probes
3–5 need separate evidence); model-visible workspace precedence (residual
observation only); capture locks, export provenance, publication,
destination admission, runtime-B and materialize integration (reused frozen
snapshot boundary); serve, launcher/runtime consumer gates, registry pins
and credential policy (explicit scope exclusions); native platforms other
than macOS arm64 (exact vendor identity). Accepted exec config-key spelling
is parser evidence only; its exclusion semantics remain unmeasured.

## Validation and release effect

Commands, actual exits, evidence identities, scope and the full mutant table
are in the [task gaps record](../../.research/muse1411/TASK-261021-mw4laf_gaps.md).
All ten patch mutants and the exec omission plant fail at named assertions.
Interactive omission survives because the real exclusion already retains
the skill; it establishes no interactive protection. The observation checker
still refuses that skill when asked to attest zero.

The new harness's full gated run, ungated controls, vet and lint were run
locally. Default source/test/config/workflow inputs match the previously
green hosted gate exactly; its complete original test stream was retrieved
again and its digest verified. This is evidence reuse, not a local full-suite
pass or an external-CI fallback. Reviewer regressions retain their names.

Exec is qualified only for these measured macOS arm64 loader rows.
Interactive remains refused: its §7.9 release gate stays pending, as on
1.4.10. Consumer owners may consume the exact-build exec record under their
own admission contract. D10 credential-mode admission is independent.
No runtime gate or pin is changed here, and no portable isolation,
model-visible removal, authentication or refresh claim follows.
