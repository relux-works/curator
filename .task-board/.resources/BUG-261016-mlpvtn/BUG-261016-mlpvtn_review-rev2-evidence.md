# Independent revision 2 review evidence

Binary archive attachment was refused by board policy; this text resource contains the same sanitized evidence.

## actionlint-configured.log

```text

```

## base-control.log

```text
=== RUN   TestLinuxCIWorkflow
    isolation_test.go:244: go-family-qualification: private Go state must precede every Go consumer
--- FAIL: TestLinuxCIWorkflow (0.00s)
FAIL
FAIL	github.com/relux-works/curator/tools/ciworkflow	0.373s
FAIL

```

## concurrency.log

```text
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/cancel_main
=== RUN   TestLinuxCIWorkflowRejectsNarrowingPlants/split_same_ref
--- PASS: TestLinuxCIWorkflowRejectsNarrowingPlants (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsNarrowingPlants/cancel_main (0.00s)
    --- PASS: TestLinuxCIWorkflowRejectsNarrowingPlants/split_same_ref (0.00s)
PASS
ok  	github.com/relux-works/curator/tools/ciworkflow	0.403s

```

## fork-routing.log

```text
=== RUN   TestLinuxRoutingKeepsForksHosted
--- PASS: TestLinuxRoutingKeepsForksHosted (12.53s)
PASS
ok  	github.com/relux-works/curator/tools/ciworkflow	12.881s

```

## hygiene.json

```text
{"added_lines_scan": "pass", "scope": "six-file candidate delta; manual diff review also performed"}
```

## narrow-module.log

```text
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
--- PASS: TestLinuxRoutingKeepsForksHosted (12.57s)
FAIL
FAIL	github.com/relux-works/curator/tools/ciworkflow	13.343s
FAIL

```

## preservation.json

```text
{"routing_and_pins": "unchanged"}
```

## prior-export.log

```text
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
--- PASS: TestGoStateInitialization (0.04s)
=== RUN   TestGoStateMissingRunnerContext
=== RUN   TestGoStateMissingRunnerContext/RUNNER_TEMP
=== RUN   TestGoStateMissingRunnerContext/GITHUB_ENV
--- PASS: TestGoStateMissingRunnerContext (0.01s)
    --- PASS: TestGoStateMissingRunnerContext/RUNNER_TEMP (0.01s)
    --- PASS: TestGoStateMissingRunnerContext/GITHUB_ENV (0.01s)
=== RUN   TestLinuxRoutingKeepsForksHosted
--- PASS: TestLinuxRoutingKeepsForksHosted (12.68s)
FAIL
FAIL	github.com/relux-works/curator/tools/ciworkflow	13.061s
FAIL

```

## remote-sanitized.json

```text
{
  "run": {
    "id": 39016112959,
    "head_sha": "5e1d61470ec525b73c216f50c531cf407cffd93e",
    "status": "completed",
    "conclusion": "success"
  },
  "linux_jobs": [
    {
      "name": "Naming gate",
      "id": 117348444644,
      "conclusion": "success",
      "runner_registration_present": true
    },
    {
      "name": "Go 1.27.1 driver (ubuntu-latest)",
      "id": 117348444653,
      "conclusion": "success",
      "runner_registration_present": true
    },
    {
      "name": "Gate self-test (ubuntu-latest)",
      "id": 117348444657,
      "conclusion": "success",
      "runner_registration_present": true
    },
    {
      "name": "Interop conformance gate",
      "id": 117348444712,
      "conclusion": "success",
      "runner_registration_present": true
    },
    {
      "name": "Test (ubuntu-latest)",
      "id": 117348444790,
      "conclusion": "success",
      "runner_registration_present": true
    },
    {
      "name": "Go 1.26.0 driver (ubuntu-latest)",
      "id": 117348444796,
      "conclusion": "success",
      "runner_registration_present": true
    },
    {
      "name": "Lint",
      "id": 117348444831,
      "conclusion": "success",
      "runner_registration_present": true
    },
    {
      "name": "Go 1.25.5 driver (ubuntu-latest)",
      "id": 117348445000,
      "conclusion": "success",
      "runner_registration_present": true
    }
  ]
}
```

## review-attack-results.json

```text
[
  {
    "name": "prior-export",
    "exit": 1,
    "intended_result": true
  },
  {
    "name": "narrow-module",
    "exit": 1,
    "intended_result": true
  },
  {
    "name": "concurrency",
    "exit": 0,
    "intended_result": true
  },
  {
    "name": "fork-routing",
    "exit": 0,
    "intended_result": true
  },
  {
    "name": "base-control",
    "exit": 1,
    "intended_result": true
  }
]
```

## review-attacks.py

