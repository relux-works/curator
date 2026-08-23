# Reviewer verdict for TASK-260810-2n3sbi

Verdict: **accepted -> done**

## Goal and scope evidence

- Reviewer run: `RUN-260811-76cafd`
- Authoritative goal at the verdict checkpoint: `GOAL-260811-6e447e` revision 1
- Resolved scope: `TASK-260810-2n3sbi`
- Review policy: `required`
- The orchestrator progress and convergence directives were observed at a safe checkpoint. No cancel, pause, reroute, or scope-change directive exists.

The goal requires exactly one evidence-backed reviewer branch. This artifact records the accepted branch; it records neither changes requested nor stop the line.

## Acceptance result

No acceptance-blocking finding remains.

1. **Closure-model comparison.** The outcome separately evaluates npm, pnpm, Yarn Classic, modern Yarn, pylock.toml, hash-complete pip requirements, and versioned tool-specific Python locks. It distinguishes graph selection, immutable raw artifacts, target/peer/marker contexts, derived caches, installed layouts, and runtime identity.
2. **Hooks, build backends, and generated code.** It disables dependency lifecycle execution during materialization; detects npm implicit binding.gyp/node-gyp, pnpm pnpmfile and side-effects surfaces, Yarn plugins and Git pack/build hooks; and permits only declared networkless build or generator DAG nodes. Shipped generated JavaScript is distinguished from locally generated JavaScript with causal lineage. Python install closure is separated from PEP 517 backend/build closure, including predeclared locked dynamic requirements and contained backend-path handling.
3. **Compiled payload and wheel policy.** Recursive raw-container inspection applies the accepted shared classifier before installation. Native Node addons, Python extensions, native wheels, objects/libraries, WebAssembly, bytecode, code caches, opaque members, and renamed or nested compiled leaves reject. Wheel tags, Root-Is-Purelib, package labels, and checksums cannot override byte classification; WHEEL, METADATA, and RECORD are verified and reconciled.
4. **Shared versus separate layers.** The recommendation shares the artifact service, trust roles, canonical graph/checkpoint schemas, sandbox/output contract, diagnostics, and semantic fixtures while keeping npm/pnpm/Yarn adapters and the independent Python resolver/backend implementation separate. Compatibility is protocol- and fixture-based and never assumes repository, cache, environment, or code co-location.
5. **Recursive immutable closure and offline proof.** The explicit admission predicate, six-step compositional proof, chained C0-C7 checkpoints, and clean replay protocol require every transitive/build edge and raw byte to be immutable and audited. Replay uses an empty ambient home/cache, derived private stores, frozen/no-resolution modes, OS-enforced network denial, write/process checks, poisoned-cache invariance, and missing-artifact failure before build.
6. **Diagnostics, unsupported cases, and fixtures.** Stable global closure codes carry structured ecosystem details and deterministic precedence. Unsupported lists fail closed for mutable locators, lock/parser gaps, undeclared manager extensions, native builds, compiled or opaque payloads, record/metadata drift, and unreproducible target selection. Shared vectors S01-S08, Node vectors N01-N13, and Python vectors P01-P13 cover positive closure, offline replay, recursive failure, hooks, generation, native/compiled rejection, caches, targets, backends, wheels, and runtime identity.
7. **Architecture fit.** The split extends the repository baseline instead of bypassing it: immutable tree identity in internal/buildsource, canonical logical inputs and exact receipts in internal/buildmeta, fingerprinted/rechecked toolchains and stable diagnostics in internal/godriver, plus the accepted language-neutral artifact taxonomy. The source-closure specification and task precondition are byte-identical.

## Independent fact check

Current primary documentation supports the risk-critical claims:

- npm package locks describe the exact generated tree, npm ci rejects manifest/lock mismatch without rewriting, binding.gyp can synthesize node-gyp rebuild, and the npm cache is explicitly non-authoritative: https://docs.npmjs.com/cli/v11/configuring-npm/package-lock-json/ ; https://docs.npmjs.com/cli/v11/commands/npm-ci/ ; https://docs.npmjs.com/cli/v10/commands/npm-rebuild/ ; https://docs.npmjs.com/cli/v11/commands/npm-cache/
- pnpm offline and frozen behavior, skipped local file dependencies in fetch, pnpmfile execution outside ignoreScripts, and hook-produced side-effects caching are documented at https://pnpm.io/cli/install ; https://pnpm.io/cli/fetch ; https://pnpm.io/pnpmfile ; https://pnpm.io/settings/build
- Yarn documents source-tarball offline mirrors, frozen/offline Classic installs, modern immutable cache and skip-build behavior, network disablement, plugin resolver/fetcher/linker/hooks, and automatic install/pack for raw Git sources: https://classic.yarnpkg.com/blog/2016/11/24/offline-mirror/ ; https://classic.yarnpkg.com/en/docs/cli/install ; https://yarnpkg.com/cli/install ; https://yarnpkg.com/configuration/yarnrc ; https://yarnpkg.com/features/extensibility ; https://yarnpkg.com/advanced/lifecycle-scripts
- Python primary specifications confirm pylock reproducible-install records, all-dependency hash mode, PEP 517 dynamic requirements and contained backend-path, sdist structure, wheel RECORD coverage/hashes, and no-index/no-deps bundle replay: https://packaging.python.org/en/latest/specifications/pylock-toml/ ; https://pip.pypa.io/en/stable/topics/secure-installs/ ; https://peps.python.org/pep-0517/ ; https://packaging.python.org/en/latest/specifications/source-distribution-format/ ; https://packaging.python.org/en/latest/specifications/binary-distribution-format/ ; https://pip.pypa.io/en/latest/topics/repeatable-installs/

No checked claim contradicted the outcome. Local probes reproduce the recorded Node 25.6.1, npm 11.10.1, Yarn Classic 1.22.22, Python 3.14.4, and pip 26.1 versions. pnpm remains unavailable with exit 127 exactly as disclosed; the outcome makes no local pnpm conformance claim.

## Artifact and verification evidence

- Project research file and board outcome are byte-identical, 818 lines, SHA-256 `68ecaad383fc3fd7b2704065f0d1e7d78446c5c7f535b4fcbfdd669e7003fe4f`.
- Required sections are present, four Markdown fences are balanced, local references resolve, and no trailing whitespace was found.
- `task-board validate --json` returned valid true with no errors or warnings.
- `go test -count=1 ./...` exited 0 across every package. Notable results include cmd/curator 506.857s, internal/buildsource 0.590s, internal/buildmeta 2.278s, internal/godriver 70.662s, internal/buildrepo 36.512s, and internal/transaction 57.309s.
- `go vet ./...` exited 0.

Reviewer environment anomaly: a default cached `go test ./...` completed its test child, then entered the already known cmd/go cache-input traversal over workspace state and grew to about 14.5 GB RSS with no test child or failure. The reviewer terminated only that reviewer-launched coordinator. The full `-count=1` suite above bypassed the cache-ID path and passed, so this unrelated workspace contamination does not weaken test evidence or the research artifact.

## Routing decision

Accept the research deliverable and route `TASK-260810-2n3sbi` to `done`. This reviewer supplies no `commit_ack`; other discovery tasks remain active, so this transition does not complete the parent Story.