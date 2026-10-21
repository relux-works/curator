# Local validation record

The qualification identity and baseline are in the companion record. All tests
use `env -u GIT_DIR -u GIT_WORK_TREE go test ...`. Run roots below are distinct
ignored task-local `.temp/` directories; the private binary locator is omitted.
Go 1.25.5 and a workspace-local module/build cache are used.

| Command or run | Exit | Claim or interpretation |
| --- | ---: | --- |
| `go test ./.research/project-exclusion-muse -count=1 -v`, no vendor gate | 0 | Every test skips before case creation; `ungated.log`. A skip is not native qualification. |
| Same command, with binary gate and fresh run root, `-timeout 8m`, gated-04 | 0 | Full startup, resume, late-root and sandbox matrix; `gated.log`. Unsupported contracts remain explicitly unmeasured. |
| `go test ./.research/project-exclusion-muse -run '^TestChecker$' -count=1 -v`, final gate | 0 | Final checker also rejects a contradictory event hidden behind another field name. |
| `go test ./.research/project-exclusion-muse -run '^TestP8LateRoot$' -count=1 -v`, final gate | 0 | Final late-root check requires two distinct native listings, each retaining the home root and reporting no project root. |
| `sh .research/project-exclusion-muse/mutants.sh` | 0 | Five intended failures detected. Each inner Go command exits 1 for the named invariant, never for an unrelated setup failure. |
| `sh -n .research/project-exclusion-muse/mutants.sh` | 0 | Control wrapper shell syntax. |
| `go vet ./.research/project-exclusion-muse ./...` | 0 | Research and default packages pass static vet. |
| `go build ./...` | 0 | Default module compiles without a product delta. |

The tests' claim comments delimit loader enumeration, actual marker-hook effects
and terminal observations. P5's approved MCP definition is a config-loading
positive, not a server-handshake positive. P6's malformed-user-file rejection
is a parser control, not a tool-permission decision.

Exploratory runs are not silently counted as passing qualification:

- gated-01, exit 1: the first session-replay event preceded catalog loading;
  extraction incorrectly stopped there. The final boundary is security-mode
  resolution and missing observations still refuse.
- gated-02, exit 1: ancestors outside the Git root were incorrectly expected to
  load; those now have their own zero-discovery control beside inside-checkout
  ancestor positives. The old resume assertion incorrectly expected a fresh root
  listing. The loopback positive also compared curl's progress meter to body-only
  output; the final command uses `--silent --show-error`.
- focus-01, exit 1: resume retained the original rule snapshot. The final probe
  measures live skill/hook/MCP catalogs separately and observes their revival
  after granting trust.
- gated-03, exit 1: all source rows passed; the remaining curl-output assertion
  above was corrected. gated-04 and the expected-red boundary controls pass.

Exact signed-head default-suite commands, package/shard coverage, actual exits,
clone removal, signature verification and branch publication are recorded in the
board outcome `TASK-261021-4bo0a7_results-r1.md`. They are delivery evidence,
separate from the native qualification matrix above. No hosted CI, Linux/Windows
native qualification, production adapter admission or release certification is
claimed here.
