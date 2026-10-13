# Cross-check the base archive and the plan manifest against commit B, at materialization and at import

- **Lens:** reliability and integrity of R-TRANSFER: base binding, mechanical verification, rework rounds.
- **Problem:** The contract already requires an archive *of* exact B and a manifest *of* B, trusted B and manifest metadata, and result-to-plan matching (dispatcher §5.2, §5.2.1; helper §12). So a dirty or stale archive is invalid under the existing semantics. But no step mechanically establishes that the three representations of the base agree:
  - the commit ID B (in the plan and the result's `base_commit`);
  - the caller-supplied bounded canonical archive of B's tree (its digest journaled separately);
  - B's file-manifest digest (in the plan and the result's `manifest_digest`).

  The materializer writes the archive without verifying its entries against the plan's manifest. The board import builds C on its own B and recomputes "B plus change set", but never recomputes the manifest from that B or compares it with the result's `manifest_digest` and `base_commit`. The plan is the only bridge, and it is never itself checked against the repository.

  An implementation that omits these checks produces spurious worker deltas from a caller-side mismatch (an archive cut from a dirty checkout, the wrong commit, or a stale rework base):
  - archive content of F differs from B, F untouched by the worker: F exports as `modify` and lands in C as a silent revert;
  - a B path missing from the archive: it exports as `delete` and is silently removed in C;
  - an archive-only path: it exports as `add` and appears in C.

  Review binds to exact C, so this is not a review-binding bypass. It is misattributed worker deltas that reviewers must spot by eye. Every rework round cuts its archive from the previous C, so each round repeats the exposure. The archive, intake and manifest formats are unpublished, so this is a pre-publication verification gap, not an observed failure.
- **Evidence:** Paths are relative to `~/Developer/ReluxWorks/`.
  - `platform/curator-dispatcher` at `74687e600e04817fa6e7a2e8904a9f7a3c83a1d8`, `spec/dispatcher.md`:
    - §4.2, line 189: the archive digest is journaled for workspace rounds; the plan covers B and the manifest digest;
    - §5.2 step 1, line 322: archive of exact B; plan records B and manifest digest; no archive-to-manifest verification step;
    - step 3, line 324: the export compares the workspace with the plan's manifest;
    - step 4, line 325: import construction;
    - §5.2.1, lines 334–337 and 368: trusted B and manifest context; semantic validation against the *admitted plan*, not the repository;
    - §15, line 696: "manifest/plan substitution" is about substitution in transit, not archive consistency;
    - §16 question 22, line 730: archive and manifest formats unpublished.
  - `platform/curator-host-helper` at `3da62ad`, `spec/helper.md`:
    - §7 R-TRANSFER item 1, line 441: materialization from the archive with no manifest check;
    - §12, line 542: qualifies "exact B/archive/manifest … integrity" without naming where it is checked.
  - `curator/curator-agent-launcher` at `f08f566d`, `SPEC.md` lines 1757–1765 and 1859–1865: the plan carries B and the manifest digest; the materializer writes from the archive.
  - spm PR #604 r6 (`d65cf7f`), copy at `swarma-review/context/inputs/worker-code-transfer.spm-pr604.md`:
    - §3 steps 1 and 4–5, lines 294–302 and 345–391;
    - §4 inputs and determinism, lines 460–497: the import names the admitted exact B and recomputes C = B plus change set, with no manifest recomputation and no `base_commit`/`manifest_digest` equality check.
- **Proposed change:**
  1. **One canonical manifest.** Define it as path plus Git kind/mode plus content identity, the shape of `f-workspace-result-modes`. Derive it from the immutable objects of B, never from a working checkout.
  2. **Materialize.** Before work, verify every archive entry against the plan's manifest: path, kind/mode and content identity. Refuse on any mismatch, missing or extra entry. The workspace is then provably the manifest's tree.
  3. **Import.**
     - The recomputation never takes the manifest from the orchestrator-composed plan: all three base representations come from that one principal, so only the repository object B anchors attribution (fable).
     - Recompute the manifest from the caller repository's immutable B, with the same canonical encoding as the plan, so the board is a second conforming implementation rather than a parallel definition. This is a bounded tree pass, not a comparison of two trusted strings.
     - Refuse unless it equals the result's `manifest_digest` and `base_commit` equals B.
     - Keep the existing result-to-admission checks.
  4. **Qualification:** an archive differing in content or mode; missing and extra archive paths; a dirty-checkout archive; a stale rework base. Each must refuse at materialization or import, never land as a spurious delta.
