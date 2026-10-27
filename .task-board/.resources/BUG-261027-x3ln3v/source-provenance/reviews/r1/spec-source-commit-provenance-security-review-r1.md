# Source-commit provenance independent security review r1

**CHANGES_REQUESTED** for [curator-spec PR202](https://github.com/relux-works/curator-spec/pull/202).
This is a complete source/security sweep with bounded controls, not acceptance, landing, trust adoption, native qualification or release authorization.

## Exact source and publication

| Boundary | Independently verified result |
| --- | --- |
| Candidate | `ec70b25b5c7b7b23be64a05dacd6534eee71ade3` |
| Base and current advertised main | `74a7dcc503c0fd7acd491299ab71290ae40e708e`; unchanged on final advertisement |
| Repository/default/feature | Canonical `relux-works/curator-spec`, default `main`, feature `fix/source-commit-provenance`; repository API, PR API, advertised and fetched OIDs agree |
| Ancestry and signature | One human-authored signed commit directly on base, signature `G`, fingerprint `SHA256:rE/sRlfdFt/vT5y5yukDCsq4aFtC48pbWsq9z4hm8MA`; exact `Codex gpt-6.1-sol` coauthor trailer; no session metadata |
| Published delta | All 12 remote file patches applied independently in memory to exact base reconstruct byte-identical candidate blobs: 585 insertions, 95 deletions. GitHub coalesces two local hunks; textual hunk layout is not itself content drift |
| Review checkout | Exact detached candidate; clean under selector-free Git inspection. Stable checkout and previous worktrees/evidence preserved |
| Reviewer context | Recorded fresh invocation selects managed `codex_cli`, `gpt-6-astra`, `xhigh`, the requested current profile home, and no resume flag. No profile repair, fallback, native config rewrite or new worker |
| Producer input | Full handover: 12,077 bytes, SHA-256 `976670ed15db5958de49b65694fdf783c5506ace1f15888afc76d27ec834b0d9`; complete published producer COMMENT 5673407186 read. These are inputs, not independent acceptance |

## Findings

### R1-01 — Medium: API error diagnostics reflect bearer-token input

Locations: `tools/verify_release_commit.py:139-155,173-175` and `tools/verify_release_merge_policy.py:34-50,60-62`.

Requirement: API uncertainty must refuse without exposing credentials. Both helpers construct `Authorization: Bearer <token>`, catch `ValueError`, then interpolate the entire exception into the public failure. Python's HTTP header validation raises `ValueError` containing the complete header value when a token contains an embedded newline. Both `main()` functions print that failure to stderr. A malformed environment value therefore turns a refusal into a credential disclosure, including in the locally documented log workflow. This is an error-path confidentiality defect in the touched transport code; it is not a source-signature or release-authority bypass. The old implementation also had an unsafe uncaught-error path; this review does not label it a newly introduced privilege escalation.

Independent reproduction: call each exact candidate helper with a synthetic token containing an embedded newline while patching `http.client.HTTPSConnection.connect` to fail if reached. Real urllib/http.client header validation occurs before connect; both exceptions contain the synthetic token, and connect is never called. `token-diagnostic-witness.json` records both reflections as true. `token-redaction-assertion.log` marks the required redaction property RED. No real token was read, copied, logged or transmitted.

Nearby valid controls: fixed-endpoint responses succeed; absent token, network/error results and actual urllib 301/302/303/307/308 redirects refuse. The redirect controls use an in-process HTTPS transport, and no second authority is contacted. The witness runner's exit 0 means the defect was reproduced, not that redaction passed.

Required correction: validate malformed token input before constructing the request and/or replace exception interpolation with a stable secret-free diagnostic. Cover both tools, including malformed header values and transport exceptions, and assert absence of the synthetic token in captured stderr and exception text while preserving nearby valid API responses and failure exits. Do not log a request/header dump or rely on hosted log masking.

### R1-02 — Medium: Decision 0033 collides with an existing independent proposal

Locations: `decisions/0033-source-commit-and-release-tag-trust.md:1`, `GOVERNANCE.md:46`, `RELEASE.md:9`, `README.md:232`, `SECURITY.md:572`.

Requirement: the trust-policy amendment needs an unambiguous decision identity before landing. Open draft [PR200](https://github.com/relux-works/curator-spec/pull/200), at independently advertised/API-confirmed head `0d3bd4e8d7c33e6ba95de30673205936bf6c4e34`, already adds `decisions/0033-launch-lock-refresh-compatible.md`. It is a different design-only proposal, not this source-trust amendment. Both would be cited as Decision 0033. The repository uses four-digit decision prefixes with meaningful slugs; historical duplicate 0005 records exist, so this is not claimed to be an executable uniqueness-rule failure. It is the explicit coordination gate for this amendment.

Bounded disposition: rename this later source-trust decision to `decisions/0034-source-commit-and-release-tag-trust.md`, change its heading to Decision 0034, and update the four references above. Update the PR description/review-facing references when publishing the correction. Current main, PR200 and the other open PR file inventories do not occupy 0034. Recheck that allocation when publishing. Preserve PR200 and its semantics; do not renumber or merge it into this decision. No source or foreign PR was modified by this review.

## Complete swept surface

| Surface | Result and limits |
| --- | --- |
| Source-key authority | Candidate public material equals the independently read owner-designated public key and the fingerprint authorized in the current mandate. The actual landed object verifies with it. Authorization is not inferred from the producer adding a file |
| Closed source parser | Exactly one non-comment entry, literal `source-commit`, literal `namespaces="git"`, Ed25519 algorithm, strict base64 and SSH wire prefix/32-byte material. Duplicate, wildcard, wrong principal/namespace/type/material, invalid UTF-8 and absent/symlink/directory/FIFO inputs refuse; paired valid release-source signature still accepts |
| Existing release-signers as source | Original file remains a source route; candidate tests use newly generated source/release fixture keys. Actual old release keys are not possessed or used for signing. The unchanged release-file tests pin both original public entries |
| Every source route | `resolve_commit` then `show-ref`/`merge-base --is-ancestor` precede SSH or delegated GitHub acceptance. Unknown/unsigned/uncontained/missing-ref/nonexistent targets and ancestry query failures refuse. PGP route is independently exercised with real local object/header dispatch plus simulated API state |
| Ambient Git trust and selectors | All inherited `GIT_*`/`XDG_*`, `SSH_AUTH_SOCK` and `SSH_ASKPASS` are removed; deliberate no-system/no-global/no-replacement settings are installed. Fixed ROOT, SSH format, `ssh-keygen` and two repository trust files are selected. Real config/repository/index/object/trace poison and a replacement ref cannot admit an unknown signer |
| State preservation | Full synthetic decoy bytes, directory/file identity, mode, ownership, size, mtime, ctime and symlink target agree before/after. Caller environment agrees exactly; no poison trace or askpass sentinel appears. Atime is intentionally excluded because observation reads files |
| Trusted local boundary | The reviewed checkout, its repository metadata/refs and executable Git/SSH toolchain remain trusted. The tool does not independently authenticate a local origin ref or protect a same-user-modifiable checkout. Workflow fetch and repository custody supply that authority. Other inherited process state is not claimed to be hermetically sandboxed |
| Ref and API selectors | Full resolved SHA is checked before API URL construction; repository/main/API roots are constants. Context drift and removed CLI override flags refuse before use. `--end-of-options` prevents target-as-option evaluation; unresolved refs refuse |
| GitHub delegated fallback | Requires exactly one local PGP header, exact response SHA, GitHub committer name/email, boolean true verification, `valid` reason and nonempty string signature/payload. Unknown shapes/material types refuse. Local PGP header recognition is not local crypto verification; API fields are trusted delegated results, not independently compared/reverified signature bytes |
| GitHub authority and limits | GitHub's documented PGP verification ties committer email to key identity and a verified account email. Fixed exact-object API plus containment prevents treating arbitrary verified users as this route. It does not prove PR review, squash topology or protected-branch configuration. Persistent GitHub verification is not a current key-custody/revocation proof |
| HTTP transport/token | Fixed HTTPS authority and 30-second request timeout; redirects refused by the actual handler controls. Error/unknown results refuse. R1-01 remains open. The timeout is not represented as a total response-size or end-to-end wall-clock resource limit |
| Optional merge-settings observer | Fixed repo/API/context, exact boolean squash/rebase/merge policy checks and error refusal. Reads settings only. It is evidence for the optional squash route, not a protection audit or a command to rewrite authorized signed FF heads |
| Main CI | Existing `release-provenance` job and `Release target provenance` name preserved; push remains main-only, workflow_dispatch remains available, PR condition still skips provenance. Fresh origin/default-branch fetch precedes exact `GITHUB_SHA` check; non-main context/containment cannot qualify main |
| Release workflow | Tag trigger, annotated SSH tag verification using only original `maintainers.allowed_signers`, explicit peeled tag/candidate equality, fresh main fetch and source verification precede all original validation/regeneration/release gates. Quoted environment inputs and full target resolution remain. Source key is never passed to tag verifier |
| Real tag controls | Source fixture key fails tag verification with `No principal matched`; release fixture key succeeds for the same exact source commit. Peeled target equals that commit and differs from the neighboring candidate. OpenSSH `git` namespace itself does not distinguish tags/commits; consumer trust separation does |
| Packaging/consumer sweep | Both distinct trust files are included alongside the prior normative archive contents. Tracked and hidden-path search found no additional source-key tag consumer. No archive/build/full workflow was run |
| Unchanged release gate and tests | Full `release_gate.py`, `test_release_gate.py`, `test_allowed_signers.py` read. Version, clean exact checkout, immutable historical pins, platform/assurance limits, stable independent reviews, ancestry and frozen normative diff remain. The gate does not itself verify signatures; workflow invokes the separate provenance checks. These source files are unchanged; release-gate tests were not executed in this bounded unit |
| Governance/Release/Security/README/Decision | Explicitly amend old squash-only procedure; preserve independent review/checks, separate release endorsement, no historical compliance fiction. Tool docs describe inputs/commands/logs. Decision identity remains blocked by R1-02 |
| Unchanged authority boundaries | No key rotation/revocation, original tag-custody expansion, branch/protection/status mutation, package/platform-key change, infra-exit change, version bump or release claim. Review/check/native requirements remain separate gates |

Primary references for the delegated boundary: [GitHub commit API](https://docs.github.com/en/rest/commits/commits), [committer email verification](https://docs.github.com/en/authentication/troubleshooting-commit-signature-verification/using-a-verified-email-address-in-your-gpg-key), and [persistent signature verification](https://docs.github.com/en/authentication/managing-commit-signature-verification/about-commit-signature-verification). The inference above relies on GitHub enforcing its documented verification contract; it is not a local PGP proof.

## Independent executable and instrument receipts

Exact commands, selected child environments, direct PID lists and output are retained in the task evidence as `*-command.json`, `*-pids.txt`, `*-runner.pid`, logs and `cleanup.json`. The following command paths are abbreviated to avoid publishing machine paths.

| Command/control | Independent result |
| --- | --- |
| `python3 -B <evidence>/run_controls.py named` | Exit 0; exact candidate modules `test_verify_release_commit test_verify_release_merge_policy test_allowed_signers`; 30 tests, 0 skips, 0.770s |
| `python3 -B <evidence>/run_controls.py plants` | Exit 0; four intended checker assertion failures detected, no setup/import error accepted |
| Byte-identical base verifier | SHA-256 `423596a105e6dc2b12c44c9d804c2d4c48017cb16fc517cd23b838509f498ee7`; old containment control fails because authorized uncontained signature is accepted; old ambient control fails because ambient key is accepted |
| Reject-all/accept-all | Reject-all fails both nearby authorized-source/existing-signer positives; accept-all fails unknown/unsigned refusals. These are instrument sensitivity controls, not exhaustive mutation testing |
| `python3 -B <evidence>/run_controls.py extra` | Exit 0; six bounded positive/negative control groups plus the recorded R1-01 confidentiality failure witness |
| Actual landed signature | Read-only exact raw object imported without changing OID; source verification exit 0, original release verification exit 1 / `No principal matched`. No missing-ancestry success is claimed; this witness is signature-only |
| Local/public patch and whitespace | All 12 candidate blobs reconstructed from remote patches; `git diff --check` exit 0; final selector-free candidate checkout clean |
| Direct owned control PIDs | 399 unique recorded direct children, including failed setup probes, all absent at cleanup; no pattern kill or other-session process operation |
| New scratch | Sampled task peak 22,695,936 bytes, plus conservatively counted recent Git common metadata/objects 688,128 bytes: 23,384,064 bytes accounted, below 48 MiB STOP and 64 MiB hard budget. Sampling is 50ms, not a continuous filesystem quota claim. Controls also use a 1 MiB per-file OS limit and a 60-second runner bound |
| Cleanup | Synthetic keys, repos, homes, temp/cache roots removed. Decoy snapshots and caller-env invariants passed; retained worktree/evidence only. No private live key, socket or credential/config copy was a fixture input |

The outer test process uses an explicit allowlist and absolute disposable HOME/cwd/TMP/TEMP/XDG/cache/GOPATH. Inherited Git/XDG/socket/askpass/Go/native selectors are absent except deliberately constructed synthetic poison. Fixture children have disposable HOME/temp/cache and disabled global/system Git config; no Go executable is invoked. Product Git subprocess environments are inspected after their own stripping, not repaired by a wrapper.

Initial independent setup attempts remain red: the macOS `/usr/bin/git` shim terminated with SIGXFSZ under the 1 MiB and 8 MiB per-file limits. A direct installed Apple Git executable (`2.54.0`) passed isolated readiness, then the unchanged named tests ran with that explicit trusted toolchain path. No installation or profile repair occurred. The first extra-control attempt exited 2 because an option-looking target was rejected by argparse; the witness was corrected to expect that refusal and retained as `extra-argparse-red.*`. Producer atime/import red receipts remain untouched and are not promoted into passes. Producer claims of 389 PIDs, 30 tests and 22.6 MB remain separate from the independent values above.

## Trust pins and remaining gates

Original `maintainers.allowed_signers` is byte-identical at base and candidate: SHA-256 `12f4bfaf1bfcd954d30b0683f6b55b8ad7e19e5f60c88c95dbf0af702db82dc8`, blob `47f54d69ce219fa871602d77c554282c7d2c6207`. Candidate `source.allowed_signers` blob is `8e3976ba08be6a9c0386193f32cf73aa00b3b397`. Complete affected and adjacent blob/hash inventory is retained in `source-pins.json`.

One hosted snapshot at exact candidate only: Specification jobs on Ubuntu/macOS/Windows IN_PROGRESS; Implementations on all three SUCCESS; Formatting and Links SUCCESS; Release target provenance SKIPPED. The GitHub connector omitted verification/check-run fields needed for these bounded reads, so selected read-only `gh` API/PR projections supplied them. No polling, waiting, rerun, cancellation or artifact/log download occurred.

Historical main run `40569031707`, job `122417945263`, remains red. Its original failure is the trusted-source policy defect, not an external CI failure. The actual base GitHub API independently reports SSH signature verification `true/valid`, consistent with the local crypto witness; that does not make the old provenance run pass.

No full discovery, Go/product/build/module/package/race/repro/make/validate/regenerate/workflow/full gate/broad lint/crosscompile was run locally. Darwin Python 3.15.1 / Apple Git 2.54.0 focused results do not qualify the hosted Python 3.12 multi-platform jobs. Following a corrected candidate, independent exact-head review and actual required hosted checks remain necessary; only after separately authorized accepted landing can a NEW actual-main provenance run legitimately establish recovery.

```json
{
  "verdict": "CHANGES_REQUESTED",
  "head": "ec70b25b5c7b7b23be64a05dacd6534eee71ade3",
  "base": "74a7dcc503c0fd7acd491299ab71290ae40e708e",
  "current_main": "74a7dcc503c0fd7acd491299ab71290ae40e708e",
  "findings": [
    {"id":"R1-01","severity":"medium","status":"open","requirement":"Secret-free API refusal diagnostics in both tools","effect":"Malformed bearer token is reflected in failure text","repro":"Synthetic newline-bearing token; real HTTP header validation; connect forbidden"},
    {"id":"R1-02","severity":"medium","status":"open","requirement":"Unambiguous source-trust decision identity before landing","effect":"Decision 0033 collides with existing PR200","disposition":"Rename this decision to 0034 and update four references; preserve PR200"}
  ],
  "source_trust": {
    "fingerprint":"SHA256:rE/sRlfdFt/vT5y5yukDCsq4aFtC48pbWsq9z4hm8MA",
    "authority":"Existing owner mandate plus independently matched public key and landed signature",
    "scope":"Commit-only alongside existing release signers; all accepted routes require main containment",
    "adopted":false
  },
  "release_trust_unchanged":true,
  "native_or_hosted_gates": {
    "focused_tests":"30 pass, zero skips",
    "specification_jobs":"in progress at the sole snapshot",
    "pr_provenance":"skipped; not qualification",
    "actual_main_recovery":"not established; historical failure remains red",
    "native_qualification":"not performed"
  },
  "next_checkpoint":"Producer corrects R1-01/R1-02; independent review of the new exact head and real required hosted checks before any separately authorized landing. New actual-main provenance remains a later recovery gate."
}
```

Publication receipt: [exact-head COMMENT review](https://github.com/relux-works/curator-spec/pull/202#pullrequestreview-5673704818), state `COMMENTED`, head `ec70b25b5c7b7b23be64a05dacd6534eee71ade3`; separate API readback matched all 19363 body bytes (SHA-256 `24576ca9941760d901e7f1147b3506eb5a3145df43178162db3f4cd355dfeb5e`). This is a COMMENT carrying CHANGES_REQUESTED, not a platform approval. The connector returned 403 for review creation; the existing authenticated CLI published the same body without any authentication change.
