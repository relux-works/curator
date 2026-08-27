# TASK-260730-1fsbqd developer commit evidence

## Handoff result

The independently accepted curator-spec rc.5 candidate was committed as one
reviewable commit in the assigned isolated worktree. This developer handoff did
not push, tag, publish a GitHub Release, sign, advance a downstream
implementation pin, or modify accepted candidate bytes.

## Exact Git identity

- Worktree:
  `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree`
- Branch: `release/curator-spec-v1.0.0-rc.5-candidate`
- Commit: `5c29c1a65bcf084c8ad27d91bcaf9d319f6146f3`
- Git tree: `78210085727ec33b79a050a807f51da253ffb0c8`
- Parent: `57c1f56846d221ecc55786bd3c2467ec32f11730`
- Subject: `Release protocol suite v1.0.0-rc.5`
- Delta: 394 paths, 30,573 insertions, 264 deletions
- Commit count after the accepted base: 1

`git fetch origin main` exited 0. The subsequent hard base gate exited 0 and
proved that worktree `HEAD`, `origin/main`, and `FETCH_HEAD` were all exactly
`57c1f56846d221ecc55786bd3c2467ec32f11730` before the commit. After the
commit, local `origin/main` remains at that base and is an ancestor of the
candidate. No tag points at the candidate commit.

## Accepted suite identity from the committed tree

The verification extracted `conformance/v1` with `git archive HEAD`, so these
values come from the committed Git tree rather than only from the working
copy:

- Protocol version: `1.0.0-rc.5`
- Manifest entries: 447
- Files under `conformance/v1`: 448 (447 listed files plus `manifest.json`)
- Manifest SHA-256:
  `b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c`
- Sorted whole-tree SHA-256:
  `e6a132157806bf747f0ad24a61bb5a9a4c8b915dac743d9465c889c0a1ad2fae`

The committed-tree verification used the accepted recipe:

```text
find . -type f -print0 | LC_ALL=C sort -z | xargs -0 shasum -a 256 | shasum -a 256
```

The shell ran with `pipefail`; the aggregate gate and all four exact assertions
exited 0.

## Validation and gate exits

Every gate below ran directly as a standalone process.

| Command / assertion | Exit | Evidence |
| --- | ---: | --- |
| `git fetch origin main` | 0 | fetched `main` into `FETCH_HEAD` |
| exact HEAD/origin/main/FETCH_HEAD base gate | 0 | all equal `57c1f56846d221ecc55786bd3c2467ec32f11730` |
| pre-commit manifest SHA-256 | 0 | exact accepted `b6f56aac...` |
| pre-commit sorted conformance tree SHA-256 | 0 | exact accepted `e6a13215...` |
| corrected pre-commit inventory assertion | 0 | 447 entries, 448 files, rc.5 |
| `git diff --cached --check` | 0 | staged delta clean |
| `git commit -m 'Release protocol suite v1.0.0-rc.5'` | 0 | exact commit above |
| task venv `python tools/validate.py` | 0 | validated 42 schemas and 447 vector files |
| task venv Python unittest discovery | 0 | 41 tests, OK |
| `go test ./tools/...` | 0 | generator package passed |
| `go vet ./tools/...` | 0 | no findings |
| `test -z "$(gofmt -l tools)"` | 0 | formatting clean |
| `go run ./tools/generate-vectors -root .` | 0 | deterministic regeneration |
| `git diff --exit-code -- conformance/v1 release/1.0.0-rc.5.json` | 0 | regeneration caused no byte drift |
| task venv `python tools/release_gate.py --version 1.0.0-rc.5 --commit HEAD` | 0 | release metadata gate passed at exact commit |
| committed `git archive HEAD` manifest/tree/count gate | 0 | exact accepted digests and counts |
| `git diff --check HEAD^ HEAD` | 0 | committed delta clean |
| `git status --porcelain=v1` | 0 | zero output after all gates |
| `git rev-list --count <base>..HEAD` | 0 | output `1` |
| `git merge-base --is-ancestor origin/main HEAD` | 0 | candidate is a direct fast-forward descendant |
| explicit rc.5 metadata assertions | 0 | exact pin, baseline, no downstream pin advancement, zero claims |

Non-green diagnostic attempts are reported separately:

- The initial board start transition exited 1 because an estimate was missing.
  A Fibonacci estimate of 3 was added (exit 0), then the required transition to
  `development` exited 0.
- The first ad hoc Python inventory-print one-liner exited 1 because of quoting
  syntax. It did not write anything; the corrected command exited 0 with the
  counts above.
- The system `python3` `jsonschema` import readiness probe exited 1
  (`ModuleNotFoundError`). The accepted task venv import exited 0 with
  jsonschema 4.25.1, and all Python/release gates used that venv.
- The first committed-archive command was rejected by the execution policy
  before it ran because its cleanup trap used `rm -rf`; it had no process exit
  and made no change. The retained task-scoped scratch version then ran and
  exited 0.

## Isolation and scope

All staging, commit, generation, validation, and Git verification commands ran
with the assigned accepted worktree as their working directory. The commit was
created with `git add -A` from that worktree; immediately before commit there
were zero unstaged tracked paths and zero untracked paths. No copy/import
command read from the dirty curator-spec primary worktree during this task.
The committed archive reproducing both independently accepted digests is the
byte-identity check.

The task's current checklist includes future reviewer and publication steps
that this developer brief explicitly forbids: independent acceptance of this
exact commit, main push, and prerelease creation. Those items are intentionally
left unchecked for the reviewer/orchestrator. Tag and GitHub Release
publication remain deferred pending a new human command.

The standalone `logbook` executable and `task-board logbook` command are not
available in this environment, so the scope/checklist anomaly is persisted
here and in task notes instead of modifying `LOGBOOK.md` directly.
