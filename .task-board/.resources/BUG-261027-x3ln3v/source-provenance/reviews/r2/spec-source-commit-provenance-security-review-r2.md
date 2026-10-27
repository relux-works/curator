# Source commit provenance — independent security rereview r2

**ACCEPT_CODE_PENDING_HOSTED_GATES** for exact head `3d5f838717e0609fa8905c37b13b117c8d974a00` in [curator-spec PR202](https://github.com/relux-works/curator-spec/pull/202). Both medium findings from [r1](https://github.com/relux-works/curator-spec/pull/202#pullrequestreview-5673704818) are closed on this successor. No new finding remains in the complete bounded sweep below. This verdict accepts the reviewed code and policy text; required hosted checks, separately authorized landing and a new actual-main provenance result remain distinct gates.

| Source/publication boundary | Independent result |
| --- | --- |
| Exact head | `3d5f838717e0609fa8905c37b13b117c8d974a00` |
| Base/current main | `74a7dcc503c0fd7acd491299ab71290ae40e708e`; initial and final remote advertisements agree |
| Append-only history | Original reviewed `ec70b25b5c7b7b23be64a05dacd6534eee71ade3`, correction `41a6121151eea2cd4229f78fb2b4377f28370ef1`, then this head; all three form a direct chain on current main |
| Canonical authority | Repository/default/feature/PR API and advertised/fetched refs agree: `relux-works/curator-spec`, `main`, `fix/source-commit-provenance`. Final PR API still names this exact head/base |
| Signatures and attribution | All three commits are cryptographically `G`, with the configured human author/committer and authorized fingerprint `SHA256:rE/sRlfdFt/vT5y5yukDCsq4aFtC48pbWsq9z4hm8MA`. Each has the exact sole `Co-authored-by: Codex gpt-6.1-sol <noreply@openai.com>` trailer; no session metadata |
| Full published parity | All 12 remote patches independently reconstruct byte-identical head blobs from exact base: 768 additions, 97 deletions. The complete corrective delta from r1 is 9 paths, 191 additions, 10 deletions. No source edit was made during review |
| Isolation | Exact detached review checkout is clean. Private Git directory, explicit work tree and command-level `core.worktree` were bound before `read-tree`/`checkout-index` materialization. Stable HEAD, branch, index, diffs and 5,440 prior file hashes agree; earlier worktrees/reports/evidence are preserved |
| Reviewer context | Selected recorded runtime fields match the requested current managed profile, `gpt-6-astra` and `xhigh`, resuming only the original reviewer thread. No whole native configuration/argv inspection, fallback, profile repair or worker launch |
| Input integrity | Full new producer handover: 13,099 bytes, SHA-256 `fec33f78e14d77c7a070a9dc20b761e3fad8a102190c34756517de485777aeee`. Full prior independent report: 19,903 bytes, SHA-256 `b6a96e80c65be5eaa9386d5a45b51e15cb939a8e0bd80f3a376e54a319a410f2`. Original review brief and complete published corrective reviews were read. Producer statements remain inputs, not acceptance |

## Finding dispositions

**R1-01 — Medium, CLOSED: credential-bearing API errors.** Both `fetch_github_payload` and `fetch_repository_payload` now discard `OSError`, `URLError`, `http.client.HTTPException` and `ValueError` details, then raise their fixed failure outside the exception handler. This removes raw exception interpolation and the implicit exception context. The public entrypoints retain failure status 1 and stable prefixed stderr. Valid fixed-endpoint responses still succeed.

Independent reproduction used newly generated synthetic token fragments held only in memory. Real `http.client` header validation refused newline-bearing headers before connect for both helpers and both entrypoints. Each helper produced its exact documented failure class/message with `__cause__` and `__context__` absent; `str`, `repr`, formatted traceback and public stderr contained no generated fragment. No external request occurred. Additional independent pairs covered OS, URL, value and HTTP status-line errors, including malformed status-line parsing through the real HTTP response parser over an in-memory socket. Candidate tests also exercise real valid HTTP parsing and 301/302/303/307/308 redirect refusals at the fixed authority, with no second connection to a redirect destination.

The detector was checked against byte-identical production files from `ec70b25`: commit verifier SHA-256 `fafabd6be4af83d6a98dccf4c2280600f91c40338144cb32d79a327e60286fb0`, settings verifier SHA-256 `a34bbe7ce41c1aa99e8c655ac18920fb754e628e9dcd02efefbf9541655f0267`. Four test methods reported the intended confidentiality violations: ten helper-diagnostic and ten public-stderr assertions, zero setup/import errors. The old verdict remains failing evidence; the enclosing detector's exit 0 means it detected those failures. Reject-all substitutions in each helper and each entrypoint failed the nearby valid-response control. Helper substitutions surface as the expected failure exception at the positive fetch assertion; they are not counted as successful product tests or unrelated setup errors.

The accepted claim covers these standard transport/header/protocol/JSON error families. It does not promise to sanitize arbitrary exceptions from hostile replacement libraries or local programming defects, nor does the 30-second request timeout establish a total response-size or end-to-end wall-clock bound. No real token, credential copy, request dump or authenticated mutation was used by fixtures.

**R1-02 — Medium, CLOSED: decision number collision.** The source-trust amendment is now `decisions/0034-source-commit-and-release-tag-trust.md`, headed Decision 0034. Its four references in `GOVERNANCE.md:46`, `RELEASE.md:9`, `README.md:232` and `SECURITY.md:572` are updated, as is the actual PR description. Exact-byte comparison shows only heading/number/link changes in these five policy documents; no source-trust semantics were merged or altered.

Current main has no Decision 0034. Current open-PR inventories for PR203, PR200, PR198, PR119 and PR89 contain none. [PR200](https://github.com/relux-works/curator-spec/pull/200) remains at `0d3bd4e8d7c33e6ba95de30673205936bf6c4e34`, owning `decisions/0033-launch-lock-refresh-compatible.md` with its separate design-only semantics. Four-digit prefixes plus descriptive slugs are the existing convention; historical duplicate 0005 entries do not turn this into an executable uniqueness-rule claim. Historical r1 references remain historical. No foreign PR was altered. Allocation is verified at this source/open-PR snapshot and must remain consistent when landing is separately considered.

## Complete swept surface

Unchanged r1 controls are reused only where exact input pins and production logic still agree. A fresh AST comparison proves all production nodes outside the two fetch helpers and the added `http.client` imports are unchanged from the reviewed predecessor. Both current production files and the complete corrective diff were read. The named suite was rerun at the new exact head.

| Surface | Verdict, evidence and limits |
| --- | --- |
| Source public-key authority | Accepted code boundary. The unchanged public material is the key independently matched in r1 to the existing owner mandate and actual landed signature. Authority is not inferred from the producer adding a trust file. The parser reads a reviewed repository trust input; it does not independently discover ownership or hardcode the fingerprint |
| Closed source-key parser | Exactly one non-comment entry, literal `source-commit`, literal `namespaces="git"`, Ed25519, strict base64 and exact SSH wire-prefix/32-byte material. Unknown, duplicate, malformed, wrong principal/namespace/type/material, invalid UTF-8 and missing/symlink/directory/FIFO trust inputs refuse. R1 nonregular/material controls are reused against unchanged logic; current named positives/refusals pass |
| Existing maintainer source route | Preserved alongside the source-only key. Fresh synthetic source and release signer positives pass; actual original release private keys are neither available nor used. Original maintainer-file tests pin the existing public entries |
| Containment before every success | Full commit resolution and fixed `origin/main` existence/ancestry checks precede SSH or delegated GitHub success. Unknown/unsigned/uncontained/nonexistent/wrong targets, missing ref and ancestry-query errors refuse. Current named checks plus unchanged r1 PGP-dispatch/containment controls cover both routes |
| Ambient Git trust/config/selectors | All inherited `GIT_*`/`XDG_*`, `SSH_AUTH_SOCK` and `SSH_ASKPASS` removed by the verifier itself; deliberate no-system/no-global/no-replacement settings follow. Fixed ROOT, SSH format, `ssh-keygen` and repository trust paths remain. Current ambient-trust tests pass; unchanged r1 full config/repository/index/object/trace poison and replacement-ref controls remain applicable |
| Local trusted assumptions | Checkout/policy files, repository metadata and refs, PATH-resolved installed Git/SSH executables remain trusted. This is not an arbitrary same-user attacker sandbox. Workflow fetch supplies fresh canonical ref authority. The source tool alone does not authenticate a locally supplied origin ref or attest review/protection settings |
| State preservation | Current test caller environments agree before/after. Synthetic roots/keys are removed. Stronger r1 full decoy byte, identity/mode/ownership/size/mtime/ctime and symlink-target snapshots remain valid for the unchanged Git path; no trace/askpass sentinel appeared. Atime equality was intentionally excluded because observation reads files |
| Ref/API selector injection | Resolved full SHA before URL construction; fixed repository, main and HTTPS API roots. Context mismatches and removed override flags refuse; `--end-of-options` prevents option evaluation. Current context/URL cases pass; unchanged r1 option/replacement/ref/ancestry cases are reused |
| GitHub PGP dispatch | Requires one actual local PGP signature header, exact response SHA, GitHub committer name/email, boolean true verification, `valid` reason, nonempty string signature/payload. Wrong identity, material types, unknown/error results refuse. Local marker recognition and API simulations are not local cryptographic PGP verification |
| Delegated GitHub authority | Fixed exact-object API plus contained target and GitHub committer constrains this route. GitHub's documented PGP verification associates committer email with the key and verified account email. The local tool delegates this result; it does not compare/reverify returned signature/payload bytes, prove squash topology/PR review or audit branch protection. Persistent verification is not proof of current key custody/revocation |
| HTTP request/error/token boundary | R1-01 closed. Fixed API authority, no caller API root, request timeout 30 seconds, actual redirect handler refuses. Current real header/parser/valid-response/redirect checks and independent error-family pairs pass. Failures contain fixed diagnostics without cause/context. Known transport-family scope and resource-limit caveat stated above |
| Optional merge-settings observer | Fixed repository/context/API and strict boolean squash/rebase/merge settings remain; errors refuse confidentially. This tool observes settings only and does not change them, audit protection or require rewriting authorized signed fast-forward heads |
| Main CI | Byte-identical to r1. Same `release-provenance` job/`Release target provenance` name, main push trigger and manual dispatch. Fresh origin/default-main fetch precedes exact `GITHUB_SHA` verification. PR condition still skips provenance; this skip cannot qualify main or establish recovery |
| Release workflow | Byte-identical to r1. Annotated SSH tag verification uses ONLY original `maintainers.allowed_signers`; peeled tag and candidate must agree. Fresh main fetch and source verification precede prior release gates. Quoted inputs/full resolution and all prior validation/regeneration steps remain |
| Tag-versus-commit authority | Fresh real synthetic paired controls: source fixture key fails tag verification; release fixture key succeeds for the same endorsed source commit, with exact peeled-target agreement and a different neighboring target. OpenSSH's `git` namespace covers both object kinds; the separate consumer trust roots enforce source-only authority |
| Packaging and every consumer | Complete tracked/hidden-path r1 consumer sweep retained; all affected consumer bytes are unchanged. Both distinct trust inputs are packaged with prior normative contents. Source trust is never supplied to tag verification. No package, archive, workflow or full-gate execution is claimed |
| Unchanged release gate/tests | Full `release_gate.py`, `test_release_gate.py` and `test_allowed_signers.py` were read in r1; fresh byte pins match. Version/clean exact checkout, historical pins, independent stable reviews, ancestry, frozen normative diff and platform/assurance limits remain. The release gate does not itself verify signatures; workflow invokes provenance separately. Release-gate tests were not run in this bounded review |
| Governance/Release/Security/README/Decision | Old squash-only source procedure is explicitly amended to allow independently reviewed signed exact-head fast-forward, retaining GitHub's signed squash alternative. Review/checks and separate release endorsement remain external gates. Decision 0034 resolves coordination without changing policy substance. No retrospective compliance claim |
| Promised versus executable behavior | Crypto/trust/containment and fixed API checks are executable. Human source-key authorization, reviewed checkout/toolchain custody, independent review, required hosted checks and protection/landing compliance require their respective governance mechanisms. Neither a signature pass nor optional settings observation alone proves those conditions |
| Unchanged authority limits | No key revocation/rotation, release-custody expansion, branch/protection/status mutation, platform-key or infra-exit change, version bump, native qualification, release, adoption or landing action |

Delegated-verification references retained from r1: [GitHub commit API](https://docs.github.com/en/rest/commits/commits), [committer email verification](https://docs.github.com/en/authentication/troubleshooting-commit-signature-verification/using-a-verified-email-address-in-your-gpg-key), [persistent signature verification](https://docs.github.com/en/authentication/managing-commit-signature-verification/about-commit-signature-verification). The delegated conclusion depends on GitHub enforcing that contract; no API simulation is credited as local PGP proof.

## Independent commands, controls and preservation

Exact argv, explicit child environments, direct PID lists, safe logs and result JSON are retained in private review evidence. Paths below are intentionally abbreviated. No synthetic token fragment is persisted; failing checker details are classified in memory before safe summaries are written.

| Command/control | Result |
| --- | --- |
| `python3 -B <evidence>/run_controls.py named` | Exit 0: exactly `test_verify_release_commit test_verify_release_merge_policy test_allowed_signers`, 36 tests, 0 failures, 0 errors, 0 skips |
| `python3 -B <evidence>/run_controls.py old` | Exit 0: four methods detect intended old-byte confidentiality failures, ten helper and ten stderr disclosures; zero unrelated errors |
| `python3 -B <evidence>/run_controls.py reject-all` | Exit 0: all four helper/entrypoint plants detected by nearby valid fixed-API responses. No setup/import failure counted as detection |
| `python3 -B <evidence>/run_controls.py audit` | Exit 0: 12 independent tool/fault pairs; stable helper types/messages, absent cause/context, secret-free str/repr/formatted trace, exact public status/stderr, no external requests. Real malformed-header and bad-status-line parsing exercised |
| `python3 -B <evidence>/verify_closure.py` | Exit 0: unchanged AST boundaries, policy-only renumbering, current decision allocation, prior 5,440 file hashes/index and clean exact-head checkout; `git diff --check` passes |
| R1 containment/ambient detector reuse | Exact main verifier SHA-256 `423596a105e6dc2b12c44c9d804c2d4c48017cb16fc517cd23b838509f498ee7` failed intended uncontained-signature and ambient-key assertions, not import/setup. A reject-all plant failed authorized-source/existing-signer positives; accept-all failed unknown/unsigned negatives. These four historical controls apply to unchanged production logic and were not rerun |
| R1 adjacent controls reuse | Six pinned groups cover nonregular/malformed trust, full ambient/decoy preservation, replacements/ref/query errors, actual landed signature source acceptance versus release rejection, PGP header/containment dispatch, and real redirect refusal. Actual landed object control is signature-only; it does not waive ancestry |
| Direct control PIDs | 87 unique fresh direct control/runner PIDs recorded; all absent at cleanup. This differs from r1's 399 and producer counts. No other-session process control or pattern kill |
| Scratch accounting | 50ms sampled task peak 23,457,792 bytes, plus conservative recent common-Git allocation 512,000 bytes: 23,969,792 bytes at control cleanup, below 48 MiB STOP/64 MiB hard budget. Small final report/publication/board receipts are additionally checked against the same budget. Sampling is not a continuous filesystem quota claim |
| Fixture limits/cleanup | Explicit allowlisted outer environment and disposable absolute HOME/cwd/TMP/TEMP/XDG/cache/GOPATH, 1 MiB per-file limit and 60-second runner bound. Fixture children have isolated config/home/temp/cache. Synthetic SSH keys generated internally and removed with owned roots. No live private key, agent socket, repository/index/home or credential/config copy used as fixture input |
| Preserved red evidence | R1 setup/import/argument and producer intermediate reds remain unchanged. An r2 signature-inspection probe initially rejected legitimate prose because an overbroad metadata pattern matched `HTTP`; it failed before materialization, was retained as red inspection evidence, and the precise session/URL check then passed. No product failure was relabeled a pass |

No full unittest discovery, Go/product/build/module/package/race/repro/make/validate/regenerate/workflow/full gate/broad lint/crosscompile was run. Focused Darwin Python 3.15.1 / Apple Git 2.54.0 results do not substitute for hosted Python 3.12 multi-platform checks. The read-only supplied actual public-key/landed-object evidence is inherited from pinned r1 controls, not used to create a real signature or release.

## Trust pins and remaining checkpoint

| Input | Exact unchanged scope/pin |
| --- | --- |
| Original release trust | `maintainers.allowed_signers`, SHA-256 `12f4bfaf1bfcd954d30b0683f6b55b8ad7e19e5f60c88c95dbf0af702db82dc8`, blob `47f54d69ce219fa871602d77c554282c7d2c6207`; unchanged from base and r1 |
| Commit-only source trust | `source.allowed_signers`, SHA-256 `a757c518fb974409858ead6c177bb332f08877d13e96420196eaf045d0f9927f`, blob `8e3976ba08be6a9c0386193f32cf73aa00b3b397`; unchanged from r1 |
| CI/release workflows | SHA-256 `e8e3a8d1133a0ac6a8a1110bec77264c76659e64a3dd632be9d495acf161b0c7` / `260c8e03830c10d838b203157aa8e72d8f5b611e59c893e33f12caabc4496bbf`; unchanged from r1 |
| Adjacent gate/tests | `release_gate.py` SHA-256 `b6342d54a1f80e68f93206b4e97690157a999bc6eb37a43d6e4c52ef03a80dcc`; `test_release_gate.py` `dfc6e27ec06bce182efee4e7ca709271ad3821777c18846e43edceae1b2170ab`; `test_allowed_signers.py` `5ab5484447c3aee4974c2aaf63a07ca0f12d42eb14260251a5dab1cf4e8415d5`; unchanged from r1 |

One exact-head hosted snapshot only: Specification Ubuntu SUCCESS; Specification macOS and Windows IN_PROGRESS; Implementations on Ubuntu/macOS/Windows SUCCESS; Formatting and Links SUCCESS; Release target provenance SKIPPED. No wait, poll, rerun, cancel, collector, log/artifact download or local broad-suite fallback occurred. Required checks remain actual hosted gates, not inferred from the review verdict.

Historical main run `40569031707`, job `122417945263`, remains genuinely failed. Its cryptographically valid authorized signature and missing source trust are not reclassified as a passing historical check. After a separately accepted and authorized landing, a NEW actual-main provenance run must legitimately pass before recovery is claimed. Current PR provenance SKIP provides no recovery evidence.

```json
{
  "verdict": "ACCEPT_CODE_PENDING_HOSTED_GATES",
  "head": "3d5f838717e0609fa8905c37b13b117c8d974a00",
  "base": "74a7dcc503c0fd7acd491299ab71290ae40e708e",
  "current_main": "74a7dcc503c0fd7acd491299ab71290ae40e708e",
  "findings": [
    {"id":"R1-01","severity":"medium","status":"closed","requirement":"Secret-free helper and public API refusals","evidence":"Real pre-connect malformed-header refusal; error/parser controls; exact old-byte detector failures; nearby valid-response controls"},
    {"id":"R1-02","severity":"medium","status":"closed","requirement":"Unambiguous source-trust decision identity","evidence":"Decision0034, heading/four links/PR description complete; current main and other open PRs unoccupied; PR200 unchanged"}
  ],
  "source_trust": {
    "fingerprint":"SHA256:rE/sRlfdFt/vT5y5yukDCsq4aFtC48pbWsq9z4hm8MA",
    "authority":"Existing owner mandate; independently matched public key and landed signature from pinned r1 evidence",
    "scope":"Commit-only alongside existing maintainer source route; every success requires main containment",
    "adopted":false
  },
  "release_trust_unchanged":true,
  "native_or_hosted_gates": {
    "focused_tests":"36 pass, zero skips",
    "specification_jobs":"Ubuntu success; macOS and Windows in progress at sole snapshot",
    "pr_provenance":"skipped; not qualification",
    "actual_main_recovery":"not established; historical failure remains red",
    "native_qualification":"not performed"
  },
  "next_checkpoint":"Real required hosted checks at this exact head, then separately authorized landing with fresh source/decision-allocation checks. A new actual-main provenance pass is required before recovery; no task closure, trust adoption or release is granted by this review."
}
```

This immutable report is the body for an exact-head COMMENT review, not a GitHub approval. Publication ID, body byte/hash readback, distinct outcome readback and append-only coordination receipts are recorded separately after publication. Previous review and producer outcomes remain intact.
