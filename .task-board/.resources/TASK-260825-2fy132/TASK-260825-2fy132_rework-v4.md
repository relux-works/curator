# Rework revision 4 — TASK-260825-2fy132

## Scope of this revision

Rev3 was asked for a one-word fix (G1: drop "private" from the candidate-prompt
sentence) but instead rewrote the Resolution section, the precedence list, the
host-narrowing sentence, and the LOGBOOK 0057 CORRECTION paragraph to deny that
an install-time HTTPS candidate prompt exists — a regression of rev1 finding
F1, flagged blocking (H1) in `TASK-260825-2fy132_review-verdict-rev3.md`.

This revision restores the rev2 wording for those four passages (rev2 was
verified correct by two independent reviews), applies the G1 word-drop, and
changes nothing else. README.md, CHANGELOG.md, and the warning block are kept
exactly as rev3 (already correct per the rev3 verdict).

## Tree used for verification

Per the landing map in `TASK-260825-1d0eo5` notes (orchestrator-verified,
supersedes earlier notes), the primary checkout `/Users/iv/Developer/ReluxWorks/curator`
is a stale partial copy that lacks `internal/install/buildhttpsprompt.go` and
the resolver wiring, and has misled three prior runs of this task. This
revision verifies every behavioural claim against the authoritative complete
code set: `/Users/iv/Developer/ReluxWorks/curator/.temp/STORY-260825-32bopo/worktree`.

## Changes made

- `docs/build-https.md`, "Resolution, precheck, and candidates": restored the
  rev2 paragraph describing the terminal candidate prompt (presence-only
  discovery offer, explicit selection, persist-or-this-run-only scope
  question, abort stops the run) and the headless/non-terminal/dry-run
  anonymous-fallback paragraph, with "private" dropped from the prompt
  sentence per G1.
- `docs/build-https.md`, "Precedence and exposure warning": restored the
  four-item precedence order (env override, scope, terminal prompt,
  anonymous) and the host-narrowing sentence mentioning the terminal prompt.
- `LOGBOOK.md` entry 0057: replaced the inverted CORRECTION paragraph (which
  denied the prompt) with rev2's CORRECTION paragraph verbatim, which
  correctly describes the shipped prompt. The SECURITY bullet keeps rev3's
  "private" removal.

## Verification against `.temp/STORY-260825-32bopo/worktree`

- `go build ./...`: clean.
- Config grammar (`internal/config/buildhttps.go`): fields limited to
  `token`/`token_env`/`username`, exactly-one-of token/token_env enforced,
  env var name regex, empty-value rejection, atomic (all-or-nothing) load,
  scope grammar/longest-match shared with `build_ssh` (`ValidBuildSSHScope`,
  `longestScope`) — matches the doc's grammar and matching sections.
- CLI (`cmd/curator/main.go` config-build-https handlers,
  `formatBuildHTTPS`): `added`/`replaced` verb selection, `source=keyring`
  vs `token_env=... username=...` formatting, tab-separated `list` output,
  exact stderr `curator: no build_https scopes are configured`, `remove`
  deleting the keyring-stored token only for `keyring` selections — all
  confirmed in source and reproduced byte-for-byte by building the
  authoritative worktree's binary and running the exact transcripts in
  the doc (see command log below).
- Resolution precedence (`internal/install/buildhttps.go`): explicit
  host-matched override, then longest scope, then (terminal only) the
  interactive resolver, then anonymous — matches the restored precedence
  list.
- **Interactive prompt** (`internal/install/buildhttpsprompt.go`,
  `InteractiveBuildHTTPSResolver`): a real, production-wired, test-covered
  menu (`buildhttpsprompt_test.go`, `main_test.go`) that calls
  presence-only `reader.Discover` for each unmatched host and lets the
  operator pick a candidate or enter a token, then persist a scope or use
  it for the current run only; abort stops the run. Wired at
  `cmd/curator/main.go` via `operatorBuildHTTPSResolver`, which returns
  `nil` (anonymous fallback, no prompt) only when the run is a dry run or
  stdin/stderr are not real terminals. This confirms the restored rev2
  text and confirms rev3's claim ("no candidate discovery or prompt … same
  in terminal, headless, and dry-run runs") was false for terminal runs.
- Askpass broker (`internal/buildrepo/httpsbroker.go`): private
  per-run-materialized broker, answers only the exact host/username Git
  credential prompts, state file holds only host+username, secret passed
  to the fetch process tree only, host-bound — matches the "Platform and
  fetch mechanism" section (unchanged from rev3).
- Spec citations (`§6.3`, `§6.1`, `Spec core §12.2`) confirmed present in
  `internal/config/buildhttps.go` and `cmd/curator/main.go` at the points
  the doc cites them.

## Command transcripts (built from the authoritative worktree)

```
$ curator config build-https list
curator: no build_https scopes are configured
$ curator config build-https add git.example.com/portals --token-env PORTALS_TOKEN --username oauth2
added build_https scope git.example.com/portals: token_env=PORTALS_TOKEN username=oauth2
$ curator config build-https add git.example.com/portals --keyring
replaced build_https scope git.example.com/portals: source=keyring
$ curator config build-https add git.example.com/tools --token-env TOOLS_TOKEN --username oauth2
added build_https scope git.example.com/tools: token_env=TOOLS_TOKEN username=oauth2
$ curator config build-https list
git.example.com/tools	token_env=TOOLS_TOKEN username=oauth2 present=false
$ curator config build-https remove git.example.com/portals
removed build_https scope git.example.com/portals
$ curator config build-https remove git.example.com/portals
curator: build_https scope "git.example.com/portals" is not configured in <config path>
```

All exit codes matched expectations (0 for success paths, 1 for the
already-removed re-removal), and separators are real tabs (confirmed with
`sed -n l`), matching the doc's transcripts.

## Local gates run in this worktree

- `make lint`: `0 issues.` (exit 0).
- `git diff --check`: exit 0 (no whitespace errors).
- Links: `docs/build-ssh.md` and `docs/build-https.md` both exist and are
  referenced correctly from `README.md`; no other document links reference
  `docs/build-https.md`.

This worktree carries no Go source changes (docs task), so `go build`/`go
test` here are not applicable to this delta; the shipped-behavior claims were
verified by building and exercising the authoritative code worktree instead,
per the landing map's tree-selection guidance.
