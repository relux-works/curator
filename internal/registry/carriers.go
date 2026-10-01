package registry

import (
	"fmt"
	"unicode/utf8"
)

// LogEntry is one authenticated registry-log row. ParseLogEntry is the
// frozen v1 reader; ParseLogEntryVersioned accepts the v2 container and its
// historical v1 or current v2 audit records.
type LogEntry struct {
	Sequence  int
	EntryHash string
	PrevHash  string
	Record    Record
}

// ParseLogEntry reads the frozen registry-log-entry-v1 shape.
func ParseLogEntry(payload map[string]any) (LogEntry, error) {
	return parseLogEntry(payload, false)
}

// ParseLogEntryVersioned reads registry-log-entry-v2, including historical
// v1 records and v2 records that carry hash_version 2.
func ParseLogEntryVersioned(payload map[string]any) (LogEntry, error) {
	return parseLogEntry(payload, true)
}

func parseLogEntry(payload map[string]any, allowVersion2 bool) (LogEntry, error) {
	if unknown := unknownKeys(payload, "seq", "entry_hash", "prev_hash", "record"); len(unknown) > 0 || len(payload) != 4 {
		return LogEntry{}, fmt.Errorf("registry log entry must contain exactly seq, entry_hash, prev_hash, and record")
	}
	sequence, ok := nonNegativeSafeInteger(payload["seq"])
	entryHash, hashOK := payload["entry_hash"].(string)
	prevHash, prevOK := payload["prev_hash"].(string)
	recordPayload, recordOK := payload["record"].(map[string]any)
	if !ok || sequence < 1 || !hashOK || !hex256RE.MatchString(entryHash) || !prevOK || !hex256RE.MatchString(prevHash) || !recordOK {
		return LogEntry{}, fmt.Errorf("registry log entry has invalid sequence, hashes, or record")
	}
	var record Record
	var err error
	if allowVersion2 {
		record, err = ParseRecordVersioned(recordPayload)
	} else {
		record, err = ParseRecord(recordPayload)
	}
	if err != nil {
		return LogEntry{}, fmt.Errorf("registry log entry record: %w", err)
	}
	return LogEntry{Sequence: sequence, EntryHash: entryHash, PrevHash: prevHash, Record: record}, nil
}

// BundleV2 is a schema-validated offline registry bundle. Signature and
// checkpoint admission remain the importing caller's responsibility.
type BundleV2 struct {
	Records   []Record
	Snapshot  map[string]any
	PublicKey string
}

// ParseBundleV2 validates the v2 envelope and parses every historical or
// v2 record with an explicit content-hash version.
func ParseBundleV2(payload map[string]any) (BundleV2, error) {
	if unknown := unknownKeys(payload, "schema_version", "records", "snapshot", "public_key"); len(unknown) > 0 || len(payload) != 4 {
		return BundleV2{}, fmt.Errorf("registry bundle v2 must contain exactly schema_version, records, snapshot, and public_key")
	}
	if !integerEquals(payload["schema_version"], 2) {
		return BundleV2{}, fmt.Errorf("registry bundle schema_version must be 2")
	}
	rawRecords, ok := payload["records"].([]any)
	if !ok || len(rawRecords) > 1_000_000 {
		return BundleV2{}, fmt.Errorf("registry bundle records must be an array of at most 1000000 items")
	}
	snapshot, ok := payload["snapshot"].(map[string]any)
	if !ok {
		return BundleV2{}, fmt.Errorf("registry bundle snapshot must be an object")
	}
	if _, err := parseSnapshot(snapshot); err != nil {
		return BundleV2{}, fmt.Errorf("registry bundle snapshot: %w", err)
	}
	publicKey, ok := payload["public_key"].(string)
	if !ok {
		return BundleV2{}, fmt.Errorf("registry bundle public_key must be a string")
	}
	if _, err := ParsePublicKey(publicKey); err != nil {
		return BundleV2{}, fmt.Errorf("registry bundle public_key: %w", err)
	}
	result := BundleV2{Records: make([]Record, 0, len(rawRecords)), Snapshot: snapshot, PublicKey: publicKey}
	for i, raw := range rawRecords {
		recordPayload, ok := raw.(map[string]any)
		if !ok {
			return BundleV2{}, fmt.Errorf("registry bundle record %d must be an object", i)
		}
		record, err := ParseRecordVersioned(recordPayload)
		if err != nil {
			return BundleV2{}, fmt.Errorf("registry bundle record %d: %w", i, err)
		}
		result.Records = append(result.Records, record)
	}
	return result, nil
}

// LogResponseV3 is one paged registry-log response at a signed snapshot
// boundary. Its entry records preserve the hash version from each audit
// record; the boundary has its independent snapshot identity.
type LogResponseV3 struct {
	Entries    []LogEntry
	NextCursor *string
	Boundary   map[string]any
}

// ParseLogResponseV3 validates the response envelope and parses each v2
// container entry without widening the frozen log-response-v2 reader.
func ParseLogResponseV3(payload map[string]any) (LogResponseV3, error) {
	if unknown := unknownKeys(payload, "entries", "next_cursor", "boundary"); len(unknown) > 0 || len(payload) != 3 {
		return LogResponseV3{}, fmt.Errorf("log response v3 must contain exactly entries, next_cursor, and boundary")
	}
	rawEntries, ok := payload["entries"].([]any)
	if !ok || len(rawEntries) > maxPageSize {
		return LogResponseV3{}, fmt.Errorf("log response entries must be an array of at most %d items", maxPageSize)
	}
	var nextCursor *string
	if rawCursor := payload["next_cursor"]; rawCursor != nil {
		cursor, ok := rawCursor.(string)
		if !ok || cursor == "" || utf8.RuneCountInString(cursor) > 4096 {
			return LogResponseV3{}, fmt.Errorf("log response next_cursor must be null or a 1-4096 character string")
		}
		nextCursor = &cursor
	}
	boundary, ok := payload["boundary"].(map[string]any)
	if !ok {
		return LogResponseV3{}, fmt.Errorf("log response boundary must be a snapshot object")
	}
	if _, err := parseSnapshot(boundary); err != nil {
		return LogResponseV3{}, fmt.Errorf("log response boundary: %w", err)
	}
	result := LogResponseV3{Entries: make([]LogEntry, 0, len(rawEntries)), NextCursor: nextCursor, Boundary: boundary}
	for i, raw := range rawEntries {
		entryPayload, ok := raw.(map[string]any)
		if !ok {
			return LogResponseV3{}, fmt.Errorf("log response entry %d must be an object", i)
		}
		entry, err := ParseLogEntryVersioned(entryPayload)
		if err != nil {
			return LogResponseV3{}, fmt.Errorf("log response entry %d: %w", i, err)
		}
		result.Entries = append(result.Entries, entry)
	}
	return result, nil
}
