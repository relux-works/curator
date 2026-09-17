# TASK-260916-1x0ogh — review verdict, revision 2

Verdict: **accepted** for CR-TASK-260916-1x0ogh-2 revision 2. F1–F4 are closed; no remaining correction required within this task's scope. Route with `accept_cr` to `integrating`, not `done`. Landing and reconciliation with advanced main remain producer integration work.

## Candidate identity and the empty runtime delta

The runtime CR has an empty **curator** repository delta: `git diff fcb4faf7ddf1e2fa0dc7faa70afcd403e423e529 35fcca9f1cb219388226d12a284e8b9491719e3e --stat` produced no output. That is the correct repository boundary: this leaf delivers a **curator-spec** revision and explicitly excludes manager/launcher implementation. Its nonempty deliverable is the attached spec patch and the separate spec Story worktree, not curator source changes. Acceptance does not claim that the spec has already merged.

Reviewed worktree: `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260916-2otjbn/worktree`, branch `task-board/story/STORY-260916-2otjbn`, HEAD `07e2b41892bf24c33b859072b2b920fe2d0a17fe`. Current origin/main `f544a017fbabcbe40ba985ce6a52af3d4b3e9ddc` includes separate S6/E2 landings. The comparison therefore uses the recorded fork base (HEAD), as the round-2 brief requires, rather than introducing unrelated reverse changes from advanced origin/main.

Using an alternate temporary Git index (`read-tree HEAD`, `add -N .`, `diff HEAD`), the complete worktree patch including all eight new files is **byte-identical** to `TASK-260916-1x0ogh_spec-patch_rev2.patch`:

- SHA-256: `2d66b496c5779a4afa982369690aac5cb81362005099db146bedc4f9f406a38a`.
- Stable patch ID: `2ac6d1969dcbed6ec9168fe85a8818637a7b2a2b`.
- 112 paths: 104 tracked modifications and 8 new files.

The candidate files and real index were not edited. Regeneration and adversarial edits ran in a disposable byte copy, with the exact candidate staged as its comparison baseline. No branch, commit, push, or integration was performed. The campaign rules, producer/rework/review briefs, producer evidence, both patches, prior verdict, and audit E4 plus Appendix B were read. `task-board spawn goal "$TASK_BOARD_RUN_ID"` reported `Active Goal: none (run is not goal-bound)`.

## Brief conformance and round-2 closure

All file references below are relative to the reviewed spec worktree.

| Requirement | Exact text / checked evidence | Result |
|---|---|---|
| Trust roots, install resolution, ordered first match | `protocol/environments.md:2167`: “The provider trust roots, in search order, are exactly”; `:2169`: “executable, resolved after symlinks”; `:2171`: “in listed order”; `:2172`: “The first executable regular file”; `:2248`: “manager searches the trust roots in order” | Pass |
| Preserve pre-existing discovery constraints | `protocol/environments.md:2156`: “MUST match the core §2”; `:2158`: “An implemented subcommand always wins”; `:2160`: “no provider registry”; `:2163`: “nothing is downloaded or installed implicitly”; `:2164`: “Profile data, marker data, and fragment data MUST NOT influence” | Pass |
| PATH exclusion and explicit S6 attack | `protocol/environments.md:2185`: “The ambient `PATH` is not a trust root”; `:2211`: “The attack this rule closes is the S6-injected `PATH` (finding E4)”; `:2212`: “a project `.agents/env.sh`” | Pass |
| Published / managed refusal preserved | `protocol/environments.md:2192`: “The resolved executable MUST NOT reside”; `:2197–2199` applies refusal to A selection and B trust-root/probe matches, “outranking dispatch in both” | Pass |
| Printed path and status posture | `protocol/environments.md:2203`: “Every dispatch reports the resolved absolute provider path”; `:2307–2314`: per-provider path/trust verdict and “always for `curator-run` and `curator-session`”; missing and unreadable reported; `:2327–2334`: refusal/failure non-current, A warning stays current | Pass |
| F2: two explicit rollout profiles with A unchanged selection | `protocol/environments.md:2230`: “Revision A (warning release)”; `:2231`: “searches the ambient `PATH` in order”; `:2242`: “provider's directory in `provider_directories`”; `:2245–2246`: “revision A never consults the trust roots for selection”; `:2247`: “Revision B (flip release). The ambient `PATH` never selects.” `cli/curator.md:123–126` names both profiles | Closed |
| F3: B dispatch/refusal/probe/absence/failure distinctions | `protocol/environments.md:2251`: “The five outcomes are mutually exclusive”; `:2257`: “diagnostic-only `PATH` probe”; `:2263`: “a `PATH`-only match never reports missing”; §11.1 `:2275–2278` contains the four diagnostic rows with path/root details | Closed |
| F4: unreadable is failure, both revisions | `protocol/environments.md:2180`: “is a read failure, never absence”; `:2181`: “MUST fail with `subcommand_provider_root_unreadable`”; `:2182`: “MUST NOT fall through”; A `:2236–2237`, B `:2264–2266`, diagnostic table `:2278`, status `:2313–2314` and `:2331–2332` agree | Closed |
| Closed knob/default/lockability | `protocol/environments.md:2380`: `provider_directories`, “list of absolute paths”, `[]`, section 11; `:2400–2402`: “`provider_directories` — a locked provider list is fleet policy”; `profiles/manager.md:57` adds `environments.provider_directories`; manager schema `:501` and system schema `:24,:51` admit the exact key | Pass |
| Closed diagnostics and schema surfaces | Four exact §11.1 diagnostic spellings equal `UMBRELLA_PROVIDER_DIAGNOSTICS` in `tools/validate.py:4598`; generated outcomes use the same set. Machine/system schemas stay closed, only the authorized knob/lock key are admitted; old schema version 1 is untouched | Pass |
| Conformance registration and schema cases | `protocol/environments.md:2446–2462` names the new suite and both profiles; `conformance/v1/manifest.json:4216` registers `vectors/umbrella-provider-resolution.json`; schema-case index registers POSIX/Windows positives, relative/duplicate negatives; system valid fixture includes and locks the list | Pass |
| F1: semantic derivation and production caller | `tools/validate.py:4684` `_umbrella_expected` derives A/B outcomes from inputs; `:4918–4939` compares resolution, diagnostic and details; `:4964` includes `validate_umbrella_provider_vectors` in `main`. `tools/test_validate.py:1988` onward contains 18 umbrella tests, including the original silent-success mutant at `:2084` | Closed; original mutant independently rejected at the real entry point |
| CHANGELOG and dependency | `CHANGELOG.md:99`: “E4: umbrella provider resolution”; `:110–118` explicitly states A keeps PATH selection/warns and B trust-root selection/refuses; `:124`: “Blocks proposal 0016 / `path_prepend`” | Pass |
| Scope and architecture | Changed paths are normative/CLI/profile docs, the authorized two schemas, generated schema fixtures/vector/manifest/release manifest pins, and spec conformance generator/validator/tests. Conformance tooling is explicitly authorized by F1. No manager/launcher implementation or proposal 0014–0018 document was changed | Pass |

