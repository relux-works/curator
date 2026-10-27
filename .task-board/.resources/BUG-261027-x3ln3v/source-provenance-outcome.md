# Source-commit provenance candidate

Ready for independent source and security-depth review; not accepted or landed.

PR: https://github.com/relux-works/curator-spec/pull/202
Base: 74a7dcc503c0fd7acd491299ab71290ae40e708e
Head: ec70b25b5c7b7b23be64a05dacd6534eee71ade3
Producer COMMENT: https://github.com/relux-works/curator-spec/pull/202#pullrequestreview-5673407186

Explicit policy amendment in Decision 0033 separates authorized source SSH commits from release tags. The source-only role uses fingerprint SHA256:rE/sRlfdFt/vT5y5yukDCsq4aFtC48pbWsq9z4hm8MA. Source verification composes two fixed trust inputs, requires fetched default-branch containment before acceptance, and preserves the fixed-API GitHub-created signed route. Release tags still require the original release trust input alone, exact endorsement and every existing release gate. Both scoped trust inputs are packaged. No release-key revocation/rotation or protection change.

Release trust SHA-256 remains 12f4bfaf1bfcd954d30b0683f6b55b8ad7e19e5f60c88c95dbf0af702db82dc8, byte-identical to base. The unchanged landed source object's signature verifies with proposed source trust and fails original release trust with No principal matched. This is read-only object verification, not a main-run qualification.

Thirty focused Python tests pass on the exact candidate. Real synthetic SSH commits/tags pair source and existing signer acceptance with unknown/unsigned/uncontained/missing-ref/trust-input refusals and source-only tag rejection. Hostile Git selectors leave the decoy byte/identity snapshot unchanged and create no external trace. Fixed API/context/network/error controls refuse ambiguity. API acceptance controls are delegated-verification simulations, not local PGP cryptographic proof. Old verifier bytes fail the intended uncontained and ambient-key rejection assertions; accept-all/reject-all plants also fail paired checks. Human signature G and exact model coauthor verified.

Hosted jobs were queued/running at publication. The main-only provenance job is skipped on PRs by its unchanged trigger; no passing current main provenance is claimed. Historical main run 40569031707 remains failed. Full local suites, native hosted gates, independent review, eventual integration and a new exact-main run remain pending. All owned fixture keys/runtime roots were disposed and recorded processes exited. Sampled task scratch peak was 22634496 bytes, below the 48 MiB stop and 64 MiB hard limit. No task closure or trust adoption is claimed.
