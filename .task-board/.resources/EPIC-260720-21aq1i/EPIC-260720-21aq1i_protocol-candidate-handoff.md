# Protocol candidate handoff after release-only block

Date: 2026-07-21

TASK-260720-3ag6pi remains blocked because the repository has no authorized landed curator-spec 1.0.0-rc.4 commit/ref and this goal forbids creating or staging one. The independent reviewer confirmed that candidate validation, manifest integrity, legacy compatibility, rejection safety, coverage mapping, and two byte-identical regenerations pass, but rejected the virtual-index release-check as publication evidence.

This is the release-only condition anticipated by the epic two-stage release model. Implementable manager work may consume the retained accepted candidate tree at /Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-3ag6pi/worktree through an explicit candidate conformance root. Candidate identity is protocol suite SHA sha256:70cd274150d629f16ca04f2073a8eae5a0f3fff4584f905d807775040af21eae; the complete regenerated 190-file tree digest recorded by verification is 41aa37774478c26377455877ee79ef74f8cb5cf5562ea5b1501e5c94fe9c1fa0.

Only the three Curator foundation dependencies on TASK-260720-3ag6pi are removed so implementation can proceed. This does not mark TASK-260720-3ag6pi done, does not relax its real release gate, and does not alter downstream manager/protocol release qualification or pin-audit dependencies. No commit, tag, release, checksum, signature, attestation, or pin is inferred.