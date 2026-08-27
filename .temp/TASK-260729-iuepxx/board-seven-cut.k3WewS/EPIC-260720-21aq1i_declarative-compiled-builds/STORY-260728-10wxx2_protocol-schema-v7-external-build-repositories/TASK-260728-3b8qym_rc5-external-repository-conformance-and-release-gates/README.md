# TASK-260728-3b8qym: rc5-external-repository-conformance-and-release-gates

## Description
Build the rc.5 conformance corpus and release qualification for external build repositories after the normative text and wire schemas are accepted. Cover positive, adversarial, lifecycle, mixed-build, substitution, cache, platform-claim, and compatibility behavior and publish authoring guidance.

## Scope
Shared vectors and fixtures, validator/test harness updates, legacy guards, authoring and operator docs, release metadata, claim-v3 qualification rules, and an integrated curator-spec verification outcome. Manager-specific implementation is out of scope.

## Acceptance Criteria
Vectors cover SHA-1/SHA-256, exact tagged and untagged resolution, moved/missing/malformed refs and objects, HTTPS/SSH identity, local substitutions, submodule/LFS/link/special-file/alternate/replace/graft/promisor/filter/helper negatives, whole-snapshot audit ordering, cache hit/miss/corruption, receipt-v2/marker-v3 mixed bytes, status/repair/GC, rollback, PATH/shim checks, and signing boundaries; claim v3 lists only evidenced platforms and drivers and excludes Linux until its later story passes; full clean validation, regeneration, compatibility, and release gates pass and exact downstream pins are recorded.
