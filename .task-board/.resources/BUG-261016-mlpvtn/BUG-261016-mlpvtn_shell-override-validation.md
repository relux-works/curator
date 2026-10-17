# BUG-261016-mlpvtn — self-hosted-linux-ci-flaky: validation bundle

Plain-text evidence; binary archives are refused by this public board. Expected-red logs preserve their real exit 1 and named failure.

## README.md

~~~~~~~~
# Rework validation evidence

The files record the candidate identified by source-identity.json. All commands
were run in a Story workspace whose HEAD was the recorded checkpoint, with the
six candidate files uncommitted. The remote gate separately snapshots that
working tree; its detached commit is not the Story HEAD.

mutants.py is the exact executed task-specific driver. It expects that same
checkpoint/candidate layout and an existing .temp/BUG-261016-mlpvtn/rework-2
scratch directory. Its pre-fix control reads HEAD:.github/workflows/ci.yml;
running it from a checkout whose HEAD already contains the candidate would
change the control and is not an equivalent replay. Use the checkpoint from
source-identity.json for the pre-fix workflow in any other checkout layout.

Each mutant log records a standalone full-package Go test command. The real
nonzero exits and named intended failures are in public-mutant-results.json.
Local absolute workspace paths are replaced with <workspace> in logs and
removed from recorded command arguments. All test/mutant inputs retain their
original bytes. Raw originals remain in ignored task scratch.

~~~~~~~~

## actionlint.exit

~~~~~~~~
0

~~~~~~~~

## actionlint.log

~~~~~~~~

~~~~~~~~

## build.exit

~~~~~~~~
0

~~~~~~~~

## build.log

~~~~~~~~

~~~~~~~~

## cancel-main.log

~~~~~~~~
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides
    isolation_test.go:135: valid workflow control: workflow concurrency must serialize each workflow/ref and cancel only pull requests
--- FAIL: TestLinuxCIWorkflowRejectsShellCacheOverrides (0.00s)
=== RUN   TestLinuxCIWorkflowShellOverrideBehavior
--- PASS: TestLinuxCIWorkflowShellOverrideBehavior (0.05s)
=== RUN   TestLinuxCIWorkflow
    isolation_test.go:244: workflow concurrency must serialize each workflow/ref and cancel only pull requests
--- FAIL: TestLinuxCIWorkflow (0.00s)
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/one_job_skips_initialization
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/one_setup_restores_shared_cache
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/cancel_main
    isolation_test.go:261: plant was not applied
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/split_same_ref
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/skip_gate_on_push
    isolation_test.go:264: want "unconditionally" refusal, got workflow concurrency must serialize each workflow/ref and cancel only pull requests
