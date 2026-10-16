# Revision 2 validation process evidence

Candidate: `a718fda37792b57b094d9cbca8984104d8e320da`.

## Focused race suite

Actual process exit: **0**. Exact command is recorded in the re-handoff resource.

```text
ok  	github.com/relux-works/curator/internal/envprofile	25.837s
ok  	github.com/relux-works/curator/internal/managerlock	1.434s
ok  	github.com/relux-works/curator/internal/transaction	1.845s

```

## Build

Actual process exit: **0**. Exact command is recorded in the re-handoff resource.

```text
(empty output)
```

## Offline local links

Actual process exit: **0**. Exact command is recorded in the re-handoff resource.

```text
🔍 48 Total (in 5ms) 🔗 44 Unique ✅ 31 OK 🚫 0 Errors 👻 17 Excluded


```

## Whitespace

Actual process exit: **0**. Exact command is recorded in the re-handoff resource.

```text
(empty output)
```

## Shell syntax

Actual process exit: **0**. Exact command is recorded in the re-handoff resource.

```text
(empty output)
```

## Previously tracked tests retained

Actual process exit: **0**. Exact command is recorded in the re-handoff resource.

```text
(empty output)
```

## Whole source/test/configuration identity

Actual process exit: **0**. Exact command is recorded in the re-handoff resource.

```text
(empty output)
```

## Full hosted gate

Actual process exit: **0**. Exact command is recorded in the re-handoff resource.

```text
remote gate: attempt 1/3: pushing a718fda37792b57b094d9cbca8984104d8e320da as gate/STORY-261016-1wj1v9/261016-094557-18367-1
remote: 
remote: Create a pull request for 'gate/STORY-261016-1wj1v9/261016-094557-18367-1' on GitHub by visiting:        
remote:      https://github.com/relux-works/curator/pull/new/gate/STORY-261016-1wj1v9/261016-094557-18367-1        
remote: 
remote gate: run 38882703220 (https://github.com/relux-works/curator/actions/runs/38882703220)
remote gate: run 38882703220 finished: success
  Test (windows-latest): success  
  Gate self-test (windows-latest): success  
  Go 1.27.1 driver (ubuntu-latest): success  
  Test (macos-latest): success  
  Gate self-test (macos-latest): success  
  Lint: success  
  Race (ubuntu-latest): success  
  Gate self-test (ubuntu-latest): success  
  Test (ubuntu-latest): success  
  Naming gate: success  
  Go 1.25.5 driver (macos-latest): success  
  Go 1.27.1 driver (windows-latest): success  
  Go 1.26.0 driver (windows-latest): success  
  Go 1.27.1 driver (macos-latest): success  
  Go 1.25.5 driver (windows-latest): success  
  Go 1.25.5 driver (ubuntu-latest): success  
  Go 1.26.0 driver (ubuntu-latest): success  
  Go 1.26.0 driver (macos-latest): success  
  Race (macos-latest): success  
  Interop conformance gate: success  
  Optional self-hosted test: skipped  
  Candidate suite (${{ matrix.os }}): skipped  

```

## Final identity

```json
{
  "head": "dccac3199aa5c543c589c129f5ef2bc437523a2a",
  "candidate": "a718fda37792b57b094d9cbca8984104d8e320da",
  "candidate_tree": "f0b08b96c097fcf57c1927355f9874079df23e5b",
  "working_tree": "f0b08b96c097fcf57c1927355f9874079df23e5b",
  "signature": "G",
  "real_index_empty": true,
  "story_checkpoint_unchanged": true
}
```

Fresh ls-remote exited 0; temporary gate branch absent.
