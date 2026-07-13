package registry

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Cache lifetimes (Spec §13.5).
const (
	DefaultCacheTTL     = time.Hour
	DefaultOfflineGrace = 7 * 24 * time.Hour
	maxResponseBytes    = 16 << 20
	maxRecordsPerQuery  = 10000
	maxPageSize         = 1000
)

var httpClient = &http.Client{Timeout: 10 * time.Second}

// HTTPGetSnapshot fetches GET <url>/v1/snapshot.
func HTTPGetSnapshot(baseURL string) (map[string]any, error) {
	endpoint := strings.TrimRight(baseURL, "/") + "/v1/snapshot"
	request, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/json")
	response, err := httpClient.Do(request) // #nosec G107 -- registry URL comes from pinned machine config
	if err != nil {
		return nil, err
	}
	defer func() { _ = response.Body.Close() }()
	payload, err := readBounded(response.Body, maxResponseBytes)
	if err != nil {
		return nil, err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("registry snapshot request failed (%d): %s", response.StatusCode, strings.TrimSpace(string(payload)))
	}
	if !isJSONResponse(response.Header.Get("Content-Type")) {
		return nil, fmt.Errorf("registry snapshot response has unsupported content type %q", response.Header.Get("Content-Type"))
	}
	var data map[string]any
	if err := decodeJSON(payload, &data); err != nil {
		return nil, fmt.Errorf("registry returned invalid JSON: %v", err)
	}
	return data, nil
}

// NewHTTPFetch builds a FetchFn over GET /v1/records with an on-disk cache:
// a fresh entry is served without a network call; on refresh failure a stale
// entry stays usable within the offline grace window (Spec §13.5).
func NewHTTPFetch(cacheDir string, ttl, grace time.Duration, now func() time.Time) FetchFn {
	if ttl == 0 {
		ttl = DefaultCacheTTL
	}
	if grace == 0 {
		grace = DefaultOfflineGrace
	}
	if now == nil {
		now = time.Now
	}
	_ = os.MkdirAll(cacheDir, 0o755)
	return func(baseURL, sourceIdentity, commit, contentSHA256 string) ([]map[string]any, error) {
		query := url.Values{}
		query.Set("source_identity", sourceIdentity)
		query.Set("commit", commit)
		query.Set("content_sha256", contentSHA256)
		query.Set("limit", fmt.Sprintf("%d", maxPageSize))
		endpoint := strings.TrimRight(baseURL, "/") + "/v1/records?" + query.Encode()
		sum := sha256.Sum256([]byte(endpoint))
		cacheFile := filepath.Join(cacheDir, "records-"+hex.EncodeToString(sum[:])[:16]+".json")

		if fetchedAt, records, ok := readRecordCache(cacheFile); ok && now().Sub(fetchedAt) < ttl {
			return records, nil
		}
		records, err := httpGetAllRecords(endpoint)
		if err != nil {
			if fetchedAt, cached, ok := readRecordCache(cacheFile); ok && now().Sub(fetchedAt) < grace {
				return cached, nil
			}
			return nil, err
		}
		writeRecordCache(cacheFile, now(), records)
		return records, nil
	}
}

func httpGetAllRecords(endpoint string) ([]map[string]any, error) {
	var records []map[string]any
	seenCursors := map[string]bool{}
	cursor := ""
	for {
		pageURL, err := url.Parse(endpoint)
		if err != nil {
			return nil, err
		}
		query := pageURL.Query()
		if cursor == "" {
			query.Del("cursor")
		} else {
			query.Set("cursor", cursor)
		}
		pageURL.RawQuery = query.Encode()
		page, next, err := httpGetRecordsPage(pageURL.String())
		if err != nil {
			return nil, err
		}
		records = append(records, page...)
		if len(records) > maxRecordsPerQuery {
			return nil, fmt.Errorf("registry returned more than %d records for one artifact", maxRecordsPerQuery)
		}
		if next == "" {
			return records, nil
		}
		if seenCursors[next] {
			return nil, fmt.Errorf("registry returned a repeated pagination cursor")
		}
		seenCursors[next] = true
		cursor = next
	}
}

