# Producer brief: stage (c) rework 2

## Verdict

Review cycle 2 returned CHANGES REQUESTED: one blocking, one major, four minor. Read
`TASK-260906-1uf713_review-findings-stage-c-2.md` in full.

Rework 1 is otherwise accepted, and the reviewer says so explicitly: **all four cycle-1 blocking
findings plus M2, M3, m3 and m4 are fixed and driven**, the B1 seam enumeration holds (a single
`CurrentFile` writer), the B2 write sweep found no write landing outside the manager tree, the CLI
rows match `cli/curator.md` at `550579d`, and the consent gate is not config-reachable. Its own
mutants R1–R5 were all killed. **Every hosted lane is green on `0bcea201`, Windows included** — the
first time in this stage. Do not disturb any of that.

Both remaining findings are invisible to every lane. That is the point of them.

## C2-B1 (blocking) — a corrupt marker turns a managed home back into a "native" surface

`readRootSurface` skips a managed home only when `envmarker.Read` returns `(marker, nil)`. An
unreadable or malformed marker returns an error, the skip never fires, and `profile import` re-detects
curator's own generated root-context document as a native surface: exit 0, no warning, no loss entry,
and the reassembled module carries a full §5.1 generation header.

This is **cycle-1 B3 reached by a different addressing mode**, which is a repeat, and it is the same
§8.4 absence-versus-failed-read class you already fixed one function away in `readSkillsLedger`
(commit `ee6743a2`). Absence and a failed read are different facts; a fallback defined for absence
must not fire on a read failure.

Fix it as the class, not as the one call site: find every place in this stage that treats a failed
read of manager-owned state as "not present" and make each distinguish the two. Then drive `import`
with a corrupt marker, an unreadable marker, and an absent marker, and show the three different
outcomes. Narrowing mutant to quote: make the marker read collapse error into absence for exactly one
adapter, and show a named test failing.

## C2-M1 (major) — the lockable-subset gate is still per-knob, and one mutant reaches secret material

Cycle 1's surviving mutant (`key != "forms"`) now dies, but three new narrowing mutants survive with
and without `CURATOR_CONFORMANCE_ROOT`: `secret_material_waivers`, `xdg_seed_allowlist`,
`in_place_mode`. The reviewer drove the first end to end: with that mutant built, a system file
injects §9.1 secret-material waivers into the effective config through `mergeSystemEnvironments`
rule 3. A gate whose coverage is a hand-written list of knob names will keep losing this race.

**Derive the gate, do not extend the list.** The §12.2 lockable subset is a closed set the spec
states; make the test enumerate it from the same source the code does, so a knob added to one and not
the other fails. Prove it by re-applying all four mutants — cycle 1's and these three — and showing
each dies.

## Minors

- **C2-m1 — the §9.5 dotfile list is POSIX-only, and item 1 of the orchestrator's Windows finding is
  still unanswered.** You fixed the sweep half and left the question. Here is the answer, so you do
  not have to guess: the production lookup is correct — it resolves the operator home through
  `os.UserHomeDir` and joins portable relative paths — and the *list* is what makes the heuristic
  inert on Windows, because chezmoi stores state under `%LOCALAPPDATA%` there and home-manager is
  Nix-only. Whether §9.5's closed list is platform-specific is a **spec** question, now filed as
  TASK-260906-vlrjo1, and inventing Windows paths for it would be recording an unverified tool fact
  as normative. So: state it as an explicit bound in code and in the report, naming that task. Do not
  add a Windows location, and do not skip the case into silence.
- **C2-m2** — `config.CheckMachineUse` is left with no production caller while its comment claims it
  is the production path. Either route the seam through it or correct the comment; a comment that
  lies about reachability is how cycle-1 M1 happened.
- **C2-m3** — ledger row 304 still asserts a machine-declared `path` overlay. The spec fix is landing
  separately (curator-spec PR #47); until it does, the row must describe what the test actually
  asserts.
- **C2-m4** — `applyPlan`'s seed and marker writes lack the remove-first hardening. The reviewer is
  explicit that this is the manager tree only and *not* the B2 class, so it is not a data-loss risk —
  but the asymmetry is worth removing while the reasoning is fresh.

## Delivery

Small signed commits on `0bcea201` in `/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-c`.
Human identity, nothing written into the control root. Do not push and do not touch PR #61.

**Do not stage with `git add -A`.** The gate run itself produces artefacts — `go build ./tools/...`
leaves binaries, `make`-style runs leave caches — and a blanket stage has now put a `.pyc` on
curator-spec main once and a 5.4 MB executable into a reviewed commit once. Stage the paths you
changed and read `git status --short` before committing.

Gates, each a standalone process with its observed exit code: `go build ./...`, `go vet ./...`,
`gofmt -l cmd internal`, `golangci-lint run ./...`, `gate-selftest.sh`, `no-broad-suppression.sh`,
`ledger-consistency.sh`, `test-gate.sh` against **both** roots (materialized `SPEC_PIN`, and
curator-spec main with `CI_REQUIRE_FULL_ROOT=1`), `go test -count=1 -timeout 30m ./cmd/curator`.
**Run the two lanes sequentially**, never concurrently and never alongside a `-race` suite.

Attach `TASK-260906-1uf713_rework-report-2.md`: finding → resolution for all six; the derived-gate
mutant table with all four knob mutants dying; the three marker outcomes driven through `import`; the
stated bound for C2-m1 naming TASK-260906-vlrjo1; the two-root gate table; and honest bounds. Then
`task-board handoff TASK-260906-1uf713 --role developer`.
