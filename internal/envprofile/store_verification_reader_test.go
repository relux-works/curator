package envprofile

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// These are qualification inputs only. No production source reads either
// selector. The explicit proposal pin is not a canonical capability claim.
const storeVerificationRootKey = "CURATOR_STORE_VERIFICATION_ROOT"
const storeVerificationPinKey = "CURATOR_STORE_VERIFICATION_MANIFEST_SHA256"
const storeVerificationUnsupported = "store_verification_corpus_unsupported"
const storeVerificationManifestPin = "61e5d95e02555764d60633aa9d6656e118a7a6c969f10202885218bc7b9d7899"
const storeVerificationCasesPin = "b103291e1dbbf964d7bf5444477f7b0211b834823bf6ea0df2bf4805a81b9ba7"
const storeVerificationPredecessorPin = "7efa5b09e450162cfb2937f25b3e29fab4699602e0990cfd295a4f695cb60413"

//go:embed testdata/store-verification-v2/cases.schema.json
var storeVerificationCaseSchema []byte

type storeVerificationObservation struct {
	Diagnostic     string   `json:"diagnostic"`
	Check          string   `json:"failing_check"`
	Fragment       bool     `json:"fragment_emitted"`
	Current        bool     `json:"row_current"`
	Writes         []string `json:"writes"`
	Rebuilt        bool     `json:"store_rebuilt"`
	LockUnchanged  bool     `json:"lock_unchanged"`
	PlantPreserved bool     `json:"plant_preserved"`
	PathNamed      bool     `json:"path_named"`
	RuleNamed      bool     `json:"rule_named"`
}
type storeVerificationPredecessor struct {
	Group string `json:"group"`
	Name  string `json:"name"`
}
type storeVerificationCase struct {
	Schema      string                        `json:"schema"`
	Capability  string                        `json:"capability"`
	Revision    int                           `json:"capability_revision"`
	Name        string                        `json:"name"`
	Kind        string                        `json:"case_kind"`
	Operation   string                        `json:"operation"`
	Condition   string                        `json:"condition"`
	Member      string                        `json:"member_kind"`
	Source      string                        `json:"source_kind"`
	Home        string                        `json:"home_state"`
	Levels      []int                         `json:"verification_levels"`
	Expected    storeVerificationObservation  `json:"expected"`
	Positive    string                        `json:"positive_control"`
	Predecessor *storeVerificationPredecessor `json:"predecessor"`
	Forbidden   []struct {
		Name        string                       `json:"name"`
		Observation storeVerificationObservation `json:"observation"`
	} `json:"forbidden_observations"`
}
type storeVerificationManifest struct {
	Schema      string `json:"schema"`
	Capability  string `json:"capability"`
	Revision    int    `json:"capability_revision"`
	Status      string `json:"status"`
	Cases       string `json:"cases"`
	Count       int    `json:"case_count"`
	Predecessor struct {
		Commit string `json:"commit"`
		Path   string `json:"path"`
		SHA256 string `json:"sha256"`
	} `json:"predecessor"`
	Migration []struct {
		Group         string  `json:"group"`
		Name          string  `json:"name"`
		Disposition   string  `json:"disposition"`
		OldDiagnostic *string `json:"old_diagnostic"`
		OldOutcome    *string `json:"old_outcome"`
		NewDiagnostic *string `json:"new_required_diagnostic"`
		Successor     string  `json:"successor_case"`
		Negative      bool    `json:"known_bad_remains_negative"`
	} `json:"migration"`
}

func storeVerificationError(err error) error {
	return fmt.Errorf("%s: %w", storeVerificationUnsupported, err)
}

