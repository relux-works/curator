# Muse 1.4.13-R6996.1 loader qualification record

Task: TASK-261026-jk47fn. This is a vendor observation record,
not runtime admission authority. **Exec loader rows qualify; interactive
P1–P4 skill exclusion FAILS. The environments §7.9 exclusion replay passes
for exec and remains pending for interactive, unchanged from 1.4.11.**
Consumer owners retain admission authority. No runtime constants, registry
pins, qualification gates or credential policy change in this leaf.

**1.4.12-R6909.1 is deliberately skipped:** it is superseded on every known
host according to the task intake. No 1.4.12 executable was inspected or run.
The previous qualified observation baseline is 1.4.11-R6667.1 (PR #152).
Task-board canary exec/TUI green and its reported additive MSP schema change
are intake context, not proof of exclusion or a substitute for this replay.

## Exact identity and observation boundary

| Identity | Value |
|---|---|
| Vendor version | `Muse Code 1.4.13 (1.4.13-R6996.1)` |
| Binary SHA-256 | `9fd40ae9c6dfb847e3ef58b61405c004d0ebbf19ceb0297be966f750b3b84aa0` |
| Binary locator | `$MUSE1413_BINARY`; installed source and private copy locators recorded only in ignored task-local `identity.json` |
| Platform | macOS 27.0.1 arm64 |
| Curator checkpoint | `c7b39b5474f68db1064256a91e59a9cdfb66b49b`, fetched main and local main equal before branching |
| Normative curator-spec main | `88b0d8615a5df8728e0e9540a9edd708c0ca5660` |
| Harness | Go 1.25.5, tmux 3.8, Apple sandbox-exec; golangci-lint 2.12.2 |
| Native root SHA-256 | `8c9af1ca52b75aaa1e67af6365f1e39dfcc3c56a274d3571270079ef669ac9d0` |
| Snapshot manifest SHA-256 | `924378006ba6e6e7db9d0277375456a958c14cbaff03d1ebd63dfa6e4b7ceac8` |
| Snapshot launch SHA-256 | `2991bc9c97904d77aef70bb5c3ea7a7044ff462382310ae613bbd31f461fd5f1` |

The installed executable was copied to an ignored task-local read-only file;
source and copy matched the digest. The updater-managed source and launcher
were untouched. Only the copy was executed, behind the harness's digest gate.
Without `MUSE1413_BINARY`, every main vendor test skips before fixture creation.
Ordinary `go test ./...` excludes this hidden evidence package; the explicit
ungated run checks skips and non-vendor controls.

Normative sources reread from fetched spec main: environments §§7.1 P1–P4,
7.9, 10.2; CIP-0017 implementation leaf 2 and minimal probes 1–5;
`conformance/environments-muse-v4/README.md`; CIP-0002's native exclusion,
resume and trusted-mode boundaries. Earlier context/credential observations
are not promoted to this build.

Every vendor invocation uses `MUSE_NO_AUTO_UPDATE=1`, `env -i`, disposable
HOME and four XDG parents under task-local `.temp/`, file credentials and
`--provider echo`. The OS sandbox denies network, security-service lookups,
Keychain files and the real native config/library directories. Paired controls
observe curl exits 0 / 7 with exactly one server hit, and ordinary versus
synthetic Keychains file read exits 0 / 1 with unchanged bytes and
`Operation not permitted`. No real Muse home, credential or session was read
or changed; no auth command, remote inference or real Keychain item was used.

Interactive runs in private background tmux with `/dev/null` configuration.
Startup, an echo sentinel, `/rules`, then `/exit` flush diagnostics without
using the owner's UI. This measures loader behavior, not model-visible
removal. Fixture inputs and native root bytes remain unchanged; main vendor
homes contain no credential file. Synthetic `{}` auth exists only in the
separate Curator resolver fixture, where Muse never runs.

## Fixtures, commands and production entrypoints

[Harness and fixtures](../../.research/muse1413/README.md) retain the 1.4.11
intake, replay and reviewer tests. `testdata` bytes are unchanged, including
the historical `muse148-probe` identity. Current Curator drives real
`profile install SOURCE`, then `env resolve muse --profile muse148-probe
--repair --format json`; the resulting symlink must match the frozen root.
The reference wrapper classifies complete physical closure without following
links, then calls published `managed_home_snapshots.check_bundle`. Listed
and unlisted dangling links refuse. Private snapshot files are 0600; source
mode alone is waived because Git cannot retain 0600. This checks representation
and loading, not export provenance or hostile concurrent filesystem mutation.

```text
exec:        $MUSE1413_BINARY exec --json --provider echo --trust-workspace [--no-foreign-personal-context] MUSE1413_OFFLINE_CONTROL
interactive: $MUSE1413_BINARY --provider echo --trust-workspace [-c context.foreign_personal_skills=false]
```

`TestMuse1413LoaderMatrix` drives 36 cases, parser refusals 2, exec config-key
observation 1, interactive regression witnesses 2, workspace residual 4:
**45 loader/parser cases**, plus isolated version inspection. Vendor exits:
43 at 0, two intended parser refusals at 2. The green gated suite pins the
interactive failure; it does not qualify interactive exclusion.

| Probe | Exec observation and vendor exit | Interactive observation and vendor exit | Compared with 1.4.11 |
|---|---|---|---|
| P1: exec exclusion spelling | Accepted; absent-root foreign rules 1 → 0 and skills 1 → 0; native-root skills 1 → 0; exit 0 | Exec-only spelling covered by P2 | Same |
| P2: form/parser boundary | Interactive config key accepted, loader reached, exit 0; parser-only, exclusion semantics unmeasured | Exec-only flag refused with `invalid TUI options`; string false refused with `the JSON type does not fit this setting`; exits 2 before rules loading. Boolean accepted, exit 0 | Same |
| P3: foreign rules fallback | Separate Claude/Codex rules: absent root `user_sources=1 foreign_sources=1`; linked/snapshot root `user_sources=1 foreign_sources=0`. Flag removes absent-root fallback; exits 0 | Same root contrast; absent-root boolean retains fallback at 1. Native root suppresses fallback with or without boolean; exits 0 | Same |
| P4: persistent skills and required exclusion | Linked/snapshot roots retain one foreign skill without flag, zero with flag; native root preserved; exits 0 | Both roots retain one skill with or without boolean, including retained-skill banner. Absent-root boolean also retains one; exits 0. **Required exclusion fails** | Same |

Workspace residual: both forms load native and workspace rules for both root
representations, exit 0. Interactive `/rules` enumerates both; project count
changes 1 → 0 on reload. This is enumeration only; model-visible precedence
and conflict resolution remain unknown. MSP schema/session behavior was not
qualified by this loader replay.

[Observations](../../.research/muse1413/observations.json) bind argv, exits,
loader records and bounded visible markers to source digests. The
[comparison](../../.research/muse1413/comparison-1.4.11.json) finds **45/45
unchanged** for normalized argv, exits, ordered event/outcome/source/skill
counts and marker sets. Timings and sentinel-dependent lengths are excluded.

### CIP-0002 resume/trust finding

**Still open, and directly reproduced on 1.4.13.** The previous
[1.4.11 P7 record](project-source-exclusion-muse.md) did not qualify complete
resume exclusion. The new [P7 replay](../../.research/muse1413/resume-trust/README.md)
uses the same native entry and synthetic sources, with new exact-byte intake.
[Native excerpts](../../.research/muse1413/evidence/resume-observations.txt)
and [full replay](../../.research/muse1413/evidence/resume-04.log) show:

| Stage | Project rules | Project skills | Runnable hooks | Project MCP sources | Effect |
|---|---:|---:|---:|---:|---|
| Fresh untrusted exec | 0 | 0 | 1 home / 2 total | 0 | Home context/skill/MCP and hook work; project hook marker absent |
| `resume --last`, before choice | Saved session | Pending trust | Pending trust | Pending trust | Offers trust or `2 Quit`; no continue-untrusted choice |
| Resume after fixture-only trust | `/rules`: `project: none` | 1 | 2 / 2 | 1 | Synthetic project hook executes; selected home sources remain |

The saved root-rule listing does not prove live catalog exclusion. The
project skill, hook and MCP revival is the same failure class as 1.4.11.
This finding concerns strict project-source admission under CIP-0002;
it is distinct from foreign **personal** P1–P4 exclusion. Exec P1–P4 success
cannot close it. Trusted-mode delivery in current CIP-0002 does not claim a
hostile-repository boundary and does not resolve this strict-mode finding.
No complete untrusted interactive/resume recipe, lazy-tool or native-child
inheritance, or permission-policy enforcement is qualified here.

## Coverage map

**6/6 task AC rows driven**, with qualification separated from successful
evidence collection. The task originally had placeholder scope/AC fields;
the board now records the six concrete rows below from the requested scope.

| AC row | Named tests / production entry | Result and narrowing evidence |
|---|---|---|
| 1: exact identity and opt-in | `TestVendorIntakeBoundary`, `TestMuse1413Identity`; real setup and sandboxed version | Wrong bytes refused, unset gate skips before fixture; foreign-digest mutant killed |
| 2: P1–P4 and residual | `TestMuse1413LoaderMatrix`, parser/config-key tests, interactive regression and workspace tests; real parsers/loaders, Curator install/resolve | Exec qualifies; interactive fails on linked/snapshot roots; exec omission caught |
| 3: evidence and guards | `TestObservationChecker`, `TestReviewContradictoryObservation`, frozen/source/reference closure and offline tests | Ten mutants killed at intended assertions; interactive omission declared survivor |
| 4: version comparison | 45 sanitized case pairs against retained 1.4.11 observations | All unchanged; 1.4.12 explicitly skipped, not qualified |
| 5: resume/trust disposition | `resume-trust/TestP7ResumeTrustGate`, `TestChecker`, `TestOfflineBoundary`; real exec → resume → trust → echo → `/rules` | Project catalogs and hook revive; pre-trust negative and home positives pass; sandbox controls killed |
| 6: delivery | Signed feature PR, exact-head hosted CI, task outcome and topic handoff | Delivery links and check results attached to the task; no runtime gate changes |

Out-of-contract rows: credentials/login/refresh, broker keys and model-visible
removal (offline scope; CIP-0017 probes 3–5 need separate evidence); actual
MCP handshake/tool use (`/usr/bin/true` is a configuration marker), lazy read
and native children; model-visible workspace precedence; export provenance,
capture locks, publication/destination admission and materialize integration;
serve, runtime consumer gates, registry pins and other platforms. The foreign
exec config-key case establishes parser acceptance only.

## Validation and release effect

[Commands and actual exits](../../.research/muse1413/evidence/commands.json),
[mutant results](../../.research/muse1413/evidence/mutants-results.json) and
[task gaps](../../.research/muse1413/TASK-261026-jk47fn_gaps.md) preserve the
verification boundary. Main replay, final resume replay, ungated controls,
production build and harness vet/lint pass. Ten patch mutants and exec
omission fail their named assertions. Interactive omission survives because
the real boolean already retains the skill. Resume sandbox-denial mutants
fail only synthetic curl/cat controls, never a weakened vendor run.

Initial resume translation and lint failures are retained with their fixes
in the logbook. No previous hosted pass is reused as current CI evidence;
the PR must have green hosted checks for its delivered head. Hidden vendor
packages require the explicit local opt-in run; default CI does not execute
installed Muse. PR check links and exact reviewed head are task attachments.

Exec qualifies only for the measured macOS arm64 loader rows. Interactive
remains refused; its §7.9 release gate remains pending. The CIP-0002 strict
resume/trust finding remains open on 1.4.13. Consumer admission remains with
its owners; no runtime gate or pin is changed and no authentication,
portable isolation or model-visible removal claim follows.
