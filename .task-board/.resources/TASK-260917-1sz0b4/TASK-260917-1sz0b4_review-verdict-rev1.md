# TASK-260917-1sz0b4 — landing review

Verdict: **accept-landing** (accepted).

Reviewed exact curator-spec PR #63 head `23dafa798fa80fc2591ddb287c1c6345e2715b3b`, parent `0da40207a70d0b6c990c8f1bc8c79e59f218bf6d`. `git ls-remote origin refs/pull/63/head` returned that head. Delivery worktree was clean; all 1352/1352 tracked file bytes independently matched Git blobs at the head. The attached rebased diff equals `git diff 0da4020..23dafa7` byte for byte. Accepted patch SHA-256: `7d0b2d785dd2802cdf201a5fe051e31c358aad03e0b359a7d34d7da62aabf788`; reconstructed by applying it to an archive of `07e2b41`. The prior acceptance resource is actually `TASK-260910-2ohnjo_review-verdict-rev4.md` (review round 3, CR revision 4), not the rev3 verdict filename in the brief. It accepts this same rev3 patch.

No candidate code, index, commits, branches or parent task were changed. Generation/mutations used a disposable byte copy with a private staged Git baseline, no scratch commit. `core.autocrlf=false` preserved fixture bytes. Scope is specification and its conformance tooling; this is not proof of curator runtime implementation or a claim that PR #63 has already merged.

## Per-file merge table

All file:line references are in the exact PR head. Compared every added/deleted patch line against accepted rev3, then inspected the overlapping sections against landed `0da4020`. All 11/11 changed paths are accounted for.

| File | Classification and merge evidence |
|---|---|
| `CHANGELOG.md` | Union, identical S4 added bytes. :10 “S6: shell-hook project env trust gate”; :61 “E2 direct-only”; :144 “E4: umbrella provider resolution”; :170 “S4 (MCP env passthrough and declaration surfacing)”. Each entry appears once. Landed entries remain intact; S4 is appended after E4. |
| `conformance/v1/manifest.json` | Regenerated. :4184 registers `vectors/environments-env-passthrough.json`, :4212 retains `vectors/manager-config-v2.json` with refreshed hash. Generator check proves exact bytes; existing S6/E2/E4 inventory remains. |
| `conformance/v1/vectors/environments-env-passthrough.json` | Entire file byte-identical to accepted rev3. 25 cases: 7 default resolution, 6 allowlist, 6 surfacing, 4 schema, 2 order; includes spaced and escaped-quote arguments. |
| `conformance/v1/vectors/manager-config-v2.json` | Regenerated union. Independent structural comparison: accepted 46 cases union landed 44 = exactly 48 names, no duplicate/additional/lost names. Every landed case equals its former value except the required absent-knob default `null` → `[]`; every new S4 case equals accepted content plus E2/E4 default keys. :1607 “schema2-provider-directories”, :1621 relative refusal, :1663 absent-empty, :1707 explicit-null-unbounded, :1751 explicit-empty, :1765 invalid-name. |
| `profiles/manager.md` | Context-only rebase: all +/- patch lines identical to accepted S4; landed changes retained. :2361 “empty list permits every network identity with the” warning at :2362; :2587 “default empty — opt-in per name; explicit `null` stays unbounded”. S6 §8 remains the landed authority. |
| `protocol/environments.md` | Union at §12 reporting/warnings and §13 surfaces; all other S4 edits identical. :2440 “`transitive_system_modules` value with every dropped system module”; :2442 “absolute provider path and trust verdict”; :2447 “reported unreadable with the directory”; :2448–2451 add S4 warning, profile/allowlist and surfacing. :2474–2475 retain `context_system_module_dropped` plus three S4 warning names. :2593–2600 preserve E2 admission cases; :2602–2606 add S4; :2607 preserves snapshot; :2608–2624 preserve full E4 cases and no-PATH revision-B claim. Differences from accepted patch edits are only the necessary conjunction/wrapping and retained E2/E4 clauses. No third rule was introduced. |
| `release/1.0.0-rc.9.json` | Regenerated candidate manifest pins only, :15 and :27 `sha256:fcdc8ecac2348375b3bc5462d3f4bec5d56e75d7f8590c61b3f653f4c8a074b2`. Historical release pins unchanged. |
| `schemas/v1/manager-config-v2.schema.json` | Context-only rebase: identical accepted +/- delta, :470 `"default": []`. E2 `transitive_system_modules` and waiver schema at :494 onward, E4 `provider_directories` at :531 onward remain. |
| `tools/generate-vectors/manager_config.go` | Union; only padding differs in accepted default edit. :22 `"passable_env_names": []any{}` alongside :26 `"transitive_system_modules": "drop"`, :27 empty waivers, :31 empty providers. Schema-example list :180–184 preserves E2 cases and :188–191 preserves E4 cases. Vector list :333–336 preserves E4 cases; :337–352 adds the exact four accepted S4 cases. No landed code deleted. |
| `tools/test_validate.py` | Identical accepted +/- delta with landed tests retained. :1731 `EnvPassthroughVectorTests`; :1927 escaped-quote parser test. No rebase-specific semantics. |
| `tools/validate.py` | Identical accepted +/- delta with landed consumers retained. :4713 S4 consumer; real `main` check list :5637 invokes it, :5641 invokes S6 shell-hook trust, :5644 invokes E4 provider resolution. |

