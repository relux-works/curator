# External build repositories: architecture decision and threat model, revision 4

**Task:** `TASK-260720-1nvomm`  
**Date:** 2026-07-27  
**Status:** Superseding architecture proposal for independent review  
**Supersedes:** `TASK-260720-1nvomm_external-build-repositories-architecture-v3.md`  
**Review input:** `TASK-260720-1nvomm_external-build-repositories-review-v3.md`  
**Scope:** Architecture only. No protocol, schema, manager, test, CLI, release, or prior decision artifact was changed.

## 1. Decision summary

External build repositories require a new, explicitly versioned protocol surface. They MUST NOT be folded into manifest schema 6 or broaden the accepted `go-v1` driver.

The recommended transition remains:

- manifest schema 7;
- a strict top-level `build_repositories` map;
- a new closed command driver, `go-repository-v1`;
- one fixed repository-root descriptor, `curator-build.json`, schema 1;
- build receipt schema 2 and install marker schema 3;
- a new conformance claim and release identity, recommended as protocol `1.0.0-rc.5` with claim schema 3;
- unchanged readers and meanings for manifest schemas 1 through 6, `go-v1`, receipt v1, and markers v1/v2.

The consuming skill owns the executable command name. The external repository owns only a logical target definition: driver, `build_root`, and `source_dir`. The manager continues to derive:

```text
artifact-relative path = bin/<manifest-command-key>[.exe]
staging path           = manager-private and implementation-specific
shim name              = <manifest-command-key>
```

`curator-build.json` therefore MUST NOT name a binary or output location. A descriptor target is a logical selector, not a filename. Allowing repository-controlled names or paths would reopen the output-selection boundary closed by schema 6.

An external source is first-class compiler input. Before an artifact-cache lookup or compiler child, the manager MUST resolve or load the exact effective Git commit, freeze its complete regular-file snapshot, validate and hash it, and audit it as a separate subject from the consuming skill. This ordering applies to real operations, cache hits, and dry-runs that claim source verification.

### 1.1 Changes required by the third independent review

This revision preserves every accepted revision-3 section and closes its final
three bounded interoperability findings:

1. Network and local sources now share one exact raw-object reader. The manager
   starts one direct, fingerprinted `git cat-file --batch` child against a
   manager-created private repository, sends only full lowercase object IDs,
   parses one byte-framed response grammar, and recomputes every consumed
   commit, tag, tree, and blob ID. `git ls-tree`, `--batch-command`, revision
   expressions, filters, textconv, mailmap, alternates, replace objects, lazy
   fetch, and every additional helper are forbidden.
2. Local packed-object admission is interoperable rather than implementation
   selected. V1 admits pack versions 2 or 3 and only pack-index version 2,
   under the repository's declared SHA-1 or SHA-256 hash family, with exact
   filename, header, fanout, object-name, offset, pack-checksum, and
   index-checksum validation. All admitted loose/pack/index bytes are copied
   inertly into a freshly initialized manager repository before trusted
   `cat-file` reads them. Every other pack/index version has one stable error.
3. Every reachable blob is scanned before audit. A byte-exact current Git LFS
   pointer or a narrowly recognized legacy/noncanonical equivalent fails
   `build_repository_git_lfs_unsupported`; the detector fixes the byte limit,
   line endings, version aliases, key and extension grammar, SHA-256 OID, size,
   ordering, and empty-blob behavior. Hydration and `git-lfs` execution remain
   forbidden.

The prior exact init/fetch/SSH controls, local repository layout rules,
declared/effective identity, mixed receipt/marker, audit ordering,
status/repair/GC, output, signing, future-driver, and board-impact decisions are
unchanged.

### 1.2 Justified gap and self-verification

**Missing piece:** schema 6 can compile only source stored in the skill snapshot. It cannot name, lock, audit, or cache compiler source from a separate Git repository.

**Requirement exposed by the scope delta:** Go CLI commands need named external build repositories, while source access, audit-before-compiler, no-hooks, manager-derived output, protected caches, currentness, rollback, signing separation, and future language drivers remain closed and deterministic.

**Consequence if omitted:** implementations would overload `go-v1`, compile an unbound checkout, or accept package-selected Git/build/output behavior. Cache keys and audit coverage would be incomplete.

**Gap closure:** schema 7, `curator-build.json`, `go-repository-v1`, receipt v2, and marker v3 define the missing source, target, identity, lifecycle, and compatibility surfaces without changing rc.4 meanings.

**Sections checked:** accepted contract sections 2 through 10, including its complete rejected/out-of-scope list. The contract excludes generic hooks, arbitrary argv/environment/output, unsafe build systems, physical cache-path standardization, receipt-only provenance, and unreviewed drivers. It does not exclude a new versioned external-source driver. This packet preserves every exclusion.

No further research or human product decision is open. Board refinement may occur only after this architecture passes independent review.

## 2. Manifest schema 7 and compatibility

Illustrative consuming manifest:

```json
{
  "schema_version": 7,
  "build_repositories": {
    "golden-tools": {
      "git": "https://github.com/example/golden-tools.git",
      "locked_commit": {
        "object_format": "sha1",
        "hex": "0123456789abcdef0123456789abcdef01234567"
      },
      "tag": "v1.4.0"
    }
  },
  "commands": {
    "golden-tool": {
      "type": "build",
      "driver": "go-repository-v1",
      "repository": "golden-tools",
      "target": "golden-tool"
    }
  },
  "capabilities": {
    "network": "none",
    "filesystem": "repo",
    "exec": "none",
    "secrets": "none",
    "env_read": []
  },
  "dependencies": {}
}
```

Normative recommendations:

1. `build_repositories` is legal only in schema 7 or later and is a strict map keyed by portable identifiers.
2. Each entry contains exactly `git` and `locked_commit`, plus OPTIONAL `tag`.
3. `locked_commit.object_format` is exactly `sha1` or `sha256`. `hex` is the full lowercase commit object ID: 40 hexadecimal characters for SHA-1 or 64 for SHA-256.
4. Branches, ranges, abbreviated IDs, `HEAD`, raw revision expressions, and package-selected local paths are forbidden.
5. Every declaration is selected by at least one command. Every repository command selects exactly one declaration and one descriptor target.
6. A command contains exactly `type`, `driver`, `repository`, and `target`. It cannot contain a program, argv, environment, output, filename, toolchain, signing identity, hook, generator, plugin, build script, or fallback.
7. The command driver equals the descriptor target driver. Unknown or mismatched drivers fail before cache lookup or compiler execution.
8. Schema 7 may mix local `go-v1` commands and external `go-repository-v1` commands. Local `build_roots` keep their schema-6 context-exclusion meaning. External repository files never enter agent context or runtime copying.
9. Schemas 1 through 6 reject all new fields and driver IDs. No rc.4 schema, receipt, marker, decision, or claim is reinterpreted.

## 3. Exact Git source grammar and immutable lock

### 3.1 Canonical network identity

Schema 7 reuses protocol core section 6.1 canonical source identity, narrowed to HTTPS and SSH:

- HTTPS form is `https://host/path`. Userinfo, password, port, query, fragment, percent escape, backslash, and empty path are forbidden.
- SSH URI form is `ssh://[user@]host/path`.
- SCP form is `[user@]host:path`.
- SSH username is optional ASCII matching `[A-Za-z0-9][A-Za-z0-9._-]{0,63}`. It affects the trusted SSH connection but is removed from canonical identity. HTTPS has no username.
- Host is ASCII `[A-Za-z0-9][A-Za-z0-9.-]*`, has no explicit port, and is lowercased.
- Repository path is valid Unicode scalar text encoded as UTF-8, has non-empty portable components, and contains no whitespace, `%`, `?`, `#`, backslash, empty component, `.` component, or `..` component.
- Leading and trailing path slashes are removed. Path case is preserved.
- Exactly one case-sensitive trailing `.git` is removed. `repo.GIT` is not changed.
- HTTPS, SSH URI, and SCP spellings that differ only by transport or permitted SSH username therefore produce the same `host/path` value.
- The canonical value is at most 4096 Unicode scalar values. Invalid network-looking forms are rejected and never treated as local.

The declared receipt identity records the resulting `{"kind":"network-git","value":"host/path"}` plus the selected transport. The effective identity applies the same algorithm to an operator network substitution.

### 3.2 Safe tag and development-ref grammar

The manifest `tag` is a tag name without the `refs/tags/` prefix. It is accepted only when:

- it is valid Unicode scalar text whose UTF-8 encoding is 1 through 255 bytes;
- it does not start or end with `/`, contain `//`, or end with `.`;
- no slash-separated component starts with `.` or ends with `.lock`;
- it contains none of NUL, ASCII control bytes, DEL, space, `~`, `^`, `:`, `?`, `*`, `[`, or backslash;
- it contains neither `..` nor `@{` and is not the single character `@`;
- the exact constructed `refs/tags/<tag>` passes the protocol-defined equivalent of `git check-ref-format` without normalization.

The manager never passes an unvalidated tag as a revision expression. Network
acquisition passes only exact `refs/tags/<tag>` or `refs/heads/<branch>` fetch
sources; manager code performs tag peeling after object-ID recomputation. If a
separately specified trusted Git verifier is ever used, its only permitted
revision spelling is `refs/tags/<tag>^{commit}` after validation and after
`--end-of-options` where supported.

Operator network substitutions use a structured ref:

- `revision` is a full lowercase object ID for the effective repository object format;
- `tag` uses the rules above and constructs `refs/tags/<value>^{commit}`;
- `branch` applies the same grammar to `refs/heads/<value>` and constructs only that ref;
- no other revision operators or reflog syntax are accepted.

### 3.3 Lock and optional tag assertion

