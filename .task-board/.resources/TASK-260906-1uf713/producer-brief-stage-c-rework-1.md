# Producer brief: stage (c) rework 1

## Verdict

Review cycle 1 returned CHANGES REQUESTED: four blocking, three major, four minor. Read
`TASK-260906-1uf713_review-findings-stage-c-1.md` in full and unpack
`TASK-260906-1uf713_review-probes-stage-c-1.tgz` — every finding was driven through the production
CLI and is reproducible from that archive. Do not re-derive them; reproduce one, then fix.

Work continues on `feat/agent-environments-stage-c` in
`/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-c`, currently `833918d2`, rebased onto
current main. Small signed commits on top.

## Before anything else — the attestation problem

Two of the review's findings are not about behaviour, they are about what the increment-2 report
claimed:

- The report lists the `path`-overlay AC row as **driven**. It is not reachable from any production
  surface; the only passing test builds a `Policy` in Go that `PolicyFromConfig` cannot produce (M1).
- The report claims **15 of 15** surfaces while silently dropping a bound increment 1 had declared —
  `env status` not reporting the locked `require_current_profile` (M3).

Stages (a) and (b) each lost a full cycle to exactly this class: a gate that exists, compiles, returns
the right value when a test calls it directly, and is never reached from `run()`. An AC row is driven
only when the production entry point drives it. A bound once declared is dropped only by fixing it,
never by omitting it from the next report. This revision's report will be read against that standard
first.

## Blocking

### B1 — `profile install --use` bypasses the locked `require_current_profile`

`Config.CheckMachineUse` has exactly one production caller, `cmdProfileUse`. `Install` performs the
same §9.2 machine-scope switch on `--use` and on first install and never consults the gate, so with
the key locked to `acme`, `profile install ./x --use` activates `p5` at exit 0 with no warning. Same
on a fresh machine, where the first install of any profile activates it.

Fix at the **activation seam** — the §9.2 machine-scope switch itself — not by adding a second call
site. Then find every other path that reaches that seam and confirm the gate covers them all;
`overlays_allowed` and `require_current_profile` are the entirety of revision 1's fleet-policy
surface, so a second bypass is a repeat finding. Narrowing mutant to apply and quote: make the gate
admit exactly the install path, and show a CLI-level test failing.

### B2 — `--takeover` of a foreign-manager symlink writes THROUGH the link

`materializeOne` writes the copied surface with `os.WriteFile` (`switch.go:498`, and the
symlink-fallback copy at `:509`) without removing an existing symlink first, and `os.WriteFile`
follows symlinks. Stage (c) is what first routes a foreign-manager symlink into that write — before
`--takeover` the entry was refused unconditionally. So after a "take over with backup" the surface is
**still a symlink**, and curator has overwritten a file it never inventoried, at a path outside every
managed home. In the field that is the other manager's own source of truth, which its next apply
propagates. §9.5 offers "abort, or take over with backup — never a silent absorption"; this absorbs
in the wrong direction.

This is the most serious finding in the set: it destroys data outside the manager's own tree. Fix the
write path so a copied surface never follows a link — remove first, exactly as `replaceLink` already
does for the linked adapters, which the reviewer confirmed correct. Then check the whole delta for
any other write that can land on a path the manager does not own.

`TestForeignSymlinkStopsSwitch` asserted only `result.OK` after the takeover — positive-path evidence
sitting directly over the hole. Its replacement must assert that the surface stopped being a symlink
**and** that the foreign file's bytes are unchanged.

### B3 — `profile import` swallows curator's own managed root-context files as "native"

`readRootSurface` has no marker check, so the import detects curator-managed root-context files as
native surfaces and reassembles four modules each carrying a full §5.1 generation header. §9.6's
detected-surface list is closed and is about *native* context; a managed surface is recorded in the
marker and reaches managed state through §9.2, never through import.

### B4 — same-named skills across adapters collapse silently, with no loss-list entry

`reassembleImport` keys `requires.skills` by bare skill name (`import.go:434`), so two adapters
carrying `foo` at different commits produce one entry, last writer wins by ascending environment
identifier, and the dropped declaration appears in neither the manifest nor the loss list. §9.6
requires one entry per mapping skills entry, and the loss list must name every loss. A machine with
one skill installed at different commits in two agent homes is the ordinary onboarding case, and the
entire point of §9.5 is reaching managed state without loss.

Decide from §9.6 what the correct outcome is — two entries, or one mapping entry plus a recorded loss
— and say which sentence decides it. Do not pick whichever is easier to implement.

## Major

### M1 — `path` overlays unreachable, and the spec contradiction behind it

