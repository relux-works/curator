# BUG-260921-1fpaij brief (orchestrator, binding)

Story STORY-260915-3w11un (cross-PR Story; publish a Change Request as usual, the orchestrator
lands it by PR). `internal/gitignore` `Missing()` (gitignore.go:19-28) treats ANY `cmd.Run()`
error of `git -C <root> check-ignore -q <probe>` as "entry not ignored". `check-ignore -q`
exits 1 for "not ignored", 128 for a fatal error (not a repository, bad root), and a spawn
failure (`fork/exec …: permission denied`, git absent) is a `*fs.PathError`/`*exec.Error` —
today all three collapse into the policy message "generated paths are not ignored by git;
missing entries: …" and the install is skipped (PR #83 run 35550731645, Test macos-latest,
`internal/install :: TestAdapterLedgerCommitsAfterTheMirrorsItClaims` while `.gitignore` was
correct; same class as BUG-260921-30ycv0, a product spawn failure masked).

Rulings:
R1 Classify: exit status 1 (`*exec.ExitError` with code 1) = not ignored (unchanged); any other
outcome — spawn error, `exec.ErrNotFound`, exit 128 or other codes — is returned from
`Missing`/`Ensure` as an error wrapped with the operation and sanitized context (shape of
`gitops.run`: `git check-ignore failed: <stderr or errno>`; no environment, no probe path
beyond the entry name). Callers keep their fail-closed behaviour: the install refuses with that
error (result `failed`/refusal, not `skipped`), never with the "not ignored" message; check
every caller of `Missing`/`Ensure` (`cmd/curator`, `internal/install`) for how the error
surfaces and keep their existing error-class mapping.
R2 No retry in product code. Success path byte-identical (existing gitignore rows). Legacy v1
lane: declared bug fix (CHANGELOG `## Unreleased` → `### Fixed`); legacy goldens green.
R3 Evidence: rows with (a) an injected non-executable git on PATH (0755 script whose shebang
interpreter is 0644 — the fixture shape) → error carrying the spawn diagnostic, (b) a root that
is not a git repository → error with the exit-128 text, (c) exit 1 still "not ignored",
(d) production entry `install.Project` (or the CLI) with (a) → refusal text contains the spawn
diagnostic and NOT "generated paths are not ignored"; mutant restoring the conflation
(`if err != nil { missing = append… }`) fails (a), (b), (d). Windows: declared skips only via the
ledger vocabulary (`.github/ci/skip-classes.tsv`) if a sibling row already skips there.
results.md: before/after at the production entry, caller table, mutant table, ratio line (no
corpus change expected). Publish the Change Request only when the configured gate is green.
