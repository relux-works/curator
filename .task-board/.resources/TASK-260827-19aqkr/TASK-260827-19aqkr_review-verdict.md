# TASK-260827-19aqkr review verdict: changes requested

Reviewer run, 2026-08-27. Branch `task-board/story/STORY-260827-3a5efk`, doc
commit `81e56a13` ("wip: docs-refresh work products (workspace record
recovery)"), `docs/build-https.md` 285 lines, md5 `c031cd47a3b6a0090e8db4f65802467f`.

Verdict: **changes requested**, route `to-dev`.

The document is not verifiable against the implementation it claims to
document, and the implementation it documents is not the merged one. Most of
the load-bearing claims (CLI flags, output strings, error identifiers, prompt
menu, off-terminal behavior, username default, one environment variable) are
contradicted by `origin/main`. One of them inverts a fail-closed claim into the
opposite of the real behavior.

## Blocking finding 0: the document was written against a tree with no HTTPS implementation

The story branch forked from `1f55f1b`, which is 55 commits behind
`origin/main`. The merged HTTPS broker does not exist anywhere in that tree, so
no claim in the document could have been grep-verified as the AC requires.

```
$ git merge-base --is-ancestor 903af23 origin/main; echo $?
1
$ git rev-list --count HEAD..origin/main
55
$ find . -name '*buildhttps*' -not -path './.git/*'
(no output)
$ git ls-tree -r --name-only origin/main -- internal/install/ | grep -i https
internal/install/buildhttps.go
internal/install/buildhttps_test.go
internal/install/buildhttpsprompt.go
internal/install/buildhttpsprompt_test.go
```

## Blocking finding 1: `docs/build-https.md` already exists on `origin/main`

The task asked to create the file. It is already there, 180 lines, written
against the real implementation, and it is correct on every point the branch
version gets wrong.

```
$ git ls-tree -r --name-only origin/main -- docs/
docs/build-https.md
docs/build-ssh.md
docs/implementation-plan.md
$ git show origin/main:docs/build-https.md | wc -l
     180
```

`docs/build-ssh.md` also moved upstream since the fork (9 insertions, 5
deletions), so the one-line cross-link added on the branch lands on a stale
copy.

```
$ git diff --stat 903af23:docs/build-ssh.md origin/main:docs/build-ssh.md
 docs/build-ssh.md | 14 +++++++++-----
```

## Blocking finding 2: off-terminal behavior is documented as the inverse of the code

The document states three times that an uncovered private HTTPS repository
fails closed before the fetch, and prints a fabricated
`build_repository_identity_invalid` diagnostic with candidate commands
(lines 5, 46, 219, and the whole block at lines 246-263).

The code does the opposite: a headless run continues anonymously.

```
$ sed -n '1338,1345p' cmd/curator/main.go
// operatorBuildHTTPSResolver offers unmatched HTTPS repositories an explicit
// operator choice. A headless run keeps this nil and therefore continues over
// anonymous HTTPS; a dry run remains read-only even on a terminal.
func operatorBuildHTTPSResolver(cfg *config.Config, dryRun bool) install.BuildHTTPSResolver {
	if dryRun || !attachedToTerminal(os.Stdin) || !attachedToTerminal(os.Stderr) {
		return nil
	}
```

```
$ sed -n '182,187p' internal/install/buildhttps.go
	if selection.Resolve == nil {
		for _, row := range missing {
			provenance = append(provenance, buildHTTPSProvenance(row, "anonymous"))
		}
		return credentials, provenance, nil
	}
```

The package comment states the rule explicitly (`internal/config/buildhttps.go:26-28`):

```
// Unlike build_ssh, an unmatched scope is not an error at resolution time:
// anonymous HTTPS is a real transport and public repositories must keep
// working, so absence of a selection simply means no credential is offered.
```

An operator following the document would expect a CI run to stop with a named
diagnostic and instead get an anonymous fetch that fails inside Git.

## Blocking finding 3: the error identifier does not exist for this surface

`build_repository_identity_invalid` is an admission code for non-absolute paths
and non-SSH transports. It has nothing to do with HTTPS credential selection.

