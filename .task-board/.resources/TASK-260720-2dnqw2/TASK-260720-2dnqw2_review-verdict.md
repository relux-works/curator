# TASK-260720-2dnqw2 review verdict

## Verdict

CHANGES REQUESTED. Route to `to-dev` for implementation rework and a fresh
review cycle.

The candidate reproduces the required rc.5 golden bytes and hashes and its
declared gates are independently green, but two reader/model false accepts
violate the portable protocol boundary.

## Reviewed scope and provenance

- Candidate worktree:
  `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-2dnqw2/worktree`
- Base SHA and unchanged worktree HEAD:
  `97a0ed870782b48eebc5a9c25a9cfa8fea5ff245`
- Candidate branch:
  `task/TASK-260720-2dnqw2-canonical-build-metadata`
- rc.5 candidate root:
  `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1`
- Candidate manifest SHA-256, retained as non-release evidence:
  `b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c`
- Independently confirmed canonical input:
  869 bytes,
  `sha256:529370122ae11e2e961d5265b1a020e046bcd43165b2eb96b05e73a51187ac9b`
- Independently confirmed exact receipt:
  1120 bytes,
  `sha256:919fbbad8e6ce95532219fd952c2309d0d7026f85209650508fd6834af4020cd`
- Reviewed all changed and new source/test files. No candidate code was edited,
  staged, committed, or discarded by this reviewer.
- `TASK_BOARD_RUN_ID=RUN-260730-c1d794` is not goal-bound and has no operator
  directives.

## Blocking finding 1: non-UTF-8 protocol JSON is accepted

`src/csk/protocol_json.py:18-35` and `:40-71` pass byte input directly to
`json.loads`. Python auto-detects UTF-16 and UTF-32 byte streams. The BOM check
at `src/csk/protocol_json.py:139-142` recognizes only a UTF-8 BOM. Consequently
both `loads` and `loads_canonical` accept UTF-16/UTF-32, with and without those
encodings' BOMs.

This reaches the task-owned marker reader at
`src/csk/install_marker.py:427-436`, and the refactored registry reader also
inherits the false accept. Independent reproduction against the candidate:

```text
utf-16 protocol_json.loads ACCEPT
utf-16 protocol_json.loads_canonical ACCEPT
utf-16 audit_registry.load_protocol_json ACCEPT
utf-16-le protocol_json.loads ACCEPT
utf-16-le protocol_json.loads_canonical ACCEPT
utf-32 protocol_json.loads ACCEPT
utf-32 protocol_json.loads_canonical ACCEPT
utf-32-le protocol_json.loads_canonical ACCEPT
```

The caller-supplied valid marker-v2 golden, re-encoded as each of `utf-16`,
`utf-16-le`, `utf-32`, and `utf-32-le`, was accepted four times as
`InstallMarkerV2`.

This violates the normative rule that protocol JSON MUST be UTF-8 without a
BOM and that parsers reject invalid UTF-8. It also contradicts the new CCJ-1
reader contract to reject every forbidden ambiguity.

Required rework:

1. Decode byte inputs explicitly and strictly as UTF-8 before JSON decoding.
2. Reject a BOM at the decoded-text boundary and never let `json.loads`
   auto-detect another byte encoding.
3. Add regression tests for UTF-16/UTF-32, with and without BOMs, through
   `protocol_json.loads`, `loads_canonical`, `read_install_marker`, and the
   shared audit-registry reader.
4. Preserve the exact accepted UTF-8 goldens and existing error wrapping.

## Blocking finding 2: malformed native target/toolchain identities are typed and hashed

`src/csk/builds/metadata.py:525-540` checks a tuning key against the union of
all Go tuning names, but does not require the key applicable to
`target.goarch`. `src/csk/builds/metadata.py:507-522` checks only the toolchain
version string's length and forbidden line characters; it does not validate
the normalized `go version` identity or bind its OS/architecture to the native
target.

Independent mutations of the shared 869-byte input produced these false
accepts from `parse_build_input`:

```text
target darwin/arm64 + tuning {"GOAMD64":"v1"}        ACCEPT
target darwin/arm64 + go_version linux/amd64         ACCEPT
```

Both returned a `GoBuildInput` that can be canonicalized and assigned a cache
key. The portable go-v1 contract requires exactly one applicable trusted
tuning variable and a native target/toolchain identity established from the
same normalized probe. The task acceptance criteria also require malformed
identities to be rejected.

Required rework:

1. Reuse one authoritative architecture-to-tuning mapping and reject an
   unsupported architecture or the wrong tuning key.
2. Reuse or expose the existing strict normalized Go-version parser and bind
   its GOOS/GOARCH to the native target.
3. Add focused parser/constructor negatives proving these combinations cannot
   be canonicalized or keyed.

## Independent green controls

- Focused accepted-root pytest:
  `170 passed in 0.32s`, exit 0.
- Audit-registry regression pytest:
  `42 passed in 2.82s`, exit 0.
- Strict mypy:
  `Success: no issues found in 61 source files`, exit 0.
- `git diff --check`: exit 0, no output.
- Existing golden, receipt, marker-v1/v2, execution-policy non-alias, duplicate
  key, unsafe integer, unknown-field, noncanonical receipt, hash, version, and
  artifact-path assertions pass.

The verdict is based on demonstrated false accepts outside the current test
matrix, not on a failing existing gate.

## Logbook note

The standalone `logbook` command is unavailable in this environment. Per the
project's established fallback, this finding is persisted in this outcome and
in task notes as the durable reviewer logbook record. The repository logbook
was not edited because this reviewer role is read-only.
