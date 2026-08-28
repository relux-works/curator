# HTTPS credentials for external build repositories

Curator can offer an operator-owned HTTPS credential to an external build
repository. This page documents the `build_https` configuration, the
`curator config build-https` command, and the fetch boundary. It follows the
[Curator Specification](https://github.com/relux-works/curator-spec): canonical
repository identity is defined by `Spec §6.3`, scope matching by `Spec §6.1`,
and credential ownership and disclosure by `Spec core §12.2`.

Credentials are selected only by the operator. A manifest, descriptor,
repository, substitution, or marker cannot choose one.

For SSH repository credentials, see [Operator SSH credentials for external build repositories](build-ssh.md).

## Configuration

`build_https` is an optional object in the machine configuration
(`~/.curator/config.json`, or `CURATOR_CONFIG`). Each key is a scope and each
value is one token-source selection. The configuration names where to read a
token; it never contains a literal token.

```json
{
  "build_https": {
    "git.example.com": {
      "token": "git-credentials"
    },
    "git.example.com/portals": {
      "token_env": "PORTALS_TOKEN",
      "username": "oauth2"
    },
    "code.example.net/tools": {
      "token": "keyring"
    }
  }
}
```

### Scope grammar and matching

A scope is a lowercase host, optionally followed by `/`-separated path
segments containing letters, digits, dots, underscores, or hyphens. It is
matched to the canonical `host/path` repository identity on whole-segment
boundaries. The longest matching scope wins.

```text
git.example.com/portals
  matches     git.example.com/portals
  matches     git.example.com/portals/app
  does not    git.example.com/portals-other
```

This makes a scope an operator choice over a known repository identity, not a
rule that package data can widen. See `Spec §6.1` and `Spec §6.3`.

### Entry grammar and token sources

Each entry is an object with only these fields:

| Field | Required | Meaning |
| --- | --- | --- |
| `token` | exactly one of `token` or `token_env` | `git-credentials` reads the operator's existing Git HTTPS credential for the scope host. `keyring` reads the manager-namespaced credential stored by `login`. |
| `token_env` | exactly one of `token` or `token_env` | An environment variable name. Its value is captured when the Curator process starts. |
| `username` | no | Username sent with the resolved token. It defaults to `token`. |

The field is parsed strictly. An invalid scope, an unknown field, a non-object
entry, an invalid environment variable name, an empty field value, or an entry
that names zero or two sources rejects the configuration. The configuration is
not partially applied.

## `curator config build-https`

The command surface is:

```text
curator config build-https add <scope> (--git-credentials | --keyring | --token-env NAME) [--username NAME]
curator config build-https login <scope> [--username NAME]
curator config build-https list
curator config build-https remove <scope>
```

`add` records exactly one source and replaces the complete entry under that
scope. It prints the source, never the token:

```console
$ curator config build-https add git.example.com/portals --token-env PORTALS_TOKEN --username oauth2
added build_https scope git.example.com/portals: token_env=PORTALS_TOKEN username=oauth2

$ curator config build-https add git.example.com/portals --keyring
replaced build_https scope git.example.com/portals: source=keyring
```

`login` reads one token from a hidden terminal prompt. When either standard
input or standard error is not a terminal, it reads one line from standard
input instead. It stores the token through the operator's Git credential
machinery under a scope-namespaced entry, then records a `keyring` source for
that scope. A token is never a command-line argument and is never written as a
literal configuration value.

`list` sorts scopes and reports whether their named source resolves now:

```console
$ curator config build-https list
git.example.com/portals	source=keyring present=true
git.example.com/tools	token_env=TOOLS_TOKEN username=oauth2 present=false
```

With no configured scopes, `list` exits zero, prints no standard output, and
reports `curator: no build_https scopes are configured` on standard error.
`remove` removes the selection. For a `keyring` selection it also removes the
manager-namespaced stored token; it never removes the operator's own Git
credential selected by `git-credentials`.

## Resolution, precheck, and candidates

Before the first external repository fetch, Curator resolves HTTPS selections
for every HTTPS build repository in the planned closure. This gives a selected
but unavailable source a deterministic failure before any repository fetch:

- `git-credentials` requires an existing operator Git credential for the
  scope host.
- `keyring` requires a token previously stored by `login` for that scope.
- `token_env` requires the named variable to have been non-empty at process
  entry.

No configured or environment source is not an error. HTTPS has an anonymous
transport, so an uncovered repository can be fetched anonymously. On an
operator terminal, however, an uncovered repository first opens a
candidate prompt before any fetch. The prompt offers an existing Git HTTPS
credential for that host when presence-only discovery finds one, or lets the
operator enter a token now. No candidate is read or used until the operator
selects it. The operator then chooses a scope to persist or a this-run-only
(`r`) choice; the latter never writes configuration or credential storage.
Aborting the prompt stops the run rather than silently falling back to
anonymous HTTPS.

Headless, non-terminal, and dry-run runs never prompt. Their uncovered HTTPS
repositories continue anonymously. `list` is a presence-only check for
configured candidates; it does not reveal tokens or select a credential for a
repository.

For an HTTPS address in a manifest, include the `.git` suffix. The service may
otherwise answer with a `301` redirect, and Curator's protected fetch rejects
redirects rather than forwarding a request or credential.

## Precedence and exposure warning

Resolution order is:

1. `CURATOR_BUILD_HTTPS_TOKEN`, optionally restricted by
   `CURATOR_BUILD_HTTPS_HOST` to one exact canonical host.
2. The longest matching `build_https` scope.
3. On an operator terminal, the credential candidate prompt for an uncovered
   repository.
4. Anonymous HTTPS when neither applies and no terminal prompt is active.

When `CURATOR_BUILD_HTTPS_HOST` is set, repositories on other hosts resolve as
though the run-wide override were absent. They may use a matching configured
scope, be offered candidates on an operator terminal, or remain anonymous.

> Warning: `CURATOR_BUILD_HTTPS_TOKEN` without
> `CURATOR_BUILD_HTTPS_HOST` is identity-unbound. It is offered to every
> HTTPS build repository host the run reaches. Set
> `CURATOR_BUILD_HTTPS_HOST` to bind the override to one host, or use a
> `build_https` scope, unless every reached host is intended to receive that
> token. This disclosure is required by `Spec core §12.2`.

## Platform and fetch mechanism

Curator uses the operator's `git credential` interface to read, store, and
delete credentials. That mechanism uses the credential helper configured by
the operator's Git installation, so the same configuration surface works on
macOS, Linux, and Windows. Curator does not ask a fetch process to discover
credentials, and it disables Git terminal prompting during the protected
fetch.

For an authenticated fetch, Curator materializes a private copy of its fixed
askpass broker. The broker receives only the exact username and password
prompts for the selected host. Its state file contains the host and username;
the secret exists only in that fetch process tree. Other Git subprocesses do
not receive the broker state or secret. The broker is bound to the resolved
host, so a credential selected for one host cannot answer a prompt for another.
