# Logbook entry (reviewer, TASK-260905-3r30t1, 2026-09-05)

Finding: `gitops.writeBlobs` (curator a46abc80) deadlocks whenever `Extract` returns early after `git cat-file --batch`
has more than a pipe buffer of output queued: `cmd.Wait()` waits for exit before closing `StdoutPipe`, git blocks on
write. The size gate is therefore a hang in production for any blob > 512 MiB; the duplicate-platform-path gate hangs
when a large blob follows the collision. Committed negative tests use 1-byte blobs and never hit it. General lesson for
this repo: any `StdoutPipe` consumer that can return before EOF needs kill/drain before `Wait`, and refusal tests must
use payloads larger than the OS pipe buffer.
