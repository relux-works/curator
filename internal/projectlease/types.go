// Package projectlease owns stable project homes and durable session leases.
package projectlease

import "github.com/relux-works/curator/internal/projectleaserecord"

// Binding preserves the existing public API and exact serialized bytes.
// Common admission and the executor share the same closed decoder.
type Binding = projectleaserecord.Binding

// Surface aliases the closed lease projection surface.
type Surface = projectleaserecord.Surface

// Composition aliases the protected composition authority record.
type Composition = projectleaserecord.Composition

// Authority aliases the lease authority transport interface.
type Authority = projectleaserecord.Authority

// Evidence aliases reconciled lease liveness evidence.
type Evidence = projectleaserecord.Evidence

// Sessions aliases the session observation transport.
type Sessions = projectleaserecord.Sessions

// Request aliases the bounded lease acquisition request.
type Request = projectleaserecord.Request

// Phase aliases the durable journal or lease phase.
type Phase = projectleaserecord.Phase

// Status aliases the current lease status projection.
type Status = projectleaserecord.Status

// Roots aliases the set of live lease roots.
type Roots = projectleaserecord.Roots

// ErrRevoked reports an explicitly revoked lease.
var ErrRevoked = projectleaserecord.ErrRevoked

// ErrArtifactsUnavailable reports missing protected artifacts.
var ErrArtifactsUnavailable = projectleaserecord.ErrArtifactsUnavailable

// ErrPendingRecovery requires durable lease recovery before continuation.
var ErrPendingRecovery = projectleaserecord.ErrPendingRecovery

// Busy reports a project view retained by a live or uncertain lease.
const Busy = projectleaserecord.Busy
const maxRecord = projectleaserecord.MaxRecord

// Prepared and its siblings enumerate durable lease phases.
const (
	Prepared = projectleaserecord.Prepared
	Active   = projectleaserecord.Active
	Unknown  = projectleaserecord.Unknown
	Released = projectleaserecord.Released
)

func tagged(s, tag string) bool         { return projectleaserecord.Tagged(s, tag) }
func homeRelative(c Composition) string { return projectleaserecord.HomeRelative(c) }
func canonical(v any) ([]byte, error)   { return projectleaserecord.Canonical(v) }
func digest(raw []byte) string          { return projectleaserecord.Digest(raw) }
func foldPath(path string) string       { return projectleaserecord.FoldPath(path) }

// ParseBinding decodes the exact bounded lease binding.
func ParseBinding(raw []byte) (Binding, error) { return projectleaserecord.ParseBinding(raw) }
