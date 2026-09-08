# Protocol launch-boundary errata — review candidate

Task: TASK-260908-2zn8fu, STORY-260908-6gwnfc.
Base: 87a0d0060bad64ab883d007dcdf35df7485368bf.
Candidate: uncommitted changes only in decisions/0013-execution-ownership-and-launch-plans.md and protocol/environments.md.
Initial worktree was clean; no previous WIP or candidate patch was present in this task's outcomes.
Parent owns independent review and signed publication; this producer does not assert that its candidate has already passed independent review.

## Accepted evidence map

| Changed statements | Accepted source and exact boundary |
| --- | --- |
| Decision 0013 §6.3 and environments §10.1: process environment enters request; full Plan.Env is untracked base; own-name extraction, collision subtraction and SHOULD-warn | A0 TASK-260908-qblycn_a0-verification-findings.md §3.5/E4; accepted TASK-260908-1c0fwn_results.md §§1–5 and review-verdict-rev1.md. skill-agents-management v0.5.10: annotated tag 13d167e2c5cabb226eb0563f591ae6c63095855a, peeled commit 12f443d10bc217ca7a48e2edab19c739f441df9c; pkg/agentic/plan.go:38/280, system.go:419/534–541, runcontext.go:61–100, systems/claude/env.go, systems/codex/env.go, internal/runtimeenv, internal/launchenv:39. |
| Decision 0013 §6.4: retain complete composed Argv; never serialize Binary or inherited Plan.Env | A0 reviewer TASK-260908-qblycn_review-verdict-rev1.md F1; A0 §3.4 tagged interactive argv; accepted environment evidence §§3–5. F1 is a documented contradiction, not a runtime reproduction. |
| Environments §§5.5/7.3: native flags suppress same-semantics discovery; trusted project file wins over agent-home; default off and managed-home-only remain | A0 §4.3/E5 and accepted A0 review: installed pi 0.84.2 dist/core/resource-loader.js:380/386/808–829. Removed the contradictory “only replace path” statement; no replace-flag descriptor was added. |
| Environments §7.3 project-local probe residual | A0 §4.3, matrix §5 and E5: launcher managed-home probes do not inspect trusted project-local .pi files. No new probe or trust decision claimed. |
| Decision 0013 open question 6 and environments §10.1 tracked unset/PATH residual | Accepted environment evidence §4 and independent review: closed Decision 0013 §3.2 grammar lacks destination env-unset/PATH-transform fields. Destination ax filtering remains unknown. |

All source analysis/probes above are accepted prior evidence, not rerun by this doc writer. No runtime launches, ax inspection, MCP-channel changes, code, schema/vector changes, protocol version changes, tags/releases, hosted CI, installs, runtime-home changes, or commits.

## Validation

- Canonical command: `PATH=/Users/iv/Developer/ReluxWorks/curator-spec/.temp/STORY-260905-2z9pw4/worktree/.venv/bin:$PATH PYTHONDONTWRITEBYTECODE=1 make validate` — exit 0.
- It ran all three existing Makefile gates: tools/validate.py (60 schemas, 1047 vector files), Python unittest discovery, and go test ./tools/....
- `git diff --check` — exit 0 after the last documentation edit.
- Self-review covered every changed hunk and searched both documents for obsolete inherited-overlay, element-zero deletion, and unconditional Pi wording; those contradictory occurrences are absent.
- Validation is local on the uncommitted candidate over the stated base; no hosted CI or new prose-mirroring tests.
- Canonical logs: task-scoped validation log outcome (first failed attempt and successful retry); local originals in .temp/TASK-260908-2zn8fu/.

## Operational record (logbook equivalent)

The explicit task brief prohibits LOGBOOK/control-root ordinary writes; this task-scoped outcome records findings instead.
Initial set_status(development) exited 1 because an estimate was required. Set Fibonacci estimate 2 and retried successfully (exit 0).
Initial make validate exited 2: default Python lacked jsonschema; no document validation ran.
System Python dependency probe exited 1 too. Reused the existing Python 3.12 validation venv after an import readiness probe exited 0; no dependencies installed.
A scratch path inventory attempt encountered zsh glob/command syntax errors and was replaced by bounded read-only directory enumeration.
The board emitted a Story-base preflight reminder; this assigned worker used the parent-provisioned frozen 87a0d006 worktree and did not branch, fetch, switch, or rebase.
