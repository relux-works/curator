// Package projectlease owns stable project homes and durable session leases.
package projectlease

import "github.com/relux-works/curator/internal/projectleaserecord"

// Transport aliases preserve the existing public API and exact serialized bytes.
// Common admission and the executor share the same closed decoder.
type Binding = projectleaserecord.Binding
type Surface = projectleaserecord.Surface
type Composition = projectleaserecord.Composition
type Authority = projectleaserecord.Authority
type Evidence = projectleaserecord.Evidence
type Sessions = projectleaserecord.Sessions
type Request = projectleaserecord.Request
type Phase = projectleaserecord.Phase
type Status = projectleaserecord.Status
type Roots = projectleaserecord.Roots

var ErrRevoked = projectleaserecord.ErrRevoked
var ErrArtifactsUnavailable = projectleaserecord.ErrArtifactsUnavailable
var ErrPendingRecovery = projectleaserecord.ErrPendingRecovery

const Busy = projectleaserecord.Busy
const maxRecord = projectleaserecord.MaxRecord
const (
	Prepared = projectleaserecord.Prepared
	Active   = projectleaserecord.Active
	Unknown  = projectleaserecord.Unknown
	Released = projectleaserecord.Released
)

func tagged(s, tag string) bool                { return projectleaserecord.Tagged(s, tag) }
func homeRelative(c Composition) string        { return projectleaserecord.HomeRelative(c) }
func canonical(v any) ([]byte, error)          { return projectleaserecord.Canonical(v) }
func digest(raw []byte) string                 { return projectleaserecord.Digest(raw) }
func foldPath(path string) string              { return projectleaserecord.FoldPath(path) }
func ParseBinding(raw []byte) (Binding, error) { return projectleaserecord.ParseBinding(raw) }