The full object ID is the immutable lock. The tag is only a human-facing assertion:

1. The effective declared-source repository object format MUST equal `locked_commit.object_format`.
2. The object named by `locked_commit.hex` MUST exist and be a commit.
3. If `tag` is present, its tag ref MUST peel to that exact commit under `--no-replace-objects`; mismatch is `build_repository_ref_moved`.
4. Install, repair, status, and audit never rewrite a lock. An authoring command may propose a manifest change, but it is outside installation.
5. A server that does not permit fetching an unadvertised locked object may require the optional tag to retrieve it. Failure to obtain the exact locked commit is an availability error, never permission to widen the fetch or accept another commit.

## 4. Repository descriptor and output boundary

The root tree of the effective commit contains exactly one descriptor named `curator-build.json`, schema 1:

```json
{
  "schema_version": 1,
  "targets": {
    "golden-tool": {
      "driver": "go-repository-v1",
      "build_root": ".",
      "source_dir": "cmd/golden-tool"
    },
    "admin-tool": {
      "driver": "go-repository-v1",
      "build_root": "tools/admin",
      "source_dir": "tools/admin/cmd/admin"
    }
  }
}
```

The descriptor and every target are strict. For `go-repository-v1`:

1. A target contains exactly `driver`, `build_root`, and `source_dir`.
2. Paths are repository-relative. The single value `.` denotes the repository root; otherwise existing portable-path rules apply.
3. `source_dir` equals or is below the selected `build_root`.
4. `build_root` contains `go.mod` directly, and that file is the nearest ancestor `go.mod` of `source_dir`.
5. Targets may share one build root. Nested module roots are allowed only through explicit target selection and the nearest-module rule; discovery is forbidden.
6. Every non-standard module/package/source/embed/vendor input in the fixed `go list` graph stays below the selected build root. No input comes from the consuming skill, another external repository, a sibling module, a parent workspace, a host module cache, or the network.
7. The whole repository snapshot participates in raw validation, audit, and conservative source identity. Only the selected build root is compiler-visible.
8. The descriptor cannot contain a command name, output basename/path, install destination, alias, PATH edit, signing policy, credential, flag, environment, build system, hook, plugin, or arbitrary target argument.

Allowing `.` for an external repository does not weaken schema 6. The entire external snapshot is compile-only and never prompt-visible or runtime-copied; a schema-6 local root still cannot be `.`.

## 5. Declared and effective source identity

### 5.1 Source states

Every external command binds two source states:

- **Declared:** canonical manifest network identity, transport, immutable object-format/commit lock, and optional tag.
- **Effective:** exact source used after operator substitutions, including identity kind/value, transport where applicable, actual object format, full commit, substitution flag/type, and `curator-build-source-v1`.

For an unsubstituted source:

- `effective.substituted` is `false`;
- `effective.substitution` is absent;
- effective identity equals declared canonical identity;
- effective object format and commit equal the lock.

For a substitution:

- `effective.substituted` is `true`;
- `effective.substitution` is required and names exactly `local-path` or `network-git`;
- the declared state remains present and unchanged;
- effective identity, object format, full commit, and digest describe the substituted snapshot.

A network substitution uses the same canonical network identity and records its transport and structured ref. A local substitution has no network identity. Its effective identity is:

```text
kind  = operator-local-git
value = sha256(CCJ-1({
  "algorithm":"curator-operator-local-git-v1",
  "project":<canonical-project-identity>,
  "selector":<normalized project-relative substitution path>
}))
```

The selector uses `/`, valid Unicode scalars without normalization, removes `.` components, cancels an ordinary component followed by `..`, preserves unmatched leading `..`, and contains no empty component. This makes equivalent operator spellings stable across managers without placing an absolute host path in a receipt. The identity is not a network source identity or authorization token. Exact resolved bytes remain bound by object format, full commit, and build-source digest.

### 5.2 Skillfile.dev schema 2

Operator substitutions live only in ignored `Skillfile.dev.json` schema 2:

```json
{
  "schema_version": 2,
  "substitutions": {},
  "build_repository_substitutions": {
    "golden-skill": {
      "golden-tools": {
        "path": "../golden-tools"
      }
    }
  }
}
```

The outer key is the consuming skill; the inner key is its repository ID. An entry is exactly one:

- operator project-relative local Git `path`; or
- operator `git` plus one structured exact `ref`.

Rules:

1. Package data cannot create or change substitutions.
2. Local paths resolve relative to the canonical project root. The manager obtains the repository's exact HEAD commit; dirty and untracked worktree bytes are not compiler input.
3. A network substitution uses the same HTTPS/SSH grammar and safe structured-ref rules.
4. The substitution replaces source acquisition only. It cannot change repository ID, descriptor target, driver, command name, output, compiler policy, credentials, or signing.
5. Strict audit fails for any substitution. Advisory audit reports it and still audits the exact effective snapshot.
6. Status compares the current effective substitution state to marker v3. Removing, changing, or moving a substitution makes the installation non-current.

## 6. Closed Git child and object boundary

### 6.1 Trusted distribution, process graph, and clean environment

The manager resolves an operator-trusted Git distribution before reading a
package or substitution path. The absolute `git` executable, its complete
`GIT_EXEC_PATH`, HTTPS helper, and every Git subprogram that the fixed fetch can
start MUST be bundled, fingerprinted, or pinned by trusted operator policy. The
selected release family MUST be one tested against the driver vectors and MUST
support `--object-format`, `--ref-format=files`, `--no-lazy-fetch`,
`--no-write-fetch-head`, and `--no-auto-maintenance`. A newer or older unknown
family is rejected, not approximated.

The permitted child graph is exactly:

- manager -> absolute trusted `git` for manager-owned `init` and `fetch`;
- manager -> the same absolute trusted `git` for one manager-owned, read-only
  `cat-file --batch` session after network acquisition or inert local-store
  copying;
- `git fetch` -> regular executables below the fingerprinted `GIT_EXEC_PATH`
  that the selected Git family uses for fetch/pack ingestion;
- HTTPS only: the fingerprinted `git-remote-https` and an optional
  manager-owned credential broker;
- SSH only: a manager-owned binary SSH wrapper and one operator-trusted
  OpenSSH executable;
- no shell, arbitrary remote helper, local/source upload-pack, maintenance,
  submodule child, checkout filter, hook, LFS client, pager, editor, proxy
  command, package program, or produced artifact.

The `cat-file` child MUST NOT start another process. An observed local child
outside that graph is a conformance failure. The
package/descriptor can select only a validated HTTPS or SSH source spelling; it
cannot select the Git binary, exec path, helper, upload-pack, credential
broker, SSH client, object-reader arguments, or child arguments.

Every Git child uses a manager-owned empty CWD and an environment constructed
from empty. It contains only indispensable OS process variables, private
home/config/temp paths, normalized locale, the fingerprinted exec path, and:

```text
GIT_CONFIG_GLOBAL=<manager-owned zero-byte regular file>
GIT_CONFIG_SYSTEM=<manager-owned zero-byte regular file>
GIT_CONFIG_NOSYSTEM=1
GIT_NO_REPLACE_OBJECTS=1
GIT_NO_LAZY_FETCH=1
GIT_OPTIONAL_LOCKS=0
GIT_TERMINAL_PROMPT=0
GIT_PAGER=cat
GIT_PROTOCOL_FROM_USER=0
GIT_LITERAL_PATHSPECS=1
GIT_ATTR_NOSYSTEM=1
GIT_EXEC_PATH=<fingerprinted trusted Git exec path>
HOME=<operation-private>/home
XDG_CONFIG_HOME=<operation-private>/config
LC_ALL=C
LANG=C
```

For SSH it additionally sets only
`GIT_SSH=<absolute-manager-wrapper>`, `GIT_SSH_VARIANT=ssh`, and the
manager-private policy-FD number. `GIT_PROTOCOL` is absent and the fixed
configuration uses protocol version 0, so Git MUST NOT add an SSH
`SendEnv=GIT_PROTOCOL` option. For HTTPS, `GIT_ASKPASS` names either the
manager credential broker or a fail-closed manager broker.

The manager unsets `GIT_DIR`, `GIT_WORK_TREE`, `GIT_COMMON_DIR`,
`GIT_OBJECT_DIRECTORY`, `GIT_ALTERNATE_OBJECT_DIRECTORIES`, `GIT_NAMESPACE`,
`GIT_REPLACE_REF_BASE`, `GIT_ATTR_SOURCE`, `GIT_SSL_NO_VERIFY`,
`GIT_SSH_COMMAND`, `SSH_AUTH_SOCK`, `SSH_ASKPASS`, `GIT_CONFIG_COUNT`,
`GIT_CONFIG_KEY_*`, `GIT_CONFIG_VALUE_*`, all trace variables, all HTTP/HTTPS
proxy variables, and every other inherited `GIT_*`/`SSH_*` variable. `PATH` is
a manager-owned empty directory except for exact helper basenames that the
fingerprinted Git family cannot launch by absolute exec-path resolution.

HTTPS verifies TLS and does not follow redirects. An operator proxy or
credential may be supplied only through manager policy/broker state, never the
manifest, descriptor, source repository, or compiler environment.

### 6.2 Exact private repository initialization

Before each network attempt the manager exclusively creates a new empty
operation-private directory, an actual zero-entry template directory, empty
hooks directory, and private bare-repository path. It invokes exactly:

```json
[
  "<absolute-trusted-git>",
  "--git-dir=<operation-private>/repo.git",
  "-c",
  "init.defaultBranch=curator-invalid",
  "init",
  "--bare",
  "--quiet",
  "--template=<operation-private>/empty-template",
  "--object-format=<sha1-or-sha256>",
  "--ref-format=files"
]
```

