# TASK-260910-31ocjt — gate fix (THE ONLY CURRENT INSTRUCTION, with 31ocjt-brief.md)

Rev1 (tree 7b224eed) does not compile its tests: `go vet ./...` fails on every lane (run 36472221878) —
internal/crossconformance/draftsources_broker_e2e_test.go:72,108,109,125: undefined: buildrepo.EnvHTTPSBrokerSecret. You removed the env
var constant but this e2e test still drives the old env-based handoff. Update that test to the new pipe/fd delivery (it must still prove the
broker receives the secret, and additionally that the environment no longer carries it). Then run `go vet ./...` and `go test ./... -run
XXX_NO_TEST` (compile every test package) plus `go test ./internal/crossconformance ./internal/buildrepo -run 'Broker|Askpass|HTTPS'` and the
race detector on the touched packages (`go test -race ./internal/buildrepo`), all with real exit codes. Set status development; update
results; handoff; END YOUR TURN. No CHANGELOG/LOGBOOK edit.
