package registry

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/protocoljson"
)

// Cache lifetimes (Spec §13.5).
const (
	DefaultCacheTTL     = time.Hour
	DefaultOfflineGrace = 7 * 24 * time.Hour
	maxResponseBytes    = 16 << 20
	maxRecordsPerQuery  = 10000
	maxPageSize         = 1000
	// RegistryMaxAttempts bounds one logical registry request, including its first attempt.
	RegistryMaxAttempts = 3
	// RegistryGetTotalDeadline bounds all GET attempts and retry delays.
	RegistryGetTotalDeadline = 30 * time.Second
	// RegistryPostTotalDeadline bounds all idempotent POST attempts and retry delays.
	RegistryPostTotalDeadline = 45 * time.Second
)

var httpClient = &http.Client{
	Timeout:       10 * time.Second,
	CheckRedirect: rejectRegistryRedirect,
}

var errRegistryResponseLimit = errors.New("registry response limit exceeded")

type registryUnavailableError struct {
	cause error
}

func (e *registryUnavailableError) Error() string { return e.cause.Error() }

func (e *registryUnavailableError) Unwrap() error { return e.cause }

func markRegistryUnavailable(err error) error {
	if err == nil {
		return nil
	}
	return &registryUnavailableError{cause: err}
}

func isRegistryUnavailable(err error) bool {
	var unavailable *registryUnavailableError
	return errors.As(err, &unavailable)
}

func isRegistryTransportFailure(err error) bool {
	if err == nil || errors.Is(err, context.Canceled) {
		return false
	}
	var networkError net.Error
	return errors.As(err, &networkError) ||
		errors.Is(err, context.DeadlineExceeded) ||
		errors.Is(err, io.EOF) ||
		errors.Is(err, io.ErrUnexpectedEOF)
}

type registryHTTPResult struct {
	status int
	header http.Header
	body   []byte
}

// RetryPermitted implements the protocol retry safety classification. Actual
// retries remain finite, preserve the exact request, and use the same
// idempotency key and body for publication requests.
func RetryPermitted(method, outcome string, hasIdempotencyKey bool) bool {
	if outcome != "network_error" && outcome != "429" && outcome != "503" {
		return false
	}
	switch method {
	case http.MethodGet:
		return true
	case http.MethodPost:
		return hasIdempotencyKey
	default:
		return false
	}
}

func rejectRegistryRedirect(_ *http.Request, _ []*http.Request) error {
	return http.ErrUseLastResponse
}

// HTTPGetSnapshot fetches GET <url>/v1/snapshot.
func HTTPGetSnapshot(baseURL string) (map[string]any, error) {
	endpoint := strings.TrimRight(baseURL, "/") + "/v1/snapshot"
	request, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/json")
	result, err := doRegistryRequest(httpClient, request, maxResponseBytes, RegistryGetTotalDeadline)
	if err != nil {
		return nil, err
	}
	if result.status < 200 || result.status >= 300 {
		return nil, registryStatusError("snapshot request", result)
	}
	if !isJSONResponse(result.header.Get("Content-Type")) {
		return nil, fmt.Errorf("registry snapshot response has unsupported content type %q", result.header.Get("Content-Type"))
	}
	var data map[string]any
	if err := decodeJSON(result.body, &data); err != nil {
		return nil, fmt.Errorf("registry returned invalid JSON: %v", err)
	}
	return data, nil
}

// FetchPolicy controls persistence independently for the response cache and
// the protected registry high-water state.
type FetchPolicy struct {
	PersistCache bool
	PersistState bool
}

// NewHTTPFetch builds a persistent records fetcher. Every network page must
// carry a verified v2 boundary before its records contribute; the boundary
// is checked against the shared snapshot high-water state (protocol §9.3).
func NewHTTPFetch(cacheDir, stateDir string, registries []Registry, ttl, grace time.Duration, now func() time.Time) FetchFn {
	if ttl == 0 {
		ttl = DefaultCacheTTL
	}
	if grace == 0 {
		grace = DefaultOfflineGrace
	}
	return newHTTPFetch(cacheDir, stateDir, registries, ttl, grace, now, FetchPolicy{PersistCache: true, PersistState: true}, nil)
}

