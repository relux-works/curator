package crossconformance

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/audit"
	"github.com/relux-works/curator/internal/buildmeta"
	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/conformancecoverage"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/marker"
	"github.com/relux-works/curator/internal/protocoljson"
	"github.com/relux-works/curator/internal/sourcelock"
)

// Schema-production-entry map. Every index.json entry is driven through
// the reader the manager actually uses on that document; the vendored
// JSON schemas document the oracle but no Go JSON Schema engine is
// executed. Cases with no byte-reader production entry are explicit
// bounds (see draftSchemaBounds), never passes.
var draftSchemaDrivers = map[string]func(t *testing.T, dir string, entry draftSchemaEntry){
	"skillfile-v2.schema.json":      driveSkillfileV2Case,
	"source-policy-v1.schema.json":  driveSourcePolicyCase,
	"source-policy-v2.schema.json":  driveSourcePolicyCase,
	"skillfile-lock-v1.schema.json": driveSkillfileLockCase,
	"local-snapshot-v1.schema.json": driveLocalSnapshotBound,
	"install-marker-v5.schema.json": driveInstallMarkerV5Case,
	"build-receipt-v3.schema.json":  driveBuildReceiptV3Case,
	"source-audit-v1.schema.json":   driveSourceAuditCase,
}

// draftSchemaBounds records every schema case that cannot be driven at
// a production entry, with the exact reason. The harness asserts this
// set exactly: a fixed gap must convert its row to a drive, never
// silently pass.
var draftSchemaBounds = map[string]string{
	// No production entry consumes local-snapshot inventory documents
	// as bytes: Capture/BuildInventory derive inventories from live
	// trees and OpenLocal re-derives from the store tree, so the wire
	// shape is proven by the snapshot vectors through Capture, not by
	// parsing these documents.
	"schema-cases/local-snapshot-v1/valid.json":                     "no byte-reader production entry for inventory documents; shape proven by Capture vectors",
	"schema-cases/local-snapshot-v1/invalid-algorithm.json":         "no byte-reader production entry for inventory documents; shape proven by Capture vectors",
	"schema-cases/local-snapshot-v1/invalid-unknown-top-level.json": "no byte-reader production entry for inventory documents; shape proven by Capture vectors",
}

// TestDraftSourcesSchemaCases drives all 116 pinned schema cases.
func TestDraftSourcesSchemaCases(t *testing.T) {
	dir := draftCorpusDir(t)
	entries := loadDraftIndex(t)
	seenBounds := map[string]bool{}
	conformancecoverage.RunOutcomes(t, "draft-sources-v1/schema-cases", entries,
		func(entry draftSchemaEntry) string { return entry.Instance }, func(t *testing.T, entry draftSchemaEntry) conformancecoverage.Observation {
			driver, ok := draftSchemaDrivers[entry.Schema]
			if !ok {
				t.Fatalf("schema %q has no production-entry driver", entry.Schema)
			}
			if reason, isBound := draftSchemaBounds[entry.Instance]; isBound {
				seenBounds[entry.Instance] = true
				driver(t, dir, entry)
				t.Logf("BOUND: %s", reason)
				return conformancecoverage.Observation{BoundReason: reason}
			}
			driver(t, dir, entry)
			return conformancecoverage.Observation{}
		})
	for instance := range draftSchemaBounds {
		if !seenBounds[instance] {
			t.Errorf("bound case %q not present in the pinned index", instance)
		}
	}
}

func driveSkillfileV2Case(t *testing.T, dir string, entry draftSchemaEntry) {
	t.Helper()
	payload := readDraftFile(t, filepath.Join(dir, filepath.FromSlash(entry.Instance)))
	project := t.TempDir()
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), payload, 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := manifest.Load(project)
	if (err == nil) != entry.Valid {
		t.Fatalf("valid=%v: Load err=%v", entry.Valid, err)
	}
}

func driveSourcePolicyCase(t *testing.T, dir string, entry draftSchemaEntry) {
	t.Helper()
	payload := readDraftFile(t, filepath.Join(dir, filepath.FromSlash(entry.Instance)))
	_, err := config.ParseSourcePolicy(payload, entry.Instance)
	if (err == nil) != entry.Valid {
		t.Fatalf("valid=%v: ParseSourcePolicy err=%v", entry.Valid, err)
	}
	if err != nil && !strings.Contains(err.Error(), config.CodeRepositoryPolicyInvalid) {
		t.Fatalf("error %q lacks class %q", err, config.CodeRepositoryPolicyInvalid)
	}
}

