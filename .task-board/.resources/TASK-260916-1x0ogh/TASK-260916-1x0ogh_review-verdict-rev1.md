# TASK-260916-1x0ogh — review verdict, revision 1

Verdict: **changes_requested**, route to `to-dev`. Candidate was reviewed read-only. No acceptance, commit acknowledgement, candidate edit, or integration was performed.

## Candidate and artifact identity

The runtime CR `CR-TASK-260916-1x0ogh-1` has an empty curator repository delta. This is appropriate for the repository boundary: the deliverable is in curator-spec, not curator implementation. It is not evidence that the leaf delivered nothing, nor proof of acceptance. The reviewed spec worktree is `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260916-2otjbn/worktree`, HEAD and origin/main `07e2b41892bf24c33b859072b2b920fe2d0a17fe`.

The complete diff, including untracked new files through an alternate temporary Git index (`GIT_INDEX_FILE`, `read-tree HEAD`, `add -N .`, `diff origin/main`), is byte-identical to the attached `TASK-260916-1x0ogh_spec-patch_rev1.patch`:

- SHA-256 both: `5d45ea7a42c06f94a88c7330da092e2ceaab7210fd5fe50c6cbf3e11d55dcce4`.
- `git patch-id --stable` both: `d01f170f9ab802d6a7af6ef2eec0b1e23360a704`.

The actual candidate index was not changed. A disposable byte copy was initialized with its own temporary Git baseline solely for regeneration and mutation checks. The producer evidence and patch, supplied campaign/producer/reviewer briefs, and audit E4 plus Appendix B were read. `task-board spawn goal "$TASK_BOARD_RUN_ID"` reported `Active Goal: none (run is not goal-bound)`.

## Required corrections

### F1 — High: semantic vector enforcement is absent

`tools/validate.py:4644–4688` checks field presence and outcome shape but never derives resolution from `install_dir`, `provider_directories`, `path_entries`, `present`, `published_dirs`, or `managed_dirs`. Its production entry point does call the checker (`main`, line 4710 vicinity), so this is an inadequate check, not a missing caller.

Reproduction in the disposable copy: change only `s6-planted-path-provider-warns-then-refuses.revision_b` to `{"resolved":"/home/operator/work/acme/.bin/curator-run","diagnostic":null}`. Refresh the vector manifest digest and the two candidate manifest pins so integrity checks remain valid. Run the unchanged `python3 tools/validate.py` entry point. It prints `validated 60 schemas and 1054 vector files` and exits **0**, incorrectly accepting the exact E4 behavior revision B must refuse. Measured adversarial bound: **0/1 semantic mutants rejected**; this does not measure manager implementation behavior.

Correction: independently compute expected outcomes from each case's inputs and compare both profiles, including precedence, executable filtering, managed/published refusal, missing/untrusted distinction, and migration directory. Add a regression that rejects the demonstrated silent-success mutant (not merely a warning diagnostic under B or a resolve-and-refuse contradiction). Extend semantic checks and negative evidence to the rules actually claimed. Refusal outcomes should also carry/check the path and consulted roots required by the prose rather than only `resolved: null`.

### F2 — High: revision A changes selection before the promised flip

`protocol/environments.md:2213–2222` says “Resolution searches the trust roots first” and consults PATH only without a root match. The settled decision says revision A **keeps resolving on PATH**, warns when that selected executable is outside roots, and preserves old behavior until revision B. A provider in `/opt/curator/providers` now silently beats a different `/usr/local/bin/curator-run` on PATH; the existing `listed-directory-provider-resolved` vector encodes this changed selection under A. An install-directory-only provider also becomes executable under A despite not being discoverable under the previous PATH convention.

