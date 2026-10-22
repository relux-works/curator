# Offline Muse composition adapter

`internal/projectadapter` supplies manager-side offline preparation, an exact
path-to-`{SHA256, Size, Mode}` receipt, typed certification reports and a refusing
launch contribution. It enables no CLI launch, released fragment revision or
native adapter capability. The common type definitions in `types.go` are marked
integration candidates for the other project adapters and I-LEASE.

## Offline preparation

After resolving retained protected I-COMP data, pass its complete `Files`,
`Generated`, policy bytes/digest and `Memory: "off"` to `MaterializeMuseOffline`.
The caller supplies a new allocation below an existing private manager parent.
Existing allocations, including links, refuse without replacement. There is no
live-home update path; I-LEASE owns future live publication and session admission.
Preparation creates `config/`, `data/`, `state/` and `cache/` using the existing
Muse envregistry layout. Its returned parent map describes offline paths only.
It is never an admitted environment map or an `envfragment.Fragment`.

Only `config/muse/AGENTS.md`, complete approved skill entries under
`data/muse/skills/<name>/`, and inert approved artifacts under
`.agent-context/commands/<name>` may be copied. No dispatcher is installed.
Bytes, size and logical private mode
must match the entire supplied receipt. The strict additive document is copied
verbatim: profile first, then project documents with I-COMP's one additional LF.
No repository/native-home discovery, configuration merge, approval lookup,
credential access, endpoint probe or provider execution occurs here. Policy bytes
are retained verbatim as inert manager metadata, never encoded into guessed
Muse settings. Memory-off describes the manager plan, not proven native behavior.

MCP inputs must already have protected approval and optionality. Every required
Muse MCP item refuses `mcp_contribution_policy_refused`, reason
`native_writer_unqualified`, before allocation. Approved optional items receive
named `unsupported` omissions; no native writer path is invented. The pinned
[adaptation sheet](https://github.com/relux-works/curator-spec/blob/f02b5aeff07f81ee7d39ca4d2bc9aba1f8e3f816/protocol/mcp-native-writer.md)
certifies no Muse writable MCP surface.

`VerifyMuseOffline` checks the exact offline file set, regular-file types,
hashes, sizes and modes without changing the receipt. It rejects links and
extra/missing files. This is an offline inventory checker, not a verifier for
native mutable state or a concurrent live-home transaction. A process interrupted
while copying can leave an incomplete inert allocation; it grants no session
authority and must not be consumed as a successful result.

## Certification and launch

`MuseCertification(version, platform, mode, form)` separates the composition
mode (`strict` or `native-project`) from the launch form (`exec`, `interactive`,
`resume`). It reports no fully certified strict surface. Its exact known tuple
is Muse `1.4.11-R6667.1` on macOS; unknown tuples remain refused.

The [project-source evidence](qualification/project-source-exclusion-muse.md)
does not qualify complete native exclusion, manager permission enforcement or
memory-off. The [exec foreign-context observations](qualification/muse-1.4.11-R6667.1.md)
do not substitute for those guarantees. Interactive and resume refuse first;
strict exec refuses native-source-exclusion. Native-project exec also refuses
because manager permissions and memory-off remain mandatory and unqualified.
Native-project status and every form's refusal say `project sources not isolated`.
Every failed `MuseLaunch` returns a typed `Unsupported` and no env or argv;
there is no native fallback or automatic downgrade.

The owner-selected native-project mode requires a profile-only managed document,
with native repository instructions reported separately. That mode-aware I-COMP
receipt is a separate follow-up. This offline copier accepts strict output only;
it never strips project bytes or rebaselines the existing composition identity.

## Validation

Use the Go test commands in [README](../README.md#tools). The real-composition
fixture retains protected project snapshots and profile skills, changes the
repository afterward, then compares every copied byte and receipt entry. Its
synthetic composition contract tests rendering, not native certification.
Unit tests pair valid offline copying with policy/receipt/path/MCP refusals,
drift detection and empty refused launch contributions. Opt-in Go-overlay
controls plant forbidden root order, repository approval import, interactive
launch and other admission/receipt violations. Compile/setup failures never
count as caught controls. No test invokes the Muse vendor binary.
