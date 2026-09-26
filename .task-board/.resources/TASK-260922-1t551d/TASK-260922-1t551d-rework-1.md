# TASK-260922-1t551d rework 1 (orchestrator, binding) — F-C1

Verdict rev1: CHANGES_REQUESTED (TASK-260922-1t551d_review-verdict-rev1.md; reviewer probes
attached: `TestReviewerDanglingExpectedTarget`, `TestReviewerLiteralCodexSelector`). Accepted:
stale-link removal, regular-file refusal, wrong-target detection, Pi agent root, no credential
reads, marker untouched. Continue from the revision-1 tree (no checkout/clean/stash); fix exactly:

R1 (high) — dangling links with the EXPECTED target are still reported current
(`managed.go:1552-1559` `checkPassthrough` compares the recorded target string but never checks
that the native target exists; `credential_link_test.go:354-366` asserts the contrary). Rule R2 of
the brief: dangling OR mis-targeted ⇒ detached with `environment_credential_conflict`-class
wording, in `StatusOf` and `Resolve`. Establish target liveness with Lstat/Stat on the target
(never read bytes); distinguish "target absent" from "inspection failed" (EACCES etc. is a
conflict/inspection diagnostic, not silence); fix the contrary test; add the reviewer's probe as a
committed row and a narrowing mutant that distinguishes expected-target-dangling from
wrong-target. Do not defer dangling coverage to F-C3.
R2 (high) — the Codex selector reader (`managed.go:552-564`) is a regexp for `key = "value"` only:
`cli_auth_credentials_store = 'keyring'` (single quotes), and other valid TOML spellings, read as
"absent" ⇒ `file` ⇒ isolated admitted under the operator-global keyring — the exact R3 hazard.
Parse `config.toml` with a real TOML parser (the repo already depends on one for other config —
reuse it; if not, `github.com/pelletier/go-toml/v2` or the BurntSushi parser already in go.mod);
absent key ⇒ `file`; `keyring`/`auto` ⇒ `environment_isolated_unsupported`; any other value ⇒
`environment_credential_unsupported`; unreadable/invalid TOML ⇒ refuse (fail closed, named
diagnostic), never "absent". Cover single-/double-/multi-line-quoted spellings and a key inside a
table at the production entry; commit the reviewer's three probe subtests; narrowing mutants per
refusal class.
Append "Revision 2" to results.md (both fixes with file:line, the new rows and mutants, the
corrected claim about dangling coverage in CHANGELOG/troubleshooting); republish only on a green
gate.
