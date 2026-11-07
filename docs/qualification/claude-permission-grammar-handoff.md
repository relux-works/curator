# Claude Code 2.1.317 / 2.1.319 permission grammar: producer handoff

TASK-261107-oe1ki9 — qualify-claude-permission-grammar-2-1-317-and-319,
under STORY-261105-fevga9 — profile-launches-on-linux-agent-hosts.
Observed 2026-11-07. Rework base: current `main` at
`470331906654c65169b88bb9579c0c929ed272df`, replacing the original
`ed5688b7121c56f23ca18e80e8e51d65d82fbbd2` base in one rebase.

**Measurements offered, no admission row proposed.** The earlier V2 proposals
are withdrawn. Current upstream main has only Claude 2.1.312 under V3, with
2.1.261 retired. The literal bypass spelling and six documented modes agree
with the stored historical text; that comparison does not establish grammar
compatibility. Upstream recorded drift by 2.1.312, and 2.1.319 still parses
`--inherit-permission-mode`. The earlier conclusion that no check-logic change
was indicated is withdrawn. No gate, launcher, dependency pin or installed
executable was changed here.

- [2.1.317 record](claude-2.1.317.md) and [exact observations](claude-2.1.317-observations.json).
- [2.1.319 record](claude-2.1.319.md) and [exact observations](claude-2.1.319-observations.json).
- [Control results](claude-permission-grammar-controls.json) and [evidence checker](claude-permission-grammar-controls.py).

## Where the gate actually lives

Read fetched `skill-agents-management` main at
`d97ec5166f479c08410881dc38c9cc945dfdc433`, rather than the earlier
`46d17fb7d8eb0de79dafd91a971de6344d34ebfb` checkpoint.
Read fetched `curator-agent-launcher` main at
`2045339f703ee8190705c116a05197b12b684b71`.
These identify source inspection, not an installed launcher's build.

