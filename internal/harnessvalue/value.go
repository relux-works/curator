// Package harnessvalue delegates bounded, immutable harness policy values to
// github.com/relux-works/curator-runtimes/pkg/harnessvalue.
package harnessvalue

import upstream "github.com/relux-works/curator-runtimes/pkg/harnessvalue"

// Value size and collection bounds delegated to curator-runtimes.
const (
	MaxTextBytes = upstream.MaxTextBytes
	MaxPathBytes = upstream.MaxPathBytes
	MaxItems     = upstream.MaxItems
)

// RelativePath is a validated harness-relative path value.
type RelativePath = upstream.RelativePath

// Bytes is an immutable bounded byte slice value.
type Bytes = upstream.Bytes

// NewRelativePath validates and constructs a relative harness path.
func NewRelativePath(value string) (RelativePath, error) { return upstream.NewRelativePath(value) }

// NewBytes validates and copies a bounded byte slice.
func NewBytes(value []byte) (Bytes, error) { return upstream.NewBytes(value) }

// Text validates bounded UTF-8 text.
func Text(value string) error { return upstream.Text(value) }