// NewHTTPFetchWithPolicy is the manager-config form. Zero TTL and grace are
// literal and disable fresh-cache or stale-fallback use respectively.
func NewHTTPFetchWithPolicy(cacheDir, stateDir string, registries []Registry, ttl, grace time.Duration, now func() time.Time, policy FetchPolicy) FetchFn {
	return newHTTPFetch(cacheDir, stateDir, registries, ttl, grace, now, policy, nil)
}

// NewHTTPFetchWithPolicyAndObserver reports accepted page boundaries to the
// same operation-scoped mirror observer used for accepted snapshots.
func NewHTTPFetchWithPolicyAndObserver(cacheDir, stateDir string, registries []Registry, ttl, grace time.Duration, now func() time.Time, policy FetchPolicy, observer *MirrorViewObserver) FetchFn {
	return newHTTPFetch(cacheDir, stateDir, registries, ttl, grace, now, policy, observer)
}

// NewHTTPFetchWithPolicyReadOnly verifies network and cached boundaries
// against existing state, but never creates or updates either persistent
// cache or rollback state.
func NewHTTPFetchWithPolicyReadOnly(cacheDir, stateDir string, registries []Registry, ttl, grace time.Duration, now func() time.Time) FetchFn {
	return newHTTPFetch(cacheDir, stateDir, registries, ttl, grace, now, FetchPolicy{}, nil)
}

// NewHTTPFetchWithPolicyReadOnlyAndObserver is the read-only counterpart of
// NewHTTPFetchWithPolicyAndObserver.
func NewHTTPFetchWithPolicyReadOnlyAndObserver(cacheDir, stateDir string, registries []Registry, ttl, grace time.Duration, now func() time.Time, observer *MirrorViewObserver) FetchFn {
	return newHTTPFetch(cacheDir, stateDir, registries, ttl, grace, now, FetchPolicy{}, observer)
}

func newHTTPFetch(cacheDir, stateDir string, registries []Registry, ttl, grace time.Duration, now func() time.Time, policy FetchPolicy, observer *MirrorViewObserver) FetchFn {
	versioned := newHTTPFetchVersioned(cacheDir, stateDir, registries, ttl, grace, now, policy, observer)
	return func(baseURL, sourceIdentity, commit, contentSHA256 string) ([]map[string]any, error) {
		return versioned(baseURL, sourceIdentity, commit, contentSHA256, hashing.VersionV1)
	}
}

// NewHTTPFetchVersionedWithPolicyAndObserver is the versioned query form used
// by v2 content identities. The hash version is included in the request and
// cache key before any registry records are considered.
func NewHTTPFetchVersionedWithPolicyAndObserver(cacheDir, stateDir string, registries []Registry, ttl, grace time.Duration, now func() time.Time, policy FetchPolicy, observer *MirrorViewObserver) FetchVersionedFn {
	return newHTTPFetchVersioned(cacheDir, stateDir, registries, ttl, grace, now, policy, observer)
}

func newHTTPFetchVersioned(cacheDir, stateDir string, registries []Registry, ttl, grace time.Duration, now func() time.Time, policy FetchPolicy, observer *MirrorViewObserver) FetchVersionedFn {
	if now == nil {
		now = time.Now
	}
	if policy.PersistCache {
		_ = os.MkdirAll(cacheDir, 0o755)
	}
	byURL := make(map[string]Registry, len(registries))
	for _, reg := range registries {
		byURL[strings.TrimRight(reg.URL, "/")] = reg
	}
	return func(baseURL, sourceIdentity, commit, contentSHA256 string, hashVersion hashing.Version) ([]map[string]any, error) {
		if hashVersion != hashing.VersionV1 && hashVersion != hashing.VersionV2 {
			return nil, fmt.Errorf("unsupported content hash version %d", hashVersion)
		}
		reg, trusted := byURL[strings.TrimRight(baseURL, "/")]
		if !trusted || len(reg.PublicKeys) == 0 {
			return nil, fmt.Errorf("registry %s is not trusted: no pinned keys to verify its page boundary", baseURL)
		}
		chain, err := openPageChain(reg, stateDir, policy.PersistState, observer)
		if err != nil {
			return nil, err
		}
		endpoint, err := recordsEndpoint(baseURL, sourceIdentity, commit, contentSHA256, hashVersion)
		if err != nil {
			return nil, err
		}
		sum := sha256.Sum256([]byte(endpoint))
		cacheFile := filepath.Join(cacheDir, "records-"+hex.EncodeToString(sum[:])[:16]+".json")

		if fetchedAt, records, boundary, rawBoundary, ok := readRecordCache(cacheFile); ok && now().Sub(fetchedAt) >= 0 && now().Sub(fetchedAt) < ttl &&
			chain.cacheBoundaryCurrent(boundary, rawBoundary) {
			if observer != nil {
				if parsed, err := parseSnapshot(boundary); err == nil {
					observer.Observe(reg, parsed)
				}
			}
			return records, nil
		}
		records, err := httpGetAllRecords(chain, endpoint, reg, observer)
		if err != nil {
			var boundaryErr *PageBoundaryError
			var stateErr *pageStateError
			if errors.As(err, &boundaryErr) || errors.As(err, &stateErr) {
				// A rejected chain contributes nothing. Keep any pre-existing
				// cache entry on disk for a later genuine outage, but do not
				// overwrite it or serve it for this rejected operation.
				return nil, err
			}
			if fetchedAt, cached, boundary, rawBoundary, ok := readRecordCache(cacheFile); ok && now().Sub(fetchedAt) >= 0 && now().Sub(fetchedAt) < grace &&
				chain.cacheBoundaryCurrent(boundary, rawBoundary) {
				return cached, nil
			}
			return nil, err
		}
		if policy.PersistCache {
			writeRecordCache(cacheFile, now(), records, chain.chain)
		}
		return records, nil
	}
}

