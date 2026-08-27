# Review verdict — CR-TASK-260825-2fy132-4 (revision 4)

**Verdict: ACCEPTED.**
Reviewer run: RUN (claude), 2026-08-25. Reviewed delta: `903af23..f7284154` (4 paths: CHANGELOG.md, LOGBOOK.md, README.md, docs/build-https.md). Story worktree state verified equal to the candidate tree (tracked files: empty diff; untracked `docs/build-https.md` blob `66ff91da` == candidate blob).

## Trees read (per the rev3 directive)

- **Docs candidate**: `.temp/STORY-260825-39h6vz/worktree` at base `903af23` + the CR-4 delta.
- **Authoritative delivery code**: `.temp/STORY-260825-32bopo/worktree` (base `903af23` + uncommitted delivery), designated authoritative by the orchestrator's landing map in TASK-260825-1d0eo5 notes. The primary checkout was **not** used for any claim. Files read there: `cmd/curator/main.go` (usage text, `cmdConfigBuildHTTPS*`, `operatorBuildHTTPSResolver`, `readBuildHTTPSToken`, `persistPromptedBuildHTTPS`), `internal/config/{buildhttps,buildssh,config}.go`, `internal/install/{buildhttps,buildhttpsprompt,buildsshprompt,external}.go`, `internal/buildrepo/{httpsbroker,admission}.go`, `internal/gitcred/gitcred.go`.
- **Spec**: `curator-spec/protocol/core.md` §6.1, §6.3, §12.2.

## AC 1 — docs match the shipped command output (verified live, not accepted from prior evidence)

Built the delivery worktree's `./cmd/curator` and ran the transcripts myself against a scratch `CURATOR_CONFIG` and scratch `HOME` with a file credential helper (`GIT_CONFIG_NOSYSTEM=1`), avoiding the operator keychain entirely. Transcripts: `.temp/TASK-260825-2fy132/review/transcript-0{1..5}.txt` + `help.txt` in the docs worktree.

Byte-identical to the doc examples:

