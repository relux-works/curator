# Review verdict — TASK-261001-3qugz9 rev1: ACCEPTED

Reviewer re-ran the spot checks from the review note on candidate tree 064f083a (base bd126a9a), using a throwaway HOME, a temp non-git dir and a binary built from current main (`go build ./cmd/curator`, exit 0).

| # | Check | Result |
|---|---|---|
| 1 | R2: `curator install` in a non-git dir skips local skills with exit 0; gitignore gate is the cause | CONFIRMED. bootstrap 0, project add 0, schema-2 local-path resolve 0 (lock in product root, bindings under `$HOME/.curator/source-bindings/`), install **exit 0** printing `generated paths are not ignored by git; missing entries: .agents/, .claude/skills/, .codex/skills/; skipped`; `.agents` absent; `git check-ignore` exits 128. Code: `internal/install/install.go:334-342` (gate comment at 334, `gitignore.Ensure` at 336) returns Status skipped; `cmd/curator/main.go` only fails on Status failed. |
| 2 | R1: v0.14.0 already has profile/env/run; launcher v0.1.0 tag-only | CONFIRMED. `git show v0.14.0:cmd/curator/main.go` lines 73-76 list profile, env, run; `git tag --contains e43dd2b8` = v0.14.0, rc.1, rc.2; v0.14.0 release has `curator_0.14.0_darwin_arm64.tar.gz`. `gh release list --repo relux-works/curator-agent-launcher` exit 0, no rows; tag v0.1.0 is an annotated tag object. |
| 3 | R6: credentials excluded from profiles; Codex file link is the operator's own auth.json | CONFIRMED. Spec 0400feab `environments.md:1386-1388` ("Credentials are never profile content and never managed surfaces"), `:40-46` (no memory/hooks/settings surfaces). `managed.go:535-540, 660-672`: keyring links nothing, file/auto link the native auth.json; unreadable/invalid TOML refuses, never reads as absence. |
| 4 | Three random citations | C4 `docs/cli.md:308` (project add registers alias/path, initializes Skillfile.json and .gitignore) OK; C9 `import.go:222` (closed revision-1 detected-surface inventory) OK; C7 `profile.go:150-157` (installed/activated/updated messages) OK. Also checked C16 `envregistry.go:28-32` (seed A/B) and C2 `umbrella.go:22` OK. |
| 5 | Draft notice does not overclaim | CONFIRMED. Marked DRAFT for tb-keeper confirmation, not sent; pins rc.3/rc.14 flagged as pending; says "op2-product should remain on its current setup"; states no full R1-R10 readiness claim; profile name `op2-product` flagged as proposed. |
| 6 | Hygiene | CLEAN. Grep over all four attached files for token/key/password/private-key/email patterns finds nothing; evidence JSON is argv/stdout of throwaway-HOME probes. No employer name found (reviewer only has the stated rule to go by). |

## Non-blocking residuals (do not block acceptance; carry into G2/G6 and the notice)

1. **Fixture validity / missing positive control (G2).** The probe's skill fixture is minimal (`SKILL.md` frontmatter name/description only). In a fresh *git* repo with the same fixture, `curator install` exits non-zero with `error: review: install marker is invalid for schema 5` (reviewer repro). So "skill absent in non-git root" is correctly attributed to the gitignore gate, but the artifact does not prove that lifting the gate alone would materialize that fixture. G2 acceptance must use a schema-5-valid skill fixture and a git-root positive control that installs the bytes, and the 1-2 day estimate should be read with that in mind.
2. **R6 wording.** "Codex file links only the operator's auth.json" — `auto` store also links the file (`managed.go:665-668`); only `keyring` links nothing. Row text is otherwise accurate.
3. **C6 line anchor.** The gate `gitignore.Ensure` call is `install.go:336`; `:334` is its comment. Cosmetic.
4. Host default Go caches were used for builds (the document states this bound itself).

Verdict: accepted. Producer side routes integration.