--- FAIL: TestLinuxCIWorkflowRejectsNarrowingPlants (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsNarrowingPlants/one_job_skips_initialization (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsNarrowingPlants/one_setup_restores_shared_cache (0.00s)
    --- FAIL: TestLinuxCIWorkflowRejectsNarrowingPlants/cancel_main (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsNarrowingPlants/split_same_ref (0.00s)
    --- FAIL: TestLinuxCIWorkflowRejectsNarrowingPlants/skip_gate_on_push (0.00s)
=== RUN   TestGoStateInitialization
--- PASS: TestGoStateInitialization (0.03s)
=== RUN   TestGoStateMissingRunnerContext
=== RUN   TestGoStateMissingRunnerContext/RUNNER_TEMP
=== RUN   TestGoStateMissingRunnerContext/GITHUB_ENV
--- PASS: TestGoStateMissingRunnerContext (0.01s)
    --- PASS: TestGoStateMissingRunnerContext/RUNNER_TEMP (0.01s)
    --- PASS: TestGoStateMissingRunnerContext/GITHUB_ENV (0.01s)
=== RUN   TestLinuxRoutingKeepsForksHosted
--- PASS: TestLinuxRoutingKeepsForksHosted (12.62s)
FAIL
FAIL	github.com/relux-works/curator/tools/ciworkflow	13.271s
FAIL

~~~~~~~~

## checker-admits-cache.log

~~~~~~~~
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/export
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/command_assignment
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/env_assignment
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/unset
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/github_env
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/quoted_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/split_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/escaped_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/continued_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/export
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/command_assignment
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/env_assignment
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/unset
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/github_env
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/quoted_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/split_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/escaped_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/continued_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/export
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/command_assignment
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/env_assignment
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/unset
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/github_env
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/quoted_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/split_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/escaped_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/continued_name
--- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides (0.05s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/export (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/command_assignment (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/env_assignment (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/unset (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/github_env (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/quoted_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/split_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/escaped_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/continued_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/export (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/command_assignment (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/env_assignment (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/unset (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/github_env (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/quoted_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/split_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/escaped_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/continued_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/export (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/command_assignment (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/env_assignment (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/unset (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/github_env (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/quoted_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/split_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/escaped_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/continued_name (0.00s)
=== RUN   TestLinuxCIWorkflowShellOverrideBehavior
--- PASS: TestLinuxCIWorkflowShellOverrideBehavior (0.05s)
=== RUN   TestLinuxCIWorkflow
--- PASS: TestLinuxCIWorkflow (0.00s)
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/one_job_skips_initialization
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/one_setup_restores_shared_cache
    isolation_test.go:264: want "shared Go cache" refusal, got <nil>
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/cancel_main
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/split_same_ref
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/skip_gate_on_push
--- FAIL: TestLinuxCIWorkflowRejectsNarrowingPlants (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsNarrowingPlants/one_job_skips_initialization (0.00s)
    --- FAIL: TestLinuxCIWorkflowRejectsNarrowingPlants/one_setup_restores_shared_cache (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsNarrowingPlants/cancel_main (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsNarrowingPlants/split_same_ref (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsNarrowingPlants/skip_gate_on_push (0.00s)
=== RUN   TestGoStateInitialization
--- PASS: TestGoStateInitialization (0.03s)
=== RUN   TestGoStateMissingRunnerContext
=== RUN   TestGoStateMissingRunnerContext/RUNNER_TEMP
=== RUN   TestGoStateMissingRunnerContext/GITHUB_ENV
--- PASS: TestGoStateMissingRunnerContext (0.01s)
    --- PASS: TestGoStateMissingRunnerContext/RUNNER_TEMP (0.01s)
    --- PASS: TestGoStateMissingRunnerContext/GITHUB_ENV (0.01s)
=== RUN   TestLinuxRoutingKeepsForksHosted
--- PASS: TestLinuxRoutingKeepsForksHosted (13.26s)
FAIL
FAIL	github.com/relux-works/curator/tools/ciworkflow	13.933s
FAIL

~~~~~~~~

## checker-admits-module-override.log

~~~~~~~~
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/export
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/command_assignment
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/env_assignment
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/unset
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/github_env
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/quoted_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/split_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/escaped_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/continued_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/export
    isolation_test.go:173: want GOMODCACHE shell override refusal, got <nil>
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/command_assignment
    isolation_test.go:173: want GOMODCACHE shell override refusal, got <nil>
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/env_assignment
    isolation_test.go:173: want GOMODCACHE shell override refusal, got <nil>
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/unset
    isolation_test.go:173: want GOMODCACHE shell override refusal, got <nil>
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/github_env
    isolation_test.go:173: want GOMODCACHE shell override refusal, got <nil>
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/quoted_name
    isolation_test.go:173: want GOMODCACHE shell override refusal, got <nil>
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/split_name
    isolation_test.go:173: want GOMODCACHE shell override refusal, got <nil>
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/escaped_name
    isolation_test.go:173: want GOMODCACHE shell override refusal, got <nil>
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/continued_name
    isolation_test.go:173: want GOMODCACHE shell override refusal, got <nil>
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/export
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/command_assignment
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/env_assignment
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/unset
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/github_env
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/quoted_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/split_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/escaped_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/continued_name
--- FAIL: TestLinuxCIWorkflowRejectsShellCacheOverrides (0.05s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/export (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/command_assignment (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/env_assignment (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/unset (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/github_env (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/quoted_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/split_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/escaped_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/continued_name (0.00s)
    --- FAIL: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/export (0.00s)
    --- FAIL: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/command_assignment (0.00s)
    --- FAIL: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/env_assignment (0.00s)
    --- FAIL: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/unset (0.00s)
    --- FAIL: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/github_env (0.00s)
    --- FAIL: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/quoted_name (0.00s)
    --- FAIL: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/split_name (0.00s)
    --- FAIL: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/escaped_name (0.00s)
    --- FAIL: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/continued_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/export (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/command_assignment (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/env_assignment (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/unset (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/github_env (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/quoted_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/split_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/escaped_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/continued_name (0.00s)
=== RUN   TestLinuxCIWorkflowShellOverrideBehavior
--- PASS: TestLinuxCIWorkflowShellOverrideBehavior (0.06s)
=== RUN   TestLinuxCIWorkflow
--- PASS: TestLinuxCIWorkflow (0.00s)
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/one_job_skips_initialization
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/one_setup_restores_shared_cache
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/cancel_main
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/split_same_ref
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/skip_gate_on_push
--- PASS: TestLinuxCIWorkflowRejectsNarrowingPlants (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsNarrowingPlants/one_job_skips_initialization (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsNarrowingPlants/one_setup_restores_shared_cache (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsNarrowingPlants/cancel_main (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsNarrowingPlants/split_same_ref (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsNarrowingPlants/skip_gate_on_push (0.00s)
=== RUN   TestGoStateInitialization
--- PASS: TestGoStateInitialization (0.05s)
=== RUN   TestGoStateMissingRunnerContext
=== RUN   TestGoStateMissingRunnerContext/RUNNER_TEMP
=== RUN   TestGoStateMissingRunnerContext/GITHUB_ENV
--- PASS: TestGoStateMissingRunnerContext (0.02s)
    --- PASS: TestGoStateMissingRunnerContext/RUNNER_TEMP (0.01s)
    --- PASS: TestGoStateMissingRunnerContext/GITHUB_ENV (0.01s)
=== RUN   TestLinuxRoutingKeepsForksHosted
--- PASS: TestLinuxRoutingKeepsForksHosted (13.07s)
FAIL
FAIL	github.com/relux-works/curator/tools/ciworkflow	13.750s
FAIL

~~~~~~~~

## checker-admits-split-name.log

~~~~~~~~
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/export
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/command_assignment
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/env_assignment
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/unset
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/github_env
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/quoted_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/split_name
    isolation_test.go:173: want GOCACHE shell override refusal, got <nil>
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/escaped_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/continued_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/export
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/command_assignment
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/env_assignment
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/unset
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/github_env
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/quoted_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/split_name
    isolation_test.go:173: want GOMODCACHE shell override refusal, got <nil>
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/escaped_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/continued_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/export
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/command_assignment
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/env_assignment
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/unset
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/github_env
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/quoted_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/split_name
    isolation_test.go:173: want GOPATH shell override refusal, got <nil>
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/escaped_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/continued_name
--- FAIL: TestLinuxCIWorkflowRejectsShellCacheOverrides (0.06s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/export (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/command_assignment (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/env_assignment (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/unset (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/github_env (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/quoted_name (0.00s)
    --- FAIL: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/split_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/escaped_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/continued_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/export (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/command_assignment (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/env_assignment (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/unset (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/github_env (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/quoted_name (0.00s)
    --- FAIL: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/split_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/escaped_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/continued_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/export (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/command_assignment (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/env_assignment (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/unset (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/github_env (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/quoted_name (0.00s)
    --- FAIL: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/split_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/escaped_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/continued_name (0.00s)
=== RUN   TestLinuxCIWorkflowShellOverrideBehavior
--- PASS: TestLinuxCIWorkflowShellOverrideBehavior (0.06s)
=== RUN   TestLinuxCIWorkflow
--- PASS: TestLinuxCIWorkflow (0.00s)
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/one_job_skips_initialization
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/one_setup_restores_shared_cache
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/cancel_main
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/split_same_ref
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/skip_gate_on_push
--- PASS: TestLinuxCIWorkflowRejectsNarrowingPlants (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsNarrowingPlants/one_job_skips_initialization (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsNarrowingPlants/one_setup_restores_shared_cache (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsNarrowingPlants/cancel_main (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsNarrowingPlants/split_same_ref (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsNarrowingPlants/skip_gate_on_push (0.00s)
=== RUN   TestGoStateInitialization
--- PASS: TestGoStateInitialization (0.06s)
=== RUN   TestGoStateMissingRunnerContext
=== RUN   TestGoStateMissingRunnerContext/RUNNER_TEMP
=== RUN   TestGoStateMissingRunnerContext/GITHUB_ENV
--- PASS: TestGoStateMissingRunnerContext (0.02s)
    --- PASS: TestGoStateMissingRunnerContext/RUNNER_TEMP (0.01s)
    --- PASS: TestGoStateMissingRunnerContext/GITHUB_ENV (0.01s)
=== RUN   TestLinuxRoutingKeepsForksHosted
--- PASS: TestLinuxRoutingKeepsForksHosted (13.14s)
FAIL
FAIL	github.com/relux-works/curator/tools/ciworkflow	14.617s
FAIL

~~~~~~~~

## diff-check.exit

~~~~~~~~
0

~~~~~~~~

## diff-check.log

~~~~~~~~

~~~~~~~~

## echo-init.log

~~~~~~~~
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides
    isolation_test.go:135: valid workflow control: go-family-qualification: private Go state must precede every Go consumer
--- FAIL: TestLinuxCIWorkflowRejectsShellCacheOverrides (0.00s)
=== RUN   TestLinuxCIWorkflowShellOverrideBehavior
    isolation_test.go:237: workflow accepted the demonstrated shell override: go-family-qualification: private Go state must precede every Go consumer
--- FAIL: TestLinuxCIWorkflowShellOverrideBehavior (0.06s)
=== RUN   TestLinuxCIWorkflow
    isolation_test.go:244: go-family-qualification: private Go state must precede every Go consumer
--- FAIL: TestLinuxCIWorkflow (0.00s)
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/one_job_skips_initialization
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/one_setup_restores_shared_cache
    isolation_test.go:264: want "shared Go cache" refusal, got go-family-qualification: private Go state must precede every Go consumer
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/cancel_main
    isolation_test.go:264: want "concurrency" refusal, got go-family-qualification: private Go state must precede every Go consumer
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/split_same_ref
    isolation_test.go:264: want "concurrency" refusal, got go-family-qualification: private Go state must precede every Go consumer
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/skip_gate_on_push
    isolation_test.go:264: want "unconditionally" refusal, got go-family-qualification: private Go state must precede every Go consumer
--- FAIL: TestLinuxCIWorkflowRejectsNarrowingPlants (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsNarrowingPlants/one_job_skips_initialization (0.00s)
    --- FAIL: TestLinuxCIWorkflowRejectsNarrowingPlants/one_setup_restores_shared_cache (0.00s)
    --- FAIL: TestLinuxCIWorkflowRejectsNarrowingPlants/cancel_main (0.00s)
    --- FAIL: TestLinuxCIWorkflowRejectsNarrowingPlants/split_same_ref (0.00s)
    --- FAIL: TestLinuxCIWorkflowRejectsNarrowingPlants/skip_gate_on_push (0.00s)
=== RUN   TestGoStateInitialization
--- PASS: TestGoStateInitialization (0.03s)
=== RUN   TestGoStateMissingRunnerContext
=== RUN   TestGoStateMissingRunnerContext/RUNNER_TEMP
=== RUN   TestGoStateMissingRunnerContext/GITHUB_ENV
--- PASS: TestGoStateMissingRunnerContext (0.01s)
    --- PASS: TestGoStateMissingRunnerContext/RUNNER_TEMP (0.01s)
    --- PASS: TestGoStateMissingRunnerContext/GITHUB_ENV (0.01s)
=== RUN   TestLinuxRoutingKeepsForksHosted
--- PASS: TestLinuxRoutingKeepsForksHosted (12.72s)
FAIL
FAIL	github.com/relux-works/curator/tools/ciworkflow	13.546s
FAIL

~~~~~~~~

## fork-bypass.log

~~~~~~~~
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/export
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/command_assignment
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/env_assignment
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/unset
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/github_env
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/quoted_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/split_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/escaped_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/continued_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/export
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/command_assignment
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/env_assignment
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/unset
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/github_env
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/quoted_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/split_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/escaped_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/continued_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/export
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/command_assignment
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/env_assignment
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/unset
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/github_env
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/quoted_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/split_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/escaped_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/continued_name
--- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides (0.05s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/export (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/command_assignment (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/env_assignment (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/unset (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/github_env (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/quoted_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/split_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/escaped_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/continued_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/export (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/command_assignment (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/env_assignment (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/unset (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/github_env (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/quoted_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/split_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/escaped_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/continued_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/export (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/command_assignment (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/env_assignment (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/unset (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/github_env (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/quoted_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/split_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/escaped_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/continued_name (0.00s)
=== RUN   TestLinuxCIWorkflowShellOverrideBehavior
--- PASS: TestLinuxCIWorkflowShellOverrideBehavior (0.05s)
=== RUN   TestLinuxCIWorkflow
--- PASS: TestLinuxCIWorkflow (0.00s)
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/one_job_skips_initialization
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/one_setup_restores_shared_cache
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/cancel_main
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/split_same_ref
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/skip_gate_on_push
--- PASS: TestLinuxCIWorkflowRejectsNarrowingPlants (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsNarrowingPlants/one_job_skips_initialization (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsNarrowingPlants/one_setup_restores_shared_cache (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsNarrowingPlants/cancel_main (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsNarrowingPlants/split_same_ref (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsNarrowingPlants/skip_gate_on_push (0.00s)
=== RUN   TestGoStateInitialization
--- PASS: TestGoStateInitialization (0.03s)
=== RUN   TestGoStateMissingRunnerContext
=== RUN   TestGoStateMissingRunnerContext/RUNNER_TEMP
=== RUN   TestGoStateMissingRunnerContext/GITHUB_ENV
--- PASS: TestGoStateMissingRunnerContext (0.01s)
    --- PASS: TestGoStateMissingRunnerContext/RUNNER_TEMP (0.01s)
    --- PASS: TestGoStateMissingRunnerContext/GITHUB_ENV (0.01s)
=== RUN   TestLinuxRoutingKeepsForksHosted
    isolation_test.go:396: go-family-qualification pull_request enabled=true fork=true ubuntu-latest: got ["self-hosted","relux-linux","curator"]
         want "ubuntu-latest"
--- FAIL: TestLinuxRoutingKeepsForksHosted (4.33s)
FAIL
FAIL	github.com/relux-works/curator/tools/ciworkflow	4.887s
FAIL

~~~~~~~~

## mutants-harness.exit

~~~~~~~~
0

~~~~~~~~

## mutants.py

~~~~~~~~
import os, pathlib, subprocess, json
root = pathlib.Path.cwd()
out = root/'.temp/BUG-261016-mlpvtn/rework-2'
workflow = (root/'.github/workflows/ci.yml').read_text()
source = (root/'tools/ciworkflow/isolation_test.go').read_text()
base = subprocess.check_output(['git','show','HEAD:.github/workflows/ci.yml'],text=True)
cache = "${{ runner.environment != 'self-hosted' || runner.os != 'Linux' }}"
expr = "(github.event_name != 'pull_request' || github.event.pull_request.head.repo.full_name == github.repository)"
results=[]
def run(name, env=None, args=None, failing=''):
    command=['go','test','-count=1','-timeout=3m','-v']+(args or [])+['./tools/ciworkflow/']
    with (out/(name+'.log')).open('w') as log:
        p=subprocess.run(command, env=env,stdout=log,stderr=subprocess.STDOUT,timeout=240)
    text=(out/(name+'.log')).read_text()
    record={'name':name,'command':command,'exit':p.returncode,'named_failure':failing,'intended_failure':p.returncode==1 and ('--- FAIL: '+failing) in text}
    results.append(record)
    (out/'mutant-results.json').write_text(json.dumps(results,indent=2)+'\n')
    print(json.dumps(record),flush=True)
def fixture(name, data, failing):
    path=out/(name+'.yml'); path.write_text(data)
    run(name,dict(os.environ,CI_WORKFLOW_UNDER_TEST=str(path)),failing=failing)
fixture('pre-fix-control',base,'TestLinuxCIWorkflow')
plants=[
('echo-init','run: bash .github/ci/isolate-go-state.sh',"run: echo 'bash .github/ci/isolate-go-state.sh'",'TestLinuxCIWorkflow'),
('restore-cache','cache: '+cache,'cache: true # '+cache,'TestLinuxCIWorkflow'),
('cancel-main',"cancel-in-progress: ${{ github.event_name == 'pull_request' }}","cancel-in-progress: ${{ github.event_name != 'workflow_dispatch' }}",'TestLinuxCIWorkflow'),
('split-ref','group: ${{ github.workflow }}-${{ github.ref }}','group: ${{ github.workflow }}-${{ github.ref }}-${{ github.run_id }}','TestLinuxCIWorkflow'),
('skip-push-gate','run: go test -count=1 ./tools/ciworkflow/',"if: github.event_name == 'pull_request'\n        run: go test -count=1 ./tools/ciworkflow/",'TestLinuxCIWorkflow'),
('fork-bypass',expr,"("+expr+" || github.event_name == 'pull_request')",'TestLinuxRoutingKeepsForksHosted'),
('shell-override','          actual="$(go version)"','          export GOCACHE="$RUNNER_TEMP/shared-build" GOMODCACHE="$RUNNER_TEMP/shared-mod" GOPATH="$RUNNER_TEMP/shared-path"\n          actual="$(go version)"','TestLinuxCIWorkflow'),
]
for name,old,new,failing in plants:
    assert old in workflow, name
    fixture(name,workflow.replace(old,new,1),failing)
script=(root/'.github/ci/isolate-go-state.sh').read_text()
old="  printf 'GOMODCACHE=%s/mod\\n' \"$state\""
new="  mkdir -p \"$RUNNER_TEMP/mod\"\n  printf 'GOMODCACHE=%s/mod\\n' \"$RUNNER_TEMP\""
assert old in script
path=out/'shared-module-cache.sh'; path.write_text(script.replace(old,new,1))
run('shared-module-cache',dict(os.environ,CI_ISOLATION_SCRIPT_UNDER_TEST=str(path)),failing='TestGoStateInitialization')
for name,old,new,failing in [
('checker-admits-cache', 's.With["cache"] != cacheCondition', 's.With["cache"] != cacheCondition && s.With["cache"] != "true"','TestLinuxCIWorkflowRejectsNarrowingPlants/one_setup_restores_shared_cache'),
('checker-admits-module-override',r'\b(GOCACHE|GOMODCACHE|GOPATH)\b',r'\b(GOCACHE|GOPATH)\b','TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/export'),
('checker-admits-split-name', 'strings.NewReplacer("\'", "", `"`, "", `\\`, "")', 'strings.NewReplacer(`\\`, "")','TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/split_name'),
]:
    assert old in source,name
    path=out/(name+'_test.go'); path.write_text(source.replace(old,new,1))
    overlay=out/(name+'.json'); overlay.write_text(json.dumps({'Replace':{str(root/'tools/ciworkflow/isolation_test.go'):str(path)}}))
    run(name,args=['-overlay',str(overlay)],failing=failing)
assert all(r['intended_failure'] for r in results),results

~~~~~~~~

## pre-fix-control.log

~~~~~~~~
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides
    isolation_test.go:135: valid workflow control: gate-selftest: private Go state must precede every Go consumer
--- FAIL: TestLinuxCIWorkflowRejectsShellCacheOverrides (0.00s)
=== RUN   TestLinuxCIWorkflowShellOverrideBehavior
    isolation_test.go:237: workflow accepted the demonstrated shell override: interop: private Go state must precede every Go consumer
--- FAIL: TestLinuxCIWorkflowShellOverrideBehavior (0.05s)
=== RUN   TestLinuxCIWorkflow
    isolation_test.go:244: go-family-qualification: private Go state must precede every Go consumer
--- FAIL: TestLinuxCIWorkflow (0.00s)
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/one_job_skips_initialization
    isolation_test.go:261: plant was not applied
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/one_setup_restores_shared_cache
    isolation_test.go:261: plant was not applied
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/cancel_main
    isolation_test.go:261: plant was not applied
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/split_same_ref
    isolation_test.go:261: plant was not applied
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/skip_gate_on_push
    isolation_test.go:261: plant was not applied
--- FAIL: TestLinuxCIWorkflowRejectsNarrowingPlants (0.00s)
    --- FAIL: TestLinuxCIWorkflowRejectsNarrowingPlants/one_job_skips_initialization (0.00s)
    --- FAIL: TestLinuxCIWorkflowRejectsNarrowingPlants/one_setup_restores_shared_cache (0.00s)
    --- FAIL: TestLinuxCIWorkflowRejectsNarrowingPlants/cancel_main (0.00s)
    --- FAIL: TestLinuxCIWorkflowRejectsNarrowingPlants/split_same_ref (0.00s)
    --- FAIL: TestLinuxCIWorkflowRejectsNarrowingPlants/skip_gate_on_push (0.00s)
=== RUN   TestGoStateInitialization
--- PASS: TestGoStateInitialization (0.04s)
=== RUN   TestGoStateMissingRunnerContext
=== RUN   TestGoStateMissingRunnerContext/RUNNER_TEMP
=== RUN   TestGoStateMissingRunnerContext/GITHUB_ENV
--- PASS: TestGoStateMissingRunnerContext (0.01s)
    --- PASS: TestGoStateMissingRunnerContext/RUNNER_TEMP (0.01s)
    --- PASS: TestGoStateMissingRunnerContext/GITHUB_ENV (0.01s)
=== RUN   TestLinuxRoutingKeepsForksHosted
--- PASS: TestLinuxRoutingKeepsForksHosted (12.77s)
FAIL
FAIL	github.com/relux-works/curator/tools/ciworkflow	13.165s
FAIL

~~~~~~~~

## public-jobs.json

~~~~~~~~
[
  {
    "job_id": 117348444502,
    "name": "Test (windows-latest)",
    "runner_alias": "runner-dcec810569",
    "self_hosted_linux": false,
    "conclusion": "success",
    "url": "https://github.com/relux-works/curator/actions/runs/39016112959/job/117348444502"
  },
  {
    "job_id": 117348444548,
    "name": "Gate self-test (windows-latest)",
    "runner_alias": "runner-6d52d8b4db",
    "self_hosted_linux": false,
    "conclusion": "success",
    "url": "https://github.com/relux-works/curator/actions/runs/39016112959/job/117348444548"
  },
  {
    "job_id": 117348444594,
    "name": "Gate self-test (macos-latest)",
    "runner_alias": "runner-6f2f4fc4dc",
    "self_hosted_linux": false,
    "conclusion": "success",
    "url": "https://github.com/relux-works/curator/actions/runs/39016112959/job/117348444594"
  },
  {
    "job_id": 117348444603,
    "name": "Test (macos-latest)",
    "runner_alias": "runner-7ef5cdc29b",
    "self_hosted_linux": false,
    "conclusion": "success",
    "url": "https://github.com/relux-works/curator/actions/runs/39016112959/job/117348444603"
  },
  {
    "job_id": 117348444644,
    "name": "Naming gate",
    "runner_alias": "runner-c55353a08c",
    "self_hosted_linux": true,
    "conclusion": "success",
    "url": "https://github.com/relux-works/curator/actions/runs/39016112959/job/117348444644"
  },
  {
    "job_id": 117348444646,
    "name": "Race (macos-latest)",
    "runner_alias": "runner-29963431fe",
    "self_hosted_linux": false,
    "conclusion": "success",
    "url": "https://github.com/relux-works/curator/actions/runs/39016112959/job/117348444646"
  },
  {
    "job_id": 117348444653,
    "name": "Go 1.27.1 driver (ubuntu-latest)",
    "runner_alias": "runner-6dd10446cf",
    "self_hosted_linux": true,
    "conclusion": "success",
    "url": "https://github.com/relux-works/curator/actions/runs/39016112959/job/117348444653"
  },
  {
    "job_id": 117348444657,
    "name": "Gate self-test (ubuntu-latest)",
    "runner_alias": "runner-5ba02dc1e7",
    "self_hosted_linux": true,
    "conclusion": "success",
    "url": "https://github.com/relux-works/curator/actions/runs/39016112959/job/117348444657"
  },
  {
    "job_id": 117348444669,
    "name": "Go 1.26.0 driver (windows-latest)",
    "runner_alias": "runner-1e347592ef",
    "self_hosted_linux": false,
    "conclusion": "success",
    "url": "https://github.com/relux-works/curator/actions/runs/39016112959/job/117348444669"
  },
  {
    "job_id": 117348444685,
    "name": "Race (ubuntu-latest)",
    "runner_alias": "runner-550dfc0c08",
    "self_hosted_linux": false,
    "conclusion": "success",
    "url": "https://github.com/relux-works/curator/actions/runs/39016112959/job/117348444685"
  },
  {
    "job_id": 117348444689,
    "name": "Go 1.27.1 driver (macos-latest)",
    "runner_alias": "runner-0f393e3742",
    "self_hosted_linux": false,
    "conclusion": "success",
    "url": "https://github.com/relux-works/curator/actions/runs/39016112959/job/117348444689"
  },
  {
    "job_id": 117348444712,
    "name": "Interop conformance gate",
    "runner_alias": "runner-c55353a08c",
    "self_hosted_linux": true,
    "conclusion": "success",
    "url": "https://github.com/relux-works/curator/actions/runs/39016112959/job/117348444712"
  },
  {
    "job_id": 117348444790,
    "name": "Test (ubuntu-latest)",
    "runner_alias": "runner-4d6e1c9351",
    "self_hosted_linux": true,
    "conclusion": "success",
    "url": "https://github.com/relux-works/curator/actions/runs/39016112959/job/117348444790"
  },
  {
    "job_id": 117348444796,
    "name": "Go 1.26.0 driver (ubuntu-latest)",
    "runner_alias": "runner-6dd10446cf",
    "self_hosted_linux": true,
    "conclusion": "success",
    "url": "https://github.com/relux-works/curator/actions/runs/39016112959/job/117348444796"
  },
  {
    "job_id": 117348444831,
    "name": "Lint",
    "runner_alias": "runner-5ba02dc1e7",
    "self_hosted_linux": true,
    "conclusion": "success",
    "url": "https://github.com/relux-works/curator/actions/runs/39016112959/job/117348444831"
  },
  {
    "job_id": 117348444911,
    "name": "Go 1.25.5 driver (windows-latest)",
    "runner_alias": "runner-50a7478ab9",
    "self_hosted_linux": false,
    "conclusion": "success",
    "url": "https://github.com/relux-works/curator/actions/runs/39016112959/job/117348444911"
  },
  {
    "job_id": 117348444936,
    "name": "Go 1.26.0 driver (macos-latest)",
    "runner_alias": "runner-b7faed0803",
    "self_hosted_linux": false,
    "conclusion": "success",
    "url": "https://github.com/relux-works/curator/actions/runs/39016112959/job/117348444936"
  },
  {
    "job_id": 117348445000,
    "name": "Go 1.25.5 driver (ubuntu-latest)",
    "runner_alias": "runner-c55353a08c",
    "self_hosted_linux": true,
    "conclusion": "success",
    "url": "https://github.com/relux-works/curator/actions/runs/39016112959/job/117348445000"
  },
  {
    "job_id": 117348445012,
    "name": "Go 1.25.5 driver (macos-latest)",
    "runner_alias": "runner-2f60fe2bd8",
    "self_hosted_linux": false,
    "conclusion": "success",
    "url": "https://github.com/relux-works/curator/actions/runs/39016112959/job/117348445012"
  },
  {
    "job_id": 117348445085,
    "name": "Go 1.27.1 driver (windows-latest)",
    "runner_alias": "runner-0022734e54",
    "self_hosted_linux": false,
    "conclusion": "success",
    "url": "https://github.com/relux-works/curator/actions/runs/39016112959/job/117348445085"
  }
]

~~~~~~~~

## public-mutant-results.json

~~~~~~~~
[
  {
    "name": "pre-fix-control",
    "command": [
      "go",
      "test",
      "-count=1",
      "-timeout=3m",
      "-v",
      "./tools/ciworkflow/"
    ],
    "exit": 1,
    "named_failure": "TestLinuxCIWorkflow",
    "intended_failure": true,
    "observed_finding": "private Go state must precede every Go consumer",
    "log_sha256": "09b98a4496fa360d1a27a82ccda96a403346a3e5841773b27915eebb2232a12e"
  },
  {
    "name": "echo-init",
    "command": [
      "go",
      "test",
      "-count=1",
      "-timeout=3m",
      "-v",
      "./tools/ciworkflow/"
    ],
    "exit": 1,
    "named_failure": "TestLinuxCIWorkflow",
    "intended_failure": true,
    "observed_finding": "private Go state must precede every Go consumer",
    "log_sha256": "98ba66b10c65dee17e6ff92d0fa43a9e697ca98c9417cf40652d57e318081b07"
  },
  {
    "name": "restore-cache",
    "command": [
      "go",
      "test",
      "-count=1",
      "-timeout=3m",
      "-v",
      "./tools/ciworkflow/"
    ],
    "exit": 1,
    "named_failure": "TestLinuxCIWorkflow",
    "intended_failure": true,
    "observed_finding": "setup-go may restore shared Go cache",
    "log_sha256": "9dd9194e83479cba52701af2a9a7e3df5f6c82b0a40003359cf3e56947d9212b"
  },
  {
    "name": "cancel-main",
    "command": [
      "go",
      "test",
      "-count=1",
      "-timeout=3m",
      "-v",
      "./tools/ciworkflow/"
    ],
    "exit": 1,
    "named_failure": "TestLinuxCIWorkflow",
    "intended_failure": true,
    "observed_finding": "workflow concurrency must serialize each workflow/ref and cancel only pull requests",
    "log_sha256": "ed27ebba61a3d9082128bffc6583a1427ff0cd26f0d1cc121fc6cc490fa0f109"
  },
  {
    "name": "split-ref",
    "command": [
      "go",
      "test",
      "-count=1",
      "-timeout=3m",
      "-v",
      "./tools/ciworkflow/"
    ],
    "exit": 1,
    "named_failure": "TestLinuxCIWorkflow",
    "intended_failure": true,
    "observed_finding": "workflow concurrency must serialize each workflow/ref and cancel only pull requests",
    "log_sha256": "e810ad5b84d01c332a6598c0b47776580067fef556e2c9a6dd944b217a2c6d09"
  },
  {
    "name": "skip-push-gate",
    "command": [
      "go",
      "test",
      "-count=1",
      "-timeout=3m",
      "-v",
      "./tools/ciworkflow/"
    ],
    "exit": 1,
    "named_failure": "TestLinuxCIWorkflow",
    "intended_failure": true,
    "observed_finding": "lint must run the isolation gate unconditionally exactly once",
    "log_sha256": "84f126bfb0350102854c4596ea6b2221050809f64e4dc71a872ccb4cb129c9f7"
  },
  {
    "name": "fork-bypass",
    "command": [
      "go",
      "test",
      "-count=1",
      "-timeout=3m",
      "-v",
      "./tools/ciworkflow/"
    ],
    "exit": 1,
    "named_failure": "TestLinuxRoutingKeepsForksHosted",
    "intended_failure": true,
    "observed_finding": "pull_request enabled=true fork=true ubuntu-latest: got [\"self-hosted\",\"relux-linux\",\"curator\"]",
    "log_sha256": "fd4e0b09ef16911019cd01462783adb10516805494622f64967d039706e823fd"
  },
  {
    "name": "shell-override",
    "command": [
      "go",
      "test",
      "-count=1",
      "-timeout=3m",
      "-v",
      "./tools/ciworkflow/"
    ],
    "exit": 1,
    "named_failure": "TestLinuxCIWorkflow",
    "intended_failure": true,
    "observed_finding": "inline shell references reserved private Go state GOCACHE",
    "log_sha256": "65cb484c0592f8ad59cf74210b523de4234aa1c6944be67e7270b11824b4cbc8"
  },
  {
    "name": "shared-module-cache",
    "command": [
      "go",
      "test",
      "-count=1",
      "-timeout=3m",
      "-v",
      "./tools/ciworkflow/"
    ],
    "exit": 1,
    "named_failure": "TestGoStateInitialization",
    "intended_failure": true,
    "observed_finding": "GOMODCACHE is shared by two jobs",
    "log_sha256": "2772b82aa699a8fd5eeba9d6ee3d031ffd77f14c44ee7898c1548c83e0693357"
  },
  {
    "name": "checker-admits-cache",
    "command": [
      "go",
      "test",
      "-count=1",
      "-timeout=3m",
      "-v",
      "-overlay",
      ".temp/BUG-261016-mlpvtn/rework-2/checker-admits-cache.json",
      "./tools/ciworkflow/"
    ],
    "exit": 1,
    "named_failure": "TestLinuxCIWorkflowRejectsNarrowingPlants/one_setup_restores_shared_cache",
    "intended_failure": true,
    "observed_finding": "want \"shared Go cache\" refusal, got <nil>",
    "log_sha256": "7949e0a03493c24957b46b291ebc767c0c56bde98d53fea2240a2a64b99ae4b8"
  },
  {
    "name": "checker-admits-module-override",
    "command": [
      "go",
      "test",
      "-count=1",
      "-timeout=3m",
      "-v",
      "-overlay",
      ".temp/BUG-261016-mlpvtn/rework-2/checker-admits-module-override.json",
      "./tools/ciworkflow/"
    ],
    "exit": 1,
    "named_failure": "TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/export",
    "intended_failure": true,
    "observed_finding": "want GOMODCACHE shell override refusal, got <nil>",
    "log_sha256": "055abc4c6228ecfda49ece290bcb64959716f331a24c65bb6b6807bba4d3076c"
  },
  {
    "name": "checker-admits-split-name",
    "command": [
      "go",
      "test",
      "-count=1",
      "-timeout=3m",
      "-v",
      "-overlay",
      ".temp/BUG-261016-mlpvtn/rework-2/checker-admits-split-name.json",
      "./tools/ciworkflow/"
    ],
    "exit": 1,
    "named_failure": "TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/split_name",
    "intended_failure": true,
    "observed_finding": "want GOCACHE shell override refusal, got <nil>",
    "log_sha256": "eafa0ce1c574debd0f6039ede165bad3a13154f14eecf996ae1ae3d4bcca58cf"
  }
]

~~~~~~~~

## public-terminal.json

~~~~~~~~
{
  "id": 39016112959,
  "head_sha": "5e1d61470ec525b73c216f50c531cf407cffd93e",
  "status": "completed",
  "conclusion": "success",
  "html_url": "https://github.com/relux-works/curator/actions/runs/39016112959",
  "created_at": "2026-10-17T08:06:20Z",
  "updated_at": "2026-10-17T08:51:07Z",
  "command": "sh scripts/remote-gate.sh",
  "exit_code": 0,
  "successful_jobs": 20,
  "optional_skips": 2,
  "self_hosted_linux_successes": 8
}

~~~~~~~~

## remote-gate.exit

~~~~~~~~
0

~~~~~~~~

## restore-cache.log

~~~~~~~~
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides
    isolation_test.go:135: valid workflow control: go-family-qualification: setup-go may restore shared Go cache
--- FAIL: TestLinuxCIWorkflowRejectsShellCacheOverrides (0.00s)
=== RUN   TestLinuxCIWorkflowShellOverrideBehavior
    isolation_test.go:237: workflow accepted the demonstrated shell override: go-family-qualification: setup-go may restore shared Go cache
--- FAIL: TestLinuxCIWorkflowShellOverrideBehavior (0.05s)
=== RUN   TestLinuxCIWorkflow
    isolation_test.go:244: go-family-qualification: setup-go may restore shared Go cache
--- FAIL: TestLinuxCIWorkflow (0.00s)
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/one_job_skips_initialization
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/one_setup_restores_shared_cache
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/cancel_main
    isolation_test.go:264: want "concurrency" refusal, got go-family-qualification: setup-go may restore shared Go cache
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/split_same_ref
    isolation_test.go:264: want "concurrency" refusal, got go-family-qualification: setup-go may restore shared Go cache
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/skip_gate_on_push
    isolation_test.go:264: want "unconditionally" refusal, got go-family-qualification: setup-go may restore shared Go cache
--- FAIL: TestLinuxCIWorkflowRejectsNarrowingPlants (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsNarrowingPlants/one_job_skips_initialization (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsNarrowingPlants/one_setup_restores_shared_cache (0.00s)
    --- FAIL: TestLinuxCIWorkflowRejectsNarrowingPlants/cancel_main (0.00s)
    --- FAIL: TestLinuxCIWorkflowRejectsNarrowingPlants/split_same_ref (0.00s)
    --- FAIL: TestLinuxCIWorkflowRejectsNarrowingPlants/skip_gate_on_push (0.00s)
=== RUN   TestGoStateInitialization
--- PASS: TestGoStateInitialization (0.03s)
=== RUN   TestGoStateMissingRunnerContext
=== RUN   TestGoStateMissingRunnerContext/RUNNER_TEMP
=== RUN   TestGoStateMissingRunnerContext/GITHUB_ENV
--- PASS: TestGoStateMissingRunnerContext (0.01s)
    --- PASS: TestGoStateMissingRunnerContext/RUNNER_TEMP (0.01s)
    --- PASS: TestGoStateMissingRunnerContext/GITHUB_ENV (0.01s)
=== RUN   TestLinuxRoutingKeepsForksHosted
--- PASS: TestLinuxRoutingKeepsForksHosted (12.64s)
FAIL
FAIL	github.com/relux-works/curator/tools/ciworkflow	13.313s
FAIL

~~~~~~~~

## scope-hygiene.json

~~~~~~~~
{
  "routes_pins_unchanged": true,
  "untouched_jobs": [
    "race",
    "test-self-hosted",
    "candidate-conformance"
  ],
  "added_text_scan": "no personal absolute paths, literal IPv4 or email addresses; manual diff inspection supplements scan"
}

~~~~~~~~

## shared-module-cache.log

~~~~~~~~
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/export
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/command_assignment
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/env_assignment
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/unset
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/github_env
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/quoted_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/split_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/escaped_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/continued_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/export
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/command_assignment
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/env_assignment
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/unset
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/github_env
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/quoted_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/split_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/escaped_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/continued_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/export
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/command_assignment
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/env_assignment
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/unset
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/github_env
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/quoted_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/split_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/escaped_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/continued_name
--- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides (0.05s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/export (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/command_assignment (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/env_assignment (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/unset (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/github_env (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/quoted_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/split_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/escaped_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/continued_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/export (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/command_assignment (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/env_assignment (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/unset (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/github_env (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/quoted_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/split_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/escaped_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/continued_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/export (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/command_assignment (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/env_assignment (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/unset (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/github_env (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/quoted_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/split_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/escaped_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/continued_name (0.00s)
=== RUN   TestLinuxCIWorkflowShellOverrideBehavior
--- PASS: TestLinuxCIWorkflowShellOverrideBehavior (0.05s)
=== RUN   TestLinuxCIWorkflow
--- PASS: TestLinuxCIWorkflow (0.00s)
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/one_job_skips_initialization
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/one_setup_restores_shared_cache
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/cancel_main
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/split_same_ref
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/skip_gate_on_push
--- PASS: TestLinuxCIWorkflowRejectsNarrowingPlants (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsNarrowingPlants/one_job_skips_initialization (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsNarrowingPlants/one_setup_restores_shared_cache (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsNarrowingPlants/cancel_main (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsNarrowingPlants/split_same_ref (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsNarrowingPlants/skip_gate_on_push (0.00s)
=== RUN   TestGoStateInitialization
    isolation_test.go:320: GOMODCACHE is shared by two jobs
--- FAIL: TestGoStateInitialization (0.03s)
=== RUN   TestGoStateMissingRunnerContext
=== RUN   TestGoStateMissingRunnerContext/RUNNER_TEMP
=== RUN   TestGoStateMissingRunnerContext/GITHUB_ENV
--- PASS: TestGoStateMissingRunnerContext (0.01s)
    --- PASS: TestGoStateMissingRunnerContext/RUNNER_TEMP (0.01s)
    --- PASS: TestGoStateMissingRunnerContext/GITHUB_ENV (0.01s)
=== RUN   TestLinuxRoutingKeepsForksHosted
--- PASS: TestLinuxRoutingKeepsForksHosted (13.13s)
FAIL
FAIL	github.com/relux-works/curator/tools/ciworkflow	13.652s
FAIL

~~~~~~~~

## shell-override.log

~~~~~~~~
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides
    isolation_test.go:135: valid workflow control: go-family-qualification: inline shell references reserved private Go state GOCACHE
--- FAIL: TestLinuxCIWorkflowRejectsShellCacheOverrides (0.00s)
=== RUN   TestLinuxCIWorkflowShellOverrideBehavior
--- PASS: TestLinuxCIWorkflowShellOverrideBehavior (0.06s)
=== RUN   TestLinuxCIWorkflow
    isolation_test.go:244: go-family-qualification: inline shell references reserved private Go state GOCACHE
--- FAIL: TestLinuxCIWorkflow (0.00s)
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/one_job_skips_initialization
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/one_setup_restores_shared_cache
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/cancel_main
    isolation_test.go:264: want "concurrency" refusal, got go-family-qualification: inline shell references reserved private Go state GOCACHE
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/split_same_ref
    isolation_test.go:264: want "concurrency" refusal, got go-family-qualification: inline shell references reserved private Go state GOCACHE
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/skip_gate_on_push
    isolation_test.go:264: want "unconditionally" refusal, got go-family-qualification: inline shell references reserved private Go state GOCACHE
--- FAIL: TestLinuxCIWorkflowRejectsNarrowingPlants (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsNarrowingPlants/one_job_skips_initialization (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsNarrowingPlants/one_setup_restores_shared_cache (0.00s)
    --- FAIL: TestLinuxCIWorkflowRejectsNarrowingPlants/cancel_main (0.00s)
    --- FAIL: TestLinuxCIWorkflowRejectsNarrowingPlants/split_same_ref (0.00s)
    --- FAIL: TestLinuxCIWorkflowRejectsNarrowingPlants/skip_gate_on_push (0.00s)
=== RUN   TestGoStateInitialization
--- PASS: TestGoStateInitialization (0.03s)
=== RUN   TestGoStateMissingRunnerContext
=== RUN   TestGoStateMissingRunnerContext/RUNNER_TEMP
=== RUN   TestGoStateMissingRunnerContext/GITHUB_ENV
--- PASS: TestGoStateMissingRunnerContext (0.01s)
    --- PASS: TestGoStateMissingRunnerContext/RUNNER_TEMP (0.01s)
    --- PASS: TestGoStateMissingRunnerContext/GITHUB_ENV (0.01s)
=== RUN   TestLinuxRoutingKeepsForksHosted
--- PASS: TestLinuxRoutingKeepsForksHosted (13.66s)
FAIL
FAIL	github.com/relux-works/curator/tools/ciworkflow	14.218s
FAIL

~~~~~~~~

## skip-push-gate.log

~~~~~~~~
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides
    isolation_test.go:135: valid workflow control: lint must run the isolation gate unconditionally exactly once
--- FAIL: TestLinuxCIWorkflowRejectsShellCacheOverrides (0.00s)
=== RUN   TestLinuxCIWorkflowShellOverrideBehavior
--- PASS: TestLinuxCIWorkflowShellOverrideBehavior (0.05s)
=== RUN   TestLinuxCIWorkflow
    isolation_test.go:244: lint must run the isolation gate unconditionally exactly once
--- FAIL: TestLinuxCIWorkflow (0.00s)
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/one_job_skips_initialization
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/one_setup_restores_shared_cache
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/cancel_main
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/split_same_ref
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/skip_gate_on_push
    isolation_test.go:264: want "unconditionally" refusal, got yaml: unmarshal errors:
          line 572: mapping key "if" already defined at line 571
--- FAIL: TestLinuxCIWorkflowRejectsNarrowingPlants (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsNarrowingPlants/one_job_skips_initialization (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsNarrowingPlants/one_setup_restores_shared_cache (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsNarrowingPlants/cancel_main (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsNarrowingPlants/split_same_ref (0.00s)
    --- FAIL: TestLinuxCIWorkflowRejectsNarrowingPlants/skip_gate_on_push (0.00s)
=== RUN   TestGoStateInitialization
--- PASS: TestGoStateInitialization (0.03s)
=== RUN   TestGoStateMissingRunnerContext
=== RUN   TestGoStateMissingRunnerContext/RUNNER_TEMP
=== RUN   TestGoStateMissingRunnerContext/GITHUB_ENV
--- PASS: TestGoStateMissingRunnerContext (0.01s)
    --- PASS: TestGoStateMissingRunnerContext/RUNNER_TEMP (0.01s)
    --- PASS: TestGoStateMissingRunnerContext/GITHUB_ENV (0.01s)
=== RUN   TestLinuxRoutingKeepsForksHosted
--- PASS: TestLinuxRoutingKeepsForksHosted (12.83s)
FAIL
FAIL	github.com/relux-works/curator/tools/ciworkflow	13.366s
FAIL

~~~~~~~~

## source-identity.json

~~~~~~~~
{
  "checkpoint": "a17e58156ac30a1629e3ff36efa259bab1555a15",
  "snapshot": "5e1d61470ec525b73c216f50c531cf407cffd93e",
  "files": {
    ".github/workflows/ci.yml": "f2d5daededfda50de425aa790338e8118334ac9002790df1b3fef45940c0378e",
    ".github/ci/isolate-go-state.sh": "1647682f0de005c2cc2def56f8ee683d2d58ce070ce20b314cf48287c6637cf0",
    "tools/ciworkflow/isolation_test.go": "844ebd3c8a724b3e8f60d05a406b962c6b9a7e19d1c5aecbe8107682a1106c6c",
    "docs/self-hosted-runner-setup.md": "4afecddfe68061dd92ce045ff72ee9abdc18eef37b95337af01ac3d11f06598d",
    "README.md": "9aa2153c622311bff029c1c556b8ca15224f321c6579ebd7cb95f5d3bc347bf9",
    "UNRESOLVED_QUESTIONS.md": "8625d747d18fdec998e7d2c4585c2654c9f471c9946c362660f8358a37c09711"
  }
}

~~~~~~~~

## split-ref.log

~~~~~~~~
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides
    isolation_test.go:135: valid workflow control: workflow concurrency must serialize each workflow/ref and cancel only pull requests
--- FAIL: TestLinuxCIWorkflowRejectsShellCacheOverrides (0.00s)
=== RUN   TestLinuxCIWorkflowShellOverrideBehavior
--- PASS: TestLinuxCIWorkflowShellOverrideBehavior (0.05s)
=== RUN   TestLinuxCIWorkflow
    isolation_test.go:244: workflow concurrency must serialize each workflow/ref and cancel only pull requests
--- FAIL: TestLinuxCIWorkflow (0.00s)
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/one_job_skips_initialization
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/one_setup_restores_shared_cache
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/cancel_main
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/split_same_ref
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/skip_gate_on_push
    isolation_test.go:264: want "unconditionally" refusal, got workflow concurrency must serialize each workflow/ref and cancel only pull requests
--- FAIL: TestLinuxCIWorkflowRejectsNarrowingPlants (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsNarrowingPlants/one_job_skips_initialization (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsNarrowingPlants/one_setup_restores_shared_cache (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsNarrowingPlants/cancel_main (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsNarrowingPlants/split_same_ref (0.00s)
    --- FAIL: TestLinuxCIWorkflowRejectsNarrowingPlants/skip_gate_on_push (0.00s)
=== RUN   TestGoStateInitialization
--- PASS: TestGoStateInitialization (0.03s)
=== RUN   TestGoStateMissingRunnerContext
=== RUN   TestGoStateMissingRunnerContext/RUNNER_TEMP
=== RUN   TestGoStateMissingRunnerContext/GITHUB_ENV
--- PASS: TestGoStateMissingRunnerContext (0.01s)
    --- PASS: TestGoStateMissingRunnerContext/RUNNER_TEMP (0.01s)
    --- PASS: TestGoStateMissingRunnerContext/GITHUB_ENV (0.01s)
=== RUN   TestLinuxRoutingKeepsForksHosted
--- PASS: TestLinuxRoutingKeepsForksHosted (12.73s)
FAIL
FAIL	github.com/relux-works/curator/tools/ciworkflow	13.617s
FAIL

~~~~~~~~

## syntax.exit

~~~~~~~~
0

~~~~~~~~

## syntax.log

~~~~~~~~

~~~~~~~~

## tests.exit

~~~~~~~~
0

~~~~~~~~

## tests.log

~~~~~~~~
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/export
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/command_assignment
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/env_assignment
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/unset
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/github_env
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/quoted_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/split_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/escaped_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/continued_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/export
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/command_assignment
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/env_assignment
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/unset
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/github_env
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/quoted_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/split_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/escaped_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/continued_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/export
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/command_assignment
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/env_assignment
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/unset
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/github_env
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/quoted_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/split_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/escaped_name
=== RUN   TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/continued_name
--- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides (0.05s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/export (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/command_assignment (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/env_assignment (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/unset (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/github_env (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/quoted_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/split_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/escaped_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/continued_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/export (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/command_assignment (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/env_assignment (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/unset (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/github_env (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/quoted_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/split_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/escaped_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/continued_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/export (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/command_assignment (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/env_assignment (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/unset (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/github_env (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/quoted_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/split_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/escaped_name (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsShellCacheOverrides/GOPATH/continued_name (0.00s)
=== RUN   TestLinuxCIWorkflowShellOverrideBehavior
--- PASS: TestLinuxCIWorkflowShellOverrideBehavior (0.05s)
=== RUN   TestLinuxCIWorkflow
--- PASS: TestLinuxCIWorkflow (0.00s)
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/one_job_skips_initialization
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/one_setup_restores_shared_cache
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/cancel_main
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/split_same_ref
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/skip_gate_on_push
--- PASS: TestLinuxCIWorkflowRejectsNarrowingPlants (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsNarrowingPlants/one_job_skips_initialization (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsNarrowingPlants/one_setup_restores_shared_cache (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsNarrowingPlants/cancel_main (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsNarrowingPlants/split_same_ref (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsNarrowingPlants/skip_gate_on_push (0.00s)
=== RUN   TestGoStateInitialization
--- PASS: TestGoStateInitialization (0.03s)
=== RUN   TestGoStateMissingRunnerContext
=== RUN   TestGoStateMissingRunnerContext/RUNNER_TEMP
=== RUN   TestGoStateMissingRunnerContext/GITHUB_ENV
--- PASS: TestGoStateMissingRunnerContext (0.01s)
    --- PASS: TestGoStateMissingRunnerContext/RUNNER_TEMP (0.01s)
    --- PASS: TestGoStateMissingRunnerContext/GITHUB_ENV (0.01s)
=== RUN   TestLinuxRoutingKeepsForksHosted
--- PASS: TestLinuxRoutingKeepsForksHosted (12.83s)
PASS
ok  	github.com/relux-works/curator/tools/ciworkflow	13.244s

~~~~~~~~

## toolchain.txt

~~~~~~~~
go version go1.25.5 darwin/arm64
v25.6.1
1.7.12
installed from Homebrew
built with go1.26.3 compiler for darwin/arm64

~~~~~~~~

## tree-identity.json

~~~~~~~~
{
  "head": "a17e58156ac30a1629e3ff36efa259bab1555a15",
  "snapshot": "5e1d61470ec525b73c216f50c531cf407cffd93e",
  "tree": "5788aaf9481965f0b22ce0257a18d6175b380b01",
  "signature": "G",
  "complete_working_tree_matches_snapshot": true
}

~~~~~~~~

## vet.exit

~~~~~~~~
0

~~~~~~~~

## vet.log

~~~~~~~~

~~~~~~~~

## wait-provenance.md

~~~~~~~~
2026-10-17T08:17:09.341851+00:00: remote-gate remains live in exec session 28040; no turn end, detached shell or background command. Original gate command is polled through bounded session calls; all Linux instances green, aggregate pending.

~~~~~~~~
