# TASK-260827-21xw9d rev2 rework: one blocking defect (D5)

Reviewer run `RUN-260828-a70495`, 2026-08-28, against
`CR-TASK-260827-21xw9d-2` revision 2 (candidate `2509c7d3`).

Rev1 defects D1, D3 and D4 are fixed and were independently reverified. D2 is
fixed for the shim contract, the script runtime target and the local build
cache path. One invented filesystem path survives.

## D5 (blocking): `cache/build_repositories` is not a real path

`docs/authoring-cli-commands.md:176` currently reads:

> `curator install` creates an executable launcher in `.agents/bin/remote_cmd`
> pointing to the compiled binary in Curator's build repository artifact store
> (`$HOME/.curator/cache/build_repositories/.../artifacts/<hash>/artifact`).

That path exists nowhere but this document:

    grep -rn "cache/build_repositories" . --exclude-dir=.git
    docs/authoring-cli-commands.md:176:...

`build_repositories` is only an `agent-skill.json` manifest key
(`internal/skillspec/parse.go:135,172,417,421,429,491`), never a path segment.

The real external store root is a sibling of `cache`, not a child of it:

    internal/install/external.go:105:
        deps.StoreRoot = filepath.Join(home, "external-build-cache")

with `home = cfg.Home()` (`internal/install/install.go:541`) = `$HOME/.curator`.
The artifact path is assembled at `internal/install/targets.go:185-186`:

    keyName := strings.TrimPrefix(entry.result.CacheKey, "sha256:")
    finalArtifact := filepath.Join(externalRoot, "artifacts", keyName, "artifact")

and garbage collection sweeps the same root
(`internal/scopes/gc.go:125`).

### Fix

1. Line 176: replace the path with

       $HOME/.curator/external-build-cache/artifacts/<sha256-hex>/artifact

2. Line 68: the trailing "/ build artifact store for compiled binaries" is
   unnamed. Name the same store there so the two mentions agree.

Verify with `grep -rn "external-build-cache" --include='*.go' internal/` before
resubmitting, and confirm `grep -rn "build_repositories" docs/authoring-cli-commands.md`
only matches the manifest key in Example 2, not a path.

## Do not change anything else

The rest of the document was reverified in full this round and holds: all 14
diagnostic identifiers, the build vectors, the network-off environment, the
native-input and directive lists, the Linux refusal, the interpreter set and
its rationale, the mutual `execution_policy`/`interpreter` errors, the
`script-worker-v1` refusal, all three worked examples (`curator skill check`
exit 0 each), all six reserved driver identities, the shim contract, the script
runtime target, the local build cache path, the `.git` restatement, the README
link, and prose style. Details and literal outputs in
`TASK-260827-21xw9d_review-verdict-rev2.md`.