There is no repository path operand, global/user/system template, initial
branch from configuration, shared mode, separate Git directory, worktree, or
source-provided config. After exit, manager code opens the result with
no-follow/rooted access and requires:

- one ordinary bare repository inside operation-private state;
- `core.repositoryformatversion`/`extensions.objectFormat` consistent with the
  requested object format and files refs;
- no remote, alternate, promisor, partial-clone, worktree, replace, graft,
  shallow, hook, filter, or unknown extension state;
- no link, reparse escape, special file, unexpected owner/ACL, or writable
  boundary outside the operation principal.

Failure discards the directory. The manager never repairs or adopts it.

### 6.3 Exact network fetch and ref flows

All fetches use a direct validated URL, a fresh private repository, one
manager-chosen destination under `refs/curator/`, and this exact common shape:

```json
[
  "<absolute-trusted-git>",
  "--git-dir=<operation-private>/repo.git",
  "--no-replace-objects",
  "--no-lazy-fetch",
  "--no-optional-locks",
  "-c",
  "protocol.allow=never",
  "-c",
  "protocol.<https-or-ssh>.allow=always",
  "-c",
  "protocol.version=0",
  "-c",
  "credential.helper=",
  "-c",
  "core.askPass=<manager-broker>",
  "-c",
  "core.hooksPath=<operation-private>/empty-hooks",
  "-c",
  "core.fsmonitor=false",
  "-c",
  "core.untrackedCache=false",
  "-c",
  "submodule.recurse=false",
  "-c",
  "fetch.recurseSubmodules=false",
  "-c",
  "maintenance.auto=false",
  "-c",
  "fetch.writeCommitGraph=false",
  "-c",
  "fetch.fsckObjects=true",
  "-c",
  "transfer.fsckObjects=true",
  "-c",
  "http.followRedirects=false",
  "-c",
  "http.sslVerify=true",
  "-c",
  "http.proxy=",
  "-c",
  "https.proxy=",
  "fetch",
  "--quiet",
  "--atomic",
  "--no-tags",
  "--no-recurse-submodules",
  "--no-auto-maintenance",
  "--no-write-fetch-head",
  "--no-write-commit-graph",
  "--refmap=",
  "--jobs=1",
  "--upload-pack=git-upload-pack",
  "--",
  "<validated-url>",
  "<one-manager-refspec>"
]
```

The selected transport is the only `protocol.*.allow=always` entry; the other
transport and every remote-helper protocol remain denied. There is no
`--filter`, `--server-option`, remote name, configured refspec, stdin refspec,
prune, mirror, depth/shallow option, tag auto-follow, or package-selected
argument. Top-level `git fetch` receives closed stdin and bounded captured
diagnostics.

The one refspec is selected as follows:

1. **Declared lock, primary attempt:**  
   `<full-locked-oid>:refs/curator/locked`. A full hexadecimal object name is
   the only non-ref source admitted. On success, the manager still parses and
   recomputes that object and requires a commit.
2. **Declared lock, exact-tag fallback:**  
   If and only if the primary attempt fails and the declaration contains a
   validated tag, discard the entire failed repository, initialize a new one,
   and fetch
   `refs/tags/<tag>:refs/curator/tag`. Manager code parses the resulting direct
   ref and any annotated-tag chain, recomputes every tag object, peels to a
   commit, and requires exact equality with the locked object ID. Mismatch is
   `build_repository_ref_moved`. There is no branch/all-tags fallback.
3. **Operator network substitution:**  
   A full `revision` fetches
   `<full-oid>:refs/curator/effective`; a `tag` fetches
   `refs/tags/<tag>:refs/curator/effective`; a `branch` fetches
   `refs/heads/<branch>:refs/curator/effective`. There is no fallback between
   forms. The declared repository lock's object format initializes the private
   repository and MUST match the substituted remote; the manager never retries
   under another hash family. The exact resulting commit becomes the effective
   receipt identity.

Every destination is fixed manager namespace text; remote data cannot choose a
local ref, refmap, refspec, helper, upload-pack, filter, or server option.
`FETCH_HEAD`, remote-tracking refs, local heads/tags, commit-graphs, and
maintenance state MUST remain absent. A nonzero fetch, unexpected local ref,
unexpected child, or any write outside the private repository fails and
discards the attempt.

### 6.4 Exact SSH wrapper and OpenSSH invocation

SSH sources retain the section-3 canonical identity but are further narrowed
for this driver: the raw repository path supplied to the remote shell MUST use
only ASCII letters, digits, `.`, `_`, `-`, and `/`, with the existing
non-empty/no-`.`/no-`..` component rules. A quote, shell metacharacter,
whitespace, escape, or non-ASCII byte is therefore rejected for SSH before
Git. HTTPS retains the section-3 path grammar.

Before fetch, the manager creates a protected read-only policy record and
passes an already-open descriptor number to the binary wrapper. The record
contains exactly:

- the expected Git destination, either `host` or `user@host`;
- the expected raw remote path from the validated SSH URI/SCP spelling;
- the expected remote command
  `git-upload-pack '<raw-path>'`;
- absolute operator-trusted SSH executable, empty SSH config, empty global
  known-hosts file, selected operator known-hosts file, timeout, and exactly
  one authentication mode.

Because the allowed path alphabet excludes `'`, the command has no escaping
branch. The wrapper does not accept a policy pathname from argv/environment and
does not consult package/project files.

`GIT_SSH_VARIANT=ssh` suppresses variant autodetection. With fixed Git protocol
0 and no port/IP-family option, the wrapper accepts exactly:

```text
argv[0] = <absolute-manager-wrapper>
argv[1] = <expected host or user@host>
argv[2] = <exact expected git-upload-pack command>
argc    = 3
```

It rejects `-G`, `-p`, `-4`, `-6`, `-o`, `SendEnv`, extra operands, a different
host/user/path/upload-pack, and every other shape. A tested Git family whose
actual wrapper call differs is unsupported until a new reviewed wrapper
revision is defined.

After byte-exact comparison, the wrapper directly executes the
operator-trusted SSH binary without a shell. The common argv is exactly:

```json
[
  "<absolute-operator-trusted-ssh>",
  "-F",
  "<manager-owned-empty-ssh-config>",
  "-T",
  "-o",
  "BatchMode=yes",
  "-o",
  "NumberOfPasswordPrompts=0",
  "-o",
  "PasswordAuthentication=no",
  "-o",
  "KbdInteractiveAuthentication=no",
  "-o",
  "PreferredAuthentications=publickey",
  "-o",
  "HostbasedAuthentication=no",
  "-o",
  "GSSAPIAuthentication=no",
  "-o",
  "StrictHostKeyChecking=yes",
  "-o",
  "UserKnownHostsFile=<operator-known-hosts>",
  "-o",
  "GlobalKnownHostsFile=<manager-owned-empty-known-hosts>",
  "-o",
  "CheckHostIP=no",
  "-o",
  "VerifyHostKeyDNS=no",
  "-o",
  "UpdateHostKeys=no",
  "-o",
  "ForwardAgent=no",
  "-o",
  "ForwardX11=no",
  "-o",
  "ClearAllForwardings=yes",
  "-o",
  "PermitLocalCommand=no",
  "-o",
  "ProxyCommand=none",
  "-o",
  "ProxyJump=none",
  "-o",
  "ControlMaster=no",
  "-o",
  "ControlPath=none",
  "-o",
  "ControlPersist=no",
  "-o",
  "RequestTTY=no",
  "-o",
  "EscapeChar=none",
  "-o",
  "EnableEscapeCommandline=no",
  "-o",
  "CanonicalizeHostname=no",
  "-o",
  "ConnectionAttempts=1",
  "-o",
  "ConnectTimeout=<manager-policy-seconds>",
  "<one-authentication-tail>",
  "<expected host or user@host>",
  "<exact expected git-upload-pack command>"
]
```

The authentication tail is exactly one of:

```text
operator identity file:
  -o IdentitiesOnly=yes -o IdentityAgent=none -i <operator-selected-identity>

operator SSH agent:
  -o IdentitiesOnly=no -o IdentityFile=none
  -o IdentityAgent=<operator-selected-agent-socket>
```

The private `HOME` contains no default identity files. Package data cannot
select the identity, agent, known-hosts, timeout, SSH config, host-key policy,
or option. The wrapper passes no ambient `SSH_AUTH_SOCK`.

The top-level fetch has no inherited terminal stdin. SSH standard input/output
are necessarily the private Git-upload-pack protocol pipes created by trusted
Git; they are not a TTY, user input, package file, or prompt channel. The
wrapper neither parses nor replaces that protocol stream. This is the only
stdin carried through SSH; closing it would make Git transport impossible.

### 6.5 Local substitution admission: deliberately narrow v1

Local substitutions execute no Git command, upload-pack, hook, filter, helper,
or LFS process from the source repository. Manager code treats repository
administration files and object stores as untrusted data.

V1 admits only an ordinary non-bare worktree whose substitution root is a
link-free directory and whose `.git` entry is a link-free directory directly
below that root. Stable pre-audit rejections are:

- `.git` is a regular gitfile: `build_repository_local_gitfile_unsupported`;
- `.git` is absent and the selected path is a bare repository:
  `build_repository_local_bare_unsupported`;
- `.git/commondir`, `.git/worktrees`, or `.git/config.worktree` exists:
  `build_repository_local_linked_worktree_unsupported`;
- `.git` or any admitted administration/object component is a link, reparse
  point, special file, multiply linked mutable file, or containment escape:
  `build_repository_local_layout_unsafe`.

The manager never follows or opens a gitfile/commondir target in v1. This makes
their containment rule exact: no path outside `<substitution-root>/.git` is
admitted. Supporting ordinary gitfiles, linked worktrees, submodule worktrees,
or bare repositories requires a new admission algorithm and conformance
vectors, not a fallback to Git discovery.

