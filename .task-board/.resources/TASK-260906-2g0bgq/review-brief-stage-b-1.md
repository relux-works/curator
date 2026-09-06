# Review brief: implementation stage (b) — cycle 1

Subject: curator worktree `/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-b`, branch
`feat/agent-environments-stage-b`, head `73174cc3` — 12 signed commits on the stage (a) head `834b40f6`
(which lands as PR #59; judge the code, the orchestrator rebases). 26 files, +6565/−40. Diff:
`git diff 834b40f6..73174cc3`. Read `producer-brief-stage-b.md` and
`TASK-260906-2g0bgq_drafting-report.md`.

Authority: curator-spec main `f39f4a9` — `protocol/environments.md` revision 1.1 §5.3, §5.5, §5.7,
§5.8, §7.1–§7.9, §8.1–§8.4, §9.5, §10.1–§10.4, §11, §12; `schemas/v1/launch-env-fragment-v1` and
`agent-environment-marker-v1`; the `referenced-*`, `system-prompt-composed` and `mcp-*` expected sets;
`profiles/manager.md` §12; `cli/curator.md`; Decision 0012 D6/D8; Decision 0013 D6.

Attack, do not read. The stage-(a) reviews found sixteen defects by driving production entry points
and by asking "what does no test drive?" — apply the same method here:
1. **Managed homes and marker** — provision homes for every adapter and form; verify each surface is
   recorded with its content hash, its form, and a copy reason where §8.1 mandates a copy (the
   claude_code root context is always copied); mutate a surface on disk and confirm drift is reported
   per §8.4; confirm no surface escapes the environments root (§10.3).
2. **Seeds and passthrough** — the seed class is non-credential, one-time and never hashed; the
   verification-sprint shapes are used (`.claude.json` with the trust entry and, for the referenced
   form, `hasClaudeMdExternalIncludesApproved`); passthrough strategies match §7.4 per adapter with the
   liveness row; `isolated` is refused where §7.4 says `environment_isolated_unsupported`; the XDG
   allowlist and `environment_seed_shadowed`; `environment_seed_unreadable` stops provisioning and an
   unreadable seed is never treated as absent.
3. **Read-only resolve** — verification is lock-free and covers exactly the marker's surfaces; a stale
   home reports `environment_home_stale` with reasons and emits **no** fragment; `--repair` takes the
   mutation lock with a bounded wait and a distinct lock diagnostic; drive a concurrent holder and
   confirm the diagnostic, not a hang.
4. **Fragment** — the closed `launch-env-fragment-v1` validates against the published schema for every
   adapter and form: `profile.lock_sha256`, the `precedence` object, `env`, `system_prompt` with
   §7.3 descriptors, `mcp` (path, sorted `env_names` union, `argument`/`with`/`name`), `path_prepend`;
   `--format env|shell` quoting; a profile whose bytes try to select a variable name or escape the root
   is refused (§10.3).
5. **MCP channel files** — bytes match the `mcp-*` expected sets exactly (CCJ-1 where JSON, the codex
   fixed layer path and `args` spelling); managed homes only, never a native home; the package
   allowlist refuses with `mcp_package_not_allowed`; `env_names` grammar and the manager-reserved
   exclusion.
6. **Conformance** — the `referenced-*`, `system-prompt-composed` and `mcp-*` sets pass byte for byte
   through `CURATOR_CONFORMANCE_ROOT=/Users/iv/Developer/ReluxWorks/curator-spec/conformance/v1`, and
   the stage-deferred skips for exactly those cases are **gone**; every remaining skip is in a
   registered class with a truthful reason (the stage-(a) F4 lesson: no phantom environment variables).
7. **Gates** — `go build ./...`, `go vet ./...`, `gofmt -l`, `golangci-lint run ./...`,
   `go test -count=1 -race` on the touched packages, `bash .github/ci/gate-selftest.sh`, the
   platform-case gate for the three GOOS values, `go test ./cmd/curator`. Note that stage (a) shipped a
   Windows-only fixture defect that no Unix lane could see (backslashes in git config values) — check
   every new fixture for the same class before the hosted lane does.
8. Twelve signed commits by the repository's human identity; scope limited to stage (b).

Read-only (scratch under the worktree's `.temp/`). Never write into the control root. Findings resource
`TASK-260906-2g0bgq_review-findings-stage-b-1.md` (severity, file:line, quote, what is wrong, fix, your
evidence). Blocking/major → `development`; else explicit ACCEPT at `to-review` with `accept_cr`. Do
not mark done. `task-board handoff TASK-260906-2g0bgq --role reviewer`.
