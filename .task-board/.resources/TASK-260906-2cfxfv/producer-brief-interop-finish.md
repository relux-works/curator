# Producer brief: finish the interop coverage leaf — commit what is already verified

## What happened

Your previous run (`RUN-260907-a1a8cf`) was terminated by the launcher for exceeding its four-hour
timeout. It was **not** a failure of the work: the worktree
`/Users/iv/Developer/ReluxWorks/.worktrees/curator-interop-coverage` holds a complete, coherent and
already-verified change, and `.temp/TASK-260906-2cfxfv/logs/` holds its evidence.

What is in the tree, uncommitted:

- `internal/interop/environments/` — the five environments conformance tests moved out of
  `internal/interop`, plus `contract_test.go` and `suite_test.go`;
- `.github/ci/root-artifacts.tsv` — the new package registered for the four vector families and the
  two byte-exact trees, with the rationale in the header;
- `.github/ci/platform-cases.tsv`, `.github/ci/gate-selftest.sh`, `internal/interop/golden_test.go`.

`go build ./...` and `go vet ./internal/interop/...` are exit 0 on this tree, and
`logs/verify-final.txt` records six negative runs — each of `vectors/environments.json`,
`vectors/context-versions.json`, `vectors/context-detectors.json`,
`vectors/snapshot-acquisition.json`, `expected/environments` and `fixtures/byte-exact` removed in
turn, every one **exit=1 naming the missing artefact**. That is broader than the brief asked for and
it is the proof the leaf exists to produce.

## Your task

**Commit and hand off. Do not redo the work, and do not redesign it.**

1. Re-run the gates on the tree as it stands, each as a standalone process, and record the observed
   exit codes: `go build ./...`, `go vet ./...`, `gofmt -l cmd internal`, `golangci-lint run ./...`,
   `bash .github/ci/gate-selftest.sh`, `bash .github/ci/no-broad-suppression.sh`,
   `bash .github/ci/ledger-consistency.sh`. If any of them is red, fix only what that gate names.
2. Re-run **one** negative case end to end to confirm the recorded evidence still reproduces on the
   tree as it stands — pick `vectors/environments.json` — and one positive run against the unmodified
   candidate root. Materialize roots as **plain checkouts** verified against `manifest.json`, never
   with `git archive`. Run the lanes **sequentially**.
3. Commit in small signed commits with the human identity. **Stage named paths** — do not use
   `git add -A`, and do not commit anything under `.temp/`. **Do not write `LOGBOOK.md`.** Do not push
   and do not open a PR.
4. Attach `TASK-260906-2cfxfv_drafting-report.md`: the shape you chose with the alternatives weighed,
   the six negative runs and the positive one, the two-root partition, the ledger rows before and
   after, and the gate table. Then `task-board handoff TASK-260906-2cfxfv --role developer`.

If you find something in the tree you believe is wrong, say so in the report and fix that one thing —
but the default is that this work is finished and needs recording, not revisiting. Budget your time
accordingly: the previous run spent four hours and left nothing on the branch.
