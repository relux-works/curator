package runtimestore

import (
	"os"
	"path/filepath"
	"strings"
)

// MatchesLegacyShim recognizes only a byte-exact generated launcher or a
// legacy symlink to an independently authorized runtime target. A name, a
// shebang, a directory prefix, or a similar-looking script is not ownership.
func MatchesLegacyShim(path, platform string, authorized func(string) bool) bool {
	if authorized == nil {
		return false
	}
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}
	if info.Mode()&os.ModeSymlink != 0 {
		if platform != "unix" {
			return false
		}
		target, err := os.Readlink(path)
		if err != nil {
			return false
		}
		if !filepath.IsAbs(target) {
			target = filepath.Join(filepath.Dir(path), target)
		}
		return authorized(filepath.Clean(target))
	}
	if !info.Mode().IsRegular() || info.Size() > 65536 {
		return false
	}
	payload, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	text := string(payload)
	var target string
	var entries []string
	if platform == "windows" {
		const start = "call \""
		const end = "\" %*\r\nexit /b %ERRORLEVEL%\r\n"
		index := strings.LastIndex(text, start)
		if index < 0 || !strings.HasSuffix(text, end) {
			return false
		}
		target = strings.ReplaceAll(text[index+len(start):len(text)-len(end)], "%%%%", "%")
		if prefix, ok := windowsPathPrefix(text); ok {
			entries = strings.Split(prefix, ";")
		}
		if WindowsShimContent(target, entries) != text {
			return false
		}
	} else {
		const end = " \"$@\"\n"
		index := strings.LastIndex(text, "\nexec ")
		if index < 0 || !strings.HasSuffix(text, end) {
			return false
		}
		raw := text[index+len("\nexec ") : len(text)-len(end)]
		var ok bool
		target, ok = unquoteShim(raw)
		if !ok {
			return false
		}
		const prefix = "  PATH="
		if at := strings.Index(text, prefix); at >= 0 {
			line, _, _ := strings.Cut(text[at+len(prefix):], "\n")
			quoted, ok := strings.CutSuffix(line, ":\"$PATH\"")
			if !ok {
				return false
			}
			value, ok := unquoteShim(quoted)
			if !ok {
				return false
			}
			entries = strings.Split(value, ":")
		}
		if UnixShimContent(target, entries) != text {
			return false
		}
	}
	return filepath.IsAbs(target) && authorized(filepath.Clean(target))
}

func unquoteShim(raw string) (string, bool) {
	if len(raw) < 2 || raw[0] != '\'' || raw[len(raw)-1] != '\'' {
		return "", false
	}
	value := strings.ReplaceAll(raw[1:len(raw)-1], "'\"'\"'", "'")
	return value, shellQuote(value) == raw
}

func windowsPathPrefix(text string) (string, bool) {
	const start = "  set \"PATH="
	index := strings.Index(text, start)
	if index < 0 {
		return "", false
	}
	line, _, _ := strings.Cut(text[index+len(start):], "\r\n")
	value, ok := strings.CutSuffix(line, ";%PATH%\"")
	if !ok {
		return "", false
	}
	return strings.ReplaceAll(value, "%%", "%"), true
}
