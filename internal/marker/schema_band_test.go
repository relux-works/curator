package marker

import "testing"

// TestSupportedSchemaBandIsExactAndNewestIsItsMaximum pins the two predicates
// every reader outside this package now bands on.
//
// Both directions matter. A band that is too narrow reports an exactly current
// installation as unreadable or as needing reinstallation the moment the
// written schema advances; a band that is too wide claims a document from a
// newer manager is understood. NewestSchemaVersion is what an operator is told
// is the newest readable schema, so it is proven to actually be the maximum of
// the band rather than a constant someone forgot to advance.
func TestSupportedSchemaBandIsExactAndNewestIsItsMaximum(t *testing.T) {
	supported := map[int]bool{
		LegacySchemaVersion:   true,
		SchemaVersion:         true,
		ExternalSchemaVersion: true,
		PolicySchemaVersion:   true,
	}
	for version := -1; version <= NewestSchemaVersion+3; version++ {
		if got := SupportedSchema(version); got != supported[version] {
			t.Fatalf("SupportedSchema(%d) = %v, want %v", version, got, supported[version])
		}
	}
	if !SupportedSchema(NewestSchemaVersion) {
		t.Fatalf("NewestSchemaVersion %d is not itself readable", NewestSchemaVersion)
	}
	if SupportedSchema(NewestSchemaVersion + 1) {
		t.Fatalf("schema %d is readable, so NewestSchemaVersion %d understates the band",
			NewestSchemaVersion+1, NewestSchemaVersion)
	}
	for version := range supported {
		if version > NewestSchemaVersion {
			t.Fatalf("readable schema %d exceeds NewestSchemaVersion %d", version, NewestSchemaVersion)
		}
	}
}

// TestBuildBearingSchemaCoversEverySchemaThatCanRecordABuild proves the
// predicate that decides whether a recorded compiled command is knowable. Only
// schema 1 predates the build record; every later readable schema carries it,
// including the schema-8 marker v4 whose absence from this band made status
// report every compiled command of a freshly installed skill as needs-install.
func TestBuildBearingSchemaCoversEverySchemaThatCanRecordABuild(t *testing.T) {
	bearing := map[int]bool{
		LegacySchemaVersion:   false,
		SchemaVersion:         true,
		ExternalSchemaVersion: true,
		PolicySchemaVersion:   true,
	}
	for version := -1; version <= NewestSchemaVersion+3; version++ {
		if got := BuildBearingSchema(version); got != bearing[version] {
			t.Fatalf("BuildBearingSchema(%d) = %v, want %v", version, got, bearing[version])
		}
	}
	// A build-bearing schema is necessarily a readable one: nothing outside the
	// band may be treated as carrying a build record.
	for version := -1; version <= NewestSchemaVersion+3; version++ {
		if BuildBearingSchema(version) && !SupportedSchema(version) {
			t.Fatalf("schema %d bears builds but is not readable", version)
		}
	}
}