func driveSkillfileLockCase(t *testing.T, dir string, entry draftSchemaEntry) {
	t.Helper()
	payload := readDraftFile(t, filepath.Join(dir, filepath.FromSlash(entry.Instance)))
	lock, err := sourcelock.Parse(payload)
	base := filepath.Base(entry.Instance)
	switch base {
	case "valid.json":
		if err != nil {
			t.Fatalf("valid lock refused: %v", err)
		}
		if lock == nil || len(lock.Members) != 1 {
			t.Fatalf("valid lock parsed to %+v", lock)
		}
	case "valid-git.json", "valid-configured-git.json":
		// Schema-valid but normatively inconsistent: the member
		// directory disagrees with the package directory, which JSON
		// Schema cannot express and sourcelock.Parse refuses before
		// any integrity consult.
		if err == nil || !strings.Contains(err.Error(), "source_member_invalid") {
			t.Fatalf("git fixture err = %v, want source_member_invalid directory disagreement", err)
		}
	default:
		if entry.Valid {
			t.Fatalf("no oracle for valid lock case %q", base)
		}
		if err == nil {
			t.Fatalf("invalid lock case %q accepted", base)
		}
		if !strings.Contains(err.Error(), "source_member_invalid") && !strings.Contains(err.Error(), "source_selection_invalid") {
			t.Fatalf("invalid lock case %q err = %v, want a source diagnostic", base, err)
		}
	}
}

func driveLocalSnapshotBound(t *testing.T, dir string, entry draftSchemaEntry) {
	t.Helper()
	// Bound by design (see draftSchemaBounds): the only assertion is
	// that the fixture is still the pinned document shape, so a
	// corpus revision that adds a reader-visible field forces this
	// row to be revisited rather than silently passing.
	payload := readDraftFile(t, filepath.Join(dir, filepath.FromSlash(entry.Instance)))
	var doc map[string]any
	if err := json.Unmarshal(payload, &doc); err != nil {
		t.Fatalf("fixture is not JSON: %v", err)
	}
	if doc["schema_version"] != float64(1) {
		t.Fatalf("fixture schema_version = %v, want 1", doc["schema_version"])
	}
	if _, ok := draftSchemaBounds[entry.Instance]; !ok {
		t.Fatalf("local-snapshot case %q drove without a bound reason", entry.Instance)
	}
}

func driveInstallMarkerV5Case(t *testing.T, dir string, entry draftSchemaEntry) {
	t.Helper()
	payload := readDraftFile(t, filepath.Join(dir, filepath.FromSlash(entry.Instance)))
	installed := t.TempDir()
	if err := os.WriteFile(filepath.Join(installed, marker.Name), payload, 0o644); err != nil {
		t.Fatal(err)
	}
	got := marker.Read(installed)
	if (got != nil) != entry.Valid {
		t.Fatalf("valid=%v: marker.Read nil=%v", entry.Valid, got == nil)
	}
}

func driveBuildReceiptV3Case(t *testing.T, dir string, entry draftSchemaEntry) {
	t.Helper()
	payload := readDraftFile(t, filepath.Join(dir, filepath.FromSlash(entry.Instance)))
	// DecodeReceipt's contract is canonical CCJ-1 bytes, the exact
	// form production writers emit; the published fixtures are
	// pretty-printed JSON, so the harness applies the production
	// canonicalizer before the production reader. The seam is explicit
	// and both directions are verified below.
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		t.Fatalf("fixture is not JSON: %v", err)
	}
	canonical, err := protocoljson.MarshalCanonical(value)
	if err != nil {
		t.Fatalf("canonicalize: %v", err)
	}
	_, err = buildmeta.DecodeReceipt(canonical)
	if (err == nil) != entry.Valid {
		t.Fatalf("valid=%v: DecodeReceipt err=%v", entry.Valid, err)
	}
}

func driveSourceAuditCase(t *testing.T, dir string, entry draftSchemaEntry) {
	t.Helper()
	payload := readDraftFile(t, filepath.Join(dir, filepath.FromSlash(entry.Instance)))
	_, err := audit.ParseSourceAudit(payload)
	if (err == nil) != entry.Valid {
		t.Fatalf("valid=%v: ParseSourceAudit err=%v", entry.Valid, err)
	}
}
