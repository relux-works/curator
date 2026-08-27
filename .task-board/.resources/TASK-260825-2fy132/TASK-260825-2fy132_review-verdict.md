# Review verdict: TASK-260825-2fy132 (https-credential-docs) — CR-TASK-260825-2fy132-1 rev 1

**Verdict: changes requested → to-dev.**

Reviewed delta: `git diff 903af23a 41fff7fd` (CHANGELOG.md, LOGBOOK.md, README.md, new docs/build-https.md).
Shipped implementation ground truth: the accepted epic delivery in the delivery checkout
`.temp/STORY-260825-32bopo/worktree` (stories STORY-260825-32bopo and STORY-260825-37cz7x are `done`;
their union is that worktree's working state; story branches all sit at 903af23, landing is
TASK-260825-1d0eo5's job). All transcript checks below ran against a binary built from that checkout
(`go build ./cmd/curator`), with an isolated `CURATOR_CONFIG`; artifacts in
`.temp/TASK-260825-2fy132/review/`.

## Blocking finding F1 — the page denies the shipped interactive candidate prompt

docs/build-https.md ("Resolution, precheck, and candidates") states:

> "There is no install-time credential discovery or interactive candidate prompt for HTTPS: the
> operator explicitly chooses a source with `add` or provides a token through `login`."

and, unconditionally, "an uncovered repository is fetched anonymously".

The accepted delivery ships exactly that prompt, and it was explicitly required by
TASK-260825-3kb532 ("install-precheck-and-candidates", status `done`): "On an operator terminal an
unmatched repository prompts with detected candidates: the operator's existing credential for that
host, and entering a token now. … Mirror the SSH prompt shape."

Shipped evidence (delivery checkout):

- `internal/install/buildhttpsprompt.go` — `InteractiveBuildHTTPSResolver`, menu: existing Git host
  credential (default) / `t` enter a token / abort; scope question with persisted or this-run-only
  (`r`) answers; prompt banner `build_repository_https_credential_missing: …`.
- `cmd/curator/main.go:1341` — `operatorBuildHTTPSResolver`: resolver active when stdin+stderr are
  terminals and the run is not a dry run; wired at `main.go:538` and `main.go:1300`.
- `internal/install/buildhttps.go` — `resolveBuildHTTPS` calls `reader.Discover` (presence-only
  install-time discovery, `gitcred.Access.Discover`) for each unmatched host and presents
  `BuildHTTPSRequest`s; aborting the prompt raises `ErrBuildHTTPSAborted` and stops the run — it is
  not an anonymous fallback. Anonymous is the fallback only when no resolver exists
  (headless / non-TTY / dry run).
- Accepted tests cover the prompt and its gating: `internal/install/buildhttpsprompt_test.go`,
  `cmd/curator/main_test.go:1035` (dry run → nil resolver), `:1041`, `:1175`.
- The same accepted delta updates docs/build-ssh.md with the shared this-run-only (`r`) scope
  machinery the HTTPS prompt reuses.

Sentences needing rework (all one root cause):

1. The "no install-time credential discovery or interactive candidate prompt" sentence — replace
   with the real behavior: on an operator terminal, a private HTTPS repository no source covers is
   prompted with detected candidates (existing host credential, entering a token now); discovery
   lists presence only; nothing is used without explicit selection; a this-run-only choice never
   reaches the config; abort stops the run. Headless, non-terminal, and dry-run runs never prompt
   and continue anonymously.
2. "Precedence and exposure warning" item 3 ("Anonymous HTTPS when neither applies") — on a
   terminal the candidate prompt sits between a non-matching scope and anonymous.
3. "When `CURATOR_BUILD_HTTPS_HOST` is set, repositories on other hosts … may use a matching
   configured scope or remain anonymous" — may also be prompted.
4. LOGBOOK.md entry 0057 records the same false claim as a delivery fact ("The delivered surface
   has no install-time HTTPS credential discovery or interactive candidate prompt … instead of
   borrowing the SSH prompt behaviour"). Future sessions treat the logbook as evidence; the rework
   must correct this (follow the logbook's own correction conventions).

This is the exact surface the task description asked to document ("the precheck and candidate
behaviour"), and it is security-relevant: an operator told a terminal install can never prompt will
mis-model when credentials can be selected and persisted.

## Minor finding F2 — list transcript separator does not match shipped output

The `list` transcript renders the separator as the two literal characters `\t`. The shipped command
emits a real TAB byte: `fmt.Printf("%s\t%s present=%t\n", …)` (`cmd/curator/main.go:2251`),
live-verified with `cat -vet` → `git.example.com/portals^Isource=keyring present=true`. AC says docs
match the shipped command output. Fix: put the actual tab character in the fenced block (optionally
note the tab separator in prose).

## Landing note F3 — README line collision (for the next producer and the orchestrator)

This CR rewrites the same README.md "Operator credentials" line that the accepted STORY-260825-32bopo
delta also rewrites, with different text: 32bopo's version adds "Private HTTPS fetches use a
manager-owned, host-pinned askpass broker; public HTTPS remains anonymous." and links only
docs/build-ssh.md; this CR links both pages but drops the broker sentence. TASK-260825-1d0eo5
assembles both deltas; the rework should produce one final line (links to both pages, ideally
keeping the broker sentence) so the composite landing does not have to invent the merge.

## Verified — everything else matches the shipped implementation

Live-verified against the shipped binary (isolated `CURATOR_CONFIG`; transcripts in
`.temp/TASK-260825-2fy132/review/`):

- `add … --token-env PORTALS_TOKEN --username oauth2` → `added build_https scope
  git.example.com/portals: token_env=PORTALS_TOKEN username=oauth2` — byte-identical to the doc.
- `add … --keyring` on the same scope → `replaced build_https scope git.example.com/portals:
  source=keyring` — byte-identical; the config file confirms complete-entry replacement (the earlier
  username is gone).
- Empty `list`: exit 0, empty stdout, stderr exactly `curator: no build_https scopes are
  configured`.
- Populated `list`: sorted scopes, `<scope><TAB><source> present=<bool>` shape, `present` flips when
  the token_env variable is set — field spellings match the doc.
- The page's JSON config example parses and lists as-is (all three scopes, including the host-only
  scope).
- All six documented rejection cases rejected by the shipped parser with field-scoped errors:
  invalid scope, unknown field, non-object entry, invalid env var name, present-but-empty value,
  zero/two sources; a literal token value in `token` is rejected by the source enumeration
  ("secrets never live in the config").

Code-verified (delivery checkout):

- Config location `~/.curator/config.json` / `CURATOR_CONFIG` (`internal/config/config.go:174-181`).
- Scope grammar wording matches `BuildSSHScopeRule`; matching is segment-aware longest-match
  (`longestScope` + `identity.MatchesPrefix`); the portals/portals-other example is correct.
- Entry grammar: exactly one of token/token_env; username default `token`
  (`BuildHTTPSDefaultUsername`).
- `login`: hidden prompt only when stdin AND stderr are terminals, else one line from stdin
  (`readBuildHTTPSToken`, main.go:2209); stores through `git credential approve` under the
  manager-namespaced username `curator-build-https:<scope>` with read-back proof
  (`gitcred.Access.StoreScoped`); records a keyring source; token never argv, never a config
  literal.
- `remove`: drops the selection; deletes the manager-stored token only for a keyring selection
  (`DeleteScoped`); never touches the operator's own host credential.
- Precheck: `planExternalBuilds` resolves SSH then HTTPS for the whole planned closure before the
  first repository fetch; selected-but-unavailable sources fail deterministically with per-source
  messages matching the doc's three bullets (`resolveConfiguredBuildHTTPS`).
- Env override: captured once (`CaptureBuildHTTPSSelection`); applies ahead of scopes for every
  host, or exactly one canonical host when `CURATOR_BUILD_HTTPS_HOST` is set (`Override.Host ==
  host`).
- Redirects: `http.followRedirects=false` in `strictFetchArgs` — backs the `.git`-suffix/301
  guidance.
- Platform mechanism: all credential IO is `git credential fill|approve|reject` against the
  operator's configured helper; `GIT_TERMINAL_PROMPT=0` during the protected fetch; private broker
  copy under `manager-wrappers/`; state file holds host+username only (0600); the secret rides
  `CURATOR_BUILD_HTTPS_ASKPASS_SECRET` on the fetch process only; broker answers only the two exact
  prompts for the bound host, and admission separately rejects a host-mismatched credential
  (`admission.go:331`).
- Spec citations: §6.1 (canonical source identity / segment matching, the convention
  docs/build-ssh.md already uses), §6.3 (external repository identity), core §12.2 verified to
  require exactly the documented disclosure ("MUST let the operator bind it to a host and MUST
  document the exposure"); "Spec core §12.2" matches the shipped CLI help's own spelling.
- Exposure warning present at every place this delta documents the override (docs page warning
  block, CHANGELOG entry); README does not mention the override; no other repo page documents it.
- No other manager implementations named in the new page.
- Links: README links to docs/build-ssh.md and docs/build-https.md, both present; the spec repo URL
  matches the existing convention. The repo has no automated link/markdown linter.
- Lint: `make lint` (golangci-lint) green in the story worktree after `git submodule update --init
  --recursive` — 0 issues. This independently closes the LOGBOOK-noted exit-2 gap on the isolated
  worktree.
- CHANGELOG entry present under Unreleased, consistent with the file's grammar.

## What I reran vs accepted

Reran myself: every transcript and negative parse check above, the shipped-binary build, and
`make lint` in this docs worktree. Accepted from the board without rerunning: the sibling stories'
own acceptance evidence (their reviews are recorded on TASK-260825-1tgpcn/2gyhq8/3kb532/3n4bjj,
168m7o/1lausy). Not exercised live: a real TTY install prompt and a keychain-backed
`present=true` keyring read (no operator-keychain writes from a spawned session); both are covered
by the accepted tests cited under F1 and by code reading.

## Routing

`to-dev` for documentation rework: fix F1 (three sentences + the LOGBOOK 0057 correction), F2
(real tab), and preferably resolve F3's README line. Everything in the "Verified" list is accurate
as written and should be preserved.
