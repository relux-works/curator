# TASK-260720-1zntv0 review verdict

Verdict: changes requested; route to to-dev.

## Stop-ship finding: native HostPolicy does not enforce the submitted isolation or resource contract

The default policy validates the shape of HostExecution and then calls Executor.Run with only Process, discarding ReadOnlyRoots, WritableRoots, AllowedExecutables, NetworkDenied, MemoryBytes, DiskBytes, ArtifactBytes, and Processes. Both platformValidateHostExecution implementations return nil. OSExecutor applies only a direct-child context timeout and combined stdout/stderr budget. MemoryBytes and Processes are never applied; disk and artifact size are inspected only after the child exits; source read-only is only a permission-bit audit; no OS filesystem confinement, network denial, descendant allowlist, process group or Windows Job Object is installed. Therefore unexpected tool children cannot be observed or denied, Go descendants are not guaranteed to die with the deadline, source/GOROOT cannot be presented through an enforced read-only boundary, and memory/process/disk exhaustion is not bounded during execution. This conflicts with the task AC and the normative manager/security rule to fail closed when fixed environment, source separation, network denial, or process graph cannot be enforced. The attached architecture diagram also claims Policy rejects shell/network/host tools, but the implementation only validates booleans and paths.

Evidence: internal/godriver/host_policy.go lines 45-52; host_policy_unix.go lines 10-16; host_policy_windows.go lines 11-15; executor.go lines 44-65; build.go lines 123-161 and 300-346. TestBuildPresentsReadonlyNetworkAndAllResourceControls only verifies values reach a mock policy. TestRealGoV1VendoredBuildIsBoundedAndNotLaunched counts the two manager-to-go calls but cannot observe or constrain Go tool descendants. There are no native denial tests for source writes, network access, unexpected children, memory/process exhaustion, or process-tree termination.

## Required rework

Implement native adapters that apply every available filesystem/network/deadline/output/artifact-disk/memory/process/process-graph control without a launcher or other widened executable. A platform lacking a required fixed-environment, source-separation, network, or process-graph primitive must fail closed before starting Go; supported optional controls must likewise fail closed when advertised but unavailable. Enforce tree-wide deadline termination and the GOROOT tool-child allowlist. Add native helper/integration tests on each supported CI platform that attempt source/toolchain writes, network, an unexpected child, process/memory/disk/output exhaustion, and descendant survival, and prove denial/termination while the valid real vendored fixture still builds.

## Independent passing evidence

- go test ./internal/godriver -count=1
- go test -race ./internal/godriver -count=1
- make check
- rc.4 candidate contract plus bounded real Go 1.25.5 vendored build
- Windows amd64 and Linux amd64 test-binary compilation
- git diff --check and gofmt clean

These passing gates do not satisfy the missing executable host-policy AC.