# BUG-261017-ygfdal review execution evidence

Candidate tree: `103755136899324f1401da18113bd6cc5c17f3e7`.

Commands: `go test -count=1 -timeout=3m ./internal/scriptworker -run ^TestLandlockProbeModeTerminatesAffectedTestBinaries$` (exit 0); baseline overlay with same final regression and install/owned filter (exit 1, intended behavioral failure); `go test -count=1 -timeout=3m ./cmd/curator -run ^TestProductionBinaryDispatchesLandlockProbe$` (exit 0). Baseline overlay replaces landlock_probe.go with an empty package and restores base landlock_linux.go; no test bytes changed. Production build and git diff --check exit 0.

## dispatch.log

```text
ok  	github.com/relux-works/curator/internal/scriptworker	5.570s

```

## baseline-red.log

```text
--- FAIL: TestLandlockProbeModeTerminatesAffectedTestBinaries (3.70s)
    --- FAIL: TestLandlockProbeModeTerminatesAffectedTestBinaries/internal/install (3.70s)
        --- FAIL: TestLandlockProbeModeTerminatesAffectedTestBinaries/internal/install/owned (3.00s)
            probe_reexec_test.go:47: probe ran tests or hung: context deadline exceeded; output --- FAIL: TestProjectManifestABAAroundTheReadRestartsClosure (0.01s)
                    aba_test.go:88: git [init -q]: exec: "git": executable file not found in $PATH
                        
                        gitfixture diagnostic (spawn failed, no git process ran):
                          argv0 "git" resolved: unresolved: exec: "git": executable file not found in $PATH
                          dir "<temporary-path>": lstat: mode=drwxr-xr-x; stat: mode=drwxr-xr-x size=64
                          cwd: <local-path>
FAIL
FAIL	github.com/relux-works/curator/internal/scriptworker	3.984s
FAIL

```

## production.log

```text
ok  	github.com/relux-works/curator/cmd/curator	63.990s

```

## packages.txt

```text
github.com/relux-works/curator/cmd/curator.test
github.com/relux-works/curator/internal/crossconformance.test
github.com/relux-works/curator/internal/install.test
github.com/relux-works/curator/internal/install/atomicity.test
github.com/relux-works/curator/internal/scriptworker.test

```

## stream-verification.json

```text
{
  "stream_sha256": "aa277143d2e1112ae4f32ab71acf0ed04a912cf0a8bf1f4b796850d221e1ae52",
  "terminal_packages": 83,
  "required_tests": 7,
  "binary_cases": 20,
  "narrowing_mutants": 3,
  "fail_records": 0,
  "all_started_packages_terminated": true
}

```

## free-hunt.json

```text
[
  {
    "argc": 2,
    "exit": 1,
    "silent": true
  },
  {
    "argc": 3,
    "exit": 1,
    "silent": true
  },
  {
    "argc": 4,
    "exit": 1,
    "silent": true
  },
  {
    "argc": 4,
    "exit": 1,
    "silent": true
  },
  {
    "argc": 130,
    "exit": 1,
    "silent": true
  },
  {
    "help_exit": 0,
    "hidden_token_absent": true
  }
]

```
