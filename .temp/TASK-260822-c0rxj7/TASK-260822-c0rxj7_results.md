# TASK-260822-c0rxj7 — Schema 8 rc.9 candidate evidence

## Candidate

- Repository: `relux-works/curator-spec`
- Branch: `candidate/schema-8-rc.9`
- Full commit SHA: `859727b103ed175ff214cbb64641f4686d8c6a68`
- Inputs: script-worker `dd9c9fc079470f03f247b71efb52d4de6b204e78`; module-roots `bac193cadb7d26aabf006c92924b4a05f6574e31`
- Suite manifest SHA-256: `sha256:782d6868f6d9725f7bf38d3fb1944f307f2d5d9c060b8816b0f55a5c2e97f11f`
- Tree SHA-256: `sha256:f88a76263040c90470be8f18c110927267ac5862f7e25b3b161bcd5ef97319f3`
- File count: `692`
- Historical `release/1.0.0-rc.8.json` SHA-256: `293f101d10665061aa049efa72141f9e3c5d608bbde300e882f6e3e095e31ede`, byte-identical to `origin/main`

The candidate adds rc.9 release metadata and a new conformance-claim-v5 surface. Conformance-claim-v4 remains byte-frozen and bound to rc.8. `SPEC_PIN` was deliberately not advanced.

## Validation

All commands below were run directly; exit codes are the real process results.

| Gate | Exit | Result |
| --- | ---: | --- |
| `make validate` | 0 | 53 schemas, 691 vector files, 97 Python tests, Go tool tests |
| `make release-check VERSION=1.0.0-rc.9` | 0 | rc.9 release gate passed at the candidate SHA |
| `gofmt -l tools` | 0 | no output |
| `git diff --check` | 0 | clean |
| generator pass 1 | 0 | regenerated both vector families |
| generator pass 2 | 0 | regenerated both vector families |
| checksum inventory comparison | 0 | both passes byte-identical |
| rc.8 comparison against `origin/main` | 0 | byte-identical |
| ancestry check: script-worker input | 0 | candidate contains `dd9c9fc` |
| ancestry check: module-roots input | 0 | candidate contains `bac193c` |

Specification CI run [32633567102](https://github.com/relux-works/curator-spec/actions/runs/32633567102) passed on the exact candidate SHA across Linux, macOS, and Windows, including formatting, links, and release provenance.

## Candidate-conformance result and blocker

Curator candidate-conformance run [32633572039](https://github.com/relux-works/curator/actions/runs/32633572039) was dispatched with the exact immutable candidate SHA and manifest digest. Its candidate matrix is definitively red for two concrete external reasons (an unrelated normal Windows test job was still running when this packet was recorded):

1. Ubuntu resolves and records the exact candidate identity, then fails `internal/install TestAuthoritativeDryRunCasesMutateNothingPersistent/compiled-cache-miss-is-read-only`: published dry-run scope `multi-project` has no executable binding. The implementation fix exists on open Curator PR #14 (`d345420`) but is not on `main`.
2. Windows fails candidate identity before the suite gate because Git for Windows `shasum` prefixes the digest with `\\` when the hashed path contains Windows escaping. The measured digest is therefore textual `\\782d...` rather than the canonical expected `782d...`; the candidate bytes themselves are unchanged.

The gate was not weakened, the expected digest was not falsified, and the unlanded implementation change was not cherry-picked merely to manufacture green evidence. Schema 8 implementation stories must consume the immutable candidate and add explicit suite-consumption coverage before qualification.

## Handoff

- Preserve branch `candidate/schema-8-rc.9` and its task-scoped worktree while downstream implementations qualify.
- Land/fix Curator's multi-project dry-run binding and the Windows candidate identity normalization.
- Re-run candidate-conformance with the same SHA and digest.
- Only the later landing task may update implementation pins, merge the spec candidate, publish rc.9, and advance `SPEC_PIN`.

Unblock notes with the exact candidate evidence were recorded on Curator story `STORY-260822-2h0v9j` and CocoaSkills story `STORY-260822-2evh3p`. The same evidence was routed to `TASK-260822-f4qv7w` and `TASK-260822-1so0ym`; their attempted transition from `blocked` to `to-review` was rejected by the board because both dependency edges point at this still-active candidate task. Their statuses must be released by the later landing workflow once the external conformance blockers are cleared.