Three gates compose into a dead feature: the config reader requires a form on every overlay,
`resolveOverlay` refuses any form on a path source, and `profile compose add` requires a form. The
root cause is a genuine spec contradiction — §6 promises a `path` overlay, §1 makes a form on a `path`
declaration `profile_source_invalid`, and §12.1 plus `manager-config-v2` require a form on every
overlay.

**The orchestrator has filed and is fixing that contradiction as TASK-260906-3x0w4y** (the form
becomes required for a `git` source only). Do not work around it here and do not invent a
discriminator of your own. Your job in this revision is: stop reporting the row as driven, carry it as
an explicit stated bound naming the spec task, and fix the ledger row
`internal/config TestPathOverlayDeclarationParses`, which currently states the exact opposite of what
the test asserts and the reader does. If the spec fix lands before you finish, consume it and make the
path overlay reachable from all three surfaces — the orchestrator will tell you; otherwise the bound
stands and `cmdComposeList`'s dead `default: form = "path"` branch is named in the report.

### M2 — the unreadable-path diagnostic is shadowed, and it refutes a declared survivor

`stateForPath` wraps every `EnsureState` error with `profile_source_invalid`, so
`profile_source_path_unreadable` never leads and the doubled prefix of m3 falls out of the same
wrapper. §1.1 requires the specific diagnostic.

Note what this does to increment 2's mutant table: M3 was declared a **survivor** on the rationale
that `pathManifestDiag` precedes the store so the store branch is unreachable except on a race. The
wrapper is why no test failed. Once the wrapper is gone, re-run that exact mutant and report what
happens — if it now kills, the survivor was an artefact of the defect, and say so plainly.

### M3 — `env status` does not report the locked `require_current_profile`

§12.2 says `env status` reports the requirement. Increment 1 declared this a bound; increment 2
dropped the bound and claimed full coverage. Implement it and drive it through `run()`.

## Minor

- **m1 — a surviving narrowing mutant, the reviewer's own.** Narrowing the `parseSystemEnvironments`
  allowed-knob gate to admit exactly `environments.forms` leaves the whole tree green, with and
  without `CURATOR_CONFORMANCE_ROOT`. The published family covers `current_profile` and
  `overlay_default_weight` but not `forms`, so the gate's coverage is per-knob-name rather than
  class-wide. Make the test class-wide — every §12.2 lockable key and a representative non-lockable
  one — so no single knob can be admitted.
- **m2 — a ledger read failure is treated as "nothing ledgered."** `readSkillsLedger` returns an empty
  recorded set when `.csk-managed.json` cannot be read or does not decode, so an entry gets imported
  that the manager cannot prove is unledgered. §8.4's absence-versus-failed-read distinction applies.
- **m3 — doubled diagnostic prefixes.** Falls out of M2's wrapper; confirm it is gone.
- **m4 — `compose list` prints declarations a locked `overlays_allowed: false` has emptied.**
  Resolution correctly drops them and the manager §1 warning fires, but the informative row gives no
  indication the declaration is inert. Judgement call: fix it or state why the informative row is
  right to print the file's contents verbatim.

## Delivery

Small signed commits on `833918d2`. Human identity, no `LOGBOOK.md`, nothing written into the control
root. Do not push and do not touch PR #61 — the orchestrator owns its head.

Gates, each a standalone process with its observed exit code: `go build ./...`, `go vet ./...`,
`gofmt -l cmd internal`, `golangci-lint run ./...`, `bash .github/ci/gate-selftest.sh`,
`bash .github/ci/no-broad-suppression.sh`, `bash .github/ci/ledger-consistency.sh`,
`bash .github/ci/test-gate.sh` against **both** roots (the materialized `SPEC_PIN` tree, and
curator-spec main with `CI_REQUIRE_FULL_ROOT=1`), and `go test -count=1 -timeout 30m ./cmd/curator`.

**Run the two lanes sequentially, never concurrently and never alongside a `-race` suite.** Increment 2
lost two attempts to that contention over the machine-wide Go test lock, and the resulting red is
indistinguishable from a real regression. A gate failure seen only under concurrent lanes is not
evidence of a defect — re-run it solo before reporting it as one.

Attach `TASK-260906-1uf713_rework-report-1.md`: a finding → resolution table for all eleven; for each
blocking fix, the narrowing mutant and the named test it kills; the re-run of increment 2's M3 mutant
after the M2 wrapper is gone; the two-root gate table; and bounds that are honest, including any row
you cannot drive. Then `task-board handoff TASK-260906-1uf713 --role developer`.
