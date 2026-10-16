# M2 — Qualify the Muse 1.4.8 loader without inventing model evidence

- **Owner repository:** curator (qualification record); agents-management
  consumes the exact evidence identity in its subsequent release.
- **Normative clauses:** environments §§7.1 P1–P4, 7.9 Muse qualification
  order, 10.2; CIP-0017 implementation leaf 2 and Test plan items 1–4;
  `conformance/environments-muse-v4/README.md`.
- **Attack surface:** vendor-build evidence authorizing form-specific exclusion.
- **Size:** M. Evidence leaf, not a new runtime or credentials migration.

## Acceptance and the no-model boundary

Record exact version/build, platform, fixture bytes, command and exit status,
observation channel, and which production entrypoint executed. Use disposable
homes/accounts and synthetic sentinels; never an active native login.

| Required observation | Possible without model calls? | Evidence that is insufficient |
|---|---|---|
| Exact exec flag accepted; interactive same flag rejected; interactive `-c` JSON false accepted | Yes, if real parser entrypoints can terminate before inference | Fake argv, `--help` alone, or generic process failure |
| Native root suppresses foreign rule fallback; foreign skills persist without exclusion | Only if the real loader exposes pre-inference records | Successful parse, expected fixture contents, or quota/auth failure before loader |
| Exec exclusion and interactive boolean remove the foreign loader sources with root present | Same condition: observable real loader before any model request | Correct argv without loader observation |
| Interactive override without native root does not suppress rules fallback | Same conditional real-loader lane | Treating a missing own file as exclusive |
| Curator link and frozen regular-file snapshot fixture both load native root | Same conditional real-loader lane; fixture is independently checked against the published snapshot contract | Calling the fixture proof that the new materialize command works |
| Workspace conflict | Record loader order/conflict if observable; model-visible precedence remains unknown without a model run | Inferring behavioral precedence solely from file enumeration |
| Credential backend/login/refresh or model-visible foreign-content removal | Separate evidence; not established by this exclusion replay | Promoting the input pack's backend facts into loader qualification |

First establish a bounded pre-inference observation path on the exact binary.
The concrete first candidate is Muse's offline `--provider echo`: released
agents-management v0.5.66 (`acae871`) README:968 and
`systems/muse/args.go:117,159` already construct it for both forms. Verify the
real 1.4.8 echo path reaches the same context loader and exposes the required
records without authentication or remote inference; do not infer that from
the plugin's fake-binary/argv tests. The current managed v4 plugin deliberately
refuses 1.4.8 with a root fact before qualification. Use a separately scoped
vendor probe to collect evidence; never lie about ToolRelease or disable the
production gate merely to make the probe start.

Observe that it reaches the loader, terminates before a model request, and
does not alter credential stores. Do not assume such a path exists. Quota
exhaustion is not a reliable network block or proof of no model request.
If no safe observable path exists, record the exact missing rows as unverified
and keep the release gate pending until the authorized live replay is possible.
Do not weaken §7.9 because the Muse quota lasts until 2026-10-19.

P1–P4's original claims are loader claims, so genuine pre-inference loader
observations can qualify those narrow claims without making model-visible
claims. A workspace/model claim that cannot be observed stays explicitly
unknown; if the owners interpret the required residual as needing a live
behavioral replay, that row keeps the gate open rather than silently passing.

## Mandatory negative controls

For each form run root absent/present, foreign rules/skills separately, and
exclusion absent/present. The control without exclusion must visibly load the
foreign source that the positive exclusion case removes. Plant one omission
of the required exclusion: the observation checker must report that foreign
source, not a timeout/unrelated error. Repeat linked and snapshot-file roots.
Check string `"false"` versus boolean `false` and the rejected interactive
exec-only flag independently. Record any unsupported spelling as refusal.

Runtime consumers must refuse required exclusion on unqualified build or
serve; absence of a fact remains a residual, never authority to inspect paths.
Tests of those consumer refusals are not substitutes for vendor replay.

## Dependencies and release effect

Can prepare fixtures in parallel with M1. For the narrow regular-file loader
row, freeze a snapshot fixture using the verified Curator root bytes, declared
XDG placement and published snapshot grammar; check its regular-file closure
and expected identity independently. Compare real loader observations against
the real Curator link. File representation and loader behavior are the tested
property; this does not qualify export provenance, capture locks, publication,
destination admission or runtime-B protection. A loose arbitrary hand-copy
without that fixture evidence is insufficient.

This scoped interpretation needs no SNAP-CLI predecessor for M2. The actual
`env materialize -> Muse loader` integration is a separate required SNAP-CLI
acceptance row when that command ships, using the same frozen expected bytes;
it must not be reported as covered by the fixture. No token-refresh or broker
qualification dependency.
M2 precedes release of the v4 producer/consumer pair and every release claiming
1.4.8 qualification. D10 credential-mode admission remains accepted meanwhile.

Paths: curator's normal task evidence/documentation location, existing Muse
probe fixtures if available, and a sanitized evidence update consumed by
registry/status and plugin qualification. Coordinate those code paths through
M1/M3; this leaf does not independently edit their qualification constants.
