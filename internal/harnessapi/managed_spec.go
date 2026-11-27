package harnessapi

import upstream "github.com/relux-works/curator-runtimes/pkg/harnessapi"

// ShadowSpec declares a native shadow file path handled during managed-home setup.
type ShadowSpec = upstream.ShadowSpec

// PassthroughSpec declares a credential or state passthrough link in a managed home.
type PassthroughSpec = upstream.PassthroughSpec

// ManagedHomeSpec declares the managed-home layout for a managed-only harness.
type ManagedHomeSpec = upstream.ManagedHomeSpec
