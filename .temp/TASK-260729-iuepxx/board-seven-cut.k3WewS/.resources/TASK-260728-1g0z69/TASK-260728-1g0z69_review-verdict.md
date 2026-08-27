# TASK-260728-1g0z69 reviewer verdict

Verdict: changes requested. Route: analysis.

## Blocking findings

1. The Go registry entry weakens the released manager-tested-family gate. The new registry defines only baseline at_least 1.23.0 and Stage A rejects only versions outside that interval. Existing profiles/manager.md lines 104-109 require every admitted Go release family to have passed the go-v1 conformance vectors and explicitly reject an otherwise unknown newer release. Under the proposed entry, an untested Go 1.99.0 satisfies the baseline. Add a manager-owned tested-release-family compatibility set or equivalent closed compatibility predicate to every registry entry, evaluate it in Stage A, define its diagnostic and guidance mapping, and add preserve/reject vectors.

2. Guidance coverage is incomplete. docs/compiled-build-toolchain-requirements.md lines 285-287 require every one of the eleven diagnostics to carry guidance_id, but lines 298-300 define reasons for only unavailable, incompatible, prerelease_unsupported, platform_unsupported, untrusted, and metadata_mismatch. No deterministic selection exists for requirement_invalid, requirement_unsatisfiable, version_undetermined, package_influence_forbidden, or changed. Define an explicit total code-to-reason mapping or extend the reason enum and totality vectors.

3. Guidance identifier lifecycle is not implementable. Lines 295-296 allow exactly toolchain plus reason plus platform, while lines 311-313 require a new guidance_id when meaning or primary-source origin changes. The tuple cannot produce a second identifier. Add an immutable revision or generation component, or redefine lifecycle/version semantics so superseded_by can name a distinct valid ID.

4. Package channel handling is contradictory. Lines 181-185 and diagnostic line 282 say package data supplying any channel or track is build_toolchain_package_influence_forbidden, while lines 239-245 retain rust-toolchain.toml channel as a compared assertion that may yield metadata mismatch or be permitted. Decision lines 133-137 and 199-204 make the same assertion-versus-selector distinction. Specify whether the forbidden category is selection/install channel data or every channel-valued field, then make the diagnostic precedence and positive/negative vectors deterministic.

5. Vector 35 conflicts with fail-fast incompatibility. Lines 255-257 and 280 require current-toolchain incompatibility to fail before cache lookup, but lines 364-365 say a cache hit with a now-incompatible toolchain is rebuilt. Split this into current resolved toolchain incompatible, where cache lookup is unreachable and the operation fails, versus a cache candidate built with a different toolchain identity while the current toolchain is compatible, where cache miss and rebuild are correct.

## Verified evidence

Task delta versus the accepted predecessor is exactly the new decision, the new reference, and one CHANGELOG bullet. Release 1.0.0-rc.5 metadata remains byte-identical. Non-mutating checks passed: project validation reported 42 schemas and 422 vector files; 29 Python tests passed; go test ./tools/... passed; go vet ./tools/... passed; gofmt -l tools produced no output; git diff --check passed. Python checks were run with the existing validation-venv because ambient python was unavailable and ambient python3 lacked jsonschema.

No code, schema, vector, release artifact, commit, or platform claim was changed by review.