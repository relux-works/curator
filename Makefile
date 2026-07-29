GO ?= go
VERSION ?= dev
LDFLAGS := -X github.com/relux-works/curator/internal/version.value=$(VERSION)

# Compiled-build CI gates. EVIDENCE is where each gate writes its raw
# `go test -json` stream, its suite plan and its platform-case report, so a
# claim about a gate can always be checked against the run that produced it.
EVIDENCE ?= .temp/ci-evidence
GO_TEST_TIMEOUT ?= 30m
TEST_GATE := .github/ci/test-gate.sh
CANDIDATE := .github/ci/candidate-suite.sh

.PHONY: build test fmt lint vet check \
	require-pin-root ci-test race race-full check-ci \
	gate-selftest ledger-check no-broad-suppression \
	candidate-verify-ref candidate-record candidate-test

build:
	$(GO) build -ldflags '$(LDFLAGS)' -o bin/curator ./cmd/curator

test:
	$(GO) test ./...

fmt:
	gofmt -l -w .

vet:
	$(GO) vet ./...

lint:
	golangci-lint run

check: vet test
	@test -z "$$(gofmt -l .)" || { echo 'gofmt: files need formatting:'; gofmt -l .; exit 1; }

# --- CI gates ---------------------------------------------------------------
#
# These targets are the local mirror of the workflow. CI calls the scripts
# directly rather than through `make`, because `make` is not a guaranteed tool
# on windows-latest; the recipes below invoke exactly the same scripts with
# exactly the same environment, so a green `make check-ci` means the same thing
# the job does.
#
# Every gate consumes a protocol suite root. CI exports it from the committed
# pin; a local run must supply it explicitly, because a gate that quietly runs
# with the conformance suite unset is a smaller gate wearing the same name.
require-pin-root:
	@test -n "$(CURATOR_CONFORMANCE_ROOT)" || { \
		echo 'CURATOR_CONFORMANCE_ROOT is required by this gate.'; \
		echo 'CI exports it from the committed SPEC_PIN checkout; locally, point it'; \
		echo 'at a materialised <curator-spec>/conformance/v1 directory.'; \
		exit 1; }

# Mirrors the `test` job's gate step: plan from the root, run the planned
# package set, enforce the platform-case ledger.
ci-test: require-pin-root
	GO='$(GO)' GO_TEST_TIMEOUT='$(GO_TEST_TIMEOUT)' \
		bash $(TEST_GATE) '$(EVIDENCE)/test'

# Mirrors the `race` job. The package set is not narrowed by hand: the same
# plan decides it, so no maintained list can drift out of coverage.
race: require-pin-root
	GO='$(GO)' GO_TEST_TIMEOUT='$(GO_TEST_TIMEOUT)' GO_TEST_FLAGS='-race' \
		bash $(TEST_GATE) '$(EVIDENCE)/race'

race-full: race

# Proves the ledger's per-platform claims against the real per-GOOS builds.
ledger-check:
	bash .github/ci/ledger-consistency.sh '$(EVIDENCE)/ledger'

no-broad-suppression:
	bash .github/ci/no-broad-suppression.sh

# Proves the gates reject what they claim to reject. Needs no conformance
# root, no network and no Go build.
gate-selftest:
	bash .github/ci/gate-selftest.sh

# Mirrors the `test` job end to end, in order, over the same paths CI checks
# (`check` keeps its wider `.` scope, which additionally walks the skills
# submodule).
check-ci: require-pin-root
	@test -z "$$(gofmt -l cmd internal)" || { echo 'gofmt: files need formatting:'; gofmt -l cmd internal; exit 1; }
	$(GO) vet ./...
	$(MAKE) ledger-check
	$(MAKE) ci-test

# --- Candidate protocol suite ----------------------------------------------
#
# Non-default by construction: nothing here reads or writes the committed pin.
# `candidate-test` sets CI_REQUIRE_FULL_ROOT=1, exactly as the workflow's
# candidate job does, so a candidate root that cannot serve a package fails
# instead of quietly deferring it.
candidate-verify-ref:
	@test -n "$(CANDIDATE_REF)" || { echo 'CANDIDATE_REF is required'; exit 1; }
	bash $(CANDIDATE) verify-ref '$(CANDIDATE_REF)'

candidate-record:
	@test -n "$(CANDIDATE_ROOT)" || { echo 'CANDIDATE_ROOT is required'; exit 1; }
	bash $(CANDIDATE) record '$(CANDIDATE_ROOT)' '$(EVIDENCE)/candidate'

candidate-test:
	@test -n "$(CANDIDATE_ROOT)" || { echo 'CANDIDATE_ROOT is required'; exit 1; }
	CURATOR_CONFORMANCE_ROOT='$(CANDIDATE_ROOT)' CI_REQUIRE_FULL_ROOT=1 \
	GO='$(GO)' GO_TEST_TIMEOUT='$(GO_TEST_TIMEOUT)' \
		bash $(TEST_GATE) '$(EVIDENCE)/candidate-test'
