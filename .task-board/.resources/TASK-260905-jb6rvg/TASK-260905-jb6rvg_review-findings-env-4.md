# Review findings: environments.md revision 1.1, cycle 4 (F14/F15 edit at a68559b)

Subject: draft worktree `/Users/iv/Developer/ReluxWorks/.worktrees/curator-spec-env-1-1`, branch `draft/environments-revision-1-1`, head `a68559b` on top of `e45c5b6`. Scope `git diff e45c5b6..a68559b`: one file (`protocol/environments.md`), +19/−11. Read-only review; managed story workspace untouched at CR revision 2 (`3ce0d5a`, accepted).

Verdict: **ACCEPT** (no blocking/major; two nits for a later edit).

## F14 / F15 disposition

| Finding | Status | Evidence |
|---|---|---|
| F14 (major) — `linked` in-place default for `claude_code` contradicts the verified link/hard-link skip | resolved by fix (a) | §8.1 mode-defaults paragraph: "One per-surface exception holds in every mode and every home and is not overridable by the registry or by machine configuration: the `claude_code` root-context surface is always a copied regular file ... the `claude_code` skills tree follows the home's mode as usual, and the marker's `surfaces` entry records the copy (section 8.2)". §5.3: guard "keys on the memory type `User`, not on which directory `CLAUDE_CONFIG_DIR` names", surface "materialized as a regular file ... in every mode and every home — managed, `linked` in-place, and `copied`". §8.2 `surfaces`: "for a `linked` or managed home — whether any entry is a copy rather than a link: the manager §5 fallback, or the always-copied `claude_code` root-context surface of section 8.1, each recorded with its reason so that section 8.4 hash drift applies to the copy". §12.1 `in_place_mode.<env-id>` row: "the `claude_code` root-context surface is always copied whatever this value says". All four anchors named in the fix are present and agree. |
| F15 (nit) — "the `copied` mode of section 8.1" | resolved | §5.3 now reads "a copied surface in the sense of section 8.1". |

## Attack pass (residual assumptions of a linked `CLAUDE.md`)

Grepped `CLAUDE.md`, `linked`, `symlink`, `link-target`, `re-point` across the whole file.

- §5.6 hash binding: per-surface, mode-agnostic; a copied root-context surface hashes like any other. No change needed.
- §8.1 `linked` definition ("symlinks into the profile store ... extended from skills to root context"): generic definition; the exception paragraph in the same section explicitly overrides it. Not a contradiction (nit N-a below).
- §8.4 drift: phrased per surface ("For `linked` surfaces ... For `copied` and `managed-home` surfaces, hash"); the §8.2 edit routes the copy to the hash branch. Consistent.
- §9.2 switching (lines 1607, 1723): re-points machine shims only; does not mention `CLAUDE.md`. Unaffected.
- §9.5 onboarding/takeover: "managed-surface paths that are already symlinks pointing outside the manager's store" is foreign-manager evidence; a native symlinked `CLAUDE.md` from another manager is correctly detected, and the manager's own copy is a regular file. Consistent.
- §10.1 link-target shortcut: scoped to "a symlinked surface whose link targets an entry of the immutable profile store"; the copied surface falls to the marker's hash. Correct, though implicit (nit N-b).
- §13 marker schema note: names `agent-environment-marker-v1` as rewritten without enumerating `surfaces` members; the new copy-reason member is within that rewrite. Still right.
- §7.4 registry row and §7.3 descriptors: no linked-`CLAUDE.md` assumption.
- Decisions 0012/0013: no rule touched; no diagnostic added, renamed, or withdrawn.

## Nits (not blocking; for a later edit)

- **N-a §8.1 `linked` bullet**: add "(except the `claude_code` root-context surface, below)" after "extended from skills to root context" so the definition does not have to be read against the later paragraph.
- **N-b §10.1**: one clause making the copied case explicit: "a copied surface — including the always-copied `claude_code` root context — is verified by its recorded hash".

## Gates

- `make validate` at `a68559b` (PATH `.temp/venv/bin` in the draft worktree): 152 unittest OK, `go test ./tools/...` ok.
- Cross-references in the delta (sections 5.3, 8.1, 8.2, 8.4, 12.1) resolve to existing headings.
- Commit `a68559b`: "Good git signature" ECDSA `SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM`, author Ivan Oparin. The "No principal matched" line is the repository-local `allowedSignersFile` pointing at a deleted temp path, identical on base `ec695ba`; not a signature defect. Diff touches one file; `tools/__pycache__/` untracked in the draft worktree only.

Routing: explicit ACCEPT at `to-review`. CR revision 2 already accepted; no `accept_cr`. Not marked done.
