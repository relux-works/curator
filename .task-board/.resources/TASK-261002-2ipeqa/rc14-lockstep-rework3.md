# THE ONLY CURRENT INSTRUCTION — TASK-261002-2ipeqa rework → revision 3 (orchestrator, binding)

The orchestrator dispatched the missing hosted candidate lane: curator CI run 37005296654, on your rev2 tree, with candidate_ref e3a88ced and sha256 6f832d81. Result: everything is green EXCEPT `internal/interop/environments TestConformanceSnapshotAcquisition`, on all three OSes:

    failing published case snapshot-acquisition/cases/byte-exact-snapshot is not listed in the gap ledger
    content hash sha256:500ea934… want sha256:ecca17aa…   (autocrlf=true and false)

Cause: the rc.14 corpus expects curator-content-v2 snapshot hashes, and the writer is intentionally still OFF here.

Do:
1. Add EXACTLY that case as a known-gap row for the rc.14 digest (6f832d81…) in `.github/ci/conformance-gaps.tsv`. Owner: **TASK-261002-1foyf3 (rc14-pin-and-v2-writer-cutover)**. Reason: "rc.14 expects curator-content-v2 writes; writer enabled at the rc.14 pin cut-over". Adjust the counts exactly.
   - Check, without guessing, whether any OTHER rc.14 case depends on v2 writes. Rely on the hosted run's evidence. The orchestrator saw only this one.
   - Do not add rows for anything that passes.
2. Do not flip the writer and do not move SPEC_PIN.
3. Hand off. The orchestrator will re-dispatch the candidate lane on your new tree.

Never edit LOGBOOK.md. Follow host-rules (-work); the hosted gate is the arbiter for the full matrix.
