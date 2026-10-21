# Review brief r2: audit Markdown scope, case-alias fix (BUG-261021-2qv4zj, curator)

Correctness re-review (packaging consistency between context publication and the content audit) at the exact head in the request, against `W-AUDITSCOPE-rework-r2-brief.md`, the r1 review `BUG-261021-2qv4zj_review-r1.md` and the producer's `BUG-261021-2qv4zj_results-r2.md`. The producer chose option (a): one filesystem-aware `WalkContext` now feeds both `CopyContext` and the Markdown audit. Protocol exclusions stay case-sensitive. Ruleset 7.

| Surface | Check |
| --- | --- |
| F1 closure | Re-run the nine alias cases from review r1 and the exact CLI strict/high install. All must refuse, with no lock, no source and no native context publication. Then extend the parity check to neighbouring name forms: mixed-case nested paths (`references/Sub/x.md`), Unicode case and normalisation variants (NFC/NFD, `ß`, Turkish dotless i), symlinked roots, and a root that exists twice in different case on a case-sensitive volume. |
| One walk | Verify there is really a single walk and no other path that publishes context (native skill materialisation, envprofile copy, install staging) outside `WalkContext`. Any publisher outside the shared walk is a finding. |
| Case-sensitive exclusions | Exclusions stay case-sensitive: is that safe? Case-sensitive exclusion means `ReadMe.md` is *not* excluded, so it is both published and scanned, which is fail-closed. Confirm that direction holds everywhere and that no exclusion path can hide a published file from the audit. |
| Ruleset 7 and cache | A real ruleset-6 allowance for an alias package is refused after the upgrade, in both hash framings. Canonical, unaffected packages still reuse correctly. |
| Behaviour change for copy | Moving `CopyContext` onto the walk must not change what canonical packages publish. Diff the published trees for the existing fixtures before and after. |
| Mutants and suite | Rerun the restored-matcher, every-Markdown and namespace-reuse mutants. Run the full default suite at the exact head with a private GOCACHE, and delete scratch afterwards. |

Verdict at the exact head; attach `BUG-261021-2qv4zj_review-r2.md` as an outcome resource.
