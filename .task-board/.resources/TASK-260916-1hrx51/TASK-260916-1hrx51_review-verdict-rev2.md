# TASK-260916-1hrx51 — independent review, revision 2

Verdict: **accepted** for CR-TASK-260916-1hrx51-2 revision 2. R1 and R2 are closed; no blocking findings remain. Acceptance routes to `integrating`, not `done`; external spec landing is still pending.

## Candidate identity and scope

Reviewed the curator-spec Story worktree `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260916-2d9coh/worktree`, based on `07e2b41892bf24c33b859072b2b920fe2d0a17fe`. The attached `TASK-260916-1hrx51_spec-patch_rev2.patch` is byte-for-byte equal to `git diff HEAD`: 223600 bytes; stable patch ID `b9bff40b930a859b540f85b6a02e2fcec4796805`. All new files already have intent-to-add entries; review did not alter the index.

`origin/main` and `main` now point to `90d50c6` (S6), so their diff is deliberately not the candidate comparison. Against current origin/main the patch ID is `f7261187eeb77c22895ed4e3dcf804e990d0b6c4`; this is the expected base movement called out in the review brief. The evidence's phrase “against origin/main” describes the old base imprecisely; the recorded base and exact patch match establish the reviewed tree without including a reversal of S6. Rebase/landing must preserve S6.

The curator-repository runtime CR has an **empty repository delta**. This is the right outcome for this leaf: it delivers the external curator-spec revision and its patch/evidence, while curator implementation belongs to TASK-260916-55g9dg. Acceptance covers the actual external spec candidate and attached patch, not an assertion that the empty curator patch implements admission or proves external landing.

Compared both attached revisions. Rev1 patch ID is `d0b6f56569cb506c7dfda0ba7fd7c7af02858e69`. Rev2 differs only in `profiles/manager.md` (R1) and `protocol/environments.md` (R2). All schemas, vectors, expected bytes, generators, validator tests, CHANGELOG and generated pins are unchanged from rev1. The full candidate has 117 changed paths, all spec documents, schemas, conformance artifacts, conformance tooling or generated release manifest pins. No product implementation, tags, release publication, or proposals 0014–0018 changed. The Go/Python files are the repository's required vector-generation/validation tooling, not curator implementation.

Read the supplied campaign rules, producer/review/rework briefs, producer evidence and both patches, previous verdict, and audit E2 (`docs/security-audit-2026-09.md:247`) plus Appendix B (`:414`). No candidate files, branch, commits or index were modified by review.

## Per-item review

Line references below are in the candidate tree.

| Item | Exact text / evidence | Result |
| --- | --- | --- |
| R1 manager closed lock set | `profiles/manager.md:56`: “`environments.require_current_profile`, `environments.isolation`, and `environments.transitive_system_modules`”; :58 “No other `environments` knob is lockable or carried by the system file.”; :63–65 “only in the direction of `error`, and schema 2 admits no other value there” | Closed. Exact key matches environments §12.2 and system schema. |
| R2 admitted-only output | `protocol/environments.md:516–518`: “The section 5.5 system-prompt output and the section 10.2 fragment `system_prompt` section MUST contain only admitted system modules.” | Closed. |
| Direct/transitive definition | `protocol/environments.md:510–516`: “root itself”, “active overlay”, “root's or an active overlay's `requires.contexts` entry”; other members “reached only through another package's `requires`” are transitive; waiver “naming the package” admits it | Matches settled decisions. §2 package shape remains appropriately unchanged. |
| R2 drop obligations | `protocol/environments.md:521–525`: “MUST be skipped at materialization”, “manager MUST emit the warning `context_system_module_dropped` naming the package and the module path”; bytes “MUST be exactly the admitted modules' bytes”; resolution/install/update “MUST NOT fail for admission” | Closed; drop warning registered in §5.7 at :843. |
| R2 strict error and lock | `protocol/environments.md:526–530`: “resolution ... MUST fail” with `context_system_module_transitive`; “manager MUST NOT write or change the lock”; install “MUST fail”; update “MUST leave the old lock in place” | Closed; diagnostic in §3.1 :545 and §5.7 :844. |
| Always-warn finding | `protocol/environments.md:532–535`: finding “MUST report every `class: system` module of every member at install and update, admitted or not” | Preserves `context-system-module-present` and its existing always-warn posture. |
| System-prompt and fragment | `protocol/environments.md:775`: “admitted applicable system modules”; :798–801 fragment presence “follows the same admitted set” and drives `works.relux.curator.system-modules`, so “`ax` resume still refuses on drift”; §10.2 :2140 says “admitted” | Consistent. Launcher runtime is out of scope. |
| Knobs/defaults | `protocol/environments.md:2315`: `transitive_system_modules`, `drop`, `error`, default `drop`; :2316: `system_module_waivers`, “list of `{ package, reason }`”, “empty”; :2327–2330 package is a “portable identifier” naming a lock member, absent member “has no effect” | Matches manager schema :494/:501 and closed waiver object :307. |
| R2 lock direction | `protocol/environments.md:2353–2356`: “a system file MUST lock `transitive_system_modules` only to `error`”, “`system_module_waivers` MUST NOT be lockable” | Closed. System schema :23/:51 explicitly admits the key and error-only value; waiver omitted from closed properties/locked enum. |
| Rollout | `protocol/environments.md:803–808`: “default `drop` policy is non-breaking”, “no warn-first split applies”, “`error` is opt-in strictness” | Explicit E2 exemption in settled brief applies; two rollout profiles are not required. |
| Status posture | `protocol/environments.md:2258–2260`: effective `transitive_system_modules` with “every dropped system module by package and path”; :2274 drop warning never makes a row non-current | Mirrored in `profiles/manager.md:2520–2522,2535`. |
| Conformance contract | `protocol/environments.md:2389–2397`: direct, transitive drop with exact bytes, error refusal, waiver, overlay-direct, knob schemas and error-direction case | All 5/5 admission vectors inspected; schema coverage below. |
| Manifest | `conformance/v1/manifest.json:140,144,148,152` names the four expected files; :4160 binds `vectors/environments.json`; :3660–3800 and :4080–4084 bind new schema cases | Registered and independently regenerated exactly. |
| CHANGELOG | `CHANGELOG.md:36`: “E2 direct-only `class: system` modules”; :49–50 “no warn-first split applies” | Rule, knobs, diagnostics, posture and rollout exemption present. |
| Evidence count | Producer evidence now says “seven new schema-case files” plus “four” expected byte directories | Correct; independently counted/read all seven cases and four files. |
| Spellings / closed sets | Both diagnostics and both knobs agree across prose/tables/schema/vectors; lock prefix `environments.transitive_system_modules` agrees in manager §1 and system schema. No corresponding knob rows exist in `cli/`. | No widened/open-ended set or unrelated frozen surface change found. |

