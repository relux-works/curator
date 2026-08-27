# TASK-260720-2284br review verdict — cycle 4

## Verdict

Changes requested. Route to `to-dev`.

R4 closes the previously demonstrated one-way project/global manifest races,
and the permanent regressions for project, global, development substitutions,
and hybrid activation pass. The implementation still does not bind an
optimistic digest to the bytes actually used to construct the closure. A
supported A → B → A writer sequence can therefore pass revalidation and commit
the transient B closure while the current declaration state is A.

## R5 — declaration observations permit an ABA stale-closure commit

Severity: high; acceptance-blocking.

The four declaration inputs all use the same sequence:

1. digest the path outside the home lock;
2. read and parse the path separately;
3. much later, under the home lock, digest the current path and compare it only
   with the first digest.

Relevant sites:

- project manifest: `internal/install/install.go:177-178`;
- project substitutions: `internal/install/install.go:217-218`;
- hybrid activation manifest: `internal/install/install.go:254-255`;
- global manifest: `internal/install/global.go:79-81`;
- comparison: `internal/install/commit.go:321-357`.

That catches A → B, but not A → B → A around the separate read. In the ABA
case, closure construction consumes B while both recorded and rechecked
digests are A. The run consequently sees no restart reason.

### Independent deterministic reproduction

The read-only overlay adds test-only hooks immediately after the project
manifest digest and immediately after `manifest.Load`; product files are not
modified. The test starts from a manifest canonicalized through the real
`manifest.AddDecl` writer, then:

1. adds `skill-b` through `manifest.AddDecl` after the A digest;
2. lets `manifest.Load` consume B;
3. removes `skill-b` through `manifest.RemoveDecl`, restoring byte-identical A;
4. requires a closure restart and absence of `skill-b` state.

Command:

```text
go test -overlay .temp/TASK-260720-2284br/reviewer/aba-overlay/overlay.json \
  ./internal/install -run '^TestReviewerProjectManifestABARestartsClosure$' \
  -count=1 -v
```

Actual result: exit 1. No restart message is emitted, and the result reports
both `skill-a` and transient `skill-b` installed. Evidence:

- `.temp/TASK-260720-2284br/reviewer/aba-overlay/repro.log`
- `.temp/TASK-260720-2284br/reviewer/aba-overlay/repro.exit`
- overlay sources and map in the same directory.

Impact: an install can commit context, shims, adapters, and consumer state for a
closure that corresponds to neither the initial nor current declaration state.
This violates the acceptance criterion that stale closure or activation state
restarts instead of applying an old plan.

### Required rework

1. Bind each observation to the exact bytes used by the parser. A suitable
   design is one read that returns both immutable bytes/digest and the parsed
   value, followed by the existing under-home-lock recheck; an equivalent
   before/read/after protocol is acceptable if it proves the parsed bytes
   matched the recorded generation.
2. Apply the invariant to all four declaration inputs, not only the project
   manifest.
3. Add permanent ABA regressions using the real supported writers where they
   exist. At minimum prove that a transient declaration/activation cannot
   commit stale context, shims, adapters, or consumer state.
4. Preserve the existing one-way mutation regressions and bounded
   `restartClosure` routing before cache publication and target staging.

## What remains closed

- R1/R2 journal ownership, entry-kind symlink restoration, deterministic
  target classes, stale removals before consumer, and reverse rollback remain
  closed by inspection and
  `go test -count=1 ./internal/transaction ./internal/staging ./internal/adapters`.
- R3 hybrid one-way activation revalidation remains closed by the four focused
  activation cases.
- R4 one-way project/global/substitution cases pass their six focused permanent
  tests. The problem is the untested ABA relation between digest and read, not
  the later under-lock ordering.

## Verification and archive assessment

Independent final-tree checks:

- `gofmt -l .` — exit 0, no output;
- `git diff --check` — exit 0;
- `go build ./...` — exit 0;
- `go vet ./...` — exit 0;
- pinned `golangci-lint` v2.4.0 `run ./...` — exit 0;
- six submitted R4 revalidation cases — exit 0;
- R1/R2 transaction/staging/adapter packages — exit 0;
- four R3 activation cases — exit 0;
- `go test -count=1 ./internal/godriver` at default parallelism — exit 0.

The submitted gate archive SHA-256 is
`81f155c786ad015ab818dc48bebb7146050a0dd9ef838cfee82571cd51fe9475`.
Its durable exits, PWD, and timestamps are internally consistent and cover
source older than the 18:42:38 start. The recorded `internal/godriver` red runs
contain only the documented 15-second `process_timeout` failures; the archive
also contains passing serial and later default-parallel reruns, and the
independent default-parallel run passes. I treat that as honest non-task
contention evidence, not this verdict's blocker.

No product code was modified, staged, committed, or published during review.
