# TASK-261017-sxydvh — authority/path regression

Candidate: 528618358696a3d498ac456667ad19b87f6fc8bd (signature G). Full local verification is green: 111 commands, 85 packages, 3476 top-level terminals (3445 pass, 31 skip). Full hosted gate is green: https://github.com/relux-works/curator/actions/runs/39077229686 (20 success, 2 conditional skips); controller exit 0. Full current logs are attached in the task evidence bundles.

The production Gate probe below failed with exit 1 before the correction: a malformed path escape erased a real host. Its paired reserved-host control passed. Authority-only extraction fixes the case. The corrected focused network suite exited 0 and killed all six narrowing mutants at their named assertions. Build and lint also exited 0.

## path-contact-probe

{
  "name": "path-contact-probe",
  "args": [
    "go",
    "test",
    "-overlay",
    ".temp/TASK-261017-sxydvh/path-contact-probe-overlay.json",
    "-count=1",
    "-run",
    "^TestNetworkContactNormalizationBoundaries$/^percent-path-",
    "-v",
    "./internal/audit"
  ],
  "exit": 1,
  "seconds": 5.73,
  "environment": "isolated-v2"
}

```text
=== RUN   TestNetworkContactNormalizationBoundaries
=== RUN   TestNetworkContactNormalizationBoundaries/percent-path-real
    network_contacts_test.go:187: URL classification: []
=== RUN   TestNetworkContactNormalizationBoundaries/percent-path-reserved
--- FAIL: TestNetworkContactNormalizationBoundaries (0.30s)
    --- FAIL: TestNetworkContactNormalizationBoundaries/percent-path-real (0.12s)
    --- PASS: TestNetworkContactNormalizationBoundaries/percent-path-reserved (0.10s)
FAIL
FAIL	github.com/relux-works/curator/internal/audit	0.596s
FAIL

```

## authority-fix-focused

{
  "name": "authority-fix-focused",
  "args": [
    "go",
    "test",
    "-count=1",
    "-timeout=8m",
    "-run",
    "^TestNetworkContact",
    "-v",
    "./internal/audit"
  ],
  "exit": 0,
  "seconds": 47.92,
  "environment": "isolated-v2"
}