Independent byte comparisons also establish unchanged landed CLI (`cli/curator.md`), S6 and E4 behavior vector files, and system-config-v2 schema. Entire environments §3, §5.5, §11 and §12.2 are byte-identical to `0da4020`, with one heading each. Full comparison output and reproducible assertion script are attached.

## Semantic and closed-set review

- S4: environments :449–455 retain reserved-name exclusions, opt-in absence and explicit unbounded null. :471–474 require the empty-package-allowlist warning at install, update and status. §2.3 :476 onward owns six columns (`package`, `version`, `transport`, `command`, `args`, `env_names`), compact JSON and one LF. :1693 and :1745 repeat the same before-lock-publication timing for install/update, after audit, before materialization. :2237–2261 owns the single rollout definition: `s4-warn` before `s4-enforce`, explicit list remains bounded, absent warn is unbounded with migration warning, absent enforce drops, explicit null stays unbounded. :2512 fixes default `[]`; §12.2 retains both S4 lockable keys. Diagnostic spelling is consistently `mcp_package_allowlist_empty`, `mcp_env_passthrough_unlisted`, `mcp_env_passthrough_dropped`; repetitions in diagnostic tables/status are references to the same rules, not duplicate or contradictory definitions.
- S6: unchanged manager §8 defines manager-home approval records `{path, sha256, approved_by, approved_at}`, closed `manager`/`operator`, exact-byte digest binding, rejection of project-supplied approval records; :1183 closes rollout to `A-warning`/`B-enforcing`. CLI :50 and :101–106 retain approve and rollout behavior. `shell_hook_env_unapproved` and `shell_hook_env_changed` retain their spellings across prose/CLI/vectors.
- E2: unchanged environments §3 admits only direct packages and named waivers. Default `drop` warns `context_system_module_dropped` at materialization, while `error` fails resolution with `context_system_module_transitive` without changing the lock. §5.5 retains exact admitted bytes and fragment/extension binding. :2516–2517 preserve knobs/defaults; :2555–2558 permit locking only toward `error` and prohibit locking waivers. No warn-first split added.
- E4: unchanged §11 defines install directory then listed absolute `provider_directories`, first executable regular file, unreadable-root failure without fallback, and published/managed directory refusals. Revision A preserves PATH selection with `subcommand_provider_outside_trust_roots`; revision B selects only trust roots, uses PATH diagnostically and refuses untrusted matches. Closed missing/untrusted/outside-trust-roots/root-unreadable diagnostics remain aligned with vectors and schema. §12 preserves warning-current versus refusal/missing/unreadable-non-current distinctions. CLI is byte-identical to landed main; no new CLI subcommand or spelling was invented for S4.

The authoritative sections, entries and case names appear once; intentional cross-references and diagnostic tables remain consistent. The solution preserves the project's normative-document → schema → generated vectors → semantic validator architecture.

