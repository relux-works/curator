# TASK-260720-2dnqw2 review rework cycle 2

## Provenance

- Product repository: `/Users/iv/Developer/intranet/cocoaskills`
- Existing task worktree:
  `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-2dnqw2/worktree`
- Task branch: `task/TASK-260720-2dnqw2-canonical-build-metadata`
- Unchanged base/HEAD:
  `97a0ed870782b48eebc5a9c25a9cfa8fea5ff245`
- The clean CocoaSkills `main` clone remained equal to `origin/main` at that
  SHA. No file was staged, committed, or discarded.
- Accepted dependency handoffs remained `done` with outcome evidence:
  `TASK-260720-3c0ss2`, `TASK-260720-3j8pp5`, and
  `TASK-260729-3nx97g`.

The inherited caller-supplied rc.5 candidate root was exported explicitly as
`CURATOR_CONFORMANCE_ROOT` for every conformance-driven test command:

`/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1`

Its `manifest.json` SHA-256 remains
`b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c`.
This is candidate-only non-release evidence; no release pin or conformance
claim changed.

## Reviewer findings closed

### Strict UTF-8 protocol boundary

Byte input is now decoded explicitly with strict UTF-8 before `json.loads`
receives text. A decoded UTF-8 BOM is rejected at the same boundary. Python can
no longer auto-detect UTF-16 or UTF-32 for protocol or CCJ-1 bytes.

Regression coverage rejects UTF-16 and UTF-32 in little- and big-endian forms,
both with and without BOMs, through:

- `protocol_json.loads`
- `protocol_json.loads_canonical`
- `audit_registry.load_protocol_json`
- `install_marker.read_install_marker`

Accepted UTF-8 vectors and the existing marker/audit error wrappers are
unchanged.

### Native target/toolchain identity binding

Portable metadata now reuses the trusted toolchain module's authoritative
`GOARCH -> tuning variable` mapping. The trusted normalized `go version`
parser is exposed as a narrow shared helper and reused by metadata validation.

Build-input parsing and direct construction now reject:

- unsupported target architectures;
- a tuning key that does not belong to the target architecture;
- malformed normalized Go-version identities;
- Go-version identities that would change during line-ending normalization;
  and
- a Go-version OS/architecture that differs from the native target.

These failures occur before an invalid input can be canonicalized or assigned
a cache key.

## Canonical identity controls

The accepted-root focused suite reconfirmed:

- 869-byte build input and cache key
  `sha256:529370122ae11e2e961d5265b1a020e046bcd43165b2eb96b05e73a51187ac9b`;
- 1120-byte exact receipt and receipt hash
  `sha256:919fbbad8e6ce95532219fd952c2309d0d7026f85209650508fd6834af4020cd`;
- required `policy.execution_policy=manager-worker-v1`;
- distinct schema-invalid legacy and reserved-hardened identities;
- marker-v1 compatibility and deterministic marker-v2 build state.

## Exact command ledger

Every gate ran as a standalone process without `tee` or a masking pipeline.

1. Test-first reviewer reproduction:

   `CURATOR_CONFORMANCE_ROOT=... ../venv/bin/python -m pytest tests/test_protocol_json.py::test_protocol_readers_reject_non_utf8_byte_encodings tests/test_build_metadata.py::test_build_input_parser_rejects_malformed_native_identities_before_keying tests/test_build_metadata.py::test_build_input_constructor_rejects_wrong_tuning_and_toolchain_target tests/test_build_metadata.py::test_marker_reader_rejects_non_utf8_byte_encodings -q`

   Exit `1`: `14 failed in 3.41s`. This was the expected-red proof that all
   reviewer false accepts remained reproducible before the implementation
   change.

2. Same reviewer regression set after the fix:

   Final-state rerun exit `0`: `14 passed in 0.08s`.

3. Focused accepted-root pytest:

   `CURATOR_CONFORMANCE_ROOT=... ../venv/bin/python -m pytest tests/test_build_metadata.py tests/test_protocol_json.py tests/test_protocol_conformance.py -q`

   Final-state rerun exit `0`: `184 passed in 0.30s`.

4. Strict type check:

   `../venv/bin/python -m mypy`

   Exit `0`: `Success: no issues found in 61 source files`.

5. Full accepted-root pytest:

   `CURATOR_CONFORMANCE_ROOT=... ../venv/bin/python -m pytest -q`

   Final-state rerun exit `0`: `849 passed, 6 skipped in 84.74s`. The six skips are existing
   platform-conditional cases.

6. Distribution build:

   `../venv/bin/python -m build`

   Exit `0`: sdist and universal wheel built successfully, including
   `csk/install_marker.py`, `csk/protocol_json.py`,
   `csk/builds/metadata.py`, and `csk/builds/toolchain.py`.

7. Distribution metadata:

   `../venv/bin/python -m twine check dist/*`

   Exit `0`: wheel and sdist both `PASSED`.

8. Repository diff hygiene:

   `git diff --check`

   Exit `0`, no output. The project defines no separate Ruff, Black, or
   Pyflakes gate.

9. Candidate manifest provenance:

   `shasum -a 256 .../conformance/v1/manifest.json`

   Exit `0`, digest
   `b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c`.

Tool readiness was reconfirmed with exit `0` for Git 2.50.1, ripgrep 15.2.0,
task-board 0.23.0, Python 3.14.4, pytest 9.1.1, mypy 2.3.0, build 1.5.0, and
Twine 7.0.0.
