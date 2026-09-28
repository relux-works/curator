package envprofile

import (
	"errors"
	"fmt"
	"io/fs"

	"github.com/relux-works/curator/internal/contextpkg"
	"github.com/relux-works/curator/internal/identity"
	"github.com/relux-works/curator/internal/pathboundary"
)

const (
	// DiagPathSourceUntrusted reports an untrusted protected-boundary source.
	DiagPathSourceUntrusted = "environment_store_untrusted"
	// DiagMCPPathSourceRefused reports an MCP declaration carried by a path source.
	DiagMCPPathSourceRefused = "mcp_declaration_path_source_refused"
)

var pathSourceOwnerLookup = pathboundary.DefaultOwnerLookup()

// validateProfilePathSources verifies only directories named as path
// sources. It does not reread their content as package state or recompute a
// state pin from the live directory.
func validateProfilePathSources(profile string, source Source, policy Policy) error {
	if source.Kind == KindPath && source.Path != "" {
		if err := validatePathDirectory(source.Path); err != nil {
			return err
		}
	}
	for _, overlay := range policy.EffectiveOverlays(profile) {
		if identity.ClassifySource(overlay.Source) != identity.SourcePath {
			continue
		}
		if err := validatePathDirectory(overlay.Source); err != nil {
			return err
		}
	}
	return nil
}

// preflightProfilePathSources reads one installed source record and proves
// its path directories before a mutating caller can establish defaults or
// write a materialized home. A proven absence leaves profile-existence and
// policy diagnostics to the operation that consumes the profile; unreadable
// and malformed records remain errors.
func preflightProfilePathSources(home, profile string, policy Policy) error {
	source, err := readSource(home, profile)
	if err != nil {
		if isStateAbsent(err) {
			return nil
		}
		return err
	}
	return validateProfilePathSources(profile, source, policy)
}

// preflightCurrentPathSources covers all profiles a sync will materialize
// before ensureDefault or the first native-home write.
func preflightCurrentPathSources(home string, policy Policy) error {
	current, err := Current(home)
	if err != nil {
		return err
	}
	if current == "" {
		current = DefaultProfile
	}
	if err := preflightProfilePathSources(home, current, policy); err != nil {
		return err
	}
	scoped, err := ScopedCurrents(home)
	if err != nil {
		return err
	}
	for _, profile := range scoped {
		if err := preflightProfilePathSources(home, profile, policy); err != nil {
			return err
		}
	}
	return nil
}

// PreflightCurrentPathSources validates the current machine and scoped profile
// path sources for a manager-home mutation that is not routed through the
// envprofile operation lock, such as the standalone garbage collector. Callers
// must hold the manager-home mutation lock while using this as a mutation gate.
func PreflightCurrentPathSources(home string, policy Policy) error {
	return preflightCurrentPathSources(home, policy)
}

func validatePathDirectory(path string) error {
	if err := pathboundary.ValidateWithOwner(path, pathSourceOwnerLookup); err != nil {
		return pathBoundaryDiagnostic(path, err)
	}
	return nil
}

// validatePathPackageDirectory proves the source boundary before parsing any
// package bytes. Missing and non-directory operands retain the established
// path-source diagnostics; an unsafe directory is environment_store_untrusted.
func validatePathPackageDirectory(path string) error {
	err := pathboundary.ValidateWithOwner(path, pathSourceOwnerLookup)
	if err == nil {
		return nil
	}
	var failure *pathboundary.Failure
	if errors.As(err, &failure) && failure.Check == pathboundary.CheckRegular &&
		(pathboundary.IsAbsent(err) || errors.Is(err, fs.ErrPermission) || failure.Reason == "declared path is not a directory") {
		return pathManifestDiag(path, err)
	}
	return pathBoundaryDiagnostic(path, err)
}

func pathBoundaryDiagnostic(path string, err error) error {
	var failure *pathboundary.Failure
	if errors.As(err, &failure) {
		return fmt.Errorf("%s: path %q fails %s boundary check at %q: %s",
			DiagPathSourceUntrusted, failure.Root, failure.Check, failure.Path, boundaryReason(failure))
	}
	return fmt.Errorf("%s: path %q boundary could not be proven: %v", DiagPathSourceUntrusted, path, err)
}

func preflightPathOverlayDeclarations(profile string, policy Policy) error {
	for _, overlay := range policy.EffectiveOverlays(profile) {
		if identity.ClassifySource(overlay.Source) != identity.SourcePath {
			continue
		}
		if overlay.Range != "" || overlay.Tag != "" || overlay.Revision != "" || overlay.Directory != "" {
			return fmt.Errorf("%s: a path overlay carries no range, tag, revision, or directory", DiagSourceInvalid)
		}
		if err := validatePathPackageDirectory(overlay.Source); err != nil {
			return err
		}
		manifest, err := contextpkg.LoadManifest(overlay.Source)
		if err != nil {
			return pathManifestDiag(overlay.Source, err)
		}
		if err := refusePathMCPDeclaration(overlay.Source, manifest.Name); err != nil {
			return err
		}
	}
	return nil
}

func boundaryReason(failure *pathboundary.Failure) string {
	if failure.Reason != "" {
		return failure.Reason
	}
	if failure.Cause != nil {
		return failure.Cause.Error()
	}
	return "check could not be proven"
}

func refusePathMCPDeclaration(root, packageName string) error {
	manifest, present, err := contextpkg.LoadMCPIfPresent(root)
	if err != nil {
		return fmt.Errorf("%s: path package at %q has an invalid MCP declaration: %w", contextpkg.DiagMCPInvalid, root, err)
	}
	if !present {
		return nil
	}
	return fmt.Errorf("%s: package %q carries MCP declaration %q from a path source",
		DiagMCPPathSourceRefused, packageName, manifest.Name)
}
