# Review verdict: TASK-260825-2fy132 (https-credential-docs) — CR-TASK-260825-2fy132-2 rev 2

**Verdict: changes requested → to-dev.** One blocking sentence; everything else verified against the shipped implementation and preserved.

Reviewed delta: `git diff 903af23a a66b0b9c` (CHANGELOG.md, LOGBOOK.md, README.md, new docs/build-https.md; patch sha256 `306df91b…`).
Shipped ground truth: the accepted epic delivery in `.temp/STORY-260825-32bopo/worktree` (stories 32bopo and 37cz7x `done`; the primary checkout lags it — it lacks `buildhttpsprompt.go` and the resolver wiring, so the story worktree is the newer accepted state; landing is TASK-260825-1d0eo5). All transcripts below were rerun by this review against `go run ./cmd/curator` from that checkout with an isolated `CURATOR_CONFIG`, isolated `HOME`, and `GIT_CONFIG_NOSYSTEM=1` + a scratch `credential.helper store`; artifacts in `.temp/TASK-260825-2fy132/reviewer-rev2/`.

## Rev1 findings: closed

- **F1 (prompt denied)** — fixed. The rewritten "Resolution, precheck, and candidates" section, precedence items 3–4, the host-pin sentence, and the LOGBOOK 0057 correction now describe the shipped terminal candidate prompt. Each statement re-verified against `buildhttpsprompt.go`, `operatorBuildHTTPSResolver` (dry-run/non-TTY → nil resolver), `resolveBuildHTTPS`, and `TestTheCredentialPromptIsWiredOnlyWhereAnOperatorCanAnswerIt`.
- **F2 (literal `\t`)** — fixed. `od -c` on docs/build-https.md lines 102–103 shows real `\t` bytes; live `list` output matches the fenced block byte-for-byte, including `present=true` for a keyring scope after a real piped `login` (this review exercised the login→store→read-back→list chain that rev1 could not).
- **F3 (README line)** — fixed. The line keeps 32bopo's broker sentence and links both pages; the composite landing has an obvious final form. (Landing still reconciles the same-line edit in the 32bopo delta and the append-point/numbering of LOGBOOK entries — 32bopo's worktree numbers its entry 0052 while the primary checkout already holds 0052–0056; this CR's 0057 is the next free number against the primary tail.)

## Blocking finding G1 — "private" misstates who gets the candidate prompt

docs/build-https.md, "Resolution, precheck, and candidates":

> "On an operator terminal, however, an uncovered **private** repository first opens a candidate prompt before any fetch."

Shipped behavior: the prompt opens for **every** uncovered HTTPS repository. Curator cannot know whether a repository is private before fetching — `resolveBuildHTTPS` collects every planned HTTPS row with no override and no matching scope into the request list, and the interactive resolver asks about each one (`internal/install/buildhttps.go` missing-rows path; `InteractiveBuildHTTPSResolver`). The menu offers only the detected host credential / `t` enter a token / abort — there is no "continue anonymously" choice, and an empty answer with no detected credential re-asks (`readBuildHTTPSChoice`). Anonymous continuation exists only where no resolver exists (headless, non-terminal, dry-run).

Why it matters: an operator with an ordinary public HTTPS build repository reads this page as "a terminal install of a public repository fetches anonymously". In reality the run stops at `build_repository_https_credential_missing:` with no anonymous option; following the page, their model of when a credential can be solicited is wrong — the same mis-modeling class rev1 blocked on, in the same section this task exists to document. The delta already states the correct rule twice: precedence item 3 says "the candidate prompt for an uncovered repository" (no qualifier), and this CR's own LOGBOOK 0057 correction says "offers **every** uncovered HTTPS repository". The page contradicts both.

Fix (one sentence): drop "private" — e.g. "On an operator terminal, however, an uncovered repository first opens a candidate prompt before any fetch." Optionally anchor the preceding transport sentence to the non-interactive paths ("so an uncovered repository can be fetched anonymously — and headless, non-terminal, and dry-run runs do exactly that"). No other change is needed.

## Non-blocking notes (do not rework; recorded for precision)

