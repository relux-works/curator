// Package ciartifact sanitizes shareable CI evidence while preserving gate inputs.
package ciartifact

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"
)

// Counts contains only aggregate redaction totals. It never includes source text.
type Counts struct {
	Files       int `json:"files"`
	Redactions  int `json:"redactions"`
	Paths       int `json:"paths"`
	Assignments int `json:"assignments"`
	Hostnames   int `json:"hostnames"`
	IPAddresses int `json:"ip_addresses"`
	Credentials int `json:"credentials"`
	BinaryFiles int `json:"binary_files"`
}

type counts struct {
	paths, assignments, hostnames, ipAddresses, credentials int
}

var (
	windowsPathRE    = regexp.MustCompile(`(?i)(^|[^a-z0-9])([a-z]:\\[^\s"'<>;,]+)`)
	fileURLPathRE    = regexp.MustCompile(`(?i)(file:///{1,3})([a-z]:)?(/[^\s"'<>;,]+)`)
	unixPathRE       = regexp.MustCompile(`(^|[^a-zA-Z0-9_:/])(/[^\s"'<>;,]+)`)
	envRE            = regexp.MustCompile(`\b([A-Z][A-Z0-9_]{1,63})=([^\s,;"']+)`)
	sensitiveFieldRE = regexp.MustCompile(`(?i)((?:access[_-]?token|api[_-]?key|auth(?:orization)?|client[_-]?secret|credential|password|private[_-]?key|secret|token)(?:\s*[:=]\s*["']?))([^\s,"';&]+)`)
	bearerRE         = regexp.MustCompile(`(?i)\bBearer\s+[A-Za-z0-9._~+/-]{12,}=?`)
	githubTokenRE    = regexp.MustCompile(`\b(?:gh[pousr]_[A-Za-z0-9]{20,}|github_pat_[A-Za-z0-9_]{20,})\b`)
	gitlabTokenRE    = regexp.MustCompile(`\bglpat-[A-Za-z0-9_-]{20,}\b`)
	openAITokenRE    = regexp.MustCompile(`\bsk-(?:proj-|live-|test-)?[A-Za-z0-9_-]{20,}\b`)
	awsKeyRE         = regexp.MustCompile(`\bAKIA[0-9A-Z]{16}\b`)
	slackTokenRE     = regexp.MustCompile(`\bxox[baprs]-[A-Za-z0-9-]{12,}\b`)
	jwtRE            = regexp.MustCompile(`\beyJ[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}\b`)
	pemRE            = regexp.MustCompile(`(?s)-----BEGIN [A-Z0-9 ]*PRIVATE KEY-----.*?-----END [A-Z0-9 ]*PRIVATE KEY-----`)
	sshKeyRE         = regexp.MustCompile(`\b(?:ssh-rsa|ssh-ed25519|ecdsa-sha2-nistp[0-9]+)\s+[A-Za-z0-9+/]{40,}={0,3}`)
	ipv4RE           = regexp.MustCompile(`(^|[^0-9.])((?:[0-9]{1,3}\.){3}[0-9]{1,3})([^0-9.]|$)`)
	ipv6RE           = regexp.MustCompile(`(?i)\b(?:[0-9a-f]{0,4}:){2,7}[0-9a-f]{0,4}\b`)
	hostRE           = regexp.MustCompile(`(?i)\b(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+(?:[a-z]{2,63})\b`)
	sensitiveKeyRE   = regexp.MustCompile(`(?i)(?:token|secret|password|credential|authorization|api[_-]?key|private[_-]?key)`)
)

const (
	curatorModulePath  = "github.com/relux-works/curator"
	redactedPath       = "[redacted path]"
	redactedAssignment = "[redacted]"
	redactedHostname   = "[redacted hostname]"
	redactedIP         = "[redacted IP address]"
	redactedCredential = "[redacted credential]" // #nosec G101 -- fixed redaction label; contains no credential material.
	redactedSubtest    = "[redacted subtest]"
)

var (
	errEvidenceFileClose   = errors.New("close evidence file")
	errEvidenceRootClose   = errors.New("close evidence source")
	errEvidenceStageRemove = errors.New("remove staged evidence copy")
)

// SanitizeTree copies src to a new dst directory, sanitizing textual evidence
// in the copy. The source remains unchanged. The destination must not exist.
func SanitizeTree(src, dst string) (Counts, error) {
	return sanitizeTree(src, dst, os.RemoveAll, openEvidenceInput)
}