#### 6.5.1 Byte-level config admission

The manager opens `.git/config` no-follow as a bounded regular file and parses
this data-only subset:

1. bytes are valid UTF-8 with LF or CRLF records, optional final line ending,
   and no BOM, NUL, lone CR, control byte other than tab in a value, or
   backslash line continuation;
2. section names and variable names are ASCII case-insensitive
   `[A-Za-z][A-Za-z0-9-]*` tokens; subsections are double-quoted UTF-8 with
   only `\\` and `\"` escapes;
3. values are one-line unquoted text or one double-quoted value with only
   `\\`, `\"`, `\n`, `\t`, and `\b` escapes; comment markers are recognized
   only outside quotes; malformed/unsupported syntax rejects the repository;
4. `[include]` and `[includeIf "..."]` are rejected before resolving a path;
5. duplicate security-relevant keys reject even when values match.

Other well-formed non-extension keys are inert: the manager records but never
applies them. Admission reads only:

- `core.repositoryformatversion`, required as integer `0` or `1`;
- `core.bare`, required `false`;
- `extensions.*`;
- every `remote "<name>".promisor` and
  `remote "<name>".partialCloneFilter`.

Allowed format states are exactly:

- SHA-1: repository format `0`, no `extensions.*`;
- SHA-256: repository format `1`, exactly
  `extensions.objectFormat=sha256`, with optional
  `extensions.refStorage=files`.

Boolean values are accepted only as case-insensitive `true`/`yes`/`on`/`1` or
`false`/`no`/`off`/`0`; an absent value means true as in Git config.

`extensions.refStorage=reftable`, `extensions.partialClone`,
`extensions.worktreeConfig`, `extensions.compatObjectFormat`, `noop`,
`preciousObjects`, every unknown extension, a true promisor value, or a
non-empty partial-clone filter fails
`build_repository_local_format_unsupported`. Reftable is deliberately rejected;
dummy `HEAD`/`refs` files are never mistaken for the files backend.

#### 6.5.2 Files-ref parsing and selected HEAD

`HEAD`, every loose ref below `.git/refs`, and optional `.git/packed-refs` are
bounded, link-free regular files. The parser:

- accepts `HEAD` as either one full lowercase object ID or
  `ref: refs/heads/<safe-name>` plus exactly one LF/CRLF or EOF;
- permits only one level of HEAD symbolic reference; the selected loose/packed
  head itself contains one full lowercase object ID, never another symref;
- validates every ref name with the section-3 safe ref grammar and the
  protocol-defined equivalent of `git check-ref-format`;
- parses `packed-refs` as an optional `# pack-refs with:` header whose unique
  space-separated traits are drawn only from `peeled`, `fully-peeled`, and
  `sorted`, then
  `<full-oid> SP <refname>` records and optional `^<full-oid>` peeled records
  only immediately after a tag record; it rejects duplicates, unsupported
  headers, wrong hash lengths/case, malformed lines, and ref/peeled conflicts;
- rejects any loose or packed `refs/replace/*` entry, regardless of whether it
  is reachable or selected.

Loose refs have Git's normal precedence over packed refs, but contradictory
duplicate loose files and platform-colliding ref paths reject rather than
choose. The selected HEAD object ID is the effective commit recorded in the
receipt. Dirty index/worktree/untracked bytes are never opened as source.

#### 6.5.3 Exact object inventory and interoperable pack admission

Before interpreting an object, the manager freezes one inventory generation and
copies the allowed bytes into operation-private inert holding state. It
re-stats and rehashes every admitted source file after copy; any addition,
removal, rename, metadata identity change, or byte change aborts.

Allowed object-store entries are exactly:

- loose objects at
  `objects/[0-9a-f]{2}/[0-9a-f]{38}` for SHA-1 or
  `objects/[0-9a-f]{2}/[0-9a-f]{62}` for SHA-256;
- zero or more exact
  `objects/pack/pack-<full-object-format-hex>.pack` and matching `.idx` pairs.

Every loose/pack/index entry is a singly linked regular file. A pack without
one matching index, duplicate basename, unexpected suffix, or hash-family
mismatch rejects. V1 rejects all other object sidecars/state, including:

```text
objects/info/*
objects/pack/*.promisor
objects/pack/*.keep
objects/pack/*.bitmap
objects/pack/*.rev
objects/pack/*.mtimes
objects/pack/multi-pack-index*
shallow
info/grafts
```

It also rejects alternates/http-alternates, cruft/promisor packs, namespaces,
and unexpected object directories. The source `.idx` is captured only as
untrusted inert input. Manager code validates its fanout, sorted names,
offsets, object count, pack checksum, and index checksum against the paired
pack before using an offset.

The portable local v1 subset is exact:

1. `H` is 20 for SHA-1 and 32 for SHA-256. Every object ID and checksum below
   has exactly `H` bytes and uses that repository hash.
2. A `.pack` begins with ASCII `PACK`, unsigned 32-bit big-endian version `2`
   or `3`, and an unsigned 32-bit big-endian object count `N`. Its last `H`
   bytes equal `HASH(all preceding pack bytes)`. The lowercase checksum hex is
   exactly the `pack-<hex>` basename.
3. A paired `.idx` is version 2 only: bytes `ff 74 4f 63`, unsigned 32-bit
   big-endian version `2`, 256 monotonically nondecreasing unsigned 32-bit
   fanout entries, and final fanout value `N`.
4. The index then contains exactly `N` object names of width `H`, strictly
   increasing by unsigned byte order; `N` unsigned 32-bit CRC32 entries; and
   `N` unsigned 32-bit offsets. An offset with its high bit clear is a pack
   offset. An offset with its high bit set indexes the following table of
   unsigned 64-bit big-endian large offsets after clearing that bit. If any
   large offset is referenced, the referenced indices are unique and exactly
   the dense range `0..M-1`; otherwise `M` is zero. The table contains exactly
   `M` entries. Every resolved pack offset is unique, is at least 12, and
   begins before the pack trailer. Sorting the resolved offsets defines each
   packed-entry byte range through the next offset or the pack trailer; the
   corresponding index CRC32 MUST equal zlib CRC32 over that complete range.
5. The index ends with the paired pack checksum followed by an index checksum
   equal to `HASH(all preceding index bytes)`. There are no omitted or trailing
   bytes, duplicate object names, duplicate offsets, decreasing fanout values,
   count mismatches, or cross-family hash widths.
6. Pack versions other than 2/3, index versions other than 2, version-1
   indexes, unknown sidecars, checksum failures, or malformed tables all fail the single stable error
   `build_repository_local_object_format_unsupported`.

This admission parser validates the physical container only; it does not trust
the index's claim about an object's identity or decompress source-selected
objects itself. After validation, the manager creates a fresh private bare
repository using section 6.2 with the admitted object format, copies only the
validated loose objects and paired `.pack`/`.idx` files beneath its own
`objects/` directory with no-follow/rooted writes, and seals that directory
read-only for the object reader. It writes manager-generated repository config
and no source config, refs, hooks, attributes, LFS data, alternates, commit
graph, multi-pack index, bitmap, reverse index, or maintenance state.

The manager then uses the common trusted `cat-file` contract in section 6.6.
An OFS_DELTA base remains within its paired pack; a REF_DELTA base may be in
another admitted pack or loose object in the same sealed private store. No base
may come from an alternate, lazy/promisor fetch, source worktree, or network.
Missing/cyclic/oversized delta resolution makes `cat-file` fail and therefore
fails complete-source proof; the manager never widens the store.
Every consumed object is independently recomputed as:

```text
object_id = HASH(type || " " || decimal_byte_length || NUL || exact_content)
```

under SHA-1 or SHA-256 and MUST equal both the requested and returned full ID.
Thus trusted Git performs bounded decompression/delta resolution from inert
bytes, while manager code retains authority over identity, graph semantics, and
materialized snapshot bytes. An implementation unable to validate or stage
this exact subset MUST reject it; it MUST NOT invoke the source repository's
Git configuration or select a private implementation-specific subset.

### 6.6 Common raw-object reader and complete-object proof

Network fetches already end in a manager-created private repository. Local
substitutions are copied into an equivalent manager-created private repository
under section 6.5.3. From this point the two source kinds use one contract.

#### 6.6.1 Exact `cat-file --batch` process

After all physical-store checks and before audit or artifact-cache lookup, the
manager launches exactly one direct child:

```json
[
  "<absolute-trusted-git>",
  "--git-dir=<operation-private>/repo.git",
  "--no-replace-objects",
  "--no-lazy-fetch",
  "--no-optional-locks",
  "-c",
  "core.hooksPath=<operation-private>/empty-hooks",
  "-c",
  "core.fsmonitor=false",
  "-c",
  "core.untrackedCache=false",
  "-c",
  "maintenance.auto=false",
  "cat-file",
  "--batch=%(objectname) %(objecttype) %(objectsize)"
]
```

The child uses section 6.1's empty CWD and clean environment. In particular
`GIT_NO_REPLACE_OBJECTS=1`, `GIT_NO_LAZY_FETCH=1`,
`GIT_OPTIONAL_LOCKS=0`, empty global/system config, no alternates, no worktree,
no object-directory override, no credential/SSH/HTTP variables, and a
manager-owned empty `PATH` remain in force. Its repository and object store are
read-only. Standard error is bounded diagnostic data. The child has a
deadline, output/object/aggregate-byte limits, process/memory/disk limits, and
OS network denial. It may not create a child process or modify repository,
audit, cache, or installation state.

