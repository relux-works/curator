# Codex composition delivery

`internal/projectadapter.RenderCodex` prepares an inspectable Codex 0.162.1
native-project plan. It renders trusted project entry, so the pinned harness
loads repository prose and config natively while retaining the selected
**profile root only** in the managed instruction document. Native config,
hooks and plugins are permitted by
[specification §5.1](https://github.com/relux-works/curator-spec/blob/2510171c/protocol/project-context.md#51-explicit-native-project-mode).
Native trust is not Curator item approval.

The native-project launch capability is **refused for Permissions**.
`-c mcp_servers={}` merges with repository servers rather than erasing their
permission settings. In a paired native fixture,
`mcp_servers.project_tool.disabled_tools=["project_tool"]` removes the actual
tool from the MCP catalog despite the complete manager argv. Removing only
that restriction restores the same server's tool and synthetic call result.
Per-tool approval settings also survive in effective config; approval-dialog
behavior is not claimed. `Plan.Report()` refuses Permissions,
`Plan.RequireLaunch()` returns typed `project_context_unsupported`, and
`Plan.PublishPrepared()` cannot reach the lease publisher.

This resolves review R1 using the owner's native-project ruling: repository
config is allowed, but surviving project permission effects require refusal.
The requested launch-time enumeration alternative was also tested using a
read-only TOML parse, rejected unknown/malformed inputs, manager tool grants
and explicit dotted overrides. It restores an admitted tool and prevents a
known denied server's process from starting. On the same app-server, a newly
added repository server loads on the next thread; toggling only
`mcp_servers.late_tool.disabled_tools` changes its actual catalog and restores
a successful call in the unrestricted positive. One-time launch enumeration
does not cover that reload/entry. The candidate stays in the research replay;
no product parser or supported-launch claim is introduced.

The r3 digest fence candidate records every fixture-enumerated config input,
including absent inputs, full file bytes and the sorted MCP server/key/value
set. Before dispatch it re-reads these inputs. Unchanged input admits;
a late server, edited `disabled_tools`, removed server/file, added config,
file/directory link, malformed TOML or unknown key returns a typed
`project_config_changed` requiring relaunch. Rejection neither dispatches nor
repairs or recaptures the baseline. Relaunch re-enumerates the new set and
re-applies manager dotted ceilings; the known unadmitted server stays stopped.

That preflight does not bind Codex's independent later read. The bounded
`TestAdapterConfigFenceReadGap` deterministically changes config **after a
successful check and before `thread/start`**, on the same app-server. The new
server loads and its `disabled_tools` changes the observed catalog. Restoring
the exact initial bytes passes another check; a real synthetic tool call on
the already loaded thread still succeeds. No fence check was skipped. Native
status enumeration itself re-reads config, so the test captures the catalog
before restoration and uses an actual call afterward to prove retained state.
This is a native read gap, not an inference from a config view.

Per the r3 fallback, Permissions remains refused; the launch/report diagnostic
names this exact path. The candidate stays in research and is not installed
as admission authority. The fixture supplies the input list and supports only
the existing bounded stdio TOML subset; it claims neither complete native
config discovery nor model-turn/approval-dialog qualification. A vendor input
snapshot consumed by the native reader, or an independently enforced policy
boundary over that reader, needs separate qualification before support can
be advertised. An advisory filesystem lock or another preflight cannot bind
a native process that does not participate in that lock.

The manager must still set both `Mode: native-project` and
`NativeProjectEnabled` to inspect a plan. Another mode, release or platform
refuses rendering. `strict` always receives a typed `Unsupported` refusal for
`native-source-exclusion`; it never falls back to native-project. The existing
strict I-COMP admission and released fragment grammar remain unchanged.
The API starts no provider and grants no admission or lease authority.

## Protected inputs and output ownership

The request contains the composition file map and SHA-256/size/mode receipt,
plus the selected profile's separate instruction documents and their receipt.
I-COMP derives `CompositionPreview.ProfileDocuments` and
`ProfileDocumentReceipt` from the same protected profile closure. These are
non-wire preview outputs; no adapter splits or guesses a profile prefix from
concatenated project prose. The strict composition's existing additive
root/system document bytes and receipt remain unchanged.

The renderer preserves the profile documents and admitted skill output bytes.
It excludes the composition's concatenated project root and project system
parts from the native-project home. I-COMP continues to own whole-skill
selection, static skill projection and provider resolution. Unsupported native
outputs refuse. The profile system document retains the adapter registry's
separate `model_instructions_file` channel.

The registry supplies `CODEX_HOME`, the native root and skill paths. A new
managed `config.toml` replaces the seed's `mcp_servers` namespace, including
an explicit empty set, with the composition's MCP union. This manager-side
replacement does not erase repository MCP policy in the native merged config,
which is the launch refusal above. Only the
measured stdio entry encoding is supported; HTTP entries refuse pending native
qualification. The intermediate `curator-mcp.config.toml` is consumed as data,
not published as a second active MCP configuration or selected with `-p`.

Revision A seeds may contain MCP servers. Revision B normally strips them
before this API; if such a table nevertheless reaches the renderer, the same
composition-owned replacement rule applies. Every affected seed server and
plan-owned key is named in a sorted `Note`, without recording values.
Unrelated seed fields remain. Alternate seed profiles and native skill
overrides are dropped with named notes. Seeds selecting alternate instruction,
permission, hook, plugin or agent control sources refuse with the key's name
rather than silently pruning unowned controls.

No repository config or approval file is read by the renderer. Native-project
uses explicit trusted entry to retain native project instructions and config;
the project instruction budget is the native default 32768 bytes. The profile
root and separate system channel remain manager-owned. The manager programs
model, effort, approval/sandbox policy, user approval review, write restrictions
and memory flags through native `-c` pairs. Supported request encodings are
`never` or `on-request` with `read-only` or `workspace-write`.

Paired native startup/config/thread probes cover all four policy combinations.
The no-manager-argv control exposes hostile repository approval, sandbox and
memory settings. Complete manager argv retains the requested approval/sandbox,
no extra writable roots, network off, both temporary-directory exclusions,
and five false memory settings. Repository plugin config and prose remain
visible. MCP permission settings survive those same argv values, so this is
not a certified Permissions capability. The probes do not qualify model-turn,
reload, resume or child behavior.

`Plan.Launch()` returns the mode, required non-isolation notice, `CODEX_HOME`
and native `-c` argument pairs. They are argv values, never shell text. A
consumer must keep them bound to the plan, display the notice, and prevent
later arguments or ambient configuration from overriding manager policy.
No released fragment schema is widened by this adapter API.

## Receipt and publication

`Plan.Receipt()` inventories final output paths as `{SHA256, Size, Mode}`.
The file, receipt, launch and report accessors return copies. The shared entry
temporarily mirrors `projectlease.Surface` in `types.go`; integration replaces
that single candidate definition with the owning import. The candidate
certification includes the published Muse fields and refusal codes; JSON
spelling and the launch mode/notice additions remain an explicit integration
reconciliation. Integration must
bind this final receipt into immutable mode-specific composition and policy
before publication, never swap a receipt under an already published record
or lease. This adapter does not advertise a strict `project-adapter/1` record.

`Plan.PublishPrepared` retains I-LEASE's published
`PublishPrepared(context, leaseID, files, dryRun)` interface, but this pin
refuses before calling the owner. The regression proves zero publisher calls
for refused permissions, invalid inputs and cancellation. It does not claim an
actual lease transaction/liveness or a supported native-project publication.
No lease implementation is imported or copied.

`Plan.Materialize` requires a live manager-home lock witness and a stable
`environments/project-views/<checkout>/<profile>/codex_cli` allocation. The
caller first enforces I-LEASE's idle-home rule. A new home is staged privately
and published by rename. An existing home is only verified, never silently
repaired. `Plan.Verify` checks generated bytes, sizes, modes, linked paths,
extra generated skill files and the complete owned instruction inventory.
`AGENTS.md` must be absent when no profile root was receipted, including empty
and system-only profiles; `AGENTS.override.md` must always be absent. Missing
owned surfaces are classified through the shared `stateread` seam, retaining
unreadable-ancestor failures instead of treating them as absence. Native mutable
state and credentials are outside the receipt and are not read. Cancellation
before publication leaves no partial native home.

The final file map includes the complete pinned native `skills/.system` bundle
before publication: 48 unchanged upstream assets plus the native fingerprint
marker. Every byte, size and private mode belongs to the ordinary immutable
receipt. The bundled archive has a fixed SHA-256 and is never read from a live
native home. Native startup recognizes the complete matching bundle and leaves
it unchanged, so normal first start and restart preserve verification. Missing,
modified, extra or linked bundle files still refuse without repair. Composition
inputs cannot replace the reserved `.system` namespace, including case aliases.
See [bundle provenance and licenses](../internal/projectadapter/codex-system-skills.NOTICE.md).

## Certification boundary

The typed report names native context, admitted skill/MCP delivery and the
programmed memory-off configuration for the pinned macOS release. Permissions
is refused because repository MCP tool restrictions remain effective.
Exclusion remains refused: repository skills, including late-created paths, survive the measured
controls. Lazy, resume and child exclusion remain unqualified. Neither the
preview nor the startup observations certify hostile-code containment.

The [profile native writer adaptation](https://github.com/relux-works/curator-spec/blob/f02b5aef/protocol/mcp-native-writer.md)
limits its ownership protocol to shared profile homes and explicitly leaves
checkout-bound Werkstatt adapters outside that protocol. This API writes only
a fresh project-view home, advertises no profile native writer capability and
edits no shared profile file.

The pinned [source-exclusion qualification](qualification/project-source-exclusion-codex.md)
records measured native controls and their limits. The exact
[Codex configuration schema](https://github.com/openai/codex/blob/rust-v0.162.1/codex-rs/core/config.schema.json)
identifies the native memory and sandbox keys. These mappings do not close the
refused strict runtime boundary.

## Tools and verification

Go and Git exercise composition, synthetic approval and temporary native home
delivery. Default tests use no provider processes. Opt-in tests start the exact
pinned native binary in isolated offline fixtures; they use no credentials,
model turns or network MCP. Keep logs and private caches under task-local
`.temp/` and remove scratch after handover.

```sh
env -u GIT_DIR -u GIT_WORK_TREE go test ./internal/projectadapter ./internal/projectcontext -count=1
env -u GIT_DIR -u GIT_WORK_TREE CURATOR_CODEX_ADAPTER_MUTANTS=1 "$CURATOR_GO" test ./internal/projectadapter -run '^TestCodexNarrowingMutants$' -count=1 -v
env -u GIT_DIR -u GIT_WORK_TREE go vet ./internal/projectadapter ./internal/projectcontext
```

The real-composition test independently assembles the strict preview's profile
and project root documents with a literal LF separator, then requires that the
native-project file equal only the complete profile document. It includes
admitted skill assets and profile MCP while ensuring repository config,
approvals and native instructions are not imported into manager-rendered bytes. Nearby controls exercise empty/nonempty
MCP, seed revisions A/B, both supported policies, profile-only system context,
explicit native-project enablement, strict refusal, the required notice,
receipt drift and publication boundaries. Narrowing plants must fail their
named assertion; a skip, setup failure or compiler error is not evidence.

Rework regressions cover empty/root/system-only instruction inventories,
required absence, native-bundle receipt provenance/reservation/drift, typed
unreadable ancestors, permission refusal and zero lease-publisher calls. The
repository-wide manager-read scanner retains its known-violation control:

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test ./internal/envprofile -count=1 -run '^(TestManagerOwnedAbsenceReadsAreGuarded|TestManagerReadScannerFindsAliasedNotExistCollapse)$'
```

Set `CODEX_EXCL_BINARY` privately to the frozen qualified executable; each run
requires a fresh directory. The opt-in regressions call production Render/Materialize/Verify APIs
with the complete emitted argv. The trust-only control restores repository
instructions and also admits repository plugin config. Paired config/thread
probes verify the measured manager overrides; an actual MCP catalog/call pair
proves the surviving repository `disabled_tools` effect and typed refusal. The lifecycle matrix covers empty/root/system-only profiles,
read-only/workspace-write policy, debug startup and app-server/thread startup,
followed by unchanged restart/reuse without receipt recapture:

```sh
CODEX_EXCL_RUN_ROOT="$PWD/.temp/project-adapter-codex/native-01" \
  env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test ./.research/project-exclusion-codex -count=1 -v -run '^TestAdapter(TrustCapability|TrustedPolicyPrecedence|RepositoryMCPPermissionEffect|EnumeratedMCPReload|NativeBundleLifecycle)$'
```

Follow the [offline harness guide](../.research/project-exclusion-codex/README.md)
for boundary controls and fixture requirements. Native process IDs and sanitized
observations go below the explicit run root. These tests establish startup and
receipt behavior; they do not qualify model turns, resume or child sessions.

The r3 fence unit pairs and real native call controls are replayed explicitly:

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test ./.research/project-exclusion-codex -count=1 -v -run '^TestConfigFence(Admission|Intake)$'
CODEX_EXCL_RUN_ROOT="$PWD/.temp/project-adapter-codex/fence-01" \
  env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test ./.research/project-exclusion-codex -count=1 -v -run '^TestAdapterConfigFenceReadGap$'
```

Run the affected packages and controls through the authorized suite-slot
wrapper. Narrowing plants cover skipping the fence, comparing filenames only,
ignoring a linked parent and omitting the measured refusal path. Each must
reach its named assertion; no skipped native entry counts as evidence.
