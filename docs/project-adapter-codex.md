# Codex composition delivery

`internal/projectadapter.RenderCodex` prepares a Codex 0.162.1 Werkstatt home
for the explicitly enabled `native-project` mode. The managed instruction
file contains the selected **profile root only**. Codex reads project
instructions and repository skills natively; Curator delivers only the
approved project skills and MCP from the admitted composition. Both the
certification report and launch contribution say **project sources not
isolated**. This mode does not claim a source-exclusion boundary.

The manager must set both `Mode: native-project` and `NativeProjectEnabled`.
A missing enablement, another mode, another release or another platform
refuses. `strict` always receives a typed `Unsupported` refusal for
`native-source-exclusion`. It never falls back to native-project. The existing
strict I-COMP admission and released fragment grammar remain unchanged.
These mode names implement the explicit wave-1 owner decision; the separate
specification/runner integration owns their eventual wire dispatch.

The API starts no provider, acquires no lease and grants no project authority.
`Plan.RequireLaunch()` accepts only the explicitly enabled native-project
capability. The delegated consumer must additionally enforce current admission,
manager permission ceilings, immutable composition binding and an active lease.
`Plan.RequireStrict()` always refuses at this vendor pin.

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
managed `config.toml` owns the complete `mcp_servers` namespace, including an
explicit empty set, and contains the composition's MCP union. Only the
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

No repository config or approval file is read by this API. Explicit native
project distrust suppresses the measured `.codex/config.toml` settings surface;
it does not suppress native project instructions or repository skills. The
project instruction budget is the native default 32768 bytes, not the strict
zero-cap control. The manager owns model, effort, approval/sandbox policy,
user approval review, write restrictions and memory settings. Supported
policies are `never` or `on-request` with `read-only` or `workspace-write`.
Workspace writes add no writable roots, network access or temporary-directory
grants. Other policy encodings refuse; project restrictions never become
manager grants. Memory generation, memory use, dedicated memory tools and both
memory feature switches are off.

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

`Plan.PublishPrepared` delegates a defensive copy of final bytes through
I-LEASE's published `PublishPrepared(context, leaseID, files, dryRun)` interface.
The owner must check current authority and the bound final receipt; owner
refusals propagate unchanged. No lease implementation is imported or copied.
The interface test proves delegation, not actual lease transaction/liveness.

`Plan.Materialize` requires a live manager-home lock witness and a stable
`environments/project-views/<checkout>/<profile>/codex_cli` allocation. The
caller first enforces I-LEASE's idle-home rule. A new home is staged privately
and published by rename. An existing home is only verified, never silently
repaired. `Plan.Verify` checks generated bytes, sizes, modes, linked paths,
extra generated skill files and home instruction overrides. Native mutable
state and credentials are outside the receipt and are not read. Cancellation
before publication leaves no partial native home.

## Certification boundary

The typed report names profile context, admitted skills, MCP, manager policy
and memory-off delivery for the pinned macOS release. Exclusion remains
refused: repository skills, including late-created paths, survive the measured
controls. Lazy, resume and child exclusion remain unqualified. Native-project
acceptance does not change those facts or certify hostile-code containment.

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
delivery. Tests use no credentials, model turns, provider processes or network
MCP. Keep logs and private caches under task-local `.temp/` and remove scratch
after handover.

```sh
env -u GIT_DIR -u GIT_WORK_TREE go test ./internal/projectadapter ./internal/projectcontext -count=1
env -u GIT_DIR -u GIT_WORK_TREE CURATOR_CODEX_ADAPTER_MUTANTS=1 "$CURATOR_GO" test ./internal/projectadapter -run '^TestCodexNarrowingMutants$' -count=1 -v
env -u GIT_DIR -u GIT_WORK_TREE go vet ./internal/projectadapter ./internal/projectcontext
```

The real-composition test independently assembles the strict preview's profile
and project root documents with a literal LF separator, then requires that the
native-project file equal only the complete profile document. It includes
admitted skill assets and profile MCP alongside forbidden repository config,
approvals and unapproved instructions. Nearby controls exercise empty/nonempty
MCP, seed revisions A/B, both supported policies, profile-only system context,
explicit native-project enablement, strict refusal, the required notice,
receipt drift and publication boundaries. Narrowing plants must fail their
named assertion; a skip, setup failure or compiler error is not evidence.
