# Claude 2.1.305 native-project qualification gaps

TASK-261022-sg5y2x r2. This is an evidence inventory and a proposed research
handoff, not native qualification. No vendor process was run for this change.

## Exact existing evidence

The only executable Claude project-source qualification harness found under
`.research/` is [project-exclusion-claude](project-exclusion-claude/README.md).
Its binary is Claude Code `2.1.305`, macOS arm64, SHA-256
`ea3ea3c08d2a9634a4651e687fcf900d410c489bce2b00f22bd7352b91ba704f`.
The [record](../docs/qualification/project-source-exclusion-claude.md) and
[observations](project-exclusion-claude/observations.json) preserve the bounded
results. The older CIP research notes describe desired policy; they do not
provide executable native qualification for these two obligations.

| Obligation | Existing source/evidence | What it proves | Missing qualification |
| --- | --- | --- | --- |
| Manager permissions | `TestP6SettingsPermissions` in `project-exclusion-claude/replay_test.go`; P6 project/local, excluded/baseline observation rows | With `--setting-sources user`, two repository settings locations stop changing the requested model and one Read deny; an approved home deny still blocks its neighboring Read. The fixture explicitly grants Read via `--allowedTools Read`. | Manager default deny with empty/exact built-in tool sets, MCP deny, conflicting project permission expansions, and native-project discovery enabled. Startup Read atoms are insufficient. |
| Memory off | `setup` sets `CLAUDE_CODE_DISABLE_AUTO_MEMORY=1` for every invocation; `collect.py` retains debug lines containing `project memory is off` | The variable was supplied as a fixture precondition; some loader diagnostics were observed. | A discriminating enabled control and disabled counterpart proving automatic memory load/write behavior. No test toggles the memory variable or asserts that behavior. Debug text alone is insufficient. |

P5's strict MCP matrix includes all-settings-source cases and preserves an
explicit approved stdio server, including explicit empty sets. It does not
qualify MCP tool-call permissions. P7/P8 cover selected safe-origin resume,
background child and instruction/skill reload paths under exclusion; they do
not close native-project permission or memory obligations.

Native-project intentionally permits native project discovery. Reusing the
strict source flag as proof would suppress the project instruction inputs the
new mode is meant to expose. Repository settings may be discovered natively;
their permission requests must not expand manager grants. The exact recipe
needs measurement, rather than assuming home settings override all projects.

## Required permission probes

Use the same exact pinned binary and isolated synthetic HOME/config/project
boundary as the existing harness, with fixed local responses and real native
tool witnesses. Inspect actual requests, native init/results and bounded
fixture effects; model claims, missing requests and sandbox-denied effects do
not establish permission enforcement.

1. Establish native-project's positive discovery control: a unique project
   instruction marker reaches an actual request while profile-only managed
   instructions remain present. The chosen permission recipe must preserve
   that behavior. Test both project and local native settings sources.
2. Program the adapter's actual default policy: `dontAsk`, empty built-in
   `--tools` argument, home manager deny for `mcp__*`, explicit strict MCP file,
   and managed config home. Attempt Read/Write/Bash and a connected synthetic
   MCP tool. Assert the intended native refusal/unavailability and absence of
   forbidden data or file effects. A home deny in a fixture that never had the
   tool available is not a discriminating deny control.
3. Add a narrow manager grant, such as Read, as a nearby positive control;
   assert actual permitted file contents. Retain neighboring manager denies
   and ensure repository/local allow rules cannot expand grants or defeat
   denies. Verify command-line controls, home/project precedence and explicit
   denies where a grant overlaps. Keep MCP tool calls denied while the
   approved synthetic server can connect.
4. Repeat the mandatory policy invariants at every form the adapter proposes
   to admit: fresh entry first, then any proposed resume/fork, reload or child
   route. Refuse forms that lack discriminating evidence; do not promote a
   headless startup result to interactive/child qualification.
5. Narrowing controls must omit the built-in ceiling, weaken the MCP deny and
   plant a conflicting repository expansion. Each must trigger its intended
   violated invariant, with a successful permitted neighbor. Failure caused
   only by the outer sandbox or setup is not successful control evidence.

## Required memory probes

1. Discover and establish the exact version's synthetic automatic-memory
   locations without reading any real native home. Seed a unique memory
   marker in a fresh managed fixture; with automatic memory enabled, prove
   the marker reaches actual native input. Observe a discriminating native
   automatic write if write suppression is part of the claimed contract.
2. Repeat under the fixed adapter environment variable. Assert absence from
   actual outbound requests, no forbidden automatic writes, and continued
   successful ordinary project/profile instruction and permitted work-file
   controls. Explicit operator-authored CLAUDE.md discovery is intentional in
   native-project and must not be mislabeled automatic memory.
3. Cover each memory scope and every launch/resume/child/reload form proposed
   for admission. Previously contaminated history must be refused unless
   separately proven safe; turning memory off does not scrub transcripts.
4. Remove the memory-off variable as a narrowing control and require the
   seeded marker/effect to fail the intended assertion. A control that never
   activates native automatic memory is unqualified, even if the off case is
   silent. Preserve the exact executable identity, source/control identity,
   invocation and sanitized request/effect evidence.

Until those probes qualify the actual recipe and forms, both mandatory
surfaces remain refused. Status and the native-project launch diagnostic carry
`project sources not isolated`; no executable argv/environment is returned.