type evidenceInput interface {
	io.Reader
	Stat() (os.FileInfo, error)
	Close() error
}

type evidenceInputOpener func(*os.Root, string) (evidenceInput, error)

func openEvidenceInput(root *os.Root, relative string) (evidenceInput, error) {
	return root.Open(relative)
}

func sanitizeTree(src, dst string, removeStage func(string) error, openInput evidenceInputOpener) (total Counts, err error) {
	source, err := filepath.Abs(src)
	if err != nil {
		return total, errors.New("resolve evidence source")
	}
	destination, err := filepath.Abs(dst)
	if err != nil {
		return total, errors.New("resolve evidence destination")
	}
	if overlaps(source, destination) {
		return total, errors.New("evidence source and destination overlap")
	}
	if _, err := os.Lstat(destination); err == nil {
		return total, errors.New("evidence destination already exists")
	} else if !errors.Is(err, os.ErrNotExist) {
		return total, errors.New("inspect evidence destination")
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return total, errors.New("create evidence destination parent")
	}
	stage, err := os.MkdirTemp(filepath.Dir(destination), ".sanitize-ci-evidence-")
	if err != nil {
		return total, errors.New("create staged evidence copy")
	}
	stageOwned := true
	defer func() {
		if stageOwned {
			if removeErr := removeStage(stage); removeErr != nil {
				err = errors.Join(err, errEvidenceStageRemove)
			}
		}
	}()

	info, err := os.Lstat(source)
	if errors.Is(err, os.ErrNotExist) {
		if err := os.Rename(stage, destination); err != nil {
			return total, errors.New("publish empty evidence copy")
		}
		stageOwned = false
		return total, nil
	}
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return total, errors.New("evidence source is not a directory")
	}
	sourceRoot, err := os.OpenRoot(source)
	if err != nil {
		return total, errors.New("open evidence source")
	}
	sourceRootOpen := true
	defer func() {
		if sourceRootOpen {
			if closeErr := sourceRoot.Close(); closeErr != nil {
				err = errors.Join(err, errEvidenceRootClose)
			}
		}
	}()

	err = filepath.WalkDir(source, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return errors.New("read evidence tree")
		}
		rel, err := filepath.Rel(source, path)
		if err != nil {
			return errors.New("resolve evidence entry")
		}
		if rel == "." {
			return nil
		}
		mode := entry.Type()
		if mode&os.ModeSymlink != 0 || (!entry.IsDir() && !entry.Type().IsRegular()) {
			return errors.New("evidence tree contains an unsupported entry")
		}
		target := filepath.Join(stage, rel)
		if entry.IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return errors.New("create evidence directory")
			}
			return nil
		}
		if err := copyEvidenceFile(sourceRoot, rel, target, &total, openInput); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return total, err
	}
	if closeErr := sourceRoot.Close(); closeErr != nil {
		sourceRootOpen = false
		return total, errEvidenceRootClose
	}
	sourceRootOpen = false
	if err := os.Rename(stage, destination); err != nil {
		return total, errors.New("publish sanitized evidence copy")
	}
	stageOwned = false
	return total, nil
}

func overlaps(a, b string) bool {
	rel, err := filepath.Rel(a, b)
	if err == nil && (rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))) {
		return true
	}
	rel, err = filepath.Rel(b, a)
	return err == nil && (rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))))
}

func copyEvidenceFile(root *os.Root, source, destination string, total *Counts, openInput evidenceInputOpener) (err error) {
	in, err := openInput(root, source)
	if err != nil {
		return errors.New("open evidence file")
	}
	defer func() {
		if closeErr := in.Close(); closeErr != nil {
			err = errors.Join(err, errEvidenceFileClose)
		}
	}()
	info, err := in.Stat()
	if err != nil || !info.Mode().IsRegular() {
		return errors.New("inspect evidence file")
	}
	data, err := io.ReadAll(in)
	if err != nil {
		return errors.New("read evidence file")
	}
	if !utf8.Valid(data) || bytes.IndexByte(data, 0) >= 0 {
		return errors.New("evidence tree contains binary content")
	}
	clean, redactions, err := sanitizeContent(data, isGoTestEventStream(source))
	if err != nil {
		return err
	}
	total.Files++
	total.Redactions += redactions.paths + redactions.assignments + redactions.hostnames + redactions.ipAddresses + redactions.credentials
	total.Paths += redactions.paths
	total.Assignments += redactions.assignments
	total.Hostnames += redactions.hostnames
	total.IPAddresses += redactions.ipAddresses
	total.Credentials += redactions.credentials
	return writeEvidenceFile(destination, clean, info.Mode().Perm())
}

