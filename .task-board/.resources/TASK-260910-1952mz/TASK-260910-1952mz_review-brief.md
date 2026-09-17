# Review brief — TASK-260910-1952mz (S6 shell-hook trust gate, curator manager), review round 4

You are the independent reviewer of a curator implementation produced for
`TASK-260910-1952mz` (story `STORY-260910-2awkzu`, wave 1 of the 2026-09 security-audit
remediation). Read, in this order: `remediation-manager-producer-rules.md`,
the producer brief `TASK-260910-1952mz_brief.md`, the producer results
`TASK-260910-1952mz_results.md`, the published Change Request patch
`TASK-260910-1952mz_change-request_rev4.patch` and its validation log, and the
landed spec sections the brief names (curator-spec checkout
`/Users/administrator/Developer/ReluxWorks/curator/curator-spec`, read-only).

## Where the candidate is
The managed Story worktree `<control-root>/.temp/STORY-260910-2awkzu/worktree` on
branch `task-board/story/STORY-260910-2awkzu` holds the exact candidate the runtime
published as revision 4 (the hosted gate `scripts/remote-gate.sh` ran
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
Record `TASK-260910-1952mz_review-verdict-rev4.md` (task outcome) with the
per-item table, transcripts, mutants and findings; then exactly one of
`task-board m 'accept_cr(TASK-260910-1952mz, revision=4, evidence=TASK-260910-1952mz_review-verdict-rev4.md)'`
or a changes-requested verdict routed with `set_status(TASK-260910-1952mz, status=to-dev)`
listing the concrete corrections. Never accept on the producer's evidence
alone; never edit the candidate.

## Round specifics (revision 4)
Revision 2 was rejected with four corrections
(`TASK-260910-1952mz_review-verdict-rev2.md`): R1 hooks accepted malformed
records; R2 the dash failure had been hidden by dropping `sh` from the test
instead of repairing the Bash-array syntax; R3 lexical-only path identity;
R4 non-atomic state replacement. Revision 3 repaired those but failed the
gate on `TestShellHookTrustResolvesSymlinkedProject/powershell` (PowerShell
hook looked up the alias spelling); revision 4 passed the gate on all lanes.
Verify each R1–R4 closure exactly as the rev-3 brief
(`TASK-260910-1952mz_rework-rev3.md`) demanded: run the generated hook under
`sh` and `dash -n`; probe malformed records (missing member, `approved_by`
outside manager|operator, bad timestamp) against the emitted hooks under
`B-enforcing`; a symlinked project under both profiles for BOTH hooks (use
`pwsh` if present on this host); a failed publication preserving the last
valid state. Confirm the §8.7 vector recipe runs for the executable cases
and skips PowerShell legitimately only where `pwsh` is absent. The sibling
`TASK-260910-3ungjy` is not part of this revision.
