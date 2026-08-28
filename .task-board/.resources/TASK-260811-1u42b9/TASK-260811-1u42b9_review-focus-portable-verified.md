# npm review focus after portable/verified correction

Independently review the complete task and all acceptance criteria. In addition,
the verdict must explicitly establish each point below from code and fresh test
evidence rather than trusting the producer summary.

1. Portable assurance is the functional default and can derive the private npm
   cache, run frozen `npm ci --offline --ignore-scripts`, reconcile the lock and
   installed graph, and invoke Node without a verified provider.
2. Portable evidence does not synthesize or claim lossless host network, read,
   write, or process observation. Default-zero fields must not be treated as
   proof of absence when portable cannot observe them.
3. Verified assurance requires an exact compatible lossless provider binding
   before any npm process start. Missing, incomplete, cross-mode, or drifted
   binding fails closed with a demonstrated zero-start negative.
4. Manager-owned no-resolver/no-ambient authority remains enforced by admitted
   inputs, fresh private roots, exact invocation, frozen lock, private cache
   receipt, installed-graph reconciliation, and declared outputs. Lossless
   provider observations are an additional verified-mode gate, not a portable
   prerequisite.
5. The vendored compiled-binary prohibition remains effective for tarball and
   materialized inputs, including native addons, Wasm, V8 cache, opaque payloads,
   bundled dependency trees, implicit `binding.gyp`/node-gyp, and lifecycle
   execution.
6. Rerun affected focused, race, coverage, vet, lint, build, diff, board
   validation, and an uncached repository-wide suite after the final assurance
   correction. Record exact exits and any deviations.

Accept only if the task reaches these semantics without narrowing portable
mode to verified-only or inflating portable evidence.