```
$ grep -rn 'build_repository_identity_invalid' .
internal/buildrepo/admission.go:31:	CodeIdentityInvalid              = "build_repository_identity_invalid"
$ grep -rn 'CodeIdentityInvalid' internal/buildrepo/credentials.go internal/buildrepo/httpsbroker.go
internal/buildrepo/httpsbroker.go:135: ... "credential broker source is not absolute"
internal/buildrepo/credentials.go:59:  ... "%s path must be absolute"
internal/buildrepo/credentials.go:87:  ... "repository transport is not SSH"
```

The two identifiers that actually exist on this surface are absent from the
document:

```
$ grep -rhno 'build_repository_https[a-z_]*' . | sed 's/.*://' | sort -u
build_repository_https_credential_missing
build_repository_https_credential_selection_aborted
```

`build_repository_https_credential_missing` is the header of the interactive
prompt, not a failure (`internal/install/buildhttpsprompt.go:37`).

## Blocking finding 4: the `add` flags do not exist

Document, line 134: `add <scope> [--token git-credentials|keyring] [--token-env VAR] [--username USER]`.
Every console example in the document uses `--token`.

```
$ sed -n '2026,2029p' cmd/curator/main.go
  curator config build-https add <scope> (--git-credentials | --keyring | --token-env NAME) [--username NAME]
  curator config build-https login <scope> [--username NAME]
  curator config build-https list
  curator config build-https remove <scope>
$ grep -n 'requires exactly one of' cmd/curator/main.go
2153: "config build-https add requires exactly one of --git-credentials, --keyring, or --token-env"
```

There is no `--token` flag. `curator config build-https add ... --token git-credentials`
exits 2.

## Blocking finding 5: `CURATOR_BUILD_HTTPS_USERNAME` does not exist

The document lists it in the precedence table (line 43) and documents it with a
default of `oauth2` (lines 207-208).

```
$ grep -rhno 'CURATOR_BUILD_HTTPS[A-Z_]*' . | sed 's/.*://' | sort -u
CURATOR_BUILD_HTTPS_ASKPASS_SECRET
CURATOR_BUILD_HTTPS_ASKPASS_STATE
CURATOR_BUILD_HTTPS_HOST
CURATOR_BUILD_HTTPS_TOKEN
```

The override username is a constant, and it is `token`, not `oauth2`:

```
$ grep -n -A3 'BuildHTTPSDefaultUsername' internal/config/buildhttps.go
38:// BuildHTTPSDefaultUsername is sent alongside the resolved secret when an
39:// entry names no username of its own.
40:const BuildHTTPSDefaultUsername = "token"
$ sed -n '163p' internal/install/buildhttps.go
			credentials[key] = NewBuildHTTPSCredentials(config.BuildHTTPSDefaultUsername, selection.Override.secret)
```

Document line 91 ("Defaults to `oauth2` or `git` when omitted") is wrong for
the same reason. For a `git-credentials` entry with no explicit username the
value comes from the host credential material, not from a default
(`internal/install/buildhttps.go:243-245`).

## Blocking finding 6: the prompt transcript is fabricated

Document lines 234-242 show options `2) login with personal access token`,
`3) use environment variable GITHUB_TOKEN`, `m) enter token manually`, and the
answer line `credential [1-3, m, q] (default 1)`.

The real menu has one conditional entry, `t`, and `q`:

```
$ sed -n '101,116p' internal/install/buildhttpsprompt.go
func writeBuildHTTPSMenu(out io.Writer, material gitcred.HostMaterial) {
	if material.HostCredential {
		say(out, "  1) existing Git HTTPS credential for this host (username %s)  [default]\n", material.HostUsername)
	} else {
		say(out, "no existing Git HTTPS credential was detected for this host\n")
	}
	say(out, "  t) enter a token now\n")
	say(out, "  %s) abort\n", abortToken)
}
...
		prompt := fmt.Sprintf("credential [t, %s]", abortToken)
		if material.HostCredential {
			prompt = fmt.Sprintf("credential [1, t, %s] (default 1)", abortToken)
		}
```

The header text also differs: the code says `can use HTTPS credentials`, the
document says `needs HTTPS credentials`.

## Blocking finding 7: candidate discovery does not scan environment variables

Document line 226 claims discovery inspects `GITHUB_TOKEN`, `GITLAB_TOKEN`,
`CI_JOB_TOKEN`.

```
$ grep -rn 'GITHUB_TOKEN\|GITLAB_TOKEN\|CI_JOB_TOKEN' .
(no output)
```

