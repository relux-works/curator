package registry

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// signer builds signed registry objects for tests.
type signer struct {
	private ed25519.PrivateKey
	pinned  string
}

func newSigner(t *testing.T) *signer {
	t.Helper()
	public, private, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	return &signer{
		private: private,
		pinned:  "ed25519:" + base64.StdEncoding.EncodeToString(public),
	}
}

func (s *signer) sign(body map[string]any) map[string]any {
	record := map[string]any{}
	for key, value := range body {
		if key != "sig" {
			record[key] = value
		}
	}
	signature := ed25519.Sign(s.private, CanonicalBytes(record))
	publicKey, _ := ParsePublicKey(s.pinned)
	record["sig"] = map[string]any{
		"key_id":    KeyID(publicKey),
		"algorithm": "ed25519",
		"signature": base64.StdEncoding.EncodeToString(signature),
	}
	return record
}

func record(status string) map[string]any {
	return map[string]any{
		"name":            "skill-a",
		"source_identity": "git.example.com/skills/skill-a",
		"commit":          "abc123",
		"content_sha256":  "sha256:deadbeef",
		"status":          status,
		"audit":           map[string]any{"auditor": "team"},
	}
}

// TestCanonicalBytes locks the canonical form: compact sorted JSON without
// sig, non-ASCII preserved.
func TestCanonicalBytes(t *testing.T) {
	payload := map[string]any{
		"b":   "два",
		"a":   1,
		"sig": map[string]any{"signature": "x"},
		"z":   map[string]any{"y": []any{"м", 2, "\b\f<>/&"}, "x": true},
	}
	got := string(CanonicalBytes(payload))
	want := `{"a":1,"b":"два","z":{"x":true,"y":["м",2,"\b\f<>/&"]}}`
	if got != want {
		t.Fatalf("canonical bytes:\n got %s\nwant %s", got, want)
	}
	if _, err := CanonicalBytesChecked(map[string]any{"fraction": 1.5}); err == nil {
		t.Fatal("CCJ-1 must reject non-integer numbers")
	}
}

func TestVerifySigned(t *testing.T) {
	s := newSigner(t)
	signed := s.sign(record(StatusAudited))
	if !VerifySigned(signed, []string{s.pinned}) {
		t.Fatal("valid signature must verify")
	}
	// wrong key
	other := newSigner(t)
	if VerifySigned(signed, []string{other.pinned}) {
		t.Fatal("wrong key must not verify")
	}
	// tampered body
	signed["status"] = StatusRevoked
	if VerifySigned(signed, []string{s.pinned}) {
		t.Fatal("tampered record must not verify")
	}
	// prefix optional
	bare := strings.TrimPrefix(s.pinned, "ed25519:")
	fresh := s.sign(record(StatusAudited))
	if !VerifySigned(fresh, []string{bare}) {
		t.Fatal("bare base64 pinned key must verify")
	}
	fresh["sig"].(map[string]any)["key_id"] = "0000000000000000"
	if VerifySigned(fresh, []string{s.pinned}) {
		t.Fatal("signature key id must bind the verifying key")
	}
	fresh = s.sign(record(StatusAudited))
	fresh["sig"].(map[string]any)["algorithm"] = "rsa"
	if VerifySigned(fresh, []string{s.pinned}) {
		t.Fatal("unsupported signature algorithm must fail")
	}
}

func TestParseRecordValidation(t *testing.T) {
	if _, err := ParseRecord(record(StatusAudited)); err != nil {
		t.Fatal(err)
	}
	bad := record("blessed")
	if _, err := ParseRecord(bad); err == nil {
		t.Fatal("unknown status must fail")
	}
	missing := record(StatusAudited)
	delete(missing, "commit")
	if _, err := ParseRecord(missing); err == nil {
		t.Fatal("missing field must fail")
	}
}

func TestMatches(t *testing.T) {
	parsed, _ := ParseRecord(record(StatusAudited))
	if !Matches(parsed, "git.example.com/skills/skill-a", "abc123", "sha256:other") {
		t.Fatal("identity+commit must match")
	}
	if !Matches(parsed, "other/repo", "zzz", "sha256:deadbeef") {
		t.Fatal("content hash must match")
	}
	if Matches(parsed, "other/repo", "zzz", "sha256:other") {
		t.Fatal("nothing matching must not match")
	}
}