Correction: revision A must retain ambient PATH resolution and apply the warning/refusal rules to that selected candidate; revision B performs install-directory-first, then configured-list resolution. Update both-profile vectors and CHANGELOG consistently. Add the case where PATH selects a different provider while a trusted provider exists. Align `cli/curator.md:123` with the two profiles or explicitly reference their distinction: its unconditional “on PATH” description cannot describe B, and is already inaccurate for the candidate's current A. The producer's claim that this CLI text remains accurate under A is false for the candidate.

### F3 — Medium: missing and untrusted overlap under revision B

`protocol/environments.md:2234` assigns `subcommand_provider_missing` whenever no executable exists in trust roots under B. But `:2223–2226` requires `subcommand_provider_untrusted` for a PATH-only provider, which also has no executable in trust roots. The table therefore assigns two outcomes to the same input. Removing the PATH “fallback” also leaves unclear whether a diagnostic-only PATH probe remains.

Correction: define disjoint lookup/diagnostic conditions: trusted candidate, forbidden candidate, PATH-only candidate, true absence, and read failure. Explicitly distinguish a diagnostic-only PATH probe from permitted dispatch, and ensure the missing row excludes PATH-only matches. Keep the diagnostic set closed and consistent with the vectors.

### F4 — Medium: unreadable is treated as absence and activates fallback

`protocol/environments.md:2174–2175` says an entry that “does not name a readable directory contributes no candidate.” The `unreadable-listed-directory-contributes-nothing` vector consequently executes the PATH provider with a warning under A. This merges a failed read with absence, contrary to the supplied Evidence That Counts rule and the same document's `:2280–2281` (“unreadable evidence is reported as unreadable, never as absence”). A failed read cannot prove the trusted provider is absent.

Correction: distinguish nonexistent directory/provider from unreadable or otherwise failed lookup; do not activate absence fallback from a read failure. Specify and test the failure outcome and status posture explicitly, admitting any necessary diagnostic into all closed surfaces. This is ordinary spec rework, not an external blocker or human-only design decision.

## Brief conformance, item by item

| Requirement | Evidence (candidate file:line; exact excerpts) | Result |
|---|---|---|
| Install directory after symlinks; explicit ordered list | `protocol/environments.md:2167–2173`: “resolved after symlinks”; “in listed order”; “The first executable regular file ... a root wins” | Present for enforcing profile |
| Preserve identifier, builtins, no registry/install, data isolation | `protocol/environments.md:2156–2166`: “MUST match the core §2 identifier grammar”; “An implemented subcommand always wins”; “no provider registry”; “nothing is downloaded or installed implicitly”; “MUST NOT influence” | Present |
| PATH exclusion and S6 composition named | `protocol/environments.md:2180`: “The ambient `PATH` is not a trust root”; `:2196`: “The attack this rule closes is the S6-injected `PATH` (finding E4)” | Present |
| Published/managed refusal wins | `protocol/environments.md:2183–2188`: “MUST NOT reside”; “the refusal outranks both trust roots” | Present; semantic gate incomplete (F1) |
| Resolved path and status posture | `protocol/environments.md:2192`: “Every dispatch reports the resolved absolute provider path”; §12: “absolute provider path and trust verdict”; `:2282`: “provider row is non-current when the active revision refuses that provider” | Present in prose; refusal-vector details not verified (F1) |
| Two labelled rollout steps and migration hint | `protocol/environments.md:2213`: “Revision A (warning release)”; `:2223`: “Revision B (flip release)”; `:2219`: “list the provider's directory in `provider_directories`” | Labels/hint present; A violates settled behavior (F2) |
| Diagnostics closed and exact | `protocol/environments.md:2234–2236` lists exactly `subcommand_provider_missing`, `subcommand_provider_untrusted`, `subcommand_provider_outside_trust_roots`; validator uses identical spellings | Names agree; conditions overlap (F3) |
| Knob/default/lockability | `protocol/environments.md:2332`: `provider_directories`, “list of absolute paths”, `[]`, section 11; `:2353`: “a locked provider list is fleet policy”; `profiles/manager.md:57`: `environments.provider_directories` | Present and consistent with schemas |
| Machine/system schema | `schemas/v1/manager-config-v2.schema.json:501` defines `provider_directories`, absolute POSIX/drive paths, unique array, default `[]`; `schemas/v1/system-config-v2.schema.json:24,51` extends lock enum and references the same definition | Present; 6 new cases (4 manager, 2 system) plus existing system valid fixture |
| Conformance surfaces, vectors and manifest | `protocol/environments.md:2399` names `vectors/umbrella-provider-resolution.json`; manifest line 4216 registers it. Opened all 10 cases × 2 profiles: install; install before list; listed provider; listed order; S6 PATH-only; non-executable; published; listed-but-published; unreadable; missing | Cases encode outcomes; semantic gate fails F1; A and unreadable outcomes need F2/F4 corrections |
| CHANGELOG E4, rollout and 0016 | `CHANGELOG.md:99`: “E4: umbrella provider resolution”; entry names “revision A”, “revision B” and “Blocks proposal 0016 / `path_prepend`” | Present; revise A description with F2 |
| Scope and faithful regeneration | Only spec/docs, schemas, vectors, generator/validator/test tooling, manifest and candidate release manifest pins changed; no curator/launcher implementation or proposal content changed | Acceptable tooling scope per review brief; no historical release metadata edits |
| Existing schema meanings preserved | Parsed old/new JSON for all 88 modified existing machine/system schema cases, removed only `environments.provider_directories` and its lock-list member from candidate, compared with origin/main | **88/88 identical**, no other semantic changes |
| Patch/evidence attachment | Producer task-scoped evidence and spec patch retrieved; patch identity verified above | Pass |
| Merge/integration AC | Spec worktree is uncommitted; CR lives in separate curator repo | Not landed, and not accepted due to findings; integration remains producer-owned |

