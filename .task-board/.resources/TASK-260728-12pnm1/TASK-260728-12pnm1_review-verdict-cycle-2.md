# TASK-260728-12pnm1 review cycle 2 verdict

## CHANGES REQUESTED

Route to `analysis`. This is research and contract rework, not implementation code rework. Reviewer supplied no `commit_ack`.

## Blocking finding 1: CRLF normalization is not total at the 4096-byte boundary

Reference section 2.3 applies `len(stdout) <= 4096` before folding CRLF to LF, while the next step claims that folding makes Windows and Unix probes of the same distribution produce the same payload. The probe implements the same order in `vvrecord.go`: raw length rejection occurs before `strings.ReplaceAll`.

A constructed valid multiline LF stream consisting of 4093 `x` bytes followed by LF, `A`, LF is exactly 4096 raw bytes and passes the size gate. Its CRLF-equivalent stream is 4098 raw bytes and is rejected before folding, even though folding it produces byte-for-byte the accepted LF stream. Therefore line-ending-equivalent inputs can have different admission verdicts, contradicting the cross-platform normalization claim and the cycle-2 review requirement to close size and CRLF boundaries.

The current green evidence does not cover this boundary. N2 uses only the 192-byte measured stream. N12 tests one 4097-byte LF stream but no LF/CRLF-equivalent pair at the limit. Required rework: define the semantic length bound after CRLF folding, with a separate larger raw capture bound if resource protection needs one; add LF and CRLF pairs whose folded forms are exactly 4096 and 4097 bytes; require equal verdicts and equal payload/digest whenever admitted; update decision, reference, probe and vector inventory. Existing measured hashes should remain unchanged.

## Blocking finding 2: Stage P requires canonicalization of a path it explicitly permits not to exist

Reference section 4.3 says a non-dependency target path may be absent and that absence is not a Stage P failure. The immediately following rule says that in both dependency and non-dependency cases the canonical form of the joined path must remain under `R`. A canonical filesystem form cannot be obtained for a non-existent leaf, so an implementation cannot obey both requirements literally.

Required rework: specify a total physical algorithm. For example, lexically admit the absent leaf, non-following-stat every existing ancestor, reject every symlink or reparse component, canonicalize the deepest existing ancestor under `R`, and canonicalize the leaf only when it exists. Add vectors for absent targets with safe existing ancestors and for absent targets below a symlink or reparse ancestor. Also correct the section 7.2 statement that every G2 through G8 row has a Stage P counterpart: an auto-discovered `build.rs` with no `package.build` member is only caught by G2 unless Stage P gains an explicit file rule.

## Independent evidence

- Recomputed V payload, V record, C payload and C record hashes exactly: `7d8e0833...`, `7fc35c11...`, `8d712854...`, `d677e668...`.
- Extracted probe: `gofmt -l` clean; `go test ./...`, `go vet ./...`, `go build ./...` exit 0.
- Full macOS arm64 host replay with the declared Rust root and SDK exits 0: 19 of 19 cases, P1 and P2 true, 24 closure checks with zero verdicts, all 10 controls failing as required, 68 vectors with zero divergences, 13 structural checks with zero divergences.
- Controls-only replay exits 0 with all 10 controls failing as required.
- Repository `go test ./...`, `go vet ./...`, and `go build ./...` exit 0; `git diff --check` exits 0.
- Producer decision, reference and probe archive SHA-256 values were identical before and after review. No producer artifact, project code, staging, commit, publication or pin was modified.

All other cycle-1 blockers are closed by the inspected evidence: the exact graph vector contains `--all-features`; Stage P precedes Cargo and covers declared outside-root and ancestor-workspace origins; the operation-private Cargo config uses a closed literal-string template plus write-back verification; local and external source mappings replay equivalently; platform claims remain macOS arm64 only with Windows and Linux expressed as qualification obligations.