func staticFetch(payloads map[string][]map[string]any, errs map[string]error) FetchFn {
	return func(url, _, _, _ string) ([]map[string]any, error) {
		if err := errs[url]; err != nil {
			return nil, err
		}
		return payloads[url], nil
	}
}

func TestResolveDenyWins(t *testing.T) {
	good := newSigner(t)
	evil := newSigner(t)
	registries := []Registry{
		{Name: "one", URL: "https://one", PublicKeys: []string{good.pinned}},
		{Name: "two", URL: "https://two", PublicKeys: []string{evil.pinned}},
	}
	fetch := staticFetch(map[string][]map[string]any{
		"https://one": {good.sign(record(StatusAudited))},
		"https://two": {evil.sign(record(StatusRevoked))},
	}, nil)
	resolution := Resolve(registries, "git.example.com/skills/skill-a", "abc123", "sha256:deadbeef", fetch)
	if resolution.Result != ResultRevoked {
		t.Fatalf("deny must win: %+v", resolution)
	}
	if resolution.Attestation.Registry != "two" {
		t.Fatalf("attestation: %+v", resolution.Attestation)
	}
}

func TestResolveWarningTaxonomy(t *testing.T) {
	s := newSigner(t)
	registries := []Registry{
		{Name: "keyless", URL: "https://keyless"},
		{Name: "down", URL: "https://down", PublicKeys: []string{s.pinned}},
		{Name: "forged", URL: "https://forged", PublicKeys: []string{s.pinned}},
		{Name: "broken", URL: "https://broken", PublicKeys: []string{s.pinned}},
		{Name: "good", URL: "https://good", PublicKeys: []string{s.pinned}},
	}
	forged := record(StatusAudited)
	forged["sig"] = map[string]any{"signature": base64.StdEncoding.EncodeToString(make([]byte, 64))}
	fetch := staticFetch(map[string][]map[string]any{
		"https://forged": {forged},
		"https://broken": {{"name": "skill-a"}},
		"https://good":   {s.sign(record(StatusAudited))},
	}, map[string]error{"https://down": errFake})
	resolution := Resolve(registries, "git.example.com/skills/skill-a", "abc123", "sha256:deadbeef", fetch)
	if resolution.Result != ResultAudited || resolution.Attestation.Registry != "good" {
		t.Fatalf("resolution: %+v", resolution)
	}
	joined := strings.Join(resolution.Warnings, "\n")
	for _, needle := range []string{"no pinned keys", "unavailable", "failed signature verification", "malformed record"} {
		if !strings.Contains(joined, needle) {
			t.Fatalf("warnings lack %q:\n%s", needle, joined)
		}
	}
}

var errFake = &fakeError{}

type fakeError struct{}

func (*fakeError) Error() string { return "connection refused" }

func TestResolveDeprecatedAndUnknown(t *testing.T) {
	s := newSigner(t)
	registries := []Registry{{Name: "one", URL: "https://one", PublicKeys: []string{s.pinned}}}
	fetch := staticFetch(map[string][]map[string]any{
		"https://one": {s.sign(record(StatusDeprecated))},
	}, nil)
	resolution := Resolve(registries, "git.example.com/skills/skill-a", "abc123", "sha256:deadbeef", fetch)
	if resolution.Result != ResultDeprecated {
		t.Fatalf("resolution: %+v", resolution)
	}
	empty := staticFetch(map[string][]map[string]any{}, nil)
	resolution = Resolve(registries, "x", "y", "z", empty)
	if resolution.Result != ResultUnknown {
		t.Fatalf("resolution: %+v", resolution)
	}
}

func snapshotBody(version int, createdAt time.Time) map[string]any {
	return map[string]any{
		"schema_version": 1, "merkle_root": strings.Repeat("a", 64), "log_size": version,
		"head": strings.Repeat("b", 64), "version": version, "created_at": createdAt.UTC().Format(time.RFC3339),
	}
}

