# Review findings: environments.md revision 1.1, cycle 4 (F14/F15 edit at a68559b)

Subject: commit `a68559b` ("Copy the claude_code root-context surface in every mode") on top of `e45c5b6`. Scope `git diff e45c5b6..a68559b`: one file, `protocol/environments.md`, +19/−11. Read-only review. The draft worktree `/Users/iv/Developer/ReluxWorks/.worktrees/curator-spec-env-1-1` was removed while this review ran (PR 40 landed; curator-spec main now points at `a68559b`), so gates were re-run on a clean detached checkout of `a68559b` under this story worktree's `.temp/review-4/` and removed afterwards. Managed story workspace untouched at `3ce0d5a` (CR revision 2, accepted).

Verdict: **ACCEPT**. F14 resolved by fix (a) at all four anchors; F15 resolved. No blocking or major finding.

## F14 / F15 disposition

| Finding | Status | Evidence |
|---|---|---|
| F14 §8.1 mode defaults | resolved | New paragraph: the `claude_code` root-context surface "is always a copied regular file" in every mode and home, "not overridable by the registry or by machine configuration"; skills tree follows the home's mode; marker records the copy (section 8.2). |
| F14 §5.3 guard scope | resolved | Text now states the guard "keys on the memory type `User`, not on which directory `CLAUDE_CONFIG_DIR` names" and applies to "managed, `linked` in-place, and `copied`" homes. Re-extracted from the installed 2.1.261 binary (`claude --version` = 2.1.261): `if(t==="User"&&!C)try{let V=await ae().lstat(e);if(d===0&&V.isSymbolicLink()||(V.nlink??1)>1&&V.isFile())return[]}catch{}` — matches the stated rule. |
| F14 §8.2 `surfaces` | resolved | Entry records, for a `linked` or managed home, whether any entry is a copy rather than a link, "each recorded with its reason so that section 8.4 hash drift applies to the copy". |
| F14 §12.1 `in_place_mode.<env-id>` | resolved | Row now says the `claude_code` root-context surface is always copied whatever the value says. |
| F15 §5.3 wording | resolved | "a copied surface in the sense of section 8.1"; no longer names `copied` as the marker mode. |

## Attack pass (stale linked-`CLAUDE.md` assumptions)

Grepped `CLAUDE.md`, `symlink`, `link-target`, `hard link`, `linked` across the whole file at `a68559b`:

- §5.7 line 949 (`<home>/CLAUDE.md` surface path) — path only, no mode claim. OK.
- §8.1 `managed-home` bullet: "symlinks into the profile store, with copies where a surface or platform requires bytes" — the exception paragraph names the case. OK.
- §8.1 `linked` bullet: "symlink-with-copy-fallback discipline extended from skills to root context" — generic wording, immediately scoped by the new exception paragraph in the same subsection. OK (nit N2 below).
- §8.4 drift: keyed per surface kind (`linked` surfaces by link target or hash; `copied` and `managed-home` surfaces by hash), so the copy inside a `linked` home is hash-drift covered, as §8.2 now says. OK.
- §9.2 switching (lines 1600–1620): re-points command shims on machine-scope switches; no re-link claim for `CLAUDE.md`. OK.
- §9.5 onboarding/takeover: inventories "managed-surface paths that are already symlinks pointing outside the manager's store" — a copied surface is simply not a symlink; no false foreign-manager positive. OK.
- §9.2 `env unmanage`: "symlinks unlinked, copied files deleted" — covers both. OK.
- §10.1 verification: the link-target shortcut is explicitly "for a symlinked surface whose link targets an entry of the immutable profile store"; a copied `CLAUDE.md` therefore falls to the recorded hash. No contradiction. OK.
- §13: names the rewritten `agent-environment-marker-v1` schema; see nit N1.

No sentence in the file still assumes a linked `claude_code` root-context surface. No contradiction with Decision 0012/0013 or the diagnostics tables; the delta adds, renames, or withdraws no diagnostic. Cross-references in the delta (sections 5.3, 8.1, 8.2, 8.4, 12.1) resolve to existing headings.

## Nits (no action required for this acceptance; for the schemas batch)

- **N1 — §13 / schemas batch**: the marker schema note does not spell out that the §8.2 `surfaces` entry now carries a per-entry copy indicator with a reason (manager §5 fallback vs the always-copied `claude_code` root context). `draft/environments-schemas-1-1` must model that member; worth one clause in §13 when the text is next touched.
- **N2 — §8.1 `linked` bullet**: "extended from skills to root context" reads unconditional until the reader reaches the exception paragraph. Cosmetic.

## Gates

- `make validate` at `a68559b` on a clean detached checkout: 152 unittest OK, `go test ./tools/...` ok. Log: `.temp/review-4/make-validate.log` in the story worktree.
- Commit `a68559b`: "Good git signature" ECDSA `SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM`, author Ivan Oparin; the "No principal matched" line comes from the repo-local `gpg.ssh.allowedSignersFile` pointing at a deleted temp path, identical on the base — verifier config, not a signature defect. One file touched.
- Observation, not a finding on this change: a `git archive a68559b` export fails `make validate` with "vector digest mismatch for fixtures/byte-exact/subst.txt" while a real checkout of the same commit passes; `check-attr` shows only `text`/`eol=lf`. Flagging for the byte-exact vector owner (archive vs checkout difference in that fixture).

Routing: explicit ACCEPT at `to-review`. CR revision 2 is already accepted; no `accept_cr`. Not marked done.
