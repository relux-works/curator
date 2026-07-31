# External build repositories — independent review verdict

Verdict: **changes requested**. Route to `analysis` for architecture rework and a new reviewer cycle.

Reviewed artifact: `TASK-260720-1nvomm_external-build-repositories-architecture.md`, SHA-256 `969de48ae9230e6696c2c5423c3d71d966ba66626c6cbd2e5d77d9a78cb6352c`. The run is not goal-bound.

## Accepted portions

- Schema 7 plus new `go-repository-v1`, receipt v2, marker v3, and claim-v3/rc.5 preserve schema 1 through 6 and frozen `go-v1` semantics.
- Logical descriptor targets plus manifest command keys correctly reject repository-controlled executable names and output paths. Manager-private staging, `bin/<command>[.exe]`, and manager-owned shims remain authoritative.
- Whole external-snapshot validation, identity, and independent audit before compiled-cache lookup or compiler execution are conservative and operationally coherent. Cache hits do not bypass source audit.
- Raw tree/blob extraction, rejection of gitlinks, links, special modes, LFS hydration, filters, hooks, and checkout transforms is the correct direction.
- Structural PATH/shim validation, locked rollback, no built-output execution, no install-time signing, release-pipeline notarization/Authenticode, and per-language closed-driver review all preserve decision 0004.
- All four fenced JSON examples parse. The task-local validator passed: `validated 30 schemas and 93 vector files`. `git diff --check` passed for the original owned schema-6 documents. Primary source claims about Git protocol policy, helper execution, filters, raw object access, object formats, revision peeling, submodules/LFS, Go vendoring/toolchain behavior, platform signing, Cargo build scripts, javac processing, and Roslyn generators were checked against the cited vendor documentation.

## Required changes

### 1. Receipt/cache identity does not bind effective substitutions or mixed schema-7 builds

Sections 5.2-5.3 say receipt v2 contains the shown input, but that input contains only the declared `canonical_git` and `locked_commit`. Section 10 permits a local or development-ref substitution and says only marker v3 records effective identity/commit/substitution. This can make the receipt/cache key describe declared provenance while compiling substituted bytes.

Define receipt-v2 source identity with both declared and effective source state: canonical identity kind/value, object format, full effective commit, substitution flag/type, and external build-source digest. Define which fields are absent for an unsubstituted source. The cache key MUST bind that exact effective state. Also define mixed schema-7 behavior: local `go-v1` keeps receipt v1 and the skill build-source identity; external commands use receipt v2; marker v3 can represent both command kinds simultaneously and preserves the schema-6 top-level build-source semantics whenever local builds are active.

### 2. The raw Git object boundary is not closed against Git-native indirection

Sections 4.2-4.3 require raw plumbing but do not prohibit replacement refs, lazy promisor fetch, object alternates, or inherited global/system repository behavior. Git replacement refs are honored by default, and partial-clone object reads can dynamically fetch missing objects. Thus `ls-tree`/`cat-file` is not by itself proof that bytes came directly and completely from the locked object.

Require a clean manager-owned Git environment/config and fixed child process graph. Raw verification/extraction MUST disable replacement objects and lazy fetch, reject/clear alternate object directories and promisor/partial-clone state, avoid grafts, and prevent ambient URL rewrites, proxy commands, SSH commands, server options, hooks, filters, helpers, and repository config except the explicitly selected operator credential/host-verification broker. Require complete local availability of every reachable tree/blob before audit. Relevant primary sources: https://git-scm.com/docs/git-replace, https://git-scm.com/docs/git, https://git-scm.com/docs/partial-clone, https://git-scm.com/docs/git-cat-file.

### 3. URL, canonical identity, SSH trust, and tag grammar are not exact enough for interoperability

The manifest accepts a `git` string, the receipt invents `canonical_git`, and transport text says validated SCP-form SSH without defining the binding. The optional `tag` is appended to Git revision syntax without an explicit protocol grammar. Git ref syntax excludes revision metacharacters such as `..`, `^`, `:`, and `@{`; this must be validated before constructing `refs/tags/<tag>^{commit}`.

Normatively reuse the existing protocol canonical-source-identity algorithm, narrowed to HTTPS and SSH, and state exactly how HTTPS userinfo, SSH usernames, ports, case, `.git`, SCP form, and host-key verification are handled. Bind declared and effective identities as required by finding 1. Validate the complete `refs/tags/<tag>` with a protocol-defined equivalent of `git check-ref-format`; never accept raw revision grammar. Define fixed SSH host verification as operator-owned policy and forbid package influence over known-hosts and SSH options. Relevant primary sources: https://git-scm.com/docs/git-check-ref-format, https://git-scm.com/docs/gitrevisions, https://git-scm.com/docs/git-config, https://git-scm.com/docs/gitcredentials.

### 4. Status/currentness behavior is missing from the access-state model

Section 4.1 defines syntax-only check, mutating/audit operations, and dry-run, but omits status. Section 5.3 says status verifies both snapshots without saying what happens when the exact external snapshot is unavailable or its protected boundary fails.

Specify status as read-only: revalidate the protected snapshot/cache boundary and exact effective identity without fetching or repairing; report non-current or the existing unknown state when the required snapshot cannot be proved; treat untrusted/corrupt snapshot state as non-current; and make repair rebuild/refetch from the declared or effective operator-owned source before mutation. Define marker-v3 and journal references as snapshot-cache and artifact-cache GC roots for both local and external commands.

## Sources and validation

- Git protocol and rewriting: https://git-scm.com/docs/git-config
- Git credentials and askpass: https://git-scm.com/docs/gitcredentials
- Git object-format reporting: https://git-scm.com/docs/git-rev-parse
- Git tag/ref grammar and peeling: https://git-scm.com/docs/git-check-ref-format and https://git-scm.com/docs/gitrevisions
- Raw tree/blob interfaces: https://git-scm.com/docs/git-ls-tree and https://git-scm.com/docs/git-cat-file
- Checkout transforms: https://git-scm.com/docs/gitattributes
- Signing sources cited by the producer remain consistent with the separation decision: https://developer.apple.com/documentation/security/notarizing-macos-software-before-distribution and https://learn.microsoft.com/en-us/windows/win32/seccrypto/signtool

No product, spec, schema, code, test, release, or producer-resource file was modified by this review.