func recordsEndpoint(baseURL, sourceIdentity, commit, contentSHA256 string, hashVersion hashing.Version) (string, error) {
	if hashVersion != hashing.VersionV1 && hashVersion != hashing.VersionV2 {
		return "", fmt.Errorf("unsupported content hash version %d", hashVersion)
	}
	query := url.Values{}
	query.Set("source_identity", sourceIdentity)
	query.Set("commit", commit)
	query.Set("content_sha256", contentSHA256)
	if hashVersion == hashing.VersionV2 {
		query.Set("hash_version", "2")
	}
	query.Set("limit", fmt.Sprintf("%d", maxPageSize))
	return strings.TrimRight(baseURL, "/") + "/v1/records?" + query.Encode(), nil
}

func httpGetAllRecords(chain *pageChain, endpoint string, reg Registry, observer *MirrorViewObserver) ([]map[string]any, error) {
	records := make([]map[string]any, 0)
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
		page, next, err := httpGetRecordsPage(chain, pageURL.String(), reg, observer)
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

func httpGetRecordsPage(chain *pageChain, endpoint string, reg Registry, observer *MirrorViewObserver) ([]map[string]any, string, error) {
	request, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, "", err
	}
	request.Header.Set("Accept", "application/json")
	result, err := doRegistryRequest(httpClient, request, maxResponseBytes, RegistryGetTotalDeadline)
	if err != nil {
		return nil, "", err
	}
	if result.status < 200 || result.status >= 300 {
		return nil, "", registryStatusError("records request", result)
	}
	if !isJSONResponse(result.header.Get("Content-Type")) {
		return nil, "", fmt.Errorf("registry records response has unsupported content type %q", result.header.Get("Content-Type"))
	}
	var rawEnvelope map[string]json.RawMessage
	if err := decodeJSON(result.body, &rawEnvelope); err != nil {
		return nil, "", fmt.Errorf("registry returned invalid JSON: %v", err)
	}
	rawBoundary, present := rawEnvelope["boundary"]
	if !present {
		return nil, "", &PageBoundaryError{Diagnostic: PageBoundaryMissing, URL: chain.url, Detail: "served a records page with no boundary"}
	}
	if len(rawEnvelope) != 3 {
		return nil, "", fmt.Errorf("registry records response has unknown or missing fields")
	}
	for key := range rawEnvelope {
		if key != "records" && key != "next_cursor" && key != "boundary" {
			return nil, "", fmt.Errorf("registry records response has unknown or missing fields")
		}
	}
	var boundary map[string]any
	if err := decodeJSON(rawBoundary, &boundary); err != nil || boundary == nil {
		return nil, "", &PageBoundaryError{Diagnostic: PageBoundaryMissing, URL: chain.url, Detail: "served a records page whose boundary failed verification"}
	}
	if err := chain.acceptPage(boundary, rawBoundary); err != nil {
		return nil, "", err
	}
	if chain.firstUse && chain.persist && observer != nil {
		observer.ReportBootstrapTOFU(reg)
		chain.firstUse = false
	}
	if observer != nil {
		if parsed, err := parseSnapshot(boundary); err == nil {
			observer.Observe(reg, parsed)
		}
	}
	var data map[string]any
	if err := decodeJSON(result.body, &data); err != nil {
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
		if !ok || next == "" || utf8.RuneCountInString(next) > 4096 {
			return nil, "", fmt.Errorf("registry records response has invalid 'next_cursor'")
		}
	}
	return records, next, nil
}

