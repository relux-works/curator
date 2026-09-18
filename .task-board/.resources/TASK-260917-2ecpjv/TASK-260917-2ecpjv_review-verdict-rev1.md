# TASK-260917-2ecpjv — independent review, CR revision 1

Reviewer run: RUN-260918-073a34. Reviewed artifacts: [qualification](board-resource://TASK-260917-2ecpjv/outcome/TASK-260917-2ecpjv_qualification.md), [handoff note](board-resource://TASK-260917-2ecpjv/outcome/TASK-260917-2ecpjv_handoff-note.md), and the empty candidate delta between 73fc8a4b264a4af6174cc1a9e487c82eedac8f2c and tree 559e88e475d813ebe6c28e3dcaa798af028c0c74.

Verdict: **accepted**. Release qualification: **qualified**. No repository change is the correct outcome: this leaf qualifies an existing upstream release and records evidence; the sibling owns the manager pin/code change. An empty diff was independently confirmed, not treated as automatic acceptance.

| Review item | Result and independent evidence |
|---|---|
| Remote tag and commit | PASS: remote annotated tag object 3d544cd47bd935d238c1c1c5faa066b92ada281a peels to dced9b8317e0e8af79edf2d0539b32bd22b6c85b; clone agrees. |
| Tagger, message, signature | PASS: full tag object quoted below; Relux Bot <bot@relux.works>. Verification exit 0 with the main spec checkout's allowed-signers file. Narrowed signer control excluding bot rejects, exit 1. |
| Validation | PASS: reviewer reran make validate at the exact commit, not reused from producer. Full result below. |
| Regeneration and clean disposable tree | PASS: reviewer reran make regenerate-check; exit 0 and empty status. |
| Manifest and release pin | PASS: ea9dd5a0030b889079cf5655517056e66a0f1d61812f8c2920b4beddc1fd24ed matches both SHA-256 fields. Full release JSON below. |
| Required families | PASS: 6/6 present, 30 vector JSON files total, two environments*.json files; all published counts match the outcome. Registry page_boundary_cases = 9. |
| Expected absence | PASS: environments-source-signers.json absent at rc.12; first addition is 684c9f1324d46b4938b2e5943f20c89e27971ec8, whose parent is the tag commit. |
| Story SPEC_PIN | PASS: branch ci.yml pins the full expected commit. |
| Cleanup and scope | PASS with stated bound below: original /tmp/qual-rc12 absent; review clone removed; main spec and Story worktree clean; main curator source clean; tag ref inventories unchanged during review. |
| Architecture/AC | PASS: evidence-only qualification preserves ownership of spec publication and manager implementation. No new production gate or implementation introduced by this leaf. |

## Precision and bounds

- The authenticated signer is the release bot key SHA256:qbALzjdB9BRgYJjDkX/p9EAPLEofB2AbskJc6Ftwhds. The general brief's phrase “maintainer key” is interpreted according to the explicit review brief, which requires this bot signature. The human ECDSA key alone does not authorize this tag. The producer quotes the correct identity throughout its evidence.
- The curator control checkout is **not literally clean**: board activity/resources were already modified when review began. These are disclosed in the producer outcome and the baseline transcript below. All non-board source paths are clean; reviewer writes are confined to its disposable directory and authorized board CLI mutations. No stronger claim about historical invisible metadata writes is made: the producer's allowed worktree/fetch workflow necessarily touched Git administrative metadata. Current cleanliness cannot prove every historical command, and no pre-producer tag inventory is available. The reviewer verifies current tag identity against the remote and unchanged main tag inventories over this review.
- Publication coverage is 6/6 requested families. The nine boundary vectors include rejection cases for stale/equivocated data, missing or invalid signatures, and chain mismatch. This qualification proves their publication and validates the spec suite; it does not independently attest manager runtime enforcement, which belongs to sibling implementation review.
- No source-code findings or rework requests. The existing evidence documents the scope caveats, so they do not invalidate qualification. Full independent gate transcripts follow; no producer test result substitutes for the reviewer run.
- Reviewer checklist basis: implementation/AC and architecture items are satisfied by the evidence-only scope; tests by the rerun gates; the conditional non-acceptance routing item is not applicable because this verdict accepts revision 1.

## Reproducible command transcripts

All commands below used GIT_OPTIONAL_LOCKS=0 and PATH prefixed with /Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin. Exit codes are recorded; expected negative signer control exits 1. Reads cite the cloned Git objects at the exact tag commit, the remote advertised tag, and the named Story branch as primary sources.

```text
$ git -C /Users/administrator/Developer/ReluxWorks/curator/curator status --short
 M .task-board/.activity/BUG-260918-18ny6p/events.ndjson
 M .task-board/.activity/EPIC-260908-2wp8wn/events.ndjson
 M .task-board/.activity/EPIC-260910-16qce1/events.ndjson
 M .task-board/.activity/EPIC-260910-2hw1xb/events.ndjson
 M .task-board/.activity/EPIC-260910-ohqchs/events.ndjson
 M .task-board/.activity/STORY-260910-20sx61/events.ndjson
 M .task-board/.activity/STORY-260910-2xe3n2/events.ndjson
 M .task-board/.activity/STORY-260910-9484i4/events.ndjson
 M .task-board/.activity/STORY-260910-stz5f0/events.ndjson
 M .task-board/.activity/STORY-260915-3w11un/events.ndjson
 M .task-board/.activity/STORY-260916-12lbww/events.ndjson
 M .task-board/.activity/STORY-260916-1ll22r/events.ndjson
 M .task-board/.activity/STORY-260917-3w3lvj/events.ndjson
 M .task-board/.activity/TASK-260910-28kmef/events.ndjson
 M .task-board/.activity/TASK-260910-2g5v17/events.ndjson
 M .task-board/.activity/TASK-260910-dufdai/events.ndjson
 M .task-board/.activity/TASK-260910-s9jz1g/events.ndjson
 M .task-board/.activity/TASK-260917-2ecpjv/events.ndjson
 M .task-board/.activity/TASK-260918-24eazm/events.ndjson
 M .task-board/.activity/TASK-260918-3moznc/events.ndjson
 M .task-board/.resources/BUG-260918-18ny6p/BUG-260918-18ny6p_results.md
 M .task-board/.resources/TASK-260910-dufdai/TASK-260910-dufdai_results.md
 M .task-board/EPIC-260908-2wp8wn_curator-run-and-infra-migration/STORY-260915-3w11un_workstation-and-checkout-readiness/BUG-260918-18ny6p_rose-air-homebrew-rustup-not-on-runner-path/progress.md
 M .task-board/EPIC-260908-2wp8wn_curator-run-and-infra-migration/STORY-260915-3w11un_workstation-and-checkout-readiness/progress.md
 M .task-board/EPIC-260908-2wp8wn_curator-run-and-infra-migration/STORY-260916-12lbww_onboarding-heuristic-platform-parity/TASK-260918-24eazm_spec-dotfile-manager-platform-table/progress.md
 M .task-board/EPIC-260908-2wp8wn_curator-run-and-infra-migration/STORY-260916-12lbww_onboarding-heuristic-platform-parity/progress.md
 M .task-board/EPIC-260908-2wp8wn_curator-run-and-infra-migration/progress.md
 M .task-board/EPIC-260910-16qce1_security-audit-remediation-registry-service/STORY-260910-2xe3n2_registry-robustness-hardening/TASK-260910-28kmef_service-recursionerror-400/progress.md
 M .task-board/EPIC-260910-16qce1_security-audit-remediation-registry-service/STORY-260910-2xe3n2_registry-robustness-hardening/progress.md
 M .task-board/EPIC-260910-16qce1_security-audit-remediation-registry-service/STORY-260910-9484i4_registry-key-management/TASK-260910-s9jz1g_service-key-passphrase-support/progress.md
 M .task-board/EPIC-260910-16qce1_security-audit-remediation-registry-service/STORY-260910-9484i4_registry-key-management/progress.md
 M .task-board/EPIC-260910-16qce1_security-audit-remediation-registry-service/STORY-260910-stz5f0_import-upstream-high-water/TASK-260910-2g5v17_service-import-high-water/progress.md
 M .task-board/EPIC-260910-16qce1_security-audit-remediation-registry-service/STORY-260910-stz5f0_import-upstream-high-water/progress.md
 M .task-board/EPIC-260910-16qce1_security-audit-remediation-registry-service/progress.md
 M .task-board/EPIC-260910-2hw1xb_security-audit-remediation-manager-and-spec/STORY-260916-1ll22r_absence-vs-read-failure-collapse/TASK-260918-3moznc_spec-absence-vs-read-failure-discipline/progress.md
 M .task-board/EPIC-260910-2hw1xb_security-audit-remediation-manager-and-spec/STORY-260916-1ll22r_absence-vs-read-failure-collapse/progress.md
 M .task-board/EPIC-260910-2hw1xb_security-audit-remediation-manager-and-spec/STORY-260917-3w3lvj_conformance-pin-rc12-promotion/TASK-260917-2ecpjv_qualify-curator-spec-rc12/progress.md
 M .task-board/EPIC-260910-2hw1xb_security-audit-remediation-manager-and-spec/STORY-260917-3w3lvj_conformance-pin-rc12-promotion/progress.md
 M .task-board/EPIC-260910-2hw1xb_security-audit-remediation-manager-and-spec/progress.md
 M .task-board/EPIC-260910-ohqchs_approved-skillfile-sources-implementation/STORY-260910-20sx61_source-audit-runtime-and-build-integration/TASK-260910-dufdai_implement-source-aware-build-receipts-and-cache/progress.md
 M .task-board/EPIC-260910-ohqchs_approved-skillfile-sources-implementation/STORY-260910-20sx61_source-audit-runtime-and-build-integration/progress.md
 M .task-board/EPIC-260910-ohqchs_approved-skillfile-sources-implementation/progress.md
?? .task-board/.activity/TASK-260918-11f9l1/
?? .task-board/.resources/BUG-260918-18ny6p/18ny6p-republish-rev1.md
?? .task-board/.resources/TASK-260910-28kmef/TASK-260910-28kmef_change-request_rev1.patch
?? .task-board/.resources/TASK-260910-28kmef/TASK-260910-28kmef_results.md
?? .task-board/.resources/TASK-260910-28kmef/TASK-260910-28kmef_review-brief.md
?? .task-board/.resources/TASK-260910-2g5v17/TASK-260910-2g5v17_change-request_rev1.patch
?? .task-board/.resources/TASK-260910-2g5v17/TASK-260910-2g5v17_results.md
?? .task-board/.resources/TASK-260910-2g5v17/TASK-260910-2g5v17_review-brief.md
?? .task-board/.resources/TASK-260910-dufdai/TASK-260910-dufdai_change-request_rev3.patch
?? .task-board/.resources/TASK-260910-dufdai/TASK-260910-dufdai_rev3-candidate.patch
?? .task-board/.resources/TASK-260910-dufdai/dufdai-reapply-rev3.md
?? .task-board/.resources/TASK-260910-s9jz1g/TASK-260910-s9jz1g_change-request_rev1.patch
?? .task-board/.resources/TASK-260910-s9jz1g/TASK-260910-s9jz1g_results.md
?? .task-board/.resources/TASK-260910-s9jz1g/TASK-260910-s9jz1g_review-brief.md
?? .task-board/.resources/TASK-260917-2ecpjv/
?? .task-board/.resources/TASK-260918-11f9l1/
?? .task-board/EPIC-260910-2hw1xb_security-audit-remediation-manager-and-spec/STORY-260917-3w3lvj_conformance-pin-rc12-promotion/TASK-260918-11f9l1_combine-rc12-union-with-moved-trunk/
exit=0
```

```text
$ git -C /Users/administrator/Developer/ReluxWorks/curator/curator show-ref --tags
ac5aa449a54027e43ee6d493786cc8689b931df8 refs/tags/v0.1.0
55c5607833b4af5c82c93288c67af494ef7eaa50 refs/tags/v0.1.1
2166113c897f54b2e5ca2a47bc2442d20a0d7f52 refs/tags/v0.12.4
b5d8206e546a138b0907a7cc954b0d32f4ee64ec refs/tags/v0.12.5
ed075901f5fd3534dc21228c591ea751a644536d refs/tags/v0.13.0
e4f693a95b2b1de499a6e7b26bd5d6dc337ba172 refs/tags/v0.14.0
9f1f7d6a5d54b10fd87d8dae31fb2bafa002593f refs/tags/v0.14.0-rc.1
2ebe79ed1647d4a66b2e35f8fe30ad92f2680c0c refs/tags/v0.14.0-rc.2
124654a051f60995e826b75ab43a3c4f3359f5ff refs/tags/v0.14.0-rc.3
5c558de07e3b0fcbc8236f7c6898c86b44e0401c refs/tags/v0.15.0-rc.1
exit=0
```

```text
$ git -C /Users/administrator/Developer/ReluxWorks/curator/curator-spec status --short
exit=0
```

```text
$ git -C /Users/administrator/Developer/ReluxWorks/curator/curator-spec show-ref --tags
720284a3bb3e41fa202420ba85540c4315a10ca1 refs/tags/v1.0.0-rc.1
68a2ce966577c6e53f41c277909a6ddee6874723 refs/tags/v1.0.0-rc.10
132437cbb77509cfd9d6cdcc144ef110e9de2a60 refs/tags/v1.0.0-rc.11
3d544cd47bd935d238c1c1c5faa066b92ada281a refs/tags/v1.0.0-rc.12
97fa47e094426edc0287c24594cdd0abf6eca8f9 refs/tags/v1.0.0-rc.2
f229b96c986642be7f341092e34565e7e8df099c refs/tags/v1.0.0-rc.3
3780f50847ce7f513436c950ec1656a8ab298185 refs/tags/v1.0.0-rc.5
de704f2951e683d52ae8e475cb690b918a94d4c5 refs/tags/v1.0.0-rc.7
ad247840292487d5d88ac44331798b6b4182a79f refs/tags/v1.0.0-rc.8
b67966449220d42218bd50420e74dac673431464 refs/tags/v1.0.0-rc.9
exit=0
```

```text
$ git -C /Users/administrator/Developer/ReluxWorks/curator/curator/.temp/STORY-260917-3w3lvj/worktree status --short
exit=0
```

```text
$ git -C /Users/administrator/Developer/ReluxWorks/curator/curator/.temp/STORY-260917-3w3lvj/worktree show-ref --tags
ac5aa449a54027e43ee6d493786cc8689b931df8 refs/tags/v0.1.0
55c5607833b4af5c82c93288c67af494ef7eaa50 refs/tags/v0.1.1
2166113c897f54b2e5ca2a47bc2442d20a0d7f52 refs/tags/v0.12.4
b5d8206e546a138b0907a7cc954b0d32f4ee64ec refs/tags/v0.12.5
ed075901f5fd3534dc21228c591ea751a644536d refs/tags/v0.13.0
e4f693a95b2b1de499a6e7b26bd5d6dc337ba172 refs/tags/v0.14.0
9f1f7d6a5d54b10fd87d8dae31fb2bafa002593f refs/tags/v0.14.0-rc.1
2ebe79ed1647d4a66b2e35f8fe30ad92f2680c0c refs/tags/v0.14.0-rc.2
124654a051f60995e826b75ab43a3c4f3359f5ff refs/tags/v0.14.0-rc.3
5c558de07e3b0fcbc8236f7c6898c86b44e0401c refs/tags/v0.15.0-rc.1
exit=0
```

```text
$ test '!' -e /tmp/qual-rc12
exit=0
```

```text
$ git diff 73fc8a4b264a4af6174cc1a9e487c82eedac8f2c 559e88e475d813ebe6c28e3dcaa798af028c0c74
exit=0
```

```text
$ git ls-remote --tags git@github.com:relux-works/curator-spec.git v1.0.0-rc.12 'v1.0.0-rc.12^{}'
3d544cd47bd935d238c1c1c5faa066b92ada281a	refs/tags/v1.0.0-rc.12
dced9b8317e0e8af79edf2d0539b32bd22b6c85b	refs/tags/v1.0.0-rc.12^{}
exit=0
```

```text
$ git clone -q git@github.com:relux-works/curator-spec.git /tmp/rc12-review.iYHRo4/repo
exit=0
```

```text
$ git -C /Users/administrator/Developer/ReluxWorks/curator/curator-spec config gpg.ssh.allowedSignersFile
/Users/administrator/developer/curator/.temp/orchestration/allowed_signers
exit=0
```

```text
$ ssh-keygen -lf /Users/administrator/developer/curator/.temp/orchestration/allowed_signers
256 SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM oparin@me.com (ECDSA)
256 SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM ivan@relux.works (ECDSA)
256 SHA256:qbALzjdB9BRgYJjDkX/p9EAPLEofB2AbskJc6Ftwhds bot@relux.works (ED25519)
exit=0
```

```text
$ cd /tmp/rc12-review.iYHRo4/repo && git fetch -q origin tag v1.0.0-rc.12
exit=0
```

```text
$ cd /tmp/rc12-review.iYHRo4/repo && git rev-parse v1.0.0-rc.12 'v1.0.0-rc.12^{commit}'
3d544cd47bd935d238c1c1c5faa066b92ada281a
dced9b8317e0e8af79edf2d0539b32bd22b6c85b
exit=0
```

```text
$ cd /tmp/rc12-review.iYHRo4/repo && git cat-file -p v1.0.0-rc.12
object dced9b8317e0e8af79edf2d0539b32bd22b6c85b
type commit
tag v1.0.0-rc.12
tagger Relux Bot <bot@relux.works> 1789729928 +0400

v1.0.0-rc.12 — the 2026-09 security audit and its wave-1/2 specifications

Sixteen commits since v1.0.0-rc.11: the 2026-09 architectural security
audit of the protocol specification with its environments/launch-plane
supplement (E1–E7) and the implementation-verification appendix, and the
first remediation revisions — S6 shell-hook trust gate for project env
files, E2 direct-only class: system modules, E4 umbrella provider
resolution from trust roots, S4 bounded MCP env passthrough and
declaration surfacing, and R1/P1 records and log pages bound to their
committed snapshot boundary (records-response-v2 / log-response-v2, with
the registry-client page_boundary_cases) — plus the repository transport
revision 2, the opt-in local and Git Skillfile sources, the claude/codex
CLI aliases and the migration and credential-mode proposals filed as
draft decisions.

Protocol version stays 1.0.0-rc.9 (release/1.0.0-rc.9.json regenerated at
each revision). This tag is the conformance suite revision the Curator
manager pins as SPEC_PIN for release qualification (TASK-260917-2ecpjv,
STORY-260917-3w3lvj on the Curator board).
-----BEGIN SSH SIGNATURE-----
U1NIU0lHAAAAAQAAADMAAAALc3NoLWVkMjU1MTkAAAAg8bvFNfTkcvVdoPhctSTT9NN6pE
0dsyPkd2eo1B0NMgMAAAADZ2l0AAAAAAAAAAZzaGE1MTIAAABTAAAAC3NzaC1lZDI1NTE5
AAAAQMOhpW8yM6SSwfe95T2ejpEX+1prM0TR30rHoOTw25GiRgCaEVileKUQgYf7beekFi
4ycFtrHY49/b424Fv2kg8=
-----END SSH SIGNATURE-----
exit=0
```

```text
$ cd /tmp/rc12-review.iYHRo4/repo && git -c gpg.ssh.allowedSignersFile=/Users/administrator/developer/curator/.temp/orchestration/allowed_signers verify-tag v1.0.0-rc.12
Good "git" signature for bot@relux.works with ED25519 key SHA256:qbALzjdB9BRgYJjDkX/p9EAPLEofB2AbskJc6Ftwhds
exit=0
```

```text
$ cd /tmp/rc12-review.iYHRo4/repo && git checkout -q dced9b8317e0e8af79edf2d0539b32bd22b6c85b
exit=0
```

```text
$ cd /tmp/rc12-review.iYHRo4/repo && git rev-parse HEAD
dced9b8317e0e8af79edf2d0539b32bd22b6c85b
exit=0
```

```text
$ cd /tmp/rc12-review.iYHRo4/repo && shasum -a 256 conformance/v1/manifest.json
ea9dd5a0030b889079cf5655517056e66a0f1d61812f8c2920b4beddc1fd24ed  conformance/v1/manifest.json
exit=0
```

```text
$ cd /tmp/rc12-review.iYHRo4/repo && cat release/1.0.0-rc.9.json
{
  "assurance": {
    "default_mode": "portable",
    "portable_execution_policy": "manager-worker-v1",
    "portable_policy": "portable-cli-policy-v1",
    "silent_downgrade_permitted": false,
    "skill_vendored_provider_allowed": false,
    "verified_execution_policy": "verified-provider-execution-v1",
    "verified_implementations": [],
    "verified_platform_claims": [],
    "verified_policy": "verified-provider-policy-v1",
    "verified_provider_contract": "host-execution-provider-v1"
  },
  "candidate_protocol_pin": {
    "manifest_sha256": "sha256:ea9dd5a0030b889079cf5655517056e66a0f1d61812f8c2920b4beddc1fd24ed",
    "suite_root": "conformance/v1"
  },
  "claim_v5": {
    "claim_protocol_version": "1.0.0-rc.9",
    "claims_emitted": [],
    "schema": "schemas/v1/conformance-claim-v5.schema.json"
  },
  "created_at": "2026-08-23T00:00:00Z",
  "downstream_consumption": {
    "committed_release_pin_advanced": false,
    "environment": "CURATOR_CONFORMANCE_ROOT",
    "required_manifest_sha256": "sha256:ea9dd5a0030b889079cf5655517056e66a0f1d61812f8c2920b4beddc1fd24ed"
  },
  "historical_release": {
    "immutable": true,
    "metadata_path": "release/1.0.0-rc.8.json",
    "metadata_sha256": "sha256:293f101d10665061aa049efa72141f9e3c5d608bbde300e882f6e3e095e31ede",
    "protocol_version": "1.0.0-rc.8",
    "source_commit": "f8c405aa3ad0a39d260c2ed93684e55c5a346359"
  },
  "legacy_release": "1.0.0-rc.8",
  "protocol_version": "1.0.0-rc.9",
  "source_baseline_commit": "f8c405aa3ad0a39d260c2ed93684e55c5a346359"
}
exit=0
```

```text
$ cd /tmp/rc12-review.iYHRo4/repo && python3 -c 'import json,pathlib,hashlib
p=pathlib.Path('"'"'conformance/v1'"'"'); m=json.loads((p/'"'"'manifest.json'"'"').read_text()); d=hashlib.sha256((p/'"'"'manifest.json'"'"').read_bytes()).hexdigest(); rel=json.loads(pathlib.Path('"'"'release/1.0.0-rc.9.json'"'"').read_text())
assert rel['"'"'candidate_protocol_pin'"'"']['"'"'manifest_sha256'"'"']==rel['"'"'downstream_consumption'"'"']['"'"'required_manifest_sha256'"'"']=='"'"'sha256:'"'"'+d
print('"'"'manifest:'"'"',m['"'"'protocol_version'"'"'],m['"'"'generator'"'"'],m['"'"'generated_at'"'"'],'"'"'files:'"'"',len(m['"'"'files'"'"']))
files=sorted((p/'"'"'vectors'"'"').glob('"'"'*.json'"'"')); print('"'"'full inventory:'"'"',len(files),[f.name for f in files])
for name in ['"'"'manager-config-v2'"'"','"'"'environments'"'"','"'"'environments-env-passthrough'"'"','"'"'umbrella-provider-resolution'"'"','"'"'shell-hook-trust'"'"','"'"'registry-client'"'"']:
 data=json.loads((p/'"'"'vectors'"'"'/(name+'"'"'.json'"'"')).read_text());print(name, '"'"'list['"'"'+str(len(data))+'"'"']'"'"' if isinstance(data,list) else {k:len(v) if isinstance(v,list) else v for k,v in data.items() if isinstance(v,list) or k in ['"'"'protocol_version'"'"','"'"'finding'"'"']})
 if name=='"'"'registry-client'"'"': print('"'"'page_boundary_cases:'"'"',data['"'"'page_boundary_cases'"'"'])
print('"'"'environments inventory:'"'"',[f.name for f in (p/'"'"'vectors'"'"').glob('"'"'environments*.json'"'"')])
assert not (p/'"'"'vectors/environments-source-signers.json'"'"').exists();print('"'"'environments-source-signers.json: absent'"'"')
'
manifest: 1.0.0-rc.9 tools/generate-vectors 2026-07-13T00:00:00Z files: 1071
full inventory: 30 ['assurance-modes.json', 'build-drivers.json', 'canonical-invalid.json', 'canonical-valid.json', 'closures.json', 'conformance-claim-v3-qualification.json', 'context-detectors.json', 'context-versions.json', 'environments-env-passthrough.json', 'environments.json', 'external-repository-acquisition.json', 'external-repository-lifecycle.json', 'go-host-execution-policy.json', 'identifiers.json', 'locale-selectors.json', 'manager-config-v2.json', 'manager-config.json', 'manager-lifecycle.json', 'module-roots.json', 'portable-paths.json', 'registry-behavior.json', 'registry-client.json', 'registry-resolution.json', 'registry-service.json', 'script-host-execution-policy.json', 'shell-hook-trust.json', 'skill-manifest-resolution.json', 'snapshot-acquisition.json', 'source-identities.json', 'umbrella-provider-resolution.json']
manager-config-v2 list[48]
environments {'header_cases': 4, 'materialization_cases': 24, 'protocol_version': '1.0.0-rc.9'}
environments-env-passthrough {'allowlist_empty_cases': 6, 'default_resolution_cases': 7, 'protocol_version': '1.0.0-rc.9', 's4_profiles': 2, 'schema_cases': 4, 'surfacing_cases': 6, 'surfacing_order_cases': 2}
umbrella-provider-resolution {'cases': 14, 'protocol_version': '1.0.0-rc.9'}
shell-hook-trust {'commands': 3, 'diagnostics': 2, 'files': 2, 'finding': 'S6', 'protocol_version': '1.0.0-rc.9', 'rollout_profiles': 2, 'execution_recipe': 7, 'cases': 14}
registry-client {'page_boundary_cases': 9, 'pagination_rejections': 4, 'retry_cases': 7, 'rollback_state_cases': 4, 'snapshot_transitions': 4}
page_boundary_cases: [{'accepted': True, 'boundary_present': True, 'boundary_version': 8, 'chain_boundary_equal': True, 'diagnostic': None, 'high_water_advanced': True, 'name': 'fresh-boundary-advances-high-water', 'registry_excluded': False, 'same_body': False, 'signature_valid': True, 'stored_version': 7}, {'accepted': True, 'boundary_present': True, 'boundary_version': 8, 'chain_boundary_equal': True, 'diagnostic': None, 'high_water_advanced': False, 'name': 'equal-version-same-body-accepted', 'registry_excluded': False, 'same_body': True, 'signature_valid': True, 'stored_version': 8}, {'accepted': False, 'boundary_present': True, 'boundary_version': 8, 'chain_boundary_equal': True, 'diagnostic': 'registry_page_boundary_stale', 'high_water_advanced': False, 'name': 'equal-version-different-body-rejected', 'registry_excluded': True, 'same_body': False, 'signature_valid': True, 'stored_version': 8}, {'accepted': False, 'boundary_present': True, 'boundary_version': 7, 'chain_boundary_equal': True, 'diagnostic': 'registry_page_boundary_stale', 'high_water_advanced': False, 'name': 'below-high-water-rejected', 'registry_excluded': True, 'same_body': False, 'signature_valid': True, 'stored_version': 8}, {'accepted': False, 'boundary_present': True, 'boundary_version': 8, 'chain_boundary_equal': False, 'diagnostic': 'registry_page_boundary_mismatch', 'high_water_advanced': False, 'name': 'chain-boundary-mismatch-rejected', 'registry_excluded': True, 'same_body': False, 'signature_valid': True, 'stored_version': 7}, {'accepted': False, 'boundary_present': False, 'boundary_version': 0, 'chain_boundary_equal': False, 'diagnostic': 'registry_page_boundary_missing', 'high_water_advanced': False, 'name': 'missing-boundary-excluded', 'registry_excluded': True, 'same_body': False, 'signature_valid': False, 'stored_version': 7}, {'accepted': False, 'boundary_present': True, 'boundary_version': 8, 'chain_boundary_equal': True, 'diagnostic': 'registry_page_boundary_missing', 'high_water_advanced': False, 'name': 'bad-signature-rejected', 'registry_excluded': True, 'same_body': False, 'signature_valid': False, 'stored_version': 7}, {'accepted': False, 'boundary_present': True, 'boundary_version': 7, 'chain_boundary_equal': False, 'diagnostic': 'registry_page_boundary_mismatch', 'high_water_advanced': False, 'name': 'stale-and-mismatch-reports-mismatch', 'registry_excluded': True, 'same_body': False, 'signature_valid': True, 'stored_version': 8}, {'accepted': False, 'boundary_present': True, 'boundary_version': 9, 'chain_boundary_equal': False, 'diagnostic': 'registry_page_boundary_mismatch', 'high_water_advanced': False, 'name': 'higher-and-mismatch-never-advances', 'registry_excluded': True, 'same_body': False, 'signature_valid': True, 'stored_version': 7}]
environments inventory: ['environments.json', 'environments-env-passthrough.json']
environments-source-signers.json: absent
exit=0
```

```text
$ cd /tmp/rc12-review.iYHRo4/repo && git log --all --oneline -- conformance/v1/vectors/environments-source-signers.json
684c9f1 Verify source signers before the lock and confirm system-changing updates (E1)
exit=0
```

```text
$ cd /tmp/rc12-review.iYHRo4/repo && git show -s --format=%H%n%P%n%s 684c9f1324d46b4938b2e5943f20c89e27971ec8
684c9f1324d46b4938b2e5943f20c89e27971ec8
dced9b8317e0e8af79edf2d0539b32bd22b6c85b
Verify source signers before the lock and confirm system-changing updates (E1)
exit=0
```

```text
$ git -C /Users/administrator/Developer/ReluxWorks/curator/curator show task-board/story/STORY-260917-3w3lvj:.github/workflows/ci.yml
name: CI

on:
  push:
    # gate/** carries the throwaway snapshots scripts/remote-gate.sh pushes
    # for the task-board landing gate; the branch is deleted after the run.
    branches: [main, 'gate/**']
  pull_request:
  workflow_dispatch:
    inputs:
      candidate_ref:
        description: >-
          Full 40-character relux-works/curator-spec revision holding the candidate
          suite. The lane is schema-agnostic: it qualifies whatever suite the revision
          publishes. Branches, tags, short hashes and placeholders are rejected.
          Leave empty to use candidate_root instead.
        required: false
        default: ""
      candidate_root:
        description: >-
          Absolute path to a conformance/v1 root already materialised on the runner.
          Use instead of candidate_ref when the candidate has no published revision.
        required: false
        default: ""
      candidate_manifest_sha256:
        description: >-
          Expected sha256 of <root>/manifest.json. When set, a mismatch aborts the run
          rather than silently re-baselining onto a different candidate.
        required: false
        default: ""

permissions:
  contents: read

env:
  # The one immutable committed protocol-suite pin, referenced by every job.
  # It stays on the currently qualified released revision; promoting it is
  # owned by TASK-260720-38l1sy, after TASK-260720-25d05o qualifies the
  # release. This pin is the revision the signed released tag v1.0.0-rc.12
  # names on curator-spec: it publishes protocol 1.0.0-rc.9, whose
  # conformance/v1/manifest.json is
  # sha256:ea9dd5a0030b889079cf5655517056e66a0f1d61812f8c2920b4beddc1fd24ed
  # -- the exact digest release/1.0.0-rc.9.json requires of a downstream
  # consumer in downstream_consumption.required_manifest_sha256. It claims no
  # later revision. A candidate suite never appears here -- it enters through
  # the non-default `candidate-conformance` job below and nowhere else.
  #
  # This root serves the WHOLE module: `suite-plan.sh` defers nothing against
  # it. The umbrella-provider-resolution, environments-env-passthrough and
  # system-module E2 families that the previous pin did not publish now run
  # with the root exported on every default lane, instead of taking their
  # root-content skip path there.
  SPEC_PIN: dced9b8317e0e8af79edf2d0539b32bd22b6c85b
  # actions/setup-go only began forcing GOTOOLCHAIN=local in v6.0.0; this
  # workflow pins @v5, so without these two lines a job may download a
  # toolchain or read a per-user go env file. The identity step below reads
  # both back rather than trusting this block.
  GOTOOLCHAIN: "local"
  # Quoted deliberately: bare `off` is a YAML 1.1 boolean and would reach the
  # runner as "false", which is a real per-user go env file path, not "off".
  GOENV: "off"
  # The one pnpm release the real-pnpm integration tests admit. This is a
  # COPY of internal/pnpmsource.SupportedPNPMVersion, kept next to the lanes
  # that install it so the installed release is visible in the workflow; the
  # copy cannot drift silently because `.github/ci/pnpm-pin-guard.sh` fails
  # any lane where the two disagree. Every lane that runs the Go test suite
  # installs exactly this release with npm into a lane-local prefix (see the
  # per-lane steps), so the TestRealPinnedPNPM* cases execute everywhere
  # instead of skipping for "pinned pnpm executable unavailable" -- and so a
  # broken pnpm shim that happens to be first on a self-hosted runner's PATH
  # can no longer fail the lane, as the macbook-iv fallback shim did on run
  # 35072267145 (`node <shim> --version` died with a SyntaxError on the
  # shim's `set -euo pipefail` line, failing `read pnpm version` in
  # newConcretePNPMRunner). The lane-local bin directory is prepended via
  # GITHUB_PATH precisely to outrank any such ambient shim.
  #
  # npm, not corepack, is the installer: the tests resolve the pnpm
  # executable and stage its package root, which a corepack dispatcher shim
  # does not provide (rehearsal: 3x FAIL MODULE_NOT_FOUND with corepack, 3x
  # PASS with npm-installed pnpm).
  PNPM_PIN: "10.33.0"

jobs:
  # ---------------------------------------------------------------------------
  # The default lane, against the committed released pin, on all three runners.
  #
  # `test-gate.sh` plans the run from that root alone. The committed pin
  # publishes every artefact `.github/ci/root-artifacts.tsv` declares, so the
  # deferred set is empty here and every package runs with
  # CURATOR_CONFORMANCE_ROOT exported; a pin that stopped publishing one would
  # defer its package and name the missing artefact in the report rather than
  # dropping it. `internal/godriver` -- which implements
  # rc5-native-control-inventory-v1, defined for macOS and Windows only -- is
  # excluded on linux with its fail-closed rejection asserted there. Nothing is
  # dropped without appearing in the uploaded evidence.
  # ---------------------------------------------------------------------------
  test:
    name: Test (${{ matrix.os }})
    runs-on: ${{ matrix.os }}
    strategy:
      fail-fast: false
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
    steps:
      - uses: actions/checkout@v4
        with:
          submodules: true

      - name: Checkout authoritative protocol suite
        uses: actions/checkout@v4
        with:
          repository: relux-works/curator-spec
          ref: ${{ env.SPEC_PIN }}
          path: protocol-spec

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      # The hosted images ship their own Node, but the pnpm the
      # TestRealPinnedPNPM* cases admit is installed here, never taken from
      # the image: setup-node guarantees an npm toolchain, and the
      # lane-local prefix below is prepended to PATH so it outranks any
      # ambient pnpm shim (macbook-iv run 35072267145: a broken fallback
      # shim first on PATH failed `read pnpm version` and turned the
      # rose-air lane red).
      - uses: actions/setup-node@v4
        with:
          node-version: "22"

      - name: Verify pnpm pin matches Go constant
        shell: bash
        run: bash .github/ci/pnpm-pin-guard.sh

      - name: Install pinned pnpm via npm
        shell: bash
        run: |
          set -euo pipefail
          # Lane-local prefix: npm stages a REAL pnpm package root here, which
          # the TestRealPinnedPNPM* cases need (they resolve the executable
          # and stage dir(dir(entrypoint))). A corepack dispatcher shim
          # provides no such root (rehearsal: MODULE_NOT_FOUND).
          npm install -g --prefix "$RUNNER_TEMP/pnpm-prefix" "pnpm@${{ env.PNPM_PIN }}"
          # npm lays out the global bin dir per platform ($prefix/bin on unix,
          # the prefix root itself on Windows); prepend both so the pinned
          # pnpm is first on PATH on every runner and outranks any ambient
          # shim (macbook-iv run 35072267145).
          echo "$RUNNER_TEMP/pnpm-prefix/bin" >>"$GITHUB_PATH"
          echo "$RUNNER_TEMP/pnpm-prefix" >>"$GITHUB_PATH"

      - name: Verify rust pin matches Go constant
        shell: bash
        run: bash .github/ci/rust-pin-guard.sh

      # rustup is the installer on every runner: the hosted images ship it,
      # and the self-hosted runner gets it once per
      # docs/self-hosted-runner-setup.md. The install step reads the channel
      # from rust-toolchain.toml (verified above against
      # internal/rustsource.SupportedRustToolchainVersion), installs exactly
      # that toolchain, and prepends the rustup shim directory to PATH, so
      # `rustc -vV` in the rustsource production cases resolves the pinned
      # release instead of failing the lane (rose-air run 35121791685).
      - name: Install pinned Rust toolchain via rustup
        shell: bash
        run: bash .github/ci/install-rust-toolchain.sh

      - name: Verify Go toolchain identity
        shell: bash
        run: bash .github/ci/toolchain-identity.sh

      - name: Verify versioned Go install compatibility
        shell: bash
        run: bash .github/ci/release-source-gate.sh

      - name: gofmt check
        if: runner.os != 'Windows'
        shell: bash
        run: |
          unformatted="$(gofmt -l cmd internal)"
          if [ -n "$unformatted" ]; then
            echo "gofmt: files need formatting:"
            echo "$unformatted"
            exit 1
          fi

      - name: go vet
        run: go vet ./...

      # Proves the ledger's platform claims against the real per-GOOS builds:
      # a case required on windows must actually compile into the windows
      # build, and a case compiled into a build the ledger never mentions is a
      # coverage gap rather than a silent pass.
      - name: Ledger consistency
        shell: bash
        run: bash .github/ci/ledger-consistency.sh .temp/ci-evidence/ledger

      # Runs the planned package set and enforces .github/ci/platform-cases.tsv:
      # the Windows runner must execute the DACL, reparse-point and `.cmd`
      # launcher cases; the unix runners must execute the ownership, no-follow,
      # read-only-source, resource-policy and executable cases; and every skip
      # anywhere in the run is classified and recorded by name.
      #
      # The per-package deadline is a hosted-runner budget, not a behavioural
      # bar: the Windows runner pays a real syscall and process-spawn cost that
      # the unix runners do not. Measured on run 32648313266 (windows-latest):
      # cmd/curator 1099s and internal/install/atomicity 1027s here against
      # 454s and 125s for the same packages on macOS. 30m leaves those two
      # under 40% headroom on the slowest runner, so Windows gets 60m and
      # every other runner keeps 30m. A hang is still bounded and still fatal.
      - name: go test + platform-case gate
        shell: bash
        env:
          CURATOR_CONFORMANCE_ROOT: ${{ github.workspace }}/protocol-spec/conformance/v1
          GO_TEST_TIMEOUT: ${{ runner.os == 'Windows' && '60m' || '30m' }}
        run: bash .github/ci/test-gate.sh .temp/ci-evidence/test

      - name: Upload gate evidence
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: test-evidence-${{ matrix.os }}
          path: .temp/ci-evidence/
          if-no-files-found: warn

  # ---------------------------------------------------------------------------
  # The same default lane on the organisation's self-hosted macOS ARM64
  # runner (rose-air). Hosted macos-latest stays authoritative for the
  # platform claim; this lane adds an Apple-silicon execution host the
  # operator controls, so a hosted-image regression and a real-machine
  # regression are distinguishable. It runs the identical gate scripts and
  # uploads its own evidence under a distinct name.
  # ---------------------------------------------------------------------------
  # Gated on the repository variable ROSE_AIR_RUNNER=true: the runner is
  # registered for the organisation's board repository only until an
  # administrator registers it for this repository, and an unmatched
  # self-hosted label would leave the job queued forever.
  test-self-hosted:
    name: Test (rose-air)
    # Runs on main pushes only: every landing gets self-hosted evidence, while
    # pull requests and gate snapshots stay on the hosted matrix, because the
    # single self-hosted slot is shared with other repositories and would
    # serialise every candidate behind unrelated 90-minute runs.
    if: ${{ vars.ROSE_AIR_RUNNER == 'true' && github.event_name == 'push' && github.ref == 'refs/heads/main' }}
    runs-on: [self-hosted, macOS, ARM64]
    timeout-minutes: 90
    steps:
      - uses: actions/checkout@v4
        with:
          submodules: true

      - name: Checkout authoritative protocol suite
        uses: actions/checkout@v4
        with:
          repository: relux-works/curator-spec
          ref: ${{ env.SPEC_PIN }}
          path: protocol-spec

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      # The hosted macOS image ships Node; the self-hosted runner does not.
      # internal/npmsource's real-npm cases skip with "npm unavailable"
      # otherwise, a reason the skip classes do not tolerate on darwin.
      # The same toolchain provisions the npm-installed pnpm below.
      - uses: actions/setup-node@v4
        with:
          node-version: "22"

      # This lane is where the incident happened (macbook-iv run 35072267145:
      # a broken fallback pnpm shim first on PATH failed `read pnpm
      # version`). The lane-local prefix bin directory is prepended to PATH
      # precisely to outrank any such ambient shim.
      - name: Verify pnpm pin matches Go constant
        shell: bash
        run: bash .github/ci/pnpm-pin-guard.sh

      - name: Install pinned pnpm via npm
        shell: bash
        run: |
          set -euo pipefail
          # Lane-local prefix: npm stages a REAL pnpm package root here, which
          # the TestRealPinnedPNPM* cases need (they resolve the executable
          # and stage dir(dir(entrypoint))). A corepack dispatcher shim
          # provides no such root (rehearsal: MODULE_NOT_FOUND).
          npm install -g --prefix "$RUNNER_TEMP/pnpm-prefix" "pnpm@${{ env.PNPM_PIN }}"
          # npm lays out the global bin dir per platform ($prefix/bin on unix,
          # the prefix root itself on Windows); prepend both so the pinned
          # pnpm is first on PATH on every runner and outranks any ambient
          # shim (macbook-iv run 35072267145).
          echo "$RUNNER_TEMP/pnpm-prefix/bin" >>"$GITHUB_PATH"
          echo "$RUNNER_TEMP/pnpm-prefix" >>"$GITHUB_PATH"

      - name: Verify rust pin matches Go constant
        shell: bash
        run: bash .github/ci/rust-pin-guard.sh

      # rustup is the installer on every runner: the hosted images ship it,
      # and the self-hosted runner gets it once per
      # docs/self-hosted-runner-setup.md. The install step reads the channel
      # from rust-toolchain.toml (verified above against
      # internal/rustsource.SupportedRustToolchainVersion), installs exactly
      # that toolchain, and prepends the rustup shim directory to PATH, so
      # `rustc -vV` in the rustsource production cases resolves the pinned
      # release instead of failing the lane (rose-air run 35121791685).
      - name: Install pinned Rust toolchain via rustup
        shell: bash
        run: bash .github/ci/install-rust-toolchain.sh

      - name: Verify Go toolchain identity
        shell: bash
        run: bash .github/ci/toolchain-identity.sh

      - name: gofmt check
        shell: bash
        run: |
          unformatted="$(gofmt -l cmd internal)"
          if [ -n "$unformatted" ]; then
            echo "gofmt: files need formatting:"
            echo "$unformatted"
            exit 1
          fi

      - name: go vet
        run: go vet ./...

      - name: Ledger consistency
        shell: bash
        run: bash .github/ci/ledger-consistency.sh .temp/ci-evidence/ledger

      - name: go test + platform-case gate
        shell: bash
        env:
          CURATOR_CONFORMANCE_ROOT: ${{ github.workspace }}/protocol-spec/conformance/v1
          GO_TEST_TIMEOUT: 30m
        run: bash .github/ci/test-gate.sh .temp/ci-evidence/test

      - name: Upload gate evidence
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: test-evidence-rose-air
          path: .temp/ci-evidence/
          if-no-files-found: warn

  # ---------------------------------------------------------------------------
  # Race.
  #
  # macos-latest is an inventory platform, so `-race` there covers the whole
  # package set including internal/godriver, satisfying the acceptance
  # criteria's race gate on a supported runner. ubuntu-latest satisfies the
  # "supported race job on at least Linux" clause and covers transaction,
  # buildcache, install, install/atomicity and interop; godriver is excluded
  # there by the protocol's own qualification, with the exclusion asserted
  # rather than assumed.
  #
  # windows-latest is deliberately absent: the Go race detector needs a C
  # toolchain there and this repository has no measurement that the hosted
  # image satisfies it. Recorded, not silently dropped.
  # ---------------------------------------------------------------------------
  race:
    name: Race (${{ matrix.os }})
    runs-on: ${{ matrix.os }}
    strategy:
      fail-fast: false
      matrix:
        os: [ubuntu-latest, macos-latest]
    steps:
      - uses: actions/checkout@v4
        with:
          submodules: true

      - name: Checkout authoritative protocol suite
        uses: actions/checkout@v4
        with:
          repository: relux-works/curator-spec
          ref: ${{ env.SPEC_PIN }}
          path: protocol-spec

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      # Same provisioning as the Test lane: the TestRealPinnedPNPM* cases
      # run under -race too, so the pinned pnpm is installed here and
      # prepended to PATH ahead of any ambient shim (macbook-iv run
      # 35072267145).
      - uses: actions/setup-node@v4
        with:
          node-version: "22"

      - name: Verify pnpm pin matches Go constant
        shell: bash
        run: bash .github/ci/pnpm-pin-guard.sh

      - name: Install pinned pnpm via npm
        shell: bash
        run: |
          set -euo pipefail
          # Lane-local prefix: npm stages a REAL pnpm package root here, which
          # the TestRealPinnedPNPM* cases need (they resolve the executable
          # and stage dir(dir(entrypoint))). A corepack dispatcher shim
          # provides no such root (rehearsal: MODULE_NOT_FOUND).
          npm install -g --prefix "$RUNNER_TEMP/pnpm-prefix" "pnpm@${{ env.PNPM_PIN }}"
          # npm lays out the global bin dir per platform ($prefix/bin on unix,
          # the prefix root itself on Windows); prepend both so the pinned
          # pnpm is first on PATH on every runner and outranks any ambient
          # shim (macbook-iv run 35072267145).
          echo "$RUNNER_TEMP/pnpm-prefix/bin" >>"$GITHUB_PATH"
          echo "$RUNNER_TEMP/pnpm-prefix" >>"$GITHUB_PATH"

      - name: Verify rust pin matches Go constant
        shell: bash
        run: bash .github/ci/rust-pin-guard.sh

      # rustup is the installer on every runner: the hosted images ship it,
      # and the self-hosted runner gets it once per
      # docs/self-hosted-runner-setup.md. The install step reads the channel
      # from rust-toolchain.toml (verified above against
      # internal/rustsource.SupportedRustToolchainVersion), installs exactly
      # that toolchain, and prepends the rustup shim directory to PATH, so
      # `rustc -vV` in the rustsource production cases resolves the pinned
      # release instead of failing the lane (rose-air run 35121791685).
      - name: Install pinned Rust toolchain via rustup
        shell: bash
        run: bash .github/ci/install-rust-toolchain.sh

      - name: Verify Go toolchain identity
        shell: bash
        run: bash .github/ci/toolchain-identity.sh

      # internal/install/atomicity is the long pole under -race: 115.687s on
      # the accepted candidate's macOS measurement, against a 480s acceptance
      # bar. The 30m per-package timeout is headroom for a slower hosted
      # runner, not a licence for an unbounded test.
      - name: go test -race + platform-case gate
        shell: bash
        env:
          CURATOR_CONFORMANCE_ROOT: ${{ github.workspace }}/protocol-spec/conformance/v1
          GO_TEST_FLAGS: -race
        run: bash .github/ci/test-gate.sh .temp/ci-evidence/race

      - name: Upload race evidence
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: race-evidence-${{ matrix.os }}
          path: .temp/ci-evidence/
          if-no-files-found: warn

  lint:
    name: Lint
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          submodules: true

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      - name: Verify Go toolchain identity
        shell: bash
        run: bash .github/ci/toolchain-identity.sh

      # `version: latest` is a mutable supply-chain input: a new golangci-lint
      # release can turn this job red with no repository change. Pinned to the
      # version `latest` resolved to when this gate was written.
      - uses: golangci/golangci-lint-action@v7
        with:
          version: v2.12.2

      # Security findings may be suppressed narrowly and by name; they may not
      # be suppressed wholesale. A suppression that names no rule silences
      # every rule, including ones written after the comment.
      - name: No broad suppression
        shell: bash
        run: bash .github/ci/no-broad-suppression.sh

  # The gates are only worth their exit codes if they reject what they claim
  # to reject. This runs on all three runners because the gate scripts must
  # behave identically under Git Bash on Windows.
  gate-selftest:
    name: Gate self-test (${{ matrix.os }})
    runs-on: ${{ matrix.os }}
    strategy:
      fail-fast: false
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
    steps:
      - uses: actions/checkout@v4

      - name: Self-test the CI gates
        shell: bash
        run: bash .github/ci/gate-selftest.sh

  interop:
    name: Interop conformance gate
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          submodules: true

      - name: Checkout authoritative protocol suite
        uses: actions/checkout@v4
        with:
          repository: relux-works/curator-spec
          ref: ${{ env.SPEC_PIN }}
          path: protocol-spec

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      - name: Verify Go toolchain identity
        shell: bash
        run: bash .github/ci/toolchain-identity.sh

      - name: Shared protocol suite
        env:
          CURATOR_CONFORMANCE_ROOT: ${{ github.workspace }}/protocol-spec/conformance/v1
        run: go test -count=1 -timeout 30m ./internal/interop/ -v

  naming-gate:
    name: Naming gate
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          submodules: true

      - name: Employer name gate
        run: |
          # Curator is an open protocol implementation. The employer under
          # whose roof some of it was written is not part of that protocol and
          # must not appear anywhere in the repository -- not in source, docs,
          # task records or commit messages. Historical records that named a
          # local checkout path or an internal remote were masked to
          # `intranet` rather than deleted.
          #
          # Both patterns are assembled from parts so this file does not trip
          # the gate it defines.
          full="$(printf '%s%s' wild berries)"
          short="$(printf '%s%s' w b)"
          status=0
          if grep -riIn --exclude-dir=.git "$full" .; then
            echo "naming gate: the employer name must not appear in this repository"
            status=1
          fi
          if grep -riIEn --exclude-dir=.git "\\b$short\\b" .; then
            echo "naming gate: the employer's short name must not appear in this repository"
            status=1
          fi
          exit "$status"
  candidate-conformance:
    name: Candidate suite (${{ matrix.os }})
    if: >-
      github.event_name == 'workflow_dispatch' &&
      (inputs.candidate_ref != '' || inputs.candidate_root != '')
    runs-on: ${{ matrix.os }}
    strategy:
      fail-fast: false
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
    steps:
      - uses: actions/checkout@v4
        with:
          submodules: true

      # A revision and a pre-materialised root are independent identities.
      # Reject the ambiguous combination before checking out either candidate
      # revision or recording any candidate evidence.
      - name: Reject ambiguous candidate inputs
        shell: bash
        env:
          CANDIDATE_REF: ${{ inputs.candidate_ref }}
          CANDIDATE_ROOT_INPUT: ${{ inputs.candidate_root }}
        run: bash .github/ci/candidate-suite.sh verify-inputs "$CANDIDATE_REF" "$CANDIDATE_ROOT_INPUT"

      - name: Reject a non-immutable candidate revision
        if: inputs.candidate_ref != ''
        shell: bash
        env:
          CANDIDATE_REF: ${{ inputs.candidate_ref }}
        run: bash .github/ci/candidate-suite.sh verify-ref "$CANDIDATE_REF"

      - name: Check out the candidate suite
        if: inputs.candidate_ref != ''
        uses: actions/checkout@v4
        with:
          repository: relux-works/curator-spec
          ref: ${{ inputs.candidate_ref }}
          path: candidate-spec

      # The committed released pin is checked out here only so the candidate
      # can be proved to differ from it. This job never runs against it and
      # never modifies it.
      - name: Checkout the committed released pin for comparison
        uses: actions/checkout@v4
        with:
          repository: relux-works/curator-spec
          ref: ${{ env.SPEC_PIN }}
          path: protocol-spec

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      # Same provisioning as the Test lane: this lane runs the full suite
      # through test-gate.sh, so the TestRealPinnedPNPM* cases execute here
      # too, against the pinned pnpm prepended to PATH ahead of any ambient
      # shim (macbook-iv run 35072267145).
      - uses: actions/setup-node@v4
        with:
          node-version: "22"

      - name: Verify pnpm pin matches Go constant
        shell: bash
        run: bash .github/ci/pnpm-pin-guard.sh

      - name: Install pinned pnpm via npm
        shell: bash
        run: |
          set -euo pipefail
          # Lane-local prefix: npm stages a REAL pnpm package root here, which
          # the TestRealPinnedPNPM* cases need (they resolve the executable
          # and stage dir(dir(entrypoint))). A corepack dispatcher shim
          # provides no such root (rehearsal: MODULE_NOT_FOUND).
          npm install -g --prefix "$RUNNER_TEMP/pnpm-prefix" "pnpm@${{ env.PNPM_PIN }}"
          # npm lays out the global bin dir per platform ($prefix/bin on unix,
          # the prefix root itself on Windows); prepend both so the pinned
          # pnpm is first on PATH on every runner and outranks any ambient
          # shim (macbook-iv run 35072267145).
          echo "$RUNNER_TEMP/pnpm-prefix/bin" >>"$GITHUB_PATH"
          echo "$RUNNER_TEMP/pnpm-prefix" >>"$GITHUB_PATH"

      - name: Verify rust pin matches Go constant
        shell: bash
        run: bash .github/ci/rust-pin-guard.sh

      # rustup is the installer on every runner: the hosted images ship it,
      # and the self-hosted runner gets it once per
      # docs/self-hosted-runner-setup.md. The install step reads the channel
      # from rust-toolchain.toml (verified above against
      # internal/rustsource.SupportedRustToolchainVersion), installs exactly
      # that toolchain, and prepends the rustup shim directory to PATH, so
      # `rustc -vV` in the rustsource production cases resolves the pinned
      # release instead of failing the lane (rose-air run 35121791685).
      - name: Install pinned Rust toolchain via rustup
        shell: bash
        run: bash .github/ci/install-rust-toolchain.sh

      - name: Verify Go toolchain identity
        shell: bash
        run: bash .github/ci/toolchain-identity.sh

      - name: Resolve and record the candidate suite identity
        id: candidate
        shell: bash
        env:
          CANDIDATE_REF: ${{ inputs.candidate_ref }}
          CANDIDATE_ROOT_INPUT: ${{ inputs.candidate_root }}
          CANDIDATE_EXPECTED_MANIFEST_SHA256: ${{ inputs.candidate_manifest_sha256 }}
        run: |
          set -u
          if [ -n "$CANDIDATE_ROOT_INPUT" ]; then
            root="$CANDIDATE_ROOT_INPUT"
          else
            root="$GITHUB_WORKSPACE/candidate-spec/conformance/v1"
          fi
          echo "root=$root" >>"$GITHUB_OUTPUT"

          pin_manifest="$GITHUB_WORKSPACE/protocol-spec/conformance/v1/manifest.json"
          if command -v shasum >/dev/null 2>&1; then
            PIN_MANIFEST_SHA256="$(shasum -a 256 <"$pin_manifest" | awk '{print $1}')"
          else
            PIN_MANIFEST_SHA256="$(sha256sum <"$pin_manifest" | awk '{print $1}')"
          fi
          export PIN_MANIFEST_SHA256
          bash .github/ci/candidate-suite.sh record "$root" .temp/ci-evidence/candidate

      # The full candidate suite on every platform. The platform-case ledger is
      # enforced exactly as in the default lane, and CI_REQUIRE_FULL_ROOT=1
      # additionally forbids the package deferral the default lane would permit
      # against a root that published less than the committed pin does.
      #
      # The committed pin now serves the whole module, so the default lane
      # already runs one served invocation and this lane's load is no longer
      # the heavier of the two. The Windows deadline stays where the measured
      # contention put it: on run 32648313266, cmd/curator and
      # internal/install/atomicity ran past 1800s in a single invocation on
      # four Windows cores with every executed case passing, which the ledger
      # then read as two required Windows cases that never ran. Windows gets
      # the same 60m as the default lane; the deadline is still fixed and
      # still fatal.
      - name: Candidate suite + platform-case gate
        shell: bash
        env:
          CURATOR_CONFORMANCE_ROOT: ${{ steps.candidate.outputs.root }}
          CI_REQUIRE_FULL_ROOT: "1"
          GO_TEST_TIMEOUT: ${{ runner.os == 'Windows' && '60m' || '30m' }}
        run: bash .github/ci/test-gate.sh .temp/ci-evidence/test

      - name: Upload candidate evidence
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: candidate-evidence-${{ matrix.os }}
          path: .temp/ci-evidence/
          if-no-files-found: warn
exit=0
```

```text
$ git -C /Users/administrator/Developer/ReluxWorks/curator/curator-spec config gpg.ssh.allowedSignersFile
/Users/administrator/developer/curator/.temp/orchestration/allowed_signers
exit=0
```

```text
$ cd /tmp/rc12-review.iYHRo4/repo && git -c gpg.ssh.allowedSignersFile=/tmp/rc12-review.iYHRo4/maintainer-only-signers verify-tag v1.0.0-rc.12
Good "git" signature with ED25519 key SHA256:qbALzjdB9BRgYJjDkX/p9EAPLEofB2AbskJc6Ftwhds
No principal matched.
exit=1
```

```text
$ cd /tmp/rc12-review.iYHRo4/repo && make validate
python3 tools/validate.py
validated 62 schemas and 1071 vector files
python3 -B -m unittest discover -s tools -p 'test_*.py'
.............................................................................................................................................................................................................................................................................................................
----------------------------------------------------------------------
Ran 301 tests in 379.791s

OK
go test ./tools/...
ok  	github.com/relux-works/curator-spec/tools/generate-vectors	4.905s
exit=0
```

```text
$ cd /tmp/rc12-review.iYHRo4/repo && make regenerate-check
go run ./tools/generate-vectors -root .
git diff --exit-code -- conformance/v1 release/1.0.0-rc.5.json release/1.0.0-rc.6.json release/1.0.0-rc.7.json release/1.0.0-rc.8.json release/1.0.0-rc.9.json
exit=0
```

```text
$ cd /tmp/rc12-review.iYHRo4/repo && git status --short
exit=0
```

```text
$ git -C /Users/administrator/Developer/ReluxWorks/curator/curator-spec show-ref --tags
720284a3bb3e41fa202420ba85540c4315a10ca1 refs/tags/v1.0.0-rc.1
68a2ce966577c6e53f41c277909a6ddee6874723 refs/tags/v1.0.0-rc.10
132437cbb77509cfd9d6cdcc144ef110e9de2a60 refs/tags/v1.0.0-rc.11
3d544cd47bd935d238c1c1c5faa066b92ada281a refs/tags/v1.0.0-rc.12
97fa47e094426edc0287c24594cdd0abf6eca8f9 refs/tags/v1.0.0-rc.2
f229b96c986642be7f341092e34565e7e8df099c refs/tags/v1.0.0-rc.3
3780f50847ce7f513436c950ec1656a8ab298185 refs/tags/v1.0.0-rc.5
de704f2951e683d52ae8e475cb690b918a94d4c5 refs/tags/v1.0.0-rc.7
ad247840292487d5d88ac44331798b6b4182a79f refs/tags/v1.0.0-rc.8
b67966449220d42218bd50420e74dac673431464 refs/tags/v1.0.0-rc.9
exit=0
```

```text
$ git -C /Users/administrator/Developer/ReluxWorks/curator/curator-spec status --short -- . ':(exclude).task-board'
exit=0
```

```text
$ git -C /Users/administrator/Developer/ReluxWorks/curator/curator show-ref --tags
ac5aa449a54027e43ee6d493786cc8689b931df8 refs/tags/v0.1.0
55c5607833b4af5c82c93288c67af494ef7eaa50 refs/tags/v0.1.1
2166113c897f54b2e5ca2a47bc2442d20a0d7f52 refs/tags/v0.12.4
b5d8206e546a138b0907a7cc954b0d32f4ee64ec refs/tags/v0.12.5
ed075901f5fd3534dc21228c591ea751a644536d refs/tags/v0.13.0
e4f693a95b2b1de499a6e7b26bd5d6dc337ba172 refs/tags/v0.14.0
9f1f7d6a5d54b10fd87d8dae31fb2bafa002593f refs/tags/v0.14.0-rc.1
2ebe79ed1647d4a66b2e35f8fe30ad92f2680c0c refs/tags/v0.14.0-rc.2
124654a051f60995e826b75ab43a3c4f3359f5ff refs/tags/v0.14.0-rc.3
5c558de07e3b0fcbc8236f7c6898c86b44e0401c refs/tags/v0.15.0-rc.1
exit=0
```

```text
$ git -C /Users/administrator/Developer/ReluxWorks/curator/curator status --short -- . ':(exclude).task-board'
exit=0
```

```text
$ git status --short
exit=0
```

```text
$ test '!' -e /tmp/qual-rc12
exit=0
```

```text
$ rm -rf /tmp/rc12-review.iYHRo4/repo
exit=0
```

```text
$ test '!' -e /tmp/rc12-review.iYHRo4/repo
exit=0
```