## Independent validation

Shell `/bin/zsh`, `set -o pipefail`; candidate cwd was the exact curator-spec Story worktree. Required command:

```sh
PATH="/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin:$PATH" make validate
```

Independent result: **make validate exit 0**. All three gates completed; no producer result substituted. The Go gate reported its own cached result.

```text
python3 tools/validate.py
validated 60 schemas and 1054 vector files
python3 -B -m unittest discover -s tools -p 'test_*.py'
.........................................................................................................................................................................................................................................
----------------------------------------------------------------------
Ran 233 tests in 525.962s

OK
go test ./tools/...
ok  	github.com/relux-works/curator-spec/tools/generate-vectors	(cached)
EXIT_CODE=0
```

The execution service temporarily stalled during this run; it recovered and the command completed before the per-call bound. Python tests took 525.962 seconds. No unfinished command is counted as passed.

Final candidate identity recheck: **111/111 changed paths** still match the reviewed disposable-copy baseline; zero mismatches.

Regeneration: `make regenerate-check` in the disposable byte copy, against its temporary baseline of the exact candidate:

```text
go run ./tools/generate-vectors -root .
git diff --exit-code -- conformance/v1 release/1.0.0-rc.5.json release/1.0.0-rc.6.json release/1.0.0-rc.7.json release/1.0.0-rc.8.json release/1.0.0-rc.9.json
EXIT_CODE=0
```

Adversarial check through the real validator entry point, in that copy after regeneration:

```text
MUTANT: S6 PATH-only provider resolves without diagnostic under revision B; integrity pins refreshed, checker unchanged.
validated 60 schemas and 1054 vector files
EXIT_CODE=0
```

This passing mutant is failed security-conformance evidence. The test corpus checks syntax/inventory but does not establish provider-resolution semantics. Runtime implementation is deliberately out of scope and was not tested. No claim is made about actual manager or launcher execution.

Important findings are persisted in this task-scoped verdict and board notes; no LOGBOOK.md edit was made because campaign rules explicitly forbid it. Next producer should correct F1–F4, regenerate, rerun validation and the semantic mutants, attach revised spec patch/evidence, and hand off for another independent review.