## Independent verification

All gates below were rerun by this reviewer on this head, not accepted from earlier transcripts. Repository venv first on PATH; `set -o pipefail`. `make validate` ran in the read-only delivery worktree with `PYTHONDONTWRITEBYTECODE=1`. `make regenerate-check` ran in its byte-identical disposable copy against a staged private Git baseline because generation writes files. No code commit was needed. `git diff 0da4020..23dafa7 --check` exited 0.

Generation command: `PATH="/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin:$PATH" make regenerate-check`.

```text
go run ./tools/generate-vectors -root .
git diff --exit-code -- conformance/v1 release/1.0.0-rc.5.json release/1.0.0-rc.6.json release/1.0.0-rc.7.json release/1.0.0-rc.8.json release/1.0.0-rc.9.json
regenerate_check_exit=0
```

Mutation command: `python3 /tmp/TASK-260917-1sz0b4-mutants.py`. Each mutant changes one vector, refreshes only manifest/release hashes (and the INVALID output length/hash), asserts mutated content is still present, and invokes the real venv `python3 -B tools/validate.py` entry point.

```text
unlisted-passed exit=1  validation failed: env-passthrough case s4-enforce-absent-drops-all: passed/dropped are not the s4-enforce rule ([]/['FIGMA_API_KEY'])

invalid-output exit=1  validation failed: env-passthrough case single-stdio-declaration: expected_bytes are stale

sourced exit=1  validation failed: shell-hook-trust case changed-env-sh-B-enforcing-not-sourced sourcing does not follow its trust state and profile

resolution exit=1  validation failed: umbrella provider case s6-planted-path-provider-warns-then-refuses revision_b resolves '/home/operator/work/acme/.bin/curator-run' but the §11 model expects None

mutants rejected: 4/4; restored scratch diff exit=0
```

Measured mutation coverage: 4/4 required attacks rejected semantically, across S4, S6 and E4. S4 vector consumer covers 25/25 declared cases; manager-config inventory comparison covers 48/48 cases. These figures concern these concrete spec fixtures, not exhaustive runtime security coverage.

Verification-harness correction: the first E4 probe returned 0 because the generator overwrote its mutated generated vector before validation. That run is discarded, not counted as a surviving semantic mutant. The final attached script updates pins without regeneration and asserts preservation; E4 then exits 1 at the §11 resolver model. The first scratch patch application also failed on macOS /var symlink traversal; applying from the resolved scratch directory succeeded. A comparison script initially treated the manager vector array as an object; corrected and rerun successfully. These were scratch-only harness issues.

Validation command: `PATH="/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin:$PATH" PYTHONDONTWRITEBYTECODE=1 make validate`. Full transcript:

```text
python3 tools/validate.py
validated 60 schemas and 1067 vector files
python3 -B -m unittest discover -s tools -p 'test_*.py'
................................................................................................................................................................................................................................................................................................
----------------------------------------------------------------------
Ran 288 tests in 439.246s

OK
go test ./tools/...
ok  	github.com/relux-works/curator-spec/tools/generate-vectors	(cached)
make_validate_exit=0
```

The Makefile Go gate used its cache; to establish fresh execution as well, I additionally ran `go test -count=1 ./tools/...`:

```text
ok  	github.com/relux-works/curator-spec/tools/generate-vectors	1.690s
go_uncached_exit=0
```

Final restoration comparison: 1352/1352 scratch files equal delivery head bytes. Delivery status remains clean. No earlier producer/reviewer validation result substitutes for these reruns.

## Lifecycle

`task-board spawn goal "$TASK_BOARD_RUN_ID"` reports “Active Goal: none (run is not goal-bound)”; no directives. Acceptance evidence is attached before changing status. This is a standalone landing review with no Change Request; following its explicit brief, route THIS task to `done` with no `commit_ack`. Do not change TASK-260910-2ohnjo. The orchestrator owns landing PR #63 and its separate `integrate_external` step. Checklist 1–6 verified; item 7 is conditional on non-acceptance and does not apply.