`gitcred.Access.Discover` reads the host credential and manager-scoped entries
only (`internal/gitcred/gitcred.go:219-233`).

## Blocking finding 8: every quoted CLI output string is wrong

`formatBuildHTTPS` renders `source=`, not `token=`:

```
$ sed -n '2304,2316p' cmd/curator/main.go
func formatBuildHTTPS(credential config.BuildHTTPSCredential) string {
	var parts []string
	switch credential.Token {
	case config.TokenSourceGitCredentials:
		parts = append(parts, "source=git-credentials")
	case config.TokenSourceKeyring:
		parts = append(parts, "source=keyring")
	}
```

`list` appends a presence field the document omits entirely:

```
$ grep -n 'present=' cmd/curator/main.go
2261:		fmt.Printf("%s\t%s present=%t\n", scope, formatBuildHTTPS(credential), present)
```

`login` prints `token: ` on stderr and reports
`logged in and selected build_https scope <scope>: source=keyring` (or
`replaced the login for ...`). The document's `Enter personal access token for
<scope>:` and `token saved to system keyring for scope <scope>` appear nowhere
(`cmd/curator/main.go:2197-2199`, `2212`). The document also omits that
`login` writes the config entry, and that it falls back to reading one line
from stdin when either stream is not a terminal.

## Blocking finding 9: the parse-error table is wrong in three rows

```
$ sed -n '42,49p' internal/config/buildhttps.go
var BuildHTTPSTokenRule = "must be one of " + strings.Join(buildHTTPSTokenSourceList, ", ") + "; secrets never live in the config"
const BuildHTTPSTokenEnvRule = "must be an environment variable name"
const BuildHTTPSSourceRule = "requires exactly one of 'token' or 'token_env'"
```

| Document row | Actual message |
| --- | --- |
| `.token: must be 'git-credentials' or 'keyring'` | `.token: must be one of git-credentials, keyring; secrets never live in the config` |
| `: cannot combine 'token' and 'token_env'` | `: requires exactly one of 'token' or 'token_env'` |
| `: requires 'token' or 'token_env'` | same single message as above |

Both the zero-source and two-source faults produce one message
(`internal/config/buildhttps.go:102-104`). The document also omits
`must be a non-empty string when present` and the `token_env` name rule.

## Non-blocking observations

The scope grammar rule string, the segment-boundary and longest-match
semantics, the `curator-build-https:` keyring namespace, the
`build_https` config shape, and the whole transport-policy list
(`credential.helper=`, `core.askPass`, `GIT_ASKPASS`, `GIT_TERMINAL_PROMPT=0`,
`http.followRedirects=false`, `http.sslVerify=true`, `http.proxy=`,
`https.proxy=`) all check out against `internal/config/buildssh.go:40`,
`internal/gitcred/gitcred.go:48`, and `internal/buildrepo/admission.go:394-429`.
The cross-links are present in both directions. The prose style holds. That is
the part worth keeping.

Two smaller corrections for the rework: the broker itself lives in
`internal/buildrepo/httpsbroker.go` (the document cites `admission.go`, which
wires it), and the scope prompt offers a this-run-only answer `r` that persists
nothing (`internal/install/buildsshprompt.go:227-239`), which contradicts the
document's "The prompt records the operator choice into `config.json` before
proceeding."

## Recommended rework

Rebase the story branch onto current `origin/main` first. Then treat
`origin/main:docs/build-https.md` as the baseline and decide whether the task
is still "create" or has become "extend": the upstream document is correct but
shorter, and the branch version has a better contents structure and a
"Limits worth knowing" section worth porting once its claims are fixed. Every
remaining claim needs a grep against the post-rebase tree, not against the
sibling CocoaSkills document, which is where several of the fabrications above
came from (the CocoaSkills menu, the `--token` flag shape, and the `oauth2`
default are all sibling-doc artifacts).

## Verification note

The story worktree at `.temp/STORY-260827-3a5efk/worktree` was removed by a
concurrent session at 06:27 during this review. The work is intact on the
branch at `81e56a13` and mirrored at `.temp/docs-backup-260827/`; the md5 of
the reviewed content matches the committed blob. All implementation greps above
ran against a `git archive origin/main` extraction of `internal/` and `cmd/`.
