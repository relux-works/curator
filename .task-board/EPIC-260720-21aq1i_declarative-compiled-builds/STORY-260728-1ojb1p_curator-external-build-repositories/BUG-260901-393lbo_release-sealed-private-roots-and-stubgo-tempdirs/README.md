# BUG-260901-393lbo: release-sealed-private-roots-and-stubgo-tempdirs

## Description
Two temp-directory leaks in TMPDIR. (1) internal/buildrepo/local.go: AdmitLocal seals the private object store with mode 0500 directories and the deferred os.RemoveAll then fails with EACCES; the error is discarded, so every local admission leaves a curator-buildrepo-local-* tree behind (about 3800 on the maintainer host). (2) internal/godriver/main_test.go: stubGoBinary builds the stub launcher into os.MkdirTemp under sync.Once and TestMain never removes it, one curator-stubgo-* directory per package test run.

## Scope
internal/buildrepo/local.go private-root release, internal/godriver/main_test.go stub launcher cleanup, regression tests. No behaviour change to admission proofs or sealing.

## Acceptance Criteria
AdmitLocal leaves no curator-buildrepo-local-* directory in TMPDIR on success and on failure paths; a removal failure is reported, not discarded. The godriver test binary removes its stub launcher directory when the package run ends. go build, go vet, go test ./... pass.