The general trust-root search language is read with the explicitly labelled A/B rollout rules: A selects on PATH; B stops at the first trusted-root match or read failure. Later roots after a B match are not searched. This matches the ordered-search model rather than requiring unrelated later roots to be read after selection.

## Vector and regression review

Opened all **14/14 cases**, reviewed both profiles (**28/28 declared outcomes**): install-only (A missing/B resolves); install over listed; listed over PATH (A warns on PATH/B listed); listed-order first match; S6 PATH-only A-warning/B-refusal; non-executable trusted candidate skipped; manager-published refused; listed-but-published still refused; managed directory refused; unreadable listed root fails; unreadable install root fails; PATH selects another provider while trusted exists; trusted provider on PATH silently resolves; true missing.

Every refusal carries its refused path and consulted roots. Warnings carry the selected path, roots and migration directory. Unreadable cases identify the failed directory and do not activate fallback. These cases exercise actual branch inputs rather than merely restating defaults.

Diffed rev1 versus rev2 patches: only nine paths differ — CHANGELOG, CLI, environments protocol, umbrella vector/generator, validator/tests, manifest, and rc.9 manifest pins. The schema and schema-case patch hunks are unchanged between revisions. Independently compared all **88/88 pre-existing changed schema fixtures** with the fork base after stripping only the new knob/lock member: their prior content is preserved exactly. Regeneration re-produced the candidate artifacts exactly.

Measured adversarial result: **1/1 original production-entry semantic mutants rejected**, compared with 0/1 in round 1. The exact S6 revision-B outcome was replaced with `{"resolved":"/home/operator/work/acme/.bin/curator-run","diagnostic":null}` in the disposable copy; the vector manifest digest and both release manifest pins were refreshed so integrity checks could not mask semantic failure. The checker itself was unchanged. Original bytes were restored afterward.

Bound: this is specification/conformance-model evidence over abstract POSIX fixture paths. It does not test real filesystem symlink/permission behavior, Windows execution, or the manager/launcher runtime; those remain implementation tasks named in the brief. No claim is made that all possible filesystem states or combinations have been exercised.

## Independent validation

Shell `/bin/zsh`, `set -o pipefail`. The main gate ran directly from the original candidate worktree:

`PATH="/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin:$PATH" make validate`

The regeneration gate ran from the disposable exact candidate copy, with that candidate staged as its Git diff baseline, using the same PATH and shell:

`make regenerate-check`

This is necessary to keep review read-only and avoid measuring the producer's intentional uncommitted delta as regeneration drift. No original-candidate regeneration writes were made. No producer-only test result is substituted for these independent gate runs.

One supplemental inspection initially used system Python and could not import `jsonschema`; it was immediately rerun successfully with the prescribed repository venv. This was not a make-gate failure.

### make validate (original candidate)

```text
python3 tools/validate.py
validated 60 schemas and 1054 vector files
python3 -B -m unittest discover -s tools -p 'test_*.py'
.....................................................................................................................................................................................................................................................
----------------------------------------------------------------------
Ran 245 tests in 297.859s

OK
go test ./tools/...
ok  	github.com/relux-works/curator-spec/tools/generate-vectors	6.178s
EXIT_CODE=0
```

### make regenerate-check (disposable exact candidate)

```text
go run ./tools/generate-vectors -root .
git diff --exit-code -- conformance/v1 release/1.0.0-rc.5.json release/1.0.0-rc.6.json release/1.0.0-rc.7.json release/1.0.0-rc.8.json release/1.0.0-rc.9.json
EXIT_CODE=0
```

### Production-entry S6 mutant

```text
Original S6 silent-success mutant; vector digest and both manifest pins refreshed.
validation failed: umbrella provider case s6-planted-path-provider-warns-then-refuses revision_b resolves '/home/operator/work/acme/.bin/curator-run' but the §11 model expects None
EXIT_CODE=1
```

Final identity recheck: the complete patch SHA-256 is unchanged; **112/112 changed paths** equal the regenerated, mutant-restored copy. No candidate changes occurred during review.

Important review results are persisted in this task-scoped verdict and board notes. No LOGBOOK.md edit was made because the campaign explicitly forbids it. No remaining findings require producer rework. The new tracked doc-writer/implementer integration run must reconcile the advanced spec main and perform the eventual landing; this reviewer makes no merged/done claim.
