# External build repositories v2 — independent review verdict

Verdict: **changes requested**. Route `TASK-260720-1nvomm` to `analysis` for architecture rework and a new reviewer cycle.

Reviewed artifact: `TASK-260720-1nvomm_external-build-repositories-architecture-v2.md`, SHA-256 `76dbea4bc3f92ec9f51d59774ff81c30d402b7f0bb880aa5ab18eda507629202`. The active run `RUN-260727-dd78a0` is not goal-bound.

## Accepted portions

- Schema 7, `go-repository-v1`, receipt v2, marker v3, and claim-v3/rc.5 preserve schemas 1 through 6 and the frozen `go-v1` meaning.
- Logical descriptor targets plus manifest command keys satisfy repository target selection without reopening executable-name or output-path selection.
- Receipt-v2/cache identity now binds declared and effective source, object format, full commit, substitution state/type, and external snapshot digest. Mixed local receipt-v1 and external receipt-v2 marker-v3 behavior is coherent.
- Whole external-snapshot validation, hashing, and independent audit before artifact-cache lookup or compiler execution correctly prevent cache-hit audit bypass.
- Replacement, graft, alternate, promisor/lazy-fetch, gitlink, LFS, link/special-file, and compiler-input boundaries are directionally correct.
- Read-only status, repair, protected-cache admission, snapshot/artifact GC roots, structural PATH/shim checks, rollback, signing/notarization separation, and future closed-driver review preserve the accepted security model.
- All six fenced JSON blocks parse. `PYTHONDONTWRITEBYTECODE=1 .temp/TASK-260720-1nvomm/venv/bin/python3 tools/validate.py` passed with `validated 30 schemas and 93 vector files`; `git diff --check` passed for the owned schema-6 documents.

## Required changes

### 1. Network acquisition is called “fixed” but no fixed init/fetch contract exists

Section 6.3 says the manager uses a fixed `git fetch` form, yet gives no exact argv, destination ref, or separate locked-ID/tag retrieval flow. This matters because Git defaults conflict with the packet’s own closed process/state claims:

- fetch auto-follows tags unless `--no-tags` is supplied;
- fetch writes `FETCH_HEAD` by default;
- fetch runs automatic maintenance by default, adding a local child omitted from section 6.1’s permitted process graph;
- submodule recursion is configurable and otherwise defaults to `on-demand`;
- `--upload-pack`, command-line refspecs, and `--refmap=` control remote program and ref selection.

Define the exact manager-owned repository creation and fetch argv for SHA-1 and SHA-256, including the absolute trusted Git path, explicit `--git-dir`, `--no-replace-objects`, `--no-lazy-fetch`, object format, `--no-tags`, `--no-recurse-submodules`, `--no-auto-maintenance`, refmap/ref destination behavior, fixed `git-upload-pack`, and whether `FETCH_HEAD`/commit-graph writes are disabled. Define distinct safe flows for fetching a full locked object ID and for fetching/verifying an optional exact tag when the server refuses unadvertised IDs. The remote must never choose a local ref name, refspec, server option, filter, helper, or upload-pack.

Primary evidence: [git-fetch](https://git-scm.com/docs/git-fetch) documents tag auto-follow, `FETCH_HEAD`, default auto-maintenance, submodule recursion, refmaps, and `--upload-pack`; [git global options](https://git-scm.com/docs/git) documents the explicit repository and no-replace/no-lazy-fetch controls.

### 2. SSH still admits ambient configuration and variant-dependent argv

Sections 6.1–6.2 use `GIT_SSH` but neither fix `GIT_SSH_VARIANT=ssh` nor define the wrapper’s accepted input argv and exact trusted-SSH argv. Git documents that the variant otherwise comes from autodetection. More importantly, a private `HOME` blocks user SSH config but not the system SSH config. OpenSSH reads system configuration by default; it can select `ProxyCommand`, `ProxyJump`, `Match exec`, `KnownHostsCommand`, `LocalCommand`, identities, control sockets, and forwarding behavior. That contradicts the packet’s claim that proxies, commands, options, and process selection come only from manager policy.

Require a fixed SSH variant and a binary wrapper that rejects every invocation shape except the expected Git probe/connection forms. Invoke the operator-trusted SSH binary with a manager-owned empty configuration (or `-F none` where supported), manager-owned known-hosts input, strict host-key checking, batch/no-prompt mode, no TTY, null stdin, disabled forwarding, proxy/jump/local-command/control-master behavior, and only the operator-owned identity/agent channel explicitly selected by manager policy. State how the wrapper validates and forwards the package-derived host/user plus the fixed `git-upload-pack` remote command.

Primary evidence: [Git environment documentation](https://git-scm.com/docs/git) describes `GIT_SSH` and `GIT_SSH_VARIANT`; [OpenSSH ssh(1)](https://man.openbsd.org/ssh.1) states that `-F` selects configuration and suppresses the system file, while [ssh_config(5)](https://man.openbsd.org/ssh_config.5) documents executable proxy/local/known-host commands and the system configuration file.

### 3. Local substitution admission does not define the repository format/ref backend it parses

Section 6.3 says source repository config is not parsed for behavior or copied, but also requires rejection of `extensions.partialClone`, promisor remotes, and object-format mismatch. It does not define a strict data-only config parse, supported repository-format versions/extensions, include handling, or the reference backend. The manual parser handles only `HEAD`, loose refs, and `packed-refs`; a reftable repository uses `extensions.refStorage=reftable` and deliberately contains dummy `HEAD`/refs files. Unknown extensions, `worktreeConfig`, SHA-256 `objectFormat`, shallow state, and config includes therefore need an explicit admission rule rather than an implicit failure.

Define a byte-level local repository admission algorithm: permitted repository format and extension keys; data-only parsing with includes rejected; files-ref backend only or a normative reftable parser; exact `.git` gitfile/`commondir`/linked-worktree containment; source `HEAD`/loose-ref/`packed-refs` grammar; common `shallow`, graft, replace, alternate, promisor, and partial-clone checks; the precise loose/pack/sidecar inventory copied into private state; and recomputation of the selected commit object ID as well as every reachable tree/blob ID. Unsupported ref/object formats must fail with a stable pre-audit error.

Primary evidence: [Git repository format](https://git-scm.com/docs/repository-version) requires readers of format version 1 to understand or reject every extension; [reftable](https://git-scm.com/docs/reftable) defines `extensions.refStorage=reftable` and its dummy files; [repository layout](https://git-scm.com/docs/gitrepository-layout) defines `commondir`, shallow, alternates, grafts, and linked-worktree state.

## Verdict boundary

These are ordinary architecture rework items, not a stop-the-line decision. No product/spec/code/schema/test/release or producer resource was modified. After the packet specifies the exact network argv/SSH wrapper/local-repository admission contracts and reruns its JSON/source checks, it should receive a fresh independent review.
