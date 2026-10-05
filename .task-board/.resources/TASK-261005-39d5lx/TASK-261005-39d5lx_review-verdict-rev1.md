# Review verdict: TASK-261005-39d5lx \u2014 Design CIP-0007: manager-provisioned CLI tools

Verdict: **changes_requested**. Route: **analysis** (research/design rework).
Review date: 2026-10-05. Reviewer role: reviewer. Change request:
CR-TASK-261005-39d5lx-1, revision 1.

Base: `6ec4158d96200fca47f7d6b999f2fa3da0053274`.
Candidate tree: `ccd78c3a9c030134c42d570b9b61ad13b8c3e89f`.
The three current worktree files were compared byte-for-byte with this candidate
using `git show`; all three match. Fresh `git ls-remote origin refs/heads/main`
returned the base OID, so no upstream divergence was observed. No repository
files were edited, no commit was created, and no commit_ack was supplied.
`task-board spawn goal` was queried before recording this verdict: the run is
not goal-bound. Ordinary design rework is sufficient; no external blocker or
human-only decision prevents producing a corrected Draft.

## Numbered findings for the next researcher

1. **Medium \u2014 Remove forbidden organisation names and correct the evidence claim.**
   `cips/CIP-0007-manager-provisioned-cli-tools.md:18-19` names two vendors.
   The task expressly forbids organisation, team, client and personal names.
   The named-token grep returned **2 occurrences on 2 lines**, both in the CIP;
   the research note returned 0. The evidence note's assertion at
   `.research/261005_manager-provisioned-cli-tools.md:75-78` that none exist is
   therefore false. Replace these examples with generic CLI descriptions and
   record an actual rescan count. Personal-path/host token scan returned 0
   matches. These are bounded scans of known names and path/host tokens, backed
   by manual reading; they are not a universal named-entity detector. Existing
   unrelated CIP titles in README context were excluded from the added-text scan.

2. **High \u2014 Separate first-match selection from federation-wide revocation.**
   CIP lines 165-170 and 270-281 prescribe the first matching signed layer while
   also claiming deny-wins semantics from `protocol/registry.md` \u00a74. That section
   requires querying every enabled trusted registry and gives any verified
   revoked record precedence (lines 130-145); hardened unreachable-registry
   refusal is separate from a warning (lines 157-169). The draft does not say
   how later-layer revocations or failures are evaluated before accepting the
   first layer. Example to resolve: the built-in layer permits a tool while a
   later trusted layer revokes the same artefact. Define record/revocation match
   keys, a complete policy scan before positive selection, and the B-lane
   unavailable-registry decision. Add proposed negative vectors for revocation
   in a later layer, an unreachable later layer, and an allow override versus
   revocation. If intentionally diverging from audit federation, describe that
   as a new proposal with its residual, rather than claiming unchanged reuse.

3. **High \u2014 Bind upstream verification to an operator-trusted publisher and subject.**
   CIP lines 127-141 and 191-215 claim a compromised metadata mirror can only
   cause denial or propose bytes that fail verification. A checksum and valid
   signature alone do not establish the intended tool: a mirror can propose a
   different legitimately signed artefact, or supply its own verification key,
   unless the expected publisher and subject are independently bound. This is
   a design counterexample, not a claimed executed exploit. `registry.md` \u00a72
   (lines 53-60) verifies against currently pinned keys, not arbitrary keys;
   \u00a77 (lines 345-354) explicitly refuses embedded-key trust bootstrap. The draft
   defines registry pins, but no independent upstream signer/provenance policy
   for C's verifier. Define the proposed binding from tool identity and version/
   platform to allowed upstream signer or attestation identity and artefact
   digest, including key/identity revocation. State the resolver's privileged
   boundary and qualify the mirror claim. Add signer-substitution and valid-
   signature/wrong-subject negatives; reflect the result in the evidence note.

4. **High \u2014 Distinguish snapshot rollback protection from tool downgrade policy and replay authorization.**
   CIP lines 184-188 describe an older signed tool as the hostile author's worst
   case, but never analyze version downgrade or explicit downgrade approval.
   `registry.md` \u00a75 (lines 188-210) protects snapshot high-water, not semantic
   tool versions; a current snapshot can still contain an old tool release.
   CIP lines 227-234 and 282-292 also make local hash presence appear sufficient
   for offline success without defining current record trust, removed keys,
   revocation, and expiry on lock/store hits. \u00a72.1 revokes objects signed solely
   by a removed key; \u00a7\u00a75 and 8 bound snapshot acceptance and offline grace.
   Manager \u00a72.1 (lines 186-189) prohibits cache hits from bypassing current
   attestation/revocation gates. Define initial resolve versus refresh/downgrade
   versus replay, current policy checks on each, and how pre-seeded air-gapped
   snapshots expire or require explicit operator action. Document intentional
   availability residuals. Add proposed negatives for a lower tool version in a
   newer snapshot, revoked locked bytes, removed signing keys, and replay after
   freshness/grace expires. Do not promise indefinite air-gapped success from
   hash presence alone while claiming unchanged registry trust semantics.

5. **Medium \u2014 Correct which manifest the lock hash covers and specify versioned extensions.**
   CIP lines 246-253 introduce tools in a skill manifest and then claim that
   adding such a declaration naturally produces `source_lock_stale` through
   `manifest_sha256`; the evidence note repeats this at lines 41-45.
   `protocol/core.md` \u00a74 (lines 160-174) identifies the skill manifest as
   `agent-skill.json`/`csk-skill.json`. `protocol/skillfile-sources.md` \u00a71 line 29
   makes skill-manifest versions independent of Skillfile versions, and \u00a73
   (lines 161-170) hashes the **declaring Skillfile**, not every member's skill
   manifest. Keeping Skillfile.json unchanged while changing a member's tool
   declaration does not change that hash. Locked package identity and explicit
   refresh govern member byte changes (\u00a73, lines 172-187). Explain where tool
   declarations live, how they bind to member/package identity and resolved
   tools, and which events stale the declaring manifest versus refresh the
   member snapshot. Explicitly version the proposed lock changes: the current
   skillfile-lock-v1 schema is closed at both root and member objects. Preserve
   historical locks without assigning tool attestation to them. Correct the
   evidence note's fact-check accordingly.

