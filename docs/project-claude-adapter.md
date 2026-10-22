# Claude project-home adapter

`internal/projectcontext` prepares Claude Code `2.1.305` control outputs from
an admitted composition. This is an internal materialization and construction
seam for a trusted Werkstatt. It advertises no complete strict launch tuple.
The [native qualification record](qualification/project-source-exclusion-claude.md)
retains the exact executable digest, macOS arm64 identity and bounded probes.

## Explicit operator modes

`native-project` intentionally allows the harness to discover project files
natively. Its status and launch refusal carry **project sources not isolated**.
The managed root document is profile-only; project instruction documents are
not concatenated into either managed instruction channel. Curator's approved
whole-skill and MCP outputs remain in the offline candidate.

`PrepareClaudeMode` requires the explicit mode, a consistent composition byte
map/receipt, matching manager controls, and a `ClaudeProfileDocumentSource`.
The source's `ProfileDocuments(ctx, compositionRef, compositionSHA256)` must
return I-COMP's dedicated `CompositionPreview.ProfileDocuments` and
`ProfileDocumentReceipt` for that protected composition. It must never split
the combined instruction text. The interface deliberately has no fallback;
`ClaudeProfileDocumentsUnavailable` returns a typed refusal until the owning
I-COMP r3 implementation is connected. Tests use a dedicated source fixture,
not an assertion that r3 protected resolution has already shipped.

Only `CLAUDE.md` and the optional additive system document are accepted from
that source, with exact receipt bytes and private data-file modes. The plan
retains whole approved skill files and MCP, regenerates the same manager
settings/marker, excludes command activation outputs, and copies all bytes.
Its candidate receipt binds mode, form and the profile-only document receipt.
It differs from the strict receipt and cannot be substituted for the original
protected composition or published through its lease. Preparation is inert:
there is no filesystem writer, lease publisher, native argv or environment.

`Authority.ClaudeLaunchMode` uses the shared launch-result shape and currently
returns no argv/environment for any mode. Native-project's mandatory manager
permissions and memory-off remain unqualified; its typed diagnostic names
both and preserves the notice even for resume/fork attempts. `strict` retains
the complete-tuple typed refusal; an absent/unknown mode is invalid rather than
an implicit native-project opt-in. The r1 strict rendering/construction seam
below remains available as internal preparation, never launch authorization.

All report fields use the candidate shared `projectadapter.Certification`
shape, including Mode, Form and Notice. `ClaudeCertification` only embeds it
to retain the provider's strict-gate method. Shared `types.go` matches the
published Muse candidate exactly; mode constants live separately. The receipt
Surface alias uses the same shared projection mirroring I-LEASE.

The [qualification gap inventory](../.research/project-claude-native-project-gaps.md)
identifies P6's exact limited permission evidence, the unprobed memory-off
precondition, and the discriminating native probes needed for either surface.
This implementation performs no native experiment or login and makes no new
native certification claim.

## Protected outputs

Set `ComposeRequest.Claude` to manager-resolved `ClaudeControls{Version: 1}`
before preview and approval of the composition digest. This opt-in requires
the exact Claude release and `macos` adapter platform; generic compositions
without it cannot use the native publication entry point. Other environments
and releases refuse. The selected profile lock remains unchanged.

Composition retains I-COMP's profile-first `CLAUDE.md`, project item ordering,
empty-document rule and one-additional-LF separator. Profile and project
system parts use the separate additive system channel. Whole effective skills
use the owning context whitelist, excluding developer/build/runtime material;
this adapter does not merge skills or change selection. The admitted MCP union
uses `.agent-context/mcp/claude_code.json` and the existing canonical stdio
encoding. An empty union writes `{"mcpServers":{}}` plus LF explicitly. HTTP
loading remains unsupported under the cited adaptation sheet's evidence.

Only manager controls generate `settings.json`: sorted unique exact built-in
tool grants/denies and `defaultMode: dontAsk`. Default grants are empty. The
construction seam also supplies `--tools` with exactly the granted built-ins,
including an empty argument for no grants. It denies `mcp__*` in settings;
MCP connections can load the approved server set, but MCP tool-call grants
have no supported input arm in this slice. MCP permission names, arbitrary
permission expressions and native settings cannot enter `ClaudeControls`.
Explicit manager denies retain native precedence over grants.