func readRecordCache(path string) (time.Time, []map[string]any, map[string]any, []byte, bool) {
	payload, err := os.ReadFile(path) // #nosec G304 -- cache under the machine home
	if err != nil {
		return time.Time{}, nil, nil, nil, false
	}
	var data struct {
		FetchedAt float64          `json:"fetched_at"`
		Records   []map[string]any `json:"records"`
		Boundary  json.RawMessage  `json:"boundary"`
	}
	if err := decodeJSON(payload, &data); err != nil || data.Records == nil || len(data.Boundary) == 0 {
		return time.Time{}, nil, nil, nil, false
	}
	var boundary map[string]any
	if err := decodeJSON(data.Boundary, &boundary); err != nil || boundary == nil {
		return time.Time{}, nil, nil, nil, false
	}
	seconds := int64(data.FetchedAt)
	nanos := int64((data.FetchedAt - float64(seconds)) * 1e9)
	return time.Unix(seconds, nanos), data.Records, boundary, data.Boundary, true
}

func writeRecordCache(path string, fetchedAt time.Time, records []map[string]any, boundary []byte) {
	if len(boundary) == 0 {
		return
	}
	payload, err := json.Marshal(map[string]any{
		"fetched_at": float64(fetchedAt.UnixNano()) / 1e9,
		"records":    records,
		"boundary":   json.RawMessage(boundary),
	})
	if err != nil {
		return
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".records-*.tmp")
	if err != nil {
		return
	}
	temporaryPath := temporary.Name()
	defer func() { _ = os.Remove(temporaryPath) }()
	if _, err := temporary.Write(payload); err != nil || temporary.Chmod(0o600) != nil || temporary.Close() != nil {
		_ = temporary.Close()
		return
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		if removeErr := os.Remove(path); removeErr != nil && !os.IsNotExist(removeErr) {
			return
		}
		_ = os.Rename(temporaryPath, path)
	}
}