func TestSnapshotVerification(t *testing.T) {
	s := newSigner(t)
	now := time.Now()
	cacheDir := t.TempDir()
	reg := Registry{Name: "one", URL: "https://one", PublicKeys: []string{s.pinned}}

	// valid snapshot passes and persists the version
	fetch := func(string) (map[string]any, error) { return s.sign(snapshotBody(5, now)), nil }
	tampered, warnings := CheckSnapshots([]Registry{reg}, cacheDir, fetch, now, 0)
	if len(tampered) != 0 || len(warnings) != 0 {
		t.Fatalf("valid snapshot rejected: %v %v", tampered, warnings)
	}

	// rollback: version below the persisted highest
	fetch = func(string) (map[string]any, error) { return s.sign(snapshotBody(3, now)), nil }
	tampered, warnings = CheckSnapshots([]Registry{reg}, cacheDir, fetch, now, 0)
	if !tampered[reg.URL] || !strings.Contains(strings.Join(warnings, "\n"), "rollback") {
		t.Fatalf("rollback not detected: %v %v", tampered, warnings)
	}

	// freeze: stale created_at
	fetch = func(string) (map[string]any, error) { return s.sign(snapshotBody(9, now.Add(-30*24*time.Hour))), nil }
	tampered, warnings = CheckSnapshots([]Registry{reg}, cacheDir, fetch, now, 0)
	if !tampered[reg.URL] || !strings.Contains(strings.Join(warnings, "\n"), "stale") {
		t.Fatalf("freeze not detected: %v %v", tampered, warnings)
	}

	// bad signature
	other := newSigner(t)
	fetch = func(string) (map[string]any, error) { return other.sign(snapshotBody(9, now)), nil }
	tampered, _ = CheckSnapshots([]Registry{reg}, cacheDir, fetch, now, 0)
	if !tampered[reg.URL] {
		t.Fatal("forged snapshot not excluded")
	}

	// unreachable warns but does not exclude
	fetch = func(string) (map[string]any, error) { return nil, errFake }
	tampered, warnings = CheckSnapshots([]Registry{reg}, cacheDir, fetch, now, 0)
	if len(tampered) != 0 || !strings.Contains(strings.Join(warnings, "\n"), "unavailable") {
		t.Fatalf("unreachable handling: %v %v", tampered, warnings)
	}
}

func TestSnapshotRequiresCompleteShapeAndRejectsEquivocation(t *testing.T) {
	s := newSigner(t)
	now := time.Now().UTC().Truncate(time.Second)
	cacheDir := t.TempDir()
	reg := Registry{Name: "one", URL: "https://one", PublicKeys: []string{s.pinned}}

	valid := snapshotBody(5, now)
	fetch := func(string) (map[string]any, error) { return s.sign(valid), nil }
	tampered, _ := CheckSnapshots([]Registry{reg}, cacheDir, fetch, now, 0)
	if len(tampered) != 0 {
		t.Fatal("valid snapshot rejected")
	}

	equivocated := snapshotBody(5, now)
	equivocated["head"] = strings.Repeat("c", 64)
	fetch = func(string) (map[string]any, error) { return s.sign(equivocated), nil }
	tampered, warnings := CheckSnapshots([]Registry{reg}, cacheDir, fetch, now, 0)
	if !tampered[reg.URL] || !strings.Contains(strings.Join(warnings, "\n"), "equivocation") {
		t.Fatalf("equivocation not detected: %v %v", tampered, warnings)
	}

	incomplete := snapshotBody(6, now)
	delete(incomplete, "merkle_root")
	fetch = func(string) (map[string]any, error) { return s.sign(incomplete), nil }
	tampered, warnings = CheckSnapshots([]Registry{reg}, cacheDir, fetch, now, 0)
	if !tampered[reg.URL] || !strings.Contains(strings.Join(warnings, "\n"), "malformed") {
		t.Fatalf("incomplete snapshot not rejected: %v %v", tampered, warnings)
	}

	future := snapshotBody(7, now.Add(DefaultSnapshotClockSkew+time.Second))
	fetch = func(string) (map[string]any, error) { return s.sign(future), nil }
	tampered, warnings = CheckSnapshots([]Registry{reg}, cacheDir, fetch, now, 0)
	if !tampered[reg.URL] || !strings.Contains(strings.Join(warnings, "\n"), "future") {
		t.Fatalf("future snapshot not rejected: %v %v", tampered, warnings)
	}
}

