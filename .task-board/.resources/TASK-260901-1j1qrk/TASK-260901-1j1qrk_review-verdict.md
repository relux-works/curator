# Review verdict: TASK-260901-1j1qrk (csk surface-naming sweep) — ACCEPT

Reviewer run: RUN-260901-c9177c. Subject: CR-TASK-260901-1j1qrk-1 rev 1
(curator base `979fa36e` → candidate `5914a9dd`, LOGBOOK.md only) plus the
companion curator-spec worktree `draft/csk-surface-naming` head `4d55698`
(base `f8d7e7a`). Inventory resource `csk-naming-inventory.md` reviewed as the
primary surface. All checks below were rerun by the reviewer, not accepted
from producer claims, except where noted.

## 1. Inventory completeness — REPRODUCED

- Reran `grep -rn -i csk` in both worktrees. curator-spec: 711 post-rewrite
  hits; sorted diff against the producer's 712-line dump shows exactly one
  removed line — `protocol/core.md:1644` (`.csk-build.json`), the claimed
  rewrite. Nothing else differs.
- curator: 174 hits vs the producer's 171; content-diff (line numbers
  stripped) shows the 3 extras are the new LOGBOOK entry's own text
  (self-referential, expected). All other lines identical.
- Miscategorization hunt: inspected every csk hit in curator non-test Go code
  (19 hits: wire constants `MarkerName`/`LedgerName`/`LegacyManifestName`/
  `Name`, `.csk-install.json` stat in adapters, `csk-skill.json` audit scope,
  `CSK_PROJECT_ROOT`/`CSK_GLOBAL_ROOT` env-file emission — all wire or
  functional), every curator docs hit (README, CONTRIBUTING, CHANGELOG,
  docs/* — all name frozen identifiers or the fixture family), and every spec
  prose hit (core.md, profiles/manager.md, README, COMPATIBILITY, SECURITY,
  schemas README — all name frozen §1.1 identifiers, schema filenames, or the
  sanctioned single external-implementation reference at README:64). Filtered
  non-test Go hits through printf/errorf/usage/help patterns: zero csk in
  human diagnostic wording beyond frozen filenames. No surface hit was
  miscategorized as wire.

## 2. The one rewrite — CONFIRMED ILLUSTRATIVE

- `protocol/core.md:1644` now reads "MUST NOT infer portable paths such as
  `build-cache` or `.agent-build.json`" — correct in context (examples of
  paths a manager must NOT infer; the agent-* family spelling is exactly
  right for a hypothetical).
- `grep -rn csk-build` in both working trees: zero hits outside the new
  LOGBOOK record. `git grep csk-build` across **all tags** of both repos:
  in curator-spec it appears only as this same illustrative sentence
  (rc.5–rc.10); in curator, nowhere outside LOGBOOK. No schema, vector,
  fixture, or implementation references it.

## 3. Wire untouched — STRONGER THAN SPOT-CHECK

- `git diff f8d7e7a 4d55698 --stat` in the spec worktree: total delta is
  `protocol/core.md | 2 +-` (1 line). schemas/, conformance/, tools/,
  .github/, profiles/ are byte-identical to base by construction, which
  subsumes the 3-category spot-check. Working tree clean (only an untracked
  `tools/__pycache__/` from the validate run).
- curator CR delta: `LOGBOOK.md | 6 +` only. No code path consumes LOGBOOK.

## 4. Gates, signatures, LOGBOOK integrity — GREEN

- Spec, all three `make validate` components rerun with a
  jsonschema-equipped venv (system python3 lacks jsonschema, as the producer
  documented; bare `make validate` dies on ModuleNotFoundError before
  touching repo content): `tools/validate.py` exit 0 ("validated 57 schemas
  and 773 vector files"); `python -B -m unittest discover -s tools` exit 0
  (147 tests OK); `go test ./tools/...` exit 0. Note: the producer's evidence
  cited only validate.py; the reviewer ran the other two components.
- curator story worktree: `go build ./...` exit 0, `go vet ./...` exit 0.
  Full test evidence (57 non-cmd ok + 96 cmd/curator tests via three -run
  slices, all exit 0) accepted from the producer's run at the same base
  `979fa36e` — the only delta since is 6 LOGBOOK prose lines that nothing
  consumes.
- Signatures: `git verify-commit` good ECDSA
  (SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM) on both `4d55698`
  (spec) and `5914a9dd` (curator).
- LOGBOOK: entry inserted directly under the preamble, newest-first
  convention kept, previous entry (TASK-260901-2pho68) fully intact, content
  factually matches the inventory.

## Negative-evidence note

This change gates/refuses/validates nothing — it is a 1-line prose rewrite
plus a logbook record; the negative-test DoD items are not applicable, and
the review's adversarial angle was instead spent trying to defeat the
*classification* (miscategorized surface hits, hidden references to the
rewritten path in tags/fixtures). None found.

## Verdict

ACCEPT. `accept_cr(TASK-260901-1j1qrk, revision=1)` recorded; the element
parks at `to-review` for the orchestrator to checkpoint/integrate with
`commit_ack=scope_committed`. The spec commit `4d55698` on
`draft/csk-surface-naming` remains unpushed per brief — integration of that
repo's branch is the orchestrator's step.
