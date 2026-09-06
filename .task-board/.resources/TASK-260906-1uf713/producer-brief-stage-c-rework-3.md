# Producer brief: stage (c) rework 3

## Verdict

Review cycle 3 returned CHANGES REQUESTED — two blocking, one major, three minor — and answered the
landing question plainly: **not safe to land at `4df4d507`**. Read
`TASK-260906-1uf713_review-findings-stage-c-3.md` in full and unpack
`TASK-260906-1uf713_review-probes-stage-c-3.tgz`; both defects are reproducible from it against a
`curator` binary built from the head.

Everything else held. The reviewer attacked the composition weight rules, both precedence primitives
against materialised bytes, the import's loss list and reassembly, and all six cycle-2 findings, and
found them correct. Eight narrowing mutants on `parseSystemEnvironments` — the four named plus four of
its own — all die. Do not disturb any of that.

Fix in the order the reviewer gives, because the second defect is the one an operator hits first.

## C3-B2 (blocking, do this first) — `profile update` re-reads a `path` source

§1: "Installation copies the directory's tree into the profile store as an immutable snapshot **and
never reads the source directory again**: later edits to the source directory change nothing **until
the operator reinstalls**."

`updateLocked` (`internal/envprofile/envprofile.go:766-776`) does exactly that. Driven: install a
`path` profile, edit a file in the source, run `profile update` — the `state_sha256` pin moves, with
no reinstall. The `KindGit` arm eleven lines above short-circuits an exactly-pinned source and returns
the old lock; the `default` arm refuses the `local` root because it "does not move". A `path` root is
exactly pinned by construction, since §1 makes every requirement form on a `path` declaration
`profile_source_invalid`, so it belongs with those two.

The second consequence is worse than the first. §9.6 reassembles into a staging directory that
`importLocked` deletes, so every imported profile's recorded `source.Path` points at a directory that
no longer exists — and because `update` re-reads it, **one onboarding import breaks
`profile update --all` for the whole machine**, with `profile_source_path_missing`. That is the
routine maintenance command failing for the profile §9.6 exists to produce.

Resolve the root from the snapshot the store already holds under the old lock's `state_sha256`, not
from `source.Path`. Note the reviewer's warning: "return the old lock" is not the whole answer,
because `resolveOverlays` runs below the switch and a `path` root with `git` overlays is a legitimate
`update`.

The evidence gap under this matters as much as the defect: "the `path` source kind snapshots
immutably" is a definition-of-done row and **no committed test drives it**. `pathkind_test.go` has
nine tests, all refusals and the `.git` exclusion, and none edits a source after install. Cycle 1
verified immutability by hand against `profile sync` only, and the untested addressing mode is where
the defect lived — the same shape as stage (a) F1, cycle-1 B3 and cycle-2 C2-B1. The replacement test
edits the source and asserts the pin across `update`, `sync` **and** `use`, and asserts
`profile update --all` succeeds with an imported profile present.

## C3-B1 (blocking) — `profile use <name> --clear` skips the locked `require_current_profile`

`profile use <name> --clear` parses with no `--env`/`--target`, takes the `clearScope` branch of
`useLocked` with `scope == ""`, and reaches the §9.2 activation seam without the §12.2 gate. A machine
under a lock can be switched to another profile from a documented command with an undocumented flag
combination, at exit 0.

This is `repeat-of` cycle-1 B1, and the rework-1 brief said in terms that a second bypass would be a
repeat. So do not fix this the way B1 was fixed — a second `if` in a second branch invites a third
bypass. **Make it structural**: the gate belongs where a machine-scope switch is *constructed*, so
that a caller cannot reach the seam in machine scope without passing through it. Then make the
enumeration comment true — it currently promises coverage the code does not have, which is exactly
the failure mode cycle-1 M1 named.

Prove it by enumerating every caller that can reach the seam and driving each through `run()`, and by
a narrowing mutant that exempts exactly one of them.

## C3-M1 (major) — the lockable **set** is pinned by one hand-written knob

The consumer gate is genuinely derived and the reviewer confirmed it with eight mutants. But the
derivation moved the unpinned list rather than removing it:
`TestSystemV2LockableSubsetIsClassWide` decides each knob's branch with
`if LockableEnvKeys["environments."+knob]`, so it reads its expectations from the very map that
encodes §12.2, and a change to that map is invisible to it by construction. The only other pin is one
hand-written knob name — precisely what cycles 1 and 2 rejected one level down. A one-line widening
of `LockableEnvKeys` reaches §9.1 secret material with a fully green suite.

This is the **third consecutive cycle on one class**. The pin must **transcribe** §12.2 — the six keys
the specification states, written out independently — and be proved by both *widening* and
*narrowing* that map.

## Minors

- **C3-m1** — `profile compose add` is silent under a locked `overlays_allowed: false`, while
  `compose list` prints the manager §1 inert warning and resolution correctly empties the list. Warn,
  or state why the write row is right to stay silent.
- **C3-m2** — no committed test drives all four §6 weight rules disagreeing at once. The reviewer
  built the case and drove it: manifest 5, two agreeing requirers at 20, root map 30, machine overlay
  40 → lock records weight 40, `overlay: true`, `required_by: [mid1, mid2]`, which is correct. The
  implementation is right; the suite would not notice it stopping being right. Commit that test.
- **C3-m3** — noted, not a hole: `TestTakeoverWarnsDotfileHeuristic` drives `UseWithPolicy`, a
  production entry point the CLI calls with the same arguments. No action required; do not "fix" it
  into a CLI test if that costs the assertion.

## Delivery

Small signed commits on `4df4d507` in `/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-c`.
Human identity, nothing written into the control root. Do not push and do not touch PR #61.

**Do not stage with `git add -A`** — the gate run itself produces artefacts, and a blanket stage has
already put a `.pyc` on curator-spec main once and a 5.4 MB executable into a reviewed commit once.
Stage named paths and read `git status --short` before committing.

Gates, each a standalone process with its observed exit code: `go build ./...`, `go vet ./...`,
`gofmt -l cmd internal`, `golangci-lint run ./...`, `gate-selftest.sh`, `no-broad-suppression.sh`,
`ledger-consistency.sh`, `test-gate.sh` against **both** roots, `go test -count=1 -timeout 30m
./cmd/curator`. **Run the two lanes sequentially.**

Attach `TASK-260906-1uf713_rework-report-3.md`: finding → resolution for all six; for C3-B2 the driven
before/after of the pin across `update`, `sync`, `use` and `update --all` with an imported profile;
for C3-B1 the caller enumeration with each driven through `run()` and the exempting mutant; for C3-M1
the transcribed pin with both a widening and a narrowing mutant dying; the two-root gate table; and
honest bounds. Then `task-board handoff TASK-260906-1uf713 --role developer`.