| Surface | Responsibility and current evidence |
| --- | --- |
| SAM [`internal/releasegrammar/release.go`](https://github.com/relux-works/skill-agents-management/blob/d97ec5166f479c08410881dc38c9cc945dfdc433/internal/releasegrammar/release.go) | `IsReleaseTriple` validates a numeric triple. No capability rows or range admission live here. |
| SAM [`pkg/agentic/systems/claude/policy.go`](https://github.com/relux-works/skill-agents-management/blob/d97ec5166f479c08410881dc38c9cc945dfdc433/pkg/agentic/systems/claude/policy.go) | `verifiedReleases` has only `{Release: "2.1.312", Grammar: agentic.PermissionGrammarV3, YoloSupported: true}`. It owns the bypass mapping, selector refusals and V3 explicit-settings scan. 2.1.261 is retired. |
| SAM [`pkg/agentic/systems/claude/probe.go`](https://github.com/relux-works/skill-agents-management/blob/d97ec5166f479c08410881dc38c9cc945dfdc433/pkg/agentic/systems/claude/probe.go) and [`pkg/agentic/system.go`](https://github.com/relux-works/skill-agents-management/blob/d97ec5166f479c08410881dc38c9cc945dfdc433/pkg/agentic/system.go) | Resolve the binary, probe its version and perform exact release/grammar lookup. `ReleaseCapability` has only `Release`, `Grammar`, `YoloSupported`; no OS, architecture or digest bound. The row also admits print classification. |
| Launcher [`cmd/curator-run/main.go`](https://github.com/relux-works/curator-agent-launcher/blob/2045339f703ee8190705c116a05197b12b684b71/cmd/curator-run/main.go) | `resolvePermission` probes then calls registry `PermissionMapping`. Unknown yolo becomes `plan_refused` before execution; native tolerates an unknown mapping and adds no override. Silence defaults to yolo across stdio shapes, so admitting a row also changes launches without an explicit permission choice. |
| Launcher [`internal/plan/plan.go`](https://github.com/relux-works/curator-agent-launcher/blob/2045339f703ee8190705c116a05197b12b684b71/internal/plan/plan.go) | Passes the request into `vendorplugin.BuildLaunch` and wraps admission errors as `plan_refused`. |
| Launcher [`go.mod`](https://github.com/relux-works/curator-agent-launcher/blob/2045339f703ee8190705c116a05197b12b684b71/go.mod) versus SAM [`v0.5.74` policy](https://github.com/relux-works/skill-agents-management/blob/v0.5.74/pkg/agentic/systems/claude/policy.go) | The launcher pins `v0.5.74`, whose only Claude row is 2.1.261 under V2. Fetched SAM main and the linked release are different gates. |

Curator's root module does not own this table. Its `internal/envregistry`
Claude `VerifiedRelease` is an environment/credential fact, not this gate.
No new row reaches a launch through the inspected consumer until SAM releases
it and the launcher moves its pin. Adoption of current main would also retire
2.1.261 and bring V3, not merely add two releases.

## What current V3 requires

These are verified source requirements, not results of running SAM's tests in
this task. V3 retains V2's closed mode/conflict checks, duplicate bypass
refusal, restricted/allow-bypass conflicts and unknown-mode refusal. Under
yolo it additionally refuses these six unqualified selectors with
`ErrNativePolicyUnknown`, respecting option ownership and the native `--`
boundary:

```text
--inherit-permission-mode
--plan-mode-required
--permission-prompts
--enable-auto-mode
--managed-settings
--project-config-root
```

Native forwarding does not acquire a bypass override and remains subject to
tool re-enable guards. The selector refusals do not claim that a selector is
invalid to Claude: they refuse unqualified interaction with the mapped bypass.

The current [`toolpolicy.go`](https://github.com/relux-works/skill-agents-management/blob/d97ec5166f479c08410881dc38c9cc945dfdc433/pkg/agentic/systems/claude/toolpolicy.go)
requires Commander and eager settings selection to agree on presence and the
last effective explicit `--settings` value; disagreement is a typed ambiguity
refusal. Inline JSON and files are read without rewriting. Unreadable or
invalid sources fail with typed settings-policy refusals. The existing
AskUserQuestion re-enable guard applies to explicit allow rules in both modes.
V3 yolo further admits only:

| Settings member | Qualified shape; everything else refused |
| --- | --- |
| `$schema`, `model` | Non-null strings |
| `env` | Empty object only; every nonempty env is unqualified, including `CLAUDE_CODE_SUBPROCESS_ENV_SCRUB` or configuration-root selectors |
| `permissions` | Non-null object containing only `allow`, `deny`, `ask`, `additionalDirectories`; each a non-null array of non-null strings |
| Other keys | Unqualified, including `hooks`, `enabledPlugins`, unknown top-level members, `permissions.disableBypassPermissionsMode` and `permissions.defaultMode` |

This scan covers explicit argv settings, not implicit user/project/managed
configuration or ambient environment. It is not a complete enforcement sandbox.
The source explains 2.1.312 bypass interactions with env, hooks and permission
defaults; this task did not execute those interactions on either newer binary.

Upstream [`bc35b595f6f7ec34eb90c853d0e041035fb46702`](https://github.com/relux-works/skill-agents-management/commit/bc35b595f6f7ec34eb90c853d0e041035fb46702)
recorded drift by 2.1.312: four unqualified selectors and hidden required
scalars `--fast-mode` and `--runner-declares`, including parser and eager-table
updates. Subsequent changes yield the six-selector/settings policy above.
Unchanged visible help does not exclude hidden-option or settings drift.

The 2.1.312 bar in [`nativegrammar_test.go`](https://github.com/relux-works/skill-agents-management/blob/d97ec5166f479c08410881dc38c9cc945dfdc433/pkg/agentic/systems/claude/nativegrammar_test.go)
includes both pinned comparisons:

- `TestPinnedClaudeNativeGrammarComparison` derives visible help arities independently and executes extracted native Commander parsing and list normalization against the module's option ownership and rule-preservation corpus.
- `TestPinnedClaudeNativeListAndSettingsComparison` executes the binary's list tokenizer over Unicode cases and eager settings scan over ownership/placement cases, checks exact scalar/variadic/optional sets in both directions, and compares last-source selection and normalized lists.

The release-specific anchors in `pinnedClaudeReleases` name only 2.1.312.
By source inspection, 2.1.317/319 skip before those comparisons; missing anchors
for a pinned release fail rather than fall back. These comparisons were
**NOT_RUN** here, and no anchors, tables, gate tests or implementation were
changed. Extending that evidence and resolving any parser/eager/settings
changes belongs in SAM's lane. Whether unchanged production check logic
suffices remains unestablished. Tuple-bound admission, if required, needs a
SAM capability-contract and consumer change; none is written here.

## What these measurements supply

The historical comparison is to stored row/comments/test text, not a tested
2.1.261 binary. Some old comments cite 2.1.273/274 and option tables 2.1.288.
Only the following literal comparison is supported:

| Compared item | Stored historical text | Both measured macOS arm64 releases |
| --- | --- | --- |
| Bypass spelling | `--dangerously-skip-permissions` | Same documented spelling; P22 reaches help |
| Documented modes | `acceptEdits`, `auto`, `bypassPermissions`, `manual`, `dontAsk`, `plan` | Same choices; separate/equals forms reach help; invalid selectors refuse with exit 1 |

There is no complete historical grammar diff and no general no-drift finding.
`YoloSupported` means only that the tool documents a permission-bypass flag;
root help supplies that narrow evidence. It does not supply a verified
`Grammar` token or permission enforcement.

Both original 57-row matrices retain exact argv, streams and exits (47 help
or version successes, 10 refusals). Both add 39 bounded rows H01–H39 for current
V3 concerns (34 successes, 5 refusals). Corresponding results agree after
normalizing version banners and disposable locators. Complete root help is
byte-identical, SHA-256
`5ca3a9dac2026e56c221fdd34380c7d7d45ebbb51c1d01b7c17c75cb076a0be8`.
Executable SHA-256 is unchanged before/after each version's supplement.

H03–H06 show inherited-mode parsing and its invalid-choice refusal on **both**
versions, including 2.1.319. H17–H19 and H38–H39 pair an invalid-selector control
with hidden-scalar token ownership. H20–H34 provide inline/file/repeated-source
help observations for V3-sensitive settings. Settings that V3 refuses still
reach native help; those successes establish neither runtime validation nor
effective source selection. P43/P44/P57 already cover permission-prompts
choices. Native `--` placement, print help, allowed/disallowed tools, settings,
MCP and launcher prefix spellings remain in P01–P57.

Unknown sandbox examples and invalid settings/MCP content reaching help are
retained counterexamples to treating help success as support. Bypass/runtime
interaction, hooks, managed settings application, complete hidden-option
coverage, effective tool rules, MCP connections and model availability are
**NOT_RUN**. Missing-value hidden scalars and session-based checks stop at that
boundary: they could consume the only help flag or start a session. No session
was attempted. Bounded help probes do not replace pinned native-function
comparisons or establish that the launcher's full grammar is correct.

The [Codex precedent](codex-0.164.0.md) also had startup schema/config-read
observations without a model turn; no equivalent effective-config observation
exists here. Adjacent Muse records likewise separate measured slices from
pending consumer qualification. Their broader conclusions do not transfer.

## Linux boundary and gate-owner decision

Another lane reported a 57-row Linux x86_64 glibc matrix for 2.1.317. The
independent review read it and reported agreement row for row with the macOS
matrix, with `qualification_accept=false`. This is attributed evidence, not a
Linux qualification claim or a Linux run here. No packet, receipts, identity
or private locator from that lane is republished. There is **no Linux 2.1.319
evidence** here, nor any Linux pinned parser/eager comparison.

The gate owner's decision is explicit: a row is version-only. A 2.1.317 row
admits other architectures, libc variants and OS tuples nobody measured;
a 2.1.319 row admits **all Linux** without Linux measurements. Either the owner
knowingly accepts that evidence boundary or changes admission to express a
narrower tuple. These documents offer measurements, not either decision or a
V2/V3 row. Independent review, SAM release, launcher pin adoption and runtime
acceptance remain separate gates.

For any further Linux work, the producing lane must supply sanitized exact
version/hash before and after, OS/kernel/architecture/libc, provenance and
published integrity, argv/streams/exits, disposable HOME and selector clearing,
Linux-native OS network-denial and sanitizer controls, plus parser/list/eager
and settings evidence for the intended admission scope. Nothing is inferred
from macOS, a matching version string or another unmeasured tuple. Stop that
version if direct execution needs installation or a session. No remote Linux
execution or provider access is authorized by this record.

## Isolation and executed controls

Every native invocation used a fresh absolute HOME and cwd, closed stdin and
a 25-second timeout. The environment started empty: only HOME, system PATH,
disposable TMPDIR, C locale, DISABLE_AUTOUPDATER, DISABLE_TELEMETRY and
CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC were supplied. All other selectors,
including provider, proxy, socket, XDG and Git selectors, were absent.
The Apple sandbox allowed default operations but denied `network*`, reads and
writes to real native configuration/Keychains locations, and security-service
mach lookups. No real configuration was read or copied. Each record includes
the sanitized invocation prefix. Only version/help paths were exercised; no
model, login, provider request, MCP process or shared daemon was started.

The 2.1.317 executable was extracted directly from the published native npm
package after SHA-512/SHA-1 verification, without install scripts. It ran
without installation. The installed 2.1.319 executable was used read-only.

The previously unsupported false-success and wrong-integrity claims were
rerun. [Actual captured outputs and exits](claude-permission-grammar-controls.json)
now accompany the [executable evidence checker](claude-permission-grammar-controls.py):

| Control | Positive output / exit | Planted or guarded result / exit |
| --- | --- | --- |
| Invalid-selector record, each version | `INVALID_SELECTOR_REFUSAL_CONFIRMED` / 0 | Set recorded exit to 0: `INVALID_SELECTOR_EXIT_MISMATCH: expected 1, observed 0` / 1 |
| Cached wrapper and native tarballs, separately | `RECORDED_PUBLISHED_INTEGRITY_CONFIRMED` / 0 | Change expected SRI digest: `PUBLISHED_SHA512_INTEGRITY_MISMATCH` / 1 |
| OS network guard | `LOOPBACK_BIND_OK` / 0 | Same loopback bind under probe sandbox: `PermissionError: [Errno 1] Operation not permitted` / 1 |
| Sanitizer | Clean symbolic record and sanitized plant accepted | Actual-home prefix planted in memory detected; result `real_home_plant_detected=true` |

These plants exercise the checker, not vendor runtime. Integrity controls use
cached intake bytes and the published metadata recorded with the original
intake, not a new registry fetch or independently verified publisher signature.
The wrong-integrity plant reaches the intended SHA-512 mismatch after the
unchanged SHA-256 check; it does not fail on setup or missing input.

Reproduce selector controls from the repository root with Python 3:

```sh
python3 docs/qualification/claude-permission-grammar-controls.py selector docs/qualification/claude-2.1.317-observations.json
python3 docs/qualification/claude-permission-grammar-controls.py selector docs/qualification/claude-2.1.317-observations.json --plant
```

Repeat with the 2.1.319 record. For integrity, `$NATIVE_TARBALL` is a locally
staged native package tarball whose hash matches the record:

```sh
python3 docs/qualification/claude-permission-grammar-controls.py integrity docs/qualification/claude-2.1.317-observations.json --tarball "$NATIVE_TARBALL" --package @anthropic-ai/claude-code-darwin-arm64
python3 docs/qualification/claude-permission-grammar-controls.py integrity docs/qualification/claude-2.1.317-observations.json --tarball "$NATIVE_TARBALL" --package @anthropic-ai/claude-code-darwin-arm64 --plant
```

Use the wrapper tarball and `@anthropic-ai/claude-code` for the other pair.
All checker subprocesses also received disposable HOME and an empty allowlist
environment. Plant commands intentionally exit 1; stdout is empty on refusals.

The sanitizer control used the same sanitizer/checker as the native probes.
It formed the actual home prefix in memory without reading that directory,
rejected it, then accepted its symbolic replacement and a clean positive.
Only the sanitized plant `$REAL_HOME/.claude/CONTROL_ONLY` and Boolean results
were persisted publicly. Its executed claim was:

```python
plant = str(pathlib.Path.home()) + "/.claude/CONTROL_ONLY"  # string only
assert not clean(plant)
assert clean(sanitize(plant))
assert clean("symbolic-positive")
```

Private intake, PID and command logs remain private. No full suite, SAM gate
plant or pinned comparison was run in this rework. The review's scratch V2
plant remains the reviewer's attributed finding; its complete V3 outputs were
not reverified here. Source inspection suffices to withdraw the proposals.
This remains a producer handoff for renewed review, not acceptance.
