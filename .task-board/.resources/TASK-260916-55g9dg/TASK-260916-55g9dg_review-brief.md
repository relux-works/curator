# Review brief — TASK-260916-55g9dg (E2 direct-only class: system modules, curator manager), review round 2

You are the independent reviewer of a curator implementation produced for
`TASK-260916-55g9dg` (story `STORY-260916-2d9coh`, wave 1 of the 2026-09 security-audit
remediation). Read, in this order: `remediation-manager-producer-rules.md`,
the producer brief `TASK-260916-55g9dg_brief.md`, the producer results
`TASK-260916-55g9dg_results.md`, the published Change Request patch
`TASK-260916-55g9dg_change-request_rev2.patch` and its validation log, and the
landed spec sections the brief names (curator-spec checkout
`/Users/administrator/Developer/ReluxWorks/curator/curator-spec`, read-only).

## Where the candidate is
The managed Story worktree `<control-root>/.temp/STORY-260916-2d9coh/worktree` on
branch `task-board/story/STORY-260916-2d9coh` holds the exact candidate the runtime
published as revision 2 (the hosted gate `scripts/remote-gate.sh` ran
green on it, see the validation log). Do not edit it; read, build and test
in place (or in a disposable copy).

## What to verify
1. **Spec conformance, item by item** against the landed normative text:
   closed diagnostics spelled identically, knob defaults, RFC 2119
   obligations, "unreadable is never absence" at new read sites, the
   rollout profile shipped as default (warn first) with both profiles
   implemented behind one option, posture rows in `curator status` /
   `env status` exactly as the spec §-posture text says. Quote file:line.
2. **Vectors are driven**: the vector-execution test reaches the production
   entry point named in the AC (not a helper-only test), consumes the family
   from `CURATOR_CONFORMANCE_ROOT`, and takes the `root-content` skip with a
   `platform-cases.tsv` row when the root lacks it. Run it yourself with
   `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1`.
3. **Independent validation**: `go build ./... && go vet ./... && gofmt -l .`
   and the narrow package tests (`set -o pipefail`, quote exit codes); attack
   the gate with at least two narrowing mutants (e.g. flip the enforcing
   branch, drop a diagnostic) and confirm a committed test catches each;
   report survivors with bounds.
4. **Scope and hygiene**: no `SPEC_PIN` change, no vendored spec bytes, no
   ax/proposal content, no unrelated edits; CHANGELOG Unreleased entry names
   the finding and the warning release; no writes outside the worktree.
5. Anything the producer reports as a spec gap: check it against the text;
   a real gap is a finding for the orchestrator, not a reason to accept
   divergent behaviour.

## Verdict
Record `TASK-260916-55g9dg_review-verdict-rev2.md` (task outcome) with the
per-item table, transcripts, mutants and findings; then exactly one of
`task-board m 'accept_cr(TASK-260916-55g9dg, revision=2, evidence=TASK-260916-55g9dg_review-verdict-rev2.md)'`
or a changes-requested verdict routed with `set_status(TASK-260916-55g9dg, status=to-dev)`
listing the concrete corrections. Never accept on the producer's evidence
alone; never edit the candidate.

## Round-2 specifics
Revision 1 failed the hosted gate only on Linux in `internal/envprofile`
`TestStatusReportsPolicyAndDropped` (a passthrough entry reported detached
made the asserted row non-current) — see
`TASK-260916-55g9dg_gate-failure-rev1.md`; revision 2 passed on all lanes.
Verify the fix addressed the fixture/assertion honestly (no blanket
warning filtering that would hide real posture regressions). Spec sections:
environments §3 (direct vs transitive; waivers), §5.5/§5.7
(`context_system_module_dropped` warning, `context_system_module_transitive`
error, lock unchanged on error), §12.1/§12.2 (`transitive_system_modules`
default `drop`, lockable to `error` only; `system_module_waivers` not
lockable), §12 posture (policy value + every dropped module by package and
path; drop warnings never make a row non-current), the five admission
vectors in `vectors/environments.json`, and the `manager-config-v2` /
`system-config-v2` schema cases for the two knobs. Also check how the
implementation keeps `TestManagerConfigV2Vectors` green against the pinned
rc.11 root (the E4 task's knob broke it): if the new knobs are omitted from
the normalized defaults output only to satisfy the old vector while the new
root's vector expects them, say so explicitly as a finding for the
orchestrator (pin-lag decision pending) rather than accepting a divergence
silently.
