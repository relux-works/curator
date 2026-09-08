# E5 Pi prompt semantics — review handoff

Task: TASK-260909-1hznm7 (correct-launcher-pi-prompt-precedence).
Base: 18aeaed9af7dc5ffbe6cc79a4731a852fbb716da, after plan/environment PR7; `git ls-remote origin refs/heads/main` returned this same OID. Working tree was clean before this task. Only SPEC.md is modified, uncommitted.

## Changes and semantic review

- §5/§5.1: native same-semantics flags suppress discovery, rather than double-apply with files. Trusted existing project files win over agent-home files when discovery applies.
- §5.1/§5.2: mandatory existing managed-home probes and unreadable refusals remain; warnings distinguish observed candidates, suppression, and applied launcher channels. An absent home file says nothing about other native inputs. Native arguments remain uninspected.
- §6: adjacent absence wording is bounded to the probe; no diagnostic changes.
- §9: native/hand launch, timing, and unprobed project-file residuals are stated. Historical additive wording is explicitly marked corrected, without a version bump.
- Version 0.3.0-draft, mapping, defaults/plan/API policies, channel registry, no-home-writes, and runtime source are unchanged. No new probe, replace descriptor, tests, permission bypass or ax operation.
- README remains accurate: system-prompt application is not implemented. Existing tests validate existing runtime only; source review, not prose tests, supports the semantic correction.

## Evidence

Accepted A0: TASK-260908-qblycn, attached accepted-A0-evidence.md §4.3 and §6 E5 (accepted evidence, not re-executed native runtime probes).
Curator-spec: `git show d019f0e:protocol/environments.md`, revision 1.1 §5.5/§7.3, read locally and linked from SPEC.
Installed package.json reports @earendil-works/pi-coding-agent version 0.84.2, independently re-read this run. Installed dist/core/resource-loader.js was read directly; excerpts below. No native runtime launched, no homes modified.

```javascript
380:         const systemPromptSource = this.systemPromptSource ?? this.discoverSystemPromptFile();
381:         const baseSystemPrompt = resolvePromptInput(systemPromptSource, "system prompt");
382:         this.systemPrompt = this.systemPromptOverride ? this.systemPromptOverride(baseSystemPrompt) : baseSystemPrompt;
383:         this.systemPromptSourcePath =
384:             systemPromptSource && existsSync(systemPromptSource) ? resolvePath(systemPromptSource) : undefined;
385:         let appendSources = this.appendSystemPromptSource;
386:         if (!appendSources) {
387:             const discoveredAppendSystemPromptFile = this.discoverAppendSystemPromptFile();
388:             appendSources = discoveredAppendSystemPromptFile ? [discoveredAppendSystemPromptFile] : [];
389:         }
808:     discoverSystemPromptFile() {
809:         const projectPath = join(this.cwd, CONFIG_DIR_NAME, "SYSTEM.md");
810:         if (this.settingsManager.isProjectTrusted() && existsSync(projectPath)) {
811:             return projectPath;
812:         }
813:         const globalPath = join(this.agentDir, "SYSTEM.md");
814:         if (existsSync(globalPath)) {
815:             return globalPath;
816:         }
817:         return undefined;
818:     }
819:     discoverAppendSystemPromptFile() {
820:         const projectPath = join(this.cwd, CONFIG_DIR_NAME, "APPEND_SYSTEM.md");
821:         if (this.settingsManager.isProjectTrusted() && existsSync(projectPath)) {
822:             return projectPath;
823:         }
824:         const globalPath = join(this.agentDir, "APPEND_SYSTEM.md");
825:         if (existsSync(globalPath)) {
826:             return globalPath;
827:         }
828:         return undefined;
829:     }
```

## Validation and operational record

- `make check` executed directly, redirected to a log, no pipe: exit 0. Includes build, fmt-check, vet, test and race. Full output below.
- `git diff --check` executed directly after the last prose edits: exit 0.
- Initial requested set_status: exit 1, estimate required. Set estimate Fibonacci 1: exit 0; repeated set_status development: exit 0.
- An initial read used the wrong curator-spec path docs/environments.md (git fatal); corrected immediately to protocol/environments.md, exit 0. No evidence inferred from the failed read.
- Logbook is intentionally not written: task-specific instructions prohibit private-record/LOGBOOK/control-root writes. Findings and this anomaly are recorded here and in board notes instead.
- No CI, commits, installs, daemon restart, tags/releases, runtime-home or control-root writes were performed. Parent owns delivery.

```text
go build ./...
go vet ./...
go test ./... -count=1
ok  	github.com/relux-works/curator-agent-launcher/cmd/curator-run	3.327s
ok  	github.com/relux-works/curator-agent-launcher/internal/cli	0.466s
ok  	github.com/relux-works/curator-agent-launcher/internal/fragment	3.448s
ok  	github.com/relux-works/curator-agent-launcher/internal/mapping	1.512s
go test ./... -count=1 -race
ok  	github.com/relux-works/curator-agent-launcher/cmd/curator-run	3.423s
ok  	github.com/relux-works/curator-agent-launcher/internal/cli	1.710s
ok  	github.com/relux-works/curator-agent-launcher/internal/fragment	3.586s
ok  	github.com/relux-works/curator-agent-launcher/internal/mapping	2.404s
```