## Conformance evidence and limits

Independently opened all 5/5 new materialization cases in `conformance/v1/vectors/environments.json` (:1066,1182,1310,1439,1576): direct admits sysmid/sysroot; drop skips sysleaf with a named warning; error refuses sysleaf/90-system.md and writes no file; waiver admits sysleaf; an active overlay's requires edge admits sysleaf as direct. This exercises the branches rather than merely restating the default.

Direct and transitive-drop files are each exactly 40 bytes, byte-equal, SHA-256 `537ebd5eed3dcaa2dec64e79ccbbdc325cdb35e49652250494ea873ba46d4a3a`. Waived and overlay-direct files are 61 bytes and include the leaf; SHA-256 `056590e92faa33f9c9b4b24513b5e1e79bc57c66a37ac282ee6fdf06bf797449`.

Read 7/7 added schema cases: valid waiver; invalid mode; invalid package grammar; missing reason; unknown waiver field; forbidden system drop direction; invalid system mode. Their registration and exact regenerated contents pass. Rev1 schema-preservation measurements (911 existing index entries and 88 modified fixtures) are accepted as attached review evidence, with zero schema/vector/generator delta in rev2 established independently; those specific counts were not separately recomputed this round.

The actual spec gate calls the independent Python admission computation (`tools/validate.py:4290` onward) from `validate_environment_vectors` (:4433 onward), checks admitted/dropped records, named strict refusal, warning inventory and exact output bytes. Negative tests at `tools/test_validate.py:1403–1420` reject emptied dropped records, missing strict refusal and false refusal under drop; :2117–2123 reject enum/default drift; :2248 rejects the broadened system direction. Go tests exercise direct-vs-transitive boundaries and waiver bytes. Full suite rerun is below.

Bound: this is spec conformance, not runtime resolution/install/update or launcher testing. Runtime lock-preservation and fragment behavior are normative obligations for TASK-260916-55g9dg, not demonstrated runtime properties here. The existing automated consistency gate still does not parse manager §1's second closed enumeration. The latest rework brief explicitly freezes tooling and requests the two document changes; this review manually checked that enumeration and its error-only direction. That automation blind spot is non-blocking for this focused revision.

## Independent validation

Shell `/bin/zsh`, `set -o pipefail`, candidate worktree above. Exact command:

```sh
PATH="/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin:$PATH" make validate
```

```text
python3 tools/validate.py
validated 60 schemas and 1058 vector files
python3 -B -m unittest discover -s tools -p 'test_*.py'
...........................................................................................................................................................................................................................................
----------------------------------------------------------------------
Ran 235 tests in 213.908s

OK
go test ./tools/...
ok  	github.com/relux-works/curator-spec/tools/generate-vectors	0.662s
EXIT=0
```

Full `make validate` exited 0: schema/vector validation, 235 Python tests, and Go conformance tooling tests passed. These commands were independently rerun, not accepted from producer evidence.

`make regenerate-check` ran in a disposable byte copy of every tracked/untracked nonignored candidate file with its own temporary Git baseline (no commit or generation in either project worktree):

```text
go run ./tools/generate-vectors -root .
git diff --exit-code -- conformance/v1 release/1.0.0-rc.5.json release/1.0.0-rc.6.json release/1.0.0-rc.7.json release/1.0.0-rc.8.json release/1.0.0-rc.9.json

EXIT=0
BYTE_DIFFERENCES=[]
```

Zero byte differences across candidate files after regeneration. `git diff --check HEAD` also exited 0. Running this generator in the live candidate would mutate files and its diff-to-HEAD check would reject the uncommitted candidate by construction, so the isolated baseline tests precisely generation drift.

## Lifecycle / record

`task-board spawn goal "$TASK_BOARD_RUN_ID"` returned “Active Goal: none (run is not goal-bound)”; directives reported none. Campaign rules prohibit LOGBOOK.md edits and no logbook executable is available; the review result, base-movement caveat and coverage bound are persisted here and in board notes. Attach this task-scoped verdict before `accept_cr(TASK-260916-1hrx51, revision=2, evidence=TASK-260916-1hrx51_review-verdict-rev2.md)`. No `commit_ack` is supplied and no `done` transition is requested.