- usage block of `config build-https --help` == the doc's command-surface block (all four subcommands, flag spellings, bracketing);
- `added build_https scope git.example.com/portals: token_env=PORTALS_TOKEN username=oauth2`;
- `replaced build_https scope git.example.com/portals: source=keyring`;
- `list`: `git.example.com/portals\tsource=keyring present=true` and `git.example.com/tools\ttoken_env=TOOLS_TOKEN username=oauth2 present=false` (tab separator, sorted);
- empty `list`: exit 0, empty stdout, stderr `curator: no build_https scopes are configured`;
- `login` via stdin on a non-terminal (doc's fallback path) exit 0; stored entry username is the scope-namespaced `curator-build-https:git.example.com/portals` (percent-encoded) in the operator credential store — "scope-namespaced entry" claim confirmed live;
- `present` flips true when the `token_env` variable is set at list time ("resolves now" claim);
- `remove` prints `removed build_https scope …` and deletes the stored keyring token (store empty afterwards); code path proves `git-credentials` material is never deleted.

Negative evidence (each documented refusal actually refuses; exit != 0, config not partially applied):

| Documented gate | Probe | Result |
| --- | --- | --- |
| exactly one source (CLI) | `add` with 0 and with 2 sources | usage error, exit 2 |
| scope grammar | `add GIT.example.com/x` | rejected with the grammar rule, exit 2 |
| env var name | `--token-env 9BAD` | rejected, exit 2 |
| empty token at login | `printf '\n' \| login` | `token must be a non-empty single line`, exit 2 |
| unknown entry field | config with `extra` | `has unsupported field(s): extra`, exit 1 |
| literal secret in `token` | config `"token":"hunter2"` | `must be one of git-credentials, keyring; secrets never live in the config`, exit 1 |
| zero/two sources (config) | both shapes | `requires exactly one of 'token' or 'token_env'`, exit 1 |
| empty field value | `"token_env":""` | `must be a non-empty string when present`, exit 1 |
| non-object entry | `"git.example.com":"keyring"` | `must be an object`, exit 1 |

Behavioural claims verified in the authoritative delivery source (not runnable without a live remote):

- precedence 1→4 exactly as documented (`resolveBuildHTTPS`): captured override (host-pinned or run-wide) → longest configured scope → terminal candidate prompt → anonymous; `CURATOR_BUILD_HTTPS_HOST` narrowing makes other hosts resolve as though the override were absent;
- capture at process entry (`CaptureBuildHTTPSSelection` reads the override and every configured `token_env` once; resolution never rereads ambient env);
- whole-closure resolution before the first repository fetch (`planExternalBuilds`: "Credentials are selected for the whole run before the first repository is reached"), with deterministic failure wording for each unavailable selected source;
- prompt semantics (`InteractiveBuildHTTPSResolver` + `readCredentialScope`): presence-only `Discover`, nothing read until explicitly selected, default `1` only when a host credential was detected, `t` token entry, scope persist or `r` this-run-only (no config/keyring write), abort `q` → `ErrBuildHTTPSAborted` stops the run; resolver wired **nil** for dry-run / non-terminal stdin / non-terminal stderr → those runs never prompt and continue anonymously;
- broker (`httpsbroker.go` + `admission.go`): private materialized copy under `manager-wrappers/`, secret-free state file `{host, username}`, secret only in the fetch invocation env (base env for all other git subprocesses), answers only the two exact Git prompts for the bound host, and a host-mismatch is refused before the fetch;
- fetch hardening backs the platform section: `credential.helper=` (empty), `GIT_TERMINAL_PROMPT=0`, `http.followRedirects=false` (the `.git`-suffix / 301 advice);
- longest whole-segment scope matching (`longestScope` + `identity.MatchesPrefix`) matches the doc's `portals` / `portals/app` / `portals-other` example.

## AC 2 — exposure warning wherever the override is documented

`CURATOR_BUILD_HTTPS_TOKEN` appears in exactly three delta surfaces — docs/build-https.md (dedicated Warning block), CHANGELOG.md entry, LOGBOOK.md 0057 SECURITY bullet — and each carries the identity-unbound warning with the `Spec core §12.2` citation. The shipped CLI usage additionally carries its own Disclosure warning (delivered by the code story; consistent wording). Spec check: core §12.2 is precisely the "MUST let the operator bind it to a host and MUST document the exposure" requirement; §6.1 (segment-aware allowlist matching) and §6.3 (canonical external repository identity) match what they are cited for.

## AC 3 — links and lint green

- Links: every relative target in the changed files exists (docs/build-ssh.md, docs/build-https.md, CONTRIBUTING.md, LICENSE, NOTICE, ci.yml, both TSVs, docs/implementation-plan.md); external `curator-spec` and releases URLs answer 200.
- Lint: `make lint` (golangci-lint 2.12.2) **exit 0, 0 issues, in this docs worktree** after `git submodule update --init --recursive` — this closes the environmental gap recorded in LOGBOOK 0057 EVIDENCE (the implementer's exit 2 was the missing tuitestkit replace target, not a finding).
- Naming gate (CI `naming-gate` job): the delta adds **zero** occurrences of the alternative implementation name; README keeps exactly one mention line. The task constraint "do not name other manager implementations" holds.

## Prior-cycle closure

rev3→rev4 interdiff (candidate trees `14423a9c..f7284154`) touches only docs/build-https.md and LOGBOOK.md and is exactly the rev3-required rework: the rev2 Resolution passage restored (with "private" dropped per G1), the four-item precedence restored, the host-narrowing sentence restored, and the LOGBOOK 0057 CORRECTION paragraph replaced with the one describing the shipped prompt. README.md and CHANGELOG.md are byte-identical to rev3, as required. All four H1 items are closed and were re-verified against the authoritative tree, not carried over.

## Minor observations (no rework required)

1. The Configuration table says username "defaults to `token`". Exact for `token_env`, `keyring`, and the env override; for `--git-credentials` with no explicit username, resolution actually sends the username stored in the operator's own Git credential (`material.Username`). The shipped CLI help makes the same simplification, so the doc matches the shipped surface; candidate for a one-line polish in a future docs pass, together with the citation-style nit (`Spec §12.2` in build-ssh.md vs `Spec core §12.2` here — same section).
2. LOGBOOK 0057 retains the original incorrect DOCUMENTATION BOUNDARY bullet followed by the CORRECTION paragraph — the logbook's record-then-correct idiom, accepted shape from rev2.

## Landing intel for the commit-owning mover (supplements the TASK-260825-1d0eo5 landing map)

- origin/main has **rewritten LOGBOOK.md** since 903af23 (−3034 lines; fresh Flight Logbook, entries numbered 04xx, via PR #27/#41). Entry 0057 cannot be patch-applied there — re-record its content in the new logbook's format; the numbering collision with the code worktree's 0052 noted in the landing map is now also a format migration.
- The naming gate on origin/main is green precisely because the cleaned LOGBOOK dropped the alternative-implementation mentions. Do not carry the base-tail LOGBOOK into the PR, or the gate goes red.
- README's Operator-credentials bullet on origin/main is byte-identical to the base line this delta replaces → applies cleanly; the same line is also edited by the code story (known collision). CHANGELOG on origin/main grew +37 lines → re-seat the new entry under Unreleased.

## Handoff

Reviewer-archetype run: no `commit_ack` supplied. Acceptance recorded via `accept_cr(TASK-260825-2fy132, revision=4, …)`; the orchestrator integrates the accepted revision and makes the `done` transition with `commit_ack=scope_committed`. No repository, board, or delivery-worktree file was modified by this review (scratch artifacts live under `.temp/TASK-260825-2fy132/review/`; the docs worktree gained only the initialized submodule checkout, which is untracked state required to run the lint gate).