The exact negative option contract is equally normative: the argv contains no
`--batch-command`, `--batch-all-objects`, `--buffer`, `--follow-symlinks`,
`--unordered`, `--textconv`, `--filters`, `--filter`, `--use-mailmap`,
`--mailmap`, `--path`, `-Z`, or `-z`. There is no `git ls-tree`, `rev-parse`,
`show`, `archive`, `checkout`, `read-tree`, `verify-pack`, or source-selected
Git command. Absence of the transformation flags means `cat-file --batch`
returns raw uncompressed object contents, not checkout/filter/textconv/mailmap
output.

The manager sends requests sequentially. Each request is exactly:

```text
<full-lowercase-object-id> LF
```

The ID is 40 lowercase hexadecimal bytes for SHA-1 or 64 for SHA-256 and is
either the already validated selected ID or an ID parsed from a recomputed
commit, tag, or tree object. No abbreviated ID, whitespace, object/path suffix,
revision operator, ref, command word, NUL, CR, or second line is admitted.

For each request, stdout is parsed as exactly:

```text
<same-full-lowercase-object-id> SP <type> SP <decimal-size> LF
<exactly decimal-size content bytes> LF
```

`type` is one ASCII token from `commit`, `tag`, `tree`, or `blob`. Size is `0`
or `[1-9][0-9]*`, fits unsigned 64-bit and manager policy, and has no sign or
leading zero. The returned ID MUST byte-equal the request. The manager reads
exactly the declared content size and one terminal LF before issuing another
request. `missing`, `ambiguous`, `excluded`, `submodule`, unexpected type,
malformed/truncated/extra bytes, early EOF, trailing output after stdin close,
nonzero exit, timeout, or limit violation fails
`build_repository_incomplete_source`. Stdin closes after the last required
object; a clean child exits zero.

The manager bounds unique object count, individual object bytes, total expanded
bytes, tree depth, path depth/length, tag peel depth, diagnostics, and wall
time. Duplicate graph references reuse already verified bytes by full ID and
do not consume another logical object slot. Limits are manager policy with
portable minimums fixed by conformance; exceeding any limit fails before audit
and never creates a partial snapshot claim.

#### 6.6.2 Manager parsing and graph proof

For both network and local sources, manager code—not Git pretty-printers—then:

1. begins with the exact full ID selected by the declared lock, fetched tag,
   structured operator ref, or manager-parsed local HEAD; no ref or revision
   spelling reaches `cat-file`;
2. for an exact-tag flow, reads and recomputes each `tag`, validates its
   `object <full-id>` and `type tag|commit` headers, rejects cycles/duplicate
   peel IDs, and requires the terminal commit to equal the manifest lock where
   the declared tag assertion applies;
3. hashes the terminal selected object as
   `type || " " || decimal_byte_length || NUL || exact_content` under the
   repository hash, requires equality with the requested/returned ID, and
   requires object type `commit`;
4. parses the commit's exact `tree <full-id>` header, then recursively requests
   only IDs found in already verified tree bytes;
5. parses each tree as the Git binary sequence
   `<octal-mode> SP <name-bytes> NUL <H-byte-object-id>`, consuming the entire
   object with no empty/duplicate name, slash, NUL, `.`/`..`, invalid UTF-8,
   nonportable component, or platform collision;
6. accepts only tree mode `40000` and blob modes `100644`/`100755`; rejects
   symlink `120000`, gitlink `160000`, unknown modes, and mode/type mismatch;
7. reads and recomputes every reachable tree and blob, proving every required
   object locally available without replace, graft, alternate, promisor,
   lazy-fetch, checkout, attribute, filter, textconv, mailmap, or archive
   transformation;
8. runs the all-blob LFS detector in section 6.6.3 before materialization or
   audit;
9. materializes exact verified blob bytes into a frozen regular-file snapshot
   with manager filesystem code, preserving only the executable/non-executable
   distinction from `100755`/`100644`.

The source repository's commit message, author, committer, parent list, and tag
message remain untrusted data. They are bounded and structurally parsed only as
needed to identify the root tree and tag target; they do not select behavior.
Gitlinks fail before reading `.gitmodules`.

#### 6.6.3 Exact all-blob Git LFS rejection

The manager scans every reachable blob exactly once, regardless of path,
extension, mode, selected build root, `.gitattributes`, or whether the blob is
currently selected by `go list`. This happens after object-ID recomputation and
before snapshot materialization, audit, cache lookup, or compiler execution.
The detector never invokes `git-lfs`, a filter, or the network.

A zero-byte blob is not rejected: its compiler bytes are already complete and
the Git LFS specification treats empty content as pass-through. A non-empty
blob is a detector candidate only when its byte length is less than 1024. A
blob of 1024 bytes or more is not a Git LFS pointer under this protocol.

The **canonical pointer grammar** is:

1. bytes are valid UTF-8 and ASCII-only, contain no BOM/NUL/CR, and end in
   exactly one LF;
2. every line is `<key> SP <value> LF` with exactly one separating ASCII space,
   non-empty key/value, and no other ASCII control byte;
3. keys are unique and match `[a-z0-9][a-z0-9.-]*`; `version` is first and all
   later keys are in strictly increasing unsigned-ASCII order;
4. required keys occur exactly once with exact values:
   `version https://git-lfs.github.com/spec/v1`,
   `oid sha256:<64-lowercase-hex>`, and
   `size <0-or-nonzero-decimal-without-leading-zero>`;
5. size fits unsigned 64-bit;
6. every optional key is an extension
   `ext-<priority>-<name>`, where priority is one ASCII digit `0` through `9`,
   name matches `[a-z0-9][a-z0-9.-]*`, priorities are unique, and the value is
   `sha256:<64-lowercase-hex>`.

For defense in depth, the manager also recognizes one
**legacy/noncanonical pointer grammar**. It has the same length, UTF-8,
ASCII-only, unique-key, required-key, extension, OID, and size rules, but:

- lines may use consistently LF or CRLF and may omit the final line ending;
- required/extension lines may be unsorted, although `version` remains first;
- `version` may be the current value or one exact historical Git LFS alias:
  `https://hawser.github.com/spec/v1` or `http://git-media.io/v/2`.

Mixed line endings, blank lines, tabs, multiple spaces, leading/trailing
whitespace, duplicate keys/priorities, unknown non-extension keys, signed or
zero-padded size, uppercase/non-64-hex OID, a different version, malformed
extension, or bytes outside this closed grammar are ordinary blob content and
do not trigger this detector. Canonical and recognized legacy/noncanonical
matches both fail the same stable pre-audit error:

```text
build_repository_git_lfs_unsupported
```

The diagnostic may identify `canonical` versus `legacy-or-noncanonical` and
the repository-relative path, but MUST NOT print the pointer OID as a fetch
instruction. The manager never hydrates it from `.git/lfs`, an LFS endpoint, a
filter, or local operator state. Rejecting all recognized pointers avoids
compiling pointer text through `go:embed`; unrelated small text remains valid.

No audit or artifact-cache lookup begins until the object proof, LFS scan,
materialization, whole-snapshot validation, and
`curator-build-source-v1` digest succeed. Object bytes are transport input, not
provenance, and the object store is never agent-facing, runtime-copied, or used
directly as build source.

## 7. Snapshot identity and audit equivalence

The consuming skill snapshot and each effective external repository snapshot are independent untrusted audit subjects.

For every external repository selected by the active plan:

1. validate the whole repository snapshot;
2. compute `curator-build-source-v1` over every regular file, including `curator-build.json` and files outside the selected build root;
3. parse the descriptor and select the target;
4. audit the complete external snapshot under declared/effective identity, object format, effective commit, build-source digest, target, and substitution state;
5. apply allowlist, revocation, registry, tag-lock, and audit policy gates;
6. only then inspect an artifact-cache candidate or start a compiler child.

Registry evidence for the skill does not attest the external repository. External evidence does not attest the skill. Audit cache reuse is allowed only when its subject tuple exactly matches the declared/effective state and snapshot digest; it does not bypass snapshot admission.

A syntax-only offline check that cannot obtain an exact protected snapshot warns `build_repository_unverified_offline` and makes no source/audit/build claim. Install, update, repair, and coverage-claiming audit fail before mutation when the exact effective snapshot cannot be obtained and audited.

Dry-run may use network and removable operation-private fetch state, but it leaves no persistent repository, snapshot, audit response, lock, receipt, marker, or cache mutation. It must validate and audit the exact snapshot before reporting cache-hit or would-build.

## 8. Receipt v2, cache identity, and mixed schema-7 builds

### 8.1 Unsubstituted external receipt input

Receipt-v2 input binds the complete source state:

```json
{
  "schema_version": 2,
  "driver": "go-repository-v1",
  "source": {
    "repository": "golden-tools",
    "declared": {
      "identity": {
        "kind": "network-git",
        "value": "github.com/example/golden-tools"
      },
      "transport": "https",
      "locked_commit": {
        "object_format": "sha1",
        "hex": "0123456789abcdef0123456789abcdef01234567"
      },
      "tag": "v1.4.0"
    },
    "effective": {
      "identity": {
        "kind": "network-git",
        "value": "github.com/example/golden-tools"
      },
      "transport": "https",
      "object_format": "sha1",
      "commit": "0123456789abcdef0123456789abcdef01234567",
      "substituted": false,
      "build_source": {
        "algorithm": "curator-build-source-v1",
        "content_sha256": "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
      }
    },
    "descriptor": {
      "path": "curator-build.json",
      "target": "golden-tool"
    }
  },
  "command": "golden-tool",
  "build_root": ".",
  "source_dir": "cmd/golden-tool",
  "target": {
    "goos": "darwin",
    "goarch": "arm64",
    "tuning": {
      "GOARM64": "v8.0"
    }
  },
  "toolchain": {
    "algorithm": "curator-go-toolchain-v1",
    "go_relpath": "bin/go",
    "go_version": "go version go1.26.1 darwin/arm64",
    "content_sha256": "sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
  },
  "policy": {
    "module_mode": "vendor",
    "network": "none",
    "workspace": false,
    "cgo": false,
    "compiler_directives": "reject-nonstandard-cgo-import-dynamic-v1",
    "target_mode": "native",
    "link_mode": "internal",
    "libgcc": "none",
    "package_assembly": false,
    "host_objects": false,
    "telemetry": "off-private",
    "source_kind": "locked-external-git-v1"
  }
}
```

