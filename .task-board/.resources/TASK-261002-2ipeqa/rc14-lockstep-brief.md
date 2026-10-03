# THE ONLY CURRENT INSTRUCTION — TASK-261002-2ipeqa: curator accepts the rc.14 candidate suite (lockstep; writer stays OFF)

The curator-spec v1.0.0-rc.14 release prep is on branch `candidate-rc14` of relux-works/curator-spec (commit e3a88ced; its core manifest is sha256 **6f832d813efc768ea154a7d5076b512ab4be6aa9409d92e11469d21ea9bc69f5**). The source suite is unchanged (061ec05d…).

Content-wise, rc.14 equals the current hash-v2 + Muse candidate. The release prep changed protocol-version metadata and therefore the manifest digest. Curator's exact case-count tables and gap ledger are keyed by manifest digest (internal/conformancecoverage, .github/ci/conformance-case-counts.tsv, conformance-gaps.tsv). They must recognise the new digest, or the spec PR's Implementations job fails.

Do:
1. Add the rc.14 digest to the digest dispatch, with EXACT per-family counts and gap rows. Derive them from the candidate corpus; never guess. Use the existing candidate rows as the template. Keep rc.13 and the previous candidate digests accepted.
2. Do NOT change `internal/hashing` EnableV2Writers: it stays false. Do NOT move .github/workflows/ci.yml SPEC_PIN: it stays rc.13. Both happen in the later cut-over leaf after the rc.14 tag.
3. Run `CURATOR_CONFORMANCE_ROOT=<checkout of curator-spec candidate-rc14>/conformance/v1 go test ./internal/conformancecoverage ./cmd/curator -run 'Conformance|Coverage' -count=1`, plus the default-pin run, and record real exit codes. Also confirm the full Implementations shape the spec workflow uses passes against the candidate: see curator-spec .github/workflows/implementations.yml for the exact command.

Read and follow the attached host-rules.md (-work, syspolicyd). One CHANGELOG line. Never edit LOGBOOK.md. Never spell any employer name. Base: curator main 20368da6.

Update the results, then run `task-board handoff TASK-261002-2ipeqa --role developer`, then END YOUR TURN.
