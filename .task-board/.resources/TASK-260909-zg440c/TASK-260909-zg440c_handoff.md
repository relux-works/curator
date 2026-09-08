# TASK-260909-zg440c — apply-launcher-plan-and-environment-errata

Ready for review. Documentation only; uncommitted candidate over `13b28c9a8916464e7253551808ae9969d6aa0186`.

## Changes

- SPEC §4.4: tagged `vendorplugin.BuildLaunch` delegates model/effort admission to the module and calls `agentic.BuildPlan`; explicit launcher-invoked `providerlimits.Store.AvailabilityFor` read keyed by runtime/model/managed Home. Empty Composition, zero Run, and full process Env input in both modes.
- SPEC §4.5: untracked environment starts at complete filtered Plan.Env. Own values are System.ChildEnv(nil, req), never an inherited-env diff or nil-Env BuildPlan. Override warnings use owned names. Literal collisions win with name-only warnings; inherited allowed names retain destination lookup.
- SPEC §4.6: every composed Argv element is retained in argv_suffix, excluding only the separately stored Binary. Literals contain plugin-owned plus fragment/channel values; no inherited HOME/PATH/secrets copying.
- SPEC §6 and §9: aligned diagnostics and explicit tracked destination unset/PATH residual, with ax filtering unknown. No schema or API expansion.
- README introduction names both module entry points. Implementation status remains partial.

## Exact evidence accepted and inspected

Accepted task attachments: `accepted-A0-evidence.md` (original `TASK-260908-qblycn_a0-verification-findings.md`, §§3.2/3.5, E3/E4), and `accepted-environment-evidence.md` (TASK-260908-1c0fwn, §§2–5). Their probes were NOT rerun. The accepted environment result reports 5/5 probe cases, exit 0; this is prior evidence, not this run's measurement.

Read local tagged source: skill-agents-management v0.5.10 resolves to `12f443d10bc217ca7a48e2edab19c739f441df9c`; pkg/vendorplugin/spawn.go and pkg/providerlimits/verdict.go. The supplied environment attachment labels its tag hash `13d167e2…`, whereas A0 and the local tag resolve to `12f443d1…`; no tag was changed. SPEC pins the entry-point citation to the latter, retaining the environment attachment as accepted semantic evidence, without claiming those hash labels match.

Read curator-spec Decision0013 from exact landed PR48 commit `d019f0e7179520b5c8dcde321c4fe51e04552f58`, D6.3/D6.4 and open question 6. SPEC includes immutable upstream Decision0013/environments links. No external state, auth, runtime homes, model calls, installs, releases or ax execution.

## Validation run directly here

| Command | Exit | Scope |
|---|---:|---|
| make check | 0 | Existing build, fmt-check, vet, tests and race targets; 4/4 listed packages pass both test invocations |
| git diff --check | 0 | Whitespace check, including after prose polish |
| Python scope comparison | 0 | Only README.md/SPEC.md differ; §§4.2–4.3, §§4.7–5, §§7–8 byte-identical to HEAD |

No new prose-mirroring tests. Existing tests validate implemented CLI, fragment and mapping behavior; they do not prove the future composer, destination filtering, or secret non-serialization at runtime. No independent review claimed.

make check output:

```text
go build ./...
go vet ./...
go test ./... -count=1
ok  	github.com/relux-works/curator-agent-launcher/cmd/curator-run	6.519s
ok  	github.com/relux-works/curator-agent-launcher/internal/cli	0.856s
ok  	github.com/relux-works/curator-agent-launcher/internal/fragment	6.868s
ok  	github.com/relux-works/curator-agent-launcher/internal/mapping	1.583s
go test ./... -count=1 -race
ok  	github.com/relux-works/curator-agent-launcher/cmd/curator-run	5.367s
ok  	github.com/relux-works/curator-agent-launcher/internal/cli	2.425s
ok  	github.com/relux-works/curator-agent-launcher/internal/fragment	6.055s
ok  	github.com/relux-works/curator-agent-launcher/internal/mapping	2.765s
```

## Operational notes

The first development transition exited 1 because the task lacked an estimate; estimate 2 and the retried transition succeeded. A guessed source filename pkg/vendorplugin/launch.go was absent; git grep located spawn.go, which was read successfully. An unsupported board resources projection and resource list attempt were not treated as proof of absent evidence. No workflow depended on those failed reads. No directives were recorded at the checkpoint.

Per task-specific prohibition, LOGBOOK/control-root writes were not made; this outcome and the board note serve as the evidence record. Scope, mapping, defaults, version metadata, Pi prompt behavior and ax fields remain unchanged. Parent owns independent Astra-medium review and signed delivery.

Artifacts were written outside the managed worktree before resource attachment. Reviewable diff: TASK-260909-zg440c_documentation.patch.
