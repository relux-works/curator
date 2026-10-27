# Source commit provenance — independent security rereview r3

**CHANGES_REQUESTED** at exact [curator-spec PR202](https://github.com/relux-works/curator-spec/pull/202) head `357f2724c9ecc22db210dfa6d43abdb451ccfa44`. The Windows environment correction is supported within the stated source-mapping limits, and all 39 permitted tests pass. One medium finding remains in the changed fixture helper's confidentiality boundary: expected launch/timeout exceptions retain sensitive material. The complete changed and cumulative surface was swept; this is not a first-hit review. No landing, source-trust adoption, release or native qualification is granted.

| Source boundary | Independently established result |
| --- | --- |
| Candidate | `357f2724c9ecc22db210dfa6d43abdb451ccfa44` |
| Accepted predecessor | `3d5f838717e0609fa8905c37b13b117c8d974a00`; [r2 acceptance](https://github.com/relux-works/curator-spec/pull/202#pullrequestreview-5674261677) remains historical and exact-head scoped |
| Base/current main | `74a7dcc503c0fd7acd491299ab71290ae40e708e`; initial/final advertisements, fetched ref and PR base agree |
| Canonical repository/default/feature | `relux-works/curator-spec`, `main`, `fix/source-commit-provenance`; remote/default API, PR API, advertised/fetched refs and detached review HEAD agree |
| Signed history | Four cumulative human-authored/committed objects: `ec70b25b5c7b7b23be64a05dacd6534eee71ade3`, `41a6121151eea2cd4229f78fb2b4377f28370ef1`, `3d5f838717e0609fa8905c37b13b117c8d974a00`, `357f2724c9ecc22db210dfa6d43abdb451ccfa44`. All verify `G`, form an append-only chain on current main, and carry the exact sole `Co-authored-by: Codex gpt-6.1-sol <noreply@openai.com>` trailer without session metadata |
| Signature fingerprint | All four use `SHA256:rE/sRlfdFt/vT5y5yukDCsq4aFtC48pbWsq9z4hm8MA`, already independently tied to the owner mandate/public material in r1 |
| Complete publication parity | All 12 remote patches reconstruct byte-identical exact-head blobs from base: 881 additions, 97 deletions. Successor correction is only `tools/test_verify_release_commit.py`, 131 additions/18 deletions. No source edit during review |
| Review isolation | Exact detached checkout, private Git directory/work tree and command `core.worktree` bound before `read-tree`/`checkout-index`. Stable HEAD/branch/status/index/diffs and 8,222 previous file hashes agree; prior reports/evidence/worktrees remain intact. Review checkout clean |
| Requested reviewer context | Selected recorded fields match current managed profile, `gpt-6-astra`, `xhigh` and the original reviewer resume. No whole native config/argv inspection, fallback, profile repair, new worker or host operation |
| Inputs | Full producer handover 14,246 bytes, SHA-256 `828a4ee0ca5b5495a443ecab921907de15acda9cf964e2a7b559ff240d64bb31`; full r2 report 22,514 bytes, SHA-256 `d315d40dec039ad044da1377485f7750d532ccbbe6e22fbeb70a86b261159270`; original r1/r2 briefs and exact-head producer COMMENT 5674595936 read. Producer claims are inputs, not acceptance |

## Finding R3-01 — Medium: fixture failure sanitization omits exception paths

**Location:** `tools/test_verify_release_commit.py:152-162`, `SourceTrustTests.run_child`.

**Requirement:** failure diagnostics must contain only safe tool/status/stream-length metadata, omitting raw argv, captured output and sensitive OS-error material, including retained exception context. Keep the 10-second child bound and exact-PID kill/reap behavior, and pair refusals with valid real subprocess controls.

**Observed behavior:** nonzero-exit messages are sanitized, but two expected error paths fall outside that boundary:

1. `communicate(..., timeout=10)` raises `TimeoutExpired`; the existing cleanup branch kills/reaps the direct child and rethrows the original exception. Its ordinary string/formatted traceback includes full command argv, and its `output`/`stderr` attributes retain captured partial streams. The new status/length-only failure code is never reached.
2. The new `except OSError` branch raises its sanitized `AssertionError` inside the handler using `from None`. Ordinary formatted traceback suppresses the original error, but `__context__` still references that raw `OSError`. Suppression is not removal. The current unlaunchable-tool test checks the outer message only, so it misses this retained material.

**Independent reproduction:** the exact current helper launched a real disposable Python child with a generated synthetic argument, echoed a separately generated synthetic stdin fragment to both streams, flushed and slept. After 10.008 seconds, the observed `TimeoutExpired` reflected the argument in both `str` and formatted traceback and retained both output fragments. The exact child was killed and reaped. A separate controlled `PermissionError` carrying an in-memory fragment produced a safe outer message and ordinary traceback, while its raw fragment remained reachable through `__context__`. No fragment, raw request, key, argv or failing traceback was persisted; receipts contain only classifications and booleans. No live credential or external connection was involved.

**Effect and scope:** timeout diagnostics can disclose command inputs, and consumers inspecting exception attributes/context can recover raw sensitive error/output data. This is a fixture error-path confidentiality issue, not a bypass of production source trust or tag authority. The timeout rethrow predates this successor, but lies directly in the helper now promising safe failure receipts. The launch-context retention is in the new branch. No actual historical CI credential disclosure is claimed.

**Nearby valid controls:** real child exit 0 preserves expected stdout; silent exit 9 retains status/count diagnostics; exit 7 with sensitive text on both streams produces safe metadata; missing tools do not dispatch; ordinary unlaunchable-tool messages are stable. All 39 committed tests pass, including real synthetic SSH operations. Those passes do not cover the two properties above.

**Required bounded correction:** after exact-PID cleanup, normalize expected launch and communication-timeout failures into stable safe diagnostics raised outside the original handlers. Do not retain raw exception context, argv or captured stream attributes. Preserve interruption semantics and existing child bounds; this does not require catching arbitrary hostile-library exceptions or redesigning unrelated subprocess code. Add focused checks for both timeout diagnostics/attributes and launch `__cause__`/`__context__`, with real valid-child and cleanup controls. The resulting new signed head needs independent rereview and actual required hosted checks.

| Earlier finding | Current disposition |
| --- | --- |
| R1-01, API token/error confidentiality | CLOSED and preserved: both production helpers and all r2 transport tests are byte/AST unchanged. Current named suite passes; pinned r2 real-header/parser/error/redirect/valid-response and old-byte/reject-all controls remain applicable |
| R1-02, Decision 0033 collision | CLOSED and preserved: Decision 0034, heading/four references and PR description remain correct; fresh main/open-PR inventory finds no other 0034. PR200 remains unchanged at `0d3bd4e8d7c33e6ba95de30673205936bf6c4e34`, owning its separate Decision 0033 |

## Actual Windows evidence and primary-source limits

The complete supplied official failed-step output is 8,941 bytes, SHA-256 `8abd43a664bfe9a8ce5828ee280d13a515be528ec1b34a33866a670c38117d40`. [Prior Windows job 122457223176](https://github.com/relux-works/curator-spec/actions/runs/40580017534/job/122457223176) failed in `SourceTrustTests.setUpClass` during `ssh-keygen`; captured stderr was empty. The selected keygen executable/version, its precise exit code and stdout are **unknown**. The overall step's exit 1 is not the keygen's independently recorded exit code. This was an introduced fixture failure, not evidence of an external outage.

Primary Microsoft files were independently read at immutable commit `274893f8d1fe86c4d128b514c3af4bc10cdc0c99`, downloaded within a small bounded read, hashed and matched against the producer's supplied bytes:

| Primary file | SHA-256 | Supported conclusion |
| --- | --- | --- |
| [wmain_common.c](https://github.com/PowerShell/openssh-portable/blob/274893f8d1fe86c4d128b514c3af4bc10cdc0c99/contrib/win32/win32compat/wmain_common.c#L40) | `248603baffe3e0403164176997a11f49c72fc66d9c1690df4f8c0ce584aab66f` | Converts existing wide argv entries to UTF-8, supplies a default agent pipe only when `SSH_AUTH_SOCK` is absent, and initializes compatibility state before keygen main |
| [w32fd.c](https://github.com/PowerShell/openssh-portable/blob/274893f8d1fe86c4d128b514c3af4bc10cdc0c99/contrib/win32/win32compat/w32fd.c#L195) | `a964fb8994b4b778fe51db359fa72ba04e4c0cead122156c21f110958f899186` | `init_prog_paths` requires `ProgramData` and fails when absent; this precedes descriptor initialization. It supports the proposed missing-variable failure class, not identification of the historical binary |
| [ssh-keygen.c](https://github.com/PowerShell/openssh-portable/blob/274893f8d1fe86c4d128b514c3af4bc10cdc0c99/ssh-keygen.c#L2716) | `895098057c6360f533eb307c1a1552ddb06005d1647c7ff42c34e382bbcab716` | Signing can probe the authentication socket before private-file fallback. The reviewed verification path verifies signed material/trust without that signing-agent lookup |

The source-mapped detector is deliberately narrower than native execution: old setup reaches two real Git initializations, then fails the independently imposed missing-`ProgramData` assertion at the keygen boundary. It does not execute an old Windows binary. The corrected setup supplies an owned directory and passes that same mapped requirement while creating three real keys and exercising real Darwin signatures. No dropping-field mock or source mapping is counted as Windows success or the exact historical root cause. New actual Windows CI remains necessary.

## Complete changed-fixture sweep

| Surface | Result and boundary |
| --- | --- |
| Environment construction | Explicit dictionary; no broad caller-environment copy. HOME, temp, cache, GOPATH and ProgramData are absolute owned fixture paths at all actual callers. ProgramData directory is created before keygen. Live machine/user profile configuration is not copied |
| Excluded selectors | Caller Git/XDG/socket/askpass/GOFLAGS/native-home/user-profile entries are discarded before intentional fixture values. Constant no-system/no-global Git policy and synthetic author/committer follow. Current environment-poison test passes; dictionary construction also excludes unspecified inherited selectors |
| SYSTEMROOT | The sole loader field is preserved when present. Python documents that Windows side-by-side assembly execution may require a valid SystemRoot. This is a loader dependency, not a general home/profile inheritance grant. It was already forwarded before this correction. [Python subprocess environment contract](https://docs.python.org/3.12/library/subprocess.html#subprocess.Popen) |
| PATH and executable roots | `shutil.which` uses the fixture PATH, then Popen receives the resolved executable and an argv sequence. Installed executable directories and the resolver's context remain trusted. The diagnostic `windows-openssh` label is inferred from directory name; it is not binary/version attestation |
| Windows PATHEXT/current-directory limits | The child dictionary does not forward PATHEXT, but `shutil.which` itself can consult the calling process's PATHEXT and Windows current-directory policy. Thus this is not hermetic executable selection under hostile loader state. Trusted CI tool paths/resolver context remain assumptions; exact historical selection remains unknown. [Python which contract](https://docs.python.org/3.12/library/shutil.html#shutil.which) |
| No shell/empty passphrase | All observed launches use sequences and no `shell=True`; three actual keygens receive a distinct empty `-N` argument. For trusted native executables this avoids a shell-string workaround. Windows batch files can still be interpreted by the OS despite shell=False, so no arbitrary PATH/PATHEXT replacement-script guarantee is claimed. [Python subprocess security notes](https://docs.python.org/3.12/library/subprocess.html#security-considerations) |
| Native default agent | Synthetic signing receives an owned nonexistent `SSH_AUTH_SOCK`, preventing the pinned native startup path from inventing the default agent pipe. All three observed keygen environments carry it; Git signing uses the same fixture environment. Product verification still strips socket variables and performs verification only. Signing and verification are not assumed equivalent |
| Missing/unlaunchable tools | Missing tool gives explicit fixed refusal without Popen. Launch OSError yields safe outer class/name metadata, but retained context is R3-01. No fallback/retry/install or platform skip |
| Nonzero output diagnostics | Exact numeric status and decoded stdout/stderr CHARACTER counts, UTF-8 replacement decoding, no raw streams in the ordinary nonzero-exit message. Real sensitive-output exit 7, silent exit 9 and valid exit 0 controls pass. Output-bearing CompletedProcess values requested with `check=False` are internal test data, not a sanitized diagnostic claim |
| Timeout and cleanup | Existing 10-second communicate bound and kill/wait(5) on the exact direct child remain. Real timeout control reached that boundary and reaped its child; raw exception propagation is R3-01. This is not a whole-process-tree or uninterrupted wall-clock guarantee |
| Real cryptography retained | Three newly generated source/release/unknown Ed25519 keys, real signed source/release/unknown objects, unsigned and uncontained negatives, real annotated tags and paired source-key rejection/release-key success. No canned keys, crypto mocks, reduced assertions, skip or platform exclusion |
| Existing tests/snapshots | AST comparison proves every pre-existing test assertion unchanged; only setup/run_child changed, with fixture_environment and three harness tests added. Fourteen fresh full synthetic decoy and caller-env snapshots agree, including decoy root/file metadata and bytes; atime intentionally excluded. Existing stronger r1 symlink/ambient controls remain pinned |
| Detector reach | Exact old setup fails the intended source-mapped startup assertion after successful real Git setup; exact old helper fails the intended sensitive-output assertion. Neither is an import/setup accident. New named tests pass but do not detect R3-01; independent error-path witnesses establish that gap |
| Promised versus executable behavior | ProgramData/socket isolation and ordinary nonzero diagnostics are implemented and locally exercised. General safe failure receipts are incomplete for context/timeout paths. Native Windows compatibility remains unverified locally and cannot be inferred from source mapping or Darwin success |

## Preserved cumulative security, consumer and policy sweep

Fresh pins show both production verifiers, both trust files, both workflows, release gate/adjacent tests, policy documents and Decision 0034 byte-identical to accepted predecessor. The complete one-file delta and AST comparison establish that prior test assertions and API transport controls were retained. The original full consumer/release-gate reads and r1/r2 independent controls are reused with exact hashes in `prior-control-pins.json`, not rerun as a broad campaign.

| Preserved surface | Evidence and practical limit |
| --- | --- |
| Source-key mandate | Public key/fingerprint still matches the independently checked owner-authorized material and landed signature from r1. Addition to a file alone never establishes authority; current trusted checkout and reviewed policy remain assumptions |
| Closed source trust | One literal source-commit principal, git namespace, Ed25519 entry with strict base64/wire material; malformed/duplicate/wrong principal/namespace/type/material and missing/nonregular inputs refuse. Current named cases and unchanged pinned r1 directory/symlink/FIFO/material controls apply |
| Existing maintainer source route | Preserved alongside source-only trust. Nearby real fixture source and release-signing positives both pass; no original private release key used |
| Containment on every route | Resolve exact commit, require fixed main ref and ancestry before SSH or delegated GitHub success. Unknown/unsigned/uncontained/missing-ref/wrong-target refusals pass; unchanged r1 query-error/replacement/PGP-route controls retained |
| Git environment/trust roots | Verifier clears every inherited GIT_/XDG_ selector plus SSH socket/askpass, installs deliberate no-system/no-global/no-replacement policy and fixes checkout/trust/SSH format/program inputs. Current poison/decoy checks pass; trusted installed Git/SSH executables and local repository metadata remain outside an arbitrary same-user attacker sandbox |
| Ref/API context | Full resolved SHA and fixed repository/main/API authority; context drift, old override flags, unknown responses, option/ref and replacement attempts refuse. Fresh origin/main fetch is the workflow authority for containment; a local ref alone does not authenticate remote state |
| GitHub-created PGP fallback | Local signed-header dispatch, exact API object identity, GitHub committer, verified true/valid and nonempty string material remain. Delegated GitHub verification is not local PGP crypto proof or proof of PR review/squash topology/protection/current key custody. Arbitrary verified user metadata is not source authority |
| Product API confidentiality | R1-01 fix remains intact: fixed errors raised outside handlers, no cause/context. R2 real-header/parser/error/redirect/valid controls are pinned; all current committed transport tests pass. No real token/request dump or authenticated HTTP mutation used by fixtures |
| Optional settings observer | Fixed repository/strict settings/context; observation only. Does not change settings, audit protections or force rewriting a signed fast-forward head |
| Main CI | Same main-only provenance job/name and fresh default-ref fetch before exact GITHUB_SHA verification. PR provenance skip is unchanged and is not qualification/recovery |
| Release workflow/tag endorsement | Original maintainer trust ONLY for annotated tag verification; exact peeled tag/candidate equality, fresh main fetch, source check and all prior release gates preserved. Source-key tag rejection paired with release-key endorsement passes again |
| Namespace distinction | OpenSSH git namespace alone does not distinguish commit from tag. Separate consumer trust inputs enforce the source-only authorization boundary |
| Packaging/every consumer | Both distinctly scoped trust files remain packaged. Prior complete tracked/hidden consumer sweep is preserved by exact unchanged tree outside this test. No source-key tag consumer was introduced; no archive/full workflow execution claimed |
| Release gate/adjacent tests | Prior full reads of release_gate.py, test_release_gate.py and test_allowed_signers.py remain applicable to exact unchanged bytes. Version, clean exact checkout, immutable historical pins, independent review, ancestry/frozen diff, platform/assurance limits preserved. Signature checks remain separate workflow steps; release-gate tests were not executed here |
| Governance/Release/Security/README/Decision | Explicit amendment of old squash-only source procedure, reviewed signed exact-head fast-forward and existing GitHub-created alternative, separate tag endorsement and no retrospective compliance fiction remain. Human authorization/review/checks/protection are external governance gates, not proved by verifier success alone |
| Decision allocation | Current main and other open PR inventories (203, 200, 198, 119, 89) contain no separate 0034. Existing PR200 owns unchanged design-only 0033. All four source-policy links/heading and PR text remain consistent; historical reviews are preserved |
| Other authority | No key revocation/rotation, tag-custody expansion, platform-key/infra-exit/protection/status change, version bump, adoption, release, landing or native-host grant |

## Commands, instrument results and cleanup

Exact commands/environments/PIDs and safe receipts are retained privately. Abbreviated paths below avoid publishing host locations. The outer child environment is allowlisted with disposable absolute HOME/cwd/TMP/TEMP/XDG/cache/GOPATH; inherited Git/SSH/Go/native selectors are absent before deliberate synthetic values. Installed Apple Git is explicitly selected for this local toolchain, consistent with the retained r1 shim failure evidence; no installed tool or profile was changed.

| Independent command/control | Exit and result |
| --- | --- |
| `python3 -B <evidence>/run_controls.py named` | 0: exactly test_verify_release_commit, test_verify_release_merge_policy, test_allowed_signers; 39 tests, 0 failures/errors/skips; three observed real keygens with owned ProgramData/socket and empty passphrase; 14 decoy/caller snapshots agree |
| `python3 -B <evidence>/run_controls.py startup-old` | 0 detector: authentic predecessor setup reaches two real successful Git children, then intended missing-ProgramData assertion FAIL at keygen boundary; old root removed. Source mapping only, not native execution |
| `python3 -B <evidence>/run_controls.py diagnostic-old` | 0 detector: exact old helper produces one intended confidentiality assertion FAIL, zero errors. Sensitive failure details inspected only in memory |
| `python3 -B <evidence>/run_controls.py errors` | 0 witness: both R3-01 properties are RED. OS context retains fragment; real 10.008-second timeout reflects argv and retains both streams; exact direct child reaped. Nearby real child succeeds. Exit 0 records successful reproduction, not confidentiality acceptance |
| `python3 -B <evidence>/verify_closure.py` | 0: all 12 complete remote patches reconstruct candidate; old test assertions unchanged; Decision0034 allocation verified; stable checkout/index and 8,222 previous files preserved; clean exact review checkout and diff --check |
| Pinned r1 detector reuse | Byte-identical main verifier SHA-256 `423596a105e6dc2b12c44c9d804c2d4c48017cb16fc517cd23b838509f498ee7` failed intended containment/ambient assertions; reject-all and accept-all plants failed the paired positive/refusal properties. No setup failure counted as detection |
| Pinned r2 detector reuse | Original API helpers failed ten helper and ten stderr confidentiality assertions; four reject-all substitutions failed valid responses; twelve independent API fault pairs passed. Production and corresponding controls unchanged at this head |
| Direct fresh control PIDs | 95 unique recorded direct children/runners, all absent at cleanup. Producer's 249 and prior reviewers' counts remain separate. Real timeout killed only its recorded owned PID; no pattern kill/foreign process operation |
| Scratch | Sampled task peak 24,518,656 bytes plus 425,984 conservatively counted recent common-Git bytes = 24,944,640 bytes at control cleanup. Final report/publication/board receipts additionally checked with reserve against 48 MiB STOP / 64 MiB hard budget. 50ms sampling is not a continuous quota claim |
| Bounds and disposal | 1 MiB per-file OS limit and 60-second outer runner bound; candidate child communicate 10 seconds / reap bound 5 seconds. Synthetic keys/repos/home/cache/runtime roots removed; retained checkout and small text evidence only. No live key/socket/home/config/repository/index used as fixture input |
| Preserved red evidence | Earlier producer/reviewer reds remain unchanged. First r3 patch-reconstruction attempt failed on API patch final-line delimiter handling; retained as a red instrument receipt, corrected without source mutation, then full blob parity required and achieved. No setup red is reclassified as a passing product control |

No full unittest discovery, Go/product/build/module/package/race/repro/make/validate/regenerate/workflow/full gate/broad lint/crosscompile was run. Local Darwin Python 3.15.1 / Apple Git 2.54.0 does not replace hosted Python 3.12 Windows or other required native jobs. No Windows runtime, live native agent pipe or existing release private key was used.

## Exact pins and next checkpoint

| Input | Pin / preservation |
| --- | --- |
| Changed fixture module | Old blob `64e85f7bf472c8209c183a4f709177c927570d74`, SHA-256 `4c57eba0d18d7d42ba52b44d1224f04985715e9c5e1ccbbc43cc2719d9f66bcb`; current blob `5f437226b8d6f736bf280b7dd33ac8a8755c564c`, SHA-256 `2e06b226de3d37f078b476bc4075463892f7084ab1a9c5b9cf25e0a37c9f74b5` |
| Production commit/settings verifiers | Unchanged blobs `6de6aba39f78b600387afc4700b24c92f6995118` / `1cd3f5e64b1a40c462cb210e5deb16595ac84eb1` |
| Original tag trust | Unchanged blob `47f54d69ce219fa871602d77c554282c7d2c6207`, SHA-256 `12f4bfaf1bfcd954d30b0683f6b55b8ad7e19e5f60c88c95dbf0af702db82dc8` |
| Source-only trust | Unchanged blob `8e3976ba08be6a9c0386193f32cf73aa00b3b397`, SHA-256 `a757c518fb974409858ead6c177bb332f08877d13e96420196eaf045d0f9927f` |
| CI/release workflows | Unchanged blobs `78113b3efcfc056b610d16434223c951216ee672` / `81701d420fb09f3a1ed9c652e53d70dc260685ed` |
| Source decision | Unchanged Decision0034 blob `5d3ad36d529630950aada62a56580f45e24dba3e`; all policy documents and adjacent gate/tests have fresh blob/SHA-256 records in unchanged-input-pins.json |

One exact-head hosted snapshot: Specification Ubuntu/macOS/Windows IN_PROGRESS; Implementations on all three SUCCESS; Formatting/Links SUCCESS; Release target provenance SKIPPED. No polling, wait for CI, rerun, cancel, collector or artifact/log download occurred. The supplied historical failed-step file was read locally. Required new Windows/other hosted checks remain separate even after R3-01 correction and independent acceptance.

Historical actual-main run `40569031707`, job `122417945263`, stays genuinely failed. PR provenance SKIP is not recovery. Only after separately accepted exact signed landing can a NEW actual-main provenance result legitimately establish recovery. No historical status or prior review is rewritten.

```json
{
  "verdict": "CHANGES_REQUESTED",
  "head": "357f2724c9ecc22db210dfa6d43abdb451ccfa44",
  "base": "74a7dcc503c0fd7acd491299ab71290ae40e708e",
  "current_main": "74a7dcc503c0fd7acd491299ab71290ae40e708e",
  "findings": [
    {"id":"R3-01","severity":"medium","status":"open","location":"tools/test_verify_release_commit.py:152-162","requirement":"Safe fixture failure diagnostics and exception state","effect":"Timeout leaks argv through ordinary traceback and retains streams; launch OSError remains in implicit context","repro":"Real 10-second synthetic child timeout plus controlled sensitive PermissionError; valid real child succeeds","disposition":"Normalize expected launch/timeout failures outside handlers after exact-PID cleanup; add context/attribute/trace and valid-child controls"},
    {"id":"R1-01","severity":"medium","status":"closed_preserved","requirement":"Secret-free production API errors","evidence":"Production and r2 transport assertions unchanged; current named tests pass"},
    {"id":"R1-02","severity":"medium","status":"closed_preserved","requirement":"Unambiguous source decision identity","evidence":"Decision0034 unchanged and unoccupied elsewhere at current snapshot; PR200 preserved"}
  ],
  "source_trust": {"fingerprint":"SHA256:rE/sRlfdFt/vT5y5yukDCsq4aFtC48pbWsq9z4hm8MA","scope":"Commit-only plus existing maintainer source route; main containment before every success","adopted":false},
  "release_trust_unchanged":true,
  "native_or_hosted_gates": {
    "focused_tests":"39 pass, zero skips",
    "windows_startup_mapping":"Supported by independently matched pinned primary source; not native execution or exact historical diagnosis",
    "historical_windows_binary_status_stdout":"unknown",
    "specification_jobs":"All three in progress at sole exact-head snapshot",
    "pr_provenance":"skipped; not recovery",
    "actual_main_recovery":"not established; historical failure remains red"
  },
  "next_checkpoint":"Producer corrects R3-01 in the bounded fixture helper, publishes a signed successor, and obtains independent exact-head rereview plus real required Windows/hosted checks. Landing and new actual-main provenance remain separate later gates; no task/goal closure or adoption."
}
```

This immutable report is the body for an exact-head COMMENT review carrying CHANGES_REQUESTED, not a platform approval. Publication/body readback and distinct outcome/append-only note receipts are retained separately. Prior report/resource bytes, declarations, notes, status and assignee are preserved.
