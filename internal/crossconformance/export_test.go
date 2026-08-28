package crossconformance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/crossconformance"
)

const exportGoldenPath = "testdata/cross-adapter-protocol-export.json"

// The export is the only thing an independent implementation needs in order to
// run the same integration contract. It is committed, so a change to the
// corpus, the obligation set, or the rejection matrix shows up as a reviewable
// diff instead of a silent protocol change.
func TestProtocolExportMatchesTheCommittedGolden(t *testing.T) {
	payload := mustExport(t)
	golden, err := os.ReadFile(exportGoldenPath)
	if err != nil {
		t.Fatalf("%v (regenerate with CURATOR_WRITE_CROSS_EXPORT=1)", err)
	}
	if os.Getenv("CURATOR_WRITE_CROSS_EXPORT") == "1" {
		if err = os.WriteFile(exportGoldenPath, payload, 0o600); err != nil {
			t.Fatal(err)
		}
		t.Fatal("regenerated the committed export; rerun without CURATOR_WRITE_CROSS_EXPORT")
	}
	if string(golden) != string(payload) {
		t.Fatalf("committed export differs from the derived one\nderived digest %s\ncommitted digest %s",
			crossconformance.ExportDigest(payload), crossconformance.ExportDigest(golden))
	}
}

// The export must be exact CCJ-1 so a peer implementation can hash it the same
// way, and it must round-trip through this package's own scanner.
func TestProtocolExportIsCanonicalAndSelfDescribing(t *testing.T) {
	payload := mustExport(t)
	value, err := crossconformance.RequireCanonical(payload)
	if err != nil {
		t.Fatalf("export is not canonical CCJ-1: %v", err)
	}
	if value.String("schema_id") != crossconformance.ExportSchemaID {
		t.Fatalf("export schema = %q", value.String("schema_id"))
	}
	if value.String("corpus", "sha256") != crossconformance.AcceptedCorpusSHA256 {
		t.Fatalf("export does not name the accepted corpus bytes")
	}
	records, _ := value.Member("corpus")
	list, _ := records.Member("records")
	if len(list.Items) != crossconformance.AcceptedRecordCount {
		t.Fatalf("export carries %d records, want %d", len(list.Items), crossconformance.AcceptedRecordCount)
	}
	for _, record := range list.Items {
		payloadValue, ok := record.Member("payload")
		if !ok {
			t.Fatalf("exported record %q has no payload", record.String("name"))
		}
		canonical, canonicalErr := crossconformance.Canonical(payloadValue)
		if canonicalErr != nil {
			t.Fatal(canonicalErr)
		}
		derived, idErr := crossconformance.DomainID(record.String("label"), canonical)
		if idErr != nil {
			t.Fatal(idErr)
		}
		if derived != record.String("id") {
			t.Errorf("exported record %q publishes %s but its payload derives %s", record.String("name"), record.String("id"), derived)
		}
	}
	obligations, _ := value.Member("obligations")
	if len(obligations.Items) != len(crossconformance.Obligations()) {
		t.Fatalf("export carries %d obligations, want %d", len(obligations.Items), len(crossconformance.Obligations()))
	}
	for _, obligation := range obligations.Items {
		if obligation.String("requirement") == "" {
			t.Errorf("exported obligation %q states no requirement", obligation.String("id"))
		}
	}
	matrix, _ := value.Member("rejection_matrix")
	if len(matrix.Items) != len(crossconformance.RejectionVectors()) {
		t.Fatalf("export carries %d rejection vectors, want %d", len(matrix.Items), len(crossconformance.RejectionVectors()))
	}
	for _, vector := range matrix.Items {
		codes := vector.Strings("codes")
		if len(codes) == 0 {
			t.Errorf("exported vector %q publishes no stable diagnostic", vector.String("id"))
		}
		for _, code := range codes {
			if !isStableCode(code) {
				t.Errorf("exported vector %q publishes non-stable diagnostic %q", vector.String("id"), code)
			}
		}
	}
}

// The export names the Go packages that own each delegated vector. Those names
// must be real packages in this repository, otherwise the delegation is a
// promise rather than a pointer.
func TestDelegatedVectorsNameRealOwningPackages(t *testing.T) {
	for _, vector := range crossconformance.RejectionVectors() {
		if vector.CrossDrivable() {
			continue
		}
		if len(vector.OwnedBy) == 0 {
			t.Errorf("%s is not cross-drivable and names no owner", vector.ID)
		}
		for _, owner := range vector.OwnedBy {
			if !strings.HasPrefix(owner, "internal/") {
				t.Errorf("%s names owner %q outside internal/", vector.ID, owner)
				continue
			}
			info, err := os.Stat(filepath.Join("..", "..", filepath.FromSlash(owner)))
			if err != nil || !info.IsDir() {
				t.Errorf("%s names owner %q, which is not a package directory: %v", vector.ID, owner, err)
			}
		}
	}
}

func mustExport(t *testing.T) []byte {
	t.Helper()
	corpus := mustCorpus(t)
	report, err := crossconformance.Validate(corpus)
	if err != nil {
		t.Fatal(err)
	}
	payload, err := crossconformance.ProtocolExport(corpus, report)
	if err != nil {
		t.Fatal(err)
	}
	return payload
}