func TestHTTPFetchCacheAndOfflineGrace(t *testing.T) {
	s := newSigner(t)
	calls := 0
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		if calls > 1 {
			// simulate outage after the first response
			server.CloseClientConnections()
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"records": []any{s.sign(record(StatusAudited))}, "next_cursor": nil})
	}))
	defer server.Close()

	current := time.Now()
	now := func() time.Time { return current }
	fetch := NewHTTPFetch(t.TempDir(), time.Minute, time.Hour, now)

	first, err := fetch(server.URL, "id", "c", "h")
	if err != nil || len(first) != 1 {
		t.Fatalf("first fetch: %v %v", first, err)
	}
	// within TTL: served from cache, no second call
	second, err := fetch(server.URL, "id", "c", "h")
	if err != nil || len(second) != 1 || calls != 1 {
		t.Fatalf("cache miss within TTL: calls=%d err=%v", calls, err)
	}
	// past TTL but within grace: refresh fails, stale cache serves
	current = current.Add(2 * time.Minute)
	third, err := fetch(server.URL, "id", "c", "h")
	if err != nil || len(third) != 1 {
		t.Fatalf("offline grace failed: %v %v", third, err)
	}
	// past grace: registry counts as unreachable
	current = current.Add(2 * time.Hour)
	if _, err := fetch(server.URL, "id", "c", "h"); err == nil {
		t.Fatal("past grace must error")
	}
}

func TestHTTPFailuresUseOfflineCacheAndRejectSnapshots(t *testing.T) {
	s := newSigner(t)
	failed := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if failed {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"records":[]}`))
			return
		}
		if strings.HasSuffix(r.URL.Path, "/v1/snapshot") {
			_ = json.NewEncoder(w).Encode(s.sign(snapshotBody(1, time.Now())))
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"records": []any{s.sign(record(StatusAudited))}, "next_cursor": nil})
	}))
	defer server.Close()

	current := time.Now()
	fetch := NewHTTPFetch(t.TempDir(), time.Minute, time.Hour, func() time.Time { return current })
	first, err := fetch(server.URL, "id", "commit", "hash")
	if err != nil || len(first) != 1 {
		t.Fatalf("initial records fetch: records=%v err=%v", first, err)
	}
	failed = true
	current = current.Add(2 * time.Minute)
	cached, err := fetch(server.URL, "id", "commit", "hash")
	if err != nil || len(cached) != 1 {
		t.Fatalf("HTTP 503 must use offline cache: records=%v err=%v", cached, err)
	}
	if _, err := HTTPGetSnapshot(server.URL); err == nil || !strings.Contains(err.Error(), "503") {
		t.Fatalf("snapshot HTTP failure = %v, want 503 error", err)
	}
}

func TestHTTPFetchFollowsBoundedPagination(t *testing.T) {
	s := newSigner(t)
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("limit") != "1000" {
			t.Errorf("limit = %q", r.URL.Query().Get("limit"))
		}
		if r.URL.Query().Get("cursor") == "" {
			_ = json.NewEncoder(w).Encode(map[string]any{"records": []any{s.sign(record(StatusAudited))}, "next_cursor": "page-2"})
			return
		}
		if r.URL.Query().Get("cursor") != "page-2" {
			t.Errorf("cursor = %q", r.URL.Query().Get("cursor"))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"records": []any{s.sign(record(StatusDeprecated))}, "next_cursor": nil})
	}))
	defer server.Close()
	fetch := NewHTTPFetch(t.TempDir(), time.Minute, time.Hour, time.Now)
	records, err := fetch(server.URL, "id", "commit", "hash")
	if err != nil || len(records) != 2 || calls != 2 {
		t.Fatalf("records=%d calls=%d err=%v", len(records), calls, err)
	}
}

func TestPublish(t *testing.T) {
	var gotAuth string
	var gotIdempotency string
	var gotBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotIdempotency = r.Header.Get("Idempotency-Key")
		gotBody, _ = json.Marshal(map[string]any{"seq": 1, "entry_hash": "abc"})
		payload, _ := json.Marshal(map[string]any{"seq": 1, "entry_hash": "abc"})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write(payload)
	}))
	defer server.Close()
	response, err := Publish(server.URL, "secret-token", []byte(`{"name": "skill-a"}`))
	if err != nil {
		t.Fatal(err)
	}
	_ = gotBody
	if gotAuth != "Bearer secret-token" {
		t.Fatalf("auth header: %q", gotAuth)
	}
	if len(gotIdempotency) != 64 {
		t.Fatalf("idempotency header: %q", gotIdempotency)
	}
	if !strings.Contains(response, "entry_hash") {
		t.Fatalf("response: %s", response)
	}
	if _, err := Publish(server.URL, "t", []byte("{broken")); err == nil {
		t.Fatal("invalid JSON must fail before the network")
	}
	reject := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "invalid auditor token", http.StatusUnauthorized)
	}))
	defer reject.Close()
	if _, err := Publish(reject.URL, "bad", []byte(`{}`)); err == nil || !strings.Contains(err.Error(), "401") {
		t.Fatalf("err = %v, want rejection with status", err)
	}
}