func isGoTestEventStream(path string) bool {
	switch filepath.Base(path) {
	case "go-test.json", "go-test-served.json", "replay.jsonl":
		return true
	default:
		return false
	}
}

func writeEvidenceFile(path string, data []byte, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return errors.New("create evidence directory")
	}
	if err := os.WriteFile(path, data, mode); err != nil {
		return errors.New("write sanitized evidence file")
	}
	return nil
}

func sanitizeTextOrJSON(data []byte) ([]byte, counts, error) {
	return sanitizeContent(data, false)
}

func sanitizeGoTestStream(data []byte) ([]byte, counts, error) {
	return sanitizeContent(data, true)
}

func sanitizeContent(data []byte, preserveEventFields bool) ([]byte, counts, error) {
	if json.Valid(data) {
		return sanitizeJSONStrings(data, preserveEventFields)
	}
	var out bytes.Buffer
	var pending bytes.Buffer
	var total counts
	flushText := func() {
		if pending.Len() == 0 {
			return
		}
		clean, found := sanitizeText(pending.String())
		out.WriteString(clean)
		addCounts(&total, found)
		pending.Reset()
	}
	reader := bufio.NewReader(bytes.NewReader(data))
	for {
		line, err := reader.ReadBytes('\n')
		if len(line) > 0 {
			payload := line
			ending := []byte(nil)
			if payload[len(payload)-1] == '\n' {
				payload, ending = payload[:len(payload)-1], []byte{'\n'}
			}
			if json.Valid(payload) {
				flushText()
				clean, found, scanErr := sanitizeJSONStrings(payload, preserveEventFields)
				if scanErr != nil {
					return nil, total, scanErr
				}
				out.Write(clean)
				out.Write(ending)
				addCounts(&total, found)
			} else {
				pending.Write(payload)
				pending.Write(ending)
			}
		}
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, total, errors.New("read evidence text")
		}
	}
	flushText()
	return out.Bytes(), total, nil
}

func sanitizeJSONStrings(data []byte, preserveEventFields bool) ([]byte, counts, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	first, err := decoder.Token()
	if err != nil {
		return nil, counts{}, errors.New("parse evidence JSON")
	}
	var changes []jsonChange
	var total counts
	if err := walkJSONValue(decoder, data, first, 0, "", preserveEventFields, &changes, &total); err != nil {
		return nil, total, err
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		return nil, total, errors.New("parse trailing evidence JSON")
	}
	if len(changes) == 0 {
		return data, total, nil
	}
	var out bytes.Buffer
	last := 0
	for _, change := range changes {
		out.Write(data[last:change.start])
		out.Write(change.raw)
		last = change.end
	}
	out.Write(data[last:])
	return out.Bytes(), total, nil
}

type jsonChange struct {
	start, end int
	raw        []byte
}

