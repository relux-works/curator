# TASK-260729-2kaopg — integrated global-status candidate: provenance and gate evidence

Date: 2026-07-29
Role: developer (integration close-out cycle)
Candidate tree: `.temp/TASK-260729-2kaopg/worktree`
Accepted currentness base: `.temp/TASK-260720-1nlmvv/worktree`
Accepted fingerprint patch: `TASK-260729-1zex8r_fingerprint-cycle3.patch`

Nothing is staged or committed. No host software was installed. No pin was
moved. Every gate below ran as a standalone process with output redirected to a
file — no `tee`, no pipe — and every exit code reported is the real process exit
code.

---

## 1. Contract decision (unchanged from the accepted implementation cycle)

`curator global status` gains `--json` and `--check`, both carrying exactly the
`curator status` meaning over exactly the `cmd/curator/builds.go` vocabulary. It
derives compiled currentness by running the same read-only plan
`curator global install --dry-run` runs, because the logical build key is a
digest over the whole build input and only a plan produces the current one.
That plan resolves the machine-wide closure and passes the read-only audit and
registry gates; it runs no compiler and writes no installation target, cache
entry, or trust state.

Two deliberate deviations from the project scope, both documented in `README.md`:

1. The machine-readable document carries `alias`, `skills`, and — only when the
   closure activates compiled commands — `builds`. It carries no `path`: the
   scope has no operator-supplied root and the manager home is never published.
2. Plain `global status` keeps its historical always-report / always-exit-zero
   contract. The declared-skill map is read straight from install markers and
   never from the plan, so a scope without compiled commands prints the lines it
   always printed even when the plan refuses; the refusal is a `warning:` on
   standard error. `--check` is the only surface that turns a verdict into a
   non-zero exit, and it fails closed twice: once for every non-current code,
   once when the plan refused before it could describe every compiled command.

README: the `curator global status` contract block at lines 232–260. The
former "excluded from `TASK-260720-1nlmvv`" note is gone, replaced by the
contract it documented as missing.

---

## 2. Integrated diff provenance

### 2.1 Tree-level identity

`diff -rq` between the accepted currentness base and the integrated candidate,
excluding VCS/board/scratch trees and the two self-referential skill symlinks:

```text
README.md                                          differ   (owned)
cmd/curator/builds_test.go                         differ   (owned, call-site rewire)
cmd/curator/main.go                                differ   (owned)
cmd/curator/status_test.go                         differ   (owned, call-site rewire)
cmd/curator/global_status_test.go                  candidate only (owned, new)
internal/godriver/fingerprint.go                   differ   (accepted patch)
internal/godriver/fingerprint_equivalence_test.go  candidate only (accepted patch)
curator                                            base only (stray compiled binary, not source)
```

Exit 1 (files differ), the expected result. Raw: `logs/provenance-diffrq.txt`.
No accepted `TASK-260720-1nlmvv` file is reverted or absent — the previously
detected stale-base hazard (2396 diff lines over 15 files plus a missing
`internal/buildcache/compensation_test.go`) is repaired and stays repaired.

The candidate's own stray `curator.test` binary was removed before the gates so
the candidate tree carries source only.

### 2.2 The fingerprint delta is exactly the accepted patch — reconstructed independently

The accepted base's `internal/godriver` was copied into a scratch directory and
the accepted patch applied to that copy:

| Step | Exit | Result |
| --- | ---: | --- |
| `shasum -a 256 TASK-260729-1zex8r_fingerprint-cycle3.patch` | 0 | `a7e0906612ce6f007bfdb3776de632dd9c7a673e9b501443be5fb3eced8f1beb` — matches the cycle-3 reviewer verdict |
| `patch -p1 --dry-run` | 0 | both files apply |
| `patch -p1` | 0 | both files apply |

Reconstruction versus candidate, byte-for-byte:

| File | Reconstructed | Candidate |
| --- | --- | --- |
| `internal/godriver/fingerprint.go` | `560d0c98c665a5a83c3a6989a7b0cdcc9f26c4fb513c7688d9b1bd6e42552d1d` | identical |
| `internal/godriver/fingerprint_equivalence_test.go` | `6390e75c9848f575f2f4b50217ebd1d53481a58d349073fb0e819491b5fed484` | identical |

So the candidate's entire `internal/godriver` delta is the accepted cycle-3
patch and nothing else. Scratch: `.temp/TASK-260729-2kaopg/provenance/`.

### 2.3 The remaining delta is exactly the owned global-status surface

Preserved as `logs/owned-delta.patch`
(sha256 `1c45ac4c1f1cce15dd871b38eaf90dec0fb214280545feec14030696d48ed65d`, 923
lines):

- `README.md` — the global-status contract section only.
- `cmd/curator/main.go` — `statusScope`, `projectStatusScope`,
  `globalStatusScope`, the `statusReport`/`installedSkillDir` rescoping,
  `cmdGlobalStatus`, `globalStatusPlan`, and the usage line.
- `cmd/curator/global_status_test.go` — new, owned.
- `cmd/curator/builds_test.go`, `cmd/curator/status_test.go` — call-site rewires
  forced by the owned API change, and nothing else. The complete non-context
  delta of both files is
  `statusStores(cfg, project)` → `scope.stores` and
  `statusReport(cfg, project, …)` → `statusReport(cfg, scope, …)`, with
  `scope := projectStatusScope(cfg, project, "app")` introduced where needed.
  **No assertion in either file was altered.**

---

<!-- GATES -->
