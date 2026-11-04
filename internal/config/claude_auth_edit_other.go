//go:build !unix

package config

import "os"

// The metadata writer currently qualifies Unix owner/mode and no-follow
// semantics. Other platforms refuse rather than assume equivalent protection.
func openClaudeAuthFile(string) (*os.File, error) { return nil, ErrClaudeAuthEditRefused }
func privateClaudeAuthInfo(os.FileInfo) bool      { return false }
func claudeAuthFileIdentity(os.FileInfo) string   { return "unavailable" }