func httpGetRecordsPage(endpoint string) ([]map[string]any, string, error) {
	request, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, "", err
	}
	request.Header.Set("Accept", "application/json")
	response, err := httpClient.Do(request) // #nosec G107 -- registry URL comes from pinned machine config
	if err != nil {
		return nil, "", err
	}
	defer func() { _ = response.Body.Close() }()
	payload, err := readBounded(response.Body, maxResponseBytes)
	if err != nil {
		return nil, "", err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, "", fmt.Errorf("registry records request failed (%d): %s", response.StatusCode, strings.TrimSpace(string(payload)))
	}
	if !isJSONResponse(response.Header.Get("Content-Type")) {
		return nil, "", fmt.Errorf("registry records response has unsupported content type %q", response.Header.Get("Content-Type"))
	}
	var data map[string]any
	if err := decodeJSON(payload, &data); err != nil {
		return nil, "", fmt.Errorf("registry returned invalid JSON: %v", err)
	}
	rawRecords, ok := data["records"].([]any)
	if !ok {
		return nil, "", fmt.Errorf("registry records response field 'records' must be an array")
	}
	if len(rawRecords) > maxPageSize {
		return nil, "", fmt.Errorf("registry records page exceeds limit %d", maxPageSize)
	}
	records := make([]map[string]any, len(rawRecords))
	for index, raw := range rawRecords {
		record, ok := raw.(map[string]any)
		if !ok {
			return nil, "", fmt.Errorf("registry record %d must be an object", index)
		}
		records[index] = record
	}
	nextRaw, present := data["next_cursor"]
	if !present {
		return nil, "", fmt.Errorf("registry records response requires 'next_cursor'")
	}
	next := ""
	if nextRaw != nil {
		var ok bool
		next, ok = nextRaw.(string)
		if !ok || next == "" {
			return nil, "", fmt.Errorf("registry records response has invalid 'next_cursor'")
		}
	}
	return records, next, nil
}

func readRecordCache(path string) (time.Time, []map[string]any, bool) {
	payload, err := os.ReadFile(path) // #nosec G304 -- cache under the machine home
	if err != nil {
		return time.Time{}, nil, false
	}
	var data struct {
		FetchedAt float64          `json:"fetched_at"`
		Records   []map[string]any `json:"records"`
	}
	if err := decodeJSON(payload, &data); err != nil {
		return time.Time{}, nil, false
	}
	seconds := int64(data.FetchedAt)
	nanos := int64((data.FetchedAt - float64(seconds)) * 1e9)
	return time.Unix(seconds, nanos), data.Records, true
}

func writeRecordCache(path string, fetchedAt time.Time, records []map[string]any) {
	payload, err := json.Marshal(map[string]any{
		"fetched_at": float64(fetchedAt.UnixNano()) / 1e9,
		"records":    records,
	})
	if err != nil {
		return
	}
	_ = os.WriteFile(path, payload, 0o644)
}

// Publish submits a signed record file to POST <url>/v1/records with a
// bearer token (Spec §13.10). Returns the registry response body.
func Publish(baseURL, token string, recordJSON []byte) (string, error) {
	var probe map[string]any
	if err := decodeJSON(recordJSON, &probe); err != nil {
		return "", fmt.Errorf("record is not valid JSON: %v", err)
	}
	canonical, err := CanonicalBytesChecked(probe)
	if err != nil {
		return "", fmt.Errorf("record is not valid CCJ-1: %v", err)
	}
	endpoint := strings.TrimRight(baseURL, "/") + "/v1/records"
	request, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(recordJSON))
	if err != nil {
		return "", err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Authorization", "Bearer "+token)
	idempotency := sha256.Sum256(canonical)
	request.Header.Set("Idempotency-Key", hex.EncodeToString(idempotency[:]))
	client := &http.Client{Timeout: 15 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return "", fmt.Errorf("could not reach registry %s: %v", baseURL, err)
	}
	defer func() { _ = response.Body.Close() }()
	body, err := readBounded(response.Body, 1<<20)
	if err != nil {
		return "", err
	}
	if response.StatusCode >= 300 {
		return "", fmt.Errorf("registry rejected the record (%d): %s", response.StatusCode, strings.TrimSpace(string(body)))
	}
	if !isJSONResponse(response.Header.Get("Content-Type")) {
		return "", fmt.Errorf("registry submission response has unsupported content type %q", response.Header.Get("Content-Type"))
	}
	return strings.TrimSpace(string(body)), nil
}

func readBounded(reader io.Reader, limit int64) ([]byte, error) {
	payload, err := io.ReadAll(io.LimitReader(reader, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(payload)) > limit {
		return nil, fmt.Errorf("registry response exceeds %d bytes", limit)
	}
	return payload, nil
}

func isJSONResponse(contentType string) bool {
	mediaType := strings.TrimSpace(strings.SplitN(contentType, ";", 2)[0])
	return mediaType == "application/json"
}

func decodeJSON(payload []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return fmt.Errorf("trailing JSON value")
		}
		return err
	}
	return nil
}
