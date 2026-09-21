# TASK-260920-3ccq6b brief (orchestrator, binding)

Story STORY-260919-37szes, draft literal-URL lane only (`internal/gitops` isolated
functions introduced by BUG-260920-3ukdk4 at 99cb352b; call sites
`cmd/curator/project_resolve.go`, `internal/closure/resolve.go`). Contract:
curator-spec `protocol/repository-transport.md` §5 (line ~235): "User `~/.ssh/config`
Host aliases, `insteadOf`, `ProxyCommand`, `core.sshCommand`, helpers, include files
and environment overrides remain NOT imported" and §2 "No fallback disables
certificate or host-key validation". Reviewer residual N2 of BUG-3ukdk4 rev2 (verdict
§6) is the defect: `GIT_SSH`, `GIT_SSH_COMMAND`, `GIT_PROXY_COMMAND`, `GIT_EXEC_PATH`
(and every other git environment override) still reach git, and the real ssh still
reads the user's `~/.ssh/config`.

## Rulings (do not re-open; record them in results.md)

R1 Environment model = allow-list, not scrub-list. Build the draft-lane git
environment from an explicit allow-list (precedent: resolved lane
`cleanDiscoveryEnvironment` in `internal/buildrepo/admission.go`). Honoured on the
draft lane, and nothing else from the ambient environment: `PATH`; `HOME` /
`USERPROFILE` (only so OpenSSH finds its default `known_hosts` and default identity
files — document); temp dirs (`TMPDIR`, `TMP`, `TEMP`); `TZ`; Windows process
essentials (`SYSTEMROOT`, `WINDIR`, `COMSPEC`, `PATHEXT`, `SYSTEMDRIVE`,
`LOCALAPPDATA`/`APPDATA` only if git-for-windows demonstrably needs them — say
which and why); `SSH_AUTH_SOCK` (agent credentials); `GIT_ASKPASS` (kept by ruling:
the only HTTPS credential channel on this lane; it supplies credentials, it cannot
redirect the endpoint — keep the docs/cli.md disclosure). Locale: pin `LANG=C`,
`LC_ALL=C` as the resolved lane does (diagnostics are class vocabulary anyway). Every
other `GIT_*` name is dropped by construction (no list to forget: `GIT_SSH*`,
`GIT_PROXY_COMMAND`, `GIT_EXEC_PATH`, `GIT_DIR`/`GIT_WORK_TREE`/`GIT_*_DIRECTORY`,
`GIT_SSL_*`, `GIT_CURL_*`, `GIT_TRACE*`, `GIT_HTTP_*`, `GIT_CONFIG*`, `GIT_TEMPLATE_DIR`,
`GIT_NAMESPACE`, `GIT_ALTERNATE_OBJECT_DIRECTORIES`…), and so are `http_proxy`/
`https_proxy`/`all_proxy`/`no_proxy` (environment overrides; a proxy is admitted on
the resolved lane only through explicit operator configuration — document the bound
"draft lane does not honour proxy environment"). Keep the existing pins
(`GIT_CONFIG_GLOBAL`/`GIT_CONFIG_SYSTEM` → fresh empty file, `GIT_CONFIG_NOSYSTEM=1`,
`GIT_ALLOW_PROTOCOL`, `GIT_TERMINAL_PROMPT=0`); add `GIT_PROTOCOL_FROM_USER=0` and
`-c http.sslVerify=true -c http.followRedirects=false -c credential.helper=` (or the
environment equivalents) unless already implied — state which.

R2 ssh isolation on the draft lane: curator sets the ssh command itself (either
`GIT_SSH_COMMAND` with shell-safe quoting of the temp path, or `GIT_SSH` → a wrapper;
your choice, say why; Windows must work with git-for-windows' bundled sh/ssh) =
`ssh -F <fresh empty config file> -o BatchMode=yes -o StrictHostKeyChecking=yes
-o ProxyCommand=none -o ProxyJump=none -o PermitLocalCommand=no -o ForwardAgent=no
-o ClearAllForwardings=yes -o RequestTTY=no -o CanonicalizeHostname=no
-o UpdateHostKeys=no -o ConnectionAttempts=1` (mirror the resolved lane's
`ExactSSHCommand` list where it applies; do NOT pin identities/known_hosts paths —
the compiled-in defaults `~/.ssh/known_hosts` and default identity files plus the
agent stay in force; never `StrictHostKeyChecking=no`/`accept-new`). The ssh binary
is resolved through the honoured `PATH` (document as the draft lane's tooling bound;
the resolved lane pins tooling). Operator-facing consequence to document
(docs/cli.md draft section + troubleshooting): host aliases, `ProxyCommand`,
`IdentityFile` and any other `~/.ssh/config` entry are not read on the draft lane;
use an agent (or a default-named key) and the literal host name; unknown hosts fail
closed (`ssh-keyscan`/first manual `ssh` to seed `known_hosts`).

R3 Legacy v1 lane and resolved lane byte-identical (no file under `internal/buildrepo`,
`internal/install`; `run` keeps `append(os.Environ(), …)`; golden/legacy pins
`TestDraftTransportLegacyGolden`, `TestProjectResolveLegacyUntouched` stay green).

R4 Fold reviewer note N6: troubleshooting `GIT_ASKPASS` wording → "answers git's
username and password prompts (or embed the username in the endpoint URL)".

## Required evidence (37szes-review-brief applies)

- Production-entry rows (through `cli.run` resolve/refresh, real git, POSIX +
  declared Windows skip only where a sibling row already skips): (a) `GIT_SSH_COMMAND`
  pointing at a script that would serve an "evil" repo → the clone still binds the
  declared literal endpoint (mutant: honour the variable → KILLED); (b)
  `GIT_PROXY_COMMAND` likewise for a `git://`-style or ssh endpoint as feasible;
  (c) `GIT_EXEC_PATH` pointing at a directory with a poisoned `git-remote-https`/
  `git-upload-pack` shim → not honoured; (d) `~/.ssh/config` with `Host <literal>` →
  `HostName evil` + `ProxyCommand` under a temp `HOME` → the draft lane ignores it
  (assert on the fail-closed class/refusal or on the evil repo being untouched;
  sanitized diagnostics, no URL/tool leak); (e) `https_proxy` set to a listener that
  records connections → zero connections from the draft lane; (f) allow-list unit
  table: every ambient name outside the list is absent, every listed name present,
  case-insensitive on Windows spelling; (g) `SSH_AUTH_SOCK` and `GIT_ASKPASS` still
  pass through (positive rows).
- Mutants: remove the allow-list filter; drop `-F`; drop `ProxyCommand=none`; pass
  the ambient `GIT_SSH_COMMAND` through — each killed by a named row.
- Corpus: if a `v2-*` corpus row is the right home for (d), register it via
  `registerDraftSemantic` and keep `executed == total`; otherwise say why local rows.
- CHANGELOG `## Unreleased` → `### Security`/`### Fixed` entry; docs/cli.md draft
  paragraph (503–509) updated to the allow-list statement; troubleshooting section
  under the class the operator will see; `TestDraftDocsPinExamples` entries for the
  new wording.
- results.md: rulings R1–R4 restated, bounds (PATH-resolved tooling, proxies not
  honoured, default known_hosts/identities), Windows proof status, mutant table,
  ratio line. Publish the Change Request only when the configured gate is green.
