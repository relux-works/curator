# land-https-credentials-composite — landing evidence

Pull request 43 merged as `9bba77de3`. 26 files, +4457 / -45.

## Assembly

Cut from `origin/main`, taking only this epic's source, test and documentation
files. The board-state edits carried by the delivery worktrees were left out,
as the landing map required. The primary checkout was deliberately not used as
a source: it holds an older partial copy without the prompt and resolver wiring.

## Two defects the Windows lane earned, both fixed on the branch

- `e8a16e2` — the broker wrapper was hard-linked to the running manager
  executable, so Windows saw one locked file identity and refused to unlink it
  during test cleanup. The wrapper is now an independent byte copy on Windows,
  returned only once both handles are closed; Unix keeps the hard-link path.
  The tests materialize through the production path and remove the wrapper
  before cleanup, so the lifecycle is asserted rather than assumed.
- `1c493e2` — the platform-case gate refused two unrecognised skip reasons.
  One was reworded onto vocabulary the ledger already carries; the other named
  a genuinely new condition and got one ledger row recording why tolerating it
  is safe, since the broker's admission and prompt boundary stay asserted on
  Windows.

## CI on the merged head

fmt, lint, vet, gate self-tests on three platforms, ledger, naming gate,
interop conformance gate, tests and races on macOS, Ubuntu and Windows: all
green. Merge state was CLEAN at merge time.

## Note for the next mover

The delivery worktrees under `.temp/STORY-260825-*` and the primary checkout
still hold the pre-landing copies of this work. They are superseded by
`9bba77de3` and should not be used as a source again.
