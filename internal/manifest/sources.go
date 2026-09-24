package manifest

import (
	"unicode/utf8"

	"github.com/relux-works/curator/internal/buildrepo"
	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/identity"
	"github.com/relux-works/curator/internal/verr"
)

// ParseOptions is reader-owned capability admission, never a manifest field.
// This opts into unreleased skillfile-sources-v1 parsing only; callers must
// implement resolution/locking before consuming selectors for installation.
type ParseOptions struct{ DraftSourcesV1 bool }

// Source is the closed acquisition union. Path is literal native filesystem
// syntax relative to the declaring Skillfile (or absolute), without expansion.
// Network arms retain the declaration separately from canonical identity.
// Exactly one of Path, Git, Repository is populated by the parser.
type Source struct {
	Path       string
	Git        string
	Repository string
	Identity   string
	Ref        Ref
}

// Selector is either an individual (Decl.Name is set, Collection is false)
// or a collection. Directory remains source-relative, including the root ".".
// Include/exclude are unexpanded: filesystem membership belongs to resolution.
type Selector struct {
	From       string
	Directory  string
	Collection bool
	Include    []string
	Exclude    []string
}

func parseSources(obj map[string]any) (map[string]Source, error) {
	raw, present := obj["sources"]
	if !present {
		return nil, nil
	}
	entries, ok := raw.(map[string]any)
	if !ok {
		return nil, verr.New("sources", "source_selection_invalid: must be an object")
	}
	sources := make(map[string]Source, len(entries))
	for alias, raw := range entries {
		label := "sources." + alias
		if !identifiers.Valid(alias) {
			return nil, verr.New(label, "source_selection_invalid: invalid alias")
		}
		entry, ok := raw.(map[string]any)
		if !ok {
			return nil, verr.New(label, "source_selection_invalid: must be an object")
		}
		if _, present := entry["path"]; present {
			value, ok := entry["path"].(string)
			if len(unknownFields(entry, "path")) != 0 || !ok || value == "" || !utf8.ValidString(value) || utf8.RuneCountInString(value) > 4096 || containsControl(value) {
				return nil, verr.New(label, "source_selection_invalid: path requires a non-empty control-free native path and no other fields")
			}
			sources[alias] = Source{Path: value}
			continue
		}
		if len(unknownFields(entry, "git", "repository", "tag", "branch", "revision")) != 0 {
			return nil, verr.New(label, "source_selection_invalid: unsupported source fields")
		}
		_, gitPresent := entry["git"]
		_, repositoryPresent := entry["repository"]
		if gitPresent == repositoryPresent {
			return nil, verr.New(label, "source_selection_invalid: requires exactly one of git or repository")
		}
		source := Source{}
		if gitPresent {
			value, _ := entry["git"].(string)
			parsed, err := buildrepo.ParseSource(value)
			if err != nil {
				return nil, verr.New(label, "source_selection_invalid: invalid repository endpoint")
			}
			source.Git, source.Identity = value, parsed.Identity
		} else {
			value, _ := entry["repository"].(string)
			// Reuse the closed HTTPS endpoint grammar, then require exact canonical
			// equality: no normalization, credentials, ports or terminal .git allowed.
			parsed, err := buildrepo.ParseSource("https://" + value)
			if err != nil || parsed.Identity != value {
				return nil, verr.New(label, "source_selection_invalid: repository must be canonical host/path")
			}
			source.Repository, source.Identity = value, value
		}
		for _, kind := range []string{"tag", "branch", "revision"} {
			raw, present := entry[kind]
			if !present {
				continue
			}
			value, ok := raw.(string)
			if source.Ref.Kind != "" || !ok || !validSourceRef(kind, value) {
				return nil, verr.New(label, "source_selection_invalid: requires exactly one valid tag, branch or revision")
			}
			source.Ref = Ref{Kind: kind, Value: value}
		}
		if source.Ref.Kind == "" {
			return nil, verr.New(label, "source_selection_invalid: missing exact reference")
		}
		sources[alias] = source
	}
	return sources, nil
}

func validSourceRef(kind, value string) bool {
	if kind != "revision" {
		return identity.DraftSourceRefName(value)
	}
	if len(value) != 40 && len(value) != 64 {
		return false
	}
	for _, c := range value {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
}

func parseSelector(entry map[string]any, sources map[string]Source, label string) (Decl, error) {
	invalid := func(reason string) (Decl, error) {
		return Decl{}, verr.New(label, "source_selection_invalid: %s", reason)
	}
	from, _ := entry["from"].(string)
	if !identifiers.Valid(from) {
		return invalid("invalid source alias")
	}
	if _, exists := sources[from]; !exists {
		return Decl{}, verr.New(label, "source_alias_unknown: %s", from)
	}
	directory, _ := entry["directory"].(string)
	if !identifiers.ValidDirectory(directory) {
		return invalid("directory must be a portable contained path or '.'")
	}
	selector := &Selector{From: from, Directory: directory}
	if raw, named := entry["name"]; named {
		name, _ := raw.(string)
		if !identifiers.Valid(name) || len(unknownFields(entry, "name", "from", "directory")) != 0 {
			return invalid("invalid individual selector or mixed arms")
		}
		return Decl{Name: name, Selector: selector}, nil
	}
	if len(unknownFields(entry, "from", "directory", "include", "exclude")) != 0 {
		return invalid("unsupported collection fields")
	}
	include, err := selectorMembers(entry["include"], true)
	if err != nil {
		return invalid("include requires unique literal identifiers or '*' and must not be empty")
	}
	var exclude []string
	if raw, present := entry["exclude"]; present {
		exclude, err = selectorMembers(raw, false)
		if err != nil {
			return invalid("exclude requires unique literal identifiers")
		}
	}
	selector.Collection, selector.Include, selector.Exclude = true, include, exclude
	return Decl{Selector: selector}, nil
}

func selectorMembers(raw any, include bool) ([]string, error) {
	list, ok := raw.([]any)
	if !ok || include && len(list) == 0 {
		return nil, verr.New("members", "invalid list")
	}
	result := make([]string, 0, len(list))
	seen := map[string]bool{}
	for _, raw := range list {
		value, ok := raw.(string)
		if !ok || seen[value] || (!include || value != "*") && !identifiers.Valid(value) {
			return nil, verr.New("members", "invalid member")
		}
		seen[value] = true
		result = append(result, value)
	}
	return result, nil
}
