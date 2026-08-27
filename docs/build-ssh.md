# Operator SSH credentials for external build repositories

A schema-7 skill may declare an external build repository: a Git repository
Curator fetches and builds a command from. When that repository is reached over
SSH, something has to decide which key authenticates the fetch. That decision is
operator-owned. Git and SSH executables, credentials, host-verification state,
proxy policy, timeouts, and authentication mode `MUST NOT` be selected by a
manifest, descriptor, repository, substitution, compiler environment, receipt
trust field, or marker (`Spec §12.2`). A package can never choose the key that
reaches its own repository.

Curator implements that rule as a fail-closed selection surface. There is no
fallback onto ambient SSH state: a fetch that succeeds because some unrelated
key happened to be loaded in the operator agent is a fetch nobody authorized.
An SSH build repository the operator has not covered stops the run with
`build_repository_ssh_credential_missing` before the first fetch.

For external build repositories fetched over HTTPS, see [docs/build-https.md](build-https.md).

Contents:

- [What needs a selection](#what-needs-a-selection)
- [Where a selection comes from](#where-a-selection-comes-from)
- [The three selection shapes](#the-three-selection-shapes)
- [The `build_ssh` configuration field](#the-build_ssh-configuration-field)
- [Scope matching](#scope-matching)
- [`curator config build-ssh`](#curator-config-build-ssh)
- [Run-wide flags and environment](#run-wide-flags-and-environment)
- [The install-time precheck](#the-install-time-precheck)
- [Host keys](#host-keys)
- [Limits worth knowing](#limits-worth-knowing)

## What needs a selection

Only a repository whose *effective* state is a network Git source on the SSH
transport. The effective state, not the declared one, is what decides it
(`Spec §6.4`): a development substitution that redirects a declared SSH
repository onto HTTPS, or onto a local path, moves the repository off the SSH
transport with it and it then needs no credentials at all.

So of these four planned repositories, three need a selection and the HTTPS one
does not:

```
git.example.test/portals/app      ssh     needs a selection
git.example.test/portals/agent    ssh     needs a selection
other.example.test/tools/kit      ssh     needs a selection
git.example.test/open/lib         https   needs nothing
```

## Where a selection comes from

Four sources, highest precedence first. The first one that names authentication
material wins for a given repository.

| Precedence | Source | Covers |
| --- | --- | --- |
| 1 | `--build-ssh-identity`, `--build-ssh-agent`, `--build-ssh-known-hosts` | every SSH repository of the run |
| 2 | `CURATOR_BUILD_SSH_IDENTITY`, `CURATOR_BUILD_SSH_AGENT`, `CURATOR_BUILD_SSH_KNOWN_HOSTS` | every SSH repository of the run |
| 3 | `build_ssh` scopes in the machine configuration | the repositories whose canonical identity a scope matches |
| 4 | the interactive precheck, on a terminal only | one repository at a time, by writing a scope through source 3 |
| - | nothing matched | the run fails closed |

Flags and environment merge field by field, flags first: `--build-ssh-agent`
with `CURATOR_BUILD_SSH_IDENTITY` set in the environment is a valid pinned-agent
selection. The merged result is the *run-wide* selection. If it names an
identity or an agent, it covers every SSH repository of the run and no
configured scope is consulted. Naming only `--build-ssh-known-hosts` is not a
selection: known hosts say who the destination is allowed to be, not what
authenticates to it.

Both flags and environment are read once at process entry, before any
project-owned state can influence them.

## The three selection shapes

An SSH invocation admits exactly three authentication tails
(`Spec profiles/manager.md §11.3`). Every selection surface below carries one of
them and nothing else.

| Shape | Config entry | CLI | Authentication tail |
| --- | --- | --- | --- |
| identity only | `{"identity": PATH}` | `--identity PATH` | `-o IdentitiesOnly=yes -o IdentityAgent=none -i PATH` |
| agent only | `{"agent": true}` or `{"agent": SOCKET}` | `--agent` or `--agent SOCKET` | `-o IdentitiesOnly=no -o IdentityFile=none -o IdentityAgent=SOCKET` |
| pinned agent | `{"agent": true, "identity": PATH}` | `--agent --identity PATH` | `-o IdentitiesOnly=yes -o IdentityAgent=SOCKET -i PATH` |

The pinned-agent shape is the recommended one whenever an agent is in use. The
agent holds the private key and the named identity, conventionally the public
half, pins which single key the agent may offer. It spends one server
authentication attempt instead of one per loaded key, discloses one public key
to the destination rather than enumerating every key the agent holds, and
authenticates a passphrase-protected key without a prompt. It is also the
default entry of the interactive precheck menu.

The identity-only and agent-only tails are `Spec profiles/manager.md §11.3` as
published on `curator-spec` main. The pinned-agent tail is admitted by the
`spec/pinned-agent-authentication-tail` revision of that section, which is not
yet merged on `curator-spec` main; Curator implements all three. Read the
current section text before treating any of this as settled protocol.

In every shape the operator names each element explicitly. No ambient agent
socket, default identity, user or system SSH configuration, prompt, TTY,
forwarding, proxy, local command, or control connection is admitted
(`Spec profiles/manager.md §11.3`).

## The `build_ssh` configuration field

`build_ssh` is an optional object in the machine configuration
(`~/.curator/config.json`, or `CURATOR_CONFIG`). Each key is a scope; each value
is one credential entry.

```json
{
  "build_ssh": {
    "build.example.net": {
      "identity": "~/.ssh/build_ed25519"
    },
    "git.example.com/portals": {
      "agent": true,
      "identity": "~/.ssh/id_ed25519.pub"
    },
    "git.example.com/portals/legacy": {
      "agent": "/run/agent.sock",
      "known_hosts": "~/.ssh/known_hosts_build"
    },
    "git.example.com/tools": {
      "agent": true,
      "identity": "~/.ssh/tools_ed25519.pub"
    }
  }
}
```

### Entry grammar

| Field | Type | Meaning |
| --- | --- | --- |
| `agent` | `true`, or a socket path | `true` resolves to the socket the operator environment advertises in `SSH_AUTH_SOCK` at run time. A string names one socket. |
| `identity` | path | The identity file offered to the destination. With `agent`, it pins the single key the agent may offer. |
| `known_hosts` | path | Host keys this scope verifies against, overriding the operator default. |

An entry must name `agent`, `identity`, or both. Path values must be absolute or
start with `~/`, carry no control character, and be at most 4096 Unicode scalar
values. A leading `~/` is resolved against the operator home when the entry is
used. Windows absolute forms (`C:\...`, `C:/...`, `\\...`) are recognised
without consulting the running platform, so one configuration file yields the
same verdict wherever it is read.

`"agent": false` is rejected rather than read as an identity-only entry: only
the affirmative spelling exists, so there is exactly one way to write each
shape.

### Fail-closed parsing

The field is parsed strictly. A credential that silently fails to apply is worse
than a configuration that refuses to load, so anything the grammar does not
spell out is an error and the whole configuration fails to load:

| Input | Result |
| --- | --- |
| `"build_ssh": []` | `build_ssh: must be an object` |
| a scope key outside the grammar | `build_ssh: scope "..." must be a lowercase host optionally followed by '/'-separated path segments of letters, digits, dots, underscores, or hyphens` |
| an entry that is not an object | `build_ssh.<scope>: must be an object` |
| an unknown entry field | `build_ssh.<scope>: has unsupported field(s): ...` |
| `"agent": false`, or a non-string non-bool `agent` | `build_ssh.<scope>.agent: must be true for the operator's own agent socket, or an agent socket path` |
| a relative, empty, or control-carrying path | `build_ssh.<scope>.<field>: must be an absolute path or start with '~/' and carry no control character` |
| an entry naming neither `agent` nor `identity` | `build_ssh.<scope>: requires 'agent', 'identity', or both` |

When several entries are faulty, the scopes are checked in sorted order, so one
configuration file always reports the same fault.

## Scope matching

A scope is a lowercase host, optionally followed by `/`-separated path segments
of letters, digits, dots, underscores, or hyphens. It is matched against the
canonical `host/path` identity of the repository (`Spec §6.3`), which is the
identity the repository is already locked to. Nothing else is ever a matching
key: the identity is the only thing a repository is matched by, which is what
keeps a package from selecting credentials at all (`Spec §12.2`).

Matching is segment-aware, exactly like allowlist matching in `Spec §6.1`:

```
scope git.example.com/portals
  matches git.example.com/portals
  matches git.example.com/portals/app
  matches git.example.com/portals/app/sub
  does not match git.example.com/portals-other
```

The longest matching scope wins. With both `git.example.com/portals` and
`git.example.com/portals/legacy` configured, the repository
`git.example.com/portals/legacy/tool` selects the `legacy` entry and
`git.example.com/portals/app` selects the other one.

A value that is not a canonical `host/path` identity matches nothing, and an
identity no scope covers selects nothing. In both cases the run fails closed
rather than falling back on ambient SSH state.

## `curator config build-ssh`

```
curator config build-ssh add <scope> [--agent [SOCKET]] [--identity PATH] [--known-hosts PATH]
curator config build-ssh list
curator config build-ssh remove <scope>
```

`--agent` is optional-valued: bare, it records `"agent": true`; followed by a
socket path, it records that socket. It only claims the next token when that
token reads as a socket path, so `curator config build-ssh add --agent
git.example.com/portals` keeps the scope positional.

`add` validates the invocation before the configuration is read, so a malformed
command is a usage error (exit 2) and not a failure attributed to the
configuration file. It replaces whatever was recorded under the same scope and
says which of the two it did:

```console
$ curator config build-ssh add git.example.com/portals --agent --identity ~/.ssh/id_ed25519.pub
added build_ssh scope git.example.com/portals: agent identity=~/.ssh/id_ed25519.pub

$ curator config build-ssh add git.example.com/tools --agent
added build_ssh scope git.example.com/tools: agent

$ curator config build-ssh add build.example.net --identity ~/.ssh/build_ed25519
added build_ssh scope build.example.net: identity=~/.ssh/build_ed25519

$ curator config build-ssh add git.example.com/portals/legacy --agent /run/agent.sock --known-hosts ~/.ssh/known_hosts_build
added build_ssh scope git.example.com/portals/legacy: agent=/run/agent.sock known_hosts=~/.ssh/known_hosts_build

$ curator config build-ssh add git.example.com/tools --agent --identity ~/.ssh/tools_ed25519.pub
replaced build_ssh scope git.example.com/tools: agent identity=~/.ssh/tools_ed25519.pub
```

`list` prints one tab-separated line per scope, sorted, in the operator's own
spelling rather than the resolved one:

```console
$ curator config build-ssh list
build.example.net	identity=~/.ssh/build_ed25519
git.example.com/portals	agent identity=~/.ssh/id_ed25519.pub
git.example.com/portals/legacy	agent=/run/agent.sock known_hosts=~/.ssh/known_hosts_build
git.example.com/tools	agent identity=~/.ssh/tools_ed25519.pub
```

With nothing configured, `list` exits 0 and says so on stderr, so a caller
parsing the listing sees an empty stdout rather than a line naming no scope:

```console
$ curator config build-ssh list
curator: no build_ssh scopes are configured
```

`remove` reports a scope that is not configured as an error (exit 1) rather than
succeeding silently:

```console
$ curator config build-ssh remove git.example.com/portals/legacy
removed build_ssh scope git.example.com/portals/legacy

$ curator config build-ssh remove git.example.com/portals/legacy
curator: build_ssh scope "git.example.com/portals/legacy" is not configured in /Users/example/.curator/config.json
```

Malformed input is refused before anything is written:

```console
$ curator config build-ssh add 'git.example.com/portals-evil/../x' --agent
curator: build_ssh: scope "git.example.com/portals-evil/../x" must be a lowercase host optionally followed by '/'-separated path segments of letters, digits, dots, underscores, or hyphens

$ curator config build-ssh add git.example.com/portals --known-hosts ~/.ssh/known_hosts
curator: build_ssh.git.example.com/portals: requires 'agent', 'identity', or both
```

## Run-wide flags and environment

`curator install`, `curator upgrade`, `curator global install`, and
`curator global upgrade` accept:

```
  -build-ssh-agent string
    	agent socket for external SSH build repositories, or "auto" for your own agent (or CURATOR_BUILD_SSH_AGENT)
  -build-ssh-identity string
    	identity file for external SSH build repositories (or CURATOR_BUILD_SSH_IDENTITY)
  -build-ssh-known-hosts string
    	host keys external SSH build repositories are verified against (or CURATOR_BUILD_SSH_KNOWN_HOSTS)
```

`--build-ssh-agent auto` adopts the socket the operator environment advertises
in `SSH_AUTH_SOCK`, which is the one thing a persisted selection cannot keep
current: a macOS agent socket is per login session. If `SSH_AUTH_SOCK` is unset,
the run fails with `build_repository_ssh_credential_missing` rather than
continuing without an agent. Any other value is taken as a socket path.

Unlike configuration entries, flag and environment paths are **not**
tilde-expanded by Curator. A shell expands `~` in `--build-ssh-identity
~/.ssh/id_ed25519` before Curator sees it, but
`CURATOR_BUILD_SSH_IDENTITY=~/.ssh/id_ed25519` reaches Curator literally and is
refused:

```
build_repository_identity_invalid: SSH identity path must be absolute
```

Use an absolute path in the environment.

Every operator-named path is proved against the filesystem before it reaches the
SSH wrapper policy: a symbolic link is resolved rather than refused, since a
live agent socket is conventionally a stable link onto a per-session rendezvous
point, and the resolved target must then be a regular file for an identity or a
known-hosts file and a socket for an agent.

## The install-time precheck

Credentials for the whole run are resolved before the first repository is
fetched. A closure holding one uncovered private repository therefore fails
without having touched the network, and names every uncovered repository at
once rather than stopping part way through.

### Candidate discovery

Once a repository is actually uncovered, Curator lists the authentication
material the operator already owns, and nothing else:

- the live `SSH_AUTH_SOCK`, with a key count from `ssh-add -l`. Every failure
  path, whether the tool is absent, the socket is stale, the agent refuses, or
  the 5 second timeout expires, degrades to "key count unavailable" rather than
  dropping the agent as a candidate. Exit status 1 is a real answer, an agent
  holding nothing, not a failure to reach one.
- `*.pub` files directly under `~/.ssh`, regular files only, sorted, spelled
  with the leading `~/` the configuration grammar accepts, capped at 8 with the
  remainder reported as a count rather than dropped silently.

Discovery reads only operator-owned state, never consults package data, and
never turns what it finds into a selection. Listing is not selecting.

### On a terminal

When both stdin and stderr are real terminals and the run is not a dry run,
Curator asks two questions per uncovered repository: what to authenticate with,
and how widely that answer applies.

```console
build_repository_ssh_credential_missing: git.example.test/portals/app needs SSH credentials (command "build-tool" of skill "portals")
  1) agent, pinned to ~/.ssh/id_ed25519.pub  [default]
  2) SSH agent at /run/agent.sock holding 2 keys, any key it holds
  3) identity ~/.ssh/id_ed25519.pub
  4) identity ~/.ssh/work.pub
  m) enter an identity path
  q) abort
credential [1-4, m, q] (default 1): scope [git.example.test/portals] (q to abort): recorded build_ssh scope git.example.test/portals
```

- Entry 1 is the default and is the pinned-agent shape, the only one that both
  reuses an already loaded key and stops the agent offering every other key to
  the destination. It still has to be accepted; nothing is pre-applied.
- `m` takes a free-form identity path, for a key outside `~/.ssh` or one the
  listing capped away. A malformed path is refused and asked again.
- An answer that parses as none of the offered choices is asked again rather
  than resolved to the default: the default authorizes a key just as much as
  any other entry does. The numeric answer is read leniently, so `1)` and
  `1abc` both select entry 1.
- The scope question defaults to the repository namespace, so a sibling
  repository of the same group is covered without naming every repository by
  hand. In the transcript above there were two uncovered repositories under
  `git.example.test/portals`; the scope chosen for the first covered the second,
  which is why only one question pair appears.
- Nothing is persisted before both questions are answered. The operator
  authorizes a scope, not just a key.
- `q` at either question, and end of input, abort with
  `build_repository_ssh_credential_missing`. An abort is a refusal to
  authenticate, not a licence to fall back on ambient SSH state.

A recorded answer is written through the same writer `curator config build-ssh
add` uses, so it lands in the machine configuration in exactly the same shape.

### Without a terminal, and on a dry run

There is no prompt when stdin or stderr is not a real terminal, and none on a
dry run, which is a read-only surface that must not write a credential mid
report. The run fails closed with the uncovered repositories, the material
detected on this host, and ready-to-run commands built from exactly those
candidates, one set per uncovered namespace, deduplicated:

```console
build_repository_ssh_credential_missing: external build repositories need SSH credentials:
  git.example.test/portals/app (command "build-tool" of skill "portals")
  git.example.test/portals/agent (command "build-agent" of skill "portals")
  other.example.test/tools/kit (command "build-kit" of skill "infra")
detected on this host:
  SSH agent at /run/agent.sock holding 2 keys
  ~/.ssh/id_ed25519.pub
  ~/.ssh/work.pub
select credentials with one of:
  curator config build-ssh add git.example.test/portals --agent --identity ~/.ssh/id_ed25519.pub
  curator config build-ssh add git.example.test/portals --agent
  curator config build-ssh add git.example.test/portals --identity ~/.ssh/id_ed25519.pub
  curator config build-ssh add git.example.test/portals --identity ~/.ssh/work.pub
  curator config build-ssh add other.example.test/tools --agent --identity ~/.ssh/id_ed25519.pub
  curator config build-ssh add other.example.test/tools --agent
  curator config build-ssh add other.example.test/tools --identity ~/.ssh/id_ed25519.pub
  curator config build-ssh add other.example.test/tools --identity ~/.ssh/work.pub
or pass --build-ssh-agent/--build-ssh-identity, or set CURATOR_BUILD_SSH_AGENT/CURATOR_BUILD_SSH_IDENTITY
```

The HTTPS repository of the same closure is absent from that list, because it
needs no selection.

A host where discovery found nothing says so, and emits a placeholder the
operator has to replace rather than inventing a path:

```console
build_repository_ssh_credential_missing: external build repositories need SSH credentials:
  git.example.test/portals/app (command "build-tool" of skill "portals")
no SSH agent and no ~/.ssh/*.pub identity were detected on this host
select credentials with one of:
  curator config build-ssh add git.example.test/portals --agent --identity ~/.ssh/<key>.pub
  curator config build-ssh add git.example.test/portals --agent
or pass --build-ssh-agent/--build-ssh-identity, or set CURATOR_BUILD_SSH_AGENT/CURATOR_BUILD_SSH_IDENTITY
```

### Dry-run provenance

A dry run reports where each repository's credentials would come from, one line
per repository that needs them, prefixed by the project alias, or by `global`
for the global scope:

```
external build ssh: git.example.test/portals/app (command "build-tool" of skill "portals") <- config scope "git.example.test/portals"
```

The source reads `operator flags/env` for a run-wide selection and
`config scope "<scope>"` for a configured one. These lines are populated only on
a dry run; an ordinary install reports through its own build rows.

## Host keys

Every SSH fetch pins `StrictHostKeyChecking=yes` and has no source of truth for
host keys other than the operator's file. The file used for one repository is,
in order: the `known_hosts` of the matched scope, then the run-wide
`--build-ssh-known-hosts` or `CURATOR_BUILD_SSH_KNOWN_HOSTS`, then
`~/.ssh/known_hosts` if it exists as a regular file. A scope selects
authentication material, not who the destination is allowed to be, which is why
an explicit run-wide known-hosts file still applies to a scoped selection that
names none.

If none of the three yields a file, the fetch fails with
`build_repository_ssh_credential_missing` rather than trusting an unverified
host.

## Limits worth knowing

**An agent-less `--identity` pointing at a `*.pub` file cannot authenticate.**
The identity-only tail is `-o IdentitiesOnly=yes -o IdentityAgent=none -i
<path>`. With a public-key file and no agent there is nothing to sign with, so
the fetch fails at authentication time. Validation only proves the path resolves
to a regular file, so the identity-only menu entries and the
`--identity ~/.ssh/<key>.pub` lines of the fail-closed diagnostic are selectable
choices that do not work when the path is the public half. Point `--identity` at
a private key for the identity-only shape, or, when the key is loaded in an
agent, use the pinned-agent shape `--agent --identity ~/.ssh/<key>.pub`, which
is what the default menu entry offers.

**A repository whose canonical identity the scope grammar rejects cannot be
covered by a scope at all.** The suggested default scope is derived from the
repository namespace, and the scope grammar is narrower than the canonical
identity grammar. For an SSH build repository the path half never diverges: an
SSH repository path is already restricted to ASCII letters, digits, `.`, `_`,
`-`, and `/` before Git or SSH starts (`Spec §6.3`), which is exactly the
alphabet a scope segment admits, and a declaration such as
`ssh://git.example.com/team+infra/app.git` is refused at parse time. The host
half can diverge: a canonical host only has to match `[A-Za-z0-9][A-Za-z0-9.-]*`
lowercased, while a scope host must be dot-separated labels that neither start
nor end with a hyphen and are never empty. Hosts such as `git.example.com.`,
`git..example.com`, and `git-.example.com` therefore produce a suggested scope
the CLI refuses, and widening it to the host does not help, because the host is
not a valid scope either. Non-interactively the printed
`curator config build-ssh add` command errors when pasted. Interactively the
scope question keeps offering the same rejected default: pressing Enter is
refused with the scope rule and asks again, and any scope that would be accepted
cannot match the identity, so the run still fails closed afterwards. Cover such a
repository with `--build-ssh-agent` / `--build-ssh-identity` or the
`CURATOR_BUILD_SSH_*` environment instead.

By contrast an HTTPS identity may well carry a namespace the scope grammar
rejects, for example `git.example.com/team+infra/app`, because the HTTPS path
keeps the wider Unicode grammar of `Spec §6.3`. That never reaches the scope
suggestion: HTTPS repositories need no SSH selection.

**A prompted credential is not folded back into the in-memory configuration.**
The prompt writes to the configuration file, but the loaded configuration a run
holds is not updated. `curator install --all` captures the selection once and
loops over targets, so a repository covered by an answer given for target 1 is
asked about again for target 2. Answering twice is harmless, since the same
scope is simply replaced, but the "one scope covers every sibling" property
holds within a target and not across them.
