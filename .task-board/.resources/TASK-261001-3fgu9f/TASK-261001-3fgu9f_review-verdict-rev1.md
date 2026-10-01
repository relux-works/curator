# TASK-261001-3fgu9f review verdict rev1 (CR-TASK-261001-3fgu9f-1) — CHANGES REQUESTED → to-dev

Reviewed d0920353..7a007577 (21 paths) against launcher-muse-brief.md and the binding scope reduction.

## Blocking
1. **LOGBOOK.md is a new file in the CR** (LOGBOOK.md:1). Producers never create or edit it. Remove it from the CR. Keep the facts in the task results and CHANGELOG.
2. **The agents-management v0.5.22 → v0.5.33 bump is not necessary for the muse rows. Revert it.**
   - Evidence: in a scratch copy of the candidate tree with go.mod pinned back to v0.5.22 and the original go.sum, `go test -p 1 ./...` passes everything except the four claude golden tests (`TestProductionPipelineGoldens/claude_code/tracked={false,true}`, `TestProductionAliasEquivalence/claude-as-claude_code/tracked={false,true}`). Those fail only because the candidate's goldens already carry `CLAUDE_CODE_ENABLE_PROMPT_SUGGESTION`. Every muse test (mapping, v3 reader, defaults, plan.Build refusal, the permission bound, composition) passes at v0.5.22. The v0.5.22 muse plugin and system-only declaration already exist.
   - Revert these:
     - go.mod:6 and go.sum:3-4
     - both goldens (pipeline-claude_code-false.golden:18, pipeline-claude_code-true.golden:29)
     - CHANGELOG.md:8-11 (drop the pin and prompt-suggestion sentences)
     - README.md:14 and :51 (`v0.5.33` → the actual pin, v0.5.22)
     - SPEC.md:386 ("pinned agents-management v0.5.33")
   - The "interactive refusal flips when the pinned plugin declares interactive" and "permission_mode_unsupported" bounds are version-agnostic. Their tests assert at runtime and still hold at v0.5.22.
   - A pin bump may be proposed as a separate, explicitly requested task.

## Ruled OK
3. **Windows vet.** `GOOS=windows go vet ./...` exits 1 on both baseline (d0920353 via git archive) and candidate. Output is 20 lines each. After stripping line:col numbers and sorting, the diff is empty. The candidate introduces no new Windows vet failure; the existing POSIX-only code is the cause.
4. **Scope**
   - Muse mapping row: `internal/mapping/mapping.go` (system muse, ax provider "muse") plus the SPEC §4.2 row and the mapping test row.
   - v3 reader (`fragment.go`):
     - 4 XDG vars must share one absolute parent, with /config /data /state /cache suffixes.
     - Muse is only accepted with v3.
     - system_prompt and mcp are rejected.
     - `TestMuseV3RejectsInvalidFragments` has 14 negatives, including HOME in env.
     - `TestV3OtherAdaptersRetainV2Rules` shows the other adapters keep their v2 rules.
     - v1/v2 are unchanged.
   - HOME is never set: the fragment rejects HOME, and composition preserves the inherited HOME (`TestMuseV3CompositionPreservesHOME`).
   - Bounds: the interactive refusal is asserted with its exact error text and flips on `SupportsMode(Interactive)`. The `permission_mode_unsupported` refusal is also asserted exactly and gated on `PermissionMapping`. `TestMuseV3PermissionTransportThroughRun` pins the --yolo path.
   - `go test -p 1 ./...` on the candidate: every package ok (cmd/curator-run 96s).
5. **Mutants, run by me in a scratch copy (real exit codes)**
   - Mutant A, HOME overlaid in composition for muse → `go test ./cmd/curator-run -run Muse` exit 1 (`TestMuseV3CompositionPreservesHOME`: HOME replaced).
   - Mutant B, v3 dropped from the fragment enum → `go test ./internal/fragment ./cmd/curator-run -run 'V3|Muse'` exit 1 (`TestMuseV3Fragment`, `TestV3OtherAdaptersRetainV2Rules`, `TestMuseV3ThroughRunPermissionBound`, `TestMuseV3PermissionTransportThroughRun`).
   - Stated bound: today the run path refuses muse before composition (permission bound). Mutant A is therefore killed by the composition test, not through the run path.
6. **Hygiene.** Only LOGBOOK.md is stray. I could not check "employer name" mechanically without the forbidden term list; I saw none in the diff.

## Rework scope
Remove LOGBOOK.md and revert the pin bump (go.mod, go.sum, 2 goldens, CHANGELOG/README/SPEC text). Re-run `go test -p 1 ./...` at the v0.5.22 pin. Do not change anything else.