func walkJSONValue(decoder *json.Decoder, data []byte, token json.Token, before int64, key string, preserveEventFields bool, changes *[]jsonChange, total *counts) error {
	switch value := token.(type) {
	case json.Delim:
		switch value {
		case '{':
			for decoder.More() {
				keyToken, err := decoder.Token()
				if err != nil {
					return errors.New("parse evidence JSON key")
				}
				field, ok := keyToken.(string)
				if !ok {
					return errors.New("invalid evidence JSON key")
				}
				valueBefore := decoder.InputOffset()
				child, err := decoder.Token()
				if err != nil {
					return errors.New("parse evidence JSON value")
				}
				if sensitiveKeyRE.MatchString(field) {
					if preserveEventFields {
						return errors.New("test event metadata contains a sensitive field")
					}
					if err := skipJSONValue(decoder, child); err != nil {
						return err
					}
					start, end, err := jsonValueSpan(data, valueBefore, decoder.InputOffset())
					if err != nil {
						return err
					}
					replacement, err := json.Marshal(redactedCredential)
					if err != nil {
						return errors.New("encode redacted evidence value")
					}
					*changes = append(*changes, jsonChange{start: start, end: end, raw: replacement})
					total.credentials++
					continue
				}
				if err := walkJSONValue(decoder, data, child, valueBefore, field, preserveEventFields, changes, total); err != nil {
					return err
				}
			}
			if _, err := decoder.Token(); err != nil {
				return errors.New("close evidence JSON object")
			}
		case '[':
			for decoder.More() {
				valueBefore := decoder.InputOffset()
				child, err := decoder.Token()
				if err != nil {
					return errors.New("parse evidence JSON array")
				}
				if err := walkJSONValue(decoder, data, child, valueBefore, key, preserveEventFields, changes, total); err != nil {
					return err
				}
			}
			if _, err := decoder.Token(); err != nil {
				return errors.New("close evidence JSON array")
			}
		}
	case string:
		if preserveEventFields && key != "Output" {
			if hasForbiddenEventIdentifier(key, value) {
				if key != "Test" {
					return errors.New("test event metadata contains a forbidden identifier")
				}
				// Downstream admission checks the exact top-level test inventory;
				// subtest names are diagnostic metadata. Preserve a safe parent and
				// normalize only the unsafe suffix. Unsafe parents and other fields
				// still refuse publication.
				parent, _, ok := strings.Cut(value, "/")
				if !ok || parent == "" || hasForbiddenEventIdentifier(key, parent) {
					return errors.New("test event metadata contains a forbidden identifier")
				}
				start, end, err := jsonStringSpan(data, before)
				if err != nil {
					return err
				}
				replacement, err := json.Marshal(parent + "/" + redactedSubtest)
				if err != nil {
					return errors.New("encode normalized test metadata")
				}
				*changes = append(*changes, jsonChange{start: start, end: end, raw: replacement})
				_, found := sanitizeText(value)
				addCounts(total, found)
			}
			return nil
		}
		start, end, err := jsonStringSpan(data, before)
		if err != nil {
			return err
		}
		clean, found := sanitizeText(value)
		if clean == value {
			return nil
		}
		replacement, err := json.Marshal(clean)
		if err != nil {
			return errors.New("encode sanitized evidence value")
		}
		*changes = append(*changes, jsonChange{start: start, end: end, raw: replacement})
		addCounts(total, found)
	}
	return nil
}

func jsonStringSpan(data []byte, before int64) (int, int, error) {
	start := int(before)
	for start < len(data) && (data[start] == ' ' || data[start] == '\n' || data[start] == '\r' || data[start] == '\t' || data[start] == ':' || data[start] == ',') {
		start++
	}
	if start >= len(data) || data[start] != '"' {
		return 0, 0, errors.New("locate evidence JSON string")
	}
	for i := start + 1; i < len(data); i++ {
		switch data[i] {
		case '\\':
			i++
		case '"':
			return start, i + 1, nil
		}
	}
	return 0, 0, errors.New("close evidence JSON string")
}

func jsonValueSpan(data []byte, before, after int64) (int, int, error) {
	start := int(before)
	for start < len(data) && (data[start] == ' ' || data[start] == '\n' || data[start] == '\r' || data[start] == '\t' || data[start] == ':' || data[start] == ',') {
		start++
	}
	end := int(after)
	if start >= end || end > len(data) {
		return 0, 0, errors.New("locate evidence JSON value")
	}
	return start, end, nil
}

func skipJSONValue(decoder *json.Decoder, token json.Token) error {
	delimiter, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delimiter {
	case '{':
		for decoder.More() {
			if _, err := decoder.Token(); err != nil {
				return errors.New("skip evidence JSON key")
			}
			child, err := decoder.Token()
			if err != nil {
				return errors.New("skip evidence JSON value")
			}
			if err := skipJSONValue(decoder, child); err != nil {
				return err
			}
		}
		if _, err := decoder.Token(); err != nil {
			return errors.New("close skipped evidence JSON object")
		}
	case '[':
		for decoder.More() {
			child, err := decoder.Token()
			if err != nil {
				return errors.New("skip evidence JSON array value")
			}
			if err := skipJSONValue(decoder, child); err != nil {
				return err
			}
		}
		if _, err := decoder.Token(); err != nil {
			return errors.New("close skipped evidence JSON array")
		}
	default:
		return errors.New("skip invalid evidence JSON value")
	}
	return nil
}