`dontAsk` alone is insufficient for default deny: working-directory reads can
still run without a prompt. The explicit built-in tool set and MCP deny are
therefore required together. See the vendor's
[permission modes](https://code.claude.com/docs/en/permissions#permission-modes)
and [CLI tool controls](https://code.claude.com/docs/en/cli-reference).
These documented control semantics are not a fresh native qualification run
of this exact generated policy. The typed report names that remaining gate.

`.agent-context/claude-managed.json` records the adapter owner and exact MCP
and settings hashes. This internal marker has no independent admission power;
only the issuing manager's protected composition and lease receipt confer
ownership. It is not the shared-profile MCP writer's per-entry ownership
carrier or capability. Project outputs never write a shared profile home.
The composition receipt includes every generated file, including settings,
the empty MCP file and marker, with exact SHA-256, size and `0600`/`0700` mode.
Native mutable state, login material and source checkout files are excluded.
Repository `.claude/settings.json`, local settings and permissions are never
read by this adapter or copied into native settings.

## Lease and construction seams

`Authority.PublishClaude` re-resolves protected composition bytes and current
admission, then calls `ClaudePublisher.PublishPrepared(ctx, leaseID, files,
false)`. This interface is exactly the I-LEASE manager's publication method.
The lease manager owns reservation, current receipt checks, transactional
publication, busy refusal and lifecycle. The adapter never writes around it.
The read lock ends before delegation; the lease owner must independently
recheck authority and compare the complete supplied output against the
reserved composition. There is no fallback publication path.

`Composition.Generated` supplies the I-LEASE projection:
`map[path]Surface{SHA256, Size, Mode}`. Native augmentation occurs before the
receipt and composition hashes are calculated, never after lease admission.
The I-COMP base still mirrors the lease result types until the owning packages
are integrated; this change neither copies nor modifies lease implementation.

`Authority.ClaudeLaunch` borrows a live same-manager `managerlock.Held`, checks
fresh admission and every generated on-disk receipt entry, and returns a
`ClaudeLaunchContribution` with the composition digest and home reference.
It selects the canonical stable home at
`environments/project-views/<checkout>/<profile>/claude_code`. Its existing
`envfragment.Fragment` carries home, profile and registered channel descriptors;
its complete `Args` carries `--setting-sources user`, explicit MCP path and
`--strict-mcp-config`, `dontAsk`, exact built-in tools and the optional append
system-prompt path. Consumers use that complete argv once, without expanding
the fragment's channel descriptors a second time.

`Environment()` adds the fixed `CLAUDE_CODE_DISABLE_AUTO_MEMORY=1` control to
the managed `CLAUDE_CONFIG_DIR`. It accepts no caller environment or argv.
MCP metadata carries only the sorted declared environment-name union, with
no values. Existing destination allowlists/ceilings still govern forwarding.
This adapter performs no login, credential lookup, provider start or MCP call.
Claude setup-token remains in its separately owned channel.

This contribution is internal construction data, not permission to start.
The destination must negotiate a project-capable outer carrier, bind context
and lease, verify the actual provider identity, and transfer the reserved
lease before execution. It must not emit these project facts as standalone
legacy v2, widen a frozen carrier or treat a prepared file set as qualification.
No CLI project launch or capability advertisement is enabled here.

## Certification and refusal

`ClaudeCertificationReport` returns shared Certified and Refused surface
lists for strict mode at the exact release/platform/architecture. Its measured entries are
bounded by the linked native record: six startup instruction locations, three
nested Read instruction paths, selected skill/reload cases, explicit stdio
MCP, startup permission atoms, excluded-origin resume and one background Task
child. They do not generalize to arbitrary imports, settings keys, tool paths,
plugins, aliases, other child modes or resumed hooks/MCP.

Late hook, plugin, MCP and settings reloads, imports/additional directories,
automatic-memory-off behavior and the newly generated default-deny policy
remain uncertified. `RequireStrict()` always returns typed
`project_context_unsupported`. Another version/platform/architecture borrows
none of the measured entries. This change does not promote these observations
to a complete certified adapter record.

Resume, fork and unknown entry modes return typed `ClaudeDiagnostic` before
any authority lookup. All resumes are conservatively refused until protected
excluded-origin provenance is integrated; changing startup flags cannot scrub
contaminated history. Receipt drift emits no launch contribution and never
replaces the protected baseline. Missing/invalid publication authority never
reaches the publisher. Lease publication errors propagate without retrying
through a direct writer or choosing another home.

## Verification

Use the configured pinned Go toolchain, unset inherited Git directory variables,
and keep private caches/logs under task-scoped `.temp/`. Use a short temporary
path on macOS. The ordinary affected-package suite and explicit narrowing lane:

```sh
env -u GIT_DIR -u GIT_WORK_TREE go test ./internal/projectcontext -count=1
env -u GIT_DIR -u GIT_WORK_TREE CURATOR_PROJECT_CLAUDE_MUTANTS=1 go test ./internal/projectcontext -run '^TestClaudeNarrowingMutants$' -count=1 -v -timeout=8m
```

The adapter fixture uses real admission, protected profile/context/MCP packages,
composition and native output construction. A disposable publisher fixture
checks the exact I-LEASE method and receipt projection; it does not claim to
exercise lease transactions, liveness or session transfer. Native contribution
assertions and their controls require macOS arm64; portable rendering and
refusal cases retain their narrower claims. Each test states its claim.
The requested producer validation excludes the full default suite and performs
no optional vendor probe or authentication operation.
