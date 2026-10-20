# Muse 1.4.10-R6427.1 loader qualification record

Task: TASK-261018-ujp9nh, rc.7. This is a vendor observation record,
not runtime admission authority. **Exec: qualified for the measured loader
rows; environments §7.9 exclusion replay passes. Interactive: refused;
P4 skill exclusion FAILS and the §7.9 release gate remains pending. The
1.4.8 interactive failure persists in 1.4.10.** No runtime constants,
registry pins, qualification gates or credential policy change in this leaf.

## Exact identity and observation boundary

| Identity | Value |
|---|---|
| Vendor version | `Muse Code 1.4.10 (1.4.10-R6427.1)` |
| Binary SHA-256 | `11cc1ae887b4975bbbe1de7358ee9260a612444df9ce4dbfffa1dd785826dbcf` |
| Binary locator | `$MUSE1410_BINARY`, task-board Muse pin store; absolute locator only in ignored task-local `identity.json` |
| Platform | macOS 27.0 arm64 |
| Curator checkpoint | `e88fdf2dbec7f94c48b0232a350a0a9f9484fdeb` |
| Normative curator-spec main | `e41f96387287f5e76604627f9e63662f7d6d86d4` |
| Harness | Go 1.25.5, tmux 3.6a, Apple sandbox-exec |
| Native root SHA-256 | `8c9af1ca52b75aaa1e67af6365f1e39dfcc3c56a274d3571270079ef669ac9d0` |
| Snapshot manifest SHA-256 | `924378006ba6e6e7db9d0277375456a958c14cbaff03d1ebd63dfa6e4b7ceac8` |
| Snapshot launch SHA-256 | `2991bc9c97904d77aef70bb5c3ea7a7044ff462382310ae613bbd31f461fd5f1` |

The exact binary is a read-only manager pin, copied into task-local `.temp/`
and independently digest-checked without modifying the pin store. The
updating launcher is never invoked. `TestVendorIntakeBoundary` drives a real
vendor-test entry in child processes: missing opt-in skips, wrong bytes
refuse, and neither creates a vendor fixture. A digest-gated, sandboxed
`TestMuse1410Identity` verifies the version with exit 0.

Normative sources reread on fetched spec main: environments §§7.1 P1–P4,
7.9, 10.2; CIP-0017 (implementation leaf 2 and minimal probes); and
`conformance/environments-muse-v4/README.md`. The advertised main OID and
fetched OID agree. These sources still separate credential admission from
exclusion qualification. Neither 1.4.7 nor 1.4.8 evidence is promoted here.

Every vendor invocation uses `MUSE_NO_AUTO_UPDATE=1`, `env -i`, disposable
HOME and four XDG parents under task-local `.temp/`,
`TBH_CREDENTIAL_BACKEND=file`, `--provider echo`, and sandbox-exec. The
sandbox denies all network, security-service lookups, Keychain file access
and real native config/library directories. The binary copy is outside
those denied native directories. No login, logout or auth command runs;
no real Keychain item is inspected and no remote inference uses shared quota.

`TestMuse1410OfflineBoundary` pairs reachable loopback (curl exit 0, exactly
one server hit) with the same endpoint denied by the sandbox (exit 7, no
additional hit). Ordinary synthetic bytes are readable (cat exit 0), while
a synthetic Keychains file is denied (exit 1, Operation not permitted),
with unchanged bytes. Interactive sends a local echo turn to flush loader
records, then `/rules` and `/exit` in a private background tmux terminal.
This is loader evidence before remote inference, not model-visible removal,
credential-backend/login/refresh qualification or proof of no local turn.

## Fixtures, commands and production entrypoints

[Harness and fixtures](../../.research/muse1410/README.md) mirror 1.4.8,
including its reviewer regressions. Unchanged fixture bytes retain the
synthetic `muse148-probe` name, frozen root and metadata digests. The snapshot
manifest's historical Curator version describes that reused fixture; it is
not a new export-provenance claim. Current Curator runs real `profile install
SOURCE`, then `env resolve muse --profile muse148-probe --repair --format
json`. The resulting real symlink must match the frozen root. A synthetic
`{}` auth target serves only the resolver's metadata check; Muse never runs
in that fixture home.

The independent published `managed_home_snapshots.check_bundle` checks the
snapshot and complete physical closure through the retained reference wrapper.
Private runtime copies are regular 0600 files with all four declared XDG
suffixes. Source mode bits alone are waived because Git cannot retain 0600.
Listed symlinks and unlisted dangling links are refused before collecting
bytes. This covers frozen representation and loading, not materialize
provenance, publication or hostile concurrent filesystem mutation.

```text
exec:        $MUSE1410_BINARY exec --json --provider echo --trust-workspace [--no-foreign-personal-context] MUSE1410_OFFLINE_CONTROL
interactive: $MUSE1410_BINARY --provider echo --trust-workspace [-c context.foreign_personal_skills=false]
```

P1–P4 observations below are driven by `TestMuse1410LoaderMatrix` (36 cases),
`TestMuse1410ParserRefusals` (2), `TestMuse1410ExecInteractiveKeyObservation`
(1), `TestMuse1410InteractiveExclusionRegression` (2), and
`TestMuse1410WorkspaceResidual` (4): **45 loader/parser entry cases** plus
one isolated version inspection. Successful vendor entries exit 0; both
interactive parser refusals exit 2. The gated suite's green status pins the
failure witness; it does not qualify interactive exclusion.