func hasForbiddenEventIdentifier(key, value string) bool {
	if sensitiveKeyRE.MatchString(key) {
		return true
	}
	if key == "Package" {
		if value == curatorModulePath {
			return false
		}
		if strings.HasPrefix(value, curatorModulePath+"/") {
			value = strings.TrimPrefix(value, curatorModulePath+"/")
		}
	}
	_, found := sanitizeText(value)
	return found != (counts{})
}

func sanitizeText(text string) (string, counts) {
	var found counts
	text = replaceMatches(text, pemRE, redactedCredential, &found.credentials)
	text = replaceMatches(text, sshKeyRE, redactedCredential, &found.credentials)
	text = replaceMatches(text, bearerRE, redactedCredential, &found.credentials)
	text = replaceMatches(text, githubTokenRE, redactedCredential, &found.credentials)
	text = replaceMatches(text, gitlabTokenRE, redactedCredential, &found.credentials)
	text = replaceMatches(text, openAITokenRE, redactedCredential, &found.credentials)
	text = replaceMatches(text, awsKeyRE, redactedCredential, &found.credentials)
	text = replaceMatches(text, slackTokenRE, redactedCredential, &found.credentials)
	text = replaceMatches(text, jwtRE, redactedCredential, &found.credentials)
	text = sensitiveFieldRE.ReplaceAllStringFunc(text, func(match string) string {
		parts := sensitiveFieldRE.FindStringSubmatch(match)
		if len(parts) != 3 {
			return match
		}
		if strings.HasPrefix(parts[2], "[redacted") {
			return match
		}
		found.credentials++
		return parts[1] + redactedCredential
	})
	text = envRE.ReplaceAllStringFunc(text, func(match string) string {
		parts := envRE.FindStringSubmatch(match)
		if len(parts) != 3 {
			return match
		}
		if strings.HasPrefix(parts[2], "[redacted") {
			return match
		}
		found.assignments++
		return parts[1] + "=" + redactedAssignment
	})
	text = replaceIPAddresses(text, &found.ipAddresses)
	text = fileURLPathRE.ReplaceAllStringFunc(text, func(string) string {
		found.paths++
		return redactedPath
	})
	text = windowsPathRE.ReplaceAllStringFunc(text, func(match string) string {
		parts := windowsPathRE.FindStringSubmatch(match)
		if len(parts) != 3 {
			return match
		}
		found.paths++
		return parts[1] + redactedPath
	})
	text = hostRE.ReplaceAllStringFunc(text, func(host string) string {
		if isFileLikeName(host) {
			return host
		}
		found.hostnames++
		return redactedHostname
	})
	text = unixPathRE.ReplaceAllStringFunc(text, func(match string) string {
		parts := unixPathRE.FindStringSubmatch(match)
		if len(parts) != 3 {
			return match
		}
		found.paths++
		return parts[1] + redactedPath
	})
	return text, found
}

func isFileLikeName(value string) bool {
	if strings.Count(value, ".") != 1 {
		return false
	}
	suffix := strings.ToLower(value[strings.LastIndexByte(value, '.')+1:])
	switch suffix {
	case "go", "mod", "sum", "json", "jsonl", "txt", "md", "sh", "yaml", "yml", "toml", "tsv", "log", "err", "out", "exe", "dll", "so", "dylib", "cache", "data", "html", "css", "js", "ts", "rs", "swift", "py", "rb", "xml", "zip", "tar", "gz", "png", "jpg", "jpeg", "gif", "wasm", "csv":
		return true
	default:
		return false
	}
}

func replaceMatches(text string, pattern *regexp.Regexp, replacement string, count *int) string {
	return pattern.ReplaceAllStringFunc(text, func(string) string {
		*count++
		return replacement
	})
}

func replaceIPAddresses(text string, count *int) string {
	text = ipv4RE.ReplaceAllStringFunc(text, func(match string) string {
		parts := ipv4RE.FindStringSubmatch(match)
		if len(parts) != 4 || net.ParseIP(parts[2]) == nil {
			return match
		}
		*count++
		return parts[1] + redactedIP + parts[3]
	})
	return ipv6RE.ReplaceAllStringFunc(text, func(candidate string) string {
		if net.ParseIP(strings.Trim(candidate, "[]")) == nil {
			return candidate
		}
		*count++
		return redactedIP
	})
}

func addCounts(total *counts, delta counts) {
	total.paths += delta.paths
	total.assignments += delta.assignments
	total.hostnames += delta.hostnames
	total.ipAddresses += delta.ipAddresses
	total.credentials += delta.credentials
}
