# Research logbook: Windows CI duration analysis

Task: TASK-261017-0t5519 — windows-test-duration-analysis. 2026-10-17 UTC.

- Reviewed three recent green main runs and all nine Windows/macOS/Linux Test streams. Source identities, artifact digests and job IDs are in the analysis and evidence manifest.
- Windows critical path is cmd/curator (45–53 minutes); its serial phase consumes 79–81%. Entirely serial envprofile is the next bottleneck. Optimizing CLI alone cannot remove the whole gap.
- Outer Go caches hit in all Windows samples. Real native builds use intentionally private cold caches and repeated whole-toolchain verification. Linux omits native compiled behavior by qualification, so its two-minute result is not an equivalent Windows target.
- Defender contribution is unknown; no telemetry supports allocating minutes to scanning. Separate repeated fixture I/O/process/compile evidence from this unmeasured OS attribution.
- Recommend reducing repeated immutable setup across the three heavy packages, test-level sharding with complete ledger aggregation, and an 8-vCPU/32-GB/250-GB local-SSD pilot. All improvement figures are conditional estimates.
- No product, test or workflow edits; repository delta is empty. Full tables and analysis are the review deliverable. Missing surface table is declared as a brief gap, not treated as measured gate coverage.
- Evidence anomaly: initial job-log downloads were refused by gh's escape-sequence filter (exit 1); explicit allowed-output retry succeeded (exit 0). Empty initial files were not accepted.
- Verification: independent CSV/event reconciliation exited 0; missing-terminal and wrong-timing controls each exited 1 with the intended assertion. No product gate, build or mutation was run locally.