| Probe | Exec observation and vendor exit | Interactive observation and vendor exit |
|---|---|---|
| P1: exec exclusion spelling | Accepted; root-absent foreign rule count 1 → 0 and skill count 1 → 0; native-root skills 1 → 0; exit 0 for both controls and exclusion | The exec-only spelling is covered by P2 |
| P2: form/parser boundary | Contrary to the brief's anticipated refusal, `-c context.foreign_personal_skills=false` is accepted, exit 0, and reaches the loader. This parser-only case makes no claim about that spelling's exec exclusion semantics | Exec-only flag rejected, exit 2, `invalid TUI options`; string `"false"` independently rejected, exit 2, `the JSON type does not fit this setting`; neither reaches rules loader. Boolean accepted in matrix, exit 0 |
| P3: foreign personal rules fallback | Without root: separate Claude/Codex rules each give `user_sources=1 foreign_sources=1`; with linked or snapshot root: `user_sources=1 foreign_sources=0`; exit 0. Explicit exclusion also removes root-absent fallback | Same separate fallback/root contrast, exit 0. Boolean without root leaves fallback at 1; native root suppresses fallback with or without boolean |
| P4: foreign skills persist and required exclusion | Both root representations retain one skill without exclusion; explicit flag yields `user_skills=0`, preserves native root, exit 0 | Both root representations retain one skill with and without boolean (`user_skills=1`), exit 0; retained-skill banner confirms it. No-root boolean also leaves one skill. **Required exclusion fails**, reproducing 1.4.8 |

Workspace residual: both forms load native and workspace rules, vendor exit
0 for both representations. Interactive `/rules` lists user then project;
project count goes from 1 at startup to 0 on the subsequent reload. This is
enumeration only; model-visible precedence/conflict resolution remains unknown.
Source bytes and native root digests stay unchanged, and credential files
remain absent in every vendor home.

[Sanitized observations](../../.research/muse1410/observations.json) preserve
per-case argv, vendor exits, loader records and bounded TUI evidence. Raw
paths remain private. [Validation and controls](../../.research/muse1410/TASK-261018-ujp9nh_gaps.md)
record commands, real exit codes, source identity and narrowing results.

## Coverage map

**6 of 6 task AC rows driven**, with loader qualification/refusal reported
separately from successful evidence collection. No catalog-keyed surface table
was supplied: this is a brief gap. There are no unreported in-scope rows.

| AC row | Named tests / production call site | Result and narrowing evidence |
|---|---|---|
| 1: P1–P4 both forms | `TestMuse1410LoaderMatrix`, `TestMuse1410ParserRefusals`, `TestMuse1410ExecInteractiveKeyObservation`; real vendor exec/TUI parser and rules/skills loaders | All rows observed with exits; exec omission plant fails `/exec/linked/skills/exclude-true`; interactive omission survives and establishes no protection |
| 2: §7.9 and 1.4.8 comparison | `TestMuse1410InteractiveExclusionRegression`; real TUI loader and retained-skill banner | Exec replay passes; interactive refused on links and snapshots, prior failure persists |
| 3: digest and missing opt-in | `TestVendorIntakeBoundary`, `TestMuse1410Identity`; real child vendor-test entry → `setup` → sandboxed vendor version | Wrong digest rejected and unset gate skips before fixture; single synthetic-digest allowance killed at `/wrong-digest` |
| 4: offline boundary | `TestMuse1410OfflineBoundary`; sandbox-exec → curl/cat; matrix → sandbox-exec → env -i → vendor echo | Positive/denied pairs pass; localhost-only allowance killed; no auth/remote command is present in executed argv |
| 5: planted foreign-skill violation | `TestMuse1410LoaderMatrix/exec/linked/skills`, `TestObservationChecker`, retained reviewer test; real vendor loader → `checkCounts` | Exec omission and one-skill allowance killed; contradictory-field mutants killed |
| 6: build/default tests/hygiene | Recorded standalone Go commands and lint; `TestFrozenSnapshot`, `TestFrozenSnapshotSourceInputs`, `TestReferenceSnapshotPhysicalClosure`; current Curator install/resolve and published snapshot checker | Snapshot entry/link mutants killed; scoped hygiene check verifies only allowed files change. Default tests and gated evidence are separately reported |

Out-of-contract rows: credentials/login/refresh and model-visible removal
(CIP-0017 minimal probes 3–5 are separate credential/broker evidence; task
scope is P1–P4 loader replay); model-visible workspace precedence (brief
workspace-residual clause); capture locks/export provenance/publication/
destination admission/runtime-B/materialize integration (reused snapshot
fixture boundary); serve, launcher/runtime consumer gates, registry constants
and credential policy (task scope explicitly excludes their changes).
Parser acceptance of the interactive key on exec is measured; its exec
semantics under foreign-context inputs is a stated unmeasured bound.

## Validation and release effect

Commands, exits, log links and mutant table are in the
[task gaps record](../../.research/muse1410/TASK-261018-ujp9nh_gaps.md).
The interactive omission is a non-discriminating survivor because valid
vendor exclusion already retains the skill. The checker still rejects that
skill when asked to attest zero; no gate is weakened to accept the build.

Exec is qualified only for the measured macOS arm64 loader rows. Consumer
owners may use the exact-build exec evidence under their own admission
contract. Interactive remains refused: keep its environments §7.9 release
gate pending. D10 credential-mode admission is independent. A corrected build
needs its own exact identity and replay; 1.4.11 is the next build to qualify.
No platform-independent isolation, model-visible removal or refresh claim
follows from this record.
