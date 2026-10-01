# External Build Repositories

A schema 7 skill can build a command from a separate Git repository. The skill names the repository and pins one exact commit, and Curator verifies that commit before it compiles anything. This guide covers the author-facing surface. The full normative rules live in the specification's [External build repositories](https://github.com/relux-works/curator-spec/blob/main/docs/external-build-repositories.md) guide.

## Declare a repository

To build from an external repository, declare it under `build_repositories` and point a `build` command at it. `build_repositories` requires `schema_version: 7`.

This `agent-skill.json` declares one repository and one command built from it:

```json
{
  "schema_version": 7,
  "capabilities": {},
  "build_repositories": {
    "tools": {
      "git": "https://git.example.com/org/tools.git",
      "locked_commit": {"object_format": "sha1", "hex": "0123456789abcdef0123456789abcdef01234567"},
      "tag": "v1.4.0"
    }
  },
  "commands": {
    "tool": {"type": "build", "driver": "go-repository-v1", "repository": "tools", "target": "tool"}
  }
}
```

The command names the repository and a target. The repository root carries `skill-build.json`, which maps each target to a Go module and command package:

```json
{"schema_version": 1, "targets": {"tool": {"driver": "go-repository-v1", "build_root": "tools", "source_dir": "tools/cmd/tool"}}}
```

## Locked commit and tag

The `locked_commit` is the identity of the build. Curator fetches that exact object and refuses any other commit, so moving a branch or a tag never changes what installs. `object_format` is `sha1` or `sha256` and must match the repository.

The `tag` is optional. When you declare it, Curator also asserts in the same operation that the tag still names the locked commit. A moved or deleted tag fails the install instead of silently building a different tree. To release a new version, update `locked_commit` and `tag` together.

The `git` URL uses `https` or `ssh`. Private repositories need operator credentials; see [SSH credentials](build-ssh.md) and [HTTPS credentials](build-https.md). A package can never select credentials itself.

## Develop against a local checkout

To build from a working copy while you develop, add a substitution to `Skillfile.dev.json` next to the project's `Skillfile.json`. The file is operator-owned and must stay out of version control.

This substitution builds the `tools` repository of `my-skill` from a local checkout:

```json
{
  "schema_version": 2,
  "substitutions": {},
  "build_repository_substitutions": {
    "my-skill": {"tools": {"path": "../tools"}}
  }
}
```

`curator install` then prints `BUILD REPOSITORY SUBSTITUTION my-skill.tools` and builds the commit at the checkout's `HEAD`. Curator reads the committed objects as inert bytes and never runs Git inside the checkout. Commit a change before you install it; uncommitted edits in the working tree are not built.

The local checkout needs a narrow layout. It must be a non-bare worktree with a direct `.git` directory, not a gitfile or a linked worktree. Inside `.git`, Curator admits only `HEAD`, `config`, `index`, `objects`, `refs`, and `packed-refs`. `git init` and `git clone` also create `hooks/`, `info/`, and `logs/` by default, and Curator refuses such a checkout with `build_repository_source_unavailable`.

A substitution can also name a network `git` URL with a `ref` of kind `tag`, `revision`, or `branch`. Substituted builds are recorded as substituted in the install marker. Strict audit refuses every install while `Skillfile.dev.json` declares a substitution.

## What Curator verifies

Every install runs the same pipeline, in order. A failure at any stage stops the install before a shim is published.

1. **Admission.** Curator fetches the exact commit into private state and proves the object graph, the declared tag, and the source files. LFS pointers and unsupported repository layouts are refused.
2. **Audit.** An independent audit of the admitted snapshot runs before any code executes.
3. **Cache.** Curator derives a cache key from the source, the toolchain, and the execution policy. A matching verified artifact is reused, and the install reports `outcome=cache-hit` without rebuilding.
4. **Compiler.** On a cache miss, the Go toolchain builds the target from the immutable snapshot under the manager-owned execution policy. Curator records a receipt next to the artifact and publishes the shim in `.agents/bin/`.

Compilation runs on macOS and Windows. On Linux and other hosts, Curator refuses the build before the worker starts with `build_execution_control_unavailable`; see Platform support in [Authoring CLI commands](authoring-cli-commands.md).

Removing the skill and reinstalling removes its shim and the build root it no longer references. For status codes and repair, see [Compiled commands](compiled-commands.md) and [Troubleshooting](troubleshooting.md).
