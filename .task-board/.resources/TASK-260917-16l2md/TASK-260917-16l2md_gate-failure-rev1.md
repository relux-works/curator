# TASK-260917-16l2md — hosted gate failure on Change Request revision 1 (run 35266385103)

Extracted by the orchestrator from the CI evidence. At the new pin
(`SPEC_PIN dced9b8`) EVERY lane's vector suites are green — the byte-exact
`TestManagerConfigV2Vectors`, the gate self-test, lint, interop, Linux and
macOS Test/Race all pass. The only red job is `Test (windows-latest)`, one
test:

```
cmd/curator TestEnvStatusMatrix   env_test.go:136: status --check after repair = 1
  provider run:     C:\Users\RUNNER~1\AppData\Local\Temp\TestEnvStatusMatrix…\003\curator-run.exe     (subcommand_provider_untrusted, non-current)
  provider session: C:\Users\RUNNER~1\AppData\Local\Temp\TestEnvStatusMatrix…\003\curator-session.exe (subcommand_provider_untrusted, non-current)
```

So the E4 half of the union misclassifies the test's providers on Windows.
Two things to fix, both in the E4 code (`cmd/curator/umbrella.go` and the
`env status` provider rows), nothing in the pin or the other candidates:

1. **Trust-root identity on Windows.** The fixture's provider directory is
   reported through the 8.3 short spelling (`RUNNER~1`) while the trust
   roots (the running manager's install directory, `provider_directories`)
   are compared in long spelling — so a provider that IS inside a trust
   root looks outside it. Compare trust roots and resolved providers by
   canonical identity: resolve both through `filepath.EvalSymlinks` AND a
   long-path normalization (`GetLongPathName`-class, e.g. via
   `os.Stat`/`SameFile` on the directory or `filepath.EvalSymlinks` which
   expands 8.3 names on Windows), fold case for the drive letter, and treat
   two spellings of one directory as the same root — exactly the identity
   rule S6 adopted for approval records. Add a Windows-executed test with an
   8.3-spelled trust root.
2. **Revision A must warn, not refuse.** The landed E4 spec ships
   revision A as the default: a provider outside the trust roots is a
   WARNING row (`subcommand_provider_outside_trust_roots`, current, with the
   `provider_directories` migration hint), and `subcommand_provider_untrusted`
   (non-current/refusal) is revision B's or the manager-published-directory
   case. The row above shows the B diagnostic under the default profile:
   verify which revision the union actually ships as default and that the
   posture row and `--check` follow revision A unless the profile option
   selects B. Do not paper over this by changing the test's expectation.

Re-run the narrow gates with the rc.12 root (`CURATOR_CONFORMANCE_ROOT` at
`/tmp/spec-rc12/conformance/v1`) incl. `go test -count=1 ./cmd/curator/...`,
and hand off again; the runtime re-runs the hosted gate.
