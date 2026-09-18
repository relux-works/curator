package buildmeta

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/identity"
)

// Frozen source-types-v1 package arms (skillfile-sources §4). The shapes are
// disjoint: a local snapshot never carries Git fields and a Git package never
// carries a snapshot digest. The values mirror sourcelock's package identity
// byte for byte so a receipt-3 `input.package` compares exactly with the lock
// member and the marker-5 `package` it was installed from; buildmeta keeps its
// own copy because the receipt model must stay a leaf below the build drivers.
const (
	PackageKindLocalSnapshot = "local-snapshot"
	PackageKindNetworkGit    = "network-git"
	PackageKindConfiguredGit = "configured-git"
)

var (
	packageDigestRE    = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
	packageSHA1RE      = regexp.MustCompile(`^[0-9a-f]{40}$`)
	packageSHA256HexRE = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

// PackageCommit is a locked Git object: format-bound lowercase hex, never a
// snapshot digest and never prefixed.
type PackageCommit struct {
	ObjectFormat string
	Hex          string
}

// Package is the frozen package identity a receipt-3 build input binds.
// Exactly one arm is populated; Validate refuses every cross-arm field.
type Package struct {
	Kind       string
	Snapshot   string         // local-snapshot only: sha256:<hex> inventory digest
	Repository string         // network-git only: canonical host/path
	Source     string         // configured-git only: legacy configured-root path
	Commit     *PackageCommit // Git arms only
	Directory  string         // Git arms only: effective per-member directory
}

// Validate enforces the disjoint arm shapes of source-types schema 1.
func (p Package) Validate() error {
	switch p.Kind {
	case PackageKindLocalSnapshot:
		if !packageDigestRE.MatchString(p.Snapshot) {
			return fmt.Errorf("package snapshot must be sha256:<64 lowercase hex>")
		}
		if p.Repository != "" || p.Source != "" || p.Commit != nil || p.Directory != "" {
			return fmt.Errorf("local-snapshot package must not carry Git identity")
		}
	case PackageKindNetworkGit:
		if p.Snapshot != "" || p.Source != "" {
			return fmt.Errorf("network-git package must not carry snapshot or configured source")
		}
		if !validPackageRepository(p.Repository) {
			return fmt.Errorf("package repository must be canonical host/path")
		}
		if err := p.Commit.validate(); err != nil {
			return err
		}
		if p.Directory != "." && (!identifiers.PortablePath(p.Directory) || strings.ContainsAny(p.Directory, "*?[]")) {
			return fmt.Errorf("package directory must be '.' or a portable contained path")
		}
	case PackageKindConfiguredGit:
		if p.Snapshot != "" || p.Repository != "" {
			return fmt.Errorf("configured-git package must not carry snapshot or network repository")
		}
		if !identifiers.PortablePath(p.Source) {
			return fmt.Errorf("package source must be a portable configured-root path")
		}
		if err := p.Commit.validate(); err != nil {
			return err
		}
		if p.Directory != "." {
			return fmt.Errorf("configured-git package directory must be '.'")
		}
	default:
		return fmt.Errorf("unknown package kind %q", p.Kind)
	}
	return nil
}

func (c *PackageCommit) validate() error {
	if c == nil {
		return fmt.Errorf("package commit is required")
	}
	switch c.ObjectFormat {
	case "sha1":
		if !packageSHA1RE.MatchString(c.Hex) {
			return fmt.Errorf("sha1 package commit must be 40 lowercase hex")
		}
	case "sha256":
		if !packageSHA256HexRE.MatchString(c.Hex) {
			return fmt.Errorf("sha256 package commit must be 64 lowercase hex")
		}
	default:
		return fmt.Errorf("package commit object_format must be sha1 or sha256")
	}
	return nil
}

func validPackageRepository(value string) bool {
	if value == "" || utf8.RuneCountInString(value) > 4096 || strings.HasSuffix(value, ".git") {
		return false
	}
	return identity.ValidCanonical(value)
}

// Object returns the closed source-types-v1 value: exactly the fields of
// one arm. Callers validate first; an invalid package renders an object no
// receipt decoder accepts.
func (p Package) Object() map[string]any { return p.object() }

// object is the closed source-types-v1 value: exactly the fields of one arm.
func (p Package) object() map[string]any {
	switch p.Kind {
	case PackageKindNetworkGit:
		return map[string]any{
			"kind":       p.Kind,
			"repository": p.Repository,
			"commit":     map[string]any{"object_format": p.Commit.ObjectFormat, "hex": p.Commit.Hex},
			"directory":  p.Directory,
		}
	case PackageKindConfiguredGit:
		return map[string]any{
			"kind":      p.Kind,
			"source":    p.Source,
			"commit":    map[string]any{"object_format": p.Commit.ObjectFormat, "hex": p.Commit.Hex},
			"directory": p.Directory,
		}
	default:
		return map[string]any{"kind": p.Kind, "snapshot": p.Snapshot}
	}
}

// parsePackage decodes the closed package object: the exact field set of the
// declared arm, string-typed members, then the same validation as Validate.
// Null, foreign-arm and unknown fields are refused before any value is used.
func parsePackage(raw map[string]any) (Package, error) {
	kind, err := stringField(raw, "kind", "package")
	if err != nil {
		return Package{}, err
	}
	var allowed []string
	switch kind {
	case PackageKindLocalSnapshot:
		allowed = []string{"kind", "snapshot"}
	case PackageKindNetworkGit:
		allowed = []string{"kind", "repository", "commit", "directory"}
	case PackageKindConfiguredGit:
		allowed = []string{"kind", "source", "commit", "directory"}
	default:
		return Package{}, fmt.Errorf("unknown package kind %q", kind)
	}
	sort.Strings(allowed)
	if err := exactFields(raw, "package", allowed...); err != nil {
		return Package{}, err
	}
	parsed := Package{Kind: kind}
	for _, field := range allowed {
		switch field {
		case "kind":
		case "commit":
			commitObject, err := objectField(raw, "commit", "package")
			if err != nil {
				return Package{}, err
			}
			if err := exactFields(commitObject, "package commit", "hex", "object_format"); err != nil {
				return Package{}, err
			}
			format, err := stringField(commitObject, "object_format", "package commit")
			if err != nil {
				return Package{}, err
			}
			hexValue, err := stringField(commitObject, "hex", "package commit")
			if err != nil {
				return Package{}, err
			}
			parsed.Commit = &PackageCommit{ObjectFormat: format, Hex: hexValue}
		default:
			value, err := stringField(raw, field, "package")
			if err != nil {
				return Package{}, err
			}
			switch field {
			case "snapshot":
				parsed.Snapshot = value
			case "repository":
				parsed.Repository = value
			case "source":
				parsed.Source = value
			case "directory":
				parsed.Directory = value
			}
		}
	}
	if err := parsed.Validate(); err != nil {
		return Package{}, err
	}
	return parsed, nil
}
