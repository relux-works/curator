# Independent rc.8 pin review verdict

Verdict: accepted.

The authoritative Curator surfaces pin v1.0.0-rc.8, commit f8c405aa3ad0a39d260c2ed93684e55c5a346359, annotated tag object ad247840292487d5d88ac44331798b6b4182a79f, conformance manifest SHA-256 d14e3a16bb4a01ff282791f08e3aefa269210234f41072beae6fe59b642595a1, release metadata SHA-256 293f101d10665061aa049efa72141f9e3c5d608bbde300e882f6e3e095e31ede, and preserved rc.7 metadata SHA-256 e5872ee4dd207bf6b190d8c8be15a9366d9c1e3638047ea983620b97c9f84d5d. CI checks out the immutable commit and runs the verifier in every released-suite job.

Independent review confirmed negative coverage for mutable, mismatched, unknown-version, conformance-digest-drifted, and claim-inflating pins. Strict metadata decoding, exact byte digests, suite cross-digests, assurance policy identities, empty published claim sets, and the immutable rc.7 compatibility mapping fail closed. The logical receipt key remains distinct from the assurance-bound protected address, and the compiled dry-run vector proves the published logical identity plus no persistent mutation.

Independent commands exited 0: exact curator-spec-pin verifier; SHA-256 checks for manifest and rc.8 metadata; focused buildrepo and pin CLI tests; focused buildcache and install compatibility tests; deterministic diff -qr against the regenerated rc.8 tree; git diff --check; gofmt; go vet ./...; go build ./...; golangci-lint v2.12.2 with 0 issues; and CURATOR_CONFORMANCE_ROOT at exact rc.8 go test -count=1 -timeout 30m ./.... Producer evidence additionally records full race, CI protocol/platform, gate self-test, canonical verifier, and binary-deny gates as passing.

No code, test, documentation, release-pin, compatibility, architecture, or unreviewed semantic-drift defect was found.