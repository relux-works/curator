# Reviewer cycle 5 verdict — changes requested

Route: analysis.

## Blocking finding

The required Go boundary probe does not isolate the semantic representability layer that the contract defines. Sections 4.2.1.2 and the probe implementation use go mod tidy and treat exit 0 as semantic acceptance and every nonzero exit as semantic rejection. go mod tidy also applies the running toolchain TooNew gate. A shape-valid, gover.Parse-valid future release is therefore falsely reported as outside Upstream.

Independent adversarial reproduction on Go 1.25.5 and Go 1.25.1: a module containing go 1.99.0 makes GOTOOLCHAIN=local GOPROXY=off GOWORK=off go mod edit -json exit 0 on both, while go mod tidy exits 1 on both with go.mod requires go >= 1.99.0. The value is representable and contract class 2 compared, but the mandated probe algorithm sets semanticAccepts false and would fail P1. This is the third layer the cycle-5 review focus asked to challenge. It makes the contract and probe obligation internally inconsistent and will create false release-gate failures as the case corpus crosses the running host version.

## Required correction

Define a semantic probe that isolates gover.Parse, either through an exactly sourced equivalent or by distinguishing TooNewError from genuine representation rejection in command output. A TooNew result must count as representable; invalid-toolchain and the patch-prerelease panic/rejection must remain rejected. Add go 1.99.0 or another future-release-on-older-host case on both available toolchains, plus an expected-red regression proving the old exit-code-only classifier fails. Update the decision, reference, probe obligation and evidence consistently. Preserve the established shape-versus-semantic boundary, P1/P2 security partition, all prior review closures, exact three-file candidate scope and frozen rc.5 bytes.

## Evidence

Independent green gates: validator exit 0 with 42 schemas and 422 vector files; 29 Python tests exit 0; Go tests, vet, gofmt and git diff --check exit 0; submitted 13+13 boundary cases on Go 1.25.5 and 1.25.1 exit 0; clean-probe-c4 regenerate-check twice and release-check VERSION=1.0.0-rc.5 exit 0 at 02e4ebc. Full directory comparison against the accepted predecessor shows exactly CHANGELOG.md, decisions/0007-compiled-build-toolchain-preflight.md and docs/compiled-build-toolchain-requirements.md differ. Existing expected-red logs correctly reject the restored patch-prerelease comparison, forced C equals Upstream, and a broken local link. The defect is not a green-gate regression; it is a missing adversarial case and an invalid layer-isolation assumption.

Recorded in LOGBOOK.md entry 2347. No candidate code or specification file was modified, staged, committed, published or pinned by the reviewer.