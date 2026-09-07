package identity

import "regexp"

// SourceKind is the declared kind of a source spelling: environments §1
// admits a `git` source and a `path` source, and core §6.1 refuses a
// spelling that looks like a network form but is not one — "Invalid network
// forms MUST be rejected, not treated as local."
type SourceKind string

const (
	// SourceGit is a network git source: a ssh/git/http/https URL, or an
	// SCP `[user@]host:path` remote.
	SourceGit SourceKind = "git"
	// SourcePath is an operator-local package directory named by an
	// absolute, project-relative, or platform-absolute spelling.
	SourcePath SourceKind = "path"
	// SourceInvalid is neither: a `://` URL on an unsupported scheme, or a
	// colon-bearing spelling that is not a valid §6.1 network form. It is
	// refused rather than silently taken as local, which is what would
	// hand it past the network allowlist.
	SourceInvalid SourceKind = "invalid"
)

// ecmaSpace spells ECMA-262 `\s` out as a character-class body. The
// discriminator below is the Go transcription of the `$defs/overlay`
// patterns committed in schemas/v1/manager-config-v2.schema.json, and those
// patterns are evaluated by an ECMA-262 engine. Go's own `\s` is only
// [\t\n\f\r ]; ECMA-262 adds the vertical tab, the Unicode space
// separators, the line/paragraph separators, and U+FEFF. Spelling the class
// out keeps a source carrying an exotic space classified here exactly as
// the schema classifies it, instead of leaving the difference as a bound.
const ecmaSpace = `\t\n\v\f\r \x{00a0}\x{1680}\x{2000}-\x{200a}\x{2028}\x{2029}\x{202f}\x{205f}\x{3000}\x{feff}`

// The five patterns of manager-config-v2 `$defs/overlay`, transcribed. The
// committed schema is the authority; the conformance root does not publish
// it, so the pin is the forty-one published overlay cases driven through
// config.Load in internal/config, not a text comparison here.
var (
	// `^[A-Za-z][A-Za-z0-9+.-]+://` — a `://` URL whose scheme is two or
	// more characters. A one-character scheme is not a URL here, which is
	// what leaves `C://Users/x` to the drive-letter rule.
	sourceURLSchemeRE = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9+.-]+://`)
	// `^([Ss][Ss][Hh]|[Gg][Ii][Tt]|[Hh][Tt][Tt][Pp][Ss]?)://` — the four
	// admitted transports in any letter case.
	sourceGitSchemeRE = regexp.MustCompile(`^([Ss][Ss][Hh]|[Gg][Ii][Tt]|[Hh][Tt][Tt][Pp][Ss]?)://`)
	// `^[^/\s]*:` — a colon reached without crossing a `/` or whitespace.
	// A colon in a later segment (`packages/team:context`) does not match.
	sourceColonRE = regexp.MustCompile(`^[^/` + ecmaSpace + `]*:`)
	// `^[A-Za-z]:[\\/]` — a Windows drive letter, carved out ahead of the
	// SCP form.
	sourceDriveRE = regexp.MustCompile(`^[A-Za-z]:[\\/]`)
	// `^(?:[^/:\s@]+@)?[A-Za-z0-9][A-Za-z0-9.-]*:[^/\s\\]` — SCP
	// `[user@]host:path` on core §6.1's host grammar, whose first path
	// character is not `/`, whitespace, or a backslash.
	sourceSCPRE = regexp.MustCompile(`^(?:[^/:` + ecmaSpace + `@]+@)?[A-Za-z0-9][A-Za-z0-9.-]*:[^/` + ecmaSpace + `\\]`)
)

// ClassifySource is the one source-kind discriminator: it decides `git`,
// `path`, or refused from the spelling alone and never touches the
// filesystem. A filesystem probe here would let a directory planted in the
// operator's working directory shadow a git identity and carry it past the
// core §6.1 network allowlist, because a `path` source bypasses that
// allowlist by design.
//
// The three arms are the three `allOf` members of manager-config-v2
// `$defs/overlay`, in the order the schema states them:
//
//  1. a `://` URL on a scheme other than ssh/git/http/https is refused;
//  2. a spelling whose first colon is reached without crossing a `/` or
//     whitespace, and which is neither a URL, a drive letter, nor a valid
//     SCP remote, is refused;
//  3. an admitted URL scheme or a valid SCP remote is `git`; everything
//     else is `path`.
//
// The caller supplies the spelling verbatim. This function does not trim:
// the schema does not, and the conformance corpus pins the untrimmed
// reading.
func ClassifySource(spelling string) SourceKind {
	if spelling == "" {
		return SourceInvalid
	}
	urlScheme := sourceURLSchemeRE.MatchString(spelling)
	gitScheme := sourceGitSchemeRE.MatchString(spelling)
	if urlScheme && !gitScheme {
		return SourceInvalid
	}
	scp := sourceSCPRE.MatchString(spelling)
	if sourceColonRE.MatchString(spelling) && !urlScheme &&
		!sourceDriveRE.MatchString(spelling) && !scp {
		return SourceInvalid
	}
	if gitScheme || scp {
		return SourceGit
	}
	return SourcePath
}
