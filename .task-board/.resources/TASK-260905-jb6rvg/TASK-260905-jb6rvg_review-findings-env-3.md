# Review findings: environments.md revision 1.1, cycle 3 (F11–F13 edit at e45c5b6)

Subject: draft worktree `/Users/iv/Developer/ReluxWorks/.worktrees/curator-spec-env-1-1`, branch `draft/environments-revision-1-1`, head `e45c5b6` on top of `db642b1`. Scope `git diff db642b1..e45c5b6`: one file, +12/−5. Read-only review.

Verdict: **CHANGES REQUESTED** (one major). F12 and F13 resolved; F11 resolved for managed homes only — the same verified guard also inerts the default `linked` in-place mode for `claude_code`, which §8.1 still prescribes.

## Findings

### F14 — major — §8.1 / §5.3: the `linked` in-place default for `claude_code` contradicts the verified skip

Quote (§8.1, unchanged by this edit): "the four adapters default to `linked` for their in-place surfaces". Quote (§5.3, new): "The same guard skips a user-level `$CLAUDE_CONFIG_DIR/CLAUDE.md` that is itself a symbolic link or a hard link (`nlink > 1`) while the key is unset".

What is wrong: the guard is not managed-home specific. Re-extracted from the 2.1.261 bundle (memory loader `aD`): for `type === "User"` with the external-includes key unset for the current project, `lstat` of the user CLAUDE.md returns early when `depth === 0 && isSymbolicLink()` or `nlink > 1 && isFile()`. In `linked` mode the native `~/.claude/CLAUDE.md` is exactly that depth-0 symlink into the store, and the native `.claude.json` carries `hasClaudeMdExternalIncludesApproved` only for projects where the operator accepted the interactive dialog. So a `linked` claude_code installation delivers no root context for every other project, silently, with the marker reporting current. The cycle-2 F11 fix line and the sprint addendum both said "the `linked` mode's `CLAUDE.md` surface must therefore be a regular file (copy), never a link, unless the seed sets the key"; the edit applied it to managed homes only. This is a contradiction between §8.1's default and §5.3's verified fact, not a missing cross-reference.

Fix (one of, stated in §8.1 mode defaults and echoed in §5.3): (a) `claude_code`'s root-context in-place surface defaults to `copied` as well (the skills tree stays linked; wording: "the `claude_code` root-context surface is `copied` in every mode"), with the §8.2 `surfaces` entry recording the per-surface copy so §8.4 hash drift applies to it; or (b) keep `linked` but state that the in-place `linked` root-context surface is inert for any launch directory whose native project entry lacks the key, and have `env status` report `environment_surface_inert` (new diagnostic, tabled once in §8.5). (a) is recommended: it is the same one-surface exception already taken for managed homes and needs no new diagnostic. Whichever is chosen, the §12.1 `in_place_mode.<env-id>` knob row should say the root-context exception is not overridable to `linked` for `claude_code`.

Verification note: the additional condition `egr()` in the bundle (`M1() !== "local-agent"`) is true for the CLI entrypoints the launcher uses, so the guard applies under both interactive and `-p` runs.

### F15 — nit — §5.3 wording "the `copied` mode of section 8.1"

Quote: "materializes its root-context surface as a regular file — the `copied` mode of section 8.1 — never as a link". `copied` is an environment mode (marker `mode` is exactly one of three, §8.2) and a managed home's marker stays `managed-home`; the sentence is a per-surface treatment. Fix: "as a regular file (a copied surface in the sense of section 8.1)". Cosmetic; no contradiction because §8.1's own default sentence is phrased per surface.

## F11–F13 disposition

| Finding | Status | Evidence |
|---|---|---|
| F11 §5.3 link/hard-link skip + §8.1 managed `claude_code` root-context `copied` | partially resolved | §5.3 rule present and verified against the 2.1.261 bundle (`aD`: `d===0&&isSymbolicLink() || nlink>1&&isFile()` under `t==="User"&&!C`); §8.1 managed-home exception present. Consistency: §8.1 `managed-home` definition already allows "copies where a surface or platform requires bytes"; §8.4 drift for `managed-home` surfaces is hash-based, §5.6 hash binding unaffected, §10.1 link-target shortcut applies only to symlinked surfaces; §9.2 linked switching text unaffected. Gap: the in-place `linked` mode (F14). |
| F12 §7.8 codex `-p` under `--strict-config` | resolved | Row says "under `--strict-config` too (0.153.2, sprint evidence)"; stat-before-launch rule unchanged. Not re-checked on this machine; the sprint reviewer's evidence stands, labeled as such. |
| F13 §7.9 `global_context_cap` adopted defaults | resolved | Both cells now "`none` recorded (docs-confidence)". |

## Gates

- `make validate` at `e45c5b6` (venv `.temp/venv` in the draft worktree): 57 schemas, 780 vector files, 152 unittest OK, go test ok. Log `.temp/review-3-make-validate.log` in the draft worktree.
- Cross-references in the delta (sections 5.3, 7.4, 8.1, 5.8) resolve to existing headings.
- Commit: `e45c5b6`, "Good git signature" ECDSA `SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM`, author Ivan Oparin; "No principal matched" comes from the repository-local `gpg.ssh.allowedSignersFile` pointing at a deleted temp path and appears identically on base `ec695ba` — a verifier-config issue, not a signature defect. Diff touches one file only; `tools/__pycache__/` untracked in the draft worktree.
- Diagnostics: no diagnostic added, renamed, or withdrawn by this delta.

Routing: `development` (major F14). The managed story workspace stays at the accepted CR revision 2; no `accept_cr` action.

## Recovery-run confirmation (RUN-260905-868670, 2026-09-05)

This run re-verified the cycle-3 verdict independently before recording it; the findings above stand unchanged.

- **F14 reproduced from the installed binary** (`/Users/iv/.local/share/claude/versions/2.1.261`, `claude --version` = 2.1.261). Extracted loader `aD`: `let C=o&&(t!=="User"||egr()); ... if(t==="User"&&!C)try{let V=await ae().lstat(e);if(d===0&&V.isSymbolicLink()||(V.nlink??1)>1&&V.isFile())return[]}catch{}` where `o` is the project entry's `hasClaudeMdExternalIncludesApproved` and `egr()` is `M1()!=="local-agent"`. The guard keys on the memory type `User`, not on which directory `CLAUDE_CONFIG_DIR` points at, so it fires for the native home exactly as for a managed one. §8.1 at `e45c5b6` still defines `linked` as "symlinks into the profile store ... extended from skills to root context" and keeps `claude_code` on the `linked` default; that surface is a depth-0 symlink and is skipped for every project lacking the key. Major stands.
- **F15** stands as a nit.
- **Gates re-read**: `make validate` log at `e45c5b6` (draft worktree `.temp/review-3-make-validate.log`) ends "Ran 152 tests ... OK" and `go test ./tools/... ok`; `git log --show-signature -1` shows a good ECDSA signature by Ivan Oparin; scope is `protocol/environments.md` only. The managed story workspace remains at `3ce0d5a` (CR revision 2, accepted), untouched.

Routing: `development` for the F14 edit (recommended fix (a): `claude_code` root-context surface `copied` in every mode, §8.2 `surfaces` entry records it, §12.1 `in_place_mode.claude_code` row notes the non-overridable exception).
