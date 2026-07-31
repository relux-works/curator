# TASK-260720-3ag6pi — review cycle 4 verdict

## Verdict

**ACCEPTED**

The cycle-3 workflow-drift defect is closed. The submitted cycle-4 change fits
the curator-spec validation architecture and is limited to:

- `.github/workflows/ci.yml`
- `.github/workflows/release.yml`
- `tools/test_validate.py`

Against reviewed product candidate
`ddb181ca3b8e243f212e90ff26fcabe2234fb669`, the delta is exactly 37
insertions and 2 deletions. All protocol, schema, vector, generated,
documentation, and release-metadata bytes are unchanged.

## Independent review evidence

- Tool readiness: task-board 0.23.0, Git 2.50.1, GNU Make 3.81,
  Go 1.25.5, Python 3.14.4 with the task-local pinned environment,
  jq 1.8.1, and actionlint 1.7.12.
- `make validate` on the exact three-file reworked tree exited 0:
  42 schemas, 447 vector files, 60 Python tests with no skips reported,
  and Go tests.
- `WorkflowRegenerationScopeTests` passed independently in verbose mode.
- `actionlint .github/workflows/ci.yml .github/workflows/release.yml`
  exited 0.
- `make regenerate-check` on the exact reworked tree exited 0 and left only
  the three intended non-generated modifications.
- The two retained independent regeneration trees compare byte-identically:
  conformance tree `diff -qr` rc 0, rc.5 metadata `cmp` rc 0, rc.6 metadata
  `cmp` rc 0. Both manifests hash to
  `72c5d717027ca096b14bc32f5d60bb740676974e9429f3d09b730897e5fba89b`.
- `make release-check VERSION=1.0.0-rc.6` independently exited 0 on clean
  product commit `ddb181ca3b8e243f212e90ff26fcabe2234fb669`, tree
  `fa0eb87cacf6cd9ec8510f7d26a91ea76b6d74fd`, parent
  `f5d7673039226ab81de2f4f87e2155ae995c4df3`. The exact three-file
  review delta was separately covered by full validation, focused parity,
  actionlint, and regenerate-check. The release clean-check was not bypassed
  or spoofed.
- The retained negative fixture proves the defect and fix: the old workflow
  scope returned 0 on rc.6-only metadata drift; the corrected scope returned
  1 and identified `release/1.0.0-rc.6.json`.
- `go vet ./...`, `go test ./... -count=1`, Go formatting, and
  `git diff --check` all exited 0.

## Inventory, compatibility, and safety

- All 447 manifest entries are sorted, unique, present, and hash-correct.
- The schema-case index has 376 exact entries: agent-skill-v6 24,
  csk-skill-v6 24, build-receipt-v1 18, install-marker-v2 14,
  claim-v2 7, and claim-v3 13.
- The manifest includes 13 Go fixture files, 11 build-driver expected files,
  all 376 schema-case files, and the required build-driver, acquisition,
  host-policy, claim-qualification, and manager-lifecycle vectors.
- Build-driver coverage remains 8 positive, 77 rejection, 10 build-source,
  and 12 toolchain cases. Manager lifecycle remains 32 named cases, including
  all 22 accepted compiled lifecycle cases.
- Against pre-v6 baseline
  `57c1f56846d221ecc55786bd3c2467ec32f11730`, the 12 frozen agent-skill
  and csk-skill schemas 1–5 plus marker-v1 and claim-v1, their 24 baseline
  cases, their schema-index rows, and their manifest rows compare unchanged.
  Marker v1 remains schema version 1; claim v1 remains schema version 1 and
  protocol `1.0.0-rc.3`.
- All 77 build-driver rejection cases require reject/no execution/no reuse.
  Compiled dry-run forbids Go list/build and records no artifact execution.
  Rc.6 emits no conformance claim and advances no committed downstream pin.
- The updated cycle-4 coverage matrix maps every story criterion and minimum
  rejection cluster to executable passing evidence.

## Release and supersession boundary

The reviewed rc.5 resolution and subsequent rc.6 rework directive supersede
the task's original literal rc.4 command. Literal rc.4 is truthfully retained
as expected-red evidence; it is not represented as passing release evidence.

Current remote provenance remains:

- `origin/main`:
  `f5d7673039226ab81de2f4f87e2155ae995c4df3`
- annotated `v1.0.0-rc.5` tag object:
  `3780f50847ce7f513436c950ec1656a8ab298185`
- peeled rc.5 commit:
  `f5d7673039226ab81de2f4f87e2155ae995c4df3`
- no remote rc.6 tag

This verdict accepts the reviewed candidate and its conformance evidence. It
does not claim that rc.6 has been committed, signed, tagged, or published.
The reviewer modified no product code, staged or committed nothing, and
supplied no `commit_ack`.