### 8.2 Substituted external receipt input

Only the source portion changes for a local substitution:

```json
{
  "repository": "golden-tools",
  "declared": {
    "identity": {
      "kind": "network-git",
      "value": "github.com/example/golden-tools"
    },
    "transport": "https",
    "locked_commit": {
      "object_format": "sha1",
      "hex": "0123456789abcdef0123456789abcdef01234567"
    },
    "tag": "v1.4.0"
  },
  "effective": {
    "identity": {
      "kind": "operator-local-git",
      "value": "sha256:eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"
    },
    "object_format": "sha1",
    "commit": "89abcdef0123456789abcdef0123456789abcdef",
    "substituted": true,
    "substitution": {
      "type": "local-path"
    },
    "build_source": {
      "algorithm": "curator-build-source-v1",
      "content_sha256": "sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
    }
  },
  "descriptor": {
    "path": "curator-build.json",
    "target": "golden-tool"
  }
}
```

A network substitution instead uses `identity.kind: "network-git"`, its canonical value and transport, and `substitution: {"type":"network-git","ref":<structured-ref>}`.

The complete receipt-v2 input is canonicalized with CCJ-1. Its SHA-256 is the artifact cache key. The key therefore cannot describe declared provenance while compiling substituted bytes.

Receipt v2 retains exact canonical stored bytes, manager-derived artifact path, artifact SHA-256, and byte size. It has no trust/provenance boolean. Protected-state admission remains out of band.

### 8.3 Mixed command behavior

One schema-7 installation may contain both command kinds:

- local `go-v1` commands keep receipt v1 and the consuming skill's raw `build_source`;
- external `go-repository-v1` commands use receipt v2 and per-command external source state;
- the two drivers never share receipt interpretation or logical cache keys;
- marker v3 represents both simultaneously;
- top-level marker `build_source` is present exactly when at least one active local `go-v1` command exists and retains schema-6 meaning;
- external-only installations omit top-level skill `build_source`; each external build entry carries its source identity and digest.

No field is inferred from a driver name when parsing a receipt or marker; schemas use explicit conditional branches.

## 9. Marker v3, status, repair, deduplication, and GC

Illustrative mixed marker excerpt:

```json
{
  "schema_version": 3,
  "skill_schema_version": 7,
  "build_source": {
    "algorithm": "curator-build-source-v1",
    "content_sha256": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
  },
  "build_roots": [
    "build"
  ],
  "builds": {
    "local-helper": {
      "driver": "go-v1",
      "receipt_schema_version": 1,
      "cache_key": "sha256:1111111111111111111111111111111111111111111111111111111111111111",
      "receipt_sha256": "sha256:2222222222222222222222222222222222222222222222222222222222222222",
      "artifact_sha256": "sha256:3333333333333333333333333333333333333333333333333333333333333333",
      "artifact_path": "bin/local-helper"
    },
    "golden-tool": {
      "driver": "go-repository-v1",
      "receipt_schema_version": 2,
      "repository": "golden-tools",
      "declared_identity": {
        "kind": "network-git",
        "value": "github.com/example/golden-tools"
      },
      "declared_locked_commit": {
        "object_format": "sha1",
        "hex": "0123456789abcdef0123456789abcdef01234567"
      },
      "declared_tag": "v1.4.0",
      "effective_identity": {
        "kind": "network-git",
        "value": "github.com/example/golden-tools"
      },
      "object_format": "sha1",
      "commit": "0123456789abcdef0123456789abcdef01234567",
      "substituted": false,
      "build_source": {
        "algorithm": "curator-build-source-v1",
        "content_sha256": "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
      },
      "descriptor_target": "golden-tool",
      "cache_key": "sha256:4444444444444444444444444444444444444444444444444444444444444444",
      "receipt_sha256": "sha256:5555555555555555555555555555555555555555555555555555555555555555",
      "artifact_sha256": "sha256:6666666666666666666666666666666666666666666666666666666666666666",
      "artifact_path": "bin/golden-tool"
    }
  }
}
```

### 9.1 Snapshot and artifact keys

- Protected external snapshot key: effective identity kind/value, effective object format, full effective commit, and external build-source digest.
- Artifact key: SHA-256 of the entire receipt input, including both declared and effective states, command key, descriptor target, build root/source directory, target, toolchain, and policy.
- Snapshot bytes may be deduplicated only when the complete snapshot key matches. Audit decisions remain subject-specific.
- Artifacts may be reused only when the complete receipt input and protected entry validate. Different declarations, substitutions, commands, targets, toolchains, or policy revisions do not alias.
- Physical object, snapshot, artifact, receipt, lock, and quarantine paths remain implementation-specific.

### 9.2 Read-only status

Status MUST NOT fetch, repair, quarantine, alter permissions, refresh a snapshot, invoke a compiler, or execute an artifact.

It:

1. derives the current effective plan from manifest and operator dev configuration;
2. validates marker v3 and substitution equality;
3. revalidates every referenced protected skill/external snapshot and artifact boundary;
4. recomputes available build-source identities, receipt inputs/keys, receipt hashes, and artifact relationships;
5. structurally validates manager-owned shims and PATH exposure.

Outcomes:

- an identity, commit, digest, target, receipt, artifact, shim, or substitution mismatch is non-current;
- an untrusted or corrupt snapshot/artifact boundary is non-current;
- a missing required snapshot is non-current when the marker claims it should exist;
- when existing profile semantics distinguish unavailable evidence from drift, a source that cannot be proved read-only may be reported `unknown`, but MUST NOT be reported current and `status --check` MUST return non-zero;
- status never contacts the remote merely to test availability or tag movement.

### 9.3 Repair

Repair resolves the current effective operator substitution if present, otherwise the declared locked source. It reacquires into operation-private state, repeats object proof, validation, identity, and audit, and rebuilds or reuses only after protected admission. It never adopts an untrusted snapshot or artifact by changing permissions.

If repair cannot obtain the exact source, it fails before installation mutation. Publication, marker/shim changes, rollback, and consumer update use the existing manager-home mutation lock and journal.

### 9.4 Garbage collection roots

GC revalidates cache boundaries and, under the manager-home mutation lock:

- valid marker v3 files mark every referenced local skill raw snapshot, external frozen snapshot key, receipt/artifact key, and manager-generated shim relation;
- in-flight transaction journals mark every staged/published snapshot and artifact needed for commit or rollback;
- valid marker v1/v2 behavior remains unchanged;
- receipt content alone is never a live root;
- unreadable markers/journals or unprovable cache boundaries retain uncertain entries conservatively;
- GC never executes or adopts source or artifact bytes.

## 10. Deterministic build, install, PATH, and rollback

`go-repository-v1` reuses the accepted `go-v1` trusted Go session, environment, full dependency checks, directive/native-input rejection, fixed `go list`/`go build` vectors, internal-link policy, artifact checks, and no-execution rule. Its source acquisition and receipt semantics are distinct, so the driver ID is distinct.

Normative order:

1. Validate schema-7 manifest, URL/lock/tag grammar, substitutions, descriptor-independent collisions, and static local build-root exclusions.
2. Resolve every exact effective external commit into operation-private state.
3. Close the Git object boundary, freeze the whole snapshot, compute its digest, parse the descriptor, and validate target/module containment.
4. Audit consuming skill and every external repository independently.
5. Probe/fingerprint the trusted Go toolchain and derive all local receipt-v1 and external receipt-v2 inputs.
6. Inspect protected artifact candidates. On real misses, run fixed preflight/build into operation-private staging. Dry-run stops before source-aware Go commands as defined by the accepted driver policy.
7. Validate/hashes artifacts and generate canonical receipts/marker.
8. Under the home mutation lock, recover/revalidate, publish immutable snapshots/artifacts, and commit all targets with the consumer ledger last.
9. On failure, reverse rollback under the same lock. Pre-publication source/audit/build failures leave live state unchanged.

PATH and shims are verified structurally:

- command key passes existing portable/collision rules;
- shim is manager-owned and points exactly to the marker-selected protected artifact;
- artifact is one regular singly linked executable matching receipt path/hash;
- no shim points into a Git object database, frozen source snapshot, checkout, staging directory, or script runtime;
- environment exposure contains only the manager-generated bin directory;
- validation never calls a shell, `command -v`, or the built output.

Several commands may select one logical descriptor target, but each manifest command key derives its own output and cache input. The repository still cannot request aliases, sidecars, post-build copies, or secondary PATH entries.

## 11. Signing and notarization boundary

Signing identities and notarization credentials are operator/platform secrets. They MUST NOT appear in a manifest, descriptor, repository, compiler environment, receipt trust field, or marker.

The first `go-repository-v1` release performs no manager post-signing. It does not invoke `codesign`, `signtool`, timestamping, or a notary service.

A future local signer requires a separately reviewed profile/driver revision with:

- operator-only opaque keychain/certificate/HSM/service identity;
- fixed signer executable, argv, entitlements/options, process graph, and network policy;
- signer policy/tool identity and signed bytes in cache/receipt identity;
- no package entitlements, certificate selector, timestamp URL, or arguments;
- private staging, no artifact execution, protected publication, and rollback.

