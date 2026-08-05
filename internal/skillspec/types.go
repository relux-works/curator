// Package skillspec parses and validates the portable skill machine manifest,
// schemas 1 through 7 (Spec §4), including legacy filename and runtime
// fallbacks.
package skillspec

import "github.com/relux-works/curator/internal/capabilities"

const (
	// CanonicalManifestName is the implementation-neutral writer filename.
	CanonicalManifestName = "agent-skill.json"
	// LegacyManifestName remains readable throughout protocol 1.x.
	LegacyManifestName = "csk-skill.json"
	// RuntimeFallbackName is consulted only when neither modern manifest exists.
	RuntimeFallbackName = "agents/runtime.json"
)

// SupportedSchemaVersions is the accepted agent skill manifest schema range.
var SupportedSchemaVersions = map[int]bool{1: true, 2: true, 3: true, 4: true, 5: true, 6: true, 7: true}

// UpgradeHint tells the user how to move to a build that understands a newer
// schema.
const UpgradeHint = "upgrade curator to a release that supports this schema"

// Command is one exported command (Spec §5.4).
type Command struct {
	Name       string
	Type       string // "script", "system", or "build"
	Command    string // system: binary name on PATH
	UnixPath   string // script
	WinPath    string // script
	Hint       string // system, optional
	Driver     string // build: "go-v1"
	SourceDir  string // build: package directory below BuildRoots
	Repository string // go-repository-v1: key in BuildRepositories
	Target     string // go-repository-v1: key in skill-build.json targets
}

// LockedCommit is an immutable Git object lock. ObjectFormat determines the
// exact accepted width of Hex; there is no mutable-ref representation.
type LockedCommit struct {
	ObjectFormat string // "sha1" or "sha256"
	Hex          string // full lowercase object id
}

// BuildRepository is one schema-7 external build repository declaration.
type BuildRepository struct {
	Name         string
	Git          string
	Identity     string // canonical network-git host/path
	Transport    string // "https" or "ssh"
	LockedCommit LockedCommit
	Tag          string // optional safe refs/tags name
}

// CommandDependency is a dependencies.commands entry (Spec §5.6).
type CommandDependency struct {
	Name    string
	Type    string // "system" or legacy "skill"
	Command string
	Skill   string // legacy "skill" type only
	Hint    string
}

// Requirement is a dependencies.skills entry (Spec §5.7).
type Requirement struct {
	Name     string
	Git      string
	RefKind  string // "tag" or "revision"
	RefValue string
	Mode     string // "full", "runtime", "context"
	Commands []string
}

// McpServer is a dependencies.mcp_servers entry (Spec §5.8).
type McpServer struct {
	Name       string
	Hint       string
	Transport  string // "", "stdio", "http"
	RequiredIn string // "any" or "all"
}

// Spec is the parsed manifest of one skill snapshot.
type Spec struct {
	SchemaVersion     int
	SourceFile        string // canonical, legacy, runtime fallback, or "" for pure context skills
	RuntimeRoots      []string
	BuildRoots        []string
	BuildRepositories map[string]BuildRepository
	Capabilities      capabilities.Manifest
	Commands          map[string]Command
	Dependencies      map[string]CommandDependency
	Requirements      map[string]Requirement
	McpServers        map[string]McpServer
}