```text
=== RUN   TestNetworkContactNarrowingMutants
=== RUN   TestNetworkContactNarrowingMutants/full-url-percent-path
    network_contacts_mutants_test.go:59: expected-red go test exit 1:
        === RUN   TestNetworkContactNormalizationBoundaries
        === RUN   TestNetworkContactNormalizationBoundaries/percent-path-real
            network_contacts_test.go:188: URL classification: []
        --- FAIL: TestNetworkContactNormalizationBoundaries (0.13s)
            --- FAIL: TestNetworkContactNormalizationBoundaries/percent-path-real (0.08s)
        FAIL
        FAIL	github.com/relux-works/curator/internal/audit	0.457s
        FAIL
=== RUN   TestNetworkContactNarrowingMutants/contains-test
    network_contacts_mutants_test.go:59: expected-red go test exit 1:
        === RUN   TestNetworkContactVectors
        === RUN   TestNetworkContactVectors/lookalike-nested-test
            network_contacts_test.go:64: Gate pass=false: warnings=[] errors=[]
        --- FAIL: TestNetworkContactVectors (0.06s)
            --- FAIL: TestNetworkContactVectors/lookalike-nested-test (0.06s)
        FAIL
        FAIL	github.com/relux-works/curator/internal/audit	0.368s
        FAIL
=== RUN   TestNetworkContactNarrowingMutants/exempt-example-com
    network_contacts_mutants_test.go:59: expected-red go test exit 1:
        === RUN   TestNetworkContactVectors
        === RUN   TestNetworkContactVectors/documentation-com
            network_contacts_test.go:64: Gate pass=false: warnings=[] errors=[]
        --- FAIL: TestNetworkContactVectors (0.21s)
            --- FAIL: TestNetworkContactVectors/documentation-com (0.21s)
        FAIL
        FAIL	github.com/relux-works/curator/internal/audit	0.526s
        FAIL
=== RUN   TestNetworkContactNarrowingMutants/exempt-bare-test
    network_contacts_mutants_test.go:59: expected-red go test exit 1:
        === RUN   TestNetworkContactVectors
        === RUN   TestNetworkContactVectors/bare-test
            network_contacts_test.go:64: Gate pass=false: warnings=[] errors=[]
        --- FAIL: TestNetworkContactVectors (0.24s)
            --- FAIL: TestNetworkContactVectors/bare-test (0.24s)
        FAIL
        FAIL	github.com/relux-works/curator/internal/audit	0.608s
        FAIL
=== RUN   TestNetworkContactNarrowingMutants/reserved-clears-file
    network_contacts_mutants_test.go:59: expected-red go test exit 1:
        === RUN   TestNetworkContactVectors
        === RUN   TestNetworkContactVectors/mixed-reserved-first
            network_contacts_test.go:64: Gate pass=false: warnings=[] errors=[]
        --- FAIL: TestNetworkContactVectors (0.30s)
            --- FAIL: TestNetworkContactVectors/mixed-reserved-first (0.30s)
        FAIL
        FAIL	github.com/relux-works/curator/internal/audit	0.645s
        FAIL
=== RUN   TestNetworkContactNarrowingMutants/reuse-ruleset-three
    network_contacts_mutants_test.go:59: expected-red go test exit 1:
        === RUN   TestNetworkContactCacheInvalidatesRulesetThree
        === RUN   TestNetworkContactCacheInvalidatesRulesetThree/v1/api.test
            network_contacts_test.go:149: stale ruleset reused: report={Skill:skill-a ContentSHA256:sha256:deef286f5b413ebff22d8bc66abd1bd6f7bb5fa035cd1b04af3baf51e36e9f11 Findings:[{ID:audit.capability.network-undeclared Severity:high File:scripts/tool Evidence:contacts api.test, which no capabilities.network entry covers Verifiable:true}] Decision:block Revoked:false Revocation: CacheHit:true ScriptPolicies:[]} error=<nil>, want hit=false decision=allow
        === RUN   TestNetworkContactCacheInvalidatesRulesetThree/v1/api.example.com
            network_contacts_test.go:149: stale ruleset reused: report={Skill:skill-a ContentSHA256:sha256:d41bb09ff88a20372b9fa89f273c24a7a67cf429141918ef108bf4245100ccbe Findings:[] Decision:allow Revoked:false Revocation: CacheHit:true ScriptPolicies:[]} error=<nil>, want hit=false decision=block
        === RUN   TestNetworkContactCacheInvalidatesRulesetThree/v2/api.test
            network_contacts_test.go:149: stale ruleset reused: report={Skill:skill-a ContentSHA256:sha256:e9ee9db15d7b4c804de1c66ba7345b0bdf35ccdd1cbb00ecbf339c0a26cdca56 Findings:[{ID:audit.capability.network-undeclared Severity:high File:scripts/tool Evidence:contacts api.test, which no capabilities.network entry covers Verifiable:true}] Decision:block Revoked:false Revocation: CacheHit:true ScriptPolicies:[]} error=<nil>, want hit=false decision=allow
        === RUN   TestNetworkContactCacheInvalidatesRulesetThree/v2/api.example.com
            network_contacts_test.go:149: stale ruleset reused: report={Skill:skill-a ContentSHA256:sha256:82f0f7718d184b66aeff82613caa7ce7718e18751593744121d51723d4b5720d Findings:[] Decision:allow Revoked:false Revocation: CacheHit:true ScriptPolicies:[]} error=<nil>, want hit=false decision=block
        --- FAIL: TestNetworkContactCacheInvalidatesRulesetThree (0.73s)
            --- FAIL: TestNetworkContactCacheInvalidatesRulesetThree/v1/api.test (0.15s)
            --- FAIL: TestNetworkContactCacheInvalidatesRulesetThree/v1/api.example.com (0.17s)
            --- FAIL: TestNetworkContactCacheInvalidatesRulesetThree/v2/api.test (0.17s)
            --- FAIL: TestNetworkContactCacheInvalidatesRulesetThree/v2/api.example.com (0.24s)
        FAIL
        FAIL	github.com/relux-works/curator/internal/audit	1.056s
        FAIL
--- PASS: TestNetworkContactNarrowingMutants (30.91s)
    --- PASS: TestNetworkContactNarrowingMutants/full-url-percent-path (5.50s)
    --- PASS: TestNetworkContactNarrowingMutants/contains-test (4.98s)
    --- PASS: TestNetworkContactNarrowingMutants/exempt-example-com (4.92s)
    --- PASS: TestNetworkContactNarrowingMutants/exempt-bare-test (5.51s)
    --- PASS: TestNetworkContactNarrowingMutants/reserved-clears-file (5.27s)
    --- PASS: TestNetworkContactNarrowingMutants/reuse-ruleset-three (4.72s)
=== RUN   TestNetworkContactVectors
=== RUN   TestNetworkContactVectors/loopback-localhost
=== RUN   TestNetworkContactVectors/loopback-ipv4
=== RUN   TestNetworkContactVectors/loopback-ipv6
=== RUN   TestNetworkContactVectors/loopback-case-dot
=== RUN   TestNetworkContactVectors/reserved-test
=== RUN   TestNetworkContactVectors/reserved-example
=== RUN   TestNetworkContactVectors/reserved-invalid
=== RUN   TestNetworkContactVectors/reserved-localhost
=== RUN   TestNetworkContactVectors/reserved-nested
=== RUN   TestNetworkContactVectors/reserved-case-dot
=== RUN   TestNetworkContactVectors/reserved-idn
=== RUN   TestNetworkContactVectors/reserved-punycode
=== RUN   TestNetworkContactVectors/lookalike-nested-test
=== RUN   TestNetworkContactVectors/lookalike-nested-example
=== RUN   TestNetworkContactVectors/lookalike-hyphen
=== RUN   TestNetworkContactVectors/bare-test
=== RUN   TestNetworkContactVectors/bare-example
=== RUN   TestNetworkContactVectors/bare-invalid
=== RUN   TestNetworkContactVectors/bare-test-dot
=== RUN   TestNetworkContactVectors/lookalike-idn-label
=== RUN   TestNetworkContactVectors/lookalike-punycode-label
=== RUN   TestNetworkContactVectors/lookalike-idn-real-tld
=== RUN   TestNetworkContactVectors/lookalike-case-dot
=== RUN   TestNetworkContactVectors/documentation-com
=== RUN   TestNetworkContactVectors/documentation-net
=== RUN   TestNetworkContactVectors/documentation-org
=== RUN   TestNetworkContactVectors/documentation-subdomain-com
=== RUN   TestNetworkContactVectors/documentation-subdomain-net
=== RUN   TestNetworkContactVectors/documentation-subdomain-org
=== RUN   TestNetworkContactVectors/documentation-case-dot
=== RUN   TestNetworkContactVectors/canary-undeclared
=== RUN   TestNetworkContactVectors/canary-declared
=== RUN   TestNetworkContactVectors/mixed-reserved-first
=== RUN   TestNetworkContactVectors/mixed-real-first
=== RUN   TestNetworkContactVectors/mixed-cross-file
=== RUN   TestNetworkContactVectors/declared-exact
=== RUN   TestNetworkContactVectors/declared-glob
=== RUN   TestNetworkContactVectors/declared-other-host
=== RUN   TestNetworkContactVectors/missing-manifest
=== RUN   TestNetworkContactVectors/missing-network-declaration
--- PASS: TestNetworkContactVectors (8.71s)
    --- PASS: TestNetworkContactVectors/loopback-localhost (0.30s)
    --- PASS: TestNetworkContactVectors/loopback-ipv4 (0.45s)
    --- PASS: TestNetworkContactVectors/loopback-ipv6 (0.31s)
    --- PASS: TestNetworkContactVectors/loopback-case-dot (0.34s)
    --- PASS: TestNetworkContactVectors/reserved-test (0.39s)
    --- PASS: TestNetworkContactVectors/reserved-example (0.37s)
    --- PASS: TestNetworkContactVectors/reserved-invalid (0.37s)
    --- PASS: TestNetworkContactVectors/reserved-localhost (0.11s)
    --- PASS: TestNetworkContactVectors/reserved-nested (0.12s)
    --- PASS: TestNetworkContactVectors/reserved-case-dot (0.14s)
    --- PASS: TestNetworkContactVectors/reserved-idn (0.16s)
    --- PASS: TestNetworkContactVectors/reserved-punycode (0.03s)
    --- PASS: TestNetworkContactVectors/lookalike-nested-test (0.03s)
    --- PASS: TestNetworkContactVectors/lookalike-nested-example (0.02s)
    --- PASS: TestNetworkContactVectors/lookalike-hyphen (0.07s)
    --- PASS: TestNetworkContactVectors/bare-test (0.10s)
    --- PASS: TestNetworkContactVectors/bare-example (0.10s)
    --- PASS: TestNetworkContactVectors/bare-invalid (0.04s)
    --- PASS: TestNetworkContactVectors/bare-test-dot (0.06s)
    --- PASS: TestNetworkContactVectors/lookalike-idn-label (0.05s)
    --- PASS: TestNetworkContactVectors/lookalike-punycode-label (0.02s)
    --- PASS: TestNetworkContactVectors/lookalike-idn-real-tld (0.11s)
    --- PASS: TestNetworkContactVectors/lookalike-case-dot (0.16s)
    --- PASS: TestNetworkContactVectors/documentation-com (0.19s)
    --- PASS: TestNetworkContactVectors/documentation-net (0.11s)
    --- PASS: TestNetworkContactVectors/documentation-org (0.15s)
    --- PASS: TestNetworkContactVectors/documentation-subdomain-com (0.04s)
    --- PASS: TestNetworkContactVectors/documentation-subdomain-net (0.14s)
    --- PASS: TestNetworkContactVectors/documentation-subdomain-org (0.08s)
    --- PASS: TestNetworkContactVectors/documentation-case-dot (0.11s)
    --- PASS: TestNetworkContactVectors/canary-undeclared (0.22s)
    --- PASS: TestNetworkContactVectors/canary-declared (0.29s)
    --- PASS: TestNetworkContactVectors/mixed-reserved-first (0.28s)
    --- PASS: TestNetworkContactVectors/mixed-real-first (0.40s)
    --- PASS: TestNetworkContactVectors/mixed-cross-file (0.58s)
    --- PASS: TestNetworkContactVectors/declared-exact (0.30s)
    --- PASS: TestNetworkContactVectors/declared-glob (0.42s)
    --- PASS: TestNetworkContactVectors/declared-other-host (0.46s)
    --- PASS: TestNetworkContactVectors/missing-manifest (0.54s)
    --- PASS: TestNetworkContactVectors/missing-network-declaration (0.55s)
=== RUN   TestNetworkContactCacheInvalidatesRulesetThree
=== RUN   TestNetworkContactCacheInvalidatesRulesetThree/v1/api.test
=== RUN   TestNetworkContactCacheInvalidatesRulesetThree/v1/api.example.com
=== RUN   TestNetworkContactCacheInvalidatesRulesetThree/v2/api.test
=== RUN   TestNetworkContactCacheInvalidatesRulesetThree/v2/api.example.com
--- PASS: TestNetworkContactCacheInvalidatesRulesetThree (0.99s)
    --- PASS: TestNetworkContactCacheInvalidatesRulesetThree/v1/api.test (0.19s)
    --- PASS: TestNetworkContactCacheInvalidatesRulesetThree/v1/api.example.com (0.24s)
    --- PASS: TestNetworkContactCacheInvalidatesRulesetThree/v2/api.test (0.26s)
    --- PASS: TestNetworkContactCacheInvalidatesRulesetThree/v2/api.example.com (0.30s)
=== RUN   TestNetworkContactNormalizationBoundaries
=== RUN   TestNetworkContactNormalizationBoundaries/reserved-port
=== RUN   TestNetworkContactNormalizationBoundaries/percent-path-real
=== RUN   TestNetworkContactNormalizationBoundaries/percent-path-reserved
=== RUN   TestNetworkContactNormalizationBoundaries/loopback-port
=== RUN   TestNetworkContactNormalizationBoundaries/real-ipv6
=== RUN   TestNetworkContactNormalizationBoundaries/credentials-are-not-host
=== RUN   TestNetworkContactNormalizationBoundaries/port-is-not-host
=== RUN   TestNetworkContactNormalizationBoundaries/one-dot-only
=== RUN   TestNetworkContactNormalizationBoundaries/invalid-idna
--- PASS: TestNetworkContactNormalizationBoundaries (1.50s)
    --- PASS: TestNetworkContactNormalizationBoundaries/reserved-port (0.19s)
    --- PASS: TestNetworkContactNormalizationBoundaries/percent-path-real (0.16s)
    --- PASS: TestNetworkContactNormalizationBoundaries/percent-path-reserved (0.16s)
    --- PASS: TestNetworkContactNormalizationBoundaries/loopback-port (0.19s)
    --- PASS: TestNetworkContactNormalizationBoundaries/real-ipv6 (0.20s)
    --- PASS: TestNetworkContactNormalizationBoundaries/credentials-are-not-host (0.12s)
    --- PASS: TestNetworkContactNormalizationBoundaries/port-is-not-host (0.06s)
    --- PASS: TestNetworkContactNormalizationBoundaries/one-dot-only (0.13s)
    --- PASS: TestNetworkContactNormalizationBoundaries/invalid-idna (0.14s)
PASS
ok  	github.com/relux-works/curator/internal/audit	42.455s

```