Until standardized, a platform policy requiring local signing rejects source-built artifacts.

Apple Developer ID signing/notarization and Windows Authenticode/timestamping remain controlled release-pipeline work. A future prebuilt driver may verify signed release artifacts under its own exact digest and attestation policy; local build receipts do not become release provenance.

Linux remains a separate later native-validation task. No Linux conformance claim is made until the same driver, sandbox/resource, shim, and lifecycle vectors pass natively.

## 12. Future closed-driver rule

The external Git envelope may be shared, but compiler semantics are never generic. Every language adds a new driver ID and independent protocol/security review covering:

- complete compiler-visible inputs and dependency containment;
- trusted toolchain/sysroot/runtime identity;
- fixed process graph, environment, flags, target, output, and signer boundary;
- rejection of hooks, plugins, macros, generators, annotation processors, build tasks, recipes, response files, package linkers, and produced-program execution;
- offline dependency model and exact source identity;
- receipt/marker/cache identity, dry-run, audit-before-build, rollback/recovery, status/repair/GC, and platform vectors.

Consequences remain:

- Rust: Cargo `build.rs` and procedural macros are rejected; any direct `rustc` driver needs a closed dependency/link model.
- Swift: SwiftPM manifests/plugins/macros are rejected; direct `swiftc` must own SDK, target, graph, linker, and macro policy.
- Kotlin/JVM: Maven/Gradle, annotation processors, and compiler plugins are rejected; a direct driver owns class/module path, packaging, and JDK/Kotlin identity.
- C/C++: Make/CMake/Meson/Ninja are rejected; a direct driver owns sources, preprocessing, headers/sysroot, response files, compiler/linker plugins, native libraries, and output.
- .NET: MSBuild, analyzers, and source generators are rejected; a direct driver owns references, runtime packaging, native host, and signing.

Adding an external repository does not make an unsafe build frontend safe.

## 13. Threat model

### 13.1 Adversary-controlled

- consuming skill and manifest;
- declared Git URL, refs/tags served by the remote, object bytes, paths/modes, descriptor, source/vendor/embed inputs, sizes, diagnostics, and compiled artifact;
- source repository configuration and object-database indirection until rejected;
- candidate snapshot/receipt/marker/artifact cache bytes until protected admission;
- network failure, replay, moved tag, malicious remote, and compiler resource-exhaustion input.

### 13.2 Trusted computing base

- manager/protocol implementation;
- operator-selected Git distribution, HTTPS/SSH transport, credential and host-verification broker;
- operating system, filesystem containment, locks, protected caches, and transaction journal;
- operator-trusted Go toolchain and fingerprint;
- canonical JSON and hash implementations;
- any future separately standardized signer and opaque signing identity.

### 13.3 Required controls

| Threat | Control |
|---|---|
| Mutable/symbolic revision | Full object-format commit lock; optional safe tag must peel exactly |
| Declared/effective substitution confusion | Both states and substitution type in receipt-v2 input/cache key |
| Replace refs or grafts | Reject refs/grafts and use `--no-replace-objects` |
| Partial clone or lazy fetch | Reject promisor state and use `--no-lazy-fetch`; prove all tree/blob objects local |
| Alternates/object-store escape | Reject alternate files/env and manager-copy into private object DB |
| Fetch default widens state/process graph | Exact fresh-repo init/fetch argv; fixed destination ref; no tags, submodules, `FETCH_HEAD`, maintenance, commit graph, refmap, filter, or server option |
| Ambient config/URL rewrite/proxy/helper | Empty global/system config, manager-owned repo, fixed config/env/process graph |
| SSH MITM, system config, or variant argv | `GIT_SSH_VARIANT=ssh`; exact wrapper shape; `-F` empty config; operator-known hosts; fixed OpenSSH argv |
| Local repository extension/ref ambiguity | Admit only ordinary files-ref layout; byte-level config/ref parser; reject gitfile, commondir, bare, reftable, or unknown extensions |
| Malicious/incomplete pack metadata | Pack v2/v3 plus index-v2 admission, hash-family filenames/trailers, inert private copying, fixed trusted `cat-file --batch`, and manager object-ID recomputation |
| Object-reader transformation or process escape | One full-OID-only batch grammar; no `ls-tree`, revision/path input, filter, textconv, mailmap, alternate, lazy fetch, or object-reader child |
| Hook/filter/LFS/submodule execution | Raw object extraction; no checkout/archive transforms; exact all-blob LFS-pointer rejection; reject gitlinks |
| Package output/argv/signing selection | Strict descriptor; command-key output; closed driver; no signing fields |
| Audit bypass on cache hit | Exact snapshot validation/hash/audit before artifact-cache lookup |
| Forged cache | Protected boundary plus exact digests; rebuild rather than adopt |
| Shim/PATH hijack | Structural no-follow/hash validation; manager-generated shims; never execute output |
| Partial/cross-project install | One manager-home journaled transaction; consumer last; reverse rollback |

Same-principal/admin compromise remains outside the local protected-cache invariant. Receipt and marker hashes are consistency/currentness identifiers, not signatures.

## 14. Exact impact on existing board work

This scope is a new schema-7/rc.5 delivery subtree. It MUST NOT reopen, relabel, or silently expand rc.4 work.

### 14.1 Existing curator-spec schema-6 story

`STORY-260720-35dck7` (“Protocol schema v6”) and all 12 child tasks remain rc.4 evidence. In particular:

- `TASK-260720-1nvomm`, `TASK-260720-17llva`, and `TASK-260720-wajgn8` remain the schema-6 core/profile/manifest contract;
- `TASK-260720-12iigs`, `TASK-260720-2zc6k1`, and `TASK-260720-37ei85` remain receipt-v1/marker-v2, claim-v2, and legacy guards;
- `TASK-260720-1s1vr6`, `TASK-260720-cw39jh`, `TASK-260720-1u7hes`, `TASK-260720-3lo9jc`, `TASK-260720-q5oy3o`, and `TASK-260720-3ag6pi` remain rc.4 vectors, gates, docs, release metadata, and integrated verification.

No existing schema-6 task should absorb schema 7. After architecture acceptance, the smallest new curator-spec set is:

1. core/security/decision contract;
2. wire schemas and generated schema cases;
3. manager profile/CLI including Git, substitutions, status/repair/GC;
4. conformance/release gates including raw-object, mixed-build, and claim-v3 vectors.

Task-local tests/docs stay with those deliverables; no ceremonial extra task is justified.

### 14.2 Existing Curator implementation story

`STORY-260720-3plyvy` (“Curator Go build driver”) remains the local schema-6 `go-v1` implementation. Its manifest, source identity, toolchain, receipt/cache, marker, shim, transaction, lifecycle, docs, and rc.4 verification tasks are prerequisites/reusable foundations, not owners of external Git semantics.

A later external-repository Curator story consumes those foundations and adds:

- schema-7/repository/descriptor/substitution models;
- clean Git acquisition and raw-object snapshot admission;
- independent external audit subjects;
- receipt v2, marker v3, mixed-command planning, snapshot-cache currentness/GC;
- end-to-end rc.5 vectors.

Do not change `TASK-260720-1zntv0` or other `go-v1` tasks to accept Git fields or arbitrary build behavior.

### 14.3 Existing cocoaskills implementation story

`STORY-260720-1uv5gi` (“csk Go build driver”) likewise remains the independent local schema-6 implementation. A new external-repository story should build on its schema model, transaction engine, source/toolchain identities, protected caches, planner, currentness/repair/GC, docs, and E2E gates.

The new work owns Python-specific clean Git/raw-object handling, external audit subjects, receipt-v2/marker-v3 mixed builds, substitutions, and rc.5 vectors. Existing `git archive` behavior is not sufficient unless independently proven byte-identical and free of attributes/filters; the recommended implementation is raw object extraction.

### 14.4 Existing interoperability story

`STORY-260720-21bsr2` (“Compiled-build interoperability”) remains rc.4/schema-6 evidence. Its 12 tasks, including shared compiled cases, native consumers, black-box runner, authoring guide, language-driver gate, release qualification, and pins, MUST NOT claim schema 7.

After both managers implement schema 7, a downstream rc.5 interoperability extension adds exact cross-language comparisons for:

- canonical URL/ref parsing and SHA-1/SHA-256 locks;
- raw snapshot bytes/digest and declared/effective identities;
- unsubstituted/substituted/mixed receipt and marker bytes;
- unavailable, moved, alternate/replace/graft/promisor, helper/filter/LFS/submodule negatives;
- cache hit/miss/corruption, status/repair/GC, shim/PATH, rollback, and release evidence.

Linux remains a separate later native-validation item and cannot be listed in claim v3 until it passes.

## 15. Rejected alternatives

