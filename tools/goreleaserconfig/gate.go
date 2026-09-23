// Package goreleaserconfig guards the GoReleaser rc channel values.
//
// A release candidate must not replace the stable install channels. In
// .goreleaser.yml that rests on three parsed values: every
// homebrew_casks/scoops entry's skip_upload and release.prerelease must be
// exactly the string "auto" (TASK-260908-2kqa77). When the cask/scoop keys
// were unset, v0.14.0-rc.1 reached the tap and the bucket as the served
// release (2026-08 incident).
//
// The check parses with gopkg.in/yaml.v3 and compares each parsed scalar
// against "auto" case-sensitively. A grep for the token "auto" cannot guard
// this (the "Auto" mutant defeats it), and neither can an indentation walk:
// a block scalar body or a first-position nested map is indistinguishable
// from a real key without a real parser (rev3 R1). yaml.v3 also rejects
// duplicate mapping keys, so a malformed document fails closed (rev3 R2).
//
// Field naming: the task AC says brews[*].skip_upload, but the config
// declares no brews stanza; the publishing surfaces are homebrew_casks,
// scoops and release. Those are what this package checks, every entry of
// each list stanza.
package goreleaserconfig

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Want is the only accepted channel value, compared case-sensitively.
const Want = "auto"

// Check parses data as a GoReleaser config and returns one failure message
// per violated value. A nil slice means every checked value parses as
// exactly "auto". Each message names the field, the entry and the observed
// value.
func Check(data []byte) []string {
	// Decode into any first: that is the decode where yaml.v3 rejects
	// duplicate mapping keys ("already defined"). A Node decode alone
	// keeps both copies (first wins), so without this pass the R2
	// fixture would be admitted.
	var strict any
	if err := yaml.Unmarshal(data, &strict); err != nil {
		return []string{"invalid YAML: " + err.Error()}
	}
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return []string{"invalid YAML: " + err.Error()}
	}
	if len(doc.Content) == 0 || doc.Content[0].Kind != yaml.MappingNode {
		return []string{`top level is not a mapping, want channel entries with skip_upload "auto"`}
	}
	root := doc.Content[0]
	var out []string
	for _, stanza := range []string{"homebrew_casks", "scoops"} {
		out = append(out, checkListStanza(root, stanza)...)
	}
	out = append(out, checkPrerelease(root)...)
	return out
}

// CheckFile reads the config at path and returns the Check findings,
// prefixed with path. An unreadable file fails closed.
func CheckFile(path string) []string {
	//nolint:gosec // The gate reads a caller-supplied config path (the repository file in CI); path traversal is not in a CI gate's threat model.
	data, err := os.ReadFile(path)
	if err != nil {
		return []string{"cannot read " + path + ": " + err.Error()}
	}
	out := Check(data)
	for i := range out {
		out[i] = path + ": " + out[i]
	}
	return out
}

func checkListStanza(root *yaml.Node, stanza string) []string {
	node := mappingValue(root, stanza)
	if node == nil {
		return []string{fmt.Sprintf("section %q is absent, want channel entries with skip_upload %q", stanza, Want)}
	}
	if node.Kind != yaml.SequenceNode {
		// A bare `stanza:` key parses as null: that is an emptied
		// stanza, not a mistyped one.
		if node.Kind == yaml.ScalarNode && node.Tag == "!!null" {
			return []string{fmt.Sprintf("section %q has no entries, want at least one with skip_upload %q", stanza, Want)}
		}
		return []string{fmt.Sprintf("section %q is not a list, want channel entries with skip_upload %q", stanza, Want)}
	}
	if len(node.Content) == 0 {
		return []string{fmt.Sprintf("section %q has no entries, want at least one with skip_upload %q", stanza, Want)}
	}
	var out []string
	for i, entry := range node.Content {
		field := fmt.Sprintf("%s[%d].skip_upload", stanza, i)
		if entry.Kind != yaml.MappingNode {
			out = append(out, fmt.Sprintf("%s: entry is not a mapping, want %q", field, Want))
			continue
		}
		val := mappingValue(entry, "skip_upload")
		if val == nil {
			out = append(out, fmt.Sprintf("%s is absent, want %q", field, Want))
			continue
		}
		if val.Kind != yaml.ScalarNode {
			out = append(out, fmt.Sprintf("%s is %s, want %q", field, kindName(val.Kind), Want))
			continue
		}
		if val.Value != Want {
			out = append(out, fmt.Sprintf("%s = %q, want %q", field, val.Value, Want))
		}
	}
	return out
}

func checkPrerelease(root *yaml.Node) []string {
	node := mappingValue(root, "release")
	if node == nil {
		return []string{fmt.Sprintf("section %q is absent, want release.prerelease %q", "release", Want)}
	}
	if node.Kind != yaml.MappingNode {
		return []string{fmt.Sprintf("section %q is not a mapping, want release.prerelease %q", "release", Want)}
	}
	val := mappingValue(node, "prerelease")
	if val == nil {
		return []string{fmt.Sprintf("release.prerelease is absent, want %q", Want)}
	}
	if val.Kind != yaml.ScalarNode {
		return []string{fmt.Sprintf("release.prerelease is %s, want %q", kindName(val.Kind), Want)}
	}
	if val.Value != Want {
		return []string{fmt.Sprintf("release.prerelease = %q, want %q", val.Value, Want)}
	}
	return nil
}

// mappingValue returns the value of key in the mapping m, or nil. Duplicate
// keys never reach this walk: yaml.Unmarshal rejects the document first.
func mappingValue(m *yaml.Node, key string) *yaml.Node {
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Kind == yaml.ScalarNode && m.Content[i].Value == key {
			return m.Content[i+1]
		}
	}
	return nil
}

func kindName(k yaml.Kind) string {
	switch k {
	case yaml.MappingNode:
		return "a mapping"
	case yaml.SequenceNode:
		return "a sequence"
	case yaml.AliasNode:
		return "an alias"
	default:
		return "not a scalar"
	}
}
