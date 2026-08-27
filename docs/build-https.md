# Operator HTTPS credentials for external build repositories

A schema-7 skill may declare an external build repository: a Git repository Curator fetches and builds a command from. When that repository is reached over HTTPS, something has to decide which token authenticates the fetch. That decision is operator-owned. Git executables, tokens, credential sources, proxy policy, timeouts, and authentication modes `MUST NOT` be selected by a manifest, descriptor, repository, substitution, compiler environment, receipt trust field, or marker (`Spec §12.2`). A package can never choose the credential that reaches its own repository.

Curator implements that rule as an isolated credential surface. Ambient credential helpers and terminal prompts do not leak into the fetch session: private HTTPS fetches pass through a manager-controlled credential broker pinned to a single host. An unselected private HTTPS repository stops the run with `build_repository_identity_invalid` before the first fetch.

For external build repositories fetched over SSH, see [docs/build-ssh.md](build-ssh.md).

Contents:

- [What needs a selection](#what-needs-a-selection)
- [Where a selection comes from](#where-a-selection-comes-from)
- [The three token sources](#the-three-token-sources)
- [The `build_https` configuration field](#the-build_https-configuration-field)
- [Scope matching](#scope-matching)
- [`curator config build-https`](#curator-config-build-https)
- [Run-wide environment variables](#run-wide-environment-variables)
- [The install-time precheck](#the-install-time-precheck)
- [Credential broker and transport policy](#credential-broker-and-transport-policy)
- [Limits worth knowing](#limits-worth-knowing)

## What needs a selection

Only a repository whose *effective* state is a private network Git source on the HTTPS transport. The effective state, not the declared one, is what decides it (`Spec §6.4`): a development substitution that redirects a declared HTTPS repository onto SSH, or onto a local path, moves the repository off the HTTPS transport with it and it then needs no HTTPS credentials at all.

Anonymous HTTPS remains a first-class transport. A public repository that accepts unauthenticated fetches needs no credential selection and fetches without prompting.

So of these four planned repositories, two need a selection, one fetches anonymously, and the SSH one uses SSH credentials:

```
gitlab.example.test/portals/app   https   needs HTTPS credentials
gitlab.example.test/vendor/lib    https   needs HTTPS credentials
github.com/example/public-tool    https   fetches anonymously
git.example.test/open/lib         ssh     needs SSH credentials
```

## Where a selection comes from

Four sources, highest precedence first. The first one that names authentication material wins for a given repository.

| Precedence | Source | Covers |
| --- | --- | --- |
| 1 | `CURATOR_BUILD_HTTPS_TOKEN`, `CURATOR_BUILD_HTTPS_USERNAME`, `CURATOR_BUILD_HTTPS_HOST` | every HTTPS repository of the run, or the single pinned host |
| 2 | `build_https` scopes in the machine configuration | the repositories whose canonical identity a scope matches |
| 3 | the interactive precheck, on a terminal only | one repository at a time, by writing a scope through source 2 |
| 4 | unauthenticated HTTPS fallback | public repositories fetch anonymously; private fetches fail closed |

Environment overrides are read once at process entry, before any project-owned state can influence them.

## The three token sources

A `build_https` scope entry specifies one of three token sources (`Spec profiles/manager.md §11.4`). Every selection surface below carries one of them.

| Source | Configuration entry | CLI option | Token origin |
| --- | --- | --- | --- |
| `git-credentials` | `{"token": "git-credentials"}` | `--token git-credentials` | Reads the operator's existing HTTPS credential via `git credential fill` |
| `keyring` | `{"token": "keyring"}` | `curator config build-https login <scope>` | Stores a personal access token in the system credential helper under username `curator-build-https:<scope>` |
| `token_env` | `{"token_env": "VAR_NAME"}` | `--token-env VAR_NAME` | Reads the token from the named environment variable at run time |

The `git-credentials` source reuses credentials the operator already configured for Git on their host. The `keyring` source isolates Curator tokens from general Git credentials by storing them under a namespaced username. The `token_env` source suits non-interactive automation and continuous integration pipelines.

## The `build_https` configuration field

`build_https` is an optional object in the machine configuration (`~/.curator/config.json`, or `CURATOR_CONFIG`). Each key is a scope; each value is one credential entry:

```json
{
  "build_https": {
    "gitlab.example.com/portals/infra": {
      "token": "git-credentials"
    },
    "gitlab.example.com/vendor": {
      "token": "keyring",
      "username": "oauth2"
    },
    "ci.example.com": {
      "token_env": "CI_REPOSITORY_TOKEN"
    }
  }
}
```

The example configuration maps three distinct scopes to their respective token sources.

### Entry grammar

| Field | Type | Meaning |
| --- | --- | --- |
| `token` | `"git-credentials"` or `"keyring"` | The token source mechanism used for authentication. |
| `token_env` | environment variable name | The environment variable containing the token. Mutually exclusive with `token`. |
| `username` | string | Optional username for basic authentication. Defaults to `oauth2` or `git` when omitted. |

An entry must set `token` or `token_env`, but not both. Token values are never stored directly in `config.json`; the configuration records only the token source reference.

### Fail-closed parsing

The field is parsed strictly. A credential configuration that silently fails to apply is worse than a configuration that refuses to load, so anything the grammar does not spell out is an error and the whole configuration fails to load:

| Input | Result |
| --- | --- |
| `"build_https": []` | `build_https: must be an object` |
| a scope key outside the grammar | `build_https: scope "..." must be a lowercase host optionally followed by '/'-separated path segments of letters, digits, dots, underscores, or hyphens` |
| an entry that is not an object | `build_https.<scope>: must be an object` |
| an unknown entry field | `build_https.<scope>: has unsupported field(s): ...` |
| an invalid `token` value | `build_https.<scope>.token: must be 'git-credentials' or 'keyring'` |
| an entry naming both `token` and `token_env` | `build_https.<scope>: cannot combine 'token' and 'token_env'` |
| an entry naming neither `token` nor `token_env` | `build_https.<scope>: requires 'token' or 'token_env'` |

When several entries are faulty, the scopes are checked in sorted order, so one configuration file always reports the same fault.

## Scope matching

A scope is a lowercase host, optionally followed by `/`-separated path segments of letters, digits, dots, underscores, or hyphens. It is matched against the canonical `host/path` identity of the repository (`Spec §6.3`), which is the identity the repository is already locked to. Nothing else is ever a matching key: the identity is the only thing a repository is matched by, which is what keeps a package from selecting credentials at all (`Spec §12.2`).

Matching is segment-aware, exactly like allowlist matching in `Spec §6.1`:

```
scope gitlab.example.com/portals
  matches gitlab.example.com/portals
  matches gitlab.example.com/portals/app
  matches gitlab.example.com/portals/app/sub
  does not match gitlab.example.com/portals-other
```

The rule matches sub-paths under the scope and excludes host-name variations.

The longest matching scope wins. With both `gitlab.example.com/portals` and `gitlab.example.com/portals/legacy` configured, the repository `gitlab.example.com/portals/legacy/tool` selects the `legacy` entry and `gitlab.example.com/portals/app` selects the other one.

A value that is not a canonical `host/path` identity matches nothing. Unmatched private HTTPS repositories attempt anonymous authentication or trigger the install-time precheck.

## `curator config build-https`

```
curator config build-https add <scope> [--token git-credentials|keyring] [--token-env VAR] [--username USER]
curator config build-https login <scope> [--username USER]
curator config build-https list
curator config build-https remove <scope>
```

`add` validates the invocation before the configuration is read, so a malformed command is a usage error (exit 2) and not a failure attributed to the configuration file. It replaces whatever was recorded under the same scope and says which of the two it did:

```console
$ curator config build-https add gitlab.example.com/portals/infra --token git-credentials
added build_https scope gitlab.example.com/portals/infra: token=git-credentials

$ curator config build-https add gitlab.example.com/vendor --token keyring --username oauth2
added build_https scope gitlab.example.com/vendor: token=keyring username=oauth2

$ curator config build-https add ci.example.com --token-env CI_REPOSITORY_TOKEN
added build_https scope ci.example.com: token_env=CI_REPOSITORY_TOKEN

$ curator config build-https add gitlab.example.com/vendor --token keyring --username git
replaced build_https scope gitlab.example.com/vendor: token=keyring username=git
```

The command configures HTTPS credential scopes in `config.json`.

`login` prompts interactively for a personal access token without echoing characters, then writes the secret directly to the system keyring under username `curator-build-https:<scope>`:

```console
$ curator config build-https login gitlab.example.com/vendor
Enter personal access token for gitlab.example.com/vendor: 
token saved to system keyring for scope gitlab.example.com/vendor
```

The command securely stores an HTTPS token for the specified scope.

`list` prints one tab-separated line per scope, sorted, in the operator's own spelling:

```console
$ curator config build-https list
ci.example.com	token_env=CI_REPOSITORY_TOKEN
gitlab.example.com/portals/infra	token=git-credentials
gitlab.example.com/vendor	token=keyring username=git
```

The command lists all configured HTTPS credential scopes.

With nothing configured, `list` exits 0 and says so on stderr:

```console
$ curator config build-https list
curator: no build_https scopes are configured
```

The command indicates that no HTTPS scopes are currently stored.

`remove` deletes a scope entry and purges any corresponding secret from the system keyring:

```console
$ curator config build-https remove gitlab.example.com/vendor
removed build_https scope gitlab.example.com/vendor

$ curator config build-https remove gitlab.example.com/vendor
curator: build_https scope "gitlab.example.com/vendor" is not configured in /Users/example/.curator/config.json
```

The command removes the scope configuration and associated keyring credentials.

## Run-wide environment variables

`curator install`, `curator upgrade`, `curator global install`, and `curator global upgrade` observe process environment variables for HTTPS credential overrides:

```
  CURATOR_BUILD_HTTPS_TOKEN string
    	token for external HTTPS build repositories
  CURATOR_BUILD_HTTPS_USERNAME string
    	username for external HTTPS build repositories (default "oauth2")
  CURATOR_BUILD_HTTPS_HOST string
    	pin the environment token override to a single HTTPS host
```

Setting `CURATOR_BUILD_HTTPS_TOKEN` overrides all scope rules for the duration of the run. Tokens are passed exclusively through environment variables; they are never accepted as command-line flags to prevent secret disclosure in process listings (`ps`).

When `CURATOR_BUILD_HTTPS_HOST` is specified, the token override applies only to repositories on that host. Repositories on other hosts fall back to configured scopes or anonymous access.

## The install-time precheck

Credentials for the whole run are resolved before the first repository is fetched. A closure holding an uncovered private repository fails before touching the network, naming every uncovered repository at once.

### Candidate discovery

When an HTTPS repository requires credentials, Curator scans for existing authentication sources on the operator host:

- existing entries in the host Git credential helper for the target domain (`git credential fill`)
- environment variables matching common token patterns (`GITHUB_TOKEN`, `GITLAB_TOKEN`, `CI_JOB_TOKEN`)

Discovery lists existing material without creating a selection. Listing is not selecting.

### On a terminal

When stdin and stderr are interactive terminals and the run is not a dry run, Curator prompts the operator to select or enter credentials for uncovered repositories:

```console
build_repository_identity_invalid: gitlab.example.test/portals/app needs HTTPS credentials (command "build-tool" of skill "portals")
  1) use existing Git credentials for gitlab.example.test [default]
  2) login with personal access token (save to keyring)
  3) use environment variable GITHUB_TOKEN
  m) enter token manually
  q) abort
credential [1-3, m, q] (default 1): scope [gitlab.example.test/portals] (q to abort): recorded build_https scope gitlab.example.test/portals
```

The prompt records the operator choice into `config.json` before proceeding.

### Without a terminal, and on a dry run

There is no prompt when stdin or stderr is not a terminal, or during a dry run. The run fails closed and prints candidate commands:

```console
build_repository_identity_invalid: external build repositories need HTTPS credentials:
  gitlab.example.test/portals/app (command "build-tool" of skill "portals")
detected on this host:
  Git credential helper entry for gitlab.example.test
  environment variable GITHUB_TOKEN
select credentials with one of:
  curator config build-https add gitlab.example.test/portals --token git-credentials
  curator config build-https login gitlab.example.test/portals
  curator config build-https add gitlab.example.test/portals --token-env GITHUB_TOKEN
or set CURATOR_BUILD_HTTPS_TOKEN
```

The output gives actionable commands to configure access off-TTY.

## Credential broker and transport policy

HTTPS fetches execute through a dedicated manager credential broker (`internal/buildrepo/admission.go`). The manager invokes Git with strict parameter restrictions:

- `-c credential.helper=`: clears ambient Git credential helpers during fetch execution
- `-c core.askPass=<broker>`: sets the manager binary as the sole credential helper
- `GIT_ASKPASS=<broker>`: sets the credential broker in the process environment
- `GIT_TERMINAL_PROMPT=0`: disables interactive TTY dialogs from Git
- `-c http.followRedirects=false`: disables HTTP redirects to prevent credential leakage
- `-c http.sslVerify=true`: enforces strict TLS certificate verification
- `-c http.proxy=` and `-c https.proxy=`: clears ambient proxy configurations

The credential broker acts as a pure answer function: it answers only Git Username and Password prompts for the pinned host. Any prompt for an unpinned host or unrecognized prompt format causes the broker to exit without output, failing the fetch closed.

## Limits worth knowing

**Tokens are never recorded in configuration files, logs, or diagnostics.** Only token source declarations (`git-credentials`, `keyring`, `token_env`) land in `config.json`. Secrets reside exclusively in the process memory of fetch children or the system keyring.

**An unpinned environment override transmits the token to all HTTPS hosts in the closure.** `CURATOR_BUILD_HTTPS_TOKEN` without `CURATOR_BUILD_HTTPS_HOST` sends basic authentication headers containing the token to every HTTPS host fetched during the run. Use `CURATOR_BUILD_HTTPS_HOST` when a closure contains multiple distinct HTTPS hosts.

**Scope path segment rules differ from full HTTPS URL path rules.** A scope host and path segment admit only lowercase ASCII letters, digits, dots, underscores, and hyphens. HTTPS repository paths carrying complex characters (such as `+`) cannot be matched by sub-path scopes; cover such hosts using host-level scopes or environment variables.
