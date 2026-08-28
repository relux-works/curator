package crossconformance

import (
	_ "embed"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// AcceptedCorpusSHA256 pins the exact accepted CGP05/CGP10 golden bytes. The
// corpus is an accepted research artifact, so a change to it is a contract
// change and must fail here before it can reach any adapter.
const AcceptedCorpusSHA256 = "fed9657b33297650e64178d605c7ee7f47dde640fefd12a99dec3651bb9cadcb"

// AcceptedRecordCount is the published labeled-record count of the accepted
// CGP05 and CGP10 corpus.
const AcceptedRecordCount = 53

//go:embed testdata/canonical-goldens.txt
var acceptedCorpus []byte

// AcceptedCorpusBytes returns the embedded accepted corpus.
func AcceptedCorpusBytes() []byte {
	return append([]byte(nil), acceptedCorpus...)
}

// Record is one published four-line golden: fixture name, exact domain label,
// exact CCJ byte string, and published identity.
type Record struct {
	Name      string
	Label     string
	Payload   []byte
	Published string

	// Derived is the identity this package computed from Label and Payload
	// without consulting Published.
	Derived string
	// Value is the independently decoded payload.
	Value Value
}

// Corpus is the indexed accepted corpus.
type Corpus struct {
	Records []Record
	byName  map[string]Record
	byID    map[string]Record
}

// ByName returns the named record.
func (corpus Corpus) ByName(name string) (Record, bool) {
	record, ok := corpus.byName[name]
	return record, ok
}

// ByID returns the record whose derived identity is id.
func (corpus Corpus) ByID(id string) (Record, bool) {
	record, ok := corpus.byID[id]
	return record, ok
}

// Names returns every fixture name in sorted order.
func (corpus Corpus) Names() []string {
	names := make([]string, 0, len(corpus.Records))
	for _, record := range corpus.Records {
		names = append(names, record.Name)
	}
	sort.Strings(names)
	return names
}

var identityPattern = regexp.MustCompile(`\Asha256:[0-9a-f]{64}\z`)

// ValidIdentity reports whether id has the canonical domain-identity shape.
func ValidIdentity(id string) bool { return identityPattern.MatchString(id) }

// ParseCorpus decodes the four-line record stream, independently canonicalizes
// and hashes every payload, and rejects any record whose published identity is
// not the identity this package derives.
func ParseCorpus(payload []byte) (Corpus, error) {
	lines := strings.Split(strings.TrimSuffix(string(payload), "\n"), "\n")
	corpus := Corpus{byName: map[string]Record{}, byID: map[string]Record{}}
	for index := 0; index < len(lines); index++ {
		line := lines[index]
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "name=") {
			return Corpus{}, fmt.Errorf("line %d is not a record header: %q", index+1, line)
		}
		if index+3 >= len(lines) {
			return Corpus{}, fmt.Errorf("record at line %d is truncated", index+1)
		}
		record := Record{
			Name:      strings.TrimPrefix(line, "name="),
			Label:     strings.TrimPrefix(lines[index+1], "label="),
			Payload:   []byte(lines[index+2]),
			Published: lines[index+3],
		}
		if !strings.HasPrefix(lines[index+1], "label=") || record.Label == "" {
			return Corpus{}, fmt.Errorf("record %q has no domain label", record.Name)
		}
		if !ValidIdentity(record.Published) {
			return Corpus{}, fmt.Errorf("record %q publishes a malformed identity %q", record.Name, record.Published)
		}
		value, err := RequireCanonical(record.Payload)
		if err != nil {
			return Corpus{}, fmt.Errorf("record %q payload: %w", record.Name, err)
		}
		record.Value = value
		derived, err := DomainID(record.Label, record.Payload)
		if err != nil {
			return Corpus{}, fmt.Errorf("record %q identity: %w", record.Name, err)
		}
		record.Derived = derived
		if record.Derived != record.Published {
			return Corpus{}, fmt.Errorf("record %q derives %s but publishes %s", record.Name, record.Derived, record.Published)
		}
		if _, exists := corpus.byName[record.Name]; exists {
			return Corpus{}, fmt.Errorf("duplicate record name %q", record.Name)
		}
		corpus.byName[record.Name] = record
		// One record may legitimately appear under two fixture names: the
		// CGP05 Darwin platform node and the CGP10 platform node are the same
		// bytes under the same label, so they are the same graph record. What
		// may never happen is two different payloads or labels colliding on
		// one identity, which would break domain separation.
		if existing, exists := corpus.byID[record.Derived]; exists {
			if existing.Label != record.Label || string(existing.Payload) != string(record.Payload) {
				return Corpus{}, fmt.Errorf("records %q and %q collide on identity %s with different bytes", existing.Name, record.Name, record.Derived)
			}
		} else {
			corpus.byID[record.Derived] = record
		}
		corpus.Records = append(corpus.Records, record)
		index += 3
	}
	if len(corpus.Records) != AcceptedRecordCount {
		return Corpus{}, fmt.Errorf("corpus holds %d labeled records, want %d", len(corpus.Records), AcceptedRecordCount)
	}
	return corpus, nil
}

// AcceptedCorpus parses the embedded accepted corpus.
func AcceptedCorpus() (Corpus, error) {
	if len(acceptedCorpus) == 0 {
		return Corpus{}, errors.New("accepted corpus is empty")
	}
	return ParseCorpus(acceptedCorpus)
}
