package envprofile

import (
	"io"
	"strings"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextmaterialize"
	"github.com/relux-works/curator/internal/contextpkg"
)

// AllowlistEmptyWarning reports the §2.2 empty-allowlist warning for the
// effective MCP package allowlist, or "" when the allowlist is non-empty.
// Profile install, profile update, and env status emit it whenever the
// effective allowlist is empty; it never fails the operation.
func AllowlistEmptyWarning(allowlist []string) string {
	if len(allowlist) > 0 {
		return ""
	}
	return DiagAllowlistEmpty + ": the MCP package allowlist is empty; " +
		"every declaration package in the closure is admitted"
}

// surfacingRows renders the §2.3 declaration rows for the lock's MCP set:
// one closed-column row per declaration package, in ascending package-name
// byte order, without the LF. The caller computes them after the audit
// gate passes and before the lock is published. A declaration whose
// manifest cannot be read is reported as unreadable — never as absence —
// and the operation continues: surfacing never fails.
func surfacingRows(home string, manager *gitManager, lock *contextlock.Lock) (rows []string, unreadable []string) {
	if lock == nil {
		return nil, nil
	}
	declarations := []contextmaterialize.MCPDeclaration{}
	for _, member := range lock.Members {
		if member.Kind != contextlock.KindMCP {
			continue
		}
		entry := manager.entryPath(home, resolvedOf(member))
		manifest, err := contextpkg.LoadMCP(packageRoot(entry, member.Directory))
		if err != nil {
			unreadable = append(unreadable, "mcp declaration "+member.Name+" cannot be read for surfacing: "+err.Error())
			continue
		}
		declarations = append(declarations, contextmaterialize.MCPDeclaration{
			Package:   member.Name,
			Version:   member.Version,
			Transport: manifest.Server.Transport,
			Command:   manifest.Server.Command,
			Args:      manifest.Server.Args,
			EnvNames:  manifest.Server.EnvNames,
		})
	}
	rendered := contextmaterialize.FormatDeclarationRows(declarations)
	if rendered == "" {
		return nil, unreadable
	}
	return strings.Split(strings.TrimSuffix(rendered, "\n"), "\n"), unreadable
}

// emitSurfacing prints the §2.3 rows to sink, one LF-terminated line each,
// at the emission point the caller has reached: after the audit gate
// passed and before the lock is published or any surface is
// (re-)materialized. Surfacing is informative (§2.3): a nil sink prints
// nothing, and a write failure never fails the operation. Callers that
// pass a sink leave Info.Surfacing empty — the rows were already
// printed — so the rows print exactly once.
func emitSurfacing(sink io.Writer, rows []string) {
	if sink == nil {
		return
	}
	for _, row := range rows {
		_, _ = io.WriteString(sink, row+"\n")
	}
}

// carrySurfacing emits rows at the §2.3 emission point and returns the
// rows Info.Surfacing must carry: with a sink the rows were already
// printed, so Info carries none and nothing prints twice; without a
// sink Info carries the rows.
func carrySurfacing(sink io.Writer, rows []string) []string {
	emitSurfacing(sink, rows)
	if sink != nil {
		return nil
	}
	return rows
}