// Publish submits a signed record file to POST <url>/v1/records with a
// bearer token (Spec §13.10). Returns the registry response body.
func Publish(baseURL, token string, recordJSON []byte) (string, error) {
	var probe map[string]any
	if err := decodeJSON(recordJSON, &probe); err != nil {
		return "", fmt.Errorf("record is not valid JSON: %v", err)
	}
	if _, err := ParseRecord(probe); err != nil {
		return "", fmt.Errorf("record does not conform to audit-record-v1: %v", err)
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
	client := &http.Client{Timeout: 15 * time.Second, CheckRedirect: rejectRegistryRedirect}
	httpResult, err := doRegistryRequest(client, request, 1<<20, RegistryPostTotalDeadline)
	if err != nil {
		return "", fmt.Errorf("could not reach registry %s: %v", baseURL, err)
	}
	if httpResult.status >= 300 {
		return "", registryStatusError("record submission", httpResult)
	}
	if !isJSONResponse(httpResult.header.Get("Content-Type")) {
		return "", fmt.Errorf("registry submission response has unsupported content type %q", httpResult.header.Get("Content-Type"))
	}
	var result map[string]any
	if err := decodeJSON(httpResult.body, &result); err != nil {
		return "", fmt.Errorf("registry returned an invalid submission response: %v", err)
	}
	if unknown := unknownKeys(result, "seq", "entry_hash"); len(unknown) > 0 || len(result) != 2 {
		return "", fmt.Errorf("registry returned a malformed submission response")
	}
	sequence, ok := nonNegativeSafeInteger(result["seq"])
	entryHash, hashOK := result["entry_hash"].(string)
	if !ok || sequence < 1 || !hashOK || !hex256RE.MatchString(entryHash) {
		return "", fmt.Errorf("registry returned a malformed submission response")
	}
	return strings.TrimSpace(string(httpResult.body)), nil
}

func doRegistryRequest(client *http.Client, request *http.Request, responseLimit int64, totalDeadline time.Duration) (registryHTTPResult, error) {
	ctx, cancel := context.WithTimeout(request.Context(), totalDeadline)
	defer cancel()
	hasIdempotencyKey := request.Header.Get("Idempotency-Key") != ""
	var lastErr error
	for attempt := 1; attempt <= RegistryMaxAttempts; attempt++ {
		attemptRequest := request.Clone(ctx)
		if request.GetBody != nil {
			body, err := request.GetBody()
			if err != nil {
				return registryHTTPResult{}, err
			}
			attemptRequest.Body = body
		}
		response, err := client.Do(attemptRequest) // #nosec G107 -- registry URL comes from pinned machine config
		if err != nil {
			if isRegistryTransportFailure(err) {
				err = markRegistryUnavailable(err)
			}
			lastErr = err
			if attempt == RegistryMaxAttempts || !RetryPermitted(request.Method, "network_error", hasIdempotencyKey) ||
				!waitRegistryRetry(ctx, retryDelay(nil, attempt)) {
				return registryHTTPResult{}, lastErr
			}
			continue
		}
		body, readErr := readBounded(response.Body, responseLimit)
		_ = response.Body.Close()
		if readErr != nil {
			if isRegistryTransportFailure(readErr) {
				readErr = markRegistryUnavailable(readErr)
			}
			if !errors.Is(readErr, errRegistryResponseLimit) && attempt < RegistryMaxAttempts &&
				RetryPermitted(request.Method, "network_error", hasIdempotencyKey) &&
				waitRegistryRetry(ctx, retryDelay(nil, attempt)) {
				lastErr = readErr
				continue
			}
			return registryHTTPResult{}, readErr
		}
		result := registryHTTPResult{status: response.StatusCode, header: response.Header.Clone(), body: body}
		outcome := strconv.Itoa(response.StatusCode)
		if attempt == RegistryMaxAttempts || !RetryPermitted(request.Method, outcome, hasIdempotencyKey) {
			return result, nil
		}
		if !waitRegistryRetry(ctx, retryDelay(response.Header, attempt)) {
			return result, nil
		}
	}
	return registryHTTPResult{}, lastErr
}

func retryDelay(header http.Header, attempt int) time.Duration {
	if header != nil {
		value := strings.TrimSpace(header.Get("Retry-After"))
		if seconds, err := strconv.ParseInt(value, 10, 64); value != "" && err == nil && seconds >= 0 {
			return time.Duration(seconds) * time.Second
		}
		if when, err := http.ParseTime(value); value != "" && err == nil {
			if delay := time.Until(when); delay > 0 {
				return delay
			}
			return 0
		}
	}
	return time.Duration(100*(1<<(attempt-1))) * time.Millisecond
}

func waitRegistryRetry(ctx context.Context, delay time.Duration) bool {
	if deadline, ok := ctx.Deadline(); ok && delay > time.Until(deadline) {
		return false
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-timer.C:
		return true
	case <-ctx.Done():
		return false
	}
}

func registryStatusError(label string, result registryHTTPResult) error {
	code := ""
	var envelope map[string]any
	if err := decodeJSON(result.body, &envelope); err == nil {
		if errorValue, ok := envelope["error"].(map[string]any); ok {
			if candidate, ok := errorValue["code"].(string); ok && stableErrorCode(candidate) {
				code = candidate
			}
		}
	}
	var err error
	if code != "" {
		err = fmt.Errorf("registry %s failed (%d, %s)", label, result.status, code)
	} else {
		err = fmt.Errorf("registry %s failed (%d)", label, result.status)
	}
	if result.status >= http.StatusInternalServerError {
		return markRegistryUnavailable(err)
	}
	return err
}

func stableErrorCode(value string) bool {
	if value == "" || len(value) > 64 {
		return false
	}
	for _, character := range value {
		if (character < 'a' || character > 'z') && (character < '0' || character > '9') && character != '_' {
			return false
		}
	}
	return true
}

func readBounded(reader io.Reader, limit int64) ([]byte, error) {
	payload, err := io.ReadAll(io.LimitReader(reader, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(payload)) > limit {
		return nil, fmt.Errorf("registry response exceeds %d bytes: %w", limit, errRegistryResponseLimit)
	}
	return payload, nil
}

func isJSONResponse(contentType string) bool {
	mediaType := strings.TrimSpace(strings.SplitN(contentType, ";", 2)[0])
	return mediaType == "application/json"
}

func decodeJSON(payload []byte, target any) error {
	if err := protocoljson.Validate(payload); err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	decoder.DisallowUnknownFields()
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