6. **Medium \u2014 Make B's runtime exposure contract internally consistent and preserve execution controls.**
   CIP lines 133-137 introduce per-skill/project/global/profile PATH composition
   as a C addition; lines 176-178 defer it to C, but line 190 relies on it to
   handle hostile aliases and implementation leaf 3 (line 317) requires it for
   B. The recommended lane needs a defined minimal executable binding and
   collision policy before it can safely promise usable provisioned tools.
   `protocol/core.md` \u00a74.1.1 (lines 245-264, 335-343) already fixes enforced
   script execution and constructs PATH from exactly the declared exec names;
   absent/none exec admits only the interpreter. Explain B's binding for
   declared-only versus enforced commands, preserve capability restrictions,
   and name the identity checks and shared-scope conflicts. Adding a dependency
   must not automatically widen enforced PATH. Clarify provisioned-only B
   versus C's system-version probes. Add proposed negatives for an undeclared
   provisioned alias, duplicate aliases and conflicting shared-scope versions.
   The probe paragraph's reference to open question 5 (line 226) should point
   to question 8; its 'C \u00a76 rule' reference should identify issue #108 \u00a76.

## Swept review surfaces

| Surface | Result and evidence |
| --- | --- |
| Candidate/scope/freshness | PASS: exactly 3 documented paths, 481 added lines; candidate matches current files; remote main equals CR base. No protocol, profile, schema, conformance, CHANGELOG or LOGBOOK changes. |
| Template/process | PASS: 10/10 required top-level sections, 5/5 metadata fields, both Design subsections, Draft status and exactly one linked Draft index row. |
| Options/recommendation | PASS: A/B/C each have mechanism, pros, cons and security; B is argued through coverage and reduced policy surface. Runtime boundary needs finding 6. |
| Author downloads | Explicitly excluded from B declarations; future proposal, not a shipped guarantee. Mirror/publisher authorization needs finding 3. |
| Registry trust/federation | NEEDS REWORK: finding 2. Key overlap, out-of-band pins and durable URL-keyed rollback are otherwise correctly cited to registry \u00a7\u00a72.1 and 5. |
| Mirror integrity | NEEDS REWORK: finding 3; hash integrity is insufficient without independently authorized expected identity. |
| Downgrade/replay/offline/air gap | NEEDS REWORK: finding 4. No behavior or exploit execution is claimed. |
| Schema/lock/markers | NEEDS REWORK: finding 5. Historical schema preservation and new markers are discussed; collection-versus-member manifest identity is incorrect. |
| Transaction | Substantially fits manager \u00a7\u00a72.1\u20132.5: private verification then locked publication/rollback. Future adoption must retain protected-boundary/preimage revalidation and distinguish immutable-store retention from mutable target rollback (\u00a72.5, lines 530-578). |
| Runtime scopes/collisions | NEEDS REWORK: finding 6; capability-aware tool exposure is missing. |
| CIP-0005 audit boundary | Compatible as a proposed extension: identity evidence and environment allowlist are named. CIP-0005 Design \u00a73 (lines 99-112) and Security explain that PATH/HOME remain trust inputs and the filter is not a filesystem sandbox. Evidence note's vague Design paragraph citation should be made precise. |
| Conformance/test plan | Future positive/negative vectors and narrowing mutants are named without claiming execution. Add findings 2\u20136 cases; no runtime implementation is in this candidate. |
| Operator decisions | PASS: concrete owner/service, signing pins, default policy, resolver location, manifest/identity/distribution/probe questions; recommended answers present. These are Draft adoption questions, not blockers to research rework. |
| Names/privacy/evidence accuracy | FAIL: 2 forbidden named examples; finding 1 and false lock inference in finding 5. Personal path/host token matches: 0. |
| Validation | PASS: independently rerun as described below. The validator does not inspect CIPs, so this cannot attest the prose or security design. |

## Verification performed by this reviewer

- Read issue #108 in full, including comments (none), through gh; confirmed
  related #100/#101/#106/#107 titles and OPEN states independently.
- Checked the exact CR delta, template, CIP process and adjacent Draft structure;
  read the changed CIP/evidence and the applicable current core, registry,
  Skillfile-source, manager and schema text, plus CIP-0005's relevant contract.
- Independently ran `/tmp/cip0007-venv/bin/python -B tools/validate.py` against
  candidate-matching files: **exit 0**, `validated 73 schemas and 1294 vector files`.
  The pre-existing producer venv supplies requirements-dev.txt's
  `jsonschema==4.25.1`; bare python is absent. This is a reviewer rerun, not
  acceptance of the producer's attached validation claim. An earlier invocation
  had its returned output omitted by the tool output budget, so the observable
  rerun is the validation evidence recorded here.
- No implementation behavior, binary provisioning, platform execution, signature
  attack or mutation test was run; this deliverable is docs-only design research.
- Important anomalies are persisted here and in board notes. Repository LOGBOOK
  was not changed, following the task's explicit prohibition.

The next producer should revise both the CIP and its evidence note for findings
1\u20136, rerun validation and the bounded name scan, and hand off a new candidate
revision for another security review. This verdict does not accept revision 1.
