# Assertion-preservation matrix — TestStatusReportsCompiledCurrentnessAndFailsCheck

Every assertion the pre-patch test made, and where it lives after the patch.

| # | Pre-patch assertion | Post-patch | Notes |
|---|---|---|---|
| 1 | initial real `install app` == exitOK | unchanged | preserved verbatim |
| 2 | clean `status app --json` == exitOK, 1 build row | unchanged | preserved verbatim |
| 3 | clean row full identity (skill/command/driver/build_root/source_dir/cache_key/artifact_path/target/build_source) | unchanged | preserved verbatim |
| 4 | clean `skills["build-skill"] == up-to-date` | unchanged | preserved verbatim |
| 5 | no manager-private location in the machine-readable surface | unchanged | preserved verbatim |
| 6 | clean `status app --check` == exitOK | unchanged | preserved verbatim |
| 7 | per case: JSON acquisition succeeds and decodes | preserved | now from the combined `status app --json --check` document |
| 8 | per case: exactly 1 build row | preserved verbatim | |
| 9 | per case: `State == want` | preserved verbatim | |
| 10 | per case: `Cause == cause` | preserved verbatim | |
| 11 | per case: `CacheOutcome == outcome` | preserved verbatim | |
| 12 | per case: `Detail != ""` | preserved verbatim | |
| 13 | per case: skill demotion `Skills["build-skill"] == skillState or want` | preserved verbatim | |
| 14 | per case: fail-closed `status app --check` == exitFail | preserved, folded | same invocation as #7, so the document and the exit are one run's verdict |
| 15 | per case: human text contains `state=<want>` | preserved for all 14 | via production `buildReport.Describe()` on the published row |
| 16 | per case: human text contains `cause=<cause>` when set | preserved for all cause-bearing cases | via production `buildReport.Describe()` |
| 17 | per case: plain `status app` == exitOK | preserved for 3 representative cases | no-cause marker drift, cause-bearing input drift, cache-boundary drift |
| 18 | — (new) | `strings.Contains(human, "app: "+described)` | strictly stronger than #15/#16 for those 3: pins `Describe()` to the exact line the command prints |

## Drift cases — all 14 retained

| Case | state | cause | outcome | fixture reset | plain-CLI |
|---|---|---|---|---|---|
| artifact hash recorded by the marker no longer matches the entry | build-artifact-drift | — | cache-hit | marker restore | yes |
| receipt recorded by the marker no longer matches the entry | corrupt-build-receipt | — | cache-hit | marker restore | — |
| logical key recorded by the marker was derived from another build input | build-input-drift | unattributed | cache-hit | marker restore | yes |
| logical key was derived under a build root the marker does not record | build-input-drift | build-root | cache-hit | marker restore | — |
| recorded artifact is not the one this target derives | build-input-drift | target | cache-hit | marker restore | — |
| recorded build-source identity no longer matches the frozen snapshot | build-source-drift | — | cache-hit | marker restore | — |
| marker records no build for the command the closure activates | build-command-drift | — | cache-hit | marker restore | — |
| protected cache entry cannot be interpreted | corrupt-build-cache | — | corrupt | `snapshotBuildCacheAfter` (was `reinstall`) | — |
| protected cache holds no entry for the recorded key | missing-build-artifact | — | would-preflight-and-build | `snapshotBuildCacheAfter` (was `reinstall`) | — |
| marker schema cannot be read by this manager | unsupported-marker | — | cache-hit | marker restore | — |
| marker records a build driver outside the closed set | unsupported-build-driver | — | cache-hit | marker restore | — |
| marker is not a marker document at all | invalid-marker | — | cache-hit | marker restore | — |
| build root reached agent-facing context | build-context-exposed | — | cache-hit | `restore` removes the exposure | — |
| protected cache boundary is no longer provable | untrusted-build-cache | — | would-rebuild-untrusted-cache | `restore` chmods the entries back | yes |

## Compiled-plan accounting

| | before | after |
|---|---|---|
| per-case `status --json` | 14 | 0 |
| per-case `status --check` | 14 | 0 |
| per-case `status --json --check` | 0 | 14 |
| per-case plain `status` | 14 | 3 |
| fixture-reset `install app` | 2 | 0 |
| **total compiled plans in the loop** | **44** | **17** |

## Cleanup ordering

`t.Cleanup` runs LIFO. The marker restore is now registered *before* `snapshotBuildCacheAfter`,
so the cache snapshot restore runs first and the marker restore last — a cache restore can never
put back bytes the marker restore had already corrected. No `t.Parallel`, no skips, no shared
mutable fixture concurrency, no timeout change.
