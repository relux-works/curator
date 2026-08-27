# Candidate 6001dc3 — three-OS green matrix

- Identity: `6001dc33281b94a4ec7442ab15278550dd0f51d9`, protocol `1.0.0-rc.9`,
  manifest `sha256:803918bf8672f76cf990985e51db213b826674cd5bb54fbf47731b8404b44403`,
  692 files under `conformance/v1`. Signed (`%G? = G`) and pushed on
  `candidate/schema-8-rc.9`. Supersedes `e66cb72`, which supersedes `edd0721`,
  which supersedes `859727b`; none rewritten.
- Delta over `e66cb72` (two files, both prose, both outside `conformance/v1`,
  so the suite digest is unchanged by construction and measured unchanged):
  - `profiles/manager.md` §11 — the external repository profile is now banded
    on marker v4: a schema-8 installation records marker v4 rather than v3, and
    every marker-v3 obligation of the profile (shim selection in §11.8,
    read-only status validation and GC rooting in §11.9) applies to a valid
    marker v4 without change. This closes review-cycle-2 item 3.1 by fixing,
    not deferring.
  - `schemas/v1/README.md` — the opening paragraph named rc.8's new closed
    objects while the body documented rc.9's; it now names rc.9's
    (`agent-skill-v8`, `csk-skill-v8`, `install-marker-v4`, conformance claim
    v5) as well. This is review-cycle-2 item 3.2.
  - Review-cycle-2 item 3.3 (`conformance/README.md` "every marker role") was
    flagged cosmetic with no action required and was left alone.

## Specification CI

Run [32659168954](https://github.com/relux-works/curator-spec/actions/runs/32659168954)
on exactly `6001dc3`: **success**, all six jobs — Formatting, Links, Release
target provenance, Specification on ubuntu-latest / macos-latest /
windows-latest.

## Candidate-conformance lane

Run [32659157687](https://github.com/relux-works/curator/actions/runs/32659157687),
dispatched from curator `main` (`e17b0f1`) with `candidate_ref=6001dc33…` and
`candidate_manifest_sha256=sha256:803918bf…b44403`: **success**, all fourteen
jobs, first attempt, no reruns.

| Job | ubuntu-latest | macos-latest | windows-latest |
| --- | --- | --- | --- |
| Candidate suite | success | success | success |
| Test | success | success | success |
| Race | success | success | n/a |
| Gate self-test | success | success | success |

Cross-runner job-log facts, read from the logs rather than from a summary:

| Runner | `CANDIDATE_REF` | `manifest_sha256` | digest match | `file_count` | `SPEC_PIN` |
| --- | --- | --- | --- | ---: | --- |
| ubuntu-latest | `6001dc33…` | `sha256:803918bf…b44403` | yes | 692 | `00b1688a…` |
| macos-latest | `6001dc33…` | `sha256:803918bf…b44403` | yes | 692 | `00b1688a…` |
| windows-latest | `6001dc33…` | `sha256:803918bf…b44403` | yes | 692 | `00b1688a…` |

All three also logged `candidate-suite: revision accepted (immutable, full
40-hex)`. `SPEC_PIN` was untouched throughout — this run qualifies a candidate
and claims no release.

## One recorded discrepancy in the same logs

`tree_sha256` is **not** consistent across runners: ubuntu and windows record
`sha256:9d5a10b6…b769`, macos records `sha256:176dc52b…2f02`, for
byte-identical content with an identical, verified `manifest_sha256`. Cause is
the unpinned `sort` locale in curator's `.github/ci/candidate-suite.sh:101`.
It does not affect this qualification because only the manifest expectation is
wired in `ci.yml:365`, but it is a latent gate hole rather than a cosmetic
one. Details and routing in `TASK-260822-c0rxj7_results.md` §6.
