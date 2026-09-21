# Review note for TASK-260920-3ccq6b revision 2 (orchestrator, binding)

Brief 3ccq6b-brief.md carries the rulings (R1 allow-list environment for the draft literal-URL
lane, R2 curator-owned ssh command with `-F <empty>`/`ProxyCommand=none`/... and default
known_hosts+agent, R3 legacy v1 and resolved lanes byte-identical, R4 askpass wording).
Revision 1 was blocked by a host exec hang mid-run (results.md §8) and then failed the gate on
a Windows skip-reason ledger miss; revision 2 (gate 35534261038 — verify the gate commit's tree
equals the exact revision-2 candidate tree) is the first complete revision. The race-lane
`attestation-evidence-wrong-key` failure on the rev1 gate is BUG-260920-2d9gfv (separate leaf),
not this change.

Judge specifically, with your own reruns (bounded local commands; the host has stall windows —
retry a hung command once rather than blocking; reviewer probes go in a disposable clone):
1. Allow-list correctness and completeness: every name the brief lists is honoured, nothing
   else reaches git (unit table + Windows case-insensitivity); confirm `HOME`/`USERPROFILE`
   honoured only for ssh defaults while git user config stays pinned empty; proxies
   (`http_proxy`/`https_proxy`/`all_proxy`/`no_proxy`, any case) dropped; `GIT_ASKPASS` and
   `SSH_AUTH_SOCK` pass through.
2. Production-entry rows through `cli.run` resolve/refresh: (a) `GIT_SSH_COMMAND`, (b)
   `GIT_PROXY_COMMAND`, (c) `GIT_EXEC_PATH`, (d) `~/.ssh/config` alias + ProxyCommand, (e)
   proxy env → zero connections, (g) agent/askpass positive. The producer found OpenSSH reads
   `~/.ssh/config` from the passwd entry, not `$HOME` — judge whether row (d) still proves the
   `-F` mechanism at the production entry (real `ssh -G` leg) or only at the unit layer, and
   whether that is enough for the AC ("ssh config isolation implemented").
3. The `-c http.sslVerify=true -c http.followRedirects=false -c credential.helper=` pins and the
   `installDraftTransportShim` test-shim change (clone detected anywhere in argv) — make sure the
   shim change does not weaken existing rows (v2 insteadOf, refresh rows).
4. Mutants observed (not predicted): allow-list filter removed; `-F` dropped; `ProxyCommand=none`
   dropped; ambient `GIT_SSH_COMMAND` passed through; `https_proxy` allowed; `GIT_ASKPASS`/
   `SSH_AUTH_SOCK` dropped. Rerun at least the first four yourself.
5. R3: `git diff --stat` against the Story base shows no file under `internal/buildrepo` or
   `internal/install`; `run` still `append(os.Environ(), …)`; legacy golden/untouched pins green.
6. Docs: docs/cli.md draft paragraph, troubleshooting section, `TestDraftDocsPinExamples`
   entries, CHANGELOG entry, N6 askpass wording; Windows skip reasons now in the ledger
   vocabulary (the rev1 gate failure) without widening the vocabulary.
Record exactly one verdict: accept_cr(TASK-260920-3ccq6b, revision=2, evidence=<your outcome
resource>) on ACCEPT, or changes_requested with file:line and reproduction.
