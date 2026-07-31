# TASK-260728-17sclp review cycle 2

Verdict: **CHANGES REQUESTED**

Route: `to-dev`

## Remaining finding

The marker-v3 generator and its Go inventory assertion do not cover every new marker schema and semantic branch, so the first task-specific DoD and acceptance criterion are not yet met. The generated marker-v3 directory has 12 files, but no case for an empty `builds` map, a network-git substituted external record, a valid SHA-256 external record, an untagged external record, or the marker-specific structured-revision-width and substitution-identity mismatch paths. Current `main_test.go` likewise does not require those cases.

Independent full-schema probes proved these are live distinct paths: empty marker, network-git substituted marker, and SHA-256 marker instances all validate; a SHA-1 marker with a 64-hex network substitution revision rejects through the marker semantic path. Because those cases are absent from generated inventory, a regression in marker-only `buildRecordV2` or `validate_effective_source` integration can pass the current suite. Receipt and Skillfile cases exercise similar shared concepts but do not satisfy deterministic generated valid/invalid coverage for the marker-v3 schema branch itself.

Required rework: add deterministic marker-v3 cases and Go inventory requirements for (1) empty builds without top-level build_source, (2) network substitution with safe tag/branch and SHA-1/SHA-256 full revision boundaries, (3) valid SHA-256 external record, (4) untagged external record, and (5) invalid marker-local/network identity-kind mismatch plus SHA-1/64 and SHA-256/40 structured revision width. Re-run clean-from-empty two-pass regeneration and all existing gates.

## Closed prior findings and independent evidence

All four cycle-1 findings are closed: original URL/ref/identity/version false accepts now reject; boundary Unicode HTTPS, ASCII SSH/SCP, 255-byte refs, and SHA-1/SHA-256 values remain valid; Skillfile.dev-v2 external maps are optional; and the requested receipt/marker/claim negatives are present. Accepted Core/Security/Decision/profile files match byte-for-byte. Frozen manifest schemas 1-6, receipt-v1, marker-v1/v2, claim-v1/v2, Skillfile.dev-v1, all existing legacy cases, and all 37 legacy common definitions match the accepted rc.4 baseline.

Green gates: 42 schemas and 389 vector files; 15 Python tests; Go tests, vet, format, and build; make validate; clean-from-empty regeneration matching the checked-in corpus; second-pass digest 29d9be4adf618eb754d784b0974b1a6f8d26d24c96537dfd05a155289e964dd8; git diff check, clean index, and pinned HEAD. No product code was modified during review.