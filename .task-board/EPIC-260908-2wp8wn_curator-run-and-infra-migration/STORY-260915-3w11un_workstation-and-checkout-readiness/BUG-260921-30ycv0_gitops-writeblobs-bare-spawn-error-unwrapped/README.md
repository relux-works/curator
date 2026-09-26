# BUG-260921-30ycv0: gitops-writeblobs-bare-spawn-error-unwrapped

## Description
Residual B3 of BUG-260920-3vfwch: gitops.Extract -> writeBlobs (git -C <repo> cat-file --batch) returns a bare cmd.Start() error with cmd.Dir unset, so a product-path spawn failure surfaces as the raw fork/exec text inside install Result.Errors with no context (seen on gate 35340496757, TestDryRunEffectBindingsSeeWhatARealOperationWrites, a sequential test with one executing goroutine). Every other product git spawn wraps or classifies its error. Wrap the writeBlobs spawn error with the operation and repository context (sanitized: no path leak beyond what gitops.run already reports) and, if the product has a spawn-observability seam, record the same diagnostic facts the test fixture now records (binary/dir stat, rlimits). Test-only retries are NOT the fix here; no retry in product code without a ruling.

## Scope
internal/gitops writeBlobs error path; no lane behaviour change beyond diagnostic text

## Acceptance Criteria
writeBlobs spawn failure is classified/wrapped like gitops.run (row with an injected non-executable git); legacy goldens unchanged; no retry added