// Token validation precedes schema and typed decoding; neither may erase a
// duplicate decoded key, invalid UTF-8 or excess nesting by normalizing JSON.
func storeVerificationWalk(d *json.Decoder, depth int) error {
	if depth > 32 {
		return errors.New("JSON depth exceeds 32")
	}
	tok, err := d.Token()
	if err != nil {
		return err
	}
	switch tok {
	case json.Delim('{'):
		seen := map[string]bool{}
		for d.More() {
			key, err := d.Token()
			if err != nil {
				return err
			}
			name, ok := key.(string)
			if !ok || seen[name] {
				return errors.New("duplicate JSON key")
			}
			seen[name] = true
			if err := storeVerificationWalk(d, depth+1); err != nil {
				return err
			}
		}
		_, err = d.Token()
	case json.Delim('['):
		for d.More() {
			if err := storeVerificationWalk(d, depth+1); err != nil {
				return err
			}
		}
		_, err = d.Token()
	}
	return err
}
func storeVerificationStrict(raw []byte, out any) error {
	if len(raw) > 4<<20 || !utf8.Valid(raw) {
		return errors.New("invalid JSON bytes")
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	if err := storeVerificationWalk(d, 0); err != nil {
		return err
	}
	if _, err := d.Token(); err != io.EOF {
		return errors.New("trailing JSON")
	}
	d = json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	d.DisallowUnknownFields()
	return d.Decode(out)
}

type storeVerificationSchemaLoader struct{}

func (storeVerificationSchemaLoader) Load(string) (any, error) {
	return nil, errors.New("external qualification schema refused")
}

func decodeStoreVerificationCases(raw []byte) ([]storeVerificationCase, error) {
	bad := func(err error) ([]storeVerificationCase, error) { return nil, storeVerificationError(err) }
	var doc, schemaDoc any
	if err := storeVerificationStrict(raw, &doc); err != nil {
		return bad(err)
	}
	if err := storeVerificationStrict(storeVerificationCaseSchema, &schemaDoc); err != nil {
		return bad(err)
	}
	c := jsonschema.NewCompiler()
	c.UseLoader(storeVerificationSchemaLoader{})
	const schemaURL = "https://example.invalid/store-verification-cases.schema.json"
	if err := c.AddResource(schemaURL, schemaDoc); err != nil {
		return bad(err)
	}
	schema, err := c.Compile(schemaURL)
	if err != nil {
		return bad(err)
	}
	if err := schema.Validate(doc); err != nil {
		return bad(err)
	}
	var cases []storeVerificationCase
	// Integer fields decode directly to int: 2.0 and 2e0 must not normalize to 2.
	if err := storeVerificationStrict(raw, &cases); err != nil {
		return bad(err)
	}
	byName := map[string]storeVerificationCase{}
	predecessors := map[string]bool{}
	for _, row := range cases {
		if _, ok := byName[row.Name]; ok {
			return bad(errors.New("duplicate case ID"))
		}
		byName[row.Name] = row
		if row.Predecessor != nil {
			key := row.Predecessor.Group + "/" + row.Predecessor.Name
			if predecessors[key] {
				return bad(errors.New("duplicate predecessor"))
			}
			predecessors[key] = true
		}
	}
	for _, row := range cases {
		positive, ok := byName[row.Positive]
		if !ok || positive.Kind != "implementation" || positive.Condition != "intact" || positive.Operation != row.Operation || positive.Member != row.Member {
			return bad(errors.New("invalid positive-control reference"))
		}
	}
	return cases, nil
}

func storeVerificationReadPinned(path, pin string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > 4<<20 {
		return nil, errors.New("qualification input is not a bounded regular file")
	}
	raw, err := io.ReadAll(io.LimitReader(f, (4<<20)+1))
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(raw)
	if len(raw) > 4<<20 || hex.EncodeToString(sum[:]) != pin {
		return nil, errors.New("qualification input pin mismatch")
	}
	return raw, nil
}

// Reading all artifacts, versions, row references and predecessor coverage is
// atomic with respect to fixture execution: this function creates no fixture.
func readStoreVerificationCorpus(lookup func(string) (string, bool)) ([]storeVerificationCase, bool, error) {
	root, present := lookup(storeVerificationRootKey)
	pin, hasPin := lookup(storeVerificationPinKey)
	if !present && !hasPin {
		return nil, false, nil
	}
	bad := func(err error) ([]storeVerificationCase, bool, error) { return nil, true, storeVerificationError(err) }
	if !present || root == "" || !filepath.IsAbs(root) || !hasPin || pin != storeVerificationManifestPin {
		return bad(errors.New("explicit root and matching exact manifest pin required"))
	}
	raw, err := storeVerificationReadPinned(filepath.Join(root, "manifest.json"), pin)
	if err != nil {
		return bad(err)
	}
	var m storeVerificationManifest
	if err := storeVerificationStrict(raw, &m); err != nil {
		return bad(err)
	}
	if m.Schema != "curator.environment-store-corpus/2" || m.Capability != "agent-environments" || m.Revision != 2 || m.Status != "proposal-unreleased" || m.Cases != "vectors/environments-store-boundary.json" || m.Count != 134 || len(m.Migration) != 38 || m.Predecessor.SHA256 != storeVerificationPredecessorPin {
		return bad(errors.New("unsupported corpus manifest"))
	}
	raw, err = storeVerificationReadPinned(filepath.Join(root, filepath.FromSlash(m.Cases)), storeVerificationCasesPin)
	if err != nil {
		return bad(err)
	}
	cases, err := decodeStoreVerificationCases(raw)
	if err != nil {
		return nil, true, err
	}
	if len(cases) != m.Count {
		return bad(errors.New("incomplete case roster"))
	}
	byName := map[string]storeVerificationCase{}
	for _, row := range cases {
		byName[row.Name] = row
	}
	seen := map[string]bool{}
	for _, old := range m.Migration {
		key := old.Group + "/" + old.Name
		row, ok := byName[old.Successor]
		if seen[key] || !ok || row.Predecessor == nil || row.Predecessor.Group != old.Group || row.Predecessor.Name != old.Name {
			return bad(errors.New("predecessor roster mismatch"))
		}
		seen[key] = true
	}
	return cases, true, nil
}

func storeVerificationCanonicalGuard(lookup func(string) (string, bool)) error {
	if _, set := lookup(storeVerificationRootKey); set {
		return errors.New("canonical conformance refuses store qualification selector")
	}
	if _, set := lookup(storeVerificationPinKey); set {
		return errors.New("canonical conformance refuses store qualification pin")
	}
	return nil
}

// A lexical guard intentionally catches a production reference even in a
// comment or alias; runtime lookup need not have one particular AST shape.
func storeVerificationProductionReferences(root string) ([]string, error) {
	var found []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == ".git" || d.Name() == ".temp" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if bytes.Contains(raw, []byte(storeVerificationRootKey)) || bytes.Contains(raw, []byte(storeVerificationPinKey)) {
			found = append(found, path)
		}
		return nil
	})
	return found, err
}