- N1: The username row says it "defaults to `token`". For a `git-credentials` source with no `username` field, the shipped fetch sends the stored credential's own username, falling back to `token` only when the helper names none (`resolveConfiguredBuildHTTPS`). The shipped `--help` carries the same simplification, so the doc matches the shipped surface; a future polish could add the inheritance clause to both.
- N2: The warning's "offered to every **private** HTTPS build repository host" matches the shipped help verbatim; strictly the secret goes to any host that issues an auth challenge (Spec core §12.2 says "every host in the closure"). Defensible as written since an unchallenging host never receives the secret; candidate for a joint help+doc tightening later, not for this CR.
- N3: LOGBOOK 0057 keeps the false rev1 boundary paragraph with an explicit superseding correction — truthful as an append-style record; the landing mover may keep or squash per logbook convention.

## Verified (rev2, rerun by this review)

- Transcripts byte-identical to the page: first `add` → `added build_https scope git.example.com/portals: token_env=PORTALS_TOKEN username=oauth2`; `--keyring` re-add → `replaced … source=keyring` (entry fully replaced); empty `list` → exit 0, empty stdout, stderr `curator: no build_https scopes are configured`; piped `login` (non-TTY one-line stdin path) → keyring source recorded, token stored under `curator-build-https:<scope>` via the operator's git credential machinery with read-back proof; populated `list` → exactly the doc's two lines including real tabs and `present=true`/`present=false`; `remove` deletes the manager-stored token only for a keyring selection and errors on an unconfigured scope.
- Config/grammar claims match `internal/config/buildhttps.go` + shared `buildssh.go` grammar: lowercase host (`scopeHostRE`), segment charset, whole-segment longest match (`longestScope`/`identity.MatchesPrefix`), strict parsing (non-object, unknown field, empty value, invalid env name, zero-or-two sources all reject; config not partially applied), `CURATOR_CONFIG` path override.
- Resolution/precedence claims match `resolveBuildHTTPS`/`CaptureBuildHTTPSSelection`: override → longest scope → terminal prompt → anonymous; host pin makes other hosts resolve as if the override were absent (test `TestBuildHTTPSHostPinMakesOtherHostsResolveWithoutTheOverride`); the three per-source precheck failures are deterministic and pre-fetch (`planExternalBuilds` orders resolution before any acquisition); token_env values frozen at capture.
- Prompt claims match `buildhttpsprompt.go` + tests: presence-only discovery (`gitcred.Discover`), selection before any read of the chosen source, scope persist vs `r` this-run-only (never writes config or credential storage), abort → `ErrBuildHTTPSAborted` stops the run; dry-run/non-TTY gating in `operatorBuildHTTPSResolver`.
- Platform/broker claims match `gitcred.go`, `httpsbroker.go`, `admission.go`: all credential IO via `git credential fill|approve|reject` under the operator's configured helper; private broker copy under `manager-wrappers/` answering only the two exact prompts for the bound host; state file host+username only (0600); secret only in the fetch process env; host mismatch rejected at admission; `GIT_TERMINAL_PROMPT=0`; `http.followRedirects=false` backs the `.git`/301 guidance.
- Spec citations verified in the curator-spec checkout: §6.1 segment-aware matching, §6.3 external repository identity, core §12.2 requires exactly the documented bind-and-disclose. Exposure warning present at every place this delta documents the override (doc warning block + CHANGELOG); README does not mention the override; no other manager implementation named anywhere in the delta.
- Links: README → both doc pages exist; spec URL matches the repo convention. Lint: `golangci-lint run` in this docs worktree (after `git submodule update --init --recursive agents/skills/skill-go-testing-tools`) → `0 issues`, exit 0, rerun by this review. No Go files change in this CR.

Accepted from the board without rerunning: the sibling stories' own acceptance evidence and their interactive-resolver test runs. Not exercised live: a real-TTY prompt (spawned session; covered by the cited tests and code reading). Environmental note, not a finding: in this sandboxed session the system-gitconfig `osxkeychain` helper hangs on keychain access; curator's own 15s `gitcred` timeout handled it exactly as designed, and the login transcript above was produced against an isolated file-store helper.

## Routing

`to-dev` for a one-sentence documentation fix (G1). Everything under "Verified" is accurate as written and must be preserved; N1–N3 are optional polish, not rework.