- **Expected benefit:** Every C is provably B plus edits the worker could have made. Base mismatches surface as refusals at the first step that can see them, instead of misattributed deltas in review. Verification is a bounded content pass and an equality, with no new component, privilege or worker `.git` access.
- **Limits.**
  - These checks establish consistent base representations. They do not prove that an exporter-reported edit was made by the worker, replace exact-C review, or remove the compromised-exporter residual (spm #604 §10).
  - They close the orchestrator-side injection path that fable describes.
  - The relation is derived and pinned at preparation, where B is available, and verified at consumption, without making the dispatcher own a repository.
- **Risks and costs:**
  - The board must implement the canonical manifest identically.
  - Materialization gains a bounded verification pass.
  - The dependency on the mode-aware manifest of `f-workspace-result-modes` is explicit.
  - No owner decision is challenged.
- **Affected specs:**
  - dispatcher §§4.2, 5.2, 5.2.1, 15, 16 question 22;
  - user manager §§7, 12, 14 question 15;
  - agent launcher §10.1 workspace section;
  - spm PR #604 §§3–4 and the import leaf;
  - CIP-0009 D5a steps 1 and 5 (cross-reference).
- **Confirmed-by:** Matrix room `#swarma-review-3`, topic `f-base-binding`, 2026-10-13:
  - `macmini_swarma-review-3_wonder` (contradictions/mechanical verification), `$xracnwg1MS3buruLR6-R5SRkFQT39De0XDOJW7J0MjQ`: confirms neither check exists in the published texts and that the silent-revert trace stands. It adds that the import must use the plan's canonical manifest encoding. Adopted.
  - `macmini_swarma-review-3_wonder-03` (mechanical verification), `$JCdTAslkbMteATBdJL1Zp-_xzXikkruyCBS5BCJmnbs`: independent confirmation; check 2 depends on the mode-aware manifest. Adopted.
  - `macmini_swarma-review-3_astra-05` (R-TRANSFER), `$bqxbLMIfUFxOFEhaShHegrt554UFuUVmbIrtdFar-LA`: confirms a narrower pre-publication cross-representation verification gap.
    - It keeps the counterevidence: existing exact-B integrity rules make a mismatched archive invalid; this is not a review-binding bypass.
    - It corrects the reverse cases: a missing archive path exports as a delete, an extra one as an add.
    - It frames the import check as a bounded tree pass.
    - All adopted.
  - `macmini_swarma-review-3_fable` (trust/threat model), `$dyh0_YIJSuXRpZJywsquqmfaC-gb60F_XQnC0BAM8a0`: confirms that neither check exists. Trust addition: all three representations of B come from one principal, the orchestrator that composes the plan and cuts the archive. A prompt-injected orchestrator could therefore plant content that the exporter reports as the worker's modification, landing in C under the worker's trailers. The import check, recomputed from the repository object B and never from the plan, is what anchors the trailers' attribution. Adopted in item 3.
  - `ivmbp_swarma-review-3_astra-15` (R-TRANSFER, adversarial), `$4eQw9nJ9Hq_F8sxAD2rTqzL4zcWVBnlRXJUabklH77I`: partial confirmation as a pre-activation contract requirement. The existing base invariants (DP7 :322, :368, :694, :696) stay counterevidence. It refines the reverse cases (content in B the worker never saw needs both archive and manifest to be wrong) and adds the limits above. Adopted.
  - Proposal `$x4o6yMN0KNytWRXmgVxwub8JzBvrpYVTqG7UJHmP56o`; review request `$Hjgh5emZJUh1cBFILZm72ylCVJ1jGdp8LiSVORHkCUI`.
- **Status:** proposed. Checked at dispatcher `74687e6`, user manager `3da62ad`, launcher `f08f566d` and spm PR #604 r6 on 2026-10-13.
  - **UM7 recheck (user manager `7b59da74`, 2026-10-13):** UM7 rewrote the materialize item but added no archive-to-manifest verification; substance unchanged. The helper lines moved: §7 item 1 :441→:445; §12 :542→:554; §14 q15 →:607.