```text
import pathlib,subprocess,os,json
p=pathlib.Path(__file__).resolve().parent
w=(p/'.github/workflows/ci.yml').read_text(); s=(p/'tools/ciworkflow/isolation_test.go').read_text()
results=[]
def run(name,env=None,args=None,expect=1,needle=''):
 r=subprocess.run(['go','test','-count=1','-v']+(args or [])+['./tools/ciworkflow/'],cwd=p,env=dict(os.environ,**(env or {})),stdout=subprocess.PIPE,stderr=subprocess.STDOUT,text=True,timeout=180)
 (p/(name+'.log')).write_text(r.stdout)
 assert r.returncode==expect and needle in r.stdout,(name,r.returncode,r.stdout[-1500:])
 results.append(dict(name=name,exit=r.returncode,intended_result=True));(p/'review-attack-results.json').write_text(json.dumps(results,indent=2))
 print(name,r.returncode,flush=True)
def row(n,text):
 with (p/'surface-sweep.md').open('a') as f:f.write(n+' | held | '+text+'\n')
# Re-test the prior mechanism and token-preserving variants through the full suite.
plant=w.replace('          actual="$(go version)"','          export GOCACHE="$RUNNER_TEMP/shared-build" GOMODCACHE="$RUNNER_TEMP/shared-mod" GOPATH="$RUNNER_TEMP/shared-path"\n          actual="$(go version)"',1)
f=p/'override.yml';f.write_text(plant)
run('prior-export',{'CI_WORKFLOW_UNDER_TEST':str(f)},needle='reserved private Go state GOCACHE')
mut=p/'narrow_test.go';mut.write_text(s.replace(r'\b(GOCACHE|GOMODCACHE|GOPATH)\b',r'\b(GOCACHE|GOPATH)\b',1));overlay=p/'narrow.json';overlay.write_text(json.dumps({'Replace':{str(p/'tools/ciworkflow/isolation_test.go'):str(mut)}}))
run('narrow-module',args=['-overlay',str(overlay)],needle='want GOMODCACHE shell override refusal, got <nil>')
row('B1','All six routed jobs inspected; full suite, real shell export regression and narrowed module-name guard executed; prior finding fixed.')
run('concurrency',args=['-run','TestLinuxCIWorkflowRejectsNarrowingPlants/(cancel_main|split_same_ref)$'],expect=0,needle='PASS')
row('B2','Main-cancellation and split-ref narrowing controls rejected; exact cancellation expression only permits pull_request.')
run('fork-routing',args=['-run','TestLinuxRoutingKeepsForksHosted$'],expect=0,needle='PASS')
row('B3','216 actual-expression routing cases, plus base/candidate preservation of Race and pins.')
f=p/'base.yml';f.write_bytes(subprocess.check_output(['git','show','a17e58156ac30a1629e3ff36efa259bab1555a15:.github/workflows/ci.yml'],cwd=p))
run('base-control',{'CI_WORKFLOW_UNDER_TEST':str(f)},args=['-run','^TestLinuxCIWorkflow$'],needle='private Go state must precede every Go consumer')
row('B4','Exact base control rejects missing initialization; fixed full suite passes. ABI race remains an explicit bound.')
row('B5','Fresh GitHub API reads: exact-tree run 39016112959 success; all eight self-hosted instances successful with registrations present.')

```

## review-tests.log

```text
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
--- PASS: TestGoStateInitialization (0.04s)
=== RUN   TestGoStateMissingRunnerContext
=== RUN   TestGoStateMissingRunnerContext/RUNNER_TEMP
=== RUN   TestGoStateMissingRunnerContext/GITHUB_ENV
--- PASS: TestGoStateMissingRunnerContext (0.01s)
    --- PASS: TestGoStateMissingRunnerContext/RUNNER_TEMP (0.01s)
    --- PASS: TestGoStateMissingRunnerContext/GITHUB_ENV (0.01s)
=== RUN   TestLinuxRoutingKeepsForksHosted
--- PASS: TestLinuxRoutingKeepsForksHosted (12.75s)
PASS
ok  	github.com/relux-works/curator/tools/ciworkflow	13.328s

```

## surface-sweep.md

```text
B1 | held | All six routed jobs inspected; full suite, real shell export regression and narrowed module-name guard executed; prior finding fixed.
B2 | held | Main-cancellation and split-ref narrowing controls rejected; exact cancellation expression only permits pull_request.
B3 | held | 216 actual-expression routing cases, plus base/candidate preservation of Race and pins.
B4 | held | Exact base control rejects missing initialization; fixed full suite passes. ABI race remains an explicit bound.
B5 | held | Fresh GitHub API reads: exact-tree run 39016112959 success; all eight self-hosted instances successful with registrations present.
B6 | held | Added-line sensitive-path/IP/email/known-host scan and complete diff review passed; no new identifiers.

```