- Add Git/ref/output fields to a build command: duplicates source declarations and reopens output control.
- Let the descriptor name a binary/output: violates manager-derived output and collision/path safety.
- Reuse `go-v1`: changes source acquisition, identity, receipt, and marker meaning under an rc.4 driver.
- Tag-only, branch, abbreviated ID, or raw revision grammar: mutable or ambiguous compiler input.
- Trust a source or artifact cache without exact protected admission and audit: bypasses source policy.
- Permit replacements, alternates, grafts, partial clones, lazy fetch, local upload-pack, submodules, or LFS: adds unbound objects, processes, credentials, or network.
- Use `git clone`, an implicit/default fetch, or a named remote: admits template/config/refspec/tag/submodule/maintenance behavior not selected by the manager.
- Inherit Git/SSH/proxy/helper/filter configuration: transfers process/network control outside the closed manager policy.
- Auto-detect the SSH variant or pass the source SSH configuration: permits variant-dependent argv and system/user proxy, command, identity, forwarding, or control behavior.
- Admit gitfiles, linked worktrees, bare repositories, reftable, unknown repository extensions, optional pack sidecars, pack versions outside 2/3, or pack indexes outside v2 in local-substitution v1: requires additional containment/parser semantics and vectors; reject with stable errors instead.
- Let each manager choose its own documented pack/index subset: makes the same local substitution non-interoperable; v1 fixes pack 2/3 plus index 2.
- Use `ls-tree`, `--batch-command`, pretty-printers, revision/path expressions, checkout, archive, filters, textconv, or mailmap for extraction: adds redundant grammars or repository-selected transformations; one full-OID `cat-file --batch` stream is sufficient.
- Treat recognized Git LFS pointer text as compiler source: a package can make a pointer visible through `go:embed` while the actual bytes remain outside the locked Git graph; v1 rejects canonical and narrowly recognized legacy/noncanonical pointers before audit.
- Store only effective substitution state: loses declared provenance; store only declared state: misdescribes compiled bytes.
- Upgrade every schema-7 command to receipt v2: unnecessarily changes local `go-v1`; mixed explicit versions preserve rc.4 semantics.
- Install-time signing/notarization: introduces secrets, post-processing children, mutable services, and new identity not covered by the first driver.
- Standardize physical cache/checkouts: machine-local layout remains implementation-specific.

## 16. Primary-source register

Checked on 2026-07-27:

- Git global options and environment isolation, including `--no-replace-objects`, `--no-lazy-fetch`, `GIT_CONFIG_GLOBAL`, `GIT_CONFIG_SYSTEM`, `GIT_CONFIG_NOSYSTEM`, `GIT_SSH`, and prompt controls: https://git-scm.com/docs/git
- Exact fetch defaults and controls, including tag auto-follow, `FETCH_HEAD`, auto-maintenance, commit-graph writing, refmap, submodule recursion, upload-pack, filters, and full-object refspecs: https://git-scm.com/docs/git-fetch
- Private repository initialization, object format, files/reftable ref format, and template selection: https://git-scm.com/docs/git-init
- Repository format version and mandatory extension recognition/rejection: https://git-scm.com/docs/repository-version
- Reftable repository extension, stack, and dummy files that make implicit files-ref parsing unsafe: https://git-scm.com/docs/reftable
- Repository layout, gitfiles, commondir/linked worktrees, loose/packed refs, shallow, graft, alternate, and object-pack state: https://git-scm.com/docs/gitrepository-layout
- Pack and index structure/checksum surfaces, including pack versions 2/3,
  index version 2, SHA-1/SHA-256 widths, fanout/offset tables, and trailers:
  https://git-scm.com/docs/gitformat-pack
- Git SSH variant selection and documented `ssh` wrapper argv: https://git-scm.com/docs/git-config
- OpenSSH `-F` configuration isolation and command-line behavior: https://man.openbsd.org/ssh.1
- OpenSSH system configuration, proxy/jump/local-command, known-hosts, identity/agent, forwarding, TTY, and control options: https://man.openbsd.org/ssh_config.5
- Replacement refs are honored by default and disabled by the fixed option/environment: https://git-scm.com/docs/git-replace
- Partial clones, promisor pack markers, promisor remotes, and dynamic missing-object fetch: https://git-scm.com/docs/partial-clone
- Alternate object stores, replace refs, grafts, shallow/commondir repository state: https://git-scm.com/docs/gitrepository-layout
- Exact ref-name exclusions used by the tag/branch grammar: https://git-scm.com/docs/git-check-ref-format
- URL rewrites, protocol policy, remote proxy/upload-pack/VCS/promisor/server-option configuration: https://git-scm.com/docs/git-config
- Credential helpers are transformed into shell-executed commands: https://git-scm.com/docs/gitcredentials
- Raw object batch request/response framing, full object-name/type/size atoms,
  and the separate filter/textconv/mailmap/follow-symlink options:
  https://git-scm.com/docs/git-cat-file
- Revision peeling semantics: https://git-scm.com/docs/gitrevisions
- Git attributes and checkout filters: https://git-scm.com/docs/gitattributes
- Git LFS pointer/content separation and normative pointer byte rules:
  https://github.com/git-lfs/git-lfs/blob/main/docs/spec.md
- Git LFS reference pointer aliases, SHA-256 OID parser, extension key/value
  handling, and canonical encoder:
  https://github.com/git-lfs/git-lfs/blob/main/lfs/pointer.go
- Go vendoring and fixed build behavior: https://go.dev/ref/mod#vendoring and https://go.dev/cmd/go/#hdr-Compile_packages_and_dependencies
- Go toolchain selection: https://go.dev/doc/toolchain
- Apple notarization/signing workflow: https://developer.apple.com/documentation/security/notarizing-macos-software-before-distribution
- Microsoft SignTool certificate/timestamp surfaces: https://learn.microsoft.com/en-us/windows/win32/seccrypto/signtool
- Cargo build scripts: https://doc.rust-lang.org/cargo/reference/build-scripts.html
- javac annotation processing: https://docs.oracle.com/en/java/javase/23/docs/specs/man/javac.html
- Roslyn source generators: https://learn.microsoft.com/en-us/dotnet/csharp/roslyn-sdk/#source-generators

## 17. Validation and independent-review checks

Required packet validation:

- every fenced JSON example parses, including the exact init/fetch/SSH and
  `cat-file --batch` arrays;
- unsubstituted source omits `substitution`;
- substituted source requires both `substituted: true` and a typed substitution;
- mixed marker v3 carries receipt v1 for local `go-v1`, receipt v2 for external commands, and top-level skill build-source only because a local build is active;
- all revision-1 findings remain mapped to sections 3, 5, 6, 8, and 9;
- all revision-2 findings map to exact network argv (6.2–6.3), SSH
  isolation (6.4), and local format/ref/object admission (6.5–6.6);
- all revision-3 findings map to the single raw-object process and byte
  protocol (6.1, 6.6.1), pack 2/3 plus index-2 local admission (6.5.3), and the
  all-blob LFS detector (6.6.3);
- no project/spec/code/schema/test/release/prior-resource file changes.

Independent review should decide:

1. Is the declared/effective receipt shape sufficient to prevent substitution/cache confusion?
2. Do the exact init/fetch vectors prevent remote/package/default selection of
   local refs, helper, refspec, tags, submodules, maintenance, `FETCH_HEAD`,
   filter, server option, or upload-pack?
3. Does the fixed SSH wrapper admit only the tested protocol-v0 destination and
   remote command, and does the OpenSSH vector suppress user/system config,
   prompts, TTY, forwarding, proxy/jump/local-command/control behavior while
   retaining only the required private Git protocol pipe?
4. Does local v1 fail stably for gitfile/linked/bare/reftable/unknown layouts
   and unambiguously parse ordinary config, files refs, loose objects, and
   exact pack/index pairs without running source Git behavior?
5. Does the single full-OID `cat-file --batch` contract close request,
   response, transformation, child-process, network, missing-object, and
   resource-limit behavior for both network and inert-copied local stores?
6. Are pack versions 2/3 and index version 2 defined identically for SHA-1 and
   SHA-256 managers, with all other physical formats rejected stably?
7. Does the all-reachable-blob LFS detector reject canonical and exactly
   recognized legacy/noncanonical pointer bytes without relying on paths,
   `.gitattributes`, filters, or hydration?
8. Are URL, SSH trust, tag, and structured development-ref rules exact enough for cross-language implementations?
9. Are whole-snapshot audit and pre-cache ordering coherent on real, cache-hit, and dry-run paths?
10. Can marker v3 represent local-only, external-only, and mixed installations without changing receipt-v1/schema-6 meaning?
11. Do status, repair, deduplication, rollback, and GC fail closed without executing source or built output?
12. Does logical target selection satisfy repository target selection while preserving manager-derived binary names and paths?
13. Are signing and future-driver boundaries still closed?

### 17.1 Producer verification evidence

Checks performed on 2026-07-27:

- all ten fenced JSON documents parsed with the host Python JSON parser;
- targeted topic checks found the exact init/fetch vectors,
  `GIT_SSH_VARIANT=ssh`, wrapper argv contract, `-F` SSH isolation, stable
  gitfile/bare/linked/reftable errors, config/ref grammar, exact pack-2/3 and
  index-2 admission, the full-OID-only batch request/response grammar, manager
  object-ID recomputation, the all-blob LFS detector, audit-before-cache/compiler
  ordering, mixed receipt/marker behavior, status/repair/GC, signing,
  future-driver rules, and board-impact mapping;
- the new Git fetch/init/repository-version/reftable/layout/config and OpenSSH
  manual pages in section 16 were opened from their primary maintainers and
  checked against the claims above;
- a temporary local-transport smoke on Git 2.50.1 exercised the exact
  init/common-fetch option shape for both full-object and exact-tag refspecs,
  verified the fixed `refs/curator/*` destinations, and verified that
  `FETCH_HEAD` was absent;
- a task-local Git 2.50.1 raw-object smoke used the exact `cat-file --batch`
  argv against SHA-1 and SHA-256 private repositories; full-ID requests
  returned the documented `<oid> <type> <size> LF`, exact content bytes, and
  terminal LF, while no filter/textconv/mailmap option or child process was
  used;
- independent fixture checks classified a canonical current Git LFS pointer,
  CRLF and no-terminal-LF legacy/noncanonical pointers, and pointer-extension
  examples as rejected while leaving unrelated small text and zero-byte blobs
  admitted;
- unchanged curator-spec validation passed:
  `validated 30 schemas and 93 vector files`;
- `git diff --check` passed in curator-spec;
- this run did not edit product/spec/schema/code/test/release files or either
  prior producer/reviewer resource.
