# Review brief — TASK-260910-1952mz (S6 shell-hook trust gate, curator manager), review round 6

You are the independent reviewer of a curator implementation produced for
`TASK-260910-1952mz` (story `STORY-260910-2awkzu`, wave 1 of the 2026-09 security-audit
remediation). Read, in this order: `remediation-manager-producer-rules.md`,
the producer brief `TASK-260910-1952mz_brief.md`, the producer results
`TASK-260910-1952mz_results.md`, the published Change Request patch
`TASK-260910-1952mz_change-request_rev6.patch` and its validation log, and the
landed spec sections the brief names (curator-spec checkout
`/Users/administrator/Developer/ReluxWorks/curator/curator-spec`, read-only).

## Where the candidate is
The managed Story worktree `<control-root>/.temp/STORY-260910-2awkzu/worktree` on
branch `task-board/story/STORY-260910-2awkzu` holds the exact candidate the runtime
published as revision 6 (the hosted gate `scripts/remote-gate.sh` ran
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
Record `TASK-260910-1952mz_review-verdict-rev6.md` (task outcome) with the
per-item table, transcripts, mutants and findings; then exactly one of
`task-board m 'accept_cr(TASK-260910-1952mz, revision=6, evidence=TASK-260910-1952mz_review-verdict-rev6.md)'`
or a changes-requested verdict routed with `set_status(TASK-260910-1952mz, status=to-dev)`
listing the concrete corrections. Never accept on the producer's evidence
alone; never edit the candidate.

## Round specifics (revision 6)
Revision 4 was rejected with one correction
(`TASK-260910-1952mz_review-verdict-rev4.md`): remove the GOOS-wide Windows
POSIX skips and reconcile the native (`C:\...`) and MSYS (`/c/...`) approval
identity for Git Bash on Windows (rev-2 R2/R3 residue). Revision 5
(`TASK-260910-1952mz_rework-rev5.md`) added the `cygpath`-based identity
mapping, interpreter probes instead of GOOS skips, and the cross-spelling
proof test `TestShellHookTrustNativeRecordAuthorizesMSYSSpelling`; it failed
the hosted gate only in that test's `changed/A-warning` subcase because the
harness compared the sourced marker against the default `1` while the
changed bytes export `2` (`TASK-260910-1952mz_gate-failure-rev5.md`).
Revision 6 is the repair; it passed the hosted gate on all lanes.

Verify, beyond the standing items above:
- the rev-6 diff against rev 5 (`git diff` of the two patch resources or
  `interdiff`) is the harness repair plus whatever the producer states —
  nothing in the hook, `hookapproval` or the identity mapping regressed;
- the Windows identity rule as implemented: which spelling is canonical in
  the record, how the hook maps the other one (`cygpath` presence/absence
  branches — what happens on a Windows host without `cygpath`, must be
  fail-closed, not a skip), and that the same file under both spellings
  yields one record key while two different files never collide;
- the interpreter probes: skips name the absent interpreter (host
  capability), never the GOOS; the `posixTrustShells` dedupe claim holds
  (report if the same binary is run twice under different names);
- the profile-A changed-bytes semantics on every shell: warn once, source
  the current bytes; profile B: refuse, warn once;
- previous R1–R4 closures still hold (spot-check: malformed record, `dash -n`,
  symlinked project for both hooks, failed publication keeps the last state).
The sibling `TASK-260910-3ungjy` is not part of this revision.
