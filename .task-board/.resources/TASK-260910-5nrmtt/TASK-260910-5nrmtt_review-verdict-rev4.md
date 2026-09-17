# TASK-260910-5nrmtt review verdict — revision 4

Verdict: CHANGES_REQUESTED. Route to to-dev. Reviewer RUN-260916-d5435f.
CR-TASK-260910-5nrmtt-4; base 38e62cb8f4b0a427625e17dc6a02cf825f4407a4;
exact candidate tree 16d6c4fed543efb915df7cb8f62e14b1c667ddcc.
All 12 candidate file blob hashes independently matched the working files.
No repository source bytes were modified: tests and mutants used Go overlays in a temporary directory.
Run goal queried: not goal-bound. No directives recorded.

## P1: Unknown facilities are silently discarded

internal/buildrepo/transport.go:245-263,297-300 treats every facility outside
fatal/error/remote/ssh as harmless environmental output. This directly contradicts
rework 2's requirement that unknown lines and mixed evidence fail closed, and
repository-transport revision 1 section 2's audit/unclassified/malformed refusal.
Production entry AcquireNetworkResolved with a ready alternate accepts:

    fatal: unable to access 'https://fixture.test/repository.git/': Could not resolve host: fixture.test
    audit: policy denied

Observed: 2 fetches, err=<nil>; required: 1 fetch and refusal. An unknown facility
cannot establish absence of an audit failure. The committed environmental-noise
positive test encodes this unsafe broad exception rather than the accepted contract.

## P1: Fallback patterns still accept arbitrary unknown reason tails

internal/buildrepo/transport.go:180-187,211,215,218-226 uses prefix matches or
unrestricted tails for fallback-eligible shapes. Two independent production probes:

    fatal: unable to access 'https://fixture.test/repository.git/': The requested URL returned error: 503; object verification failed

    ssh: Could not resolve hostname fixture.test: unexpected resolver protocol response

Each observed: 2 fetches, err=<nil>; required: 1 fetch and refusal. The HTTP regex
stops at a word boundary after 503; the SSH resolver regex accepts any reason.
Whole-line anchoring around .* does not establish a complete known diagnostic.

Required rework: implement a closed grammar for the full diagnostic including
known reason structure, fail closed on any unconsumed/unknown text, and remove
the arbitrary-facility ignore rule. Retain only explicitly justified fixed framing.
Do not add these examples to a deny list. Preserve the attached three production
regressions and previous reviewer probes; correct tests that require unsafe broad
exceptions. Keep Windows typed pre-process refusal and strict admission intact.

## Independently executed checks

Shell: zsh; direct commands, no pipelines masking exit status.
- go test -p 1 -count=1 ./internal/buildrepo ./internal/gitcred: exit 0;
  buildrepo 45.885s, gitcred 3.440s.
- go vet ./internal/buildrepo ./internal/gitcred: exit 0.
- gofmt -l internal/buildrepo internal/gitcred: exit 0, empty output.
- git diff --check: exit 0.
- go test -p 1 -count=1 -overlay <review-temp>/overlay.json ./internal/buildrepo
  -run '^TestReviewRev4ClosedGrammar$' -v: exit 1, all 3/3 refusal probes failed;
  all report "fail-closed diagnostic produced 2 fetches, err=<nil>".
  The attached patch adds only this regression test, using existing fixtures.

Narrowing mutants, independently executed with separate overlays:
1. Remove only remote from diagnosticFacilities (retain fatal/error/ssh).
   TestResolvedTransportUnknownTailMustNotFallbackWhenAlternateReady: exit 1,
   "unknown-tailed failure fell back to a ready alternate and succeeded". KILLED.
2. Narrow Windows platform refusal to goos == windows AND runtime.GOOS != darwin.
   TestTransportResolutionPlatformGate: exit 1, nil instead of typed code. KILLED.
   This is helper-level evidence only; no Windows runtime replay locally.
2/2 mutants killed; candidate source unchanged throughout.

## Existing evidence and bounds

Read producer results and revision-4 validation log. Hosted run 35118798492
(https://github.com/relux-works/curator/actions/runs/35118798492) reports exit 0:
Linux/macOS/Windows Test, Linux/macOS Race, Lint, interop and gate self-tests green;
rose-air and Candidate suite skipped. Accepted as attached hosted evidence, not
independently rerun here. Full module suite was not run locally.

Scope is the 12 declared paths in buildrepo/gitcred plus draft documentation;
frozen v1 schemas and sibling policy checkpoint untouched. Windows uses allowed
option (b): typed refusal before provider/admission/process creation. Its native
runtime behavior relies on hosted evidence; local platform helper test passes.
Existing package tests cover production admission refusals, fail-closed classes,
shared deadline/non-exec child pipes, truncation, locked alternate content,
credential binding, sanitized errors and strict lane environment. These passing
checks do not establish complete grammar coverage: the added negative shapes
are 0/3 correctly refused. No full AC acceptance claim is made. Caller wiring,
real credential export, live network authentication and rose-air remain outside
this independent replay. No human-only blocker; this is implementation rework.
