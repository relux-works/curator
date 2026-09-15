# Local skill sources, managed outputs, and runtime handling

Date: 2026-09-10.
Related discussion: TASK-260910-16vtxi.
Status: proposed requirements; local-source Skillfile syntax is not implemented or accepted as a normative change.

## Authored files and installation outputs

The current manager contract installs context under `.agents/skills/<name>/`
and command shims under `.agents/bin/`. Codex and Claude adapter skill paths
are `.codex/skills/` and `.claude/skills/`. These are different from the
conventional authored directory `agents/skills/`, without a leading dot.

```text
project/
  Skillfile.json
  agents/skills/review/          authored package
    SKILL.md
    agent-skill.json
    scripts/
  .agents/skills/review/         installed context
  .agents/bin/                  installed command shims
  .codex/skills/review/          adapter-managed entry
  .claude/skills/review/         adapter-managed entry
```

Script runtime files live in the manager's protected runtime store. Compiled
commands target immutable build artifacts. They are not installed back into
the authored skill directory.

## Proposed overlap rules

1. Resolve relative source paths against the declaring Skillfile directory.
   Select packages before evaluating installation overlap; a source alias
   with `path: "."` is not itself an error.
2. Resolve physical paths, symlinks and the filesystem's path equivalence
   rules. Validate again at the write boundary rather than relying only on
   string-prefix comparisons made during planning.
3. Reject selected source packages located within active managed output
   trees. Reject destination writes that would overwrite selected inputs,
   including indirect overlap through symlinks or adapter paths.
4. Exclude generated output trees from source traversal and snapshots before
   reading or copying them. A package selected at the project root can be
   supported only with deterministic input rules that exclude these outputs;
   if effective input and output separation cannot be established, reject it.
5. Apply the same checks to context, runtime, build inputs, dependency sources,
   adapter outputs and generated context artifacts. Existing unmanaged output
   conflicts remain errors; this feature does not grant takeover authority.

Examples:

| Selection | Proposed outcome |
| --- | --- |
| Source root `.`; package `agents/skills/review` | Allowed; authored and installed paths differ. |
| Source root `../shared`; package `skills/review` | Allowed when its effective inputs do not overlap managed outputs. |
| Source `/opt/team-skills`; package `review` | Allowed under the same checks. |
| Package `.agents/skills/review` | Rejected as an installed artifact used as an authored source. |
| Package path resolving through a symlink into the managed destination | Rejected. |
| Package `.` with generated output descendants | Requires explicit snapshot exclusion and disjoint effective inputs; never recursively ingest its own outputs. |

## Full package behavior

Local acquisition must feed the same package validation and installation
pipeline as Git acquisition. Copying only `SKILL.md` does not satisfy P1.

| Package content | Required handling |
| --- | --- |
| `SKILL.md` and eligible references/assets | Existing context projection and audit rules. |
| `agent-skill.json` | Existing schema, capabilities and dependency validation. |
| Declared script commands and `runtime_roots` | Protected runtime snapshot and `.agents/bin` shims; existing execution boundaries remain in force. |
| Supported compiled commands and `build_roots` | Existing closed build driver, toolchain admission, cache and build receipts. |
| Declared skill dependencies and system commands | Existing closure, conflict and readiness validation. |
| Undeclared arbitrary scripts | No new promise that they become commands or execute during installation. |

Under the current context projection, `scripts/` is eligible context only for
packages that export no commands. Runtime and build roots are excluded from
context; build roots are not copied into installed script runtime.

The current script runtime identity is keyed by skill name and resolved Git
commit. A plain local directory has no equivalent commit. The extension needs
a content-based package snapshot identity covering all admitted context,
runtime and build inputs, with unambiguous source-kind separation. A change to
only a runtime script must update the installed runtime even when `SKILL.md`
is unchanged. A directory containing `.git` still means filesystem bytes when
selected as a local path, including admitted dirty and untracked inputs.

Local builds retain the existing compiler/toolchain requirements. Prebuilt
CLI dependency distribution remains a separate proposed feature. Local
acquisition does not authorize package-provided install hooks, generators or
arbitrary build programs.

## Evidence

Inspected curator-spec commit: `d019f0e7179520b5c8dcde321c4fe51e04552f58`.
Inspected adjacent implementation HEAD: `683364ce233df872d6cbb194e0e6b205127f5bca`;
the existing working checkout was read without changing it.

- `protocol/core.md`, sections 3, 3.1 and 4: package, context projection,
  runtime/command manifests and installation execution restrictions.
- `profiles/manager.md`, section 3: context destination, protected runtime,
  command shims and compiled artifact activation.
- `../curator/internal/adapters/adapters.go`: current adapter paths and
  unmanaged output conflict handling.
- `../curator/internal/runtimestore/scripts.go` and `runtimestore.go`:
  staged runtime preparation and skill/commit runtime identity. Existing
  staging guards do not establish the new local-source overlap contract.

Validation: read-only specification and implementation inspection. No local
Skillfile-v2 installation or runtime execution has been performed.
