# Reviewer verdict for TASK-260811-i3154q

Verdict: **changes requested -> to-dev**

## Goal, scope, and reviewed snapshot

- Reviewer run: `RUN-260811-774fc7`
- Authoritative reviewer goal at the last pre-verdict checkpoint:
  `GOAL-260811-258815` revision 1
- Resolved scope: `TASK-260811-i3154q`
- Review policy: `required`
- Reviewed rework evidence:
  `TASK-260811-i3154q_rework-evidence_RUN-260811-07310a.md`, SHA-256
  `1b58d0af5525fb0dcbef4bc5fcc9016ee4e547dae37c869fb5c8246be1f0d1f9`
- Reviewed `internal/closuregraph` sorted per-file SHA-256 manifest:
  `bba7611d282747ae0a9d6dc77e9eb26e67db2efa3f5ed6a66296ac3a4594b725`
- Normative graph/checkpoint decision SHA-256:
  `874c3a40b9ff9fcf130f37f22e9f5aa2bdd6d9f246403e059c98e84736e27bbc`
- Exact CGP05/CGP10 corpus SHA-256:
  `fed9657b33297650e64178d605c7ee7f47dde640fefd12a99dec3651bb9cadcb`

This artifact records only the changes-requested branch. Both findings are
package-local implementation defects with direct code and regression-test
remedies. They are not a human-only or external blocker, so `blocked` is not
appropriate.

## Acceptance-blocking findings

### 1. A selected node accepts an undeclared extra platform-role binding

The accepted contract requires exact typed `targets` bindings for each declared
product/target/action/toolchain/output/boundary platform slot and rejects
role-mismatched platform references. `validatePlatformRoleRecords` verifies only
that a binding edge's role exists in `SelectionContext` and points to that
role's platform (`internal/closuregraph/validation.go:461-481`).
`validatePlatformBindings` then counts each declared role but never rejects a
selected node's additional `targets` edge for a role that node did not declare
(`internal/closuregraph/validation.go:792-815`).

Independent overlay probe
`TestReviewerProbeRejectsUndeclaredExtraHostTarget` added a valid host platform
and host selection role to CGP10, retained every required target binding, and
added a host `targets` edge from the command product whose intrinsic payload
declares only the target role. `ProjectActive` returned success. The extra edge
therefore enters the binding and active identities despite having no declared
slot on its source node.

Required rework: for every selected `targets` edge, require its normalized
binding role to be present in the source node's closed declared-platform-role
set, while preserving the specified host-to-target fallback when no distinct
host role is selected. Retain the existing exactly-once check for every
declared role. Add product, target, action, toolchain, output, and boundary
negative/permutation regressions for undeclared extra target and host roles.

### 2. Duplicate semantic edges are accepted when their evidence origins differ

The accepted contract separately requires unique edge identities and rejection
of duplicate semantic edges, with all conflicting evidence origins in the
diagnostic. `semanticEdgeKey` currently hashes the complete edge payload,
including `EvidenceOrigin` (`internal/closuregraph/validation.go:1135-1137`).
Consequently the same relationship can be repeated under different provenance
fields and is treated as two different semantics rather than one duplicated
semantic relation.

Independent overlay probe
`TestReviewerProbeRejectsDuplicateSemanticRequiresAcrossOrigins` added two
binding-owned `requires(scope=toolchain)` edges with identical action and
toolchain endpoints, scope, condition, and dependency kind. Only `edge_key` and
`origin.field` differed. `ProjectActive` returned success, so both copies enter
the canonical binding and active graph.

Required rework: derive the duplicate-semantic key from edge kind, endpoints,
and the relationship-defining payload fields while excluding provenance-only
origin fields; retain the distinct edge IDs for diagnostic evidence and report
all conflicting origins deterministically. Add cross-origin duplicates for all
origin-bearing edge payloads plus ordering/permutation regressions.

## Independent verification

The following current-snapshot gates passed:

- `go test -count=1 ./internal/closuregraph`
- `go test -race -count=1 ./internal/closuregraph` — 108.016s
- `go test -count=1 -cover ./internal/closuregraph` — 81.9%
- `go test -shuffle=on -count=10 ./internal/closuregraph` — 96.942s
- `go vet ./internal/closuregraph`
- `go build ./internal/closuregraph`
- pinned `golangci-lint` v2.12.2 on `./internal/closuregraph` — 0 issues
- `gofmt -l internal/closuregraph` — no output
- accepted Ruby canonical oracle — 53 labeled records and all references pass
- `git diff --check`
- `task-board validate`
- released, uncached `go test -count=1 ./...` — exit 0; notable timings:
  `cmd/curator` 353.398s, `artifactpolicy` 127.662s, `closuregraph` 15.811s,
  `install` 112.798s, and `install/atomicity` 113.925s

The repository-wide run used an independently fingerprinted 350-file Go-source
snapshot. Its sorted per-file fingerprint was identical before and after at
`8a90026917c29cf08191abb2f408770fd6f2555e9598f082c010eb1a6cc837e1`.

The overlay-only adversarial command
`go test -count=1 -overlay=.temp/reviewer-probes/RUN-260811-774fc7/overlay.json ./internal/closuregraph -run '^TestReviewerProbe'`
exited 1 with both expected failures because each invalid graph was accepted.
The probe source SHA-256 is
`6bddd1efc80da2b7cba4fdd799b4c6bf1f9714bca717e6ec662b82faba11cb52`.
No product source file was modified by this reviewer.

## Verdict route

Route `TASK-260811-i3154q` to `to-dev`. The next producer should close both
findings, preserve the exact accepted corpus, add permanent regressions, refresh
task-scoped evidence, and return the task for another independent reviewer
cycle.
