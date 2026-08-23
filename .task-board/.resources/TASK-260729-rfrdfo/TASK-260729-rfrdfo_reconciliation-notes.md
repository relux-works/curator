# TASK-260729-rfrdfo — deviations from the written plan, and why

## 1. Section 8.3 rsync flags did not reproduce verifier-3's artifact

The diagnosis prescribes:

    rsync -rin --delete --exclude='.git/' <accepted>/ <candidate>/

Section 8.3 also instructs the producer to run it against the *pre-patch*
candidate first and, if the output is not byte-identical to
`candidate-source-delta-post.txt`, to reconcile the flags and report rather than
proceed. It is not identical: the prescribed flags emit **119 lines** against
the recorded **23**.

Three causes, each visible in the recorded artifact:

* The recorded itemization carries the `c` bit (`>fcsT....`). That bit only
  appears under `--checksum`. Without it, rsync compares size and mtime, so
  every file whose content matches but whose mtime differs is reported.
* `--exclude='.git/'` has a trailing slash, which matches directories only. In
  both worktrees `.git` is a *pointer file*, not a directory, so it was never
  excluded — and its content genuinely differs between the two worktrees.
* `.task-board/` and `.temp/` are agent scratch, present in one tree and not the
  other. Neither is candidate source; both are absent from the recorded output.

Reconciled command (`bin/accepted-delta.sh`), which reproduces
`candidate-source-delta-post.txt` **byte for byte**, exit 0:

    rsync -rin --checksum --delete \
      --exclude='.git' --exclude='.temp' --exclude='.task-board' \
      <accepted>/ <candidate>/

Two informational lines about unfollowed symlinks
(`.claude/skills/skill-go-testing-tools`, `.codex/skills/skill-go-testing-tools`)
are stripped; this host runs `openrsync` (rsync 2.6.9 compatible) which emits
them on stdout. They carry no delta information.

## 2. `git diff --check` was deliberately not run

Section 7 lists `git -C <candidate> diff --check -- internal/install`. It was
not run, in either tree:

* The prototype's `.git` is a pointer file reading
  `gitdir: <repo>/.git/worktrees/worktree19`. A git command run from the copy
  resolves `core.worktree` from that admin directory and therefore operates on
  the **source candidate**, not on the copy — and `git diff` refreshes the
  shared index as a side effect. That is a mutation of state this task is
  required not to touch.
* Its purpose — whitespace damage — is covered without git: `gofmt -l` is clean
  on both packages, and the generated patch was scanned for trailing whitespace
  and CR bytes on added lines.

## 3. One membership disagreement with the written exclusion list, resolved

`bin/hazards.py` derives the sequential set from source rather than trusting the
diagnosis: it flags every helper touching process-global state and closes over
intra-package calls. A first pass treated `os.Args[0]` as a hazard on the
*caller*, which put `TestConcurrentProjectInstallsPreserveBothConsumers` in the
sequential set and left `TestInstallHelperProcess` parallel — the opposite of
the sanctioned split.

The sanctioned split is correct. Reading `os.Args[0]` and spawning a child
process mutates nothing in the caller's own process, and the parent is an
expensive test worth parallelising. The hazard belongs to the test *named* in
the `-test.run=^...$` anchor: `TestInstallHelperProcess` is meaningful only as
the sole test of a re-executed binary. The rule was corrected to attribute the
hazard to the named test, after which the derivation reproduces the sanctioned
19 / 88 split exactly (`evidence/derived-exclusions.txt`).

## 4. Cosmetic: sweep subtest paths repeat the class name

Partitioning one injected class per entry, while keeping the per-class `t.Run`
that provides the `passed`/`break` semantics and keeps `injectClasses` a real
slice-valued selector, yields paths of the form

    TestFailureAtEveryTargetClassRestoresPriorStateInReverseOrder/project-hybrid-auto/10-context/10-context

The repetition is the honest consequence of an entry whose chain has length one.
Removing the inner `t.Run` would remove the generality the reviewer asked for in
correction 3, so it was kept.
