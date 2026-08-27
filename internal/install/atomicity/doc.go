// Package atomicity holds the cross-scope install commit acceptance suite.
//
// It has no production code. The suite lives beside internal/install rather
// than inside it because every case here drives a complete real installation —
// a baseline plus one run per injected failure — and the combined runtime would
// otherwise push the internal/install test binary past the default per-package
// test timeout. Splitting the binary gives each suite its own budget and keeps
// `go test ./...` honest without a custom timeout.
//
// The suite deliberately uses only the exported installation API, so it also
// serves as a check that the atomicity contract is observable from outside the
// package that implements it.
package atomicity
