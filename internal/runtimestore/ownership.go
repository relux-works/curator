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
	payload, err := os.ReadFile(path) // #nosec G304 -- read-only ownership probe; bounded regular bytes are compared, never executed.
	if err != nil {
		return false
	}
	text := string(payload)
	var target string
	var entries []string
	if platform == "windows" {
		return matchesLegacyWindowsShim(text, authorized)
	}
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

// Recognize the frozen templates from 669fe11c, b52c0842 and cfffd7cd,
// as well as the current two-pass call renderer (3cee1c19). Every candidate
// must reproduce all bytes and point to the independently recorded runtime.
func matchesLegacyWindowsShim(text string, authorized func(string) bool) bool {
	const originalStart = "@echo off\r\n\""
	const originalEnd = "\" %*\r\n"
	if strings.HasPrefix(text, originalStart) && strings.HasSuffix(text, originalEnd) {
		target := text[len(originalStart) : len(text)-len(originalEnd)]
		if filepath.IsAbs(target) && !strings.ContainsAny(target, "\"\r\n") && authorized(filepath.Clean(target)) {
			return true
		}
	}
	const start = "call \""
	const end = "\" %*\r\nexit /b %ERRORLEVEL%\r\n"
	index := strings.LastIndex(text, start)
	if index < 0 || !strings.HasSuffix(text, end) {
		return false
	}
	raw := text[index+len(start) : len(text)-len(end)]
	var entries []string
	if prefix, ok := windowsPathPrefix(text); ok {
		entries = strings.Split(prefix, ";")
	} else if after, ok := strings.CutPrefix(text, "@echo off\r\nsetlocal DisableDelayedExpansion\r\nset \"PATH="); ok {
		line, _, _ := strings.Cut(after, "\r\n")
		prefix, ok := strings.CutSuffix(line, ";%PATH%\"")
		if !ok {
			return false
		}
		entries = strings.Split(strings.ReplaceAll(prefix, "%%", "%"), ";")
	}
	for _, target := range []string{strings.ReplaceAll(raw, "%%%%", "%"), strings.ReplaceAll(raw, "%%", "%")} {
		if !filepath.IsAbs(target) || strings.ContainsAny(target, "\"\r\n") {
			continue
		}
		current := WindowsShimContent(target, entries)
		singlePass := strings.Replace(current, "call \""+escapeCMDCallValue(target)+"\" %*\r\n", "call \""+escapeCMDValue(target)+"\" %*\r\n", 1)
		lifecycle := "@echo off\r\nsetlocal DisableDelayedExpansion\r\n"
		if len(entries) > 0 {
			escaped := make([]string, len(entries))
			for i, entry := range entries {
				escaped[i] = escapeCMDValue(entry)
			}
			lifecycle += "set \"PATH=" + strings.Join(escaped, ";") + ";%PATH%\"\r\n"
		}
		lifecycle += "call \"" + escapeCMDValue(target) + "\" %*\r\nexit /b %ERRORLEVEL%\r\n"
		if (text == current || text == singlePass || text == lifecycle) && authorized(filepath.Clean(target)) {
			return true
		}
	}
	return false
}
