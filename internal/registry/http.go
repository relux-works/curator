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
	payload, err := io.ReadAll(io.LimitReader(response.Body, 8<<20))
	if err != nil {
		return nil, err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("registry snapshot request failed (%d): %s", response.StatusCode, strings.TrimSpace(string(payload)))
	}
	var data map[string]any
	if err := json.Unmarshal(payload, &data); err != nil {
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
		endpoint := strings.TrimRight(baseURL, "/") + "/v1/records?" + query.Encode()
		sum := sha256.Sum256([]byte(endpoint))
		cacheFile := filepath.Join(cacheDir, "records-"+hex.EncodeToString(sum[:])[:16]+".json")

		if fetchedAt, records, ok := readRecordCache(cacheFile); ok && now().Sub(fetchedAt) < ttl {
			return records, nil
		}
		records, err := httpGetRecords(endpoint)
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

func httpGetRecords(endpoint string) ([]map[string]any, error) {
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
	payload, err := io.ReadAll(io.LimitReader(response.Body, 8<<20))
	if err != nil {
		return nil, err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("registry records request failed (%d): %s", response.StatusCode, strings.TrimSpace(string(payload)))
	}
	var data struct {
		Records []map[string]any `json:"records"`
	}
	if err := json.Unmarshal(payload, &data); err != nil {
		return nil, fmt.Errorf("registry returned invalid JSON: %v", err)
	}
	return data.Records, nil
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
	if err := json.Unmarshal(payload, &data); err != nil {
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
	var probe any
	if err := json.Unmarshal(recordJSON, &probe); err != nil {
		return "", fmt.Errorf("record is not valid JSON: %v", err)
	}
	endpoint := strings.TrimRight(baseURL, "/") + "/v1/records"
	request, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(recordJSON))
	if err != nil {
		return "", err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+token)
	client := &http.Client{Timeout: 15 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return "", fmt.Errorf("could not reach registry %s: %v", baseURL, err)
	}
	defer func() { _ = response.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return "", err
	}
	if response.StatusCode >= 300 {
		return "", fmt.Errorf("registry rejected the record (%d): %s", response.StatusCode, strings.TrimSpace(string(body)))
	}
	return strings.TrimSpace(string(body)), nil
}
