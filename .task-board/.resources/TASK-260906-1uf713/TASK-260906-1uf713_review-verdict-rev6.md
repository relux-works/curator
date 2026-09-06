# TASK-260906-1uf713 — review verdict, Change Request revision 6

**ACCEPT.** Reviewer run `RUN-260906-a1742f`, cycle 6 (rework 5).
Full findings: `TASK-260906-1uf713_review-findings-stage-c-6.md`.
Probes and mutants: `TASK-260906-1uf713_review-probes-stage-c-6.tgz`.

## Subject

- `CR-TASK-260906-1uf713-6` revision 6, element `TASK-260906-1uf713`, integration scope
  `STORY-260906-1a2i5a`, base `87a0d0060bad64ab883d007dcdf35df7485368bf`.
- The reviewed work is `feat/agent-environments-stage-c` at
  `8b8aa0417f49c11246a0a53d2006ef75708d4ea5` in
  `/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-c`, 32 signed commits, and the leaf's
  eighteen outcome resources. The producer worktree was read-only throughout: HEAD `8b8aa041` and
  `git status --short` empty before and after.

## Why an empty repository delta is the right outcome here

The CR reports `repository_delta: empty`, and I verified that rather than assuming it:

- The story worktree is a checkout of **curator-spec** (`origin = curator-spec.git`), while this
  leaf's entire scope is the **curator** repository on a different branch in a different worktree.
  Every brief forbids pushing, touching PR #61, or writing into the control root. The leaf
  structurally cannot produce a delta here, and revisions 1–4 were empty for the same reason.
- Candidate tree `2e6ca472ae8542a95e446ed4afb11cb75820536d` equals
  `git rev-parse 87a0d0060bad…^{tree}`, and the patch resource's sha256 `e3b0c442…b855` is the
  sha256 of the empty string. Both checked directly.
- Cycle 5 warned that **revision 5** was not empty and that its delta was another element's work
  (the five curator-spec commits of TASK-260906-3x0w4y's `path`-overlay reconciliation). Revision 6
  resolves that: the base advanced to `87a0d00`, so those commits are in the base, not the delta.
  **Nothing foreign is being accepted.**

Accepting an empty delta here therefore means accepting the leaf's work where it actually lives —
the 32 signed commits on the curator branch — which is what I reviewed and attacked directly.

## Verdict basis

Twenty narrowing/widening mutants applied by me, seventeen killed. All three of cycle 5's surviving
mutants (`my1`, `my4`, `my2`) are dead; cycle 4's `P2`/`P4` re-run and dead; C3-M1's §12.2 pin dies
under both a widening that would reach §9.1 secret material and a narrowing; the C3-B1 seam dies when
exempting exactly `--clear`; the §8.4 class dies under two fresh mutants including one nobody had
tried; the §9.6 divergent-skill loss and the `overlays_allowed` forbid each die under a narrowing.

C5-M1 is fixed and driven for `path` roots, with the takeover proved to reach the switch by a mutant
that strips it from the policy; the B2 class holds through rework 5's new write path in both the
copied and the linked adapter shape, foreign bytes intact, backup before the first write; every §1
path refusal fires through the new reinstall arm with the pin unmoved; the two precedence primitives
flip the emission order independently in the materialized bytes; §9.5's read-only rule holds under a
whole-home diff; not one of the 81 functions this stage adds is unreachable from production; the
Windows class does not recur under cross-compilation and a pattern sweep of all 27 stage files.

Locally: `go build`, `go vet`, `gofmt`, `golangci-lint run` (0 issues), `gate-selftest.sh` (94/94),
`no-broad-suppression.sh`, `ledger-consistency.sh` (212 rows), `go test -race` on all four stage
packages, and `go test -timeout 30m ./cmd/curator` (`ok 311.812s`) all exit 0.

## Hosted lanes and the candidate dispatch

**Every hosted lane was green on the exact head I accepted, `8b8aa041`.** `gh pr checks 61`,
run 34056559695: Lint, Naming gate, Interop conformance gate, Gate self-test ×3, Test
(ubuntu/macos/**windows**), Race (ubuntu/macos) — 11 of 11 pass.

**Candidate dispatch 34058365116 against the task authority `550579d`** with
`CI_REQUIRE_FULL_ROOT=1`, read from the `candidate-evidence-<os>` artefacts:

| Job | Result | suite-plan | failures |
|---|---|---|---|
| ubuntu-latest | success | `served=71 deferred=0 excluded=1` | 0 |
| macos-latest | success | `served=72 deferred=0 excluded=0` | 0 |
| windows-latest | success (on re-run) | `served=72 deferred=0 excluded=0` | 0 (5465 passing cases) |

The Windows job's first attempt failed on **infrastructure** — GitHub's annotation is "The hosted
runner lost communication with the server", it died mid-step with no evidence artefact. That is not
a test result, so I re-ran the single job rather than infer a pass or a failure. The re-run serves
and passes all six conformance families the default `SPEC_PIN` lane defers, on Windows, against the
authority. That was the stage's last remaining blind spot and it is now closed.

The candidate lane against curator-spec **main** (`87a0d00`) remains red on the 30
`path`-overlay-reconciliation subcases. Per the cycle-6 brief that is a decided, filed matter
(**TASK-260906-19gjyw**); I confirmed only that the bound is worded truthfully and that ledger rows
303–305 describe what their tests assert.

## Findings — two minors, no blocking, no major

- **C6-m1** (`repeat-of: cycle-5 C5-M1`) — `profile install <git-url> --use --takeover` after the
  §9.5 stop still accepts both flags, does nothing, and exits 0 saying `updated profile <name>`.
  **Driven on `origin/main` (`db444157`) with the identical result**, so it is a pre-existing trunk
  defect, not this stage's; the rework-5 brief scoped the fix to `path` roots deliberately, and the
  flag works correctly on a git *first* install. What belongs to this leaf is one false clause in
  the bound that carries it — "undrivable hermetically here" — which is wrong: it drives in ~30
  lines using the repository's own `insteadOf` fixture. Correct the clause and file the behaviour
  against trunk.
- **C6-m2** (`repeat-of: cycle-5 C5-m1`) — `reinstallActivation`'s first-install clause
  (`machine == "" → activate`) is unpinned: a mutant exempting exactly that clause survives both
  suites. Production is correct and operator-reachable (driven, including its refusal under a locked
  `require_current_profile`). One subtest closes it.

Observations, none blocking: two further pre-existing §8.4 collapse sites beyond the one cycle 5
named (`status.go:432`, `status.go:521`, alongside `purgeHomes`); the `CheckMachineUse` doc
enumeration omits rework 5's new seam caller while its universal claim stays true; one
stage-introduced skip reason (`no git on PATH`) is unclassified but fails closed on a branch dead on
all three runners; and `sameStoreTree`'s case-sensitive containment check is stage (b) code that
fails closed.

## Landing

**Stage (c) is safe to land.** Neither finding is a defect this stage introduced into behaviour a
green lane hides — C6-m1 is on trunk today and rejecting stage (c) would not remove it, and C6-m2 is
a missing subtest over correct, driven behaviour. Recommended follow-ups: a trunk task for the git
same-source reinstall dropping `--use`/`--takeover`; a trunk task closing the three §8.4 collapse
sites as a class; and one subtest for C6-m2 on the next touch of `reinstallActivation`.

Per the reviewer contract I supply no `commit_ack` and leave every `done` transition to the producer
side; `accept_cr` routes this element to `integrating